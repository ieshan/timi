package timi

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFrom(t *testing.T) {
	ti1 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	ti2 := Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	if ti1 != ti2.Time {
		t.Fatalf("Expected %v, got %v", ti1, ti2)
	}
}

func TestTime_String(t *testing.T) {
	ti := Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	if ti.String() != "2021-01-01 00:00:00 +0000 UTC" {
		t.Fatalf("Expected %v, got %v", "2021-01-01 00:00:00 +0000 UTC", ti.String())
	}
}

func TestTime_MarshalJSON(t *testing.T) {
	type TimeTestStruct struct {
		Time1 Time `json:"time1"`
		Time2 Time `json:"time2"`
	}
	ti := Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	jsonVal, err := json.Marshal(&TimeTestStruct{Time1: ti, Time2: NilTime})
	if err != nil {
		t.Fatalf("Got error while marshaling to JSON %v", err)
	}
	actualVal := string(jsonVal)
	expectedVal := `{"time1":"2021-01-01T00:00:00Z","time2":null}`
	if actualVal != expectedVal {
		t.Fatalf("Original Time (%s) did not match with the Time in generated JSON %s", expectedVal, actualVal)
	}
}

func TestTime_UnmarshalJSON(t *testing.T) {
	type TimeTestStruct struct {
		Time1 Time `json:"time1"`
		Time2 Time `json:"time2"`
	}
	jsonStrs := []string{
		`{"time1":"2021-01-01T00:00:00Z","time2":"2021-01-01T00:00:00Z"}`,
		`{"time1":"2021-01-01T00:00:00Z","time2":null}`,
		`{"time1":"2021-01-01T00:00:00Z","time2":""}`,
		`{"time1":"2021-01-01T00:00:00Z","time2":"null"}`,
	}
	ti := Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	timeVals := []TimeTestStruct{
		{Time1: ti, Time2: ti},
		{Time1: ti, Time2: NilTime},
	}
	hasErr := []bool{
		false,
		false,
		true,
		true,
	}
	unmVal := TimeTestStruct{}
	for index, jsonStr := range jsonStrs {
		err := json.Unmarshal([]byte(jsonStr), &unmVal)
		if hasErr[index] && err == nil {
			t.Fatalf("Case %d: Expecting error, but got nil", index)
		} else if !hasErr[index] && err != nil {
			t.Fatalf("Case %d: Got error while unmarshaling JSON %v", index, err)
		}
		if hasErr[index] {
			continue
		}
		if !unmVal.Time1.Equal(timeVals[index].Time1) {
			t.Fatalf("Case %d: Original Time (%v) did not match with the Time from JSON %v", index, timeVals[index], unmVal)
		}
		if unmVal.Time2.IsNull() != timeVals[index].Time2.IsNull() {
			t.Fatalf("Case %d: Time2 is null mismatch (%v : %v)", index, unmVal.Time2.IsNull(), timeVals[index].Time2.IsNull())
		}
		if !unmVal.Time2.IsNull() && !unmVal.Time2.Equal(timeVals[index].Time2) {
			t.Fatalf("Case %d: Original Time (%v) did not match with the Time from JSON %v", index, timeVals[index].Time2, unmVal.Time2)
		}
	}
}

func TestTime_Scan(t *testing.T) {
	base := time.Date(2024, time.January, 15, 10, 30, 45, 123456000, time.UTC)

	testCases := []struct {
		name     string
		input    interface{}
		expected Time
		hasErr   bool
	}{
		{"nil", nil, NilTime, false},
		{"time.Time", base, Time{Time: base, Valid: true}, false},
		{"time.Time non-UTC", base.In(time.FixedZone("EST", -5*3600)), Time{Time: base, Valid: true}, false},
		{"empty string", "", NilTime, false},
		{"empty bytes", []byte(""), NilTime, false},
		{"RFC3339Nano", "2024-01-15T10:30:45.123456789Z", Time{Time: time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC), Valid: true}, false},
		{"RFC3339", "2024-01-15T10:30:45Z", Time{Time: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC), Valid: true}, false},
		{"mattn storage format", "2024-01-15 10:30:45.123456+00:00", Time{Time: base, Valid: true}, false},
		{"space sep + Z", "2024-01-15 10:30:45Z", Time{Time: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC), Valid: true}, false},
		{"naive space sep frac", "2024-01-15 10:30:45.123456", Time{Time: base, Valid: true}, false},
		{"naive space sep", "2024-01-15 10:30:45", Time{Time: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC), Valid: true}, false},
		{"naive T sep", "2024-01-15T10:30:45", Time{Time: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC), Valid: true}, false},
		{"minute precision space", "2024-01-15 10:30", Time{Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), Valid: true}, false},
		{"minute precision T", "2024-01-15T10:30", Time{Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), Valid: true}, false},
		{"date only", "2024-01-15", Time{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Valid: true}, false},
		{"positive offset", "2024-01-15 10:30:45+05:30", Time{Time: time.Date(2024, 1, 15, 5, 0, 45, 0, time.UTC), Valid: true}, false},
		{"[]byte RFC3339", []byte("2024-01-15T10:30:45Z"), Time{Time: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC), Valid: true}, false},
		{"[]byte naive", []byte("2024-01-15 10:30:45"), Time{Time: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC), Valid: true}, false},
		{"invalid string", "not-a-time", Time{}, true},
		{"invalid date", "2024-13-40", Time{}, true},
		{"int64", int64(1705315845), Time{}, true},
		{"float64", float64(2460315.5), Time{}, true},
		{"bool", true, Time{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var scanned Time
			err := scanned.Scan(tc.input)
			if tc.hasErr {
				if err == nil {
					t.Fatalf("Expected error for input %v, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error for input %v: %v", tc.input, err)
			}
			if scanned.Valid != tc.expected.Valid {
				t.Fatalf("Valid mismatch for %v: expected %v, got %v", tc.input, tc.expected.Valid, scanned.Valid)
			}
			if scanned.Valid && !scanned.Time.Equal(tc.expected.Time) {
				t.Fatalf("Time mismatch for %v: expected %v, got %v", tc.input, tc.expected.Time, scanned.Time)
			}
			if scanned.Valid && scanned.Time.Location() != time.UTC {
				t.Fatalf("Location should be UTC for %v, got %v", tc.input, scanned.Time.Location())
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		parsed, err := ParseTime("2024-01-15 10:30:45")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		expected := Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
		if !parsed.Equal(expected) {
			t.Fatalf("Expected %v, got %v", expected, parsed)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if _, err := ParseTime(""); err == nil {
			t.Fatalf("Expected error for empty string, got nil")
		}
	})
	t.Run("invalid", func(t *testing.T) {
		if _, err := ParseTime("not-a-time"); err == nil {
			t.Fatalf("Expected error for invalid string, got nil")
		}
	})
}
