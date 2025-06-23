package duckdb

import (
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID        uint   `gorm:"primarykey;autoIncrement"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:255;uniqueIndex"`
	Age       uint8
	Birthday  *time.Time
	CreatedAt *time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime:false"`
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
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Birthday:  nil,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	result := db.Create(&user)
	if result.Error != nil {
		t.Fatalf("Failed to create user: %v", result.Error)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set after creation")
	}

	// Test querying
	var retrievedUser User
	result = db.First(&retrievedUser, user.ID)
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
	testData := TestModel{
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
		TimeField:   time.Now(),
		BytesField:  []byte("binary data"),
	}

	result := db.Create(&testData)
	if result.Error != nil {
		t.Fatalf("Failed to create test data: %v", result.Error)
	}

	// Verify the data was stored correctly
	var retrieved TestModel
	result = db.First(&retrieved, testData.ID)
	if result.Error != nil {
		t.Fatalf("Failed to retrieve test data: %v", result.Error)
	}

	if retrieved.BoolField != testData.BoolField {
		t.Errorf("Bool field mismatch: expected %v, got %v", testData.BoolField, retrieved.BoolField)
	}

	if retrieved.StringField != testData.StringField {
		t.Errorf("String field mismatch: expected %s, got %s", testData.StringField, retrieved.StringField)
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
