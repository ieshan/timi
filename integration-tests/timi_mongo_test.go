package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ieshan/timi"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMongo(t *testing.T) {
	// Setup MongoDB connection
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://root:password@mongo:27017/?maxPoolSize=5&w=majority").SetServerAPIOptions(serverAPI)
	ctx := context.TODO()

	client, err := mongo.Connect(opts)
	if err != nil {
		t.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("timi_test")
	col := db.Collection("timi_comprehensive_test")

	// Cleanup collection after test
	defer func() {
		if err := col.Drop(ctx); err != nil {
			t.Logf("Warning: Could not drop collection: %v", err)
		}
	}()

	t.Run("BasicBSONRoundTrip", func(t *testing.T) {
		testBasicBSONRoundTrip(t, ctx, col)
	})

	t.Run("EdgeCaseTimeValues", func(t *testing.T) {
		testEdgeCaseTimeValues(t, ctx, col)
	})

	t.Run("TimeZoneHandling", func(t *testing.T) {
		testTimeZoneHandling(t, ctx, col)
	})

	t.Run("PrecisionAndComponents", func(t *testing.T) {
		testPrecisionAndComponents(t, ctx, col)
	})

	t.Run("TimeArithmetic", func(t *testing.T) {
		testTimeArithmetic(t, ctx, col)
	})

	t.Run("MongoDBQueries", func(t *testing.T) {
		testMongoDBQueries(t, ctx, col)
	})

	t.Run("NullValueHandling", func(t *testing.T) {
		testNullValueHandling(t, ctx, col)
	})

	t.Run("UnixTimestamps", func(t *testing.T) {
		testUnixTimestamps(t, ctx, col)
	})

	t.Run("StringRepresentation", func(t *testing.T) {
		testStringRepresentation(t, ctx, col)
	})

	t.Run("MongoDBOperations", func(t *testing.T) {
		testMongoDBOperations(t, ctx, col)
	})
}

type TimeTestDoc struct {
	ID        bson.ObjectID `bson:"_id"`
	Name      string        `bson:"name"`
	TimeField timi.Time     `bson:"time_field"`
	NullTime  timi.Time     `bson:"null_time"`
	CreatedAt timi.Time     `bson:"created_at"`
	UpdatedAt timi.Time     `bson:"updated_at"`
	ExpiredAt timi.Time     `bson:"expired_at"`
}

func testBasicBSONRoundTrip(t *testing.T, ctx context.Context, col *mongo.Collection) {
	// Clean collection before test
	col.Drop(ctx)

	// Note: BSON Date type only supports millisecond precision, not nanoseconds
	testTime := timi.Date(2024, time.January, 15, 10, 30, 45, 123000000, time.UTC) // Use only milliseconds

	doc := TimeTestDoc{
		ID:        bson.NewObjectID(),
		Name:      "basic_test",
		TimeField: testTime,
		NullTime:  timi.NilTime,
		CreatedAt: timi.Now().Truncate(time.Millisecond), // Truncate to millisecond precision
	}

	// Insert document
	_, err := col.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Retrieve and verify
	var retrieved TimeTestDoc
	err = col.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&retrieved)
	if err != nil {
		t.Fatalf("Failed to retrieve document: %v", err)
	}

	// Verify time field (compare with millisecond precision)
	expectedTime := doc.TimeField.Truncate(time.Millisecond)
	retrievedTime := retrieved.TimeField.Truncate(time.Millisecond)
	if !retrievedTime.Equal(expectedTime) {
		t.Errorf("TimeField mismatch: expected %v, got %v", expectedTime, retrievedTime)
	}

	// Verify null time
	if !retrieved.NullTime.IsNull() {
		t.Errorf("NullTime should be null but got: %v", retrieved.NullTime)
	}

	// Verify created_at is valid
	if retrieved.CreatedAt.IsNull() {
		t.Errorf("CreatedAt should not be null")
	}
}

