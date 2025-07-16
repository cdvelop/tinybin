package tinybin

import (
	"encoding/binary"
	"io"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

// Encode encodes a slice of structs to an io.Writer.
func (h *TinyBin) Encode(w io.Writer, data any) error {
	bytes, _, err := h.EncodeToBytes(data)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}

// EncodeToBytes encodes a struct or slice of structs to a byte slice and returns the type ID.
func (h *TinyBin) EncodeToBytes(data any) ([]byte, uint32, error) {
	v := tinyreflect.ValueOf(data)

	// Get the actual struct type and determine if it's a slice
	structType, count, err := h.analyzeDataType(v)
	if err != nil {
		return nil, 0, err
	}

	// Find the struct type ID using StructID
	structID := structType.StructID()
	var typeID uint32
	var found bool
	for _, obj := range h.stObjects {
		if obj.stID == structID {
			typeID = obj.stID
			found = true
			break
		}
	}

	if !found {
		return nil, 0, Err(D.Type, D.Not, D.Found)
	}

	// Create buffer with protocol header
	var buf []byte
	buf = append(buf, 1, 0) // Major=1, Minor=0

	// Type ID (4 bytes, little-endian)
	typeIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(typeIDBytes, typeID)
	buf = append(buf, typeIDBytes...)

	// Count (varint)
	countBytes := encodeVarint(count)
	buf = append(buf, countBytes...)

	// Encode the data recursively
	dataBytes, err := h.encodeValue(v)
	if err != nil {
		return nil, 0, err
	}
	buf = append(buf, dataBytes...)

	return buf, typeID, nil
}

// analyzeDataType determines the struct type and count from the input data
func (h *TinyBin) analyzeDataType(v tinyreflect.Value) (*tinyreflect.Type, uint32, error) {
	t := v.Type()

	if t.Kind() == K.Struct {
		return t, 1, nil
	}

	if t.Kind() == K.Slice {
		// Get slice length using tinyreflect
		length, err := v.Len()
		if err != nil {
			return nil, 0, Errf("getting slice length: %w", err)
		}

		if length == 0 {
			return nil, 0, Err(D.Slice, D.Empty)
		}

		// Get the first element to determine the element type
		firstElem, err := v.Index(0)
		if err != nil {
			return nil, 0, Errf("getting first slice element: %w", err)
		}

		elemType := firstElem.Type()
		if elemType.Kind() != K.Struct {
			return nil, 0, Err(D.Type, "slice element", D.Not, D.Struct)
		}

		return elemType, uint32(length), nil
	}

	return nil, 0, Err(D.Type, D.Not, D.Supported)
}

// encodeValue recursively encodes any value (struct, slice, or field)
func (h *TinyBin) encodeValue(v tinyreflect.Value) ([]byte, error) {
	t := v.Type()

	switch t.Kind() {
	case K.Struct:
		return h.encodeStruct(v)
	case K.Slice:
		return h.encodeSlice(v)
	default:
		return h.encodeField(v)
	}
}

// encodeSlice recursively encodes a slice by encoding each element
func (h *TinyBin) encodeSlice(v tinyreflect.Value) ([]byte, error) {
	var buf []byte

	// Get slice length using the new Len() method
	length, err := v.Len()
	if err != nil {
		return nil, Errf("getting slice length: %w", err)
	}

	// Encode each element using the new Index() method
	for i := 0; i < length; i++ {
		elem, err := v.Index(i)
		if err != nil {
			return nil, Errf("getting slice element %d: %w", i, err)
		}

		elemBytes, err := h.encodeValue(elem)
		if err != nil {
			return nil, Errf("encoding slice element %d: %w", i, err)
		}
		buf = append(buf, elemBytes...)
	}

	return buf, nil
}

// encodeStruct encodes a single struct value
func (h *TinyBin) encodeStruct(v tinyreflect.Value) ([]byte, error) {
	var buf []byte

	numFields, err := v.NumField()
	if err != nil {
		return nil, err
	}

	for i := 0; i < numFields; i++ {
		field, err := v.Field(i)
		if err != nil {
			return nil, err
		}

		fieldBytes, err := h.encodeField(field)
		if err != nil {
			return nil, err
		}
		buf = append(buf, fieldBytes...)
	}

	return buf, nil
}

// encodeField encodes a single field value using Interface() and type assertions
func (h *TinyBin) encodeField(v tinyreflect.Value) ([]byte, error) {
	var buf []byte

	// Get the actual value using Interface()
	value, err := v.Interface()
	if err != nil {
		return nil, err
	}

	// Type assert based on the actual type
	switch val := value.(type) {
	case bool:
		if val {
			buf = append(buf, 1)
		} else {
			buf = append(buf, 0)
		}

	case int8:
		buf = append(buf, byte(val))

	case int16:
		bytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(bytes, uint16(val))
		buf = append(buf, bytes...)

	case int32:
		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, uint32(val))
		buf = append(buf, bytes...)

	case int64:
		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, uint64(val))
		buf = append(buf, bytes...)

	case uint8:
		buf = append(buf, byte(val))

	case uint16:
		bytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(bytes, val)
		buf = append(buf, bytes...)

	case uint32:
		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, val)
		buf = append(buf, bytes...)

	case uint64:
		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, val)
		buf = append(buf, bytes...)

	case float32:
		bytes := make([]byte, 4)
		bits := uint32(val)
		binary.LittleEndian.PutUint32(bytes, bits)
		buf = append(buf, bytes...)

	case float64:
		bytes := make([]byte, 8)
		bits := uint64(val)
		binary.LittleEndian.PutUint64(bytes, bits)
		buf = append(buf, bytes...)

	case string:
		lenBytes := encodeVarint(uint32(len(val)))
		buf = append(buf, lenBytes...)
		buf = append(buf, []byte(val)...)

	default:
		// For complex types (struct, slice), use recursive encoding
		t := v.Type()
		if t.Kind() == K.Struct {
			return h.encodeStruct(v)
		} else if t.Kind() == K.Slice {
			return h.encodeSlice(v)
		} else {
			return nil, Errf("unsupported type: %T", val)
		}
	}

	return buf, nil
}
