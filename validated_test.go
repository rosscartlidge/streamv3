package ssql

import (
	"testing"
	"time"
)

func TestSetValidated_ValidTypes(t *testing.T) {
	mut := MakeMutableRecord()
	
	// Test all valid types
	mut.setValidated("int64", int64(42))
	mut.setValidated("float64", 3.14)
	mut.setValidated("bool", true)
	mut.setValidated("string", "test")
	mut.setValidated("time", time.Now())
	
	rec := mut.Freeze()
	
	if GetOr(rec, "int64", int64(0)) != int64(42) {
		t.Error("int64 field not set correctly")
	}
	if GetOr(rec, "float64", 0.0) != 3.14 {
		t.Error("float64 field not set correctly")
	}
	if GetOr(rec, "bool", false) != true {
		t.Error("bool field not set correctly")
	}
	if GetOr(rec, "string", "") != "test" {
		t.Error("string field not set correctly")
	}
}

func TestSetValidated_InvalidType_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid type, but didn't panic")
		} else {
			// Check panic message contains useful info
			msg := r.(string)
			if len(msg) == 0 {
				t.Error("Panic message was empty")
			}
			t.Logf("Caught expected panic: %v", r)
		}
	}()
	
	mut := MakeMutableRecord()
	// This should panic - []int is not a valid Value type
	mut.setValidated("bad", []int{1, 2, 3})
}
