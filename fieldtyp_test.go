package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

func TestFieldTypNil(t *testing.T) {
	// Test the exact struct from the failing test
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

	// Check each field individually
	numFields, err := typ.NumField()
	if err != nil {
		t.Fatalf("NumField() failed: %v", err)
	}

	t.Logf("Struct has %d fields", numFields)

	for i := 0; i < numFields; i++ {
		field, err := typ.Field(i)
		if err != nil {
			t.Fatalf("Field(%d) failed: %v", i, err)
		}

		t.Logf("Field %d: Name=%s, Typ=%p", i, field.Name, field.Typ)

		if field.Typ == nil {
			t.Errorf("❌ Field %d (%s) has nil Typ!", i, field.Name)
		} else {
			t.Logf("✅ Field %d (%s) has Typ: %p, Kind: %v", i, field.Name, field.Typ, field.Typ.Kind())

			// Test scanType on this field
			codec, err := scanType(field.Typ)
			if err != nil {
				t.Errorf("❌ scanType failed for field %d (%s): %v", i, field.Name, err)
			} else {
				t.Logf("✅ scanType succeeded for field %d (%s): %T", i, field.Name, codec)
			}
		}
	}
}
