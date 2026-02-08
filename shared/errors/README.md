# FeCIM Error Handling Package

This package provides user-friendly error handling utilities for the FeCIM Lattice Tools project.

## Features

### User-Friendly Errors (`errors.go`)

The `FeCIMError` type provides:

- **Categorized errors**: Classify errors as Config, IO, Data, Resource, Internal, or User errors
- **User-friendly messages**: Clear, actionable error messages
- **Recovery suggestions**: Tell users how to fix the problem
- **Technical details**: Include debugging info without cluttering the user message
- **Error wrapping**: Preserve error chains with `errors.Is` and `errors.As` support

```go
import "fecim-lattice-tools/shared/errors"

// Create a configuration error with recovery hint
err := errors.ConfigNotFound("physics")
// Output: Configuration 'physics' not found
// Recovery: Create the configuration file or check the path. Expected locations: config/physics.yaml

// Create a file error with cause
err := errors.FileNotFound("/path/to/file").WithCause(originalErr)

// Create a user input error with details
err := errors.InvalidParameter("level", -5, "positive integer (1-100)")
```

### Recovery Utilities (`recovery.go`)

Safe execution patterns for graceful degradation:

```go
// Run a function with panic recovery
err := errors.RecoverFunc(func() error {
    // code that might panic
    return nil
})

// Graceful degradation with fallbacks
result, err := (&errors.GracefulDegradation[int]{
    Primary: func() (int, error) { return loadFromNetwork() },
    Fallbacks: []func() (int, error){
        func() (int, error) { return loadFromCache() },
        func() (int, error) { return loadFromDefaults() },
    },
    Default: 0,
}).Execute()

// Collect errors from batch operations
collector := errors.NewErrorCollector()
for _, item := range items {
    collector.AddFromRecover("process", func() {
        processItem(item)
    })
}
if collector.HasErrors() {
    return collector.Combined()
}
```

## Error Categories

| Category | Description | Recoverable |
|----------|-------------|-------------|
| CategoryConfig | Configuration/setup errors | Yes (with hint) |
| CategoryIO | File/network I/O errors | Yes (with hint) |
| CategoryData | Invalid or corrupt data | Yes (with hint) |
| CategoryResource | Resource exhaustion (memory, GPU) | Usually no |
| CategoryInternal | Programming errors, bugs | No |
| CategoryUser | User input validation errors | Yes (with hint) |

## Best Practices

1. **Use specific constructors** instead of generic `NewError()`:
   ```go
   // Good
   return errors.FileNotFound("/path/to/file")
   
   // Less informative
   return errors.NewError(errors.CategoryIO, "file not found")
   ```

2. **Always include recovery hints** for user-facing errors:
   ```go
   return errors.ConfigError("missing API key").
       WithRecovery("Set FECIM_API_KEY environment variable")
   ```

3. **Wrap errors to preserve context**:
   ```go
   if err := loadFile(path); err != nil {
       return errors.Wrap(err, "failed to load simulation")
   }
   ```

4. **Use graceful degradation** for non-critical features:
   ```go
   cfg := physics.LoadWithDefaults() // Never fails
   ```

5. **Use SafeGo for goroutines** that should not crash the app:
   ```go
   utils.SafeGo("animation", func() {
       runAnimation() // Panics are caught and logged
   })
   ```

## Migration from log.Fatal/panic

### Before
```go
cfg, err := loadConfig()
if err != nil {
    log.Fatalf("failed to load config: %v", err)
}
```

### After
```go
cfg, err := loadConfig()
if err != nil {
    return errors.ConfigError("failed to load configuration").
        WithCause(err).
        WithRecovery("Check config/physics.yaml exists and is valid")
}
```

### Panics That Are OK

Some panics are acceptable for programming errors that should never happen:
- Struct alignment checks (GPU shader compatibility)
- Type assertions that are guaranteed by the code structure
- Initialization invariants

For these, use clear error messages explaining what went wrong:
```go
if unsafe.Sizeof(Params{}) != 32 {
    panic("Params struct size mismatch: GPU shaders require 32 bytes")
}
```

## Testing

```bash
go test ./shared/errors/... -v
```
