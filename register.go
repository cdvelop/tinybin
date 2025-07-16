package tinybin

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

// AddStructs registers one or more struct types for encoding/decoding.
// The order of registration must be identical on both client and server.
func (h *TinyBin) AddStructs(structs ...any) error {
	for _, s := range structs {
		v := tinyreflect.ValueOf(s)
		t := v.Type()

		if t.Kind() != K.Struct {
			return Err(D.Type, "input", D.Not, D.Struct)
		}

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
		return Err(D.Type, D.Exceeds, D.Maximum)
	}

	// Use StructID from tinyreflect to get unique struct identifier
	structID := t.StructID()

	// Check if already registered by StructID
	for _, obj := range h.stObjects {
		if obj.stID == structID {
			return nil // Already analyzed
		}
	}

	// Create new stObject using StructID
	stObj := stObject{
		stID: structID,
	}

	// Analyze fields
	numFields, err := t.NumField()
	if err != nil {
		return Err(D.Struct, D.Invalid, D.Fields)
	}

	for i := 0; i < numFields; i++ {
		field, err := t.Field(i)
		if err != nil {
			return Err(D.Field, D.At, D.Index, i)
		}

		// Skip unexported fields (tinyreflect handles this internally)
		fieldName, err := t.NameByIndex(i)
		if err != nil {
			continue // Skip fields we can't access
		}

		// Get field type kind using tinyreflect
		fieldType := field.Typ
		fieldKind := fieldType.Kind()

		stFld := stField{
			name:     fieldName,
			typeKind: fieldKind,
		}

		// Recursively analyze nested structs
		if fieldKind == K.Struct {
			err := h.analyzeStruct(fieldType, depth+1)
			if err != nil {
				return err
			}
		}

		stObj.stFields = append(stObj.stFields, stFld)
	}

	h.stObjects = append(h.stObjects, stObj)
	return nil
}
