package mongodb

import (
	"testing"
	"time"

	"github.com/ieshan/timi"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBSONHelpers(t *testing.T) {
	// Test MarshalTimiBSON with valid time
	ti := timi.Date(2021, 1, 1, 12, 30, 45, 0, time.UTC)
	bType, data, err := MarshalTimiBSON(ti)
	if err != nil {
		t.Fatalf("MarshalTimiBSON with valid time failed: %v", err)
	}
	if bType == bson.TypeNull {
		t.Fatalf("Expected non-null BSON type for valid time, got TypeNull")
	}
	if len(data) == 0 {
		t.Fatalf("Expected non-empty data for valid time")
	}

	// Test MarshalTimiBSON with null time
	bTypeNull, dataNull, err := MarshalTimiBSON(timi.NilTime)
	if err != nil {
		t.Fatalf("MarshalTimiBSON with null time failed: %v", err)
	}
	if bTypeNull != bson.TypeNull {
		t.Fatalf("Expected TypeNull for null time, got %v", bTypeNull)
	}
	if dataNull != nil {
		t.Fatalf("Expected nil data for null time, got %v", dataNull)
	}

	// Test UnmarshalTimiBSON with valid data (round trip)
	unmarshaled, err := UnmarshalTimiBSON(bType, data)
	if err != nil {
		t.Fatalf("UnmarshalTimiBSON with valid data failed: %v", err)
	}
	if !unmarshaled.Equal(ti) {
		t.Fatalf("Round trip failed: expected %v, got %v", ti, unmarshaled)
	}
	if unmarshaled.IsNull() {
		t.Fatalf("Unmarshaled time should not be null")
	}

	// Test UnmarshalTimiBSON with null data
	unmarshaledNull, err := UnmarshalTimiBSON(bson.TypeNull, nil)
	if err != nil {
		t.Fatalf("UnmarshalTimiBSON with null data failed: %v", err)
	}
	if !unmarshaledNull.IsNull() {
		t.Fatalf("Unmarshaled null time should be null")
	}

	// Test TimiBSONWrapper marshaling/unmarshaling
	wrapper := TimiBSONWrapper{Time: ti}
	wrapperBType, wrapperData, err := wrapper.MarshalBSONValue()
	if err != nil {
		t.Fatalf("TimiBSONWrapper MarshalBSONValue failed: %v", err)
	}
	if wrapperBType == bson.TypeNull {
		t.Fatalf("Expected non-null BSON type for wrapper with valid time")
	}

	// Test unmarshaling into wrapper
	var newWrapper TimiBSONWrapper
	err = newWrapper.UnmarshalBSONValue(wrapperBType, wrapperData)
	if err != nil {
		t.Fatalf("TimiBSONWrapper UnmarshalBSONValue failed: %v", err)
	}
	if !newWrapper.Time.Equal(ti) {
		t.Fatalf("Wrapper round trip failed: expected %v, got %v", ti, newWrapper.Time)
	}

	// Test TimiBSONWrapper with null time
	nullWrapper := TimiBSONWrapper{Time: timi.NilTime}
	nullWrapperBType, nullWrapperData, err := nullWrapper.MarshalBSONValue()
	if err != nil {
		t.Fatalf("TimiBSONWrapper MarshalBSONValue with null time failed: %v", err)
	}
	if nullWrapperBType != bson.TypeNull {
		t.Fatalf("Expected TypeNull for wrapper with null time, got %v", nullWrapperBType)
	}

	// Test unmarshaling null into wrapper
	var nullNewWrapper TimiBSONWrapper
	err = nullNewWrapper.UnmarshalBSONValue(nullWrapperBType, nullWrapperData)
	if err != nil {
		t.Fatalf("TimiBSONWrapper UnmarshalBSONValue with null failed: %v", err)
	}
	if !nullNewWrapper.Time.IsNull() {
		t.Fatalf("Wrapper with null time should be null after round trip")
	}
}
