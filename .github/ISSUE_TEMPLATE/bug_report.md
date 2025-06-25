---
name: Bug Report
about: Create a detailed bug report to help us improve the DuckDB GORM driver
title: '[BUG] Brief description of the issue'
labels: 'bug'
assignees: ''
---

# DuckDB GORM Driver Bug Report

**Report Date:** <!-- Current date -->  
**Project:** <!-- Your project name -->  
**Driver Version:** `github.com/greysquirr3l/gorm-duckdb-driver vX.X.X`  
**Issue Type:** <!-- Critical/High/Medium/Low - Application Crash/Data Loss/Performance/Enhancement -->  
**Severity:** <!-- Blocker/Critical/Major/Minor -->  

## Summary

<!-- Provide a clear and concise description of the bug -->

## Environment

### Software Versions

- **Go Version:** <!-- e.g., 1.21.5 -->
- **GORM Version:** <!-- e.g., v1.25.12 -->
- **DuckDB Driver:** `github.com/greysquirr3l/gorm-duckdb-driver vX.X.X`
- **DuckDB Bindings:** <!-- e.g., github.com/marcboeker/go-duckdb/v2 -->
- **Operating System:** <!-- e.g., macOS 14.2, Ubuntu 22.04, Windows 11 -->

### Dependencies

```go
// Include relevant go.mod entries
require (
    gorm.io/gorm vX.X.X
    gorm.io/driver/duckdb vX.X.X
)

// Include any replace directives
replace gorm.io/driver/duckdb => github.com/greysquirr3l/gorm-duckdb-driver vX.X.X
```

## Error Details

### Stack Trace

```plaintext
<!-- Paste the complete stack trace here -->
```

### Error Location

```go
// File: path/to/file.go:line
// Include the relevant code snippet where the error occurs
```

## Root Cause Analysis

### Primary Issue

<!-- Describe what you believe is causing the issue -->

### Technical Details

1. **Description of the problem:**
   - <!-- Step-by-step breakdown of what happens -->

2. **Driver Failure Point:**
   - <!-- Where specifically the driver fails -->

3. **Missing/Broken Implementation:**
   - <!-- What functionality is missing or broken -->

## Models/Schemas Being Used

### Model Definition

```go
type YourModel struct {
    // Include the complete model definition that causes the issue
}
```

## Impact Assessment

### Severity: **[SEVERITY LEVEL]**

- **Application Startup:** ✅/❌ <!-- Works/Fails -->
- **Database Operations:** ✅/❌ <!-- Works/Fails -->
- **Production Deployment:** ✅/❌ <!-- Possible/Impossible -->
- **Development Testing:** ✅/❌ <!-- Works/Blocked -->

### Business Impact

- <!-- Describe the business impact of this issue -->

## Reproduction Steps

1. **Setup:**

   ```bash
   # Include setup commands
   ```

2. **Run:**

   ```bash
   # Include commands to reproduce
   ```

3. **Observe:**
   - <!-- What happens when you run the above -->

4. **Minimal Reproduction Case:**

   ```go
   package main
   
   import (
       "gorm.io/gorm"
       duckdb "gorm.io/driver/duckdb"
   )
   
   func main() {
       // Minimal code that reproduces the issue
   }
   ```

## Expected Behavior

<!-- Describe what you expected to happen -->

### Required Functionality

- **Feature 1:** <!-- What should work -->
- **Feature 2:** <!-- What should work -->

## Current Behavior

<!-- Describe what actually happens -->

## Workaround Attempts

### 1. **Attempted Solution 1** ✅/❌

```go
// Code for workaround
```

**Result:** <!-- What happened -->

### 2. **Attempted Solution 2** ✅/❌

```go
// Code for workaround
```

**Result:** <!-- What happened -->

## Proposed Solutions

### 1. **[Solution Name]**

**Priority:** <!-- HIGH/MEDIUM/LOW -->  
**Effort:** <!-- High/Medium/Low -->

```go
// Proposed code changes or approach
```

**Description:** <!-- Explain the solution -->

### 2. **Alternative Approach**

**Priority:** <!-- HIGH/MEDIUM/LOW -->  
**Effort:** <!-- High/Medium/Low -->

<!-- Describe alternative solution -->

## Additional Context

### Configuration

```go
// Include any relevant configuration
db, err := gorm.Open(duckdb.Open("database.db"), &gorm.Config{
    // Your config
})
```

### Extensions Used

<!-- List any DuckDB extensions that are loaded -->
- Extension 1
- Extension 2

### Performance Context

<!-- If performance-related -->
- **Data Size:** <!-- Amount of data involved -->
- **Query Complexity:** <!-- Simple/Complex queries -->
- **Concurrent Connections:** <!-- Number of connections -->

## Testing Information

### Test Case

```go
func TestReproduceBug(t *testing.T) {
    // Test case that reproduces the bug
}
```

### Expected Test Result

<!-- What the test should do when the bug is fixed -->

## Screenshots/Logs

<!-- Include any relevant screenshots or additional log output -->

## Checklist

- [ ] I have searched existing issues for similar problems
- [ ] I have provided a minimal reproduction case
- [ ] I have included all relevant version information
- [ ] I have described the expected behavior
- [ ] I have included stack traces and error messages
- [ ] I have attempted basic troubleshooting steps

## Additional Information

<!-- Any other information that might be relevant -->

---

**Reporter:** <!-- Your name/handle -->  
**Contact:** <!-- Email or preferred contact method -->  
**Priority for Your Project:** <!-- High/Medium/Low -->  
**Willing to Contribute Fix:** <!-- Yes/No/Maybe -->
