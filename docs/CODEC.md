# Codec Interface and Utilities

## Codec Interface

For custom types, implement the `Codec` interface:

```go
type Codec interface {
    EncodeTo(*encoder, reflect.Value) error
    DecodeTo(*decoder, reflect.Value) error
}
```

## Utility Functions

### `ToString(b *[]byte) string`
Converts a byte slice to string without allocation (unsafe operation).

### `ToBytes(v string) []byte`
Converts a string to byte slice without allocation (unsafe operation).