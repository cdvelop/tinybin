package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestGetElementTypeFunction(t *testing.T) {
	// Test the exact function that's failing in scanner.go
	type simpleStruct struct {
		Name      string
		Timestamp int64
		Payload   []byte
		Ssid      []uint32
	}

	s := &simpleStruct{}

	// This is exactly what decoder.go does
	rv := tinyreflect.Indirect(tinyreflect.ValueOf(s))
	typ := rv.Type()

	if typ == nil {
		t.Fatal("rv.Type() returned nil")
	}

	t.Logf("rv.Type() = %p, Kind: %v", typ, typ.Kind())

	// Test with original pointer type
	originalV := tinyreflect.ValueOf(s)
	originalTyp := originalV.Type()

	if originalTyp == nil {
		t.Fatal("ValueOf(s).Type() returned nil")
	}

	t.Logf("Original Type: %p, Kind: %v", originalTyp, originalTyp.Kind())

	// Test the actual getElementType function from scanner.go
	elem, err := getElementType(originalTyp)
	if err != nil {
		t.Fatalf("getElementType failed: %v", err)
	}

	if elem == nil {
		t.Fatal("getElementType returned nil element")
	}

	t.Logf("getElementType returned: %p, Kind: %v", elem, elem.Kind())
}
