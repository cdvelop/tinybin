package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestScanToCacheWithNilType(t *testing.T) {
	// Test what happens when we pass nil to ScanType
	// This should fail
	_, err := ScanType(nil)
	if err != nil {
		t.Logf("✅ ScanType correctly failed with nil: %v", err)
	} else {
		t.Error("❌ ScanType should fail with nil type")
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

	t.Logf("Testing ScanType with valid type: %p", typ)

	codec, err := ScanType(typ)
	if err != nil {
		t.Fatalf("ScanType failed: %v", err)
	}

	if codec == nil {
		t.Fatal("ScanType returned nil codec")
	}

	t.Logf("✅ ScanType succeeded with codec: %T", codec)

	// Test that it works consistently (ScanType doesn't cache, so different instances are expected)
	codec2, err := ScanType(typ)
	if err != nil {
		t.Fatalf("ScanType failed on second call: %v", err)
	}

	if codec2 == nil {
		t.Error("❌ ScanType returned nil codec on second call")
	} else {
		t.Logf("✅ ScanType returned consistent codec type on second call: %T", codec2)
	}
}
