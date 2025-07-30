package duckdb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	duckdb "github.com/greysquirr3l/gorm-duckdb-driver"
)

func setupSimpleArrayTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func TestArrayLiteral_Simple(t *testing.T) {
	db := setupSimpleArrayTestDB(t)

	// Test very simple array insertion using raw SQL
	err := db.Exec("CREATE TABLE test_arrays (id INTEGER, floats FLOAT[], strings VARCHAR[])").Error
	require.NoError(t, err)

	// Test array literal conversion
	floatArray := []float64{1.1, 2.2, 3.3}
	stringArray := []string{"hello", "world"}

	literal1 := duckdb.ArrayLiteral{Data: floatArray}
	val1, err := literal1.Value()
	require.NoError(t, err)
	t.Logf("Float array literal: %s", val1)

	literal2 := duckdb.ArrayLiteral{Data: stringArray}
	val2, err := literal2.Value()
	require.NoError(t, err)
	t.Logf("String array literal: %s", val2)

	// Test insertion with array literals
	err = db.Exec("INSERT INTO test_arrays (id, floats, strings) VALUES (?, ?, ?)", 1, val1, val2).Error
	require.NoError(t, err)

	// Test retrieval using Raw - scan into proper slice types
	var id int
	var floats []float64
	var strings []string

	// Use SimpleArrayScanner for proper array scanning
	floatScanner := &duckdb.SimpleArrayScanner{Target: &floats}
	stringScanner := &duckdb.SimpleArrayScanner{Target: &strings}

	err = db.Raw("SELECT id, floats, strings FROM test_arrays WHERE id = ?", 1).Row().Scan(&id, floatScanner, stringScanner)
	require.NoError(t, err)

	t.Logf("Retrieved: id=%d, floats=%v, strings=%v", id, floats, strings)
}
