package tinybin

import (
	"encoding/binary"
	"unsafe"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

// Constants
var (
	LittleEndian = binary.LittleEndian
	BigEndian    = binary.BigEndian
)

// sliceHeader is the runtime representation of a slice.
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// Codec represents a single part Codec, which can encode and decode something.
type Codec interface {
	EncodeTo(*Encoder, tinyreflect.Value) error
	DecodeTo(*Decoder, tinyreflect.Value) error
}

// ------------------------------------------------------------------------------

type reflectArrayCodec struct {
	elemCodec Codec // The codec of the array's elements
}

// Encode encodes a value into the encoder.
func (c *reflectArrayCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	l, err := rv.Len()
	if err != nil {
		return err
	}
	for i := 0; i < l; i++ {
		idx, err := rv.Index(i)
		if err != nil {
			return err
		}
		// Use the element directly without Addr() - it should already be the right type
		if err = c.elemCodec.EncodeTo(e, idx); err != nil {
			return err
		}
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectArrayCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	l, err := rv.Len()
	if err != nil {
		return err
	}
	for i := 0; i < l; i++ {
		idx, err := rv.Index(i)
		if err != nil {
			return err
		}
		// Don't use Indirect here - use the indexed value directly
		if err = c.elemCodec.DecodeTo(d, idx); err != nil {
			return err
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type reflectSliceCodec struct {
	elemCodec Codec // The codec of the slice's elements
}

// Encode encodes a value into the encoder.
func (c *reflectSliceCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	l, err := rv.Len()
	if err != nil {
		return err
	}
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		idx, err := rv.Index(i)
		if err != nil {
			return err
		}

		// Try using the element directly without Addr() - it should already be the right type
		if err = c.elemCodec.EncodeTo(e, idx); err != nil {
			return err
		}
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectSliceCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		// Debug: log the length read
		// fmt.Printf("DEBUG: reflectSliceCodec reading length: %d\n", l)

		typ := rv.Type()
		newSlice, err := tinyreflect.MakeSlice(typ, int(l), int(l))
		if err != nil {
			return err
		}

		// Debug: check slice was created correctly
		// sliceLen, _ := newSlice.Len()
		// fmt.Printf("DEBUG: Created slice with length: %d\n", sliceLen)

		if err = rv.Set(newSlice); err != nil {
			return err
		}

		// Debug: verify rv now contains the slice
		// rvLen, _ := rv.Len()
		// fmt.Printf("DEBUG: rv after Set has length: %d\n", rvLen)

		for i := 0; i < int(l); i++ {
			idx, err := rv.Index(i)
			if err != nil {
				return err
			}
			v := tinyreflect.Indirect(idx)
			if err = c.elemCodec.DecodeTo(d, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type reflectSliceOfPtrCodec struct {
	elemCodec Codec             // The codec of the slice's elements
	elemType  *tinyreflect.Type // The type of the element
}

// Encode encodes a value into the encoder.
func (c *reflectSliceOfPtrCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	l, err := rv.Len()
	if err != nil {
		return err
	}
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		v, err := rv.Index(i)
		if err != nil {
			return err
		}
		isNil, err := v.IsNil()
		if err != nil {
			return err
		}
		e.writeBool(isNil)
		if !isNil {
			indirect := tinyreflect.Indirect(v)
			if err = c.elemCodec.EncodeTo(e, indirect); err != nil {
				return err
			}
		}
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectSliceOfPtrCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var l uint64
	var isNil bool
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		typ := rv.Type()
		newSlice, err := tinyreflect.MakeSlice(typ, int(l), int(l))
		if err != nil {
			return err
		}
		if err = rv.Set(newSlice); err != nil {
			return err
		}
		for i := 0; i < int(l); i++ {
			if isNil, err = d.ReadBool(); !isNil {
				if err != nil {
					return err
				}

				ptr, err := rv.Index(i)
				if err != nil {
					return err
				}
				// Create new pointer value and decode directly to it
				newPtr := tinyreflect.NewValue(c.elemType)
				indirect := tinyreflect.Indirect(newPtr)
				if err = c.elemCodec.DecodeTo(d, indirect); err != nil {
					return err
				}
				// Now copy the decoded value to the slice element
				if err = ptr.Set(newPtr); err != nil {
					// If Set fails due to type incompatibility, try direct assignment
					elem, elemErr := ptr.Elem()
					if elemErr == nil {
						decodedElem := tinyreflect.Indirect(newPtr)
						if copyErr := elem.Set(decodedElem); copyErr != nil {
							return err // Return original Set error
						}
					} else {
						return err // Return original Set error
					}
				}
			}
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type byteSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *byteSliceCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	// Get the slice length
	l, err := rv.Len()
	if err != nil {
		return err
	}

	e.WriteUvarint(uint64(l))
	if l > 0 {
		// Read each byte from the slice
		data := make([]byte, l)
		for i := 0; i < l; i++ {
			idx, err := rv.Index(i)
			if err != nil {
				return err
			}
			uintVal, err := idx.Uint()
			if err != nil {
				return err
			}
			data[i] = byte(uintVal)
		}
		e.Write(data)
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *byteSliceCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		data := make([]byte, int(l))
		if _, err = d.Read(data); err == nil {
			// Create a new slice with the correct type
			typ := rv.Type()
			newSlice, err := tinyreflect.MakeSlice(typ, int(l), int(l))
			if err != nil {
				return err
			}

			// Set each byte in the slice
			for i := 0; i < int(l); i++ {
				idx, err := newSlice.Index(i)
				if err != nil {
					return err
				}
				if err = idx.SetUint(uint64(data[i])); err != nil {
					return err
				}
			}

			if err = rv.Set(newSlice); err != nil {
				return err
			}
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type boolSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolSliceCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	l, err := rv.Len()
	if err != nil {
		return err
	}
	e.WriteUvarint(uint64(l))
	if l > 0 {
		// TODO: Need to implement proper interface access for []bool
		// For now, this is a placeholder
		dummy := make([]byte, l)
		e.Write(dummy)
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *boolSliceCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		buf := make([]byte, l)
		_, err = d.Read(buf)
		if err != nil {
			return err
		}
		// TODO: Need to implement proper bool slice creation
		// For now, create empty slice
		bools := make([]bool, l)
		boolsValue := tinyreflect.ValueOf(bools)
		if err = rv.Set(boolsValue); err != nil {
			return err
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type varintSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintSliceCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	l, err := rv.Len()
	if err != nil {
		return err
	}
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		idx, err := rv.Index(i)
		if err != nil {
			return err
		}
		intVal, err := idx.Int()
		if err != nil {
			return err
		}
		e.WriteVarint(intVal)
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *varintSliceCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var l uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		typ := rv.Type()
		newSlice, err := tinyreflect.MakeSlice(typ, int(l), int(l))
		if err != nil {
			return err
		}
		if err = rv.Set(newSlice); err != nil {
			return err
		}
		for i := 0; i < int(l); i++ {
			var v int64
			if v, err = d.ReadVarint(); err == nil {
				idx, err := rv.Index(i)
				if err != nil {
					return err
				}
				if err = idx.SetInt(v); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type varuintSliceCodec struct{}

// Encode encodes a value into the encoder.
func (c *varuintSliceCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	l, err := rv.Len()
	if err != nil {
		return err
	}
	e.WriteUvarint(uint64(l))
	for i := 0; i < l; i++ {
		idx, err := rv.Index(i)
		if err != nil {
			return err
		}
		uintVal, err := idx.Uint()
		if err != nil {
			return err
		}
		e.WriteUvarint(uintVal)
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *varuintSliceCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var l, v uint64
	if l, err = d.ReadUvarint(); err == nil && l > 0 {
		typ := rv.Type()
		newSlice, err := tinyreflect.MakeSlice(typ, int(l), int(l))
		if err != nil {
			return err
		}
		if err = rv.Set(newSlice); err != nil {
			return err
		}
		for i := 0; i < int(l); i++ {
			if v, err = d.ReadUvarint(); err == nil {
				idx, err := rv.Index(i)
				if err != nil {
					return err
				}
				if err = idx.SetUint(v); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type reflectPointerCodec struct {
	elemCodec Codec
}

// Encode encodes a value into the encoder.
func (c *reflectPointerCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	isNil, err := rv.IsNil()
	if err != nil {
		return err
	}
	if isNil {
		e.writeBool(true)
		return nil
	}

	e.writeBool(false)
	elem, err := rv.Elem()
	if err != nil {
		return err
	}
	err = c.elemCodec.EncodeTo(e, elem)
	if err != nil {
		return err
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *reflectPointerCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	isNil, err := d.ReadBool()
	if err != nil {
		return err
	}
	if isNil {
		return nil
	}

	// Check if the pointer is nil and create a new value if needed
	rvIsNil, err := rv.IsNil()
	if err != nil {
		return err
	}
	if rvIsNil {
		typ := rv.Type()
		// Get the element type using the Type.Elem() method
		elemType := typ.Elem()
		if elemType == nil {
			return Err(D.Binary, "pointer", D.Type, D.Nil)
		}
		newPtr := tinyreflect.NewValue(elemType)
		if err = rv.Set(newPtr); err != nil {
			// DEBUG: This is where the error is happening
			return Err("DEBUG", "reflectPointerCodec", "Set", "failed", err.Error())
		}
	}

	elem, err := rv.Elem()
	if err != nil {
		return err
	}
	return c.elemCodec.DecodeTo(d, elem)
}

// ------------------------------------------------------------------------------

type reflectStructCodec []fieldCodec

type fieldCodec struct {
	Index int   // The index of the field
	Codec Codec // The codec to use for this field
}

// Encode encodes a value into the encoder.
func (c reflectStructCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) (err error) {
	for _, i := range c {
		field, err := rv.Field(i.Index)
		if err != nil {
			return err
		}
		if err = i.Codec.EncodeTo(e, field); err != nil {
			return err
		}
	}
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c reflectStructCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	for _, fieldCodec := range c {
		v, err := rv.Field(fieldCodec.Index)
		if err != nil {
			return err
		}

		// Debug: Check if codec is nil
		if fieldCodec.Codec == nil {
			return Err(D.Field, fieldCodec.Index, "codec", D.Nil)
		}

		// Follow the original logic: handle pointers vs regular fields differently
		switch v.Kind() {
		case K.Pointer:
			// For pointer fields, pass the value directly to the codec
			err = fieldCodec.Codec.DecodeTo(d, v)
		default:
			// For non-pointer fields that can be set, use Indirect
			// TODO: Implement CanSet() check when available
			indirect := tinyreflect.Indirect(v)
			err = fieldCodec.Codec.DecodeTo(d, indirect)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type stringCodec struct{}

// Encode encodes a value into the encoder.
func (c *stringCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) error {
	s := rv.String()
	e.WriteString(s)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *stringCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var s string
	if s, err = d.ReadString(); err == nil {
		if err = rv.SetString(s); err != nil {
			return err
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type boolCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) error {
	boolVal, err := rv.Bool()
	if err != nil {
		return err
	}
	e.writeBool(boolVal)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *boolCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var out bool
	if out, err = d.ReadBool(); err == nil {
		if err = rv.SetBool(out); err != nil {
			return err
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type varintCodec struct{}

// Encode encodes a value into the encoder.
func (c *varintCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) error {
	intVal, err := rv.Int()
	if err != nil {
		return err
	}
	e.WriteVarint(intVal)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *varintCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var v int64
	if v, err = d.ReadVarint(); err != nil {
		return err
	}
	if err = rv.SetInt(v); err != nil {
		return err
	}
	return nil
}

// ------------------------------------------------------------------------------

type varuintCodec struct{}

// Encode encodes a value into the encoder.
func (c *varuintCodec) EncodeTo(e *Encoder, rv tinyreflect.Value) error {
	uintVal, err := rv.Uint()
	if err != nil {
		return err
	}
	e.WriteUvarint(uintVal)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *varuintCodec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var v uint64
	if v, err = d.ReadUvarint(); err != nil {
		return err
	}
	if err = rv.SetUint(v); err != nil {
		return err
	}
	return nil
}

// ------------------------------------------------------------------------------

type float32Codec struct{}

// Encode encodes a value into the encoder.
func (c *float32Codec) EncodeTo(e *Encoder, rv tinyreflect.Value) error {
	floatVal, err := rv.Float()
	if err != nil {
		return err
	}
	e.WriteFloat32(float32(floatVal))
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float32Codec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var v float32
	if v, err = d.ReadFloat32(); err == nil {
		if err = rv.SetFloat(float64(v)); err != nil {
			return err
		}
	}
	return nil
}

// ------------------------------------------------------------------------------

type float64Codec struct{}

// Encode encodes a value into the encoder.
func (c *float64Codec) EncodeTo(e *Encoder, rv tinyreflect.Value) error {
	floatVal, err := rv.Float()
	if err != nil {
		return err
	}
	e.WriteFloat64(floatVal)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float64Codec) DecodeTo(d *Decoder, rv tinyreflect.Value) (err error) {
	var v float64
	if v, err = d.ReadFloat64(); err == nil {
		if err = rv.SetFloat(v); err != nil {
			return err
		}
	}
	return nil
}
