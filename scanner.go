package tinybin

import (
	"reflect"
	"sync"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

// Map of all the schemas we've encountered so far
var schemas = new(sync.Map)

// scanToCache scans the type and caches in the local cache.
func scanToCache(t reflect.Type, cache map[reflect.Type]Codec) (Codec, error) {
	if c, ok := cache[t]; ok {
		return c, nil
	}

	// Convert reflect.Type to tinyreflect.Type
	tinyT := convertToTinyReflectType(t)
	c, err := scanWithTinyReflect(tinyT)
	if err != nil {
		return nil, err
	}

	cache[t] = c
	return c, nil
}

// convertToTinyReflectType converts a reflect.Type to tinyreflect.Type
func convertToTinyReflectType(t reflect.Type) *tinyreflect.Type {
	// Create a zero value of the type and get its tinyreflect.Type
	zeroVal := reflect.Zero(t).Interface()
	return tinyreflect.TypeOf(zeroVal)
}

// getElementType returns the element type for pointer, array, and slice types
// This is a helper function to work with tinyreflect's limited API
func getElementType(t *tinyreflect.Type) (*tinyreflect.Type, error) {
	kind := t.Kind()
	switch kind {
	case K.Pointer:
		ptrType := t.PtrType()
		if ptrType == nil {
			return nil, Err(D.Type, t.String(), D.Not, "a pointer type")
		}
		return ptrType.Elem, nil
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
// This is needed for compatibility with existing codecs
func convertTinyReflectToReflectType(t *tinyreflect.Type) reflect.Type {
	// This is a temporary solution - we need to find a way to convert back
	// For now, we'll create a value and get its reflect.Type
	// Note: This is not ideal but necessary for gradual migration
	
	// Create a zero value using the type information
	switch t.Kind() {
	case K.Bool:
		return reflect.TypeOf(false)
	case K.Int:
		return reflect.TypeOf(int(0))
	case K.Int8:
		return reflect.TypeOf(int8(0))
	case K.Int16:
		return reflect.TypeOf(int16(0))
	case K.Int32:
		return reflect.TypeOf(int32(0))
	case K.Int64:
		return reflect.TypeOf(int64(0))
	case K.Uint:
		return reflect.TypeOf(uint(0))
	case K.Uint8:
		return reflect.TypeOf(uint8(0))
	case K.Uint16:
		return reflect.TypeOf(uint16(0))
	case K.Uint32:
		return reflect.TypeOf(uint32(0))
	case K.Uint64:
		return reflect.TypeOf(uint64(0))
	case K.Float32:
		return reflect.TypeOf(float32(0))
	case K.Float64:
		return reflect.TypeOf(float64(0))
	case K.String:
		return reflect.TypeOf("")
	default:
		// For complex types, we might need to create them differently
		// This is a placeholder - in a full migration, we'd eliminate this function
		return reflect.TypeOf((*interface{})(nil)).Elem()
	}
}

// scanWithTinyReflect scans using tinyreflect.Type (new implementation)
func scanWithTinyReflect(t *tinyreflect.Type) (Codec, error) {
	// Attempt to load from cache first using a string key
	key := t.String()
	if f, ok := schemas.Load(key); ok {
		c := f.(Codec)
		return c, nil
	}

	// Scan for the first time
	c, err := scanTypeWithTinyReflect(t)
	if err != nil {
		return nil, err
	}

	// Load or store again
	if f, ok := schemas.LoadOrStore(key, c); ok {
		c = f.(Codec)
		return c, nil
	}
	return c, nil
}

// Scan gets a codec for the type and uses a cached schema if the type was
// previously scanned.
func scan(t reflect.Type) (c Codec, err error) {

	// Attempt to load from cache first
	if f, ok := schemas.Load(t); ok {
		c = f.(Codec)
		return
	}

	// Scan for the first time
	c, err = scanType(t)
	if err != nil {
		return
	}

	// Load or store again
	if f, ok := schemas.LoadOrStore(t, c); ok {
		c = f.(Codec)
		return
	}
	return
}

// ScanType scans the type (legacy function using reflect.Type)
func scanType(t reflect.Type) (Codec, error) {
	tinyT := convertToTinyReflectType(t)
	return scanTypeWithTinyReflect(tinyT)
}

// scanTypeWithTinyReflect scans the type using tinyreflect.Type
func scanTypeWithTinyReflect(t *tinyreflect.Type) (Codec, error) {
	kind := t.Kind()
	switch kind {
	case K.Pointer:
		elem, err := getElementType(t)
		if err != nil {
			return nil, err
		}
		elemCodec, err := scanTypeWithTinyReflect(elem)
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
		elemCodec, err := scanTypeWithTinyReflect(elem)
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
			elemCodec, err := scanTypeWithTinyReflect(elemElem)
			if err != nil {
				return nil, err
			}

			return &reflectSliceOfPtrCodec{
				elemType:  convertTinyReflectToReflectType(elemElem), // Convert back for compatibility
				elemCodec: elemCodec,
			}, nil
		default:
			elemCodec, err := scanTypeWithTinyReflect(elem)
			if err != nil {
				return nil, err
			}

			return &reflectSliceCodec{
				elemCodec: elemCodec,
			}, nil
		}

	case K.Struct:
		s, err := scanStructWithTinyReflect(t)
		if err != nil {
			return nil, err
		}
		v := make(reflectStructCodec, 0, len(s.fields))
		for _, i := range s.fields {
			field, err := t.Field(i)
			if err != nil {
				return nil, err
			}
			codec, err := scanTypeWithTinyReflect(field.Typ)
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

type scannedStruct struct {
	fields []int
}

// scanStructWithTinyReflect scans a struct using tinyreflect.Type
func scanStructWithTinyReflect(t *tinyreflect.Type) (meta *scannedStruct, err error) {
	numFields, err := t.NumField()
	if err != nil {
		return nil, err
	}
	
	meta = new(scannedStruct)
	for i := 0; i < numFields; i++ {
		field, err := t.Field(i)
		if err != nil {
			return nil, err
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
	return meta, nil
}
