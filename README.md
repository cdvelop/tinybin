# TinyBin Protocol Specification

**Version:** 1.0  
**Target:** Go fullstack applications with WebAssembly support  
**TinyGo Compatible:** Yes  

## Overview

TinyBin is a binary data transfer protocol designed specifically for Go applications, with first-class support for WebAssembly and TinyGo compilation. The protocol focuses on transferring slices of structs with maximum efficiency and minimal overhead.

## Design Principles

- **Struct-only encoding**: Only supports `[]struct` and `struct` types
- **Index-based field encoding**: Fields are encoded by position, not name
- **Zero reflection overhead**: Uses cached struct analysis with `tinyreflect`
- **WebAssembly optimized**: Little-endian, streaming interface
- **TinyGo compatible**: No maps, minimal heap allocations
- **Strict validation**: Exact struct matching required

## Supported Data Types

- **Numeric**: `int8`, `int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`
- **Boolean**: `bool`
- **Text**: `string` (UTF-8 encoded)
- **Binary**: `[]byte`

## Wire Format

### Message Structure
```
[Protocol Header][Data Payload]
```

### Protocol Header (9 bytes)
```
[1 byte: Major Version]
[1 byte: Minor Version]  
[6 bytes: Struct Identifier]
[varint: Struct Count]
```

### Struct Identifier (6 bytes)
```
[3 bytes: Package Name][3 bytes: Struct Name]
```

**Examples:**
- Package `home`, Struct `login` → `[hom][log]`
- Package `api`, Struct `user` → `[api][use]`

**Requirements:**
- Package name must be ≥ 3 characters
- Struct name must be ≥ 3 characters
- Names shorter than 3 characters are rejected

### Data Payload
```
[Struct 0][Struct 1]...[Struct N]
```

Each struct contains fields in declaration order:
```
[Field 0][Field 1]...[Field N]
```

### Field Encoding by Type

| Type | Encoding |
|------|----------|
| `int8` | 1 byte, little-endian |
| `int16` | 2 bytes, little-endian |
| `int32` | 4 bytes, little-endian |
| `int64` | 8 bytes, little-endian |
| `uint8` | 1 byte |
| `uint16` | 2 bytes, little-endian |
| `uint32` | 4 bytes, little-endian |
| `uint64` | 8 bytes, little-endian |
| `float32` | 4 bytes, IEEE 754, little-endian |
| `float64` | 8 bytes, IEEE 754, little-endian |
| `bool` | 1 byte (`0x00` = false, `0x01` = true) |
| `string` | `[varint length][UTF-8 bytes]` |
| `[]byte` | `[varint length][raw bytes]` |

### Variable Integer Encoding (LEB128)

Uses Little Endian Base 128 (LEB128) encoding for:
- String/byte slice lengths
- Struct counts in header

**LEB128 Rules:**
- 7 bits of data per byte
- MSB = 1 indicates continuation
- MSB = 0 indicates final byte
- Maximum value: 2^32 (4GB limit)

## API Design

### Core Types

```go
type stObject struct {
    stName   string      // "package.struct" format
    stFields []stField   // field metadata
    offsets  []uintptr   // precalculated field offsets
}

type stField struct {
    name     string      // field name
    typeKind reflect.Kind // field type
    offset   uintptr     // field offset in struct
}

var stObjects = []stObject{} // global cache
```

### Primary API

```go
// Create new TinyBin handler
h := tinybin.New()

// Register struct types (must be done on both ends)
err := h.AddStructsObject(stObjects ...any)

// Encode slice of structs to writer
err := h.Encode(w io.Writer, structs ...any)

// Decode from reader to slice of structs  
structs, err := h.Decode(r io.Reader)
```

## File Structure

```
tinybin/
├── tinybin.go        // Main API, New(), AddStructsObject()
├── encoder.go        // Encode() and serialization functions
├── decoder.go        // Decode() and deserialization functions  
├── reflect.go        // tinyreflect, caching, struct analysis
├── types.go          // stObject, stField type definitions
├── errors.go         // Protocol-specific error types
├── varint.go         // LEB128 implementation
└── README.md         // Protocol documentation
```

## Error Handling

### Error Types

- **Version Mismatch**: Major version differences
- **Type Mismatch**: Struct identifier doesn't match registered types
- **Field Count Mismatch**: Different number of fields in struct
- **Field Type Mismatch**: Field type differs at specific index
- **Unsupported Type**: Non-primitive type encountered
- **Length Overflow**: String/byte slice exceeds 4GB limit
- **Corrupted Data**: Invalid varint or malformed data