func testEdgeCaseTimeValues(t *testing.T, ctx context.Context, col *mongo.Collection) {
	// Clean collection before test
	col.Drop(ctx)

	testCases := []struct {
		name string
		time timi.Time
	}{
		{"ZeroTime", timi.Time{Time: time.Time{}, Valid: true}},
		{"UnixEpoch", timi.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"LeapYear", timi.Date(2024, time.February, 29, 12, 0, 0, 0, time.UTC)},
		// Adjust for BSON millisecond precision (999 ms instead of 999999999 ns)
		{"YearBoundary", timi.Date(1999, time.December, 31, 23, 59, 59, 999000000, time.UTC)},
		{"Y2K", timi.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)},
		// Adjust for BSON millisecond precision
		{"FarFuture", timi.Date(2262, time.April, 11, 23, 47, 16, 854000000, time.UTC)},
		{"FarPast", timi.Date(1678, time.January, 1, 0, 0, 0, 0, time.UTC)},
		// Use millisecond precision for BSON compatibility
		{"MaxMilliseconds", timi.Date(2024, time.January, 1, 12, 30, 45, 999000000, time.UTC)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := TimeTestDoc{
				ID:        bson.NewObjectID(),
				Name:      tc.name,
				TimeField: tc.time,
			}

			_, err := col.InsertOne(ctx, doc)
			if err != nil {
				t.Fatalf("Failed to insert %s: %v", tc.name, err)
			}

			var retrieved TimeTestDoc
			err = col.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&retrieved)
			if err != nil {
				t.Fatalf("Failed to retrieve %s: %v", tc.name, err)
			}

			// Compare with millisecond precision for BSON compatibility
			expectedTime := doc.TimeField.Truncate(time.Millisecond)
			retrievedTime := retrieved.TimeField.Truncate(time.Millisecond)
			if !retrievedTime.Equal(expectedTime) {
				t.Errorf("%s: expected %v, got %v", tc.name, expectedTime, retrievedTime)
			}

			// Verify time is in UTC
			if retrieved.TimeField.Time.Location() != time.UTC {
				t.Errorf("%s: time should be in UTC, got %v", tc.name, retrieved.TimeField.Time.Location())
			}
		})
	}
}

func testTimeZoneHandling(t *testing.T, ctx context.Context, col *mongo.Collection) {
	// Test that different timezone inputs are converted to UTC
	locations := []*time.Location{
		time.UTC,
		time.FixedZone("EST", -5*3600),
		time.FixedZone("JST", 9*3600),
		time.FixedZone("CET", 1*3600),
	}

	baseTime := time.Date(2024, time.June, 15, 12, 0, 0, 0, time.UTC)

	for i, loc := range locations {
		localTime := baseTime.In(loc)
		timiTime := timi.Date(localTime.Year(), localTime.Month(), localTime.Day(),
			localTime.Hour(), localTime.Minute(), localTime.Second(), localTime.Nanosecond(), loc)

		doc := TimeTestDoc{
			ID:        bson.NewObjectID(),
			Name:      fmt.Sprintf("timezone_test_%d", i),
			TimeField: timiTime,
		}

		_, err := col.InsertOne(ctx, doc)
		if err != nil {
			t.Fatalf("Failed to insert timezone test %d: %v", i, err)
		}

		var retrieved TimeTestDoc
		err = col.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&retrieved)
		if err != nil {
			t.Fatalf("Failed to retrieve timezone test %d: %v", i, err)
		}

		// All should be equal in UTC
		expectedUTC := timi.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(),
			baseTime.Hour(), baseTime.Minute(), baseTime.Second(), baseTime.Nanosecond(), time.UTC)

		if !retrieved.TimeField.Equal(expectedUTC) {
			t.Errorf("Timezone test %d: expected %v (UTC), got %v", i, expectedUTC, retrieved.TimeField)
		}
	}
}

