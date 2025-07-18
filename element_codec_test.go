package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

// TestElementCodecFlow tests the element codec flow for slice elements
// This test helps ensure proper codec detection and encoding for slice elements
func TestElementCodecFlow(t *testing.T) {
	// Test slice of structs
	input := []simpleStruct{{
		Name:      "Roman",
		Timestamp: 1357092245000000006,
		Payload:   []byte("hi"),
		Ssid:      []uint32{1, 2, 3},
	}}

	// Step 1: Get the slice
	rv := tinyreflect.Indirect(tinyreflect.ValueOf(&input))
	t.Logf("Slice type: %p, kind: %s, length: %d", rv.Type(), rv.Type().Kind().String(), mustGetLen(rv))

	// Step 2: Get the slice codec
	sliceCodec, err := scanType(rv.Type())
	if err != nil {
		t.Fatalf("Failed to scan slice type: %v", err)
	}
	
	reflectSliceCodec, ok := sliceCodec.(*reflectSliceCodec)
	if !ok {
		t.Fatalf("Expected reflectSliceCodec, got %T", sliceCodec)
	}
	t.Logf("Slice codec: %T", reflectSliceCodec)

	// Step 3: Test the element codec
	elemCodec := reflectSliceCodec.elemCodec
	t.Logf("Element codec: %T", elemCodec)

	// Step 4: Get the first element and test it
	if length, _ := rv.Len(); length > 0 {
		elem, err := rv.Index(0)
		if err != nil {
			t.Fatalf("Failed to index element 0: %v", err)
		}
		
		t.Logf("Element type: %p, kind: %s", elem.Type(), elem.Type().Kind().String())
		
		// Verify element type matches what elemCodec expects
		if elem.Type().Kind().String() != "struct" {
			t.Errorf("Expected struct element, got %s", elem.Type().Kind().String())
		}

		// Test that we can scan the element type independently
		elemTypeCodec, err := scanType(elem.Type())
		if err != nil {
			t.Fatalf("Failed to scan element type: %v", err)
		}
		t.Logf("Element type codec: %T", elemTypeCodec)
		
		// Verify codec types match
		if elemCodecType := getCodecTypeName(elemCodec); elemCodecType != getCodecTypeName(elemTypeCodec) {
			t.Errorf("Codec type mismatch: slice elemCodec=%s, direct scan=%s", 
				elemCodecType, getCodecTypeName(elemTypeCodec))
		}
	}
}

// mustGetLen is a helper to get length without error handling for cleaner logs
func mustGetLen(v tinyreflect.Value) int {
	if length, err := v.Len(); err == nil {
		return length
	}
	return -1
}

// getCodecTypeName returns the type name of a codec for comparison
func getCodecTypeName(c Codec) string {
	switch c.(type) {
	case *reflectStructCodec:
		return "struct"
	case *reflectSliceCodec:
		return "slice"
	case *reflectPointerCodec:
		return "pointer"
	case *stringCodec:
		return "string"
	case *varintCodec:
		return "varint"
	case *byteSliceCodec:
		return "byteSlice"
	case *varuintSliceCodec:
		return "varuintSlice"
	default:
		return "unknown"
	}
}