### Error Behavior

- **Fail Fast**: First error aborts entire operation
- **Detailed Messages**: Include field names, indices, and expected vs actual types
- **Strict Validation**: Exact struct matching required between encoder and decoder

## Version Compatibility

### Semantic Versioning
- **Major Version** (1 byte): Breaking changes, incompatible formats
- **Minor Version** (1 byte): Compatible additions, backwards compatible

### Compatibility Rules
- **Strict Compatibility**: Major version must match exactly
- **Minor Differences**: Encoder/decoder warn but continue processing
- **Forward Compatibility**: Not supported, exact version matching preferred

## Performance Optimizations

### Memory Management
- **Global Struct Cache**: Reuse analyzed struct metadata
- **Precalculated Offsets**: Avoid reflection during encoding/decoding
- **Stack Allocation**: Minimize heap allocations where possible
- **Buffer Reuse**: Encoder/decoder instances reuse internal buffers

### WebAssembly Specific
- **Little-Endian Native**: No byte order conversion overhead
- **Streaming Interface**: Memory-efficient for large datasets
- **TinyGo Compatible**: No unsupported reflection features

## Usage Examples

### Basic Usage

```go
package main

import "tinybin"

type User struct {
    ID   int64  `html:"id"`
    Name string `html:"name"`
    Age  int32  `html:"age"`
}

func main() {
    // Initialize
    h := tinybin.New()
    err := h.AddStructsObject(User{})
    if err != nil {
        panic(err)
    }
    
    // Encode
    users := []User{
        {ID: 1, Name: "Alice", Age: 30},
        {ID: 2, Name: "Bob", Age: 25},
    }
    
    var buf bytes.Buffer
    err = h.Encode(&buf, users)
    if err != nil {
        panic(err)
    }
    
    // Decode
    decoded, err := h.Decode(&buf)
    if err != nil {
        panic(err)
    }
    
    users2 := decoded.([]User)
    fmt.Printf("Decoded: %+v\n", users2)
}
```

### WebAssembly Integration

```go
//go:build wasm
// +build wasm

package main

import (
    "syscall/js"
    "tinybin"
)

var handler *tinybin.TinyBin

func init() {
    handler = tinybin.New()
    handler.AddStructsObject(ApiRequest{}, ApiResponse{})
}

func encodeData(this js.Value, args []js.Value) interface{} {
    // Receive data from JavaScript
    // Encode with TinyBin
    // Return binary data to JavaScript
}

func decodeData(this js.Value, args []js.Value) interface{} {
    // Receive binary data from JavaScript  
    // Decode with TinyBin
    // Return structured data to JavaScript
}

func main() {
    js.Global().Set("encodeData", js.FuncOf(encodeData))
    js.Global().Set("decodeData", js.FuncOf(decodeData))
    select {} // Keep alive
}
```

## Limitations

1. **No Maps**: Maps are not supported due to TinyGo compatibility
2. **No Interfaces**: Only concrete struct types supported
3. **No Pointers**: Direct pointer encoding not supported
4. **No Nested Structs**: Only flat struct hierarchies
5. **Field Order Dependency**: Adding/removing/reordering fields breaks compatibility
6. **Name Length Requirement**: Package and struct names must be ≥ 3 characters

## Development Environment

- **Platform**: Windows 10 with GitBash
- **Go Version**: Compatible with standard Go and TinyGo
- **Dependencies**: Standard library only
- **Testing**: Unit tests for all encoding/decoding scenarios

## Security Considerations

- **Input Validation**: All inputs validated before processing
- **Buffer Overflow Protection**: Length limits enforced
- **Type Safety**: Strict type checking prevents data confusion
- **No Code Execution**: Pure data protocol, no executable content

## Future Considerations

- **Compression**: Optional compression for large payloads
- **Checksums**: Optional data integrity verification
- **Batch Operations**: Multi-struct-type encoding in single message
- **Schema Evolution**: Controlled field addition/removal strategies

---

**TinyBin Protocol v1.0** - Optimized for Go fullstack applications with WebAssembly support.

# USER

seria muy complejo que soportaramos strcuturas anidadas por ej ?

type User struct{
  Name string
  Addresses []Address
}

type Address struct{
Streeth string
Number uint16
}