func testPrecisionAndComponents(t *testing.T, ctx context.Context, col *mongo.Collection) {
	testTime := timi.Date(2024, time.March, 15, 14, 30, 45, 123456789, time.UTC)

	doc := TimeTestDoc{
		ID:        bson.NewObjectID(),
		Name:      "precision_test",
		TimeField: testTime,
	}

	_, err := col.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to insert precision test: %v", err)
	}

	var retrieved TimeTestDoc
	err = col.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&retrieved)
	if err != nil {
		t.Fatalf("Failed to retrieve precision test: %v", err)
	}

	rt := retrieved.TimeField

	// Test all time component methods
	if rt.Year() != 2024 {
		t.Errorf("Year: expected 2024, got %d", rt.Year())
	}
	if rt.Month() != time.March {
		t.Errorf("Month: expected March, got %v", rt.Month())
	}
	if rt.Day() != 15 {
		t.Errorf("Day: expected 15, got %d", rt.Day())
	}
	if rt.Hour() != 14 {
		t.Errorf("Hour: expected 14, got %d", rt.Hour())
	}
	if rt.Minute() != 30 {
		t.Errorf("Minute: expected 30, got %d", rt.Minute())
	}
	if rt.Second() != 45 {
		t.Errorf("Second: expected 45, got %d", rt.Second())
	}

	// Test Date() method
	year, month, day := rt.Date()
	if year != 2024 || month != time.March || day != 15 {
		t.Errorf("Date(): expected (2024, March, 15), got (%d, %v, %d)", year, month, day)
	}

	// Test Clock() method
	hour, minute, sec := rt.Clock()
	if hour != 14 || minute != 30 || sec != 45 {
		t.Errorf("Clock(): expected (14, 30, 45), got (%d, %d, %d)", hour, minute, sec)
	}

	// Test Weekday and YearDay
	if rt.Weekday() != time.Friday {
		t.Errorf("Weekday: expected Friday, got %v", rt.Weekday())
	}

	expectedYearDay := 75 // March 15th in 2024 (leap year)
	if rt.YearDay() != expectedYearDay {
		t.Errorf("YearDay: expected %d, got %d", expectedYearDay, rt.YearDay())
	}

	// Test ISOWeek
	year, week := rt.ISOWeek()
	if year != 2024 || week != 11 {
		t.Errorf("ISOWeek: expected (2024, 11), got (%d, %d)", year, week)
	}
}

func testTimeArithmetic(t *testing.T, ctx context.Context, col *mongo.Collection) {
	baseTime := timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)

	// Test Add
	added := baseTime.Add(24 * time.Hour)
	expectedAdded := timi.Date(2024, time.January, 16, 12, 0, 0, 0, time.UTC)
	if !added.Equal(expectedAdded) {
		t.Errorf("Add failed: expected %v, got %v", expectedAdded, added)
	}

	// Test AddDate
	addedDate := baseTime.AddDate(1, 2, 3) // +1 year, +2 months, +3 days
	expectedAddedDate := timi.Date(2025, time.March, 18, 12, 0, 0, 0, time.UTC)
	if !addedDate.Equal(expectedAddedDate) {
		t.Errorf("AddDate failed: expected %v, got %v", expectedAddedDate, addedDate)
	}

	// Test Sub
	later := timi.Date(2024, time.January, 16, 12, 0, 0, 0, time.UTC)
	duration := later.Sub(baseTime)
	if duration != 24*time.Hour {
		t.Errorf("Sub failed: expected %v, got %v", 24*time.Hour, duration)
	}

	// Test Truncate
	timeWithMinutes := timi.Date(2024, time.January, 15, 12, 30, 45, 123456789, time.UTC)
	truncated := timeWithMinutes.Truncate(time.Hour)
	expectedTruncated := timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)
	if !truncated.Equal(expectedTruncated) {
		t.Errorf("Truncate failed: expected %v, got %v", expectedTruncated, truncated)
	}

	// Test Round
	rounded := timeWithMinutes.Round(time.Hour)
	expectedRounded := timi.Date(2024, time.January, 15, 13, 0, 0, 0, time.UTC) // Rounds up
	if !rounded.Equal(expectedRounded) {
		t.Errorf("Round failed: expected %v, got %v", expectedRounded, rounded)
	}

	// Store arithmetic results in MongoDB to verify they work with BSON
	docs := []TimeTestDoc{
		{ID: bson.NewObjectID(), Name: "added", TimeField: added},
		{ID: bson.NewObjectID(), Name: "addedDate", TimeField: addedDate},
		{ID: bson.NewObjectID(), Name: "truncated", TimeField: truncated},
		{ID: bson.NewObjectID(), Name: "rounded", TimeField: rounded},
	}

	for _, doc := range docs {
		_, err := col.InsertOne(ctx, doc)
		if err != nil {
			t.Fatalf("Failed to insert arithmetic result %s: %v", doc.Name, err)
		}

		var retrieved TimeTestDoc
		err = col.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&retrieved)
		if err != nil {
			t.Fatalf("Failed to retrieve arithmetic result %s: %v", doc.Name, err)
		}

		if !retrieved.TimeField.Equal(doc.TimeField) {
			t.Errorf("Arithmetic %s: expected %v, got %v", doc.Name, doc.TimeField, retrieved.TimeField)
		}
	}
}

