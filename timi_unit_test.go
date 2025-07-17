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
