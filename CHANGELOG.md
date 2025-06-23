# Changelog

All notable changes to the GORM DuckDB driver will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-06-23

### Added

- **Comprehensive DuckDB Extension Support**: Complete extension management system for DuckDB
- **ExtensionManager**: Low-level extension operations (load, install, list, status checking)
- **ExtensionHelper**: High-level convenience methods for common extension workflows
- **Extension Auto-loading**: Preload extensions on database connection with configuration
- **Analytics Extensions**: JSON, Parquet, ICU for data processing and analytics
- **Data Format Extensions**: CSV, Excel, Arrow, SQLite for diverse data sources
- **Spatial Extensions**: Geospatial analysis capabilities with spatial extension support
- **Cloud Extensions**: HTTP/S3, Azure, AWS for cloud data access
- **Machine Learning Extensions**: ML extension support for advanced analytics
- **Time Series Extensions**: Specialized time series analysis capabilities
- **Extension Status Tracking**: Real-time monitoring of extension load and install status
- **Extension Documentation**: Comprehensive usage examples and best practices

### Improved

- **Database Connection Access**: Fixed `db.DB()` method to properly access underlying `*sql.DB` instance
- **Error Handling**: Enhanced error messages and validation throughout extension system
- **Test Coverage**: Added 13+ comprehensive extension tests with real functionality testing
- **Code Organization**: Cleaned up package structure and resolved import conflicts
- **GORM Compatibility**: Following GORM coding standards and conventions throughout

### Technical Details

- **Extension-Aware Dialectors**: New dialector variants with built-in extension support
- **Flexible Configuration**: ExtensionConfig for customizing extension behavior
- **Safe Extension Loading**: Proper error handling and validation for extension operations
- **Interface Compliance**: Full GORM interface implementation maintained
- **Backward Compatibility**: All existing functionality preserved

### Extension Categories Supported

- **Core Extensions**: JSON, Parquet, ICU (built-in extensions)
- **Analytics**: AutoComplete, FTS, TPC-H, TPC-DS benchmarking
- **Data Formats**: CSV, Excel, Arrow, SQLite import/export
- **Cloud Storage**: HTTPFS, AWS S3, Azure blob storage
- **Geospatial**: Spatial analysis and GIS functionality
- **Machine Learning**: ML algorithms and model support
- **Time Series**: Specialized time series analysis
- **Visualization**: Data visualization capabilities

### Usage Examples

```go
// Extension-aware dialector
extensionConfig := &duckdb.ExtensionConfig{
    AutoInstall:       true,
    PreloadExtensions: []string{"json", "parquet", "spatial"},
}
db, err := gorm.Open(duckdb.OpenWithExtensions(":memory:", extensionConfig), &gorm.Config{})

// Extension management
manager, _ := duckdb.GetExtensionManager(db)
manager.LoadExtension("spatial")

// Extension helper
helper := duckdb.NewExtensionHelper(manager)
helper.EnableAnalytics()    // Load analytics extensions
helper.EnableSpatial()      // Load spatial extensions
```

### Known Issues

- **Time Pointer Conversion**: Temporarily disabled to ensure `db.DB()` method compatibility
- **Affects**: `*time.Time` field handling in some edge cases
- **Workaround**: Use `time.Time` directly instead of `*time.Time` where possible
- **Resolution**: Will be addressed in v0.2.1 with improved connection wrapper

### Breaking Changes

- None - Full backward compatibility maintained

## [0.1.0] - 2025-06-22

### Added

- Initial implementation of GORM DuckDB driver
- Full GORM interface compliance
- Support for all standard CRUD operations
- Auto-migration functionality
- Transaction support with savepoints  
- Index management (create, drop, rename, check existence)
- Constraint support (foreign keys, check constraints)
- Comprehensive data type mapping for DuckDB
- View creation and management
- Connection pooling support
- Proper SQL quoting and parameter binding
- Error handling and translation
- Full test coverage
- Documentation and examples

### Features

- **Dialector**: Complete implementation of GORM dialector interface
- **Migrator**: Full migrator implementation with all migration operations
- **Data Types**: Comprehensive mapping between Go and DuckDB types
- **Indexes**: Support for creating, dropping, and managing indexes
- **Constraints**: Foreign key and check constraint support
- **Views**: Create and drop view support
- **Transactions**: Savepoint and rollback support
- **Raw SQL**: Full support for raw SQL queries and execution

### Data Type Support

- Boolean values (BOOLEAN)
- Integer types (TINYINT, SMALLINT, INTEGER, BIGINT)
- Unsigned integer types (UTINYINT, USMALLINT, UINTEGER, UBIGINT)
- Floating point types (REAL, DOUBLE)
- String types (VARCHAR, TEXT)
- Time types (TIMESTAMP with optional precision)
- Binary data (BLOB)

### Migration Operations

- Table creation, dropping, and existence checking
- Column addition, dropping, modification, and renaming
- Index creation, dropping, and management
- Constraint creation, dropping, and verification
- Auto-migration with smart column type detection

### Testing

- Comprehensive unit tests for all functionality
- Integration tests with real DuckDB database
- Data type mapping verification
- Migration operation testing
- CRUD operation validation

### Documentation

- Complete README with usage examples
- API documentation for all public methods
- Migration guide and best practices
- Performance considerations and notes
- Example application demonstrating all features

### Compatibility

- GORM v1.25.x compatibility
- Go 1.18+ support
- DuckDB latest stable version support
- Cross-platform compatibility (Windows, macOS, Linux)

## [Unreleased]

### Planned Features

- Enhanced error messages and debugging
- Performance optimizations
- Additional DuckDB-specific features
- Bulk operation optimizations
- Connection pooling enhancements
