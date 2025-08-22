package duckdb

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

const (
	// SQL data type constants
	sqlTypeBigInt  = "BIGINT"
	sqlTypeInteger = "INTEGER"
)

// isAlreadyExistsError checks if an error indicates that an object already exists
func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "duplicate")
}

// Migrator implements gorm.Migrator interface for DuckDB database.
type Migrator struct {
	migrator.Migrator
}

// isAutoIncrementField checks if a field is an auto-increment field
func (m Migrator) isAutoIncrementField(field *schema.Field) bool {
	return field.AutoIncrement || (!field.HasDefaultValue && field.DataType == schema.Uint)
}

// CurrentDatabase returns the current database name.
func (m Migrator) CurrentDatabase() (name string) {
	_ = m.DB.Raw("SELECT current_database()").Row().Scan(&name)
	return
}

// FullDataTypeOf returns the full data type for a field including constraints.
// Override FullDataTypeOf to prevent GORM from adding duplicate PRIMARY KEY clauses
func (m Migrator) FullDataTypeOf(field *schema.Field) clause.Expr {
	// Get the base data type from our dialector
	dataType := m.Dialector.DataTypeOf(field)

	expr := clause.Expr{SQL: dataType}

	// For primary key fields, ensure clean type definition without duplicate PRIMARY KEY
	if field.PrimaryKey {
		// DuckDB doesn't support native AUTO_INCREMENT, so we use sequences to emulate this behavior for auto-increment primary keys
		// Check if this is an auto-increment field (no default value specified)
		if m.isAutoIncrementField(field) {
			// Use BIGINT with a default sequence value
			expr.SQL = "BIGINT DEFAULT nextval('seq_" + strings.ToLower(field.Schema.Table) + "_" + strings.ToLower(field.DBName) + "')"
		} else {
			// Make sure the data type is clean for non-auto-increment primary keys
			upperDataType := strings.ToUpper(dataType)
			switch {
			case strings.Contains(upperDataType, sqlTypeBigInt):
				expr.SQL = sqlTypeBigInt
			case strings.Contains(upperDataType, sqlTypeInteger):
				expr.SQL = sqlTypeInteger
			default:
				expr.SQL = dataType
			}
		}

		// Add NOT NULL for primary keys
		expr.SQL += " NOT NULL"

		// Do NOT add PRIMARY KEY here - let GORM handle it in the table definition
		return expr
	}

	// For non-primary key fields, add constraints
	if field.NotNull {
		expr.SQL += " NOT NULL"
	}

	if field.Unique {
		expr.SQL += " UNIQUE"
	}

	// Handle defaults for non-primary key fields only
	if field.HasDefaultValue && (field.DefaultValueInterface != nil || field.DefaultValue != "") {
		if field.DefaultValueInterface != nil {
			defaultStmt := &gorm.Statement{Vars: []interface{}{field.DefaultValueInterface}}
			m.BindVarTo(defaultStmt, defaultStmt, field.DefaultValueInterface)
			expr.SQL += " DEFAULT " + m.Explain(defaultStmt.SQL.String(), field.DefaultValueInterface)
		} else if field.DefaultValue != "(-)" {
			expr.SQL += " DEFAULT " + field.DefaultValue
		}
	}

	if field.Comment != "" {
		expr.SQL += " COMMENT '" + field.Comment + "'"
	}

	return expr
}

// AlterColumn modifies a column definition in DuckDB, handling syntax limitations.
func (m Migrator) AlterColumn(value interface{}, field string) error {
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				// For ALTER COLUMN, only use the base data type without defaults
				baseType := m.Dialector.DataTypeOf(field)

				// Clean the base type - remove any DEFAULT clauses
				baseType = strings.Split(baseType, " DEFAULT")[0]

				return m.DB.Exec(
					"ALTER TABLE ? ALTER COLUMN ? TYPE ?",
					m.CurrentTable(stmt), clause.Column{Name: field.DBName}, clause.Expr{SQL: baseType},
				).Error
			}
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	})
	if err != nil {
		return fmt.Errorf("failed to alter column: %w", err)
	}
	return nil
}

