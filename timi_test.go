package timi

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
			fmt.Println(unmVal.Time2, timeVals[index].Time2, jsonStr)
			t.Fatalf("Case %d: Time2 is null mismatch (%v : %v)", index, unmVal.Time2.IsNull(), timeVals[index].Time2.IsNull())
		}
		if !unmVal.Time2.IsNull() && !unmVal.Time2.Equal(timeVals[index].Time2) {
			t.Fatalf("Case %d: Original Time (%v) did not match with the Time from JSON %v", index, timeVals[index].Time2, unmVal.Time2)
		}
	}
}

func TestMongo(t *testing.T) {
	type TimeTestStruct struct {
		ID    primitive.ObjectID `bson:"_id"`
		Time1 Time               `bson:"time1"`
		Time2 Time               `bson:"time2"`
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://root:password@mongo:27017/?maxPoolSize=5&w=majority").SetServerAPIOptions(serverAPI)
	c := context.TODO()
	client, err := mongo.Connect(c, opts)
	if err != nil {
		t.Fatalf("Error connecting server: %v", err)
	}
	col := client.Database("timi_test").Collection("timi_test")
	defer func() {
		if err = col.Drop(c); err != nil {
			panic(err)
		}
		if err = client.Disconnect(c); err != nil {
			panic(err)
		}
	}()

	ti := Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	data := TimeTestStruct{
		ID:    primitive.NewObjectID(),
		Time1: ti,
		Time2: NilTime,
	}
	if _, err = col.InsertOne(c, &data); err != nil {
		t.Fatalf("Error inserting record: %v", err)
	}
	var unmVal TimeTestStruct
	// Test retrieve with "time2=nil"
	if err = col.FindOne(c, primitive.M{"_id": data.ID, "time2": nil}).Decode(&unmVal); err != nil {
		t.Fatalf("Error retrieving record: %v", err)
	}
	if !unmVal.Time1.Equal(data.Time1) {
		t.Fatalf("Original time (%v) did not match with the expected time %v", data.Time1, unmVal.Time1)
	}
	if unmVal.Time2.IsNull() != data.Time2.IsNull() {
		t.Fatalf("Time2 is null mismatch (%v : %v)", unmVal.Time2.IsNull(), data.Time2.IsNull())
	}
	// Test retrieve with "time2=NilTime"
	if err = col.FindOne(c, primitive.M{"_id": data.ID, "time2": NilTime}).Decode(&unmVal); err != nil {
		t.Fatalf("Error retrieving record: %v", err)
	}
	if !unmVal.Time1.Equal(data.Time1) {
		t.Fatalf("Original time (%v) did not match with the expected time %v", data.Time1, unmVal.Time1)
	}
	if unmVal.Time2.IsNull() != data.Time2.IsNull() {
		t.Fatalf("Time2 is null mismatch (%v : %v)", unmVal.Time2.IsNull(), data.Time2.IsNull())
	}
	// Test with update
	if _, err = col.UpdateOne(c, primitive.M{"_id": data.ID}, primitive.M{"$set": primitive.M{"time2": ti}}); err != nil {
		t.Fatalf("Error updating record: %v", err)
	}
	// Test retrieve with not nil "time2"
	if err = col.FindOne(c, primitive.M{"_id": data.ID, "time2": ti}).Decode(&unmVal); err != nil {
		t.Fatalf("Error retrieving record: %v", err)
	}
	if !unmVal.Time1.Equal(data.Time1) {
		t.Fatalf("Original time (%v) did not match with the expected time %v", data.Time1, unmVal.Time1)
	}
	if unmVal.Time2.IsNull() {
		t.Fatalf("Time2 is null")
	}
}

type TimeTestSqlStruct struct {
	ID    int64 `gorm:"column:id"`
	Time1 Time  `gorm:"column:time1"`
	Time2 Time  `gorm:"column:time2"`
}

func (TimeTestSqlStruct) TableName() string {
	return "timi_test"
}

func TestMySQL(t *testing.T) {
	dsn := "root:password@tcp(mariadb:3306)/?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("MySQL Open error: %v", err)
	}
	if err = db.Exec("CREATE DATABASE IF NOT EXISTS `timi_test` COLLATE 'utf8mb4_unicode_ci';").Error; err != nil {
		t.Fatalf("MySQL database creation error: %v", err)
	}

	dsn = "root:password@tcp(mariadb:3306)/timi_test?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("MySQL Open error: %v", err)
	}
	defer func() {
		if err = db.Exec("DROP TABLE IF EXISTS `timi_test`;").Error; err != nil {
			t.Fatalf("MySQL table drop error: %v", err)
		}
		if err = db.Exec("DROP DATABASE IF EXISTS `timi_test`;").Error; err != nil {
			t.Fatalf("MySQL database drop error: %v", err)
		}
	}()
	table := `
	CREATE TABLE IF NOT EXISTS timi_test (
  		id int(10) NOT NULL,
		time1 datetime NOT NULL,
		time2 datetime DEFAULT NULL,
		PRIMARY KEY (id)
	)ENGINE=InnoDB;
	`
	if err = db.Exec(table).Error; err != nil {
		t.Fatalf("MySQL table creation error: %v", err)
	}
	// With values for both time and time pointer
	ti := Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	actual := TimeTestSqlStruct{
		ID:    1,
		Time1: ti,
		Time2: NilTime,
	}
	if err = db.Create(&actual).Error; err != nil {
		t.Fatalf("MySQL record creation error: %v", err)
	}
	var expected TimeTestSqlStruct
	if err = db.First(&expected, "id = ? AND time1 = ? AND time2 IS NULL", actual.ID, ti).Error; err != nil {
		t.Fatalf("MySQL record retrieval error: %v", err)
	}
	if actual.ID != expected.ID || !actual.Time1.Equal(expected.Time1) || !actual.Time2.IsNull() {
		t.Fatalf("MySQL record mismatch: %v", expected)
	}
}