no deseo que crees nada solo dame un analisi tecnico pro y contras con alternativas para tomar una decision, esto inmediatamnete hace persar que tambien deberia poder manejar todos los slices [] de number/string/struct y que nivel de anidamiento es recomendsble soportar? por favor ayudame a resolver estas interrogantes basado en el protocolo que estamos creando

# RESPONSE

## Análisis Técnico: Soporte de Estructuras Anidadas

### **Complejidad Técnica**

**PROS:**
- **Expresividad**: Modelos de datos más ricos y realistas
- **Reutilización**: `Address` se puede usar en múltiples structs
- **Semántica clara**: Relaciones de datos más naturales
- **Escalabilidad**: Permite diseños de API más sofisticados

**CONTRAS:**
- **Complejidad de implementación**: 
  - Recursión en encoding/decoding
  - Análisis de dependencias entre tipos
  - Gestión de referencias circulares potenciales
- **Cache management**: 
  - Necesitas cachear múltiples tipos relacionados
  - Validación de que todos los tipos necesarios estén registrados
- **Wire format más complejo**:
  - Necesitas identificadores para cada tipo anidado
  - Metadata adicional por cada nivel de anidamiento
- **TinyGo constraints**: 
  - Reflexión más compleja puede no ser soportada
  - Mayor uso de memoria para análisis recursivo

---

## **Alternativas de Implementación**

### **Alternativa A: Sin soporte (actual)**
```go
// SOLO esto funciona
type User struct {
    Name string
    Street string  // campos planos
    Number uint16
}
```
**Pros**: Simple, rápido, compatible con TinyGo
**Contras**: Modelos de datos limitados, duplicación

### **Alternativa B: Soporte limitado - 1 nivel**
```go
type User struct {
    Name string
    Address Address  // SOLO structs directos, NO slices
}

type Address struct {
    Street string
    Number uint16
}
```
**Wire format**: `[User fields][Address fields]` (aplanado)
**Complejidad**: Media, manejable

### **Alternativa C: Soporte completo con slices**
```go
type User struct {
    Name string
    Addresses []Address  // slices de structs
    Tags []string       // slices de primitivos
    Scores []int32      // todos los slice types
}
```
**Wire format**: Necesita length prefixes para cada slice
**Complejidad**: Alta

### **Alternativa D: Referencias por ID**
```go
type User struct {
    Name string
    AddressIDs []int64  // referencias, no objetos anidados
}
// Addresses se transfieren por separado
```
**Pros**: Evita anidamiento, control de relaciones
**Contras**: Lógica de resolución en aplicación

---

## **Nivel de Anidamiento Recomendado**

### **Para tu contexto (WebAssembly + TinyGo):**

**Nivel 0 (Actual)**: Solo primitivos
- ✅ Máxima compatibilidad TinyGo
- ✅ Performance óptima
- ❌ Expresividad limitada

**Nivel 1**: Structs anidados (sin slices)
- ✅ 80% de casos de uso cubiertos
- ✅ Complejidad manejable
- ✅ Compatible con TinyGo
- ❌ No permite listas relacionadas

**Nivel 2**: Slices de primitivos + structs anidados
```go
type User struct {
    Name string
    Tags []string      // slices de primitivos ✅
    Address Address    // struct anidado ✅
    Scores []int32     // más slices primitivos ✅
}
```
- ✅ Cubre 95% de casos de uso
- ⚠️ Complejidad media-alta
- ⚠️ Posibles issues con TinyGo

**Nivel 3**: Slices de structs anidados
```go
Addresses []Address  // slices de structs
```
- ✅ Máxima expresividad
- ❌ Complejidad muy alta
- ❌ Probable incompatibilidad TinyGo
- ❌ Gestión de memoria compleja en WASM

---

## **Impacto en Wire Format**

### **Actual (Nivel 0)**:
```
[version][struct_id][count][flat_fields...]
```

### **Nivel 1** (structs anidados):
```
[version][main_struct_id][nested_struct_ids...][count][fields_with_nested...]
```

### **Nivel 2+** (con slices):
```
[version][type_registry][count][complex_nested_data...]
```

---

## **Recomendación Específica**

**Para tu protocolo TinyBin, recomiendo NIVEL 1:**

```go
type User struct {
    Name string
    Address Address     // ✅ struct anidado
    ContactInfo Contact // ✅ múltiples structs anidados
    
    // NO soportar:
    // Addresses []Address  ❌ slice de structs
    // Tags []string        ❌ slices de primitivos por ahora
}
```

**Justificación:**
1. **Cubre el 80% de casos de uso** reales en APIs
2. **Mantiene compatibilidad TinyGo** (reflexión simple)
3. **Wire format manejable** (solo necesitas aplanar campos)
4. **Performance predecible** en WebAssembly
5. **Evolución gradual**: Puedes añadir slices en v2.0

**Implementación sugerida:**
- Analizar structs anidados recursivamente al registrar
- Aplanar todos los campos en el wire format
- Mantener metadata de "dónde empieza cada struct anidado"
- Un solo `struct_id` compuesto que incluya todas las dependencias

¿Te parece razonable empezar con Nivel 1 y evaluar Nivel 2 en una versión futura del protocolo?

# USER

nivel 3  slice de struct anidado un solo nivel , primitivos string/ numbers. justificacion las estructuras seran cacheadas . nota al margen el numero de identificacion para las estructuras deberia ser el mismo indice de el slice stObjects a cambio de usar [hom][log] solo numero uint16 o 32 para el wire format que numero me recomiendas? porfavor aclarame todas las dudas antes de continuar

# RESPONSE

## Análisis Técnico: Nivel 3 con Restricciones

### **Soporte Propuesto:**
```go
type User struct {
    Name string
    Age int32
    Addresses []Address  // ✅ slice de structs (1 nivel)
    Tags []string        // ✅ slice de primitivos
    Scores []int32       // ✅ slice de primitivos
}

type Address struct {
    Street string        // ✅ primitivos
    Number uint16        // ✅ primitivos
    // NO permitir: Contacts []Contact ❌ (sería nivel 2)
}
```

**Justificación válida**: Cache global hace viable el análisis recursivo.

---

## **Identificación Numérica: uint16 vs uint32**

### **uint16 (2 bytes)**
- **Rango**: 0-65,535 tipos únicos
- **Pros**: 
  - Menor overhead (2 bytes vs 6 bytes actuales)
  - Suficiente para aplicaciones típicas
  - Comparación de enteros más rápida que strings
- **Contras**: 
  - Límite de 65K tipos por aplicación
  - Requiere coordinación de IDs entre front/back

### **uint32 (4 bytes)**
- **Rango**: 0-4,294,967,295 tipos únicos  
- **Pros**:
  - Prácticamente ilimitado
  - Futuro-proof
- **Contras**:
  - 4 bytes vs 2 bytes overhead
  - Overkill para la mayoría de casos

**Recomendación: uint16** - 65K tipos es más que suficiente para cualquier aplicación realista.

---

## **Dudas Críticas a Resolver**

### **1. Gestión de IDs en el Cache**
```go
type stObject struct {
    stID     uint16      // índice en slice = ID del tipo
    stName   string      // mantener para debugging
    stFields []stField
    offsets  []uintptr
}
```

**¿Cómo sincronizas los IDs entre frontend y backend?**

**Opción A: Manual**
```go
// Mismo orden requerido en ambos lados
h.AddStructsObject(User{}, Address{}, Contact{}) // User=0, Address=1, Contact=2
```

**Opción B: Determinística**
```go
// Ordenar por nombre antes de asignar IDs
sort.Slice(stObjects, func(i, j int) bool {
    return stObjects[i].stName < stObjects[j].stName
})
```

### **2. Wire Format con Dependencias**

**¿Cómo manejas las dependencias en el wire format?**

**Ejemplo**: `User` contiene `[]Address`
- ¿Envías solo el ID de `User` (0) y asumes que el decoder conoce `Address`?
- ¿O envías ambos IDs en el header?

**Opción A: Solo tipo principal**
```
[version][main_type_id=0][count][user_data_with_addresses...]
```

**Opción B: Lista de dependencias**
```
[version][dep_count=2][type_id=0][type_id=1][count][data...]
```

### **3. Validación de Dependencias**

**¿Qué pasa si el decoder no tiene `Address` registrado cuando llega `User`?**

**Opción A: Error inmediato**
```go
if !h.hasType(addressID) {
    return fmt.Errorf("dependency type %d not registered", addressID)
}
```

**Opción B: Lazy loading**
```go
// Permitir decode parcial, error solo al acceder Address
```

### **4. Detección de Dependencias**

**¿Cómo detectas automáticamente las dependencias al registrar?**

