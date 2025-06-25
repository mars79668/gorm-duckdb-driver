package main

import (
	"fmt"
	"log"

	duckdb "gorm.io/driver/duckdb"
	"gorm.io/gorm"
)

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:100;not null"`
	Email string `gorm:"size:255;uniqueIndex"`
	Age   uint8
}

func main() {
	// Open DuckDB connection
	db, err := gorm.Open(duckdb.Open("test_transaction.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	db.AutoMigrate(&User{})

	// Test 1: Check if DuckDB supports transactions at all
	fmt.Println("=== Testing DuckDB Transaction Support ===")

	sqlDB, _ := db.DB()
	tx, err := sqlDB.Begin()
	if err != nil {
		fmt.Printf("❌ DuckDB doesn't support Begin(): %v\n", err)
		return
	}

	fmt.Println("✅ DuckDB supports Begin()")

	err = tx.Commit()
	if err != nil {
		fmt.Printf("❌ DuckDB doesn't support Commit(): %v\n", err)
		return
	}

	fmt.Println("✅ DuckDB supports Commit()")

	// Test 2: Try GORM Transaction
	fmt.Println("\n=== Testing GORM Transaction ===")
	err = db.Transaction(func(tx *gorm.DB) error {
		fmt.Println("📝 Inside transaction...")

		newUser := User{
			Name:  "Transaction Test User",
			Email: "test@transaction.com",
			Age:   30,
		}

		if err := tx.Create(&newUser).Error; err != nil {
			fmt.Printf("❌ Create failed: %v\n", err)
			return err
		}

		fmt.Printf("✅ Created user ID: %d\n", newUser.ID)
		return nil
	})

	if err != nil {
		fmt.Printf("❌ GORM Transaction failed: %v\n", err)
	} else {
		fmt.Println("✅ GORM Transaction succeeded!")
	}

	// Test 3: Manual transaction with raw SQL
	fmt.Println("\n=== Testing Manual Transaction ===")

	tx2, err := sqlDB.Begin()
	if err != nil {
		fmt.Printf("❌ Manual Begin() failed: %v\n", err)
		return
	}

	_, err = tx2.Exec("INSERT INTO users (name, email, age) VALUES (?, ?, ?)", "Manual User", "manual@test.com", 25)
	if err != nil {
		fmt.Printf("❌ Manual Insert failed: %v\n", err)
		tx2.Rollback()
		return
	}

	err = tx2.Commit()
	if err != nil {
		fmt.Printf("❌ Manual Commit failed: %v\n", err)
		return
	}

	fmt.Println("✅ Manual transaction succeeded!")

	// Check results
	var count int64
	db.Model(&User{}).Count(&count)
	fmt.Printf("\nTotal users after tests: %d\n", count)
}
