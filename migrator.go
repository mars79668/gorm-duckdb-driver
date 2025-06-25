package duckdb

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Migrator struct {
	migrator.Migrator
}

func (m Migrator) CurrentDatabase() (name string) {
	_ = m.DB.Raw("SELECT current_database()").Row().Scan(&name)
	return
}

func (m Migrator) FullDataTypeOf(field *schema.Field) clause.Expr {
	expr := clause.Expr{}
	expr.SQL = m.DataTypeOf(field)

	if field.NotNull {
		expr.SQL += " NOT NULL"
	}

	if field.Unique {
		expr.SQL += " UNIQUE"
	}

	// Handle auto-increment for primary key fields using sequences
	if field.AutoIncrement && field.PrimaryKey {
		sequenceName := m.DB.NamingStrategy.IndexName(field.Schema.Table, field.DBName) + "_seq"
		expr.SQL += " DEFAULT nextval('" + sequenceName + "')"
	}

	if field.HasDefaultValue && (field.DefaultValueInterface != nil || field.DefaultValue != "") {
		if field.DefaultValueInterface != nil {
			defaultStmt := &gorm.Statement{Vars: []interface{}{field.DefaultValueInterface}}
			m.Dialector.BindVarTo(defaultStmt, defaultStmt, field.DefaultValueInterface)
			expr.SQL += " DEFAULT " + m.Dialector.Explain(defaultStmt.SQL.String(), field.DefaultValueInterface)
		} else if field.DefaultValue != "(-)" {
			expr.SQL += " DEFAULT " + field.DefaultValue
		}
	}

	if field.Comment != "" {
		expr.SQL += " COMMENT '" + field.Comment + "'"
	}

	return expr
}

func (m Migrator) AlterColumn(value interface{}, field string) error {
	return m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				fileType := m.FullDataTypeOf(field)
				return m.Migrator.DB.Exec(
					"ALTER TABLE ? ALTER COLUMN ? TYPE ?",
					m.Migrator.CurrentTable(stmt), clause.Column{Name: field.DBName}, fileType,
				).Error
			}
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	})
}

func (m Migrator) RenameColumn(value interface{}, oldName, newName string) error {
	return m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(oldName); field != nil {
				oldName = field.DBName
			}

			if field := stmt.Schema.LookUpField(newName); field != nil {
				newName = field.DBName
			}
		}

		return m.Migrator.DB.Exec(
			"ALTER TABLE ? RENAME COLUMN ? TO ?",
			m.Migrator.CurrentTable(stmt), clause.Column{Name: oldName}, clause.Column{Name: newName},
		).Error
	})
}

func (m Migrator) RenameIndex(value interface{}, oldName, newName string) error {
	return m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.Migrator.DB.Exec(
			"ALTER INDEX ? RENAME TO ?",
			clause.Column{Name: oldName}, clause.Column{Name: newName},
		).Error
	})
}

func (m Migrator) DropIndex(value interface{}, name string) error {
	return m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}

		return m.Migrator.DB.Exec("DROP INDEX IF EXISTS ?", clause.Column{Name: name}).Error
	})
}

func (m Migrator) DropConstraint(value interface{}, name string) error {
	return m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, table := m.Migrator.GuessConstraintInterfaceAndTable(stmt, name)
		if constraint != nil {
			name = constraint.GetName()
		}
		return m.Migrator.DB.Exec("ALTER TABLE ? DROP CONSTRAINT ?", clause.Table{Name: table}, clause.Column{Name: name}).Error
	})
}

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

func (m Migrator) GetTables() (tableList []string, err error) {
	err = m.DB.Raw(
		"SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE'",
	).Scan(&tableList).Error
	return
}

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

func (m Migrator) HasConstraint(value interface{}, name string) bool {
	var count int64
	m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, table := m.Migrator.GuessConstraintInterfaceAndTable(stmt, name)
		if constraint != nil {
			name = constraint.GetName()
		}

		return m.Migrator.DB.Raw(
			"SELECT count(*) FROM information_schema.table_constraints WHERE table_name = ? AND constraint_name = ?",
			table, name,
		).Row().Scan(&count)
	})

	return count > 0
}

// CreateTable creates table with auto-increment sequence support
func (m Migrator) CreateTable(values ...interface{}) error {
	for _, value := range values {
		if err := m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
			// Create sequences for auto-increment fields before creating the table
			for _, field := range stmt.Schema.Fields {
				if field.AutoIncrement && field.PrimaryKey {
					sequenceName := m.Migrator.DB.NamingStrategy.IndexName(stmt.Table, field.DBName) + "_seq"
					createSeqSQL := "CREATE SEQUENCE IF NOT EXISTS " + sequenceName
					if err := m.Migrator.DB.Exec(createSeqSQL).Error; err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	// Now call the embedded migrator's CreateTable method
	return m.Migrator.CreateTable(values...)
}

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
	m.Migrator.QuoteTo(sql, name)
	sql.WriteString(" AS ")

	m.Migrator.DB.Statement.AddVar(sql, option.Query)

	if option.CheckOption != "" {
		sql.WriteString(" ")
		sql.WriteString(option.CheckOption)
	}

	return m.Migrator.DB.Exec(m.Migrator.Explain(sql.String(), m.Migrator.DB.Statement.Vars...)).Error
}

func (m Migrator) DropView(name string) error {
	return m.Migrator.DB.Exec("DROP VIEW IF EXISTS ?", clause.Table{Name: name}).Error
}

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

	if typeAliases, ok := aliases[strings.ToLower(databaseTypeName)]; ok {
		return typeAliases
	}

	return []string{}
}