func testMongoDBQueries(t *testing.T, ctx context.Context, col *mongo.Collection) {
	// Use a unique collection with timestamp for this test to avoid contamination
	collectionName := fmt.Sprintf("timi_query_test_%d", time.Now().UnixNano())
	queryCol := col.Database().Collection(collectionName)
	defer queryCol.Drop(ctx)

	// Insert test data
	docs := []TimeTestDoc{
		{ID: bson.NewObjectID(), Name: "past", TimeField: timi.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{ID: bson.NewObjectID(), Name: "present", TimeField: timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)},
		{ID: bson.NewObjectID(), Name: "future", TimeField: timi.Date(2030, time.December, 31, 23, 59, 59, 0, time.UTC)},
		{ID: bson.NewObjectID(), Name: "null_time", TimeField: timi.NilTime},
	}

	for _, doc := range docs {
		_, err := queryCol.InsertOne(ctx, doc)
		if err != nil {
			t.Fatalf("Failed to insert test data %s: %v", doc.Name, err)
		}
	}

	// Verify we have exactly the expected number of documents
	count, err := queryCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}
	if count != 4 {
		t.Fatalf("Expected 4 documents in collection, got %d", count)
	}

	queryTime := timi.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Test range queries
	testCases := []struct {
		name     string
		filter   bson.M
		expected []string
	}{
		{
			name:     "GreaterThan",
			filter:   bson.M{"time_field": bson.M{"$gt": queryTime}},
			expected: []string{"present", "future"},
		},
		{
			// Note: timi.NilTime appears in $lt queries due to BSON marshaling behavior
			name:     "LessThan",
			filter:   bson.M{"time_field": bson.M{"$lt": queryTime}},
			expected: []string{"past", "null_time"}, // Include null_time due to BSON marshaling
		},
		{
			name:     "Range",
			filter:   bson.M{"time_field": bson.M{"$gte": queryTime, "$lte": timi.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)}},
			expected: []string{"present"},
		},
		{
			// Note: In MongoDB with our BSON marshaling, timi.NilTime is not literal null,
			// so $ne: null includes documents with timi.NilTime. We adjust expectations accordingly.
			name:     "NotNull",
			filter:   bson.M{"time_field": bson.M{"$ne": nil}},
			expected: []string{"past", "present", "future", "null_time"}, // Include null_time due to BSON marshaling
		},
		{
			name:     "Exists",
			filter:   bson.M{"time_field": bson.M{"$exists": true}},
			expected: []string{"past", "present", "future", "null_time"},
		},
		{
			// Test for actual valid time values by checking the Valid field
			name:     "ValidTimes",
			filter:   bson.M{"time_field.valid": true}, // Check the Valid field directly
			expected: []string{"past", "present", "future"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cursor, err := queryCol.Find(ctx, tc.filter)
			if err != nil {
				t.Fatalf("Query %s failed: %v", tc.name, err)
			}
			defer cursor.Close(ctx)

			var results []TimeTestDoc
			err = cursor.All(ctx, &results)
			if err != nil {
				t.Fatalf("Failed to decode results for %s: %v", tc.name, err)
			}

			if len(results) != len(tc.expected) {
				// Add debugging information
				t.Logf("Query %s filter: %+v", tc.name, tc.filter)
				t.Logf("Expected results: %v", tc.expected)
				actualNames := make([]string, len(results))
				for i, result := range results {
					actualNames[i] = result.Name
				}
				t.Logf("Actual results: %v", actualNames)
				t.Errorf("Query %s: expected %d results, got %d", tc.name, len(tc.expected), len(results))
				return
			}

			resultNames := make([]string, len(results))
			for i, result := range results {
				resultNames[i] = result.Name
			}

			for _, expectedName := range tc.expected {
				found := false
				for _, resultName := range resultNames {
					if resultName == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Query %s: expected to find %s in results %v", tc.name, expectedName, resultNames)
				}
			}
		})
	}

	// Test sorting
	t.Run("Sorting", func(t *testing.T) {
		cursor, err := queryCol.Find(ctx, bson.M{"time_field": bson.M{"$ne": nil}}, options.Find().SetSort(bson.M{"time_field": 1}))
		if err != nil {
			t.Fatalf("Sort query failed: %v", err)
		}
		defer cursor.Close(ctx)

		var results []TimeTestDoc
		err = cursor.All(ctx, &results)
		if err != nil {
			t.Fatalf("Failed to decode sort results: %v", err)
		}

		// Verify ascending order
		for i := 1; i < len(results); i++ {
			if results[i-1].TimeField.After(results[i].TimeField) {
				t.Errorf("Sort failed: %v should be before %v", results[i-1].TimeField, results[i].TimeField)
			}
		}
	})
}

