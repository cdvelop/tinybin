package tinybin

// LEB128 Variable-length integer encoding implementation for TinyBin protocol.
// Uses Little Endian Base 128 encoding for efficient integer serialization.

// encodeVarint encodes a uint32 as a LEB128 varint
func encodeVarint(x uint32) []byte {
	var buf []byte
	for x >= 0x80 {
		buf = append(buf, byte(x)|0x80)
		x >>= 7
	}
	buf = append(buf, byte(x))
	return buf
}

// decodeVarint decodes a LEB128 varint from bytes, returns the value and bytes consumed
func decodeVarint(buf []byte) (uint32, int, error) {
	var x uint32
	var s uint
	for i, b := range buf {
		if i == 5 {
			// Overflow: uint32 can have at most 5 bytes in varint encoding
			return 0, 0, ErrVarintOverflow
		}
		if b < 0x80 {
			return x | uint32(b)<<s, i + 1, nil
		}
		x |= uint32(b&0x7f) << s
		s += 7
	}
	return 0, 0, ErrVarintTruncated
}

// encodeVarint64 encodes a uint64 as a LEB128 varint
func encodeVarint64(x uint64) []byte {
	var buf []byte
	for x >= 0x80 {
		buf = append(buf, byte(x)|0x80)
		x >>= 7
	}
	buf = append(buf, byte(x))
	return buf
}

// decodeVarint64 decodes a LEB128 varint from bytes, returns the value and bytes consumed
func decodeVarint64(buf []byte) (uint64, int, error) {
	var x uint64
	var s uint
	for i, b := range buf {
		if i == 10 {
			// Overflow: uint64 can have at most 10 bytes in varint encoding
			return 0, 0, ErrVarintOverflow
		}
		if b < 0x80 {
			return x | uint64(b)<<s, i + 1, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, 0, ErrVarintTruncated
}
