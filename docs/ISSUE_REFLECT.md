# TinyBin Reflect Migration Plan

## 🚨 CRITICAL RESTRICTION: Standard Library Replacement

### **Zero Standard Library Dependencies**
**TinyBin MUST NOT use any standard library packages that are not optimized for frontend/WebAssembly:**

#### ❌ **PROHIBITED Packages:**
- `fmt` - Too heavy for WebAssembly, use `tinystring.Fmt()` instead
- `strconv` - Not optimized for TinyGo, use `tinystring.Convert()` instead  
- `errors` - Use `tinystring.Err()` for multilingual error handling
- `strings` - Use `tinystring.Convert()` methods instead
- `reflect` - Being replaced with `tinyreflect` (lightweight alternative)

#### ✅ **ALLOWED Standard Library:**
- `unsafe` - Required for low-level operations
- `encoding/binary` - Essential for binary protocol
- `io` - Core I/O interfaces
- `bytes` - Buffer operations
- `sync` - Concurrency primitives
- `math` - Mathematical operations

### **Why This Restriction Exists:**
1. **WebAssembly Optimization**: Standard library packages are designed for server-side Go, not frontend/WebAssembly
2. **Binary Size**: Standard packages add significant bloat to WebAssembly binaries
3. **Performance**: TinyString provides faster, more efficient implementations for our use case
4. **Consistency**: Using TinyString throughout ensures consistent error handling and formatting

### **Migration Requirements:**
```go
// ❌ NEVER use these:
import "fmt"
import "strconv" 
import "errors"
import "strings"

// ✅ ALWAYS use TinyString instead:
import . "github.com/cdvelop/tinystring"

// Examples:
fmt.Sprintf("error: %s", msg)     // ❌ Forbidden
Fmt("error: %s", msg)             // ✅ Correct

strconv.Itoa(42)                  // ❌ Forbidden  
Convert(42).String()              // ✅ Correct

errors.New("message")             // ❌ Forbidden
Err("message")                    // ✅ Correct
```

---

## 🎯 Current Status

### Phase 1: Remove Unsupported Types ✅ **COMPLETED**
- Removed complex64/128, custom marshaling, maps from codecs.go
- Updated scanner.go to reject unsupported types
- **Results**: 27/30 tests passing (90%) - Only failed tests use unsupported types

### Phase 2: Replace reflect API with tinyreflect ✅ **COMPLETED**
- Added SliceType and ArrayType to tinyreflect package
- Replaced reflect.TypeOf() with tinyreflect.TypeOf() in scanner.go
- **Results**: All supported types now work correctly with tinyreflect

### Phase 3: Implement tinyreflect.Value API ✅ **COMPLETED**
- ✅ **Set Methods**: SetString, SetBool, SetInt, SetUint, SetFloat, SetBytes
- ✅ **Access Methods**: Index, Len, Cap, IsNil, Addr, Elem, Set
- ✅ **Core Methods**: Type, Field, Interface, NumField
- ✅ **Package Functions**: ValueOf, TypeOf, Indirect, MakeSlice, New, Zero

### Phase 4: Standard Library Implementation Adaptation ✅ **COMPLETED**
- ✅ **Elem() Method**: Implemented following `/usr/local/go/src/reflect/value.go` line 1218
- ✅ **Index() Method**: Implemented for arrays, slices, strings following standard library
- ✅ **Len() Method**: Implemented for arrays, slices, strings following standard library
- ✅ **Cap() Method**: Implemented for arrays, slices following standard library
- ✅ **Set() Method**: Implemented with byte-level copying following standard library
- ✅ **PtrType Structure**: Defined following `/usr/local/go/src/internal/abi/type.go` line 549
- ✅ **Flag Constants**: Updated following `/usr/local/go/src/reflect/value.go` line 76
- ✅ **Test Coverage**: All tinyreflect tests passing (15/15)

### Phase 5: Debug tinybin Integration ✅ **MAJOR PROGRESS** 
- ✅ **Root Cause Identified**: Issue was in `reflectSliceCodec.EncodeTo()` using `Addr()` incorrectly
- ✅ **Fix Applied**: Avoid `Addr()` and use slice elements directly in encoding
- ✅ **Key Tests Fixed**: `TestBinarySimpleStructSlice`, `TestSliceOfStructWithStruct` now pass
- ✅ **Method Verification**: All tinyreflect methods working correctly in isolation
- ✅ **PtrType Implementation**: Correctly following `/usr/local/go/src/internal/abi/type.go`
- ✅ **StructField.Typ**: All field types properly initialized
- ❌ **Remaining Issues**: Pointer handling ("value type nil" errors), Array codec
- ⏳ **Next**: Fix `Addr()` implementation following `/usr/local/go/src/reflect/value.go:269`
- [ ] Test compilation with TinyGo
- [ ] Run WebAssembly tests
- [ ] Validate binary size improvement

---

## 📊 Current Implementation Status

### ✅ **tinyreflect API Complete - Standard Library Adapted**
All essential methods implemented following Go standard library patterns:

