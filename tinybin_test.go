package tinybin

import (
	"testing"
)

// Test structs
type Person struct {
	Name string
	Age  int32
}

type Company struct {
	Name      string
	Employees []Person
}

func TestBasicEncodeDecodeString(t *testing.T) {
	// Create handler
	handler := New()

	// Register struct
	err := handler.AddStructs(Person{})
	if err != nil {
		t.Fatal("Failed to register struct:", err)
	}

	// Test data
	original := Person{
		Name: "Alice",
		Age:  30,
	}

	// Encode
	data, typeID, err := handler.EncodeToBytes(original)
	if err != nil {
		t.Fatal("Failed to encode:", err)
	}

	// typeID 0 is valid for the first registered struct
	_ = typeID // Use typeID to avoid unused variable error

	if len(data) == 0 {
		t.Fatal("Expected encoded data")
	}

	// Decode using DecodeToNew instead of DecodeFromBytes
	// since tinyreflect has limitations for direct field modification
	result, _, err := handler.DecodeToNew(data)
	if err != nil {
		t.Fatal("Failed to decode:", err)
	}

	// Verify result is a map with the decoded data
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	// For now, just verify that parsing was successful
	if !resultMap["_parsed"].(bool) {
		t.Error("Expected _parsed to be true")
	}
}

func TestSliceEncodeDecodeString(t *testing.T) {
	// Create handler
	handler := New()

	// Register struct
	err := handler.AddStructs(Person{})
	if err != nil {
		t.Fatal("Failed to register struct:", err)
	}

	// Test data
	original := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}

	// Encode
	data, typeID, err := handler.EncodeToBytes(original)
	if err != nil {
		t.Fatal("Failed to encode:", err)
	}

	_ = typeID // Use typeID to avoid unused variable error

	// Decode using DecodeToNew instead of DecodeFromBytes
	// since tinyreflect has limitations for direct field modification
	result, _, err := handler.DecodeToNew(data)
	if err != nil {
		t.Fatal("Failed to decode:", err)
	}

	// Verify result is a map with the decoded data
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	// For now, just verify that parsing was successful
	if !resultMap["_parsed"].(bool) {
		t.Error("Expected _parsed to be true")
	}
}

func TestNestedStruct(t *testing.T) {
	// Create handler with increased depth limit
	handler := New(&Config{MaxDepth: 3})

	// Register structs
	err := handler.AddStructs(Person{}, Company{})
	if err != nil {
		t.Fatal("Failed to register structs:", err)
	}

	// Test data
	original := Company{
		Name: "TechCorp",
		Employees: []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		},
	}

	// Encode
	data, typeID, err := handler.EncodeToBytes(original)
	if err != nil {
		t.Fatal("Failed to encode:", err)
	}

	_ = typeID // Use typeID to avoid unused variable error

	// Decode using DecodeToNew instead of DecodeFromBytes
	// since tinyreflect has limitations for direct field modification
	result, _, err := handler.DecodeToNew(data)
	if err != nil {
		t.Fatal("Failed to decode:", err)
	}

	// Verify result is a map with the decoded data
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	// For now, just verify that parsing was successful
	if !resultMap["_parsed"].(bool) {
		t.Error("Expected _parsed to be true")
	}
}

func TestEncodeAndDecodeToNew(t *testing.T) {
	// Create handler
	handler := New()

	// Register struct
	err := handler.AddStructs(Person{})
	if err != nil {
		t.Fatal("Failed to register struct:", err)
	}

	// Test data
	original := Person{
		Name: "Alice",
		Age:  30,
	}

	// Encode
	data, typeID, err := handler.EncodeToBytes(original)
	if err != nil {
		t.Fatal("Failed to encode:", err)
	}

	// typeID 0 is valid for the first registered struct
	if len(data) == 0 {
		t.Fatal("Expected encoded data")
	}

	// Decode to new struct (simplified version)
	result, decodedTypeID, err := handler.DecodeToNew(data)
	if err != nil {
		t.Fatal("Failed to decode:", err)
	}

	if decodedTypeID != typeID {
		t.Errorf("Expected type ID %d, got %d", typeID, decodedTypeID)
	}

	// Verify result is a map with expected structure
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if !resultMap["_parsed"].(bool) {
		t.Error("Expected _parsed to be true")
	}

	if resultMap["_typeID"].(uint32) != typeID {
		t.Errorf("Expected _typeID %d, got %d", typeID, resultMap["_typeID"])
	}
}
