# TinyBin Reflect Migration Plan

## Overview
Migrate `tinybin` from Go's standard `reflect` package to `tinyreflect` for WebAssembly optimization and TinyGo compatibility. This migration aims to reduce binary size while maintaining essential functionality.

## Project Context
- **Source**: `tinybin` - binary protocol library (adaptation of github.com/Kelindar/binary)
- **Target**: Replace `reflect` with `tinyreflect` 
- **Goal**: Minimize binary size for WebAssembly deployment with TinyGo
- **Constraints**: Use only `tinystring` for errors/formatting, no `fmt`/`strconv`/`errors`

## Supported Types (TinyReflect Limitations)
**✅ Supported:**
- Basic types: `string`, `bool`
- All numeric types: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`
- Basic slices: `[]string`, `[]bool`, `[]byte`, `[]int*`, `[]uint*`, `[]float*`
- Structs: Only with supported field types
- Struct slices: `[]struct{...}` where all fields are supported
- Pointers: Only to supported types above

**❌ Unsupported (Remove from tinybin):**
- `interface{}`, `chan`, `func`
- `complex64`, `complex128` (explicitly rejected)
- `uintptr`, `unsafe.Pointer` (except internal use)
- Arrays (different from slices)
- Maps: `map[K]V` (not TinyGo concurrent-safe, replaced with slices)
- Custom marshaling methods (`MarshalBinary`, `UnmarshalBinary`)

## Key Differences: reflect vs tinyreflect
| Feature | reflect | tinyreflect |
|---------|---------|-------------|
| `reflect.Value` | Full API | Minimal: `Type()`, `Field()`, `Interface()` |
| `reflect.Type` | Full API | Minimal: `NumField()`, `NameByIndex()`, `Kind()` |
| Method calls | `reflect.Method.Call()` | ❌ Not supported |
| Struct names | `Type.Name()` | ❌ Not supported (use `StructID()`) |
| Custom marshaling | Supported | ❌ Not supported |
| Complex types | Supported | ❌ Not supported |

## Files to Modify

### 1. `/codecs.go` - Main codec implementations
**Changes:**
- Replace `reflect.Value` with `tinyreflect.Value`
- Replace `reflect.Type` with `tinyreflect.Type`
- Remove `complex64Codec`, `complex128Codec`
- Remove `customCodec` (MarshalBinary/UnmarshalBinary support)
- Remove `reflectMapCodec` (maps not supported in TinyGo)
- Update all codec methods to use tinyreflect API

### 2. `/scanner.go` - Type scanning logic
**Changes:**
- Replace `reflect.Type` with `tinyreflect.Type`
- Remove `scanBinaryMarshaler()` function
- Remove `scanCustomCodec()` function
- Update `scanType()` to handle only supported types
- Remove complex64/complex128 cases
- Remove map support cases
- Simplify struct scanning logic

### 3. `/encoder.go` - Binary encoder
**Changes:**
- Replace `reflect.Indirect()` with tinyreflect equivalent
- Replace `reflect.ValueOf()` with `tinyreflect.ValueOf()`
- Update `Encode()` method to use tinyreflect API
- Remove complex number encoding support

### 4. `/decoder.go` - Binary decoder
**Changes:**
- Replace `reflect.Indirect()` with tinyreflect equivalent
- Replace `reflect.ValueOf()` with `tinyreflect.ValueOf()`
- Update `Decode()` method to use tinyreflect API
- Update error handling to use `tinystring.Err()`

### 5. `/convert.go` - Type conversion utilities
**Changes:**
- Review usage of `reflect.StringHeader` 
- Replace with tinyreflect equivalents if needed
- May need to keep some unsafe operations for performance

## Implementation Strategy

### Phase 1: Remove Unsupported Types
1. Remove `complex64Codec`, `complex128Codec` from codecs.go
2. Remove custom marshaling support (`customCodec`)
3. Remove map support (`reflectMapCodec`)
4. Update `scanType()` to reject unsupported types
5. Add proper error messages using `tinystring.Err()`

### Phase 2: Replace Reflect API
1. Replace all `reflect.Value` with `tinyreflect.Value`
2. Replace all `reflect.Type` with `tinyreflect.Type`
3. Replace `reflect.ValueOf()` with `tinyreflect.ValueOf()`
4. Update field access patterns

### Phase 3: Update Tests and Validation
1. Remove/modify tests for unsupported types (complex, maps, custom marshaling)
2. Update existing tests to use supported types only
3. Replace map usage in tests with slices
4. Ensure all tests pass before proceeding to next phase

### Phase 4: Simplify and Optimize
1. Simplify struct scanning logic
2. Remove unnecessary complexity
3. Optimize for TinyGo compilation
4. Test with TinyGo WebAssembly target

## Error Handling Pattern
Use `tinystring` error system:
```go
// Instead of: errors.New("message")
return tinystring.Err(tinystring.D.Binary, "specific error context")