```go
type User struct {
    Addresses []Address  // necesita detectar que Address es dependencia
    Contact Contact      // necesita detectar que Contact es dependencia
}
```

**¿Analizas recursivamente todos los campos durante `AddStructsObject()`?**

### **5. Slices de Primitivos vs Slices de Structs**

**Wire format diferente:**
```go
// Slice de primitivos
Tags []string → [varint_count][string1][string2]...

// Slice de structs  
Addresses []Address → [varint_count][address1_fields][address2_fields]...
```

**¿Mismo encoding para ambos o diferenciado?**

### **6. Recursión de Análisis**

**¿Límites de recursión para evitar stack overflow?**
```go
type A struct { B []B }
type B struct { A []A }  // ciclo infinito potencial
```

**¿Detección de ciclos o prohibición explícita?**

### **7. Orden de Registro**

**¿Importa el orden de registro?**
```go
// ¿Esto falla?
h.AddStructsObject(User{})    // User referencia Address
h.AddStructsObject(Address{}) // Address registrado después

// ¿O esto es obligatorio?
h.AddStructsObject(Address{}) // dependencias primero
h.AddStrut{})    // tipos que dependen después
```

---

## **Decisiones Inmediatas Necesarias**

1. **uint16 o uint32** para type IDs?
2. **¿Cómo sincronizar IDs** entre front/back?
3. **¿Wire format incluye dependencias** o solo tipo principal?
4. **¿Validación inmediata** de dependencias o lazy?
5. **¿Orden de registro** importa?
6. **¿Detección automática** de dependencias o manual?
7. **¿Límites de recursión** para evitar ciclos?

**¿Cuáles son tus preferencias para cada una de estas decisiones?**

# USER

0-la identifacion numerica uint16 Ok

1-A manual, ej arq propuesta para cargar la configuracion manual
```md
AppName/                        
│
├── modules/                    # 🔧 Lógica modular
│   ├── modules.go              #init tinybin and h.AddStrctsObjects here
│   │
│   ├── home/                   # 🏠 Módulo
│   │   ├── auth.go             # lógica de autenticación
│   │   ├── b.api.go            # 🔙 Backend API (// +build !wasm)
│   │   ├── f.auth.go           # 🌐 Frontend auth (// +build wasm)
│   │   └── handlers.go         # Handlers compartidos
│   │
│   ├── users/                  # 👥 Módulo de usuarios
│   │   ├── user.go             # Modelos de datos
│   │   ├── b.api.go            # 🔙 Backend API endpoints
│   │   ├── f.users.go          # 🌐 Frontend usuarios (// +build wasm)
│   │   └── f.events.go         # 🌐 Frontend pub/sub
│   │
│   └── medical/                # 🏥 Módulo médico
│       ├── patient.go          # Modelo de paciente
│       ├── b.api.go            # 🔙 Backend API
│       ├── f.medical.go        # 🌐 Frontend (// +build wasm)
│       └── handlers.go         
│
├── pwa/                       
│   ├── theme/              
│   │   ├── css/               
│   │   └── js/                
│   │
│   ├── public/ 
│   │   ├── main.js
│   │   ├── style.css          
│   │   ├── AppName.wasm
│   │   └── index.html
│   │
│   ├── main.server.go          # 🔙 use tinybin from modules.go // +build !wasm
│   └── main.wasm.go            # 🌐 use tinybin from modules.go // +build wasm
```

2- tinybin cuando almacena las estructuras anidadas o suletas debe de hacerlo por separado como stObject{} en stObjects solo debe enviarce la estructura padre para el ejemplo que me pregundas seria solo id User(0) al buscarla debemos identificar el tipo de campo luego si es otra estructua buscar su estructura para encode/decode

3- A error inmediato eso no da seguridad de que solo resibiremos estructuras que se codificaron previamemte (comenta esta como beneficio de esta libreria)

4- analizar recursivamente tolos los camoos durante AddStructObject

5- en este punto 5 necesito que me guies para tomar una buena decision

6- en este punto 6 necesito orientacion para tomar una buena decision

7- si importa front como backen deben compilar el mismo orden como te lo aclare en el punto 1

necesito que me orientes en las preguntas que te respondi y que me agregues mas si aun faltan otras punto. si añades mas por favor agragalas con alternativa y recomendacion segun esta proyecto para tomar una buena desicion

# RESPONSE