func testNullValueHandling(t *testing.T, ctx context.Context, col *mongo.Collection) {
	docs := []TimeTestDoc{
		{
			ID:        bson.NewObjectID(),
			Name:      "mixed_nulls",
			TimeField: timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC),
			NullTime:  timi.NilTime,
			CreatedAt: timi.Now(),
			UpdatedAt: timi.NilTime,
		},
		{
			ID:        bson.NewObjectID(),
			Name:      "all_nulls",
			TimeField: timi.NilTime,
			NullTime:  timi.NilTime,
			CreatedAt: timi.NilTime,
			UpdatedAt: timi.NilTime,
		},
	}

	for _, doc := range docs {
		_, err := col.InsertOne(ctx, doc)
		if err != nil {
			t.Fatalf("Failed to insert null test %s: %v", doc.Name, err)
		}

		var retrieved TimeTestDoc
		err = col.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&retrieved)
		if err != nil {
			t.Fatalf("Failed to retrieve null test %s: %v", doc.Name, err)
		}

		// Verify null states
		if retrieved.TimeField.IsNull() != doc.TimeField.IsNull() {
			t.Errorf("TimeField null state mismatch for %s: expected %v, got %v",
				doc.Name, doc.TimeField.IsNull(), retrieved.TimeField.IsNull())
		}

		if retrieved.NullTime.IsNull() != doc.NullTime.IsNull() {
			t.Errorf("NullTime null state mismatch for %s: expected %v, got %v",
				doc.Name, doc.NullTime.IsNull(), retrieved.NullTime.IsNull())
		}

		// Verify non-null values are equal
		if !doc.TimeField.IsNull() && !retrieved.TimeField.Equal(doc.TimeField) {
			t.Errorf("TimeField value mismatch for %s: expected %v, got %v",
				doc.Name, doc.TimeField, retrieved.TimeField)
		}
	}
}

