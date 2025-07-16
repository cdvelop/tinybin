package tinybin

import (
	"fmt"
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestDebugRegistration(t *testing.T) {
	// Create handler
	handler := New()

	// Test struct
	person := Person{Name: "test", Age: 25}

	// Check struct ID before registration
	v := tinyreflect.ValueOf(person)
	structID := v.Type().StructID()
	fmt.Printf("Person StructID: %d\n", structID)

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

	// Try to encode and see what happens
	data, typeID, err := handler.EncodeToBytes(person)
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
	} else {
		fmt.Printf("Encode successful: typeID=%d, data length=%d\n", typeID, len(data))
	}
}
