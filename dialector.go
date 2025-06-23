package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Dialector struct {
	*Config
}

type Config struct {
	DriverName        string
	DSN               string
	Conn              gorm.ConnPool
	DefaultStringSize uint
}

func Open(dsn string) gorm.Dialector {
	return &Dialector{Config: &Config{DSN: dsn}}
}

func New(config Config) gorm.Dialector {
	return &Dialector{Config: &config}
}

func (dialector Dialector) Name() string {
	return "duckdb"
}

func (dialector Dialector) Initialize(db *gorm.DB) (err error) {
	if dialector.DriverName == "" {
		dialector.DriverName = "duckdb"
	}

	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else {
		db.ConnPool, err = sql.Open(dialector.DriverName, dialector.DSN)
	}

	if dialector.DefaultStringSize == 0 {
		dialector.DefaultStringSize = 256
	}

	// Register callbacks first
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{
		CreateClauses: []string{"INSERT", "VALUES", "ON CONFLICT", "RETURNING"},
		UpdateClauses: []string{"UPDATE", "SET", "WHERE", "RETURNING"},
		DeleteClauses: []string{"DELETE", "FROM", "WHERE", "RETURNING"},
		QueryClauses:  []string{"SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT", "FOR"},
	})

	// Wrap the connection to handle time pointer conversion
	if db.ConnPool != nil {
		db.ConnPool = &duckdbConnPoolWrapper{db.ConnPool}
	}

	return
}

func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		migrator.Migrator{
			Config: migrator.Config{
				DB:                          db,
				Dialector:                   dialector,
				CreateIndexAfterCreateTable: true,
			},
		},
	}
}

func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "BOOLEAN"
	case schema.Int:
		switch field.Size {
		case 8:
			return "TINYINT"
		case 16:
			return "SMALLINT"
		case 32:
			return "INTEGER"
		default:
			return "BIGINT"
		}
	case schema.Uint:
		switch field.Size {
		case 8:
			return "UTINYINT"
		case 16:
			return "USMALLINT"
		case 32:
			return "UINTEGER"
		default:
			return "UBIGINT"
		}
	case schema.Float:
		if field.Size == 32 {
			return "REAL"
		}
		return "DOUBLE"
	case schema.String:
		size := field.Size
		if size == 0 {
			size = int(dialector.DefaultStringSize)
		}
		if size > 0 && size < 65536 {
			return fmt.Sprintf("VARCHAR(%d)", size)
		}
		return "TEXT"
	case schema.Time:
		// Handle time types similar to other GORM dialectors
		// DuckDB supports TIMESTAMP for datetime values
		if field.NotNull || field.PrimaryKey {
			return "TIMESTAMP"
		}
		return "TIMESTAMP"
	case schema.Bytes:
		return "BLOB"
	}

	return string(field.DataType)
}

func (dialector Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	if field.HasDefaultValue && (field.DefaultValueInterface != nil || field.DefaultValue != "") {
		if field.DefaultValueInterface != nil {
			switch field.DefaultValueInterface.(type) {
			case bool:
				if field.DefaultValueInterface.(bool) {
					return clause.Expr{SQL: "TRUE"}
				}
				return clause.Expr{SQL: "FALSE"}
			default:
				return clause.Expr{SQL: fmt.Sprintf("'%v'", field.DefaultValueInterface)}
			}
		} else if field.DefaultValue != "" && field.DefaultValue != "(-)" {
			if field.DataType == schema.Bool {
				if strings.ToLower(field.DefaultValue) == "true" {
					return clause.Expr{SQL: "TRUE"}
				}
				return clause.Expr{SQL: "FALSE"}
			}
			return clause.Expr{SQL: field.DefaultValue}
		}
	}
	return clause.Expr{}
}

func (dialector Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	writer.WriteByte('?')
}

func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	var (
		underQuoted, selfQuoted bool
		continuousBacktick      int8
		shiftDelimiter          int8
	)

	for _, v := range []byte(str) {
		switch v {
		case '"':
			continuousBacktick++
			if continuousBacktick == 2 {
				writer.WriteString(`""`)
				continuousBacktick = 0
			}
		case '.':
			if continuousBacktick > 0 || !selfQuoted {
				shiftDelimiter = 0
				underQuoted = false
				continuousBacktick = 0
				writer.WriteByte('"')
			}
			writer.WriteByte(v)
			continue
		default:
			if shiftDelimiter-continuousBacktick <= 0 && !underQuoted {
				writer.WriteByte('"')
				underQuoted = true
				if selfQuoted = continuousBacktick > 0; selfQuoted {
					continuousBacktick -= 1
				}
			}

			for ; continuousBacktick > 0; continuousBacktick -= 1 {
				writer.WriteString(`""`)
			}

			writer.WriteByte(v)
		}
		shiftDelimiter++
	}

	if continuousBacktick > 0 && !selfQuoted {
		writer.WriteString(`""`)
	}
	writer.WriteByte('"')
}

func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `"`, vars...)
}

func (dialector Dialector) SavePoint(tx *gorm.DB, name string) error {
	return tx.Exec("SAVEPOINT " + name).Error
}

func (dialector Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return tx.Exec("ROLLBACK TO SAVEPOINT " + name).Error
}

// convertTimePointers converts *time.Time values to time.Time for DuckDB compatibility
func convertTimePointers(args []interface{}) []interface{} {
	if args == nil {
		return args
	}

	converted := make([]interface{}, len(args))
	for i, arg := range args {
		if arg == nil {
			converted[i] = nil
			continue
		}

		// Convert *time.Time to time.Time for DuckDB driver
		if timePtr, ok := arg.(*time.Time); ok {
			if timePtr == nil {
				converted[i] = nil
			} else {
				converted[i] = *timePtr
			}
		} else {
			converted[i] = arg
		}
	}
	return converted
}

// duckdbConnPoolWrapper wraps the connection pool to return wrapped connections
type duckdbConnPoolWrapper struct {
	gorm.ConnPool
}

func (p *duckdbConnPoolWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.ConnPool.PrepareContext(ctx, query)
}

func (p *duckdbConnPoolWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	convertedArgs := convertTimePointers(args)
	return p.ConnPool.ExecContext(ctx, query, convertedArgs...)
}

func (p *duckdbConnPoolWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	convertedArgs := convertTimePointers(args)
	return p.ConnPool.QueryContext(ctx, query, convertedArgs...)
}

func (p *duckdbConnPoolWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	convertedArgs := convertTimePointers(args)
	return p.ConnPool.QueryRowContext(ctx, query, convertedArgs...)
}

// Implement GetDBConnector interface to allow access to underlying *sql.DB
func (p *duckdbConnPoolWrapper) GetDBConnector() (*sql.DB, error) {
	if dbConnector, ok := p.ConnPool.(interface{ GetDBConnector() (*sql.DB, error) }); ok {
		return dbConnector.GetDBConnector()
	}

	// If the wrapped ConnPool is directly *sql.DB, return it
	if db, ok := p.ConnPool.(*sql.DB); ok {
		return db, nil
	}

	return nil, fmt.Errorf("unable to get underlying *sql.DB from connection pool")
}
