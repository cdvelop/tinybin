package tinybin

// Encode encodes a slice of structs to an io.Writer.
func (h *TinyBin) Encode(w any, data any) error {
	// TODO: Implement encoding logic using stObjects schema.
	return nil
}

// EncodeToBytes encodes a slice of structs to a byte slice and returns the type ID.
func (h *TinyBin) EncodeToBytes(data any) ([]byte, uint16, error) {
	// TODO: Implement encoding to bytes.
	return nil, 0, nil
}
