package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestDecoderFullFlow(t *testing.T) {
	tb := New()
	// This replicates the exact flow from decoder.go
	type simpleStruct struct {
		Name      string
		Timestamp int64
		Payload   []byte
		Ssid      []uint32
	}

	s := &simpleStruct{}

	// decoder.go line 46: rv := tinyreflect.Indirect(tinyreflect.ValueOf(v))
	rv := tinyreflect.Indirect(tinyreflect.ValueOf(s))

	// Check CanAddr
	canAddr := rv.CanAddr()
	if !canAddr {
		t.Fatal("rv.CanAddr() returned false")
	}

	// decoder.go line 52: scanToCache(rv.Type(), d.schemas)
	typ := rv.Type()
	if typ == nil {
		t.Fatal("rv.Type() returned nil - this is the problem!")
	}

	t.Logf("rv.Type() = %p, Kind: %v", typ, typ.Kind())

	// decoder.go line 52: scan(rv.Type())
	codec, err := tb.scan(typ)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if codec == nil {
		t.Fatal("ScanType returned nil codec")
	}

	t.Logf("ScanType succeeded, codec: %T", codec)
}
