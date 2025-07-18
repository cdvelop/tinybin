package tinybin

import (
	"testing"

	"github.com/cdvelop/tinyreflect"
)

// TestDecoderScenario tests the specific reflection patterns used in decoder.go
// This ensures that decoder operations work correctly with tinyreflect
func TestDecoderScenario(t *testing.T) {
	// This replicates the exact scenario from decoder.go line 46
	type simpleStruct struct {
		Name      string
		Timestamp int64
		Payload   []byte
		Ssid      []uint32
	}

	// Test scenario like in decoder: pointer to struct
	s := &simpleStruct{}

	// This is exactly what decoder.go does
	rv := tinyreflect.Indirect(tinyreflect.ValueOf(s))

	// Check that rv has a valid type
	typ := rv.Type()
	if typ == nil {
		t.Error("rv.Type() returned nil - this is the 'value type nil' error")
	} else {
		t.Logf("rv.Type() returned %p, Kind: %v", typ, typ.Kind())
	}

	// Check CanAddr
	canAddr := rv.CanAddr()
	t.Logf("rv.CanAddr() = %v", canAddr)

	// Check if the original value has a type
	originalV := tinyreflect.ValueOf(s)
	if originalV.Type() == nil {
		t.Error("ValueOf(s).Type() returned nil")
	} else {
		t.Logf("ValueOf(s).Type() returned %p, Kind: %v", originalV.Type(), originalV.Type().Kind())
	}

	// Compare with direct struct (not pointer)
	directStruct := simpleStruct{}
	directRv := tinyreflect.ValueOf(directStruct)
	if directRv.Type() == nil {
		t.Error("Direct struct ValueOf returned nil type")
	} else {
		t.Logf("Direct struct Type() returned %p, Kind: %v", directRv.Type(), directRv.Type().Kind())
	}

	// Test the decoding workflow that would happen
	t.Run("DecodingWorkflow", func(t *testing.T) {
		// Test that we can access struct fields for decoding
		if typ != nil {
			numFields, err := typ.NumField()
			if err != nil {
				t.Errorf("NumField() failed: %v", err)
				return
			}

			if numFields > 0 {
				for i := 0; i < numFields; i++ {
					field, err := rv.Field(i)
					if err != nil {
						t.Errorf("Field(%d) failed: %v", i, err)
						continue
					}
					if field.Type() == nil {
						t.Errorf("Field(%d) has nil type", i)
					}
				}
			}
		}
	})
}