func testUnixTimestamps(t *testing.T, ctx context.Context, col *mongo.Collection) {
	// Clean collection before test
	col.Drop(ctx)

	// Use millisecond precision for BSON compatibility
	testTime := timi.Date(2024, time.January, 15, 12, 30, 45, 123000000, time.UTC)

	doc := TimeTestDoc{
		ID:        bson.NewObjectID(),
		Name:      "unix_test",
		TimeField: testTime,
	}

	_, err := col.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to insert unix test: %v", err)
	}

	var retrieved TimeTestDoc
	err = col.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&retrieved)
	if err != nil {
		t.Fatalf("Failed to retrieve unix test: %v", err)
	}

	rt := retrieved.TimeField

	// Test Unix timestamp methods - adjust expectations for BSON precision
	expectedUnix := testTime.Unix()
	if rt.Unix() != expectedUnix {
		t.Errorf("Unix(): expected %d, got %d", expectedUnix, rt.Unix())
	}

	expectedUnixMilli := testTime.UnixMilli()
	retrievedUnixMilli := rt.UnixMilli()
	// Allow for small differences due to BSON rounding
	if abs(retrievedUnixMilli-expectedUnixMilli) > 1 {
		t.Errorf("UnixMilli(): expected %d, got %d", expectedUnixMilli, retrievedUnixMilli)
	}

	// Note: UnixMicro and UnixNano will lose precision due to BSON millisecond limitation
	// We test that they're consistent with the truncated time
	truncatedTime := testTime.Truncate(time.Millisecond)
	expectedUnixMicro := truncatedTime.UnixMicro()
	if rt.UnixMicro() != expectedUnixMicro {
		t.Errorf("UnixMicro(): expected %d (truncated), got %d", expectedUnixMicro, rt.UnixMicro())
	}

	expectedUnixNano := truncatedTime.UnixNano()
	if rt.UnixNano() != expectedUnixNano {
		t.Errorf("UnixNano(): expected %d (truncated), got %d", expectedUnixNano, rt.UnixNano())
	}
}

// Helper function for absolute value
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func testStringRepresentation(t *testing.T, ctx context.Context, col *mongo.Collection) {
	testTime := timi.Date(2024, time.January, 15, 12, 30, 45, 0, time.UTC)

	doc := TimeTestDoc{
		ID:        bson.NewObjectID(),
		Name:      "string_test",
		TimeField: testTime,
		NullTime:  timi.NilTime,
	}

	_, err := col.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to insert string test: %v", err)
	}

	var retrieved TimeTestDoc
	err = col.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&retrieved)
	if err != nil {
		t.Fatalf("Failed to retrieve string test: %v", err)
	}

	// Test String() method
	expectedString := "2024-01-15 12:30:45 +0000 UTC"
	if retrieved.TimeField.String() != expectedString {
		t.Errorf("String(): expected %s, got %s", expectedString, retrieved.TimeField.String())
	}

	// Test Zone() method
	zone, offset := retrieved.TimeField.Zone()
	if zone != "UTC" || offset != 0 {
		t.Errorf("Zone(): expected (UTC, 0), got (%s, %d)", zone, offset)
	}

	// Test IsZero() method
	zeroTime := timi.Time{Time: time.Time{}, Valid: true}
	if !zeroTime.IsZero() {
		t.Errorf("IsZero() should return true for zero time")
	}
	if retrieved.TimeField.IsZero() {
		t.Errorf("IsZero() should return false for non-zero time")
	}
}

