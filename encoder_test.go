package tinybin

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"unsafe"
)

var testMsg = msg{
	Name:      "Roman",
	Timestamp: 1242345235,
	Payload:   []byte("hi"),
	Ssid:      []uint32{1, 2, 3},
}

type composite map[string]column

type column struct {
	Varchar columnVarchar
	Float64 columnFloat64
	Float32 columnFloat32
}

type columnVarchar struct {
	Nulls []bool
	Sizes []uint32
	Bytes []byte
}

type columnFloat64 struct {
	Nulls  []bool
	Floats []float64
}

type columnFloat32 struct {
	Nulls  []bool
	Floats []float32
}

// Test_Full removed - uses map[string]column which is not supported
// Maps are intentionally not supported in TinyBin for WebAssembly optimization
// Use slice of structs instead: []struct{Key string; Value column}
/*
func Test_Full(t *testing.T) {
	v := composite{}
	v["a"] = column{
		Varchar: columnVarchar{
			Nulls: []bool{false, false, false, true, false},
			Sizes: []uint32{2, 2, 2, 0, 2},
			Bytes: []byte{10, 10, 10, 10, 10, 10, 10, 10},
		},
	}
	v["b"] = column{
		Float64: columnFloat64{
			Nulls:  []bool{false, false, false, true, false},
			Floats: []float64{1.1, 2.2, 3.3, 0, 4.4},
		},
	}

	b, err := Marshal(&v)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if b == nil {
		t.Error("Expected non-nil bytes")
	}

	var o composite
	err = Unmarshal(b, &o)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(v, o) {
		t.Errorf("Expected %v, got %v", v, o)
	}
}
*/

func newComposite() composite {
	v := composite{}
	v["a"] = column{
		Varchar: columnVarchar{
			Nulls: []bool{false, false, false, true, false, false, false, false, true, false, false, false, false, true, false},
			Sizes: []uint32{2, 2, 2, 0, 2, 2, 2, 2, 0, 2, 2, 2, 2, 0, 2, 2, 2, 2, 0, 2, 2, 2, 2, 0, 2, 2, 2, 2, 0, 2},
			Bytes: []byte{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		},
	}
	v["b"] = column{
		Float64: columnFloat64{
			Nulls:  []bool{false, false, false, true, false},
			Floats: []float64{1.1, 2.2, 3.3, 0, 4.4},
		},
	}
	return v
}

/*
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
Benchmark_Binary/marshal-12         	 5074890	       227.4 ns/op	     112 B/op	       2 allocs/op
Benchmark_Binary/marshal-to-12      	 7011523	       162.3 ns/op	      30 B/op	       0 allocs/op
Benchmark_Binary/unmarshal-12       	 4224048	       283.0 ns/op	      72 B/op	       5 allocs/op
*/
func BenchmarkBinary(b *testing.B) {
	v := testMsg
	enc, _ := Marshal(&v)

	b.Run("marshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			Marshal(&v)
		}
	})

	var buffer bytes.Buffer
	b.Run("marshal-to", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			buffer.Reset()
			MarshalTo(&v, &buffer)
		}
	})

	b.Run("unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		var out msg
		for n := 0; n < b.N; n++ {
			Unmarshal(enc, &out)
		}
	})
}

func BenchmarkJSON(b *testing.B) {
	v := testMsg
	enc, _ := json.Marshal(&v)

	b.Run("marshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			json.Marshal(&v)
		}
	})

	b.Run("unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		var out msg
		for n := 0; n < b.N; n++ {
			json.Unmarshal(enc, &out)
		}
	})
}

func TestBinaryEncodeStruct(t *testing.T) {
	b, err := Marshal(s0v)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !bytes.Equal(s0b, b) {
		t.Errorf("Expected %v, got %v", s0b, b)
	}
}

func TestEncoderSizeOf(t *testing.T) {
	var e Encoder
	size := int(unsafe.Sizeof(e))
	if size != 56 {
		t.Errorf("Expected %v, got %v", 56, size)
	}
}

func TestMarshalWithCustomCodec(t *testing.T) {
	v := testCustom("custom codec")

	b, err := Marshal(v)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if b == nil {
		t.Error("Expected non-nil bytes")
	}

	var out testCustom
	err = Unmarshal(b, &out)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(v, out) {
		t.Errorf("Expected %v, got %v", v, out)
	}
}
