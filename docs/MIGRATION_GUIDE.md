# TinyBin Migration Guide: From Global Functions to Constructor Pattern

## Overview

TinyBin has been refactored from a global variable-based architecture to a constructor pattern with isolated instances. This is a **breaking change** that provides better isolation, testing capabilities, and configuration options.

## What Changed

### Before (Global Architecture)
```go
// Global functions - shared state across all operations
data, err := tinybin.Encode(myStruct)
err = tinybin.Decode(data, &myStruct)

// Global encoder/decoder pools
// Global schema cache (sync.Map)
```

### After (Constructor Pattern)
```go
// Instance-based - isolated state per instance
tb := tinybin.New()
data, err := tb.Encode(myStruct)
err = tb.Decode(data, &myStruct)

// Private encoder/decoder pools per instance
// Slice-based schema cache per instance (TinyGo compatible)
```

## Migration Steps

### 1. Update Imports
No changes needed to imports - the package name remains the same.

### 2. Replace Global Function Calls

#### Simple Encoding/Decoding
```go
// OLD
data, err := tinybin.Encode(myData)

// NEW
tb := tinybin.New()
data, err := tb.Encode(myData)
```

#### Encoding to Specific Writer
```go
// OLD
var buffer bytes.Buffer
err := tinybin.EncodeTo(myData, &buffer)

// NEW
tb := tinybin.New()
var buffer bytes.Buffer
err := tb.EncodeTo(myData, &buffer)
```

#### Decoding from Byte Slice
```go
// OLD
err := tinybin.Decode(data, &target)

// NEW
tb := tinybin.New()
err := tb.Decode(data, &target)
```

### 3. Update Multiple Instance Usage

#### Before (Problematic - Shared State)
```go
// These would interfere with each other
go func() {
    data, _ := tinybin.Encode(struct1)
    // ...
}()
go func() {
    data, _ := tinybin.Encode(struct2) // Could interfere!
    // ...
}()
```

#### After (Isolated Instances)
```go
// Each goroutine gets its own instance
go func() {
    tb := tinybin.New()
    data, _ := tb.Encode(struct1)
    // ...
}()
go func() {
    tb := tinybin.New()
    data, _ := tb.Encode(struct2) // Completely isolated!
    // ...
}()
```

### 4. Custom Logging (Optional)

#### No Logging (Default)
```go
tb := tinybin.New()
```

#### With Custom Logging
```go
tb := tinybin.New(func(msg ...any) {
    log.Printf("TinyBin: %v", msg)
})
```

#### With Test Logging
```go
func TestMyFunction(t *testing.T) {
    tb := tinybin.New(func(msg ...any) {
        t.Logf("TinyBin: %v", msg)
    })

    // Completely isolated test
    result, err := tb.Encode(testData)
    assert.NoError(t, err)
}
```

## Benefits of the New Architecture

### 1. Complete Isolation
- No shared global state between instances
- No test contamination
- Safe for concurrent use with multiple instances

### 2. Better Testing
- Each test can have its own TinyBin instance
- No interference between tests
- Custom logging per test instance

### 3. Multiple Protocol Support
```go
type ProtocolManager struct {
    httpTinyBin  *tinybin.TinyBin
    grpcTinyBin  *tinybin.TinyBin
    kafkaTinyBin *tinybin.TinyBin
}

func NewProtocolManager() *ProtocolManager {
    return &ProtocolManager{
        httpTinyBin:  tinybin.New(), // No logging for production
        grpcTinyBin:  tinybin.New(),
        kafkaTinyBin: tinybin.New(),
    }
}
```

### 4. TinyGo Compatibility
- Slice-based caching instead of sync.Map
- No global variables
- Memory-efficient for embedded targets

### 5. Configuration Options
- Optional custom logging function
- Future: cache size limits, TTL settings, etc.

## Common Patterns

### Basic Usage
```go
// Simple instantiation (no logging)
tb := tinybin.New()
data, err := tb.Encode(myStruct)
if err != nil {
    return err
}
err = tb.Decode(data, &myStruct)
```

### With Error Handling
```go
tb := tinybin.New()
data, err := tb.Encode(myStruct)
if err != nil {
    return fmt.Errorf("encode failed: %w", err)
}

var result MyStruct
err = tb.Decode(data, &result)
if err != nil {
    return fmt.Errorf("decode failed: %w", err)
}
```

### In Structs/Services
```go
type MyService struct {
    tinyBin *tinybin.TinyBin
}

func NewMyService() *MyService {
    return &MyService{
        tinyBin: tinybin.New(),
    }
}

func (s *MyService) Process(data MyStruct) ([]byte, error) {
    return s.tinyBin.Encode(data)
}
```

## Troubleshooting

### "undefined: tinybin.Encode" Error
This means you're calling the old global function. Replace:
```go
// OLD
tinybin.Encode(data)

// NEW
tb := tinybin.New()
tb.Encode(data)
```

### Performance Concerns
The new architecture has the same performance characteristics as the old one:
- Same encoding/decoding speed
- Similar memory usage (slice-based cache vs sync.Map)
- Same object pooling benefits

### Large Number of Types
The slice-based cache has a default limit of 1000 entries. For applications with many types:
- The cache uses LRU-style eviction (removes oldest entries)
- Consider creating multiple TinyBin instances for different type domains

## Migration Checklist

- [ ] Replace all `tinybin.Encode()` calls with `tinybin.New().Encode()`
- [ ] Replace all `tinybin.Decode()` calls with `tinybin.New().Decode()`
- [ ] Replace all `tinybin.EncodeTo()` calls with `tinybin.New().EncodeTo()`
- [ ] Update tests to use instance-based pattern
- [ ] Add custom logging where beneficial
- [ ] Test concurrent usage with multiple instances
- [ ] Verify no performance regression

## Questions?

If you encounter issues during migration:
1. Check this guide for the appropriate pattern
2. Review the usage examples in the main documentation
3. Consider the benefits of instance isolation for your use case

The new architecture provides better encapsulation and testing capabilities while maintaining the same performance characteristics as the previous version.