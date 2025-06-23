# Changelog

All notable changes to the GORM DuckDB driver will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] - 2025-06-22

### Fixed

- **Code Quality**: Reduced cyclomatic complexity in key functions (`DataTypeOf`, `DefaultValueOf`, `QuoteTo`, `FullDataTypeOf`)
- **Error Handling**: Fixed all unchecked error returns from writer operations
- **Type Safety**: Resolved integer overflow concerns in size conversions
- **Formatting**: Applied consistent code formatting with gofumpt
- **Linting**: Addressed all golangci-lint issues including errcheck, gocritic, gosec, and staticcheck warnings

### Changed

- Refactored `DataTypeOf` function to use helper methods for better maintainability
- Simplified `QuoteTo` function logic while maintaining functionality
- Improved error handling throughout codebase with proper return value checking
- Enhanced type switch optimization for better performance
- Removed redundant embedded field selectors in migrator

### Technical Details

This release focuses on code quality improvements and linting compliance:

- Split complex functions into smaller, more manageable helper functions
- Added proper error handling for all writer operations
- Implemented type assertions and error checking best practices
- Reduced cyclomatic complexity from 22+ to under 10 for all functions
- Ensured all golangci-lint checks pass with zero issues

## [0.1.1] - 2025-06-22

### Fixed

- **Critical**: Fixed `db.DB()` method access issue by implementing `GetDBConnector()` interface in connection wrapper
- Resolved "sql: unknown driver duckdb" error by adding proper DuckDB driver import
- Cleaned up package conflicts and removed large binary files from repository
- Updated `.gitignore` to prevent future binary commits

### Changed

- Improved connection pool wrapper to properly expose underlying `*sql.DB` instance
- Enhanced example application with DB access testing
- Updated import paths in example to use correct module reference

### Technical Details

The `duckdbConnPoolWrapper` now properly implements the interface needed for GORM to access the underlying `*sql.DB` through the `db.DB()` method. This enables:

- Connection pool configuration (`SetMaxIdleConns`, `SetMaxOpenConns`)
- Database monitoring (`db.DB().Stats()`)
- Health checks (`db.DB().Ping()`)
- All other standard `*sql.DB` operations

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
