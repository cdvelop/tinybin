package tinybin

import (
	"io"
	"math"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

// Note: Encoder pool is now managed by TinyBin instance

// Encoder represents a binary encoder.
type Encoder struct {
	scratch [10]byte
	tb      *TinyBin // Reference to the TinyBin instance for schema caching
	out     io.Writer
	err     error
}

// NewEncoder creates a new encoder (deprecated - use TinyBin instance methods).
func NewEncoder(out io.Writer) *Encoder {
	return &Encoder{
		out: out,
	}
}

// Reset resets the encoder and makes it ready to be reused.
func (e *Encoder) Reset(out io.Writer, tb *TinyBin) {
	e.out = out
	e.err = nil
	e.tb = tb
}

// Buffer returns the underlying writer.
func (e *Encoder) Buffer() io.Writer {
	return e.out
}

// Encode encodes the value to the binary format.
func (e *Encoder) Encode(v any) (err error) {

	// Scan the type (this will load from cache)
	rv := tinyreflect.Indirect(tinyreflect.ValueOf(v))
	typ := rv.Type()
	if typ == nil {
		return Errf("cannot encode nil value")
	}

	var c Codec
	if c, err = e.scanToCache(typ); err != nil {
		return
	}

	// Encode the value
	if err = c.EncodeTo(e, rv); err == nil {
		err = e.err
	}
	return
}

// Write writes the contents of p into the buffer.
func (e *Encoder) Write(p []byte) {
	if e.err == nil {
		_, e.err = e.out.Write(p)
	}
}

// WriteVarint writes a variable size integer
func (e *Encoder) WriteVarint(v int64) {
	x := uint64(v) << 1
	if v < 0 {
		x = ^x
	}

	i := 0
	for x >= 0x80 {
		e.scratch[i] = byte(x) | 0x80
		x >>= 7
		i++
	}
	e.scratch[i] = byte(x)
	e.Write(e.scratch[:(i + 1)])
}

// WriteUvarint writes a variable size unsigned integer
func (e *Encoder) WriteUvarint(x uint64) {
	i := 0
	for x >= 0x80 {
		e.scratch[i] = byte(x) | 0x80
		x >>= 7
		i++
	}
	e.scratch[i] = byte(x)
	e.Write(e.scratch[:(i + 1)])
}

// WriteUint16 writes a Uint16
func (e *Encoder) WriteUint16(v uint16) {
	e.scratch[0] = byte(v)
	e.scratch[1] = byte(v >> 8)
	e.Write(e.scratch[:2])
}

// WriteUint32 writes a Uint32
func (e *Encoder) WriteUint32(v uint32) {
	e.scratch[0] = byte(v)
	e.scratch[1] = byte(v >> 8)
	e.scratch[2] = byte(v >> 16)
	e.scratch[3] = byte(v >> 24)
	e.Write(e.scratch[:4])
}

// WriteUint64 writes a Uint64
func (e *Encoder) WriteUint64(v uint64) {
	e.scratch[0] = byte(v)
	e.scratch[1] = byte(v >> 8)
	e.scratch[2] = byte(v >> 16)
	e.scratch[3] = byte(v >> 24)
	e.scratch[4] = byte(v >> 32)
	e.scratch[5] = byte(v >> 40)
	e.scratch[6] = byte(v >> 48)
	e.scratch[7] = byte(v >> 56)
	e.Write(e.scratch[:8])
}

// WriteFloat32 a 32-bit floating point number
func (e *Encoder) WriteFloat32(v float32) {
	e.WriteUint32(math.Float32bits(v))
}

// WriteFloat64 a 64-bit floating point number
func (e *Encoder) WriteFloat64(v float64) {
	e.WriteUint64(math.Float64bits(v))
}

// WriteBool writes a single boolean value into the buffer
func (e *Encoder) writeBool(v bool) {
	e.scratch[0] = 0
	if v {
		e.scratch[0] = 1
	}
	e.Write(e.scratch[:1])
}

// WriteString writes a string prefixed with a variable-size integer size.
func (e *Encoder) WriteString(v string) {
	e.WriteUvarint(uint64(len(v)))
	e.Write(ToBytes(v))
}

// scanToCache scans the type and caches it in the TinyBin instance
func (e *Encoder) scanToCache(t *tinyreflect.Type) (Codec, error) {
	if e.tb == nil {
		return nil, Err("Encoder", "scanToCache", "TinyBin", "nil")
	}

	// Use the TinyBin instance's schema caching mechanism
	return e.tb.scanToCache(t)
}