// RenameColumn renames a column in the database table.
func (m Migrator) RenameColumn(value interface{}, oldName, newName string) error {
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(oldName); field != nil {
				oldName = field.DBName
			}

			if field := stmt.Schema.LookUpField(newName); field != nil {
				newName = field.DBName
			}
		}

		return m.DB.Exec(
			"ALTER TABLE ? RENAME COLUMN ? TO ?",
			m.CurrentTable(stmt), clause.Column{Name: oldName}, clause.Column{Name: newName},
		).Error
	})
	if err != nil {
		return fmt.Errorf("failed to rename column: %w", err)
	}
	return nil
}

// RenameIndex renames an index in the database.
func (m Migrator) RenameIndex(value interface{}, oldName, newName string) error {
	err := m.RunWithValue(value, func(_ *gorm.Statement) error {
		return m.DB.Exec(
			"ALTER INDEX ? RENAME TO ?",
			clause.Column{Name: oldName}, clause.Column{Name: newName},
		).Error
	})
	if err != nil {
		return fmt.Errorf("failed to rename index: %w", err)
	}
	return nil
}

// DropIndex drops an index from the database.
func (m Migrator) DropIndex(value interface{}, name string) error {
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}

		return m.DB.Exec("DROP INDEX IF EXISTS ?", clause.Column{Name: name}).Error
	})
	if err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}
	return nil
}

// DropConstraint drops a constraint from the database.
func (m Migrator) DropConstraint(value interface{}, name string) error {
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, table := m.GuessConstraintInterfaceAndTable(stmt, name)
		if constraint != nil {
			name = constraint.GetName()
		}
		return m.Migrator.DB.Exec("ALTER TABLE ? DROP CONSTRAINT ?", clause.Table{Name: table}, clause.Column{Name: name}).Error
	})
	if err != nil {
		return fmt.Errorf("failed to drop constraint: %w", err)
	}
	return nil
}

// HasTable checks if a table exists in the database.
func (m Migrator) HasTable(value interface{}) bool {
	var count int64

	_ = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.DB.Raw(
			"SELECT count(*) FROM information_schema.tables WHERE table_name = ? AND table_type = 'BASE TABLE'",
			stmt.Table,
		).Row().Scan(&count)
	})

	return count > 0
}

// GetTables returns a list of all table names in the database.
func (m Migrator) GetTables() (tableList []string, err error) {
	err = m.DB.Raw(
		"SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE'",
	).Scan(&tableList).Error
	return
}

// HasColumn checks if a column exists in the database table.
func (m Migrator) HasColumn(value interface{}, field string) bool {
	var count int64
	_ = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		name := field
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				name = field.DBName
			}
		}

		return m.DB.Raw(
			"SELECT count(*) FROM information_schema.columns WHERE table_name = ? AND column_name = ?",
			stmt.Table, name,
		).Row().Scan(&count)
	})

	return count > 0
}

// HasIndex checks if an index exists in the database.
func (m Migrator) HasIndex(value interface{}, name string) bool {
	var count int64
	_ = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}

		return m.DB.Raw(
			"SELECT count(*) FROM information_schema.statistics WHERE table_name = ? AND index_name = ?",
			stmt.Table, name,
		).Row().Scan(&count)
	})

	return count > 0
}

// HasConstraint checks if a constraint exists in the database.
func (m Migrator) HasConstraint(value interface{}, name string) bool {
	var count int64
	_ = m.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, table := m.GuessConstraintInterfaceAndTable(stmt, name)
		if constraint != nil {
			name = constraint.GetName()
		}

		return m.DB.Raw(
			"SELECT count(*) FROM information_schema.table_constraints WHERE table_name = ? AND constraint_name = ?",
			table, name,
		).Row().Scan(&count)
	})

	return count > 0
}

// CreateView creates a database view.
func (m Migrator) CreateView(name string, option gorm.ViewOption) error {
	if option.Query == nil {
		return gorm.ErrSubQueryRequired
	}

	sql := new(strings.Builder)
	sql.WriteString("CREATE ")
	if option.Replace {
		sql.WriteString("OR REPLACE ")
	}
	sql.WriteString("VIEW ")
	m.QuoteTo(sql, name)
	sql.WriteString(" AS ")

	m.DB.Statement.AddVar(sql, option.Query)

	if option.CheckOption != "" {
		sql.WriteString(" ")
		sql.WriteString(option.CheckOption)
	}

	return m.DB.Exec(m.Explain(sql.String(), m.DB.Statement.Vars...)).Error
}

// DropView drops a database view.
func (m Migrator) DropView(name string) error {
	return m.DB.Exec("DROP VIEW IF EXISTS ?", clause.Table{Name: name}).Error
}

// GetTypeAliases returns type aliases for the given database type name.
func (m Migrator) GetTypeAliases(databaseTypeName string) []string {
	aliases := map[string][]string{
		"boolean":   {"bool"},
		"tinyint":   {"int8"},
		"smallint":  {"int16"},
		"integer":   {"int", "int32"},
		"bigint":    {"int64"},
		"utinyint":  {"uint8"},
		"usmallint": {"uint16"},
		"uinteger":  {"uint", "uint32"},
		"ubigint":   {"uint64"},
		"real":      {"float32"},
		"double":    {"float64", "float"},
		"varchar":   {"string"},
		"text":      {"string"},
		"blob":      {"bytes"},
		"timestamp": {"time"},
	}

	return aliases[databaseTypeName]
}

// ColumnTypes returns comprehensive column type information for the given value
func (m Migrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	var columnTypes []gorm.ColumnType

	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		rows, err := m.DB.Raw(`
			SELECT 
				column_name,
				data_type,
				CASE 
					WHEN character_maximum_length IS NOT NULL THEN data_type || '(' || character_maximum_length || ')'
					WHEN numeric_precision IS NOT NULL AND numeric_scale IS NOT NULL THEN data_type || '(' || numeric_precision || ',' || numeric_scale || ')'
					WHEN numeric_precision IS NOT NULL THEN data_type || '(' || numeric_precision || ')'
					ELSE data_type
				END as column_type,
				CASE WHEN is_nullable = 'YES' THEN true ELSE false END as nullable,
				column_default,
				CASE WHEN column_key = 'PRI' THEN true ELSE false END as is_primary_key,
				CASE WHEN extra LIKE '%auto_increment%' OR column_default LIKE '%nextval%' THEN true ELSE false END as is_auto_increment,
				character_maximum_length,
				numeric_precision,
				numeric_scale,
				CASE WHEN column_key = 'UNI' THEN true ELSE false END as is_unique,
				column_comment
			FROM information_schema.columns 
			WHERE table_name = ?
			ORDER BY ordinal_position
		`, stmt.Table).Rows()

		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				columnName, dataType, columnTypeStr, columnDefault, columnComment string
				isNullable, isPrimaryKey, isAutoIncrement, isUnique               bool
				charMaxLength, numericPrecision, numericScale                     sql.NullInt64
			)

			err = rows.Scan(
				&columnName, &dataType, &columnTypeStr, &isNullable, &columnDefault,
				&isPrimaryKey, &isAutoIncrement, &charMaxLength, &numericPrecision,
				&numericScale, &isUnique, &columnComment,
			)
			if err != nil {
				continue
			}

			columnType := &migrator.ColumnType{
				NameValue:          sql.NullString{String: columnName, Valid: true},
				DataTypeValue:      sql.NullString{String: dataType, Valid: true},
				ColumnTypeValue:    sql.NullString{String: columnTypeStr, Valid: true},
				NullableValue:      sql.NullBool{Bool: isNullable, Valid: true},
				PrimaryKeyValue:    sql.NullBool{Bool: isPrimaryKey, Valid: true},
				AutoIncrementValue: sql.NullBool{Bool: isAutoIncrement, Valid: true},
				UniqueValue:        sql.NullBool{Bool: isUnique, Valid: true},
				CommentValue:       sql.NullString{String: columnComment, Valid: columnComment != ""},
				DefaultValueValue:  sql.NullString{String: columnDefault, Valid: columnDefault != ""},
			}

			// Set length information
			if charMaxLength.Valid {
				columnType.LengthValue = charMaxLength
			}

			// Set decimal size information
			if numericPrecision.Valid {
				columnType.DecimalSizeValue = numericPrecision
				if numericScale.Valid {
					columnType.ScaleValue = numericScale
				}
			}

			columnTypes = append(columnTypes, columnType)
		}

		return rows.Err()
	})

	return columnTypes, err
}

