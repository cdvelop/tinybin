package tinybin

import (
	"bytes"
	"io"
	"sync"

	"github.com/cdvelop/tinyreflect"
)

// TinyBin es la estructura principal que encapsula toda la funcionalidad
// de serialización binaria, eliminando las variables globales públicas.
type TinyBin struct {
	// Pools para reutilización de encoders y decoders
	encoders *sync.Pool
	decoders *sync.Pool

	// Cache local de schemas para cada instancia
	schemas *sync.Map
}

// New crea una nueva instancia de TinyBin con toda la funcionalidad
// encapsulada, eliminando las dependencias de variables globales.
func New() *TinyBin {
	return &TinyBin{
		encoders: &sync.Pool{
			New: func() any {
				return &Encoder{
					schemas: make(map[*tinyreflect.Type]Codec),
					tinyBin: nil, // Will be set when retrieved from pool
				}
			},
		},
		decoders: &sync.Pool{
			New: func() any {
				return &Decoder{
					reader:  newReader(nil),
					schemas: make(map[*tinyreflect.Type]Codec),
					tinyBin: nil, // Will be set when retrieved from pool
				}
			},
		},
		schemas: new(sync.Map),
	}
}

// Encode serializa un valor a formato binario usando la instancia de TinyBin.
func (tb *TinyBin) Encode(v any) (output []byte, err error) {
	var buffer bytes.Buffer
	buffer.Grow(64)

	// Encode and set the buffer if successful
	if err = tb.encodeTo(v, &buffer); err == nil {
		output = buffer.Bytes()
	}
	return
}

// encodeTo serializa un valor a un destino específico.
func (tb *TinyBin) encodeTo(v any, dst io.Writer) (err error) {
	// Get the encoder from the pool, reset it
	e := tb.encoders.Get().(*Encoder)
	e.tinyBin = tb // Set the TinyBin reference
	e.Reset(dst)

	// Encode and set the buffer if successful
	err = e.Encode(v)

	// Put the encoder back when we're finished
	tb.encoders.Put(e)
	return
}

// Decode deserializa datos binarios a un valor usando la instancia de TinyBin.
func (tb *TinyBin) Decode(b []byte, v any) (err error) {
	// Get the decoder from the pool, reset it
	d := tb.decoders.Get().(*Decoder)
	d.tinyBin = tb                   // Set the TinyBin reference
	d.reader.(*sliceReader).Reset(b) // Reset the reader

	// Decode and set the buffer if successful and free the decoder
	err = d.Decode(v)
	tb.decoders.Put(d)
	return
}

// NewEncoder crea un nuevo encoder usando la instancia de TinyBin.
func (tb *TinyBin) NewEncoder(out io.Writer) *Encoder {
	return &Encoder{
		out:     out,
		schemas: make(map[*tinyreflect.Type]Codec),
		tinyBin: tb,
	}
}

// NewDecoder crea un nuevo decoder usando la instancia de TinyBin.
func (tb *TinyBin) NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader:  newReader(r),
		schemas: make(map[*tinyreflect.Type]Codec),
		tinyBin: tb,
	}
}

// scanToCache escanea el tipo y lo cachea en la instancia local.
func (tb *TinyBin) scanToCache(t *tinyreflect.Type, cache map[*tinyreflect.Type]Codec) (Codec, error) {
	if c, ok := cache[t]; ok {
		return c, nil
	}

	c, err := tb.scan(t)
	if err != nil {
		return nil, err
	}

	cache[t] = c
	return c, nil
}

// scan obtiene un codec para el tipo usando el cache de la instancia.
func (tb *TinyBin) scan(t *tinyreflect.Type) (c Codec, err error) {
	// Attempt to load from instance cache first
	if f, ok := tb.schemas.Load(t); ok {
		c = f.(Codec)
		return
	}

	// Scan for the first time
	c, err = ScanType(t)
	if err != nil {
		return
	}

	// Load or store again
	if f, ok := tb.schemas.LoadOrStore(t, c); ok {
		c = f.(Codec)
		return
	}
	return
}
