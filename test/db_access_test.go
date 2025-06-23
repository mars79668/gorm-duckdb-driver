package main

import (
	"fmt"
	"log"

	"github.com/greysquirr3l/gorm-duckdb-driver"
	"gorm.io/driver/duckdb"
	"gorm.io/gorm"
)

func main() {
	// Test the db.DB() method access
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test accessing the underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get *sql.DB:", err)
	}

	// Test that we can use the sql.DB methods
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Get stats to verify it works
	stats := sqlDB.Stats()
	fmt.Printf("✅ Successfully accessed *sql.DB!\n")
	fmt.Printf("   - Max open connections: %d\n", stats.MaxOpenConnections)
	fmt.Printf("   - Open connections: %d\n", stats.OpenConnections)
	fmt.Printf("   - In use: %d\n", stats.InUse)
	fmt.Printf("   - Idle: %d\n", stats.Idle)

	// Test that we can still use GORM normally
	type User struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	user := User{Name: "Test User"}
	if err := db.Create(&user).Error; err != nil {
		log.Fatal("Failed to create user:", err)
	}

	fmt.Printf("✅ GORM operations work correctly!\n")
	fmt.Printf("   - Created user with ID: %d\n", user.ID)

	fmt.Println("✅ All tests passed! The db.DB() method works correctly.")
}
