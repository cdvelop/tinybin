package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestScanToCacheWithNilType(t *testing.T) {
	tb := New()
	// Test what happens when we pass nil to scan
	// This should fail
	_, err := tb.scan(nil)
	if err != nil {
		t.Logf("✅ scan correctly failed with nil: %v", err)
	} else {
		t.Error("❌ scan should fail with nil type")
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

	t.Logf("Testing scan with valid type: %p", typ)

	codec, err := tb.scan(typ)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if codec == nil {
		t.Fatal("scan returned nil codec")
	}

	t.Logf("✅ scan succeeded with codec: %T", codec)

	// Test that it works consistently (scan doesn't cache, so different instances are expected)
	codec2, err := tb.scan(typ)
	if err != nil {
		t.Fatalf("scan failed on second call: %v", err)
	}

	if codec2 == nil {
		t.Error("❌ scan returned nil codec on second call")
	} else {
		t.Logf("✅ scan returned consistent codec type on second call: %T", codec2)
	}
}
