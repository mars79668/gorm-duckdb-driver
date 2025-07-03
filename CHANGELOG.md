# Changelog

All notable changes to the GORM DuckDB driver will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.4] - 2025-06-26

### 📚 Documentation Enhancements

This release focuses on improving user experience with comprehensive installation guidance and enhanced documentation.

### ✨ Added

- **📋 Enhanced Installation Instructions**: Complete step-by-step installation guide with proper `go.mod` setup
- **🔗 Replace Directive Documentation**: Detailed explanation of the required `replace` directive for module path compatibility
- **📝 Installation Examples**: Real-world examples showing correct `go.mod` configuration
- **🚀 Quick Start Improvements**: Streamlined getting-started experience with clear dependency management

### 📖 Improved

- **README.md Structure**: Better organization with clear sections for installation, usage, and migration
- **Module Path Clarity**: Comprehensive explanation of why the replace directive is necessary
- **Version Reference**: Updated all documentation to reference v0.2.4
- **User Guidance**: Added notes about seamless migration to official GORM driver once available

### 🔧 Technical Details

#### Replace Directive Implementation

```go
// Required in go.mod for proper functionality
replace gorm.io/driver/duckdb => github.com/greysquirr3l/gorm-duckdb-driver v0.2.4
```

#### Documentation Structure

- **Installation Guide**: Step-by-step process with dependency management
- **Module Configuration**: Clear examples of proper `go.mod` setup
- **Migration Path**: Explanation of future transition to official GORM driver
- **Compatibility Notes**: Version compatibility and upgrade guidance

### 🎯 User Experience Improvements

- **Clearer Setup Process**: Reduced confusion around module installation
- **Better Onboarding**: New users can get started faster with improved documentation
- **Version Consistency**: All examples and references updated to v0.2.4
- **Future Compatibility**: Documentation prepared for eventual official GORM integration

### 🔄 Breaking Changes

None. This release is fully backward compatible with v0.2.3.

### 🐛 Fixed

- **Documentation Gaps**: Filled missing information about proper installation process
- **Module Path Confusion**: Clarified the relationship between hosted location and module path
- **Installation Examples**: Corrected and enhanced code examples for better clarity

---

## [0.2.3] - 2025-06-26

### 🎉 Major Feature: Production-Ready Array Support

This release brings **first-class array support** to the GORM DuckDB driver, making it the first GORM driver with native, type-safe array functionality.

### ✨ Added

- **🎨 Array Types**: Native support for `StringArray`, `IntArray`, `FloatArray` with full type safety
- **🔄 Valuer/Scanner Interface**: Proper `driver.Valuer` and `sql.Scanner` implementation for seamless database integration
- **🏗️ GORM Integration**: Custom types implement `GormDataType()` interface for automatic schema generation
- **📊 Schema Migration**: Automatic DDL generation for `TEXT[]`, `BIGINT[]`, `DOUBLE[]` column types
- **🧪 Comprehensive Testing**: Full test suite covering array creation, updates, edge cases, and error handling
- **📚 Documentation**: Complete array usage examples and best practices

### 🔧 Technical Implementation

#### Array Type System

```go
type StringArray []string  // Maps to TEXT[]
type IntArray []int64      // Maps to BIGINT[]  
type FloatArray []float64  // Maps to DOUBLE[]
```

#### GORM Integration

- Automatic schema migration with proper array column types
- Full CRUD support (Create, Read, Update, Delete) for array fields
- Type-safe operations with compile-time checking
- Seamless marshaling/unmarshaling between Go and DuckDB array syntax

#### Database Features

- **Array Literals**: Automatic conversion to DuckDB format `['a', 'b', 'c']`
- **Null Handling**: Proper nil array support
- **Empty Arrays**: Correct handling of zero-length arrays
- **String Escaping**: Safe handling of special characters in string arrays
- **Query Support**: Compatible with DuckDB array functions and operators

### 🎯 Usage Examples

#### Model 

```go
type Product struct {
    ID         uint                `gorm:"primaryKey"`
    Name       string              `gorm:"size:100;not null"`
    Categories duckdb.StringArray  `json:"categories"`
    Scores     duckdb.FloatArray   `json:"scores"`
    ViewCounts duckdb.IntArray     `json:"view_counts"`
}
```

#### Array Operations

```go
// Create with arrays
product := Product{
    Categories: duckdb.StringArray{"software", "analytics"},
    Scores:     duckdb.FloatArray{4.5, 4.8, 4.2},
    ViewCounts: duckdb.IntArray{1250, 890, 2340},
}
db.Create(&product)

// Update arrays
product.Categories = append(product.Categories, "premium")
db.Save(&product)

// Query with array functions
db.Where("array_length(categories) > ?", 2).Find(&products)
```

### 🏆 Key Benefits

- **Type Safety**: Compile-time checking prevents array type mismatches
- **Performance**: Native DuckDB array support for optimal query performance  
- **Simplicity**: Natural Go slice syntax with automatic database conversion
- **Compatibility**: Full integration with existing GORM patterns and workflows
- **Robustness**: Comprehensive error handling and edge case support

