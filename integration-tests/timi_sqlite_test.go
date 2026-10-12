package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/ieshan/timi"
	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteStringScan(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	t.Run("TextColumnFormats", func(t *testing.T) {
		if _, err := db.Exec("CREATE TABLE scan_test (id INTEGER PRIMARY KEY, val TEXT)"); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		testCases := []struct {
			name     string
			input    string
			expected time.Time
		}{
			{"naive datetime", "2024-01-15 10:30:45", time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)},
			{"naive frac", "2024-01-15 10:30:45.123456", time.Date(2024, 1, 15, 10, 30, 45, 123456000, time.UTC)},
			{"mattn storage", "2024-01-15 10:30:45.123456+00:00", time.Date(2024, 1, 15, 10, 30, 45, 123456000, time.UTC)},
			{"RFC3339", "2024-01-15T10:30:45Z", time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)},
			{"offset", "2024-01-15 10:30:45+05:30", time.Date(2024, 1, 15, 5, 0, 45, 0, time.UTC)},
			{"date only", "2024-01-15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		}

		for _, tc := range testCases {
			res, err := db.Exec("INSERT INTO scan_test (val) VALUES (?)", tc.input)
			if err != nil {
				t.Fatalf("%s: insert failed: %v", tc.name, err)
			}
			id, _ := res.LastInsertId()

			var scanned timi.Time
			if err := db.QueryRow("SELECT val FROM scan_test WHERE id = ?", id).Scan(&scanned); err != nil {
				t.Fatalf("%s: scan failed: %v", tc.name, err)
			}
			if !scanned.Time.Equal(tc.expected) {
				t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, scanned.Time)
			}
			if scanned.Time.Location() != time.UTC {
				t.Errorf("%s: expected UTC, got %v", tc.name, scanned.Time.Location())
			}
		}
	})

	t.Run("ExpressionValues", func(t *testing.T) {
		// Bare expressions have no declared column type, so the driver
		// returns strings instead of time.Time.
		var scanned timi.Time
		if err := db.QueryRow("SELECT '2024-01-15 10:30:45'").Scan(&scanned); err != nil {
			t.Fatalf("string literal scan failed: %v", err)
		}
		if !scanned.Time.Equal(time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)) {
			t.Errorf("Expected 2024-01-15 10:30:45 UTC, got %v", scanned.Time)
		}

		var now timi.Time
		if err := db.QueryRow("SELECT datetime('now')").Scan(&now); err != nil {
			t.Fatalf("datetime('now') scan failed: %v", err)
		}
		if now.IsNull() {
			t.Errorf("datetime('now') should not be null")
		}
	})

	t.Run("NullAndEmpty", func(t *testing.T) {
		var nullVal timi.Time
		if err := db.QueryRow("SELECT NULL").Scan(&nullVal); err != nil {
			t.Fatalf("NULL scan failed: %v", err)
		}
		if !nullVal.IsNull() {
			t.Errorf("NULL should scan as null")
		}

		var emptyVal timi.Time
		if err := db.QueryRow("SELECT ''").Scan(&emptyVal); err != nil {
			t.Fatalf("empty string scan failed: %v", err)
		}
		if !emptyVal.IsNull() {
			t.Errorf("empty string should scan as null")
		}
	})

	t.Run("ValueRoundTrip", func(t *testing.T) {
		if _, err := db.Exec("CREATE TABLE rt (id INTEGER PRIMARY KEY, val TEXT)"); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
		stored := timi.Date(2024, 1, 15, 10, 30, 45, 123456000, time.UTC)
		res, err := db.Exec("INSERT INTO rt (val) VALUES (?)", stored)
		if err != nil {
			t.Fatalf("insert failed: %v", err)
		}
		id, _ := res.LastInsertId()

		var scanned timi.Time
		if err := db.QueryRow("SELECT val FROM rt WHERE id = ?", id).Scan(&scanned); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		if !scanned.Equal(stored) {
			t.Errorf("Round trip mismatch: expected %v, got %v", stored, scanned)
		}
	})

	t.Run("NumericRejection", func(t *testing.T) {
		var scanned timi.Time
		if err := db.QueryRow("SELECT unixepoch('now')").Scan(&scanned); err == nil {
			t.Errorf("Expected error scanning unixepoch INTEGER, got nil")
		}
		if err := db.QueryRow("SELECT julianday('now')").Scan(&scanned); err == nil {
			t.Errorf("Expected error scanning Julian REAL, got nil")
		}
	})
}