// TableType returns comprehensive table type information
func (m Migrator) TableType(value interface{}) (gorm.TableType, error) {
	var result *migrator.TableType

	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		var schemaName, tableName, tableTypeStr, tableComment string

		err := m.DB.Raw(`
			SELECT 
				table_schema,
				table_name,
				table_type,
				COALESCE(table_comment, '') as table_comment
			FROM information_schema.tables 
			WHERE table_name = ?
		`, stmt.Table).Row().Scan(&schemaName, &tableName, &tableTypeStr, &tableComment)

		if err != nil {
			return err
		}

		result = &migrator.TableType{
			SchemaValue:  schemaName,
			NameValue:    tableName,
			TypeValue:    tableTypeStr,
			CommentValue: sql.NullString{String: tableComment, Valid: tableComment != ""},
		}

		return nil
	})

	if result == nil {
		return nil, err
	}

	return result, err
}

// DuckDBIndex implements gorm.Index interface for DuckDB
type DuckDBIndex struct {
	TableName   string
	IndexName   string
	ColumnNames []string
	IsUnique    bool
	IsPrimary   bool
	Options     string
}

func (idx DuckDBIndex) Table() string {
	return idx.TableName
}

func (idx DuckDBIndex) Name() string {
	return idx.IndexName
}

func (idx DuckDBIndex) Columns() []string {
	return idx.ColumnNames
}

func (idx DuckDBIndex) PrimaryKey() (isPrimaryKey bool, ok bool) {
	return idx.IsPrimary, true
}

func (idx DuckDBIndex) Unique() (unique bool, ok bool) {
	return idx.IsUnique, true
}

func (idx DuckDBIndex) Option() string {
	return idx.Options
}

// GetIndexes returns comprehensive index information for the given value
func (m Migrator) GetIndexes(value interface{}) ([]gorm.Index, error) {
	var indexes []gorm.Index

	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		// DuckDB may not have complete information_schema.statistics support
		// For now, return empty indexes to avoid errors
		return nil
	})

	return indexes, err
}

// BuildIndexOptions builds index options for DuckDB
func (m Migrator) BuildIndexOptions(opts []schema.IndexOption, stmt *gorm.Statement) (results []interface{}) {
	for _, opt := range opts {
		str := stmt.Quote(opt.DBName)
		if opt.Expression != "" {
			str = opt.Expression
		} else if opt.Length > 0 {
			str += fmt.Sprintf("(%d)", opt.Length)
		}

		if opt.Collate != "" {
			str += " COLLATE " + opt.Collate
		}

		if opt.Sort != "" {
			str += " " + opt.Sort
		}
		results = append(results, clause.Expr{SQL: str})
	}
	return
}

// CreateTable overrides the default CreateTable to handle DuckDB-specific auto-increment sequences
func (m Migrator) CreateTable(values ...interface{}) error {
	for _, value := range values {
		if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
			// First, create sequences for auto-increment primary key fields
			if stmt.Schema != nil {
				for _, field := range stmt.Schema.Fields {
					if field.PrimaryKey && (field.AutoIncrement || (!field.HasDefaultValue && field.DataType == schema.Uint)) {
						sequenceName := "seq_" + strings.ToLower(stmt.Schema.Table) + "_" + strings.ToLower(field.DBName)
						createSeqSQL := fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s START 1", sequenceName)
						if err := m.DB.Exec(createSeqSQL).Error; err != nil {
							// Ignore "already exists" errors
							if !isAlreadyExistsError(err) {
								return fmt.Errorf("failed to create sequence %s: %w", sequenceName, err)
							}
						}
					}
				}
			}

			// Now create the table using the parent method
			return m.Migrator.CreateTable(value)
		}); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	return nil
}
