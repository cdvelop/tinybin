package tinybin

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

// AddStructs registers one or more struct types for encoding/decoding.
// The order of registration must be identical on both client and server.
func (h *TinyBin) AddStructs(structs ...any) error {
	for _, s := range structs {
		t := tinyreflect.TypeOf(s)
		err := h.analyzeStruct(t, 1)
		if err != nil {
			return err
		}
	}
	return nil
}

// analyzeStruct recursively analyzes a struct type and adds it to the cache.
func (h *TinyBin) analyzeStruct(t *tinyreflect.Type, depth int) error {
	if depth > h.MaxDepth {
		return Err(D.Type, t.Name(), D.Exceeds, D.Struct)
	}
	// TODO: Implement recursive analysis, cycle detection, and stObject creation.
	return nil
}
