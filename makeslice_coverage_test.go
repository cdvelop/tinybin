package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

// TestMakeSliceDirectUsage tests MakeSlice functionality directly
// This provides coverage for the MakeSlice function and slice element access
func TestMakeSliceDirectUsage(t *testing.T) {
	// Test MakeSlice directly
	structType := tinyreflect.TypeOf(simpleStruct{})
	t.Logf("Original struct type: %p, Kind: %s", structType, structType.Kind().String())

	// Create a slice type manually (this is what should be happening)
	sliceOfStructs := []simpleStruct{}
	sliceType := tinyreflect.TypeOf(sliceOfStructs)
	t.Logf("Slice type: %p, Kind: %s", sliceType, sliceType.Kind().String())

	// Try to create a slice using MakeSlice
	newSlice, err := tinyreflect.MakeSlice(sliceType, 2, 2)
	if err != nil {
		t.Fatalf("MakeSlice failed: %v", err)
	}

	t.Logf("newSlice created: typ=%p, Kind: %s", newSlice.Type(), newSlice.Type().Kind().String())

	// Try to get length
	length, err := newSlice.Len()
	if err != nil {
		t.Fatalf("Len() failed: %v", err)
	}
	t.Logf("Slice length: %d", length)

	// Try to index into it
	for i := 0; i < length; i++ {
		elem, err := newSlice.Index(i)
		if err != nil {
			t.Fatalf("Index(%d) failed: %v", i, err)
		}
		t.Logf("Element %d: typ=%p, Kind: %s", i, elem.Type(), elem.Type().Kind().String())
	}
}
