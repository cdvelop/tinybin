package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestExactDecodeFlow(t *testing.T) {
	tb := New()
	// Reproduce the exact flow from Decode -> Decode -> scanToCache
	s := &simpleStruct{
		Name:      "Roman",
		Timestamp: 1357092245000000006,
		Payload:   []byte("hi"),
		Ssid:      []uint32{1, 2, 3},
	}

	// Encode first (this should work)
	b, err := tb.Encode(s)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	t.Logf("Encode succeeded: %v", b)

	// Now test the decode flow step by step
	dest := &simpleStruct{}

	// Step 1: Get the reflect value (decoder.go line 46)
	rv := tinyreflect.Indirect(tinyreflect.ValueOf(dest))
	t.Logf("rv.Type(): %p", rv.Type())

	// Step 2: Check if Type is nil
	if rv.Type() == nil {
		t.Fatal("rv.Type() is nil - this is the problem")
	}

	// Step 3: Call scan directly (this is the public API)
	codec, err := tb.scan(rv.Type())
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	t.Logf("scanToCache succeeded: %T", codec)

	// If this passes, the problem is elsewhere
	t.Logf("Test passed - problem might be in the actual decode flow")
}
