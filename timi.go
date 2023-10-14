package timi

import (
	"database/sql"
	"database/sql/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"time"
)

type Time sql.NullTime

var NilTime = Time{Time: time.Time{}, Valid: false}

func (t Time) Zone() (string, int) {
	return "UTC", 0
}

func (t Time) String() string {
	return t.Time.String()
}

func (t *Time) IsNull() bool {
	return !t.Valid
}

// After reports whether the time instant t is after u.
func (t Time) After(u Time) bool {
	return time.Time(t.Time).After(time.Time(u.Time))
}

// Before reports whether the time instant t is before u.
func (t Time) Before(u Time) bool {
	return time.Time(t.Time).Before(time.Time(u.Time))
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t Time) Compare(u Time) int {
	return time.Time(t.Time).Compare(time.Time(u.Time))
}

// Equal reports whether t and u represent the same time instant.
// Two times can be equal even if they are in different locations.
// For example, 6:00 +0200 and 4:00 UTC are Equal.
// See the documentation on the Time type for the pitfalls of using == with
// Time values; most code should use Equal instead.
func (t Time) Equal(u Time) bool {
	return t.Time.Equal(u.Time)
}

func (t *Time) Scan(value interface{}) (err error) {
	if err = (*sql.NullTime)(t).Scan(value); err != nil {
		return err
	}
	if t.Valid {
		t.Time = t.Time.UTC()
	}
	return
}

func (t Time) Value() (driver.Value, error) {
	return sql.NullTime(t).Value()
}

func (t Time) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte{'n', 'u', 'l', 'l'}, nil
	}
	return t.Time.MarshalJSON()
}

func (t *Time) UnmarshalJSON(data []byte) error {
	if len(data) == 4 && data[0] == 'n' && data[1] == 'u' && data[2] == 'l' && data[3] == 'l' {
		t.Time, t.Valid = time.Time{}, false
		return nil
	}
	t.Valid = true
	if err := t.Time.UnmarshalJSON(data); err != nil {
		return err
	}
	t.Time = t.Time.UTC()
	return nil
}

func (t Time) MarshalText() ([]byte, error) {
	return t.Time.MarshalText()
}

func (t *Time) UnmarshalText(data []byte) error {
	if len(data) == 4 && data[0] == 'n' && data[1] == 'u' && data[2] == 'l' && data[3] == 'l' {
		t.Time, t.Valid = time.Time{}, false
		return nil
	}
	t.Valid = true
	if err := t.Time.UnmarshalText(data); err != nil {
		return err
	}
	t.Time = t.Time.UTC()
	return nil
}

func (t Time) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if !t.Valid {
		return bsontype.Type('\x0A'), nil, nil
	}
	return bson.MarshalValue(t.Time)
}

func (t *Time) UnmarshalBSONValue(bType bsontype.Type, data []byte) error {
	if data == nil {
		t.Time, t.Valid = time.Time{}, false
		return nil
	}
	var tv time.Time
	if err := bson.UnmarshalValue(bType, data, &tv); err != nil {
		return err
	}
	t.Time = tv.UTC()
	t.Valid = true
	return nil
}

func Now() Time {
	return Time{Time: time.Now().UTC(), Valid: true}
}

func Date(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) Time {
	return Time{Time: time.Date(year, month, day, hour, min, sec, nsec, loc).UTC(), Valid: true}
}
