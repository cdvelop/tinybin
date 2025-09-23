package tinybin

import (
	"testing"
)

// Simple struct with only basic types (no slices, arrays, or pointers)
type basicStruct struct {
	Name    string
	Age     int
	Height  float64
	IsAdult bool
}

func TestBasicStruct(t *testing.T) {
	v := &basicStruct{
		Name:    "John",
		Age:     25,
		Height:  1.75,
		IsAdult: true,
	}

	b, err := Encode(v)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	s := &basicStruct{}
	err = Decode(b, s)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if v.Name != s.Name || v.Age != s.Age || v.Height != s.Height || v.IsAdult != s.IsAdult {
		t.Errorf("Expected %+v, got %+v", v, s)
	}
}
