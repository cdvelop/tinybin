# TinyBin Cache Integration Guide

## Executive Summary

This document describes how TinyBin will integrate with the new TinyReflect cache system. The changes outlined here complement the main TinyReflect cache proposal and provide specific guidance for migrating TinyBin's current caching architecture to leverage TinyReflect's new instance-based cache system.

For the main TinyReflect cache proposal, see [CACHE_TINYREFLECT.md](CACHE_TINYREFLECT.md).

## Current TinyBin Cache System

### Architecture Overview

El sistema actual implementa un caché de tipos eficiente con la siguiente arquitectura:

```go
type TinyBin struct {
    schemas *sync.Map  // Caché global de la instancia
    // ...
}

type Encoder struct {
    schemas map[*tinyreflect.Type]Codec  // Caché local del encoder
    // ...
}

type Decoder struct {
    schemas map[*tinyreflect.Type]Codec  // Caché local del decoder
    // ...
}
```

### Current Operation Flow

**Funcionamiento actual:**
1. **Caché de dos niveles:** Global (`sync.Map`) + local (`map`)
2. **Operación `scanToCache`:** Verifica caché local → global → genera nuevo codec
3. **Reutilización:** Los codecs se almacenan y reutilizan para tipos idénticos
4. **Thread-safety:** `sync.Map` para acceso concurrente seguro

### Problems with Current System

- **TinyGo incompatibility:** `sync.Map` is not supported in TinyGo
- **Redundant caching:** Both TinyBin and TinyReflect implement separate cache systems
- **Memory overhead:** Double caching of type information
- **Synchronization complexity:** Managing two cache levels increases complexity

## Migration to TinyReflect Cache Integration

### New Architecture (TinyGo-Compatible)

```go
// Integración con TinyBin (simplificada)
type TinyBin struct {
    reflect *TinyReflect  // Instancia optimizada para TinyGo
    // ... otros campos sin maps ni variables globales
}

func NewTinyBin() *TinyBin {
    return &TinyBin{
        reflect: tinyreflect.New(
            tinyreflect.WithMaxStructs(64),
            tinyreflect.WithMaxBasics(32),
            tinyreflect.WithMaxNames(32),  // Para struct name cache
        ),
    }
}

// Ejemplo de migración: todas las variables globales ahora son parte de la instancia
func (tb *TinyBin) processStruct(s any) {
    // Antes: usaba variable global structNameCache
    // Ahora: usa instancia tb.reflect con cache interno
    value := tb.reflect.ValueOf(s)
    typ := tb.reflect.TypeOf(s)
    name := typ.Name() // Usa cache interno, no variable global
}
```

### Benefits of Integration

1. **TinyGo Compatibility:** No more `sync.Map` - uses TinyReflect's array-based cache
2. **Single Source of Truth:** Eliminates redundant caching between libraries
3. **Memory Efficiency:** Shared cache reduces overall memory usage
4. **Performance:** Leverages TinyReflect's optimized cache for all type operations
5. **Simplicity:** Removes complex two-level caching logic

## Migration Strategy for TinyBin

### Phase 1: Preparation
1. Update TinyBin to depend on the new TinyReflect cache API
2. Replace `sync.Map` usage with TinyReflect instance methods
3. Remove local cache maps from Encoder/Decoder structs
4. Update constructor to initialize TinyReflect instance

### Phase 2: Implementation Changes

#### Before (Current TinyBin)
```go
type TinyBin struct {
    schemas *sync.Map
}

type Encoder struct {
    schemas map[*tinyreflect.Type]Codec
}

func (e *Encoder) scanToCache(typ *tinyreflect.Type) {
    // Check local cache
    if codec, exists := e.schemas[typ]; exists {
        return codec
    }
    
    // Check global cache
    if val, ok := e.tinybin.schemas.Load(typ); ok {
        codec := val.(Codec)
        e.schemas[typ] = codec
        return codec
    }
    
    // Generate new codec...
}
```

#### After (Integrated with TinyReflect)
```go
type TinyBin struct {
    reflect *TinyReflect  // Single cache source
}

type Encoder struct {
    tinybin *TinyBin  // Reference to parent
}

func (e *Encoder) scanToCache(i any) {
    // Use TinyReflect's optimized cache
    value := e.tinybin.reflect.ValueOf(i)
    typ := e.tinybin.reflect.TypeOf(i)
    
    // TinyReflect handles all caching internally
    // No need for explicit cache management
}
```

### Phase 3: Testing and Validation

1. **TinyGo Compatibility Testing:** Ensure compilation and execution on TinyGo
2. **Performance Benchmarks:** Compare before/after performance
3. **Memory Usage Analysis:** Validate reduced memory footprint
4. **WebAssembly Testing:** Test in actual WebAssembly environments

## Configuration Recommendations

### Recommended Cache Sizes for TinyBin

```go
func NewTinyBin() *TinyBin {
    return &TinyBin{
        reflect: tinyreflect.New(
            // Typical struct types in serialization
            tinyreflect.WithMaxStructs(64),
            
            // Basic types (int, string, float, etc.)
            tinyreflect.WithMaxBasics(32),
            
            // Slice and array types
            tinyreflect.WithMaxSlices(24),
            
            // Struct names cache
            tinyreflect.WithMaxNames(32),
        ),
    }
}
```

### Memory Usage Estimation

| Cache Type | Entry Size | Max Entries | Total Memory |
|------------|------------|-------------|--------------|
| Structs    | ~512 bytes | 64          | ~32KB        |
| Basics     | ~64 bytes  | 32          | ~2KB         |
| Slices     | ~32 bytes  | 24          | ~768 bytes   |
| Names      | ~72 bytes  | 32          | ~2.3KB       |
| **Total**  |            |             | **~37KB**    |

## API Changes Summary

### Removed Components
- ❌ `sync.Map` usage
- ❌ Local cache maps in Encoder/Decoder
- ❌ Two-level cache management logic
- ❌ Manual cache synchronization

### Added Components
- ✅ TinyReflect instance integration
- ✅ Simplified cache access through TinyReflect
- ✅ TinyGo-compatible operation
- ✅ Automatic cache management

### Migration Checklist

- [ ] Replace `sync.Map` with TinyReflect instance
- [ ] Remove local cache maps from structs
- [ ] Update all type operations to use TinyReflect methods
- [ ] Test compilation with TinyGo
- [ ] Benchmark performance improvements
- [ ] Update documentation and examples
- [ ] Validate WebAssembly compatibility

## Performance Expectations

### Before (Current System)
- Cache hit: O(1) map lookup + O(1) sync.Map lookup
- Cache miss: O(n) type processing + double cache storage
- Memory: Duplicate storage in both caches
- TinyGo: ❌ Not compatible

### After (TinyReflect Integration)
- Cache hit: O(1) array lookup (faster than map)
- Cache miss: O(n) type processing + single cache storage
- Memory: Single cache storage, shared across operations
- TinyGo: ✅ Fully compatible

## Conclusion

The integration with TinyReflect's cache system eliminates TinyGo compatibility issues while improving performance and reducing memory usage. The simplified architecture removes the complexity of dual-level caching and provides a more maintainable codebase.

This migration is essential for TinyGo/WebAssembly support and aligns with the minimalist philosophy of the tiny* ecosystem.

---

**Related Documents:**
- [TinyReflect Cache Proposal](CACHE_TINYREFLECT.md) - Main cache system design
- [TinyGo Compatibility Guide](../README.md#tinygo-support) - General TinyGo considerations