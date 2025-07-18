package tinybin

import (
	"bytes"
	"testing"

	"github.com/cdvelop/tinyreflect"
)

type testCustom string

// GetBinaryCodec retrieves a custom binary codec.
func (s *testCustom) GetBinaryCodec() Codec {
	return new(stringCodec)
}

func TestScanner(t *testing.T) {
	rt := tinyreflect.Indirect(tinyreflect.ValueOf(s0v)).Type()
	codec, err := scanType(rt)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if codec == nil {
		t.Fatal("Expected non-nil codec")
	}

	var b bytes.Buffer
	e := NewEncoder(&b)
	err = codec.EncodeTo(e, tinyreflect.Indirect(tinyreflect.ValueOf(s0v)))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !bytes.Equal(s0b, b.Bytes()) {
		t.Errorf("Expected %v, got %v", s0b, b.Bytes())
	}
}

func TestScanner_Custom(t *testing.T) {
	v := testCustom("test")
	rt := tinyreflect.Indirect(tinyreflect.ValueOf(v)).Type()
	codec, err := scanType(rt)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if codec == nil {
		t.Fatal("Expected non-nil codec")
	}
}

func TestScannerComposed(t *testing.T) {
	codec, err := scanType(tinyreflect.TypeOf(Partition{}))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if codec == nil {
		t.Fatal("Expected non-nil codec")
	}
}

type Partition struct {
	Strings
	Filters map[uint32][]uint64
}

type Strings struct {
	Key  string
	Fill []uint64
	Hash []uint32
	Data map[uint64][]byte
}
