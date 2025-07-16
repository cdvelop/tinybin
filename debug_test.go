package tinybin

import (
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
	println("Person StructID:", structID)

	// Register struct
	err := handler.AddStructs(person)
	if err != nil {
		t.Fatal("Failed to register struct:", err)
	}

	// Check what was registered
	println("Registered objects count:", len(handler.stObjects))
	for i, obj := range handler.stObjects {
		println("Object", i, ": ID=", obj.stID, ", Fields=", len(obj.stFields))
	}

	// Try to encode and see what happens
	data, typeID, err := handler.EncodeToBytes(person)
	if err != nil {
		println("Encode error:", err)
	} else {
		println("Encode successful: typeID=", typeID, ", data length=", len(data))
	}
}
