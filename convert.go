package tinybin

import (
	"unsafe"
)

// ToBytes converts a string to a byte slice without allocating.
func ToBytes(v string) []byte {
	// Use unsafe.StringData to get the data pointer directly
	data := unsafe.StringData(v)
	bytesData := unsafe.Slice(data, len(v))

	return bytesData
}
