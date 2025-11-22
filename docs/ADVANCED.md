# Advanced Usage

## Multiple Instance Usage

```go
// Create multiple isolated instances
httpTB := tinybin.New()
grpcTB := tinybin.New()
kafkaTB := tinybin.New()

// Each instance maintains its own cache and pools
httpData, _ := httpTB.Encode(data)
grpcData, _ := grpcTB.Encode(data)
kafkaData, _ := kafkaTB.Encode(data)
```

## Custom Instance with Logging

```go
// Create instance with custom logging for debugging
tb := tinybin.New(func(msg ...any) {
    log.Printf("TinyBin Debug: %v", msg)
})

// Use like normal
data, err := tb.Encode(myStruct)
if err != nil {
    log.Printf("Encoding failed: %v", err)
}
```

## Concurrent Usage

```go
tb := tinybin.New()

// Safe concurrent usage - internal pooling handles synchronization
go func() {
    data, _ := tb.Encode(data1)
    process(data)
}()

go func() {
    data, _ := tb.Encode(data2)
    process(data)
}()
```

## Error Handling

```go
tb := tinybin.New()

data, err := tb.Encode(myValue)
if err != nil {
    // Handle encoding error
    log.Printf("Encoding failed: %v", err)
}

var result MyType
err = tb.Decode(data, &result)
if err != nil {
    // Handle decoding error
    log.Printf("Decoding failed: %v", err)
}
```

## Multiple Instance Patterns

**Microservices Pattern**: Different services can use separate instances for complete isolation.

```go
type ProtocolManager struct {
    httpTinyBin  *tinybin.TinyBin
    grpcTinyBin  *tinybin.TinyBin
    kafkaTinyBin *tinybin.TinyBin
}

func NewProtocolManager() *ProtocolManager {
    return &ProtocolManager{
        httpTinyBin:  tinybin.New(), // Production: no logging
        grpcTinyBin:  tinybin.New(),
        kafkaTinyBin: tinybin.New(),
    }
}
```

**Concurrent Processing**: Multiple instances can be used safely across goroutines.

```go
// Each goroutine gets its own instance for complete isolation
go func() {
    tb := tinybin.New()
    data, _ := tb.Encode(data1)
    process(data)
}()

go func() {
    tb := tinybin.New()
    data, _ := tb.Encode(data2) // Completely independent
    process(data)
}()
```