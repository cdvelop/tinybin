package tinybin

import (
	"bytes"
	"io"
	"reflect"
	"sync"

	. "github.com/cdvelop/tinystring"
)

// TinyBin represents a binary encoder/decoder with isolated state.
// This replaces the previous global variable-based architecture.
type TinyBin struct {
	// log is an optional custom logging function
	log func(msg ...any)

	// schemas is a slice-based cache for TinyGo compatibility (no maps allowed)
	schemas []schemaEntry

	// encoders is a private pool for encoder instances
	encoders *sync.Pool

	// decoders is a private pool for decoder instances
	decoders *sync.Pool

	// Mutex to protect schemas slice
	mu sync.RWMutex
}

// schemaEntry represents a cached schema with its type and codec
type schemaEntry struct {
	Type  reflect.Type
	Codec Codec
}

// New creates a new TinyBin instance with optional configuration.
// The first argument can be an optional logging function.
// If no logging function is provided, a no-op logger is used.
// eg: tb := tinybin.New(func(msg ...any) { fmt.Println(msg...) })

func New(args ...any) *TinyBin {
	var logFunc func(msg ...any) // Default: no logging

	for _, arg := range args {
		if log, ok := arg.(func(msg ...any)); ok {
			logFunc = log
		}
	}

	tb := &TinyBin{log: logFunc}

	tb.schemas = make([]schemaEntry, 0, 100) // Pre-allocate reasonable size
	tb.encoders = &sync.Pool{
		New: func() any {
			return &encoder{
				tb: tb,
			}
		},
	}
	tb.decoders = &sync.Pool{
		New: func() any {
			return &decoder{
				tb: tb,
			}
		},
	}

	return tb
}

// Encode encodes the payload into binary format using this TinyBin instance.
func (tb *TinyBin) Encode(data any) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Grow(64)

	// Encode and set the buffer if successful
	if err := tb.EncodeTo(data, &buffer); err == nil {
		return buffer.Bytes(), nil
	} else {
		return nil, err
	}
}

// EncodeTo encodes the payload into a specific destination using this TinyBin instance.
func (tb *TinyBin) EncodeTo(data any, dst io.Writer) error {
	// Get the encoder from the pool, reset it
	e := tb.encoders.Get().(*encoder)
	e.Reset(dst, tb)

	// Encode and set the buffer if successful
	err := e.Encode(data)

	// Put the encoder back when we're finished
	tb.encoders.Put(e)
	return err
}

// Decode decodes the payload from the binary format using this TinyBin instance.
func (tb *TinyBin) Decode(data []byte, target any) error {
	// Get the decoder from the pool, reset it
	d := tb.decoders.Get().(*decoder)
	d.Reset(data, tb)

	// Decode and free the decoder
	err := d.Decode(target)
	tb.decoders.Put(d)
	return err
}

// findSchema performs a linear search in the slice-based cache for TinyGo compatibility
func (tb *TinyBin) findSchema(t reflect.Type) (Codec, bool) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	for _, entry := range tb.schemas {
		if entry.Type == t {
			return entry.Codec, true
		}
	}
	return nil, false
}

// addSchema adds a new schema to the slice-based cache
func (tb *TinyBin) addSchema(t reflect.Type, codec Codec) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	// Simple cache size limit (optional, for memory control)
	if len(tb.schemas) >= 1000 { // Reasonable default limit
		// Simple eviction: remove oldest (first) entry
		tb.schemas = tb.schemas[1:]
	}

	tb.schemas = append(tb.schemas, schemaEntry{
		Type:  t,
		Codec: codec,
	})
}

// scanToCache scans the type and caches it in the TinyBin instance using slice-based cache
func (tb *TinyBin) scanToCache(t reflect.Type) (Codec, error) {
	if t == nil {
		return nil, Err("scanToCache", "type", "nil")
	}

	// Check if we already have this schema cached
	if c, found := tb.findSchema(t); found {
		return c, nil
	}

	// Scan for the first time
	c, err := scan(t)
	if err != nil {
		return nil, err
	}

	// Cache the schema
	tb.addSchema(t, c)

	return c, nil
}
