package tinybin

import (
	"testing"
	"unsafe"

	"github.com/cdvelop/tinyreflect"
)

func TestStructFieldTypInitializationInTinybin(t *testing.T) {
	// Use the exact same struct as the failing test
	s := &simpleStruct{
		Name:      "Roman",
		Timestamp: 1357092245000000006,
		Payload:   []byte("hi"),
		Ssid:      []uint32{1, 2, 3},
	}

	rv := tinyreflect.Indirect(tinyreflect.ValueOf(s))
	typ := rv.Type()

	if typ == nil {
		t.Fatal("rv.Type() returned nil")
	}

	t.Logf("Type: %p, Kind: %v", typ, typ.Kind())

	// Check if it's a struct
	if typ.Kind().String() != "struct" {
		t.Fatalf("Expected struct, got %v", typ.Kind())
	}

	// Cast to StructType to access Fields directly
	st := (*tinyreflect.StructType)(unsafe.Pointer(typ))
	t.Logf("StructType.Fields length: %d", len(st.Fields))

	// Check each field
	for i, field := range st.Fields {
		t.Logf("Field %d: Name=%s, Typ=%p", i, field.Name, field.Typ)
		if field.Typ == nil {
			t.Errorf("❌ Field %d (%s) has nil Typ!", i, field.Name)
		} else {
			t.Logf("✅ Field %d (%s) has Typ: %p, Kind: %v", i, field.Name, field.Typ, field.Typ.Kind())
		}
	}

	// Also test Field() method
	numFields, err := typ.NumField()
	if err != nil {
		t.Fatalf("NumField() failed: %v", err)
	}
	t.Logf("NumField() returned: %d", numFields)

	for i := 0; i < numFields; i++ {
		field, err := typ.Field(i)
		if err != nil {
			t.Fatalf("Field(%d) failed: %v", i, err)
		}
		t.Logf("Field(%d) via method: Name=%s, Typ=%p", i, field.Name, field.Typ)

		// Test ScanType on each field
		if field.Typ == nil {
			t.Errorf("❌ Field %d (%s) has nil Typ!", i, field.Name)
		} else {
			codec, err := ScanType(field.Typ)
			if err != nil {
				t.Errorf("❌ ScanType failed for field %d (%s): %v", i, field.Name, err)
			} else {
				t.Logf("✅ ScanType succeeded for field %d (%s): %T", i, field.Name, codec)
			}
		}
	}
}
