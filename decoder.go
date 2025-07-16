package tinybin

// Decode decodes from an io.Reader into a new slice of structs.
func (h *TinyBin) Decode(r any) (any, error) {
	// TODO: Implement decoding logic using stObjects schema.
	return nil, nil
}

// DecodeFromBytes decodes from a byte slice into a provided slice pointer.
func (h *TinyBin) DecodeFromBytes(b []byte, typeID uint16, out any) error {
	// TODO: Implement decoding from bytes.
	return nil
}