### 🔄 Breaking Changes

None. This release is fully backward compatible.

### 🐛 Fixed

- **Schema Migration**: Arrays now properly migrate with correct DDL syntax
- **Type Recognition**: GORM correctly identifies and handles custom array types
- **Value Conversion**: Seamless conversion between Go slices and DuckDB array literals

### 🧪 Testing

- ✅ **Array CRUD Operations**: Full create, read, update, delete testing
- ✅ **Type Safety**: Compile-time and runtime type checking
- ✅ **Edge Cases**: Nil arrays, empty arrays, special characters
- ✅ **Integration**: End-to-end testing with real DuckDB operations
- ✅ **Performance**: Benchmark testing for array operations

### 📊 Impact

This release positions the GORM DuckDB driver as the **most advanced GORM driver** with unique array capabilities perfect for:

- **Analytics Workloads**: Store and query multi-dimensional data efficiently
- **Data Science**: Handle complex datasets with array-based features
- **Modern Applications**: Leverage DuckDB's advanced array functionality through GORM's familiar ORM interface

---

## [0.2.2] - 2025-06-25

### Fixed

- **Time Pointer Conversion**: Completely resolved the critical "*time.Time to time.Time" cast error that occurred in transaction contexts
- **Transaction Support**: Fixed time pointer conversion for all operations executed within GORM transactions
- **Universal Compatibility**: Implemented comprehensive driver-level wrapper ensuring time pointer conversion works in all contexts
- **RETURNING Clause**: Removed problematic RETURNING clauses from default callbacks to eliminate transaction bypass issues

#### Technical Implementation

- **Driver-Level Wrapper**: Registered custom "duckdb-gorm" driver that intercepts all database operations at the lowest level
- **Dual-Layer Protection**: Combined connection wrapper and driver wrapper ensure time pointer conversion works universally
- **Transaction Compatibility**: Driver wrapper handles time pointer conversion even when GORM uses raw `*sql.Tx` objects
- **Backward Compatibility**: All existing functionality preserved while fixing the core time pointer issue

#### Impact

- ✅ **All CRUD operations** now work seamlessly with `*time.Time` fields
- ✅ **Transaction operations** properly handle time pointer conversion
- ✅ **Full GORM compatibility** maintained for all standard operations
- ✅ **Production ready** - can serve as drop-in replacement for official GORM driver

### Technical Details

The driver now includes a comprehensive wrapper system that ensures time pointer conversion happens at the most fundamental level:

```go
// Custom driver registration
sql.Register("duckdb-gorm", &convertingDriver{&duckdb.Driver{}})

// Automatic time pointer conversion
func convertDriverValues(args []driver.Value) []driver.Value {
    for i, arg := range args {
        if timePtr, ok := arg.(*time.Time); ok {
            if timePtr == nil {
                converted[i] = nil
            } else {
                converted[i] = *timePtr
            }
        }
    }
    return converted
}
```

This ensures that **all** database operations, including those within transactions, properly handle `*time.Time` to `time.Time` conversion without any manual intervention required.

## [0.2.1] - 2025-06-24

### Added

- **Extension Support**: Comprehensive DuckDB extension management system
- **Extension Manager**: Programmatic loading and management of DuckDB extensions
- **Helper Functions**: Convenience functions for common extension sets (analytics, data formats, spatial)
- **Extension Validation**: Automatic validation of extension availability and loading status

### Fixed

- **Core Functionality**: Resolved fundamental GORM compatibility issues
- **Data Type Mapping**: Improved type mapping for better DuckDB compatibility
- **Schema Migration**: Enhanced auto-migration with better error handling
- **Connection Handling**: More robust connection management and pooling

### Enhanced

- **Test Coverage**: Comprehensive test suite for all functionality
- **Documentation**: Improved examples and usage documentation
- **Error Handling**: Better error messages and debugging information

## [0.2.0] - 2025-06-23

### Added

- **Full GORM Compatibility**: Complete implementation of GORM dialector interface
- **Auto-Migration**: Full schema migration support with table, column, and index management
- **Transaction Support**: Complete transaction support with savepoints
- **Connection Pooling**: Optimized connection handling
- **Type Safety**: Comprehensive Go ↔ DuckDB type mapping

### Initial Features

- **CRUD Operations**: Full Create, Read, Update, Delete support
- **Relationships**: Foreign keys and associations
- **Indexes**: Index creation and management
- **Constraints**: Primary keys, unique constraints, foreign keys
- **Schema Introspection**: Complete database schema discovery

## [0.1.0] - 2025-06-22

### Added

- **Initial Release**: Basic DuckDB driver for GORM
- **Core Functionality**: Basic database operations
- **Foundation**: Solid foundation for GORM DuckDB integration

---

**Legend:**

- 🎉 Major Feature
- ✨ Added
- 🔧 Technical
- 🔄 Changed  
- 🐛 Fixed
- 🏆 Achievement
- 📊 Impact
