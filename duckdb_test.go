package duckdb

import (
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:255;uniqueIndex"`
	Age       uint8
	Birthday  time.Time `gorm:"autoCreateTime:false"` // Change from *time.Time to time.Time
	CreatedAt time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
}

func TestDialector(t *testing.T) {
	// Test creating a dialector with DSN
	dialector := Open(":memory:")
	if dialector.Name() != "duckdb" {
		t.Errorf("Expected dialector name to be 'duckdb', got %s", dialector.Name())
	}
}

func TestConnection(t *testing.T) {
	// Test connecting to DuckDB
	db, err := gorm.Open(Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Test auto migration
	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	// Test creating a record with explicit timestamps
	now := time.Now()
	user := User{
		ID:        1, // Set ID manually since we don't have autoIncrement
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Birthday:  time.Time{}, // Use zero time instead of nil
		CreatedAt: now,         // Use time.Time directly
		UpdatedAt: now,         // Use time.Time directly
	}

	result := db.Create(&user)
	if result.Error != nil {
		t.Fatalf("Failed to create user: %v", result.Error)
	}

	// Test querying by ID since we now have a proper primary key
	var retrievedUser User
	result = db.First(&retrievedUser, 1)
	if result.Error != nil {
		t.Fatalf("Failed to retrieve user: %v", result.Error)
	}

	if retrievedUser.Name != "John Doe" {
		t.Errorf("Expected name to be 'John Doe', got %s", retrievedUser.Name)
	}

	// Test updating
	result = db.Model(&retrievedUser).Update("name", "Jane Doe")
	if result.Error != nil {
		t.Fatalf("Failed to update user: %v", result.Error)
	}

	// Test deleting
	result = db.Delete(&retrievedUser)
	if result.Error != nil {
		t.Fatalf("Failed to delete user: %v", result.Error)
	}
}

func TestDataTypes(t *testing.T) {
	db, err := gorm.Open(Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	type TestModel struct {
		ID          uint `gorm:"primarykey"`
		BoolField   bool
		IntField    int
		Int8Field   int8
		Int16Field  int16
		Int32Field  int32
		Int64Field  int64
		UintField   uint
		Uint8Field  uint8
		Uint16Field uint16
		Uint32Field uint32
		Uint64Field uint64
		Float32     float32
		Float64     float64
		StringField string `gorm:"size:255"`
		TextField   string `gorm:"type:text"`
		TimeField   time.Time
		BytesField  []byte
	}

	err = db.AutoMigrate(&TestModel{})
	if err != nil {
		t.Fatalf("Failed to auto migrate test model: %v", err)
	}

	// Test creating record with various data types
	now := time.Now()
	testData := TestModel{
		ID:          1, // Set explicit ID to avoid auto-increment issues in test
		BoolField:   true,
		IntField:    123,
		Int8Field:   12,
		Int16Field:  1234,
		Int32Field:  123456,
		Int64Field:  1234567890,
		UintField:   456,
		Uint8Field:  45,
		Uint16Field: 4567,
		Uint32Field: 456789,
		Uint64Field: 4567890123,
		Float32:     123.45,
		Float64:     123.456789,
		StringField: "test string",
		TextField:   "long text field content",
		TimeField:   now,
		BytesField:  []byte("binary data"),
	}

	result := db.Create(&testData)
	if result.Error != nil {
		t.Fatalf("Failed to create test data: %v", result.Error)
	}

	// Verify the data was stored correctly by querying with the known ID
	var retrieved TestModel
	result = db.First(&retrieved, 1) // Use the explicit ID we set
	if result.Error != nil {
		t.Fatalf("Failed to retrieve test data: %v", result.Error)
	}

	// Verify field values
	if retrieved.ID != testData.ID {
		t.Errorf("ID field mismatch: expected %d, got %d", testData.ID, retrieved.ID)
	}

	if retrieved.BoolField != testData.BoolField {
		t.Errorf("Bool field mismatch: expected %v, got %v", testData.BoolField, retrieved.BoolField)
	}

	if retrieved.StringField != testData.StringField {
		t.Errorf("String field mismatch: expected %s, got %s", testData.StringField, retrieved.StringField)
	}

	if retrieved.IntField != testData.IntField {
		t.Errorf("Int field mismatch: expected %d, got %d", testData.IntField, retrieved.IntField)
	}

	if retrieved.Float32 != testData.Float32 {
		t.Errorf("Float32 field mismatch: expected %f, got %f", testData.Float32, retrieved.Float32)
	}

	if string(retrieved.BytesField) != string(testData.BytesField) {
		t.Errorf("Bytes field mismatch: expected %s, got %s", string(testData.BytesField), string(retrieved.BytesField))
	}
}

func TestMigration(t *testing.T) {
	db, err := gorm.Open(Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Test table creation
	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	// Test if table exists
	if !db.Migrator().HasTable(&User{}) {
		t.Error("Expected table to exist after migration")
	}

	// Test adding column
	type UserWithExtra struct {
		User
		Extra string
	}

	err = db.AutoMigrate(&UserWithExtra{})
	if err != nil {
		t.Fatalf("Failed to migrate with extra column: %v", err)
	}

	// Test if column exists
	if !db.Migrator().HasColumn(&UserWithExtra{}, "extra") {
		t.Error("Expected extra column to exist after migration")
	}
}

func TestDBMethod(t *testing.T) {
	// Test that db.DB() method works correctly
	db, err := gorm.Open(Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Check if ConnPool is set
	if db.ConnPool == nil {
		t.Fatal("db.ConnPool is nil")
	}

	// Test getting the underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		t.Logf("ConnPool type: %T", db.ConnPool)
		t.Fatalf("Failed to get *sql.DB: %v", err)
	}

	if sqlDB == nil {
		t.Fatal("db.DB() returned nil - this should not happen")
	}

	// Test ping
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Test setting connection pool settings
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	// Test getting stats
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 10 {
		t.Errorf("Expected MaxOpenConnections to be 10, got %d", stats.MaxOpenConnections)
	}

	// Test close (this should work for cleanup)
	defer func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()
}
