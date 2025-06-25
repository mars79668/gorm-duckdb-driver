package duckdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArrayLiteral_Simple(t *testing.T) {
	db := setupTestDB(t)

	// Test very simple array insertion using raw SQL
	err := db.Exec("CREATE TABLE test_arrays (id INTEGER, floats FLOAT[], strings VARCHAR[])").Error
	require.NoError(t, err)

	// Test array literal conversion
	floatArray := []float64{1.1, 2.2, 3.3}
	stringArray := []string{"hello", "world"}

	literal1 := ArrayLiteral{Data: floatArray}
	val1, err := literal1.Value()
	require.NoError(t, err)
	t.Logf("Float array literal: %s", val1)

	literal2 := ArrayLiteral{Data: stringArray}
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
	floatScanner := &SimpleArrayScanner{Target: &floats}
	stringScanner := &SimpleArrayScanner{Target: &strings}

	err = db.Raw("SELECT id, floats, strings FROM test_arrays WHERE id = ?", 1).Row().Scan(&id, floatScanner, stringScanner)
	require.NoError(t, err)

	t.Logf("Retrieved: id=%d, floats=%v, strings=%v", id, floats, strings)
}
