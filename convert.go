//go:build !wasm
// +build !wasm

package tinybin

import (
	"reflect"
	"unsafe"
)

// ToString converts byte slice to a string without allocating.
func ToString(b *[]byte) string {
	return *(*string)(unsafe.Pointer(b))
}

// ToBytes converts a string to a byte slice without allocating.
func ToBytes(v string) []byte {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&v))
	bytesData := unsafe.Slice((*byte)(unsafe.Pointer(strHeader.Data)), len(v))

	return bytesData
}

func binaryToBools(b *[]byte) []bool {
	return *(*[]bool)(unsafe.Pointer(b))
}

func boolsToBinary(v *[]bool) []byte {
	return *(*[]byte)(unsafe.Pointer(v))
}
