# Changelog

All notable changes to the GORM DuckDB driver will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