**Implementation Sources:**
- **Core Value Methods**: Adapted from `/usr/local/go/src/reflect/value.go`
- **Type Structures**: Adapted from `/usr/local/go/src/internal/abi/type.go`
- **Flag Constants**: Adapted from `/usr/local/go/src/reflect/value.go`
- **Original Code Reference**: Compare with `/home/cesar/Dev/Pkg/Other/binary/` for faithful migration patterns

**Completed Methods:**
- **Value methods**: 16 methods implemented (Set*, Index, Len, Cap, Elem, etc.)
- **Package functions**: 6 functions implemented (ValueOf, TypeOf, Indirect, etc.)
- **Error handling**: All methods return errors instead of panic
- **Test coverage**: 15/15 tests passing in tinyreflect

**Key Implementations:**
- `Elem()`: Line 1218 from reflect/value.go - handles pointer dereferencing
- `Index()`: Array/slice/string indexing with bounds checking
- `Len()`/`Cap()`: Array/slice length and capacity
- `Set()`: Byte-level memory copying for value assignment
- `PtrType`: Following internal/abi/type.go structure

### ❌ **Next Steps Required**
1. **Compare with original**: Use `/home/cesar/Dev/Pkg/Other/binary/` as reference implementation
2. **Test original first**: Verify `cd /home/cesar/Dev/Pkg/Other/binary && go test -v -run TestName` passes
3. **Debug "value type nil" errors**: Investigate why Values have `typ_` nil in tinybin
4. **Follow standard library debugging**: Use reflect tests to identify missing methods
5. **Side-by-side comparison**: Compare original vs tinyreflect method implementations
6. **Systematic approach**: Find failing method → locate in standard library → implement → test
7. **Maintain minimal implementation**: Adapt existing code, don't reinvent
8. **Test with TinyGo**: Validate WebAssembly compilation after fixes

## Key Differences: reflect vs tinyreflect

### 🚨 Critical Error Handling Difference
- **reflect**: Uses `panic()` for error conditions
- **tinyreflect**: Returns `error` instead of panicking

### Migration Pattern
```go
// ❌ OLD (reflect - can panic):
field := t.Field(i)
value := v.Interface()

// ✅ NEW (tinyreflect - returns errors):
field, err := t.Field(i)
if err != nil {
    return nil, err
}
value, err := v.Interface()
if err != nil {
    return nil, err
}
```

## Supported Types
**✅ Supported:**
- Basic types: `string`, `bool`, all numeric types
- Slices and structs with supported field types
- Pointers to supported types

**❌ Unsupported:**
- `complex64`, `complex128`
- `interface{}`, `chan`, `func` 
- Maps (replaced with slices)
- Custom marshaling methods

## Files to Modify

### 1. `/codecs.go` - Main codec implementations
- Replace `reflect.Value` with `tinyreflect.Value`
- Add error handling for all tinyreflect calls
- Update all codec methods to use tinyreflect API

### 2. `/encoder.go` & `/decoder.go` - Binary encoder/decoder
- Replace `reflect.Indirect()` with `tinyreflect.Indirect()`
- Replace `reflect.ValueOf()` with `tinyreflect.ValueOf()`
- Add error handling for tinyreflect calls

### 3. `/convert.go` - Type conversion utilities
- Review usage of `reflect.StringHeader` 
- Replace with tinyreflect equivalents if needed

## Implementation Strategy

### Next Phase: Debug Integration Issues

**Current Status:**
- ✅ tinyreflect implementation: Complete (all methods adapted from standard library)
- ✅ tinyreflect tests: All passing (15/15)
- ❌ tinybin integration: Failing with "value type nil" errors

**Debugging Strategy:**
1. **Review original binary**: Compare original binary package code in VSC (matching filenames) to validate migration  
2. **Identify failing method**: Find specific method call that produces nil type  
3. **Locate in standard library**: Find implementation in `/usr/local/go/src/reflect/`  
4. **Implement following standard**: Adapt exact logic, don’t reinvent  
5. **Add corresponding test**: Create test based on standard library tests  
6. **Verify and iterate**: Ensure all tests pass before next method  


**Implementation Philosophy:**
- **Minimal adaptation**: Use existing standard library logic
- **No reinvention**: Follow proven patterns from Go reflect package
- **Systematic approach**: One method at a time, test-driven
- **Error-based**: Convert panics to errors for tinyreflect compatibility
- **Debug tests as coverage**: Keep useful debug tests as coverage tests, improve informational ones, eliminate redundant ones

**Debug Test Management Strategy:**
1. **Useful for coverage**: Rename and keep tests that increase test coverage (e.g., `debug_field_test.go` → `field_access_test.go`)
2. **Informational but important**: Adjust tests that provide valuable debugging info to be more useful for coverage
3. **Redundant**: Eliminate tests that duplicate existing functionality without adding value
4. **Test naming**: Use descriptive names that indicate the specific functionality being tested

## Success Criteria
- ✅ Compiles successfully with TinyGo
- ✅ Maintains API compatibility for supported types
- ✅ Reduces WebAssembly binary size
- ✅ Passes all relevant tests
- ✅ No dependencies on `fmt`, `strconv`, `errors`

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

## Dependencies
- `github.com/cdvelop/tinyreflect` - Main reflection replacement
- `github.com/cdvelop/tinystring` - Error handling and utilities
- Standard library: `unsafe`, `encoding/binary`, `io`, `bytes`, `sync`
