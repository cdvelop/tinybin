# TinyBin Refactoring Prompt: Breaking Change to Constructor Pattern

## Overview
Refactor TinyBin from current global variable-based architecture to a constructor pattern with isolated instances. This is a **breaking change** - no backward compatibility required.

## Current Architecture Issues
- Global variables for schemas cache (`sync.Map`)
- Global pools for encoders and decoders
- Shared state across all operations
- Testing contamination between tests
- Difficult integration with multiple protocols
- No configuration options
- Side effects between operations

## Target Architecture: Constructor Pattern

### Core Requirements
```go
// Minimal public API - ONLY these methods should be public
type TinyBin struct {
    log func(msg ...any) // Optional custom logging function
}

// Variadic constructor with optional arguments
func New(args ...any) *TinyBin
func (tb *TinyBin) Encode(data any) ([]byte, error)
func (tb *TinyBin) Decode(data []byte, target any) error

// Everything else MUST be private
```

## Technical Specifications

### 1. Schema Cache Implementation (TinyGo Compatible)
**CRITICAL**: Use slice-based cache instead of `sync.Map` for TinyGo compatibility.

```go
type TinyBin struct {
    schemas []schemaEntry // Slice-based cache (NO maps)
    // ... other private fields
}

type schemaEntry struct {
    TypeID uint32 // From tinyreflect.StructID()
    Codec  Codec
}
```

**Cache Operations** (all private methods):
- `findSchema(typeID uint32) (Codec, bool)` - Linear search in slice
- `addSchema(typeID uint32, codec Codec)` - Append to slice
- `cleanupCache()` - Remove expired entries based on TTL

### 2. Object Pooling (Private)
```go
type TinyBin struct {
    encoders *sync.Pool // Private pool
    decoders *sync.Pool // Private pool
}
```

## Breaking Changes Required

### 1. Remove All Global Variables
- ❌ Remove global `schemas sync.Map`
- ❌ Remove global encoder/decoder pools
- ❌ Remove all global functions (`Encode()`, `Decode()`, etc.)

### 2. Replace Global Functions with Instance Methods
```go
// OLD (to be removed)
func Encode(data any) ([]byte, error)
func Decode(data []byte, target any) error

// NEW (instance methods only)
func (tb *TinyBin) Encode(data any) ([]byte, error)
func (tb *TinyBin) Decode(data []byte, target any) error
```

### 3. Remove Public Access to Internal Structures
- ❌ No public access to cache
- ❌ No public access to pools
- ❌ No public access to metrics (unless through getter method)
- ❌ No public access to internal types

## Implementation Approach

### Basic Constructor Pattern
```go
func New(args ...any) *TinyBin {
    var logFunc func(msg ...any)

    // Check if logging function is provided
    if len(args) > 0 {
        if log, ok := args[0].(func(msg ...any)); ok {
            logFunc = log
        }
    }

    // Default: no logging
    if logFunc == nil {
        logFunc = func(msg ...any) {} // No-op logger
    }

    return &TinyBin{
        log:      logFunc,
        schemas:  make([]schemaEntry, 0, 100), // Pre-allocate reasonable size
        encoders: &sync.Pool{New: func() any { return &encoder{} }},
        decoders: &sync.Pool{New: func() any { return &decoder{} }},
    }
}
```

### Usage Examples
```go
// Simple instantiation (no logging)
tb := tinybin.New()
data, err := tb.Encode(myStruct)

// With custom logging (useful for testing)
tb := tinybin.New(func(msg ...any) {
    t.Logf("TinyBin: %v", msg)
})
```

## Integration Examples

### Basic Usage
```go
// Simple instantiation (no logging)
tb := tinybin.New()
data, err := tb.Encode(myStruct)

// With custom logging (useful for debugging/tests)
tb := tinybin.New(func(msg ...any) {
    fmt.Printf("TinyBin: %v\n", msg)
})
```

### Testing (Isolated)
```go
func TestMyFunction(t *testing.T) {
    // With test logging
    tb := tinybin.New(func(msg ...any) {
        t.Logf("TinyBin: %v", msg)
    })

    // Completely isolated test
    result, err := tb.Encode(testData)
    assert.NoError(t, err)
}
```

### Multiple Protocol Integration
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

## TinyGo Compatibility Requirements

### Critical Constraints
1. **No maps allowed** - Use slices for all caching
2. **Limited memory** - Implement cache size limits
3. **No reflection abuse** - Use tinyreflect efficiently

### Slice-Based Cache Implementation
```go
// Linear search is acceptable for < 1000 types
func (tb *TinyBin) findSchema(typeID uint32) (Codec, bool) {
    for _, entry := range tb.schemas {
        if entry.TypeID == typeID {
            return entry.Codec, true
        }
    }
    return nil, false
}

func (tb *TinyBin) addSchema(typeID uint32, codec Codec) {
    // Simple cache size limit (optional, for memory control)
    if len(tb.schemas) >= 1000 { // Reasonable default limit
        // Simple eviction: remove oldest (first) entry
        tb.schemas = tb.schemas[1:]
    }

    tb.schemas = append(tb.schemas, schemaEntry{
        TypeID: typeID,
        Codec:  codec,
    })
}
```

## Migration Strategy

### For Existing Users
1. **Update all imports** to use instance methods
2. **Replace global calls** with constructor instantiation
3. **Configure as needed** for specific use cases

### Example Migration
```go
// OLD
import "github.com/tinylib/tinybin"
data, err := tinybin.Encode(myData)

// NEW
import "github.com/tinylib/tinybin"
tb := tinybin.New()
data, err := tb.Encode(myData)
```

## Success Criteria

### Functional Requirements
- ✅ All existing functionality preserved through instance methods
- ✅ Complete isolation between instances
- ✅ TinyGo compatibility maintained
- ✅ Performance characteristics preserved
- ✅ Memory usage controlled
- ✅ Optional custom logging support

### Non-Functional Requirements
- ✅ Clean, minimal public API
- ✅ Comprehensive test coverage with isolated instances
- ✅ Clear documentation and examples
- ✅ No global state pollution
- ✅ Thread-safe operation within instances

## Testing Strategy

### Unit Tests
- Each test gets fresh instance
- No test interference
- Custom logging for test output
- Error condition handling

### Integration Tests
- Multiple instance coordination
- Memory usage validation
- Performance benchmarking
- TinyGo compilation verification

### Benchmark Tests
- Encoding/decoding performance
- Memory allocation tracking
- Concurrent access patterns

## Deliverables

1. **Refactored TinyBin struct** with minimal public API
2. **Constructor function** with configuration support
3. **Instance methods** for Encode/Decode operations
4. **Slice-based cache** implementation for TinyGo compatibility
5. **Comprehensive tests** demonstrating isolated instances
6. **Migration examples** for existing users
7. **Performance benchmarks** validating no regression

## Constraints and Limitations

- **Breaking change** - no backward compatibility
- **Minimal public API** - only essential methods exposed
- **TinyGo compatible** - no map usage, efficient memory usage
- **Instance isolation** - no shared global state
- **Simple logging** - optional custom logging function support

This refactoring transforms TinyBin from a convenient but limited global library into a professional, enterprise-ready solution with proper isolation and testing capabilities.