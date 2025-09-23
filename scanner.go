package tinybin

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

// scanType scans the type
func scanType(t *tinyreflect.Type) (Codec, error) {
	if t == nil {
		return nil, Err(D.Value, D.Type, D.Nil)
	}

	// TODO: Implement custom codec scanning when needed
	// if custom, ok := scanCustomCodec(t); ok {
	//     return custom, nil
	// }

	// TODO: Implement binary marshaler scanning when needed
	// if custom, ok := scanBinaryMarshaler(t); ok {
	//     return custom, nil
	// }

	switch t.Kind() {
	case K.Pointer:
		elem, err := getElementType(t)
		if err != nil {
			return nil, err
		}
		elemCodec, err := scanType(elem)
		if err != nil {
			return nil, err
		}

		return &reflectPointerCodec{
			elemCodec: elemCodec,
		}, nil

	case K.Array:
		elem, err := getElementType(t)
		if err != nil {
			return nil, err
		}
		elemCodec, err := scanType(elem)
		if err != nil {
			return nil, err
		}

		return &reflectArrayCodec{
			elemCodec: elemCodec,
		}, nil

	case K.Slice:
		elem, err := getElementType(t)
		if err != nil {
			return nil, err
		}
		elemKind := elem.Kind()

		// Fast-paths for simple numeric slices and string slices
		switch elemKind {
		case K.Uint8:
			return new(byteSliceCodec), nil
		case K.Bool:
			return new(boolSliceCodec), nil
		case K.Uint, K.Uint16, K.Uint32, K.Uint64:
			return new(varuintSliceCodec), nil
		case K.Int, K.Int8, K.Int16, K.Int32, K.Int64:
			return new(varintSliceCodec), nil
		case K.Pointer:
			elemElem, err := getElementType(elem)
			if err != nil {
				return nil, err
			}
			elemCodec, err := scanType(elemElem)
			if err != nil {
				return nil, err
			}

			return &reflectSliceOfPtrCodec{
				elemType:  elemElem,
				elemCodec: elemCodec,
			}, nil
		default:
			elemCodec, err := scanType(elem)
			if err != nil {
				return nil, err
			}

			return &reflectSliceCodec{
				elemCodec: elemCodec,
			}, nil
		}

	case K.Struct:
		s := scanStruct(t)
		v := make(reflectStructCodec, 0, len(s.fields))
		for _, i := range s.fields {
			field, err := t.Field(i)
			if err != nil {
				return nil, err
			}
			if field.Typ == nil {
				// Debug: Print information about the field
				return nil, Err(D.Field, D.Type, D.Nil, "field", Convert(i).String(), "name", field.Name)
			}
			codec, err := scanType(field.Typ)
			if err != nil {
				return nil, err
			}

			// Append since unexported fields are skipped
			v = append(v, fieldCodec{
				Index: i,
				Codec: codec,
			})
		}

		return &v, nil

	case K.String:
		return new(stringCodec), nil
	case K.Bool:
		return new(boolCodec), nil
	case K.Int8, K.Int16, K.Int32, K.Int, K.Int64:
		return new(varintCodec), nil
	case K.Uint8, K.Uint16, K.Uint32, K.Uint, K.Uint64:
		return new(varuintCodec), nil
	case K.Float32:
		return new(float32Codec), nil
	case K.Float64:
		return new(float64Codec), nil
	}

	return nil, Err(D.Type, D.Binary, t.String(), D.Not, D.Supported)
}

// getElementType returns the element type for pointer, array, and slice types
func getElementType(t *tinyreflect.Type) (*tinyreflect.Type, error) {
	kind := t.Kind()
	switch kind {
	case K.Pointer:
		// Use the Type.Elem() method instead of accessing PtrType.Elem directly
		elemType := t.Elem()
		if elemType == nil {
			return nil, Err(D.Type, t.String(), D.Not, "a valid pointer type")
		}
		return elemType, nil
	case K.Array:
		arrayType := t.ArrayType()
		if arrayType == nil {
			return nil, Err(D.Type, t.String(), D.Not, "an array type")
		}
		return arrayType.Element(), nil
	case K.Slice:
		sliceType := t.SliceType()
		if sliceType == nil {
			return nil, Err(D.Type, t.String(), D.Not, "a slice type")
		}
		return sliceType.Element(), nil
	default:
		return nil, Err(D.Type, t.String(), D.Not, "have", "element type")
	}
}

type scannedStruct struct {
	fields []int
}

// scanStruct scans a struct using tinyreflect.Type
func scanStruct(t *tinyreflect.Type) *scannedStruct {
	numFields, err := t.NumField()
	if err != nil {
		return &scannedStruct{fields: []int{}}
	}

	meta := &scannedStruct{fields: make([]int, 0, numFields)}
	for i := 0; i < numFields; i++ {
		field, err := t.Field(i)
		if err != nil {
			continue
		}

		// Get field name
		fieldName := field.Name.Name()
		if fieldName != "_" {
			// Check if field should be skipped
			tag := field.Tag()
			if tag.Get("binary") != "-" {
				meta.fields = append(meta.fields, i)
			}
		}
	}
	return meta
}
