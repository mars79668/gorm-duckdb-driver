
package main

import (
"fmt"
"log"
"gorm.io/driver/duckdb"
"gorm.io/gorm"
)

func main() {
db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})
if err != nil {
log.Fatal("Failed to connect:", err)
}

sqlDB, err := db.DB()
if err != nil {
log.Fatal("Failed to get *sql.DB:", err)
}

if err := sqlDB.Ping(); err != nil {
log.Fatal("Failed to ping:", err)
}

fmt.Println("✅ db.DB() method works correctly!")
}
