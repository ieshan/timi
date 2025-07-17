package timi

import (
	"database/sql"
	"database/sql/driver"
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

// IsZero reports whether t represents the zero time instant,
// January 1, year 1, 00:00:00 UTC.
func (t Time) IsZero() bool {
	return t.Time.IsZero()
}

// After reports whether the time instant t is after u.
func (t Time) After(u Time) bool {
	return t.Time.After(u.Time)
}

// Before reports whether the time instant t is before u.
func (t Time) Before(u Time) bool {
	return t.Time.Before(u.Time)
}

// Compare compares the time instant t with u. If t is before u, it returns -1;
// if t is after u, it returns +1; if they're the same, it returns 0.
func (t Time) Compare(u Time) int {
	return t.Time.Compare(u.Time)
}

// Equal reports whether t and u represent the same time instant.
// Two times can be equal even if they are in different locations.
// For example, 6:00 +0200 and 4:00 UTC are Equal.
// See the documentation on the Time type for the pitfalls of using == with
// Time values; most code should use Equal instead.
func (t Time) Equal(u Time) bool {
	return t.Time.Equal(u.Time)
}

// Date returns the year, month, and day in which t occurs.
func (t Time) Date() (year int, month time.Month, day int) {
	return t.Time.Date()
}

// Year returns the year in which t occurs.
func (t Time) Year() int {
	return t.Time.Year()
}

// Month returns the month of the year specified by t.
func (t Time) Month() time.Month {
	return t.Time.Month()
}

// Day returns the day of the month specified by t.
func (t Time) Day() int {
	return t.Time.Day()
}

// Weekday returns the day of the week specified by t.
func (t Time) Weekday() time.Weekday {
	return t.Time.Weekday()
}

// ISOWeek returns the ISO 8601 year and week number in which t occurs.
// Week ranges from 1 to 53. Jan 01 to Jan 03 of year n might belong to
// week 52 or 53 of year n-1, and Dec 29 to Dec 31 might belong to week 1
// of year n+1.
func (t Time) ISOWeek() (year, week int) {
	return t.Time.ISOWeek()
}

// Clock returns the hour, minute, and second within the day specified by t.
func (t Time) Clock() (hour, min, sec int) {
	return t.Time.Clock()
}

// Hour returns the hour within the day specified by t, in the range [0, 23].
func (t Time) Hour() int {
	return t.Time.Hour()
}

// Minute returns the minute offset within the hour specified by t, in the range [0, 59].
func (t Time) Minute() int {
	return t.Time.Minute()
}

// Second returns the second offset within the minute specified by t, in the range [0, 59].
func (t Time) Second() int {
	return t.Time.Second()
}

// Nanosecond returns the nanosecond offset within the second specified by t,
// in the range [0, 999999999].
func (t Time) Nanosecond() int {
	return t.Time.Nanosecond()
}

// YearDay returns the day of the year specified by t, in the range [1,365] for non-leap years,
// and [1,366] in leap years.
func (t Time) YearDay() int {
	return t.Time.YearDay()
}

// AddDate returns the time corresponding to adding the
// given number of years, months, and days to t.
// For example, AddDate(-1, 2, 3) applied to January 1, 2011
// returns March 4, 2010.
//
// AddDate normalizes its result in the same way that Date does,
// so, for example, adding one month to October 31 yields
// December 1, the normalized form for November 31.
func (t Time) AddDate(years int, months int, days int) Time {
	t.Time = t.Time.AddDate(years, months, days)
	return t
}

// Truncate returns the result of rounding t down to a multiple of d (since the zero time).
// If d <= 0, Truncate returns t stripped of any monotonic clock reading but otherwise unchanged.
//
// Truncate operates on the time as an absolute duration since the
// zero time; it does not operate on the presentation form of the
// time. Thus, Truncate(Hour) may return a time with a non-zero
// minute, depending on the time's Location.
func (t Time) Truncate(d time.Duration) Time {
	t.Time = t.Time.Truncate(d)
	return t
}

// Round returns the result of rounding t to the nearest multiple of d (since the zero time).
// The rounding behavior for halfway values is to round up.
// If d <= 0, Round returns t stripped of any monotonic clock reading but otherwise unchanged.
//
// Round operates on the time as an absolute duration since the
// zero time; it does not operate on the presentation form of the
// time. Thus, Round(Hour) may return a time with a non-zero
// minute, depending on the time's Location.
func (t Time) Round(d time.Duration) Time {
	t.Time = t.Time.Round(d)
	return t
}

// Add returns the time t+d.
func (t Time) Add(d time.Duration) Time {
	t.Time = t.Time.Add(d)
	return t
}

// Sub returns the duration t-u. If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration, the maximum (or minimum) duration
// will be returned.
// To compute t-d for a duration d, use t.Add(-d).
func (t Time) Sub(u Time) time.Duration {
	return t.Time.Sub(u.Time)
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
// Unix-like operating systems often record time as a 32-bit
// count of seconds, but since the method here returns a 64-bit
// value it is valid for billions of years into the past or future.
func (t Time) Unix() int64 {
	return t.Time.Unix()
}

// UnixMilli returns t as a Unix time, the number of milliseconds elapsed since
// January 1, 1970 UTC. The result is undefined if the Unix time in
// milliseconds cannot be represented by an int64 (a date more than 292 million
// years before or after 1970). The result does not depend on the
// location associated with t.
func (t Time) UnixMilli() int64 {
	return t.Time.UnixMilli()
}

// UnixMicro returns t as a Unix time, the number of microseconds elapsed since
// January 1, 1970 UTC. The result is undefined if the Unix time in
// microseconds cannot be represented by an int64 (a date before year -290307 or
// after year 294246). The result does not depend on the location associated
// with t.
func (t Time) UnixMicro() int64 {
	return t.Time.UnixMicro()
}

// UnixNano returns t as a Unix time, the number of nanoseconds elapsed
// since January 1, 1970 UTC. The result is undefined if the Unix time
// in nanoseconds cannot be represented by an int64 (a date before the year
// 1678 or after 2262). Note that this means the result of calling UnixNano
// on the zero Time is undefined. The result does not depend on the
// location associated with t.
func (t Time) UnixNano() int64 {
	return t.Time.UnixNano()
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

func Now() Time {
	return Time{Time: time.Now().UTC(), Valid: true}
}

func Date(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) Time {
	return Time{Time: time.Date(year, month, day, hour, min, sec, nsec, loc).UTC(), Valid: true}
}
