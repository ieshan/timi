package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ieshan/timi"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TimeTestSqlStruct struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	TimeField timi.Time `gorm:"column:time_field"`
	NullTime  timi.Time `gorm:"column:null_time"`
	CreatedAt timi.Time `gorm:"column:created_at"`
	UpdatedAt timi.Time `gorm:"column:updated_at"`
}

func (TimeTestSqlStruct) TableName() string {
	return "timi_test"
}

// Database configuration for each supported database
type dbConfig struct {
	name           string
	setupDSN       string
	connectDSN     string
	createTableSQL string
	setupFunc      func(*gorm.DB) error
	cleanupFunc    func(*gorm.DB) error
}

func TestMySQL(t *testing.T) {
	config := dbConfig{
		name:       "MySQL",
		setupDSN:   "root:password@tcp(mariadb:3306)/?charset=utf8mb4&parseTime=True&loc=UTC",
		connectDSN: "root:password@tcp(mariadb:3306)/timi_test?charset=utf8mb4&parseTime=True&loc=UTC",
		createTableSQL: `
			CREATE TABLE IF NOT EXISTS timi_test (
				id BIGINT NOT NULL AUTO_INCREMENT,
				name VARCHAR(255) NOT NULL,
				time_field DATETIME(6) NOT NULL,
				null_time DATETIME(6) DEFAULT NULL,
				created_at DATETIME(6) DEFAULT NULL,
				updated_at DATETIME(6) DEFAULT NULL,
				PRIMARY KEY (id)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		setupFunc: func(db *gorm.DB) error {
			return db.Exec("CREATE DATABASE IF NOT EXISTS `timi_test` COLLATE 'utf8mb4_unicode_ci';").Error
		},
		cleanupFunc: func(db *gorm.DB) error {
			if err := db.Exec("DROP TABLE IF EXISTS `timi_test`;").Error; err != nil {
				return err
			}
			return db.Exec("DROP DATABASE IF EXISTS `timi_test`;").Error
		},
	}

	testDatabase(t, config, func() (*gorm.DB, error) {
		return gorm.Open(mysql.Open(config.setupDSN), &gorm.Config{})
	}, func() (*gorm.DB, error) {
		return gorm.Open(mysql.Open(config.connectDSN), &gorm.Config{})
	})
}

func TestPostgres(t *testing.T) {
	config := dbConfig{
		name:       "PostgreSQL",
		setupDSN:   "host=postgres user=postgres password=password port=5432 sslmode=disable",
		connectDSN: "host=postgres user=postgres password=password dbname=timi_test port=5432 sslmode=disable TimeZone=UTC",
		createTableSQL: `
			CREATE TABLE IF NOT EXISTS timi_test (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				time_field TIMESTAMPTZ NOT NULL,
				null_time TIMESTAMPTZ DEFAULT NULL,
				created_at TIMESTAMPTZ DEFAULT NULL,
				updated_at TIMESTAMPTZ DEFAULT NULL
			);`,
		setupFunc: func(db *gorm.DB) error {
			return db.Exec("CREATE DATABASE timi_test;").Error
		},
		cleanupFunc: func(db *gorm.DB) error {
			if err := db.Exec("DROP TABLE IF EXISTS timi_test;").Error; err != nil {
				return err
			}
			sqlDB, _ := db.DB()
			_ = sqlDB.Close()
			// Note: We need to connect to the setup DB to drop the database
			setupDB, err := gorm.Open(postgres.Open("host=postgres user=postgres password=password port=5432 sslmode=disable"), &gorm.Config{})
			if err != nil {
				return err
			}
			return setupDB.Exec("DROP DATABASE IF EXISTS timi_test;").Error
		},
	}

	testDatabase(t, config, func() (*gorm.DB, error) {
		return gorm.Open(postgres.Open(config.setupDSN), &gorm.Config{})
	}, func() (*gorm.DB, error) {
		return gorm.Open(postgres.Open(config.connectDSN), &gorm.Config{})
	})
}

func TestSQLite(t *testing.T) {
	config := dbConfig{
		name:       "SQLite",
		setupDSN:   ":memory:",
		connectDSN: ":memory:",
		createTableSQL: `
			CREATE TABLE IF NOT EXISTS timi_test (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				time_field DATETIME NOT NULL,
				null_time DATETIME DEFAULT NULL,
				created_at DATETIME DEFAULT NULL,
				updated_at DATETIME DEFAULT NULL
			);`,
		setupFunc:   func(db *gorm.DB) error { return nil }, // No setup needed for in-memory
		cleanupFunc: func(db *gorm.DB) error { return nil }, // No cleanup needed for in-memory
	}

	testDatabase(t, config, func() (*gorm.DB, error) {
		return gorm.Open(sqlite.Open(config.setupDSN), &gorm.Config{})
	}, func() (*gorm.DB, error) {
		return gorm.Open(sqlite.Open(config.connectDSN), &gorm.Config{})
	})
}

func testDatabase(t *testing.T, config dbConfig, setupDBFunc, connectDBFunc func() (*gorm.DB, error)) {
	// Setup database if needed
	if config.setupFunc != nil {
		setupDB, err := setupDBFunc()
		if err != nil {
			t.Fatalf("%s setup connection error: %v", config.name, err)
		}
		if err = config.setupFunc(setupDB); err != nil {
			t.Fatalf("%s setup error: %v", config.name, err)
		}
	}

	// Connect to the actual test database
	db, err := connectDBFunc()
	if err != nil {
		t.Fatalf("%s connection error: %v", config.name, err)
	}

	// Get underlying SQL DB for cleanup
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("%s DB connection error: %v", config.name, err)
	}
	defer sqlDB.Close()

	// Cleanup
	defer func() {
		if config.cleanupFunc != nil {
			if err := config.cleanupFunc(db); err != nil {
				t.Logf("Warning: %s cleanup failed: %v", config.name, err)
			}
		}
	}()

	// Create table
	if err = db.Exec(config.createTableSQL).Error; err != nil {
		t.Fatalf("%s table creation error: %v", config.name, err)
	}

	// Run comprehensive tests
	t.Run("BasicRoundTrip", func(t *testing.T) {
		testSQLBasicRoundTrip(t, db, config.name)
	})

	t.Run("EdgeCaseTimeValues", func(t *testing.T) {
		testSQLEdgeCaseTimeValues(t, db, config.name)
	})

	t.Run("TimeZoneHandling", func(t *testing.T) {
		testSQLTimeZoneHandling(t, db, config.name)
	})

	t.Run("NullValueHandling", func(t *testing.T) {
		testSQLNullValueHandling(t, db, config.name)
	})

	t.Run("TimeArithmetic", func(t *testing.T) {
		testSQLTimeArithmetic(t, db, config.name)
	})

	t.Run("SQLQueries", func(t *testing.T) {
		testSQLQueries(t, db, config.name)
	})

	t.Run("CRUDOperations", func(t *testing.T) {
		testCRUDOperations(t, db, config.name)
	})

	t.Run("ComponentMethods", func(t *testing.T) {
		testSQLComponentMethods(t, db, config.name)
	})
}

func testSQLBasicRoundTrip(t *testing.T, db *gorm.DB, dbName string) {
	// Clear table
	db.Exec("DELETE FROM timi_test")

	testTime := timi.Date(2024, time.January, 15, 10, 30, 45, 123456000, time.UTC)

	record := TimeTestSqlStruct{
		Name:      "basic_test",
		TimeField: testTime,
		NullTime:  timi.NilTime,
		CreatedAt: timi.Now(),
	}

	// Create record
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("%s: Failed to create record: %v", dbName, err)
	}

	// Retrieve record
	var retrieved TimeTestSqlStruct
	if err := db.First(&retrieved, "name = ?", "basic_test").Error; err != nil {
		t.Fatalf("%s: Failed to retrieve record: %v", dbName, err)
	}

	// Verify time field
	if !retrieved.TimeField.Equal(testTime) {
		t.Errorf("%s: TimeField mismatch: expected %v, got %v", dbName, testTime, retrieved.TimeField)
	}

	// Verify null time
	if !retrieved.NullTime.IsNull() {
		t.Errorf("%s: NullTime should be null but got: %v", dbName, retrieved.NullTime)
	}

	// Verify created_at is valid
	if retrieved.CreatedAt.IsNull() {
		t.Errorf("%s: CreatedAt should not be null", dbName)
	}

	// Verify timezone is UTC
	if retrieved.TimeField.Time.Location() != time.UTC {
		t.Errorf("%s: Time should be in UTC, got %v", dbName, retrieved.TimeField.Time.Location())
	}
}

func testSQLEdgeCaseTimeValues(t *testing.T, db *gorm.DB, dbName string) {
	testCases := []struct {
		name string
		time timi.Time
	}{
		{"ZeroTime", timi.Time{Time: time.Time{}, Valid: true}},
		{"UnixEpoch", timi.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"LeapYear", timi.Date(2024, time.February, 29, 12, 0, 0, 0, time.UTC)},
		{"YearBoundary", timi.Date(1999, time.December, 31, 23, 59, 59, 999000000, time.UTC)},
		{"Y2K", timi.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"FarFuture", timi.Date(2099, time.December, 31, 23, 59, 59, 0, time.UTC)},
		{"FarPast", timi.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"HighPrecision", timi.Date(2024, time.January, 1, 12, 30, 45, 999999000, time.UTC)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear table
			db.Exec("DELETE FROM timi_test")

			record := TimeTestSqlStruct{
				Name:      tc.name,
				TimeField: tc.time,
				NullTime:  timi.NilTime,
			}

			if err := db.Create(&record).Error; err != nil {
				t.Fatalf("%s: Failed to insert %s: %v", dbName, tc.name, err)
			}

			var retrieved TimeTestSqlStruct
			if err := db.First(&retrieved, "name = ?", tc.name).Error; err != nil {
				t.Fatalf("%s: Failed to retrieve %s: %v", dbName, tc.name, err)
			}

			// For SQL databases, we expect some precision loss might occur
			// but the times should be approximately equal
			if !timesApproximatelyEqual(retrieved.TimeField, tc.time) {
				t.Errorf("%s %s: expected %v, got %v", dbName, tc.name, tc.time, retrieved.TimeField)
			}

			// Verify time is in UTC
			if retrieved.TimeField.Time.Location() != time.UTC {
				t.Errorf("%s %s: time should be in UTC, got %v", dbName, tc.name, retrieved.TimeField.Time.Location())
			}
		})
	}
}

func testSQLTimeZoneHandling(t *testing.T, db *gorm.DB, dbName string) {
	// Test that different timezone inputs are converted to UTC
	locations := []*time.Location{
		time.UTC,
		time.FixedZone("EST", -5*3600),
		time.FixedZone("JST", 9*3600),
		time.FixedZone("CET", 1*3600),
	}

	baseTime := time.Date(2024, time.June, 15, 12, 0, 0, 0, time.UTC)

	for i, loc := range locations {
		// Clear table
		db.Exec("DELETE FROM timi_test")

		localTime := baseTime.In(loc)
		timiTime := timi.Date(localTime.Year(), localTime.Month(), localTime.Day(),
			localTime.Hour(), localTime.Minute(), localTime.Second(), localTime.Nanosecond(), loc)

		record := TimeTestSqlStruct{
			Name:      fmt.Sprintf("timezone_test_%d", i),
			TimeField: timiTime,
		}

		if err := db.Create(&record).Error; err != nil {
			t.Fatalf("%s: Failed to insert timezone test %d: %v", dbName, i, err)
		}

		var retrieved TimeTestSqlStruct
		if err := db.First(&retrieved, "name = ?", record.Name).Error; err != nil {
			t.Fatalf("%s: Failed to retrieve timezone test %d: %v", dbName, i, err)
		}

		// All should be equal in UTC
		expectedUTC := timi.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(),
			baseTime.Hour(), baseTime.Minute(), baseTime.Second(), 0, time.UTC) // SQL may lose nanosecond precision

		if !timesApproximatelyEqual(retrieved.TimeField, expectedUTC) {
			t.Errorf("%s: Timezone test %d: expected %v (UTC), got %v", dbName, i, expectedUTC, retrieved.TimeField)
		}
	}
}

func testSQLNullValueHandling(t *testing.T, db *gorm.DB, dbName string) {
	testCases := []struct {
		name      string
		timeField timi.Time
		nullTime  timi.Time
	}{
		{
			name:      "mixed_nulls",
			timeField: timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC),
			nullTime:  timi.NilTime,
		},
		{
			name:      "all_nulls",
			timeField: timi.NilTime,
			nullTime:  timi.NilTime,
		},
		{
			name:      "no_nulls",
			timeField: timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC),
			nullTime:  timi.Date(2024, time.January, 16, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear table
			db.Exec("DELETE FROM timi_test")

			record := TimeTestSqlStruct{
				Name:      tc.name,
				TimeField: tc.timeField,
				NullTime:  tc.nullTime,
			}

			// Handle the case where timeField is null - we need to handle this specially for NOT NULL columns
			if tc.timeField.IsNull() {
				// For NOT NULL columns, we'll test by trying to insert and expecting an error
				err := db.Create(&record).Error
				if err == nil {
					t.Errorf("%s %s: Expected error when inserting null into NOT NULL column", dbName, tc.name)
					return
				}
				// This is expected behavior - null values cannot be inserted into NOT NULL columns
				return
			}

			if err := db.Create(&record).Error; err != nil {
				t.Fatalf("%s: Failed to insert null test %s: %v", dbName, tc.name, err)
			}

			var retrieved TimeTestSqlStruct
			if err := db.First(&retrieved, "name = ?", tc.name).Error; err != nil {
				t.Fatalf("%s: Failed to retrieve null test %s: %v", dbName, tc.name, err)
			}

			// Verify null states
			if retrieved.TimeField.IsNull() != tc.timeField.IsNull() {
				t.Errorf("%s %s: TimeField null state mismatch: expected %v, got %v",
					dbName, tc.name, tc.timeField.IsNull(), retrieved.TimeField.IsNull())
			}

			if retrieved.NullTime.IsNull() != tc.nullTime.IsNull() {
				t.Errorf("%s %s: NullTime null state mismatch: expected %v, got %v",
					dbName, tc.name, tc.nullTime.IsNull(), retrieved.NullTime.IsNull())
			}

			// Verify non-null values are approximately equal
			if !tc.timeField.IsNull() && !timesApproximatelyEqual(retrieved.TimeField, tc.timeField) {
				t.Errorf("%s %s: TimeField value mismatch: expected %v, got %v",
					dbName, tc.name, tc.timeField, retrieved.TimeField)
			}

			if !tc.nullTime.IsNull() && !timesApproximatelyEqual(retrieved.NullTime, tc.nullTime) {
				t.Errorf("%s %s: NullTime value mismatch: expected %v, got %v",
					dbName, tc.name, tc.nullTime, retrieved.NullTime)
			}
		})
	}
}

func testSQLTimeArithmetic(t *testing.T, db *gorm.DB, dbName string) {
	baseTime := timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)

	// Test Add
	added := baseTime.Add(24 * time.Hour)
	expectedAdded := timi.Date(2024, time.January, 16, 12, 0, 0, 0, time.UTC)
	if !added.Equal(expectedAdded) {
		t.Errorf("%s: Add failed: expected %v, got %v", dbName, expectedAdded, added)
	}

	// Test AddDate
	addedDate := baseTime.AddDate(1, 2, 3) // +1 year, +2 months, +3 days
	expectedAddedDate := timi.Date(2025, time.March, 18, 12, 0, 0, 0, time.UTC)
	if !addedDate.Equal(expectedAddedDate) {
		t.Errorf("%s: AddDate failed: expected %v, got %v", dbName, expectedAddedDate, addedDate)
	}

	// Test Sub
	later := timi.Date(2024, time.January, 16, 12, 0, 0, 0, time.UTC)
	duration := later.Sub(baseTime)
	if duration != 24*time.Hour {
		t.Errorf("%s: Sub failed: expected %v, got %v", dbName, 24*time.Hour, duration)
	}

	// Test Truncate
	timeWithMinutes := timi.Date(2024, time.January, 15, 12, 30, 45, 123456789, time.UTC)
	truncated := timeWithMinutes.Truncate(time.Hour)
	expectedTruncated := timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)
	if !truncated.Equal(expectedTruncated) {
		t.Errorf("%s: Truncate failed: expected %v, got %v", dbName, expectedTruncated, truncated)
	}

	// Store arithmetic results in database to verify they work with SQL
	testCases := []struct {
		name string
		time timi.Time
	}{
		{"added", added},
		{"addedDate", addedDate},
		{"truncated", truncated},
	}

	for _, tc := range testCases {
		// Clear table
		db.Exec("DELETE FROM timi_test")

		record := TimeTestSqlStruct{
			Name:      tc.name,
			TimeField: tc.time,
		}

		if err := db.Create(&record).Error; err != nil {
			t.Fatalf("%s: Failed to insert arithmetic result %s: %v", dbName, tc.name, err)
		}

		var retrieved TimeTestSqlStruct
		if err := db.First(&retrieved, "name = ?", tc.name).Error; err != nil {
			t.Fatalf("%s: Failed to retrieve arithmetic result %s: %v", dbName, tc.name, err)
		}

		if !timesApproximatelyEqual(retrieved.TimeField, tc.time) {
			t.Errorf("%s: Arithmetic %s: expected %v, got %v", dbName, tc.name, tc.time, retrieved.TimeField)
		}
	}
}

func testSQLQueries(t *testing.T, db *gorm.DB, dbName string) {
	// Clear table and insert test data
	db.Exec("DELETE FROM timi_test")

	testData := []TimeTestSqlStruct{
		{Name: "past", TimeField: timi.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "present", TimeField: timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)},
		{Name: "future", TimeField: timi.Date(2030, time.December, 31, 23, 59, 59, 0, time.UTC)},
		{Name: "null_time", TimeField: timi.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), NullTime: timi.NilTime},
	}

	for _, record := range testData {
		if err := db.Create(&record).Error; err != nil {
			t.Fatalf("%s: Failed to insert test data %s: %v", dbName, record.Name, err)
		}
	}

	queryTime := timi.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Test range queries
	t.Run("GreaterThan", func(t *testing.T) {
		var results []TimeTestSqlStruct
		err := db.Where("time_field > ?", queryTime).Find(&results).Error
		if err != nil {
			t.Fatalf("%s: GreaterThan query failed: %v", dbName, err)
		}
		if len(results) != 2 { // present, future
			t.Errorf("%s: GreaterThan expected 2 results, got %d", dbName, len(results))
		}
	})

	t.Run("LessThan", func(t *testing.T) {
		var results []TimeTestSqlStruct
		err := db.Where("time_field < ?", queryTime).Find(&results).Error
		if err != nil {
			t.Fatalf("%s: LessThan query failed: %v", dbName, err)
		}
		if len(results) != 1 { // past
			t.Errorf("%s: LessThan expected 1 result, got %d", dbName, len(results))
		}
	})

	t.Run("NullChecks", func(t *testing.T) {
		var nullResults []TimeTestSqlStruct
		err := db.Where("null_time IS NULL").Find(&nullResults).Error
		if err != nil {
			t.Fatalf("%s: NULL check query failed: %v", dbName, err)
		}
		if len(nullResults) != 4 { // All records have null_time as NULL
			t.Errorf("%s: NULL check expected 4 results, got %d", dbName, len(nullResults))
		}
	})

	t.Run("Ordering", func(t *testing.T) {
		var results []TimeTestSqlStruct
		err := db.Order("time_field ASC").Find(&results).Error
		if err != nil {
			t.Fatalf("%s: Ordering query failed: %v", dbName, err)
		}

		// Verify ascending order
		for i := 1; i < len(results); i++ {
			if results[i-1].TimeField.After(results[i].TimeField) {
				t.Errorf("%s: Ordering failed: %v should be before %v", dbName, results[i-1].TimeField, results[i].TimeField)
			}
		}
	})
}

func testCRUDOperations(t *testing.T, db *gorm.DB, dbName string) {
	// Clear table
	db.Exec("DELETE FROM timi_test")

	// Test Create
	originalTime := timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)
	record := TimeTestSqlStruct{
		Name:      "crud_test",
		TimeField: originalTime,
		CreatedAt: timi.Now(),
	}

	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("%s: Create failed: %v", dbName, err)
	}

	// Test Read
	var retrieved TimeTestSqlStruct
	if err := db.First(&retrieved, "name = ?", "crud_test").Error; err != nil {
		t.Fatalf("%s: Read failed: %v", dbName, err)
	}

	// Test Update
	newTime := originalTime.Add(24 * time.Hour)
	updateTime := timi.Now()
	if err := db.Model(&retrieved).Updates(TimeTestSqlStruct{
		TimeField: newTime,
		UpdatedAt: updateTime,
	}).Error; err != nil {
		t.Fatalf("%s: Update failed: %v", dbName, err)
	}

	// Verify update
	var updated TimeTestSqlStruct
	if err := db.First(&updated, "name = ?", "crud_test").Error; err != nil {
		t.Fatalf("%s: Read after update failed: %v", dbName, err)
	}

	if !timesApproximatelyEqual(updated.TimeField, newTime) {
		t.Errorf("%s: Update verification failed: expected %v, got %v", dbName, newTime, updated.TimeField)
	}

	// Test Delete
	if err := db.Delete(&updated).Error; err != nil {
		t.Fatalf("%s: Delete failed: %v", dbName, err)
	}

	// Verify deletion
	var count int64
	db.Model(&TimeTestSqlStruct{}).Where("name = ?", "crud_test").Count(&count)
	if count != 0 {
		t.Errorf("%s: Delete verification failed: expected 0 records, got %d", dbName, count)
	}
}

func testSQLComponentMethods(t *testing.T, db *gorm.DB, dbName string) {
	// Clear table
	db.Exec("DELETE FROM timi_test")

	testTime := timi.Date(2024, time.March, 15, 14, 30, 45, 123456789, time.UTC)
	record := TimeTestSqlStruct{
		Name:      "component_test",
		TimeField: testTime,
	}

	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("%s: Failed to insert component test: %v", dbName, err)
	}

	var retrieved TimeTestSqlStruct
	if err := db.First(&retrieved, "name = ?", "component_test").Error; err != nil {
		t.Fatalf("%s: Failed to retrieve component test: %v", dbName, err)
	}

	rt := retrieved.TimeField

	// Test time component methods
	if rt.Year() != 2024 {
		t.Errorf("%s: Year: expected 2024, got %d", dbName, rt.Year())
	}
	if rt.Month() != time.March {
		t.Errorf("%s: Month: expected March, got %v", dbName, rt.Month())
	}
	if rt.Day() != 15 {
		t.Errorf("%s: Day: expected 15, got %d", dbName, rt.Day())
	}
	if rt.Hour() != 14 {
		t.Errorf("%s: Hour: expected 14, got %d", dbName, rt.Hour())
	}
	if rt.Minute() != 30 {
		t.Errorf("%s: Minute: expected 30, got %d", dbName, rt.Minute())
	}
	if rt.Second() != 45 {
		t.Errorf("%s: Second: expected 45, got %d", dbName, rt.Second())
	}

	// Test Date() method
	year, month, day := rt.Date()
	if year != 2024 || month != time.March || day != 15 {
		t.Errorf("%s: Date(): expected (2024, March, 15), got (%d, %v, %d)", dbName, year, month, day)
	}

	// Test Clock() method
	hour, minute, sec := rt.Clock()
	if hour != 14 || minute != 30 || sec != 45 {
		t.Errorf("%s: Clock(): expected (14, 30, 45), got (%d, %d, %d)", dbName, hour, minute, sec)
	}

	// Test Weekday
	if rt.Weekday() != time.Friday {
		t.Errorf("%s: Weekday: expected Friday, got %v", dbName, rt.Weekday())
	}

	// Test YearDay
	expectedYearDay := 75 // March 15th in 2024 (leap year)
	if rt.YearDay() != expectedYearDay {
		t.Errorf("%s: YearDay: expected %d, got %d", dbName, expectedYearDay, rt.YearDay())
	}

	// Test String representation - SQL databases may preserve different precision
	actualString := rt.String()
	// Check that the string contains the expected date and time parts
	expectedPrefix := "2024-03-15 14:30:45"
	expectedSuffix := "+0000 UTC"
	if !strings.Contains(actualString, expectedPrefix) || !strings.Contains(actualString, expectedSuffix) {
		t.Errorf("%s: String(): expected to contain '%s' and '%s', got %s", dbName, expectedPrefix, expectedSuffix, actualString)
	}

	// Test Zone
	zone, offset := rt.Zone()
	if zone != "UTC" || offset != 0 {
		t.Errorf("%s: Zone(): expected (UTC, 0), got (%s, %d)", dbName, zone, offset)
	}

	// Test Unix timestamps
	expectedUnix := testTime.Unix()
	if rt.Unix() != expectedUnix {
		t.Errorf("%s: Unix(): expected %d, got %d", dbName, expectedUnix, rt.Unix())
	}
}

// Helper function to compare times with some tolerance for database precision differences
func timesApproximatelyEqual(t1, t2 timi.Time) bool {
	if t1.IsNull() && t2.IsNull() {
		return true
	}
	if t1.IsNull() || t2.IsNull() {
		return false
	}

	// Allow up to 1-millisecond difference to account for database precision
	diff := t1.Sub(t2)
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Millisecond
}
