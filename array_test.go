package duckdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type TestProductWithArrays struct {
	ID         uint        `gorm:"primaryKey"`
	Name       string      `gorm:"size:100"`
	Categories StringArray `json:"categories"`
	Scores     FloatArray  `json:"scores"`
	ViewCounts IntArray    `json:"view_counts"`
	CreatedAt  time.Time   `gorm:"autoCreateTime:false"`
}

func TestArraySupport(t *testing.T) {
	db, err := gorm.Open(Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate
	err = db.AutoMigrate(&TestProductWithArrays{})
	assert.NoError(t, err)

	now := time.Now()

	// Test creating with arrays
	product := TestProductWithArrays{
		ID:         1,
		Name:       "Test Product",
		Categories: StringArray{"electronics", "computers", "laptops"},
		Scores:     FloatArray{4.5, 4.8, 4.2},
		ViewCounts: IntArray{100, 250, 75},
		CreatedAt:  now,
	}

	result := db.Create(&product)
	assert.NoError(t, result.Error)

	// Test retrieving with arrays
	var retrieved TestProductWithArrays
	result = db.First(&retrieved, 1)
	assert.NoError(t, result.Error)

	assert.Equal(t, "Test Product", retrieved.Name)
	assert.Equal(t, []string{"electronics", "computers", "laptops"}, []string(retrieved.Categories))
	assert.Equal(t, []float64{4.5, 4.8, 4.2}, []float64(retrieved.Scores))
	assert.Equal(t, []int64{100, 250, 75}, []int64(retrieved.ViewCounts))

	// Test updating arrays
	retrieved.Categories = StringArray{"electronics", "computers", "laptops", "gaming"}
	retrieved.Scores = append(retrieved.Scores, 4.9)
	retrieved.ViewCounts = append(retrieved.ViewCounts, 300)

	result = db.Save(&retrieved)
	assert.NoError(t, result.Error)

	// Verify updates
	var updated TestProductWithArrays
	result = db.First(&updated, 1)
	assert.NoError(t, result.Error)

	assert.Equal(t, 4, len(updated.Categories))
	assert.Equal(t, 4, len(updated.Scores))
	assert.Equal(t, 4, len(updated.ViewCounts))
	assert.Equal(t, "gaming", updated.Categories[3])
	assert.Equal(t, 4.9, updated.Scores[3])
	assert.Equal(t, int64(300), updated.ViewCounts[3])
}

func TestArrayValuerScanner(t *testing.T) {
	// Test StringArray
	t.Run("StringArray", func(t *testing.T) {
		arr := StringArray{"hello", "world", "test"}

		// Test Value()
		val, err := arr.Value()
		assert.NoError(t, err)
		assert.Equal(t, "['hello', 'world', 'test']", val)

		// Test Scan()
		var scanned StringArray
		err = scanned.Scan("['foo', 'bar', 'baz']")
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo", "bar", "baz"}, []string(scanned))

		// Test empty array
		err = scanned.Scan("[]")
		assert.NoError(t, err)
		assert.Equal(t, 0, len(scanned))

		// Test nil
		err = scanned.Scan(nil)
		assert.NoError(t, err)
		assert.Nil(t, scanned)
	})

	// Test IntArray
	t.Run("IntArray", func(t *testing.T) {
		arr := IntArray{1, 2, 3, 42}

		// Test Value()
		val, err := arr.Value()
		assert.NoError(t, err)
		assert.Equal(t, "[1, 2, 3, 42]", val)

		// Test Scan()
		var scanned IntArray
		err = scanned.Scan("[10, 20, 30]")
		assert.NoError(t, err)
		assert.Equal(t, []int64{10, 20, 30}, []int64(scanned))
	})

	// Test FloatArray
	t.Run("FloatArray", func(t *testing.T) {
		arr := FloatArray{1.5, 2.7, 3.14}

		// Test Value()
		val, err := arr.Value()
		assert.NoError(t, err)
		assert.Equal(t, "[1.5, 2.7, 3.14]", val)

		// Test Scan()
		var scanned FloatArray
		err = scanned.Scan("[4.5, 6.7, 8.9]")
		assert.NoError(t, err)
		assert.Equal(t, []float64{4.5, 6.7, 8.9}, []float64(scanned))
	})
}

func TestArrayEdgeCases(t *testing.T) {
	// Test empty arrays
	t.Run("EmptyArrays", func(t *testing.T) {
		emptyStr := StringArray{}
		val, err := emptyStr.Value()
		assert.NoError(t, err)
		assert.Equal(t, "[]", val)

		emptyInt := IntArray{}
		val, err = emptyInt.Value()
		assert.NoError(t, err)
		assert.Equal(t, "[]", val)

		emptyFloat := FloatArray{}
		val, err = emptyFloat.Value()
		assert.NoError(t, err)
		assert.Equal(t, "[]", val)
	})

	// Test nil arrays
	t.Run("NilArrays", func(t *testing.T) {
		var nilStr StringArray
		val, err := nilStr.Value()
		assert.NoError(t, err)
		assert.Nil(t, val)

		var nilInt IntArray
		val, err = nilInt.Value()
		assert.NoError(t, err)
		assert.Nil(t, val)

		var nilFloat FloatArray
		val, err = nilFloat.Value()
		assert.NoError(t, err)
		assert.Nil(t, val)
	})

	// Test string escaping
	t.Run("StringEscaping", func(t *testing.T) {
		arr := StringArray{"hello's", "world\"test", "normal"}
		val, err := arr.Value()
		assert.NoError(t, err)
		assert.Equal(t, "['hello''s', 'world\"test', 'normal']", val)
	})
}
