package tinybin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvert_String(t *testing.T) {
	v := "hi there"

	b := ToBytes(v)
	assert.NotEmpty(t, b)
	assert.Equal(t, v, string(b))

	o := ToString(&b)
	assert.NotEmpty(t, b)
	assert.Equal(t, v, o)
}

func TestConvert_Bools(t *testing.T) {
	v := []bool{true, false, true, true, false, false}

	b := boolsToBinary(&v)
	assert.NotEmpty(t, b)
	assert.Equal(t, []byte{0x1, 0x0, 0x1, 0x1, 0x0, 0x0}, b)

	o := binaryToBools(&b)
	assert.NotEmpty(t, b)
	assert.Equal(t, v, o)
}
