package tinybin

import (
	"reflect"
	"sync"

	. "github.com/cdvelop/tinystring"
)

// Map of all the schemas we've encountered so far
var schemas = new(sync.Map)

// scanToCache scans the type and caches in the local cache.
func scanToCache(t reflect.Type, cache map[reflect.Type]Codec) (Codec, error) {
	if c, ok := cache[t]; ok {
		return c, nil
	}

	c, err := scan(t)
	if err != nil {
		return nil, err
	}

	cache[t] = c
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

// ScanType scans the type
func scanType(t reflect.Type) (Codec, error) {
	switch t.Kind() {
	case reflect.Ptr:
		elemCodec, err := scanType(t.Elem())
		if err != nil {
			return nil, err
		}

		return &reflectPointerCodec{
			elemCodec: elemCodec,
		}, nil

	case reflect.Array:
		elemCodec, err := scanType(t.Elem())
		if err != nil {
			return nil, err
		}

		return &reflectArrayCodec{
			elemCodec: elemCodec,
		}, nil

	case reflect.Slice:

		// Fast-paths for simple numeric slices and string slices
		switch t.Elem().Kind() {
		case reflect.Uint8:
			return new(byteSliceCodec), nil
		case reflect.Bool:
			return new(boolSliceCodec), nil
		case reflect.Uint:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			return new(varuintSliceCodec), nil
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			return new(varintSliceCodec), nil
		case reflect.Ptr:
			elemCodec, err := scanType(t.Elem().Elem())
			if err != nil {
				return nil, err
			}

			return &reflectSliceOfPtrCodec{
				elemType:  t.Elem().Elem(),
				elemCodec: elemCodec,
			}, nil
		default:
			elemCodec, err := scanType(t.Elem())
			if err != nil {
				return nil, err
			}

			return &reflectSliceCodec{
				elemCodec: elemCodec,
			}, nil
		}

	case reflect.Struct:
		s := scanStruct(t)
		v := make(reflectStructCodec, 0, len(s.fields))
		for _, i := range s.fields {
			field := t.Field(i)
			codec, err := scanType(field.Type)
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

	case reflect.String:
		return new(stringCodec), nil
	case reflect.Bool:
		return new(boolCodec), nil
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int64:
		return new(varintCodec), nil
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint64:
		return new(varuintCodec), nil
	case reflect.Float32:
		return new(float32Codec), nil
	case reflect.Float64:
		return new(float64Codec), nil
	}

	return nil, Err(D.Type, D.Binary, t.String(), D.Not, D.Supported)
}

type scannedStruct struct {
	fields []int
}

func scanStruct(t reflect.Type) (meta *scannedStruct) {
	l := t.NumField()
	meta = new(scannedStruct)
	for i := 0; i < l; i++ {
		if t.Field(i).Name != "_" {
			if t.Field(i).Tag.Get("binary") != "-" {
				meta.fields = append(meta.fields, i)
			}
		}
	}
	return
}
