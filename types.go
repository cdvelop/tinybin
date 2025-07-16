package tinybin

import . "github.com/cdvelop/tinystring"

// stObject holds the cached metadata for a single struct type.
type stObject struct {
	stID     uint32    // Unique ID (index in the stObjects slice)
	stFields []stField // Ordered list of field metadata
}

// stField holds the metadata for a single struct field.
type stField struct {
	name     string // Field name
	typeKind Kind   // The kind of the field (int, string, slice, etc.) from tinyreflect
	// Additional fields for offsets, nested type IDs, etc.
}