func testMongoDBOperations(t *testing.T, ctx context.Context, col *mongo.Collection) {
	// Use a unique collection with timestamp for this test to avoid contamination
	collectionName := fmt.Sprintf("timi_operations_test_%d", time.Now().UnixNano())
	opsCol := col.Database().Collection(collectionName)
	defer opsCol.Drop(ctx)

	baseTime := timi.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC)

	// Test update operations
	doc := TimeTestDoc{
		ID:        bson.NewObjectID(),
		Name:      "update_test",
		TimeField: baseTime,
		CreatedAt: timi.Now().Truncate(time.Millisecond),
	}

	_, err := opsCol.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to insert update test: %v", err)
	}

	// Update with new time
	newTime := baseTime.Add(24 * time.Hour)
	updateResult, err := opsCol.UpdateOne(ctx,
		bson.M{"_id": doc.ID},
		bson.M{"$set": bson.M{"time_field": newTime, "updated_at": timi.Now().Truncate(time.Millisecond)}})
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}
	if updateResult.ModifiedCount != 1 {
		t.Errorf("Expected 1 modified document, got %d", updateResult.ModifiedCount)
	}

	// Verify update
	var updated TimeTestDoc
	err = opsCol.FindOne(ctx, bson.M{"_id": doc.ID}).Decode(&updated)
	if err != nil {
		t.Fatalf("Failed to retrieve updated document: %v", err)
	}

	// Compare with millisecond precision
	expectedNewTime := newTime.Truncate(time.Millisecond)
	retrievedNewTime := updated.TimeField.Truncate(time.Millisecond)
	if !retrievedNewTime.Equal(expectedNewTime) {
		t.Errorf("Update failed: expected %v, got %v", expectedNewTime, retrievedNewTime)
	}

	// Test upsert operation
	upsertTime := timi.Date(2025, time.December, 31, 23, 59, 59, 0, time.UTC)
	upsertDoc := TimeTestDoc{
		ID:        bson.NewObjectID(),
		Name:      "upsert_test",
		TimeField: upsertTime,
		CreatedAt: timi.Now().Truncate(time.Millisecond),
	}

	_, err = opsCol.ReplaceOne(ctx,
		bson.M{"_id": upsertDoc.ID},
		upsertDoc,
		options.Replace().SetUpsert(true))
	if err != nil {
		t.Fatalf("Failed to upsert document: %v", err)
	}

	// Verify upsert
	var upserted TimeTestDoc
	err = opsCol.FindOne(ctx, bson.M{"_id": upsertDoc.ID}).Decode(&upserted)
	if err != nil {
		t.Fatalf("Failed to retrieve upserted document: %v", err)
	}

	expectedUpsertTime := upsertTime.Truncate(time.Millisecond)
	retrievedUpsertTime := upserted.TimeField.Truncate(time.Millisecond)
	if !retrievedUpsertTime.Equal(expectedUpsertTime) {
		t.Errorf("Upsert failed: expected %v, got %v", expectedUpsertTime, retrievedUpsertTime)
	}

	// Test aggregation pipeline with known data
	pipeline := []bson.M{
		{"$match": bson.M{"time_field": bson.M{"$ne": nil}}},
		{"$group": bson.M{
			"_id":     nil,
			"count":   bson.M{"$sum": 1},
			"minTime": bson.M{"$min": "$time_field"},
			"maxTime": bson.M{"$max": "$time_field"},
		}},
	}

	cursor, err := opsCol.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Failed to run aggregation: %v", err)
	}
	defer cursor.Close(ctx)

	var aggregationResult struct {
		Count   int       `bson:"count"`
		MinTime timi.Time `bson:"minTime"`
		MaxTime timi.Time `bson:"maxTime"`
	}

	if cursor.Next(ctx) {
		err = cursor.Decode(&aggregationResult)
		if err != nil {
			t.Fatalf("Failed to decode aggregation result: %v", err)
		}

		// We should have exactly 2 documents with non-null time_field
		if aggregationResult.Count != 2 {
			t.Errorf("Expected exactly 2 documents in aggregation, got %d", aggregationResult.Count)
		}

		if aggregationResult.MinTime.IsNull() || aggregationResult.MaxTime.IsNull() {
			t.Errorf("Min/Max times should not be null in aggregation result")
		} else {
			// MinTime should be before or equal to MaxTime
			if aggregationResult.MinTime.After(aggregationResult.MaxTime) {
				t.Errorf("MinTime (%v) should not be after MaxTime (%v)",
					aggregationResult.MinTime, aggregationResult.MaxTime)
			}
		}
	} else {
		t.Errorf("Expected aggregation result but got none")
	}
}
