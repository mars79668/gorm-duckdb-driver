package duckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	duckdb_driver "github.com/marcboeker/go-duckdb/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

func (dialector Dialector) Initialize(db *gorm.DB) error {
	if dialector.DriverName == "" {
		dialector.DriverName = "duckdb-gorm" // Use our custom driver wrapper
	}

	var err error
	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else {
		db.ConnPool, err = sql.Open(dialector.DriverName, dialector.DSN)
		if err != nil {
			return err
		}
	}

	if dialector.DefaultStringSize == 0 {
		dialector.DefaultStringSize = 256
	}

	// Register callbacks
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{
		CreateClauses: []string{"INSERT", "VALUES", "ON CONFLICT"}, // Remove RETURNING
		UpdateClauses: []string{"UPDATE", "SET", "WHERE"},          // Remove RETURNING
		DeleteClauses: []string{"DELETE", "FROM", "WHERE"},         // Remove RETURNING
		QueryClauses:  []string{"SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT", "FOR"},
	})

	// Register custom callback for time pointer conversion
	err = db.Callback().Create().Before("gorm:create").Register("duckdb:convert_time_pointers", convertTimePointersCallback)
	if err != nil {
		return err
	}
	err = db.Callback().Update().Before("gorm:update").Register("duckdb:convert_time_pointers", convertTimePointersCallback)
	if err != nil {
		return err
	}
	err = db.Callback().Delete().Before("gorm:delete").Register("duckdb:convert_time_pointers", convertTimePointersCallback)
	if err != nil {
		return err
	}
	err = db.Callback().Query().Before("gorm:query").Register("duckdb:convert_time_pointers", convertTimePointersCallback)
	if err != nil {
		return err
	}

	// Wrap the connection pool to handle time pointer conversion and db.DB() access
	if sqlDB, ok := db.ConnPool.(*sql.DB); ok {
		wrapper := &duckdbConnPoolWrapper{
			ConnPool: sqlDB,
			db:       sqlDB,
		}
		db.ConnPool = wrapper
	}

	return nil
}

// Custom driver wrapper to handle time pointer conversion at the lowest level
type duckdbDriverWrapper struct {
	driver.Driver
}

type duckdbConnWrapper struct {
	driver.Conn
}

// Register our custom driver wrapper
func init() {
	// Get the original DuckDB driver
	originalDriver := &duckdb_driver.Driver{}

	// Register our wrapper with a custom name
	sql.Register("duckdb-gorm", &duckdbDriverWrapper{Driver: originalDriver})
}

func (d *duckdbDriverWrapper) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &duckdbConnWrapper{Conn: conn}, nil
}

// convertDriverArgs converts *time.Time to time.Time in driver arguments
func convertDriverArgs(args []driver.NamedValue) []driver.NamedValue {
	converted := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		converted[i] = arg
		if timePtr, ok := arg.Value.(*time.Time); ok && timePtr != nil {
			converted[i].Value = *timePtr
		}
	}
	return converted
}

func (c *duckdbConnWrapper) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	convertedArgs := convertDriverArgs(args)

	if execer, ok := c.Conn.(driver.ExecerContext); ok {
		return execer.ExecContext(ctx, query, convertedArgs)
	}
	return nil, fmt.Errorf("driver does not support ExecContext")
}

func (c *duckdbConnWrapper) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	convertedArgs := convertDriverArgs(args)

	if queryer, ok := c.Conn.(driver.QueryerContext); ok {
		return queryer.QueryContext(ctx, query, convertedArgs)
	}
	return nil, fmt.Errorf("driver does not support QueryContext")
}

// Ensure the wrapper implements all necessary interfaces
func (c *duckdbConnWrapper) Prepare(query string) (driver.Stmt, error) {
	return c.Conn.Prepare(query)
}

func (c *duckdbConnWrapper) Close() error {
	return c.Conn.Close()
}

func (c *duckdbConnWrapper) Begin() (driver.Tx, error) {
	// Use BeginTx with default options if available
	if connBeginTx, ok := c.Conn.(driver.ConnBeginTx); ok {
		return connBeginTx.BeginTx(context.Background(), driver.TxOptions{})
	}
	// Fallback to deprecated Begin for compatibility
	return c.Conn.Begin() //nolint:staticcheck
}

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
			if dialector.DefaultStringSize > 0 && dialector.DefaultStringSize <= 65535 {
				size = int(dialector.DefaultStringSize)
			} else {
				size = 256 // Safe default
			}
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
			switch v := field.DefaultValueInterface.(type) {
			case bool:
				if v {
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
	_ = writer.WriteByte('?')
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
				_, _ = writer.WriteString(`""`)
				continuousBacktick = 0
			}
		case '.':
			if continuousBacktick > 0 || !selfQuoted {
				shiftDelimiter = 0
				underQuoted = false
				continuousBacktick = 0
				_ = writer.WriteByte('"')
			}
			_ = writer.WriteByte(v)
			continue
		default:
			if shiftDelimiter-continuousBacktick <= 0 && !underQuoted {
				_ = writer.WriteByte('"')
				underQuoted = true
				if selfQuoted = continuousBacktick > 0; selfQuoted {
					continuousBacktick -= 1
				}
			}

			for ; continuousBacktick > 0; continuousBacktick -= 1 {
				_, _ = writer.WriteString(`""`)
			}

			_ = writer.WriteByte(v)
		}
		shiftDelimiter++
	}

	if continuousBacktick > 0 && !selfQuoted {
		_, _ = writer.WriteString(`""`)
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

// convertTimePointersCallback is a GORM callback that converts *time.Time to time.Time in statement vars
func convertTimePointersCallback(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Vars == nil {
		return
	}

	// Convert any *time.Time values to time.Time
	for i, v := range db.Statement.Vars {
		if v == nil {
			continue
		}
		if timePtr, ok := v.(*time.Time); ok {
			if timePtr == nil {
				db.Statement.Vars[i] = nil
			} else {
				db.Statement.Vars[i] = *timePtr
			}
		}
	}
}

// duckdbConnPoolWrapper wraps the connection pool to handle time pointer conversion and provide db.DB() access
type duckdbConnPoolWrapper struct {
	gorm.ConnPool
	db *sql.DB // Store reference to underlying *sql.DB
}

func (p *duckdbConnPoolWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.ConnPool.PrepareContext(ctx, query)
}

// ExecContext delegates to the underlying connection pool with time pointer conversion
func (p *duckdbConnPoolWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	convertedArgs := convertTimePointers(args)
	return p.ConnPool.ExecContext(ctx, query, convertedArgs...)
}

// QueryContext delegates to the underlying connection pool with time pointer conversion
func (p *duckdbConnPoolWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	convertedArgs := convertTimePointers(args)
	return p.ConnPool.QueryContext(ctx, query, convertedArgs...)
}

// QueryRowContext delegates to the underlying connection pool with time pointer conversion
func (p *duckdbConnPoolWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	convertedArgs := convertTimePointers(args)
	return p.ConnPool.QueryRowContext(ctx, query, convertedArgs...)
}

// BeginTx implements gorm.TxBeginner to support transactions
func (p *duckdbConnPoolWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if p.db != nil {
		tx, err := p.db.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return tx, nil
	}
	return nil, fmt.Errorf("no database connection available")
}

// Implement GORM's GetDBConnector interface
func (p *duckdbConnPoolWrapper) GetDBConn() (*sql.DB, error) {
	// Return the stored reference to the original *sql.DB
	if p.db != nil {
		return p.db, nil
	}
	return nil, fmt.Errorf("no database connection available")
}