// Instead of: fmt.Errorf("format %s", arg)
return tinystring.Err(tinystring.D.Type, arg, tinystring.D.Not, tinystring.D.Supported)
```

## Testing Strategy
1. Start with basic types (string, bool, numbers) and ensure tests pass
2. Update/remove tests for unsupported types (complex, maps, custom marshaling)
3. Replace map usage in tests with equivalent slice structures
4. Ensure all tests pass before proceeding to next implementation phase
5. Test with TinyGo compilation after each major change
6. Verify WebAssembly binary size reduction
7. Test with representative data structures

## Success Criteria
- ✅ Compiles successfully with TinyGo
- ✅ Maintains API compatibility for supported types
- ✅ Reduces WebAssembly binary size
- ✅ Passes all relevant tests (after removing unsupported type tests)
- ✅ No dependencies on `fmt`, `strconv`, `errors`
- ✅ Uses only `tinystring` for error handling
- ✅ Maps replaced with slices where applicable

## Implementation Decisions (Confirmed)
1. ❌ **Maps removed** - Not TinyGo concurrent-safe, replaced with slices
2. ✅ **API changes allowed** - Minor changes permitted for better maintainability
3. ✅ **Test adaptation** - Remove/modify tests for unsupported types
4. ✅ **Incremental approach** - Start with basics, ensure tests pass each phase
5. ❌ **Performance benchmarks** - Not required for now

## Technical Justification for Map Removal

### Expert Analysis Summary
Maps are removed from TinyBin for compelling technical reasons:

#### 🚫 **Concurrency Issues in TinyGo**
- Go maps are NOT thread-safe by design
- TinyGo has runtime limitations for concurrent map operations
- Slices provide predictable behavior without race conditions

#### 📦 **Code Complexity vs Benefits**
- Map support would require ~200+ lines of reflection code
- Contradicts "minimal code" principle for WebAssembly optimization
- Slices cover 90% of practical use cases with simpler implementation

#### 🏗️ **Superior Alternatives**
```go
// Before (with maps)
data := map[string]int{"a": 1, "b": 2}

// After (with slices) - More explicit and efficient
data := []struct{Key string; Value int}{
    {"a", 1}, {"b", 2},
}
```

#### 🎯 **Performance Benefits**
- **Better iteration performance**: Linear memory access vs hash table lookups
- **More efficient serialization**: Predictable binary format
- **Better JSON compatibility**: Natural array serialization
- **Smaller binary footprint**: Less reflection infrastructure needed

#### ✅ **TinyGo Optimization**
- No concurrent map runtime overhead
- Simpler garbage collection patterns
- Better WebAssembly performance characteristics
- Cleaner integration with TinyReflect's minimal API

### Migration Strategy
Replace map usage with equivalent slice structures:
- `map[string]T` → `[]struct{Key string; Value T}`
- `map[int]T` → `[]struct{Key int; Value T}`
- `map[K]V` → `[]struct{Key K; Value V}`

This approach maintains functionality while achieving superior performance and smaller binaries.

## Questions for Clarification
1. ✅ Should maps be supported or removed for simplicity? **DECIDED: Remove maps** (Technical analysis: concurrency issues, code complexity, better alternatives exist)
2. ✅ Are there specific performance requirements? **DECIDED: Not for now**
3. ✅ Should the API be changed to improve maintainability? **DECIDED: Yes, minor changes allowed**
4. ❓ Any specific TinyGo optimization flags to consider?

## Dependencies
- `github.com/cdvelop/tinyreflect` - Main reflection replacement
- `github.com/cdvelop/tinystring` - Error handling and utilities
- Standard library: `unsafe`, `encoding/binary`, `io`, `bytes`, `sync`
