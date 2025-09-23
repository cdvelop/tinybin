package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

// TestStructPointerFieldAccess verifies that struct fields containing pointers
// can be properly accessed and marshaled/unmarshaled through tinyreflect
func TestStructPointerFieldAccess(t *testing.T) {
	tb := New()
	type InnerStruct struct {
		V int
	}
	type OuterStruct struct {
		Ptr *InnerStruct
	}
	
	// Test case 1: Non-nil pointer
	t.Run("NonNilPointer", func(t *testing.T) {
		original := &OuterStruct{Ptr: &InnerStruct{V: 42}}
		
		// Verify basic field access works correctly
		rv := tinyreflect.ValueOf(original)
		elem, err := rv.Elem()
		if err != nil {
			t.Fatalf("rv.Elem() failed: %v", err)
		}
		
		// Verify we can access the pointer field
		ptrField, err := elem.Field(0)
		if err != nil {
			t.Fatalf("elem.Field(0) failed: %v", err)
		}
		
		// Verify the field has correct type and kind
		if ptrField.Type() == nil {
			t.Fatal("ptrField.Type() returned nil")
		}
		if ptrField.Kind() != K.Pointer {
			t.Errorf("Expected pointer kind, got %v", ptrField.Kind())
		}
		
		// Verify marshal/unmarshal roundtrip
		payload, err := tb.Encode(original)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		decoded := &OuterStruct{}
		err = tb.Decode(payload, decoded)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		
		// Verify the result
		if decoded.Ptr == nil {
			t.Fatal("Decoded pointer is nil")
		}
		if decoded.Ptr.V != original.Ptr.V {
			t.Errorf("Expected V=%d, got V=%d", original.Ptr.V, decoded.Ptr.V)
		}
	})
	
	// Test case 2: Nil pointer
	t.Run("NilPointer", func(t *testing.T) {
		original := &OuterStruct{Ptr: nil}

		payload, err := tb.Encode(original)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		decoded := &OuterStruct{}
		err = tb.Decode(payload, decoded)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		
		// Verify nil pointer is preserved
		if decoded.Ptr != nil {
			t.Error("Expected nil pointer, but got non-nil")
		}
	})
}
