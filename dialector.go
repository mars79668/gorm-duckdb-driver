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
		return dialector.getIntType(field.Size)
	case schema.Uint:
		return dialector.getUintType(field.Size)
	case schema.Float:
		return dialector.getFloatType(field.Size)
	case schema.String:
		return dialector.getStringType(field.Size)
	case schema.Time:
		return "TIMESTAMP"
	case schema.Bytes:
		return "BLOB"
	default:
		return string(field.DataType)
	}
}

func (dialector Dialector) getIntType(size int) string {
	switch size {
	case 8:
		return "TINYINT"
	case 16:
		return "SMALLINT"
	case 32:
		return "INTEGER"
	default:
		return "BIGINT"
	}
}

func (dialector Dialector) getUintType(size int) string {
	switch size {
	case 8:
		return "UTINYINT"
	case 16:
		return "USMALLINT"
	case 32:
		return "UINTEGER"
	default:
		return "UBIGINT"
	}
}

func (dialector Dialector) getFloatType(size int) string {
	if size == 32 {
		return "REAL"
	}
	return "DOUBLE"
}

func (dialector Dialector) getStringType(size int) string {
	if size == 0 {
		if dialector.DefaultStringSize > 0 && dialector.DefaultStringSize <= 65535 {
			// #nosec G115 - bounds already checked above
			size = int(dialector.DefaultStringSize)
		}
	}
	if size > 0 && size < 65536 {
		return fmt.Sprintf("VARCHAR(%d)", size)
	}
	return "TEXT"
}

func (dialector Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	if !field.HasDefaultValue {
		return clause.Expr{}
	}

	if field.DefaultValueInterface != nil {
		return dialector.getDefaultFromInterface(field.DefaultValueInterface)
	}

	if field.DefaultValue != "" && field.DefaultValue != "(-)" {
		return dialector.getDefaultFromString(field.DefaultValue, field.DataType)
	}

	return clause.Expr{}
}

func (dialector Dialector) getDefaultFromInterface(defaultValue interface{}) clause.Expression {
	switch v := defaultValue.(type) {
	case bool:
		if v {
			return clause.Expr{SQL: "TRUE"}
		}
		return clause.Expr{SQL: "FALSE"}
	default:
		return clause.Expr{SQL: fmt.Sprintf("'%v'", v)}
	}
}

func (dialector Dialector) getDefaultFromString(defaultValue string, dataType schema.DataType) clause.Expression {
	if dataType == schema.Bool {
		if strings.ToLower(defaultValue) == "true" {
			return clause.Expr{SQL: "TRUE"}
		}
		return clause.Expr{SQL: "FALSE"}
	}
	return clause.Expr{SQL: defaultValue}
}

func (dialector Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	_ = writer.WriteByte('?')
}

func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	_ = writer.WriteByte('"')

	for _, v := range []byte(str) {
		switch v {
		case '"':
			_, _ = writer.WriteString(`""`)
		case '.':
			_ = writer.WriteByte('"')
			_ = writer.WriteByte(v)
			_ = writer.WriteByte('"')
		default:
			_ = writer.WriteByte(v)
		}
	}

	_ = writer.WriteByte('"')
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
