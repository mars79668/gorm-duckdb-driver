# GORM RowQuery Callback Workaround

This document explains the GORM RowQuery callback workaround implemented in this DuckDB driver.

## Background

**GORM Bug**: GORM versions up to v1.30.2 have a critical bug in the `RowQuery` callback that causes `Raw().Row()` to return `nil` instead of a valid `*sql.Row`. This leads to nil pointer panics when trying to scan results.

**Impact**: Any code using `Raw().Row()` will crash with a nil pointer dereference:

```go
// This crashes with GORM v1.30.2 and earlier
row := db.Raw("SELECT 1").Row()  // Returns nil instead of *sql.Row
err := row.Scan(&result)         // Panic: nil pointer dereference
```

## The Workaround

This driver automatically applies a workaround that replaces GORM's broken `RowQuery` callback with a working implementation.

### Default Behavior (Recommended)

```go
// Workaround is automatically enabled
db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{})

// This now works correctly:
row := db.Raw("SELECT 1").Row()
var result int
err := row.Scan(&result) // ✅ Works!
```

### Explicit Control

You can explicitly control the workaround:

```go
// Explicitly enable workaround (current GORM versions)
db := gorm.Open(duckdb.OpenWithRowCallbackWorkaround(dsn, true), &gorm.Config{})

// Explicitly disable workaround (future GORM versions)
db := gorm.Open(duckdb.OpenWithRowCallbackWorkaround(dsn, false), &gorm.Config{})
```

### Advanced Configuration

```go
config := &duckdb.Config{
    RowCallbackWorkaround: &[]bool{true}[0], // Enable workaround
    DefaultStringSize:     512,
}
db := gorm.Open(duckdb.OpenWithConfig(dsn, config), &gorm.Config{})
```

## When to Disable the Workaround

**Current Status**: Keep the workaround enabled for all GORM versions up to v1.30.2.

**Future**: When GORM releases a version that fixes the `RowQuery` callback bug (likely v1.31+), you can:

1. **Test without workaround**:

   ```go
   db := gorm.Open(duckdb.OpenWithRowCallbackWorkaround(dsn, false), &gorm.Config{})
   
   // Test if Raw().Row() works
   row := db.Raw("SELECT 1").Row()
   if row != nil {
       // GORM bug is fixed! You can disable the workaround
   }
   ```

2. **Update your code**:

   ```go
   // Remove workaround when GORM bug is fixed
   db := gorm.Open(duckdb.OpenWithRowCallbackWorkaround(dsn, false), &gorm.Config{})
   ```

## Compatibility

### What Works

✅ **Single row queries**: `Raw().Row().Scan()`  
✅ **Multiple row queries**: `Raw().Rows()`  
✅ **Direct scanning**: `Raw().Scan()`  
✅ **Error handling**: Proper error propagation  
✅ **Parameter binding**: `Raw("SELECT ?", value)`  

### What's Different

🔄 **Callback replacement**: We replace GORM's `gorm:row` callback  
🔄 **Logging**: Additional debug logs show callback registration  

### Unaffected Features

The workaround only affects `Raw().Row()` calls. All other GORM functionality works normally:

- Model operations (`Create`, `Find`, `Update`, `Delete`)
- Query builder (`Where`, `Select`, `Join`, etc.)
- Migrations and schema operations
- Transactions and connections

## Technical Details

### How It Works

1. **Detection**: We detect if the workaround should be applied
2. **Replacement**: Replace `gorm:row` callback with our implementation
3. **Execution**: Our callback properly calls `QueryRowContext()`
4. **Assignment**: Result is correctly assigned to `db.Statement.Dest`

### The Fix

```go
func rowQueryCallback(db *gorm.DB) {
    // Skip if error, no SQL, or dry run
    if db.Error != nil || db.Statement.SQL.Len() == 0 || db.DryRun {
        return
    }

    // Handle both single row and multiple rows
    if isRows, ok := db.Get("rows"); ok && isRows.(bool) {
        // Multiple rows
        db.Statement.Dest, db.Error = db.Statement.ConnPool.QueryContext(...)
    } else {
        // Single row - this is what GORM's callback fails to do correctly
        db.Statement.Dest = db.Statement.ConnPool.QueryRowContext(...)
    }
}
```

## Migration Guide

### From Other Dialectors

If you're migrating from other GORM dialectors, your `Raw().Row()` code will start working automatically:

```go
// This might have worked in other dialectors
row := db.Raw("SELECT COUNT(*) FROM users").Row()
var count int
err := row.Scan(&count) // Now works with DuckDB too!
```

### To Future GORM Versions

When GORM fixes the bug:

```go
// Step 1: Test if workaround is still needed
db := gorm.Open(duckdb.OpenWithRowCallbackWorkaround(dsn, false), &gorm.Config{})
if db.Raw("SELECT 1").Row() != nil {
    // Step 2: Bug is fixed, update all your code to disable workaround
    // Use OpenWithRowCallbackWorkaround(dsn, false) everywhere
}
```

## Troubleshooting

### Issue: Raw().Row() still returns nil

**Cause**: Workaround might not be applied or callback registration failed.

**Solution**: Check logs for callback registration messages:

```text
[DEBUG] Successfully applied RowQuery callback workaround for GORM bug
```

### Issue: Performance concerns

```text
[DEBUG] RowQuery callback executed in 5ms
```

**Impact**: Minimal - the workaround only affects `Raw().Row()` calls, not regular GORM operations.

**Monitoring**: Enable debug logging to monitor callback usage.

### Issue: Compatibility with other extensions

**Solution**: Our callback is designed to be compatible with other GORM extensions. If you encounter issues, disable the workaround temporarily:

```go
db := gorm.Open(duckdb.OpenWithRowCallbackWorkaround(dsn, false), &gorm.Config{})
```

## References

- **Bug Analysis**: `docs/GORM_ROW_CALLBACK_BUG_ANALYSIS.md`
- **GORM Issue #7575**: https://github.com/go-gorm/gorm/issues/7575 (Filed September 2, 2025)
- **Test Examples**: `row_callback_config_test.go`
- **Usage Examples**: `example/row_callback_examples.go`

---

**Note**: This workaround will be removed in a future version once GORM fixes the underlying bug. We'll provide clear migration instructions when that time comes.
