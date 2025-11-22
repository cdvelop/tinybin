# TinyBin API Reference

## Core API

### Encoding Functions

#### `(*TinyBin) Encode(v any) ([]byte, error)`
Encodes any value into binary format and returns the resulting bytes.

```go
tb := tinybin.New()
data, err := tb.Encode(myStruct)
```

#### `(*TinyBin) EncodeTo(v any, dst io.Writer) error`
Encodes a value directly to an `io.Writer`.

```go
tb := tinybin.New()
var buf bytes.Buffer
err := tb.EncodeTo(myStruct, &buf)
```

### Decoding Functions

#### `(*TinyBin) Decode(b []byte, v any) error`
Decodes binary data into a value. The destination must be a pointer.

```go
tb := tinybin.New()
var result MyStruct
err := tb.Decode(data, &result)
```

### encoder Type

**Note**: Encoders are now managed internally by TinyBin instances through object pooling for better performance and resource management. Direct creation of encoders is deprecated.

#### `(*encoder) Encode(v any) error`
Encodes a value using the encoder instance.

```go
tb := tinybin.New()
var buffer bytes.Buffer
err := tb.EncodeTo(myValue, &buffer) // Uses pooled encoder internally
```

#### `(*encoder) Buffer() io.Writer`
Returns the underlying writer.

```go
tb := tinybin.New()
var buffer bytes.Buffer
tb.EncodeTo(myValue, &buffer)
writer := buffer // Direct access to buffer
```

### encoder Write Methods

The `encoder` type provides methods for writing primitive types:

- `Write(p []byte)` - writes raw bytes
- `WriteVarint(v int64)` - writes a variable-length signed integer
- `WriteUvarint(x uint64)` - writes a variable-length unsigned integer
- `WriteUint16(v uint16)` - writes a 16-bit unsigned integer
- `WriteUint32(v uint32)` - writes a 32-bit unsigned integer
- `WriteUint64(v uint64)` - writes a 64-bit unsigned integer
- `WriteFloat32(v float32)` - writes a 32-bit floating point number
- `WriteFloat64(v float64)` - writes a 64-bit floating point number
- `WriteBool(v bool)` - writes a boolean value
- `WriteString(v string)` - writes a string with length prefix

### decoder Type

**Note**: Decoders are now managed internally by TinyBin instances through object pooling for better performance and resource management. Direct creation of decoders is deprecated.

#### `(*TinyBin) Decode(data []byte, v any) error`
Decodes binary data into a value using the TinyBin instance. The destination must be a pointer.

```go
tb := tinybin.New()
var result MyStruct
err := tb.Decode(data, &result)
```

### decoder Read Methods

The `decoder` type provides methods for reading primitive types:

- `Read(b []byte) (int, error)` - reads raw bytes
- `ReadVarint() (int64, error)` - reads a variable-length signed integer
- `ReadUvarint() (uint64, error)` - reads a variable-length unsigned integer
- `ReadUint16() (uint16, error)` - reads a 16-bit unsigned integer
- `ReadUint32() (uint32, error)` - reads a 32-bit unsigned integer
- `ReadUint64() (uint64, error)` - reads a 64-bit unsigned integer
- `ReadFloat32() (float32, error)` - reads a 32-bit floating point number
- `ReadFloat64() (float64, error)` - reads a 64-bit floating point number
- `ReadBool() (bool, error)` - reads a boolean value
- `ReadString() (string, error)` - reads a length-prefixed string
- `Slice(n int) ([]byte, error)` - returns a slice of the next n bytes
- `ReadSlice() ([]byte, error)` - reads a variable-length byte slice

## TinyBin Constructor and Instance Architecture

### Creating Instances

#### `New(args ...any) *TinyBin`
Creates a new TinyBin instance with optional configuration. Each instance is completely isolated from others.

```go
// Basic instance (no logging)
tb := tinybin.New()

// With custom logging
tb := tinybin.New(func(msg ...any) {
    log.Printf("TinyBin: %v", msg)
})
```

### Instance Isolation Benefits

**Complete State Isolation**: Each TinyBin instance maintains its own:
- Schema cache (slice-based for TinyGo compatibility)
- encoder and decoder object pools
- Optional logging function

**Thread Safety**: Multiple goroutines can safely use the same instance concurrently without external synchronization.

**Testing Benefits**: Each test can create its own instance with custom logging for complete isolation.

```go
func TestMyFunction(t *testing.T) {
    // Completely isolated test instance
    tb := tinybin.New(func(msg ...any) {
        t.Logf("TinyBin: %v", msg)
    })

    data, err := tb.Encode(testData)
    assert.NoError(t, err)
}
```