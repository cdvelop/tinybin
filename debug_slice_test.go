package tinybin

import (
	"fmt"
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestDebugSliceRegistration(t *testing.T) {
	// Create handler
	handler := New()

	// Test struct
	person := Person{Name: "test", Age: 25}
	slice := []Person{{Name: "test1", Age: 25}, {Name: "test2", Age: 30}}

	// Check struct ID for single struct
	v1 := tinyreflect.ValueOf(person)
	structID1 := v1.Type().StructID()
	fmt.Printf("Person StructID: %d\n", structID1)

	// Check struct ID for slice
	v2 := tinyreflect.ValueOf(slice)
	t2 := v2.Type()
	structID2 := t2.StructID()
	fmt.Printf("[]Person Type Kind: %v\n", t2.Kind())
	fmt.Printf("[]Person StructID: %d\n", structID2)

	// Register struct
	err := handler.AddStructs(person)
	if err != nil {
		t.Fatal("Failed to register struct:", err)
	}

	// Check what was registered
	fmt.Printf("Registered objects count: %d\n", len(handler.stObjects))
	for i, obj := range handler.stObjects {
		fmt.Printf("Object %d: ID=%d, Fields=%d\n", i, obj.stID, len(obj.stFields))
	}

	// Try to encode slice and see what happens
	data, typeID, err := handler.EncodeToBytes(slice)
	if err != nil {
		fmt.Printf("Slice encode error: %v\n", err)
	} else {
		fmt.Printf("Slice encode successful: typeID=%d, data length=%d\n", typeID, len(data))
	}
}
