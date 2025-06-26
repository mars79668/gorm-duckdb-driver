# GORM DuckDB Driver

A comprehensive, production-ready DuckDB driver for [GORM](https://gorm.io), bringing high-performance analytical database capabilities to the Go ecosystem with full ORM support.

[![Go Reference](https://pkg.go.dev/badge/gorm.io/driver/duckdb.svg)](https://pkg.go.dev/gorm.io/driver/duckdb)
[![Go Report Card](https://goreportcard.com/badge/gorm.io/driver/duckdb)](https://goreportcard.com/report/gorm.io/driver/duckdb)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🎯 Why DuckDB + GORM?

- **High-Performance Analytics**: DuckDB's columnar storage and vectorized execution
- **OLAP Workloads**: Perfect for data science, analytics, and reporting
- **Full ORM Support**: All GORM features work seamlessly with DuckDB
- **Array Support**: First-class support for array types with type safety
- **Zero Dependencies**: Embedded database, no external server required

## ✨ Features

- ✅ **Complete GORM Compatibility** - All dialector and migrator interfaces
- ✅ **Array Support** - Native support for `TEXT[]`, `BIGINT[]`, `DOUBLE[]` with type safety
- ✅ **Auto-Migration** - Full schema introspection and migration support
- ✅ **Transactions** - Complete transaction support with savepoints
- ✅ **Connection Pooling** - Optimized connection handling with value conversion
- ✅ **Type Safety** - Comprehensive Go ↔ DuckDB type mapping
- ✅ **Extension Support** - DuckDB extension management system
- ✅ **Time Handling** - Robust time and timestamp support
- ✅ **Index Management** - Full index creation and management
- ✅ **Constraint Support** - Foreign keys, unique constraints, etc.

## 🚀 Quick Start

### Install

```bash
go get -u gorm.io/driver/duckdb
go get -u gorm.io/gorm
```

### Connect to Database

```go
import (
  duckdb "gorm.io/driver/duckdb"
  "gorm.io/gorm"
)

// In-memory database (perfect for testing)
db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})

// File-based database
db, err := gorm.Open(duckdb.Open("analytics.db"), &gorm.Config{})

// With custom configuration
db, err := gorm.Open(duckdb.New(duckdb.Config{
  DSN: "analytics.db",
  DefaultStringSize: 256,
}), &gorm.Config{})
```

## 🎨 Array Support (New!)

DuckDB's powerful array types are now fully supported with type safety:

```go
import duckdb "gorm.io/driver/duckdb"

type Product struct {
  ID         uint                `gorm:"primaryKey"`
  Name       string              `gorm:"size:100;not null"`
  Categories duckdb.StringArray  `json:"categories"`  // TEXT[]
  Scores     duckdb.FloatArray   `json:"scores"`      // DOUBLE[]
  ViewCounts duckdb.IntArray     `json:"view_counts"` // BIGINT[]
}

// Create with arrays
product := Product{
  ID:         1,
  Name:       "Analytics Software",
  Categories: duckdb.StringArray{"software", "analytics", "business"},
  Scores:     duckdb.FloatArray{4.5, 4.8, 4.2, 4.9},
  ViewCounts: duckdb.IntArray{1250, 890, 2340, 567},
}

db.Create(&product)

// Query with array functions
var products []Product
db.Where("array_length(categories) > ?", 2).Find(&products)
```

### Array Types

| Go Type | DuckDB Type | Description |
|---------|-------------|-------------|
| `duckdb.StringArray` | `TEXT[]` | Array of strings |
| `duckdb.IntArray` | `BIGINT[]` | Array of integers |
| `duckdb.FloatArray` | `DOUBLE[]` | Array of floats |

## 📊 Data Type Mapping

| Go Type | DuckDB Type | Notes |
|---------|-------------|-------|
| `bool` | `BOOLEAN` | |
| `int8` | `TINYINT` | |
| `int16` | `SMALLINT` | |
| `int32` | `INTEGER` | |
| `int64` | `BIGINT` | |
| `uint8` | `TINYINT` | Mapped to signed for FK compatibility |
| `uint16` | `SMALLINT` | Mapped to signed for FK compatibility |
| `uint32` | `INTEGER` | Mapped to signed for FK compatibility |
| `uint64` | `BIGINT` | Mapped to signed for FK compatibility |
| `float32` | `REAL` | |
| `float64` | `DOUBLE` | |
| `string` | `VARCHAR(n)` / `TEXT` | |
| `time.Time` | `TIMESTAMP` | |
| `[]byte` | `BLOB` | |
| `duckdb.StringArray` | `TEXT[]` | **New!** |
| `duckdb.IntArray` | `BIGINT[]` | **New!** |
| `duckdb.FloatArray` | `DOUBLE[]` | **New!** |

## 💡 Usage Examples

### Define Models

```go
type User struct {
  ID        uint               `gorm:"primaryKey" json:"id"`
  Name      string             `gorm:"size:100;not null" json:"name"`
  Email     string             `gorm:"size:255;uniqueIndex" json:"email"`
  Age       uint8              `json:"age"`
  Birthday  time.Time          `json:"birthday"`
  CreatedAt time.Time          `gorm:"autoCreateTime:false" json:"created_at"`
  UpdatedAt time.Time          `gorm:"autoUpdateTime:false" json:"updated_at"`
  Tags      duckdb.StringArray `json:"tags"` // Array support!
}

type Post struct {
  ID       uint   `gorm:"primaryKey"`
  Title    string `gorm:"size:200;not null"`
  Content  string `gorm:"type:text"`
  UserID   uint
  User     User   `gorm:"foreignKey:UserID"`
}
```

### Auto Migration

```go
db.AutoMigrate(&User{}, &Post{})
```

### CRUD Operations

```go
// Create with arrays
user := User{
  ID:   1,
  Name: "Alice Johnson", 
  Email: "alice@example.com",
  Age:  28,
  Tags: duckdb.StringArray{"developer", "golang", "analytics"},
  CreatedAt: time.Now(),
  UpdatedAt: time.Now(),
}
db.Create(&user)

// Read
var user User
db.First(&user, 1)
fmt.Printf("User tags: %v\n", []string(user.Tags))

// Update arrays
user.Tags = append(user.Tags, "expert")
db.Save(&user)

// Query with array conditions
var developers []User
db.Where("array_to_string(tags, ',') LIKE ?", "%developer%").Find(&developers)
```

### Relationships

```go
// One-to-many with preloading
var userWithPosts User
db.Preload("Posts").First(&userWithPosts)

// Many-to-many
type Tag struct {
  ID    uint   `gorm:"primaryKey"`
  Name  string `gorm:"uniqueIndex"`
  Posts []Post `gorm:"many2many:post_tags;"`
}

var tag Tag
db.Model(&tag).Association("Posts").Append(&post)
```

### Transactions

```go
err := db.Transaction(func(tx *gorm.DB) error {
  // Create user
  if err := tx.Create(&user).Error; err != nil {
    return err
  }
  
  // Create posts
  for _, post := range posts {
    post.UserID = user.ID
    if err := tx.Create(&post).Error; err != nil {
      return err
    }
  }
  
  return nil
})
```

### Analytics Queries

```go
// Analytical aggregations
type Result struct {
  AgeGroup string
  Count    int64
  AvgAge   float64
}

var results []Result
db.Model(&User{}).
  Select("CASE WHEN age < 30 THEN 'Young' ELSE 'Mature' END as age_group, COUNT(*) as count, AVG(age) as avg_age").
  Group("age_group").
  Scan(&results)

// Array aggregations
var categoryStats []struct {
  Category string
  Count    int64
}
db.Raw(`
  SELECT UNNEST(categories) as category, COUNT(*) as count 
  FROM products 
  GROUP BY category 
  ORDER BY count DESC
`).Scan(&categoryStats)
```

## 🔧 Extension Support

DuckDB extensions can be managed programmatically:

```go
// Get extension manager
extManager := duckdb.GetExtensionManager(db)

// Load extensions
extManager.LoadExtension("json")
extManager.LoadExtension("parquet")

// Check if extension is loaded
if extManager.IsExtensionLoaded("json") {
  // Use JSON functions
  db.Raw("SELECT json_extract(data, '$.name') FROM documents").Scan(&names)
}

// Helper functions for common extension sets
duckdb.EnableAnalyticsExtensions(db)    // spatial, stats, etc.
duckdb.EnableDataFormatExtensions(db)   // parquet, csv, json, etc.
```

## 🎯 Perfect For

- **Data Analytics & OLAP**: High-performance analytical queries
- **Data Science**: Perfect for ML pipelines and data exploration  
- **ETL Processes**: Fast data transformation and loading
- **Reporting Dashboards**: Real-time analytics with complex aggregations
- **Time Series Analysis**: Efficient temporal data processing
- **Embedded Analytics**: No external database server required

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Your Go App   │───▶│   GORM Driver    │───▶│     DuckDB      │
│                 │    │                  │    │   (Embedded)    │
│  - Models       │    │  - Dialector     │    │  - Columnar     │
│  - Queries      │    │  - Migrator      │    │  - Vectorized   │
│  - Arrays       │    │  - Type Mapping  │    │  - Analytics    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 📈 Performance

DuckDB excels at analytical workloads:

- **Columnar Storage**: Optimal for analytical queries
- **Vectorized Execution**: SIMD-optimized query processing  
- **Parallel Processing**: Multi-threaded query execution
- **Advanced Optimizations**: Cost-based query optimizer
- **Compression**: Efficient data storage and transfer

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
git clone https://github.com/greysquirr3l/gorm-duckdb-driver.git
cd gorm-duckdb-driver
go mod tidy
go test -v
```

### Running Examples

```bash
cd example
go run main.go
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [GORM](https://gorm.io) - The fantastic ORM library for Go
- [DuckDB](https://duckdb.org) - High-performance analytical database
- [go-duckdb](https://github.com/marcboeker/go-duckdb) - Go bindings for DuckDB

## 📞 Support

- 🐛 **Issues**: [GitHub Issues](https://github.com/greysquirr3l/gorm-duckdb-driver/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/greysquirr3l/gorm-duckdb-driver/discussions)
- 📖 **Documentation**: [pkg.go.dev](https://pkg.go.dev/gorm.io/driver/duckdb)

---

**Made with ❤️ for the Go and DuckDB communities**
