module github.com/ieshan/timi/mongodb

go 1.25

require (
	github.com/ieshan/timi v0.0.0
	go.mongodb.org/mongo-driver/v2 v2.3.0
)

// Use local timi package
replace github.com/ieshan/timi => ../