func TestPostgres(t *testing.T) {
	dsn := "host=postgres user=postgres password=password port=5432 sslmode=disable"
	dbOp, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Postgres Open error: %v", err)
	}
	if err = dbOp.Exec("CREATE DATABASE timi_test;").Error; err != nil {
		t.Fatalf("Postgres database creation error: %v", err)
	}

	dsn = "host=postgres user=postgres password=password dbname=timi_test port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Postgres Open error: %v", err)
	}
	defer func() {
		if err = db.Exec("DROP TABLE IF EXISTS timi_test;").Error; err != nil {
			t.Fatalf("Postgres table drop error: %v", err)
		}
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		if err = dbOp.Exec("DROP DATABASE IF EXISTS timi_test;").Error; err != nil {
			t.Fatalf("Postgres database drop error: %v", err)
		}
	}()
	table := `
	CREATE TABLE timi_test (
		id integer NOT NULL,
		time1 timestamptz NOT NULL,
		time2 timestamptz DEFAULT NULL,
		CONSTRAINT timi_test_id PRIMARY KEY (id)
	) WITH (oids = false);
	`
	if err = db.Exec(table).Error; err != nil {
		t.Fatalf("Postgres table creation error: %v", err)
	}
	// With values for both time and time pointer
	ti := Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	actual := TimeTestSqlStruct{
		ID:    1,
		Time1: ti,
		Time2: NilTime,
	}
	if err = db.Create(&actual).Error; err != nil {
		t.Fatalf("Postgres record creation error: %v", err)
	}
	var expected TimeTestSqlStruct
	if err = db.First(&expected, "id = ? AND time1 = ? AND time2 IS NULL", actual.ID, ti).Error; err != nil {
		t.Fatalf("Postgres record retrieval error: %v", err)
	}
	if actual.ID != expected.ID || !actual.Time1.Equal(expected.Time1) || !actual.Time2.IsNull() {
		t.Fatalf("Postgres record mismatch: %v", expected)
	}
}
