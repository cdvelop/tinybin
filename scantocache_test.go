package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestScanToCacheWithNilType(t *testing.T) {
	// Test what happens when we pass nil to scanToCache
	cache := make(map[*tinyreflect.Type]Codec)

	// This should fail
	_, err := scanToCache(nil, cache)
	if err != nil {
		t.Logf("✅ scanToCache correctly failed with nil: %v", err)
	} else {
		t.Error("❌ scanToCache should fail with nil type")
	}

	// Test with valid type
	type simpleStruct struct {
		Name      string
		Timestamp int64
		Payload   []byte
		Ssid      []uint32
	}

	s := &simpleStruct{}
	rv := tinyreflect.Indirect(tinyreflect.ValueOf(s))
	typ := rv.Type()

	if typ == nil {
		t.Fatal("rv.Type() returned nil")
	}

	t.Logf("Testing scanToCache with valid type: %p", typ)

	codec, err := scanToCache(typ, cache)
	if err != nil {
		t.Fatalf("scanToCache failed: %v", err)
	}

	if codec == nil {
		t.Fatal("scanToCache returned nil codec")
	}

	t.Logf("✅ scanToCache succeeded with codec: %T", codec)

	// Test that it's cached
	codec2, err := scanToCache(typ, cache)
	if err != nil {
		t.Fatalf("scanToCache failed on second call: %v", err)
	}

	if codec != codec2 {
		t.Error("❌ scanToCache returned different codec on second call")
	} else {
		t.Log("✅ scanToCache returned cached codec")
	}
}
