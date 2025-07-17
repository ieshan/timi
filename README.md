# Timi - Nullable Time for Go

A Go package that provides a nullable time type with support for JSON, SQL databases, and optional BSON marshaling for MongoDB.

## 🎯 **Zero Dependencies Approach**

**Timi uses Go workspaces to provide a clean, dependency-free core package while offering optional integrations and comprehensive testing.**

```go
import "github.com/ieshan/timi"
// ✅ Zero external dependencies added to your project!
```

## 📁 **Project Structure**

```
timi/
├── go.work                     # Go workspace definition (3 workspaces)
├── go.mod                      # Main package (zero dependencies)
├── timi.go                     # Core nullable time functionality
├── timi_unit_test.go          # Unit tests (no external deps)
├── integration-tests/          # Database integration testing workspace
│   ├── go.mod                  # Integration dependencies isolated here
│   ├── timi_mongo_test.go     # MongoDB integration tests
│   └── timi_gorm_test.go      # SQL database tests (MySQL, PostgreSQL, SQLite)
├── mongodb/                    # MongoDB utilities workspace
│   ├── go.mod                  # MongoDB driver dependencies
│   ├── bson_helpers.go        # BSON marshaling utilities
│   └── bson_helpers_test.go   # BSON helpers tests
└── docker-compose.yml         # Test databases and Go runtime setup
```

## 🚀 **Installation**

```bash
go get github.com/ieshan/timi
```

## 💡 **Core Features**

### **Zero Dependencies Core**
- ✅ JSON marshaling/unmarshaling
- ✅ SQL database support via `database/sql`  
- ✅ Time manipulation methods (After, Before, Add, etc.)
- ✅ UTC timezone enforcement
- ✅ Null value handling

### **Optional Integrations**
- 🔧 MongoDB BSON support (dedicated workspace)
- 🧪 Full database integration tests (MongoDB, MySQL, PostgreSQL, SQLite)

## 📖 **Usage Examples**

### **Basic Usage (Zero Dependencies)**

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/ieshan/timi"
)

func main() {
    // Create times
    now := timi.Now()
    christmas := timi.Date(2024, time.December, 25, 15, 30, 0, 0, time.UTC)
    nilTime := timi.NilTime
    
    fmt.Printf("Current time: %s\n", now.String())
    fmt.Printf("Christmas: %s\n", christmas.String()) 
    fmt.Printf("Nil time is null: %t\n", nilTime.IsNull())
    
    // JSON marshaling - works out of the box
    type Event struct {
        Name      string    `json:"name"`
        StartTime timi.Time `json:"start_time"`
        EndTime   timi.Time `json:"end_time,omitempty"`
    }
    
    event := Event{
        Name:      "Holiday Party",
        StartTime: christmas,
        EndTime:   nilTime, // Will be null in JSON
    }
    
    jsonData, _ := json.Marshal(event)
    fmt.Printf("JSON: %s\n", jsonData)
    // Output: {"name":"Holiday Party","start_time":"2024-12-25T15:30:00Z","end_time":null}
}
```

### **SQL Database Usage**

```go
import (
    "database/sql"
    "github.com/ieshan/timi"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, _ := sql.Open("sqlite3", ":memory:")
    
    // Create table
    db.Exec(`CREATE TABLE events (
        id INTEGER PRIMARY KEY,
        name TEXT,
        start_time DATETIME,
        end_time DATETIME
    )`)
    
    // Insert with timi.Time (handles NULL values automatically)
    _, err := db.Exec(
        "INSERT INTO events (name, start_time, end_time) VALUES (?, ?, ?)",
        "Meeting", timi.Now(), timi.NilTime,
    )
    
    // Query back
    var name string
    var startTime, endTime timi.Time
    err = db.QueryRow("SELECT name, start_time, end_time FROM events WHERE id = 1").
        Scan(&name, &startTime, &endTime)
    
    fmt.Printf("Event: %s, Start: %s, End null: %t\n", 
        name, startTime.String(), endTime.IsNull())
}
```

### **MongoDB BSON Support (mongodb/ workspace)**

The mongodb workspace provides ready-to-use BSON helpers:

```go
// In your project, import the MongoDB driver:
// go get go.mongodb.org/mongo-driver/v2

// Copy the pattern from mongodb/bson_helpers.go:
import (
    "github.com/ieshan/timi"
    "go.mongodb.org/mongo-driver/v2/bson"
)

// Use the ToTimiBSON and FromTimiBSON helpers
type Event struct {
    ID        bson.ObjectID `bson:"_id"`
    Name      string        `bson:"name"`
    StartTime bson.Raw      `bson:"start_time"`
    EndTime   bson.Raw      `bson:"end_time"`
}

// Convert timi.Time to BSON
startTimeBSON, _ := ToTimiBSON(timi.Now())
endTimeBSON, _ := ToTimiBSON(timi.NilTime)

event := Event{
    ID:        bson.NewObjectID(),
    Name:      "Meeting",
    StartTime: startTimeBSON,
    EndTime:   endTimeBSON,
}

// Convert BSON back to timi.Time
startTime := FromTimiBSON(event.StartTime)
endTime := FromTimiBSON(event.EndTime)
```

## 🧪 **Testing**

### **Docker-First Approach**

**All Go commands and tests are designed to run inside Docker containers for consistency.**

#### **Quick Testing (Recommended)**
```bash
# Test everything with one command - runs in Docker
docker-compose run --rm all-tests
```

#### **Individual Test Suites**
```bash
# Unit tests only (zero dependencies)
docker-compose run --rm unit-test

# Database integration tests only
docker-compose run --rm integration-test

# MongoDB workspace tests only
docker-compose run --rm mongodb-test
```

#### **Go Commands in Docker**
```bash
# Run any Go command inside Docker
docker-compose run --rm go-workspace go version
docker-compose run --rm go-workspace go mod tidy
docker-compose run --rm go-workspace go work sync

# Work in specific workspaces
docker-compose run --rm go-workspace bash -c "cd integration-tests && go test -v"
docker-compose run --rm go-workspace bash -c "cd mongodb && go test -v"

# Interactive shell in Docker
docker-compose run --rm go-workspace bash
```

### **Docker Services**

| Service | Purpose | Dependencies | Working Directory |
|---------|---------|--------------|-------------------|
| `unit-test` | Unit tests only | None | `/app` (main package) |
| `integration-test` | SQL + MongoDB integration tests | MongoDB, MySQL, PostgreSQL | `/app/integration-tests` |
| `mongodb-test` | MongoDB workspace tests | MongoDB | `/app/mongodb` |
| `all-tests` | All tests across all workspaces | MongoDB, MySQL, PostgreSQL | All directories |
| `go-workspace` | General Go commands | MongoDB, MySQL, PostgreSQL | `/app` (configurable) |

### **Example Docker Output**

When running `docker-compose run --rm all-tests`:

```bash
=== Running Unit Tests ===
=== RUN   TestFrom
--- PASS: TestFrom (0.00s)
=== RUN   TestTime_String
--- PASS: TestTime_String (0.00s)
=== RUN   TestTime_MarshalJSON
--- PASS: TestTime_MarshalJSON (0.00s)
=== RUN   TestTime_UnmarshalJSON
--- PASS: TestTime_UnmarshalJSON (0.00s)
PASS
ok      github.com/ieshan/timi  0.002s

=== Running Integration Tests ===
=== RUN   TestMySQL
--- PASS: TestMySQL (0.06s)
=== RUN   TestPostgres
--- PASS: TestPostgres (0.07s)
=== RUN   TestSQLite
--- PASS: TestSQLite (0.00s)
=== RUN   TestMongo
--- PASS: TestMongo (0.06s)
PASS
ok      timi-integration-tests  0.188s

=== Running MongoDB Workspace Tests ===
=== RUN   TestBSONHelpers
--- PASS: TestBSONHelpers (0.00s)
PASS
ok      github.com/ieshan/timi/mongodb  0.002s

=== All Tests Complete ===
```

### **Local Testing (Optional)**

If you prefer local development:

```bash
# Unit tests (fast, no setup required)
go test -v

# Integration tests (requires manual database setup)
cd integration-tests && go test -v

# MongoDB workspace tests
cd mongodb && go test -v
```

### **Test Coverage**

#### **Unit Tests** (`timi_unit_test.go`)
- JSON marshaling/unmarshaling
- Time creation and manipulation
- String representation
- Null value handling

#### **Integration Tests** 
- **SQL Databases** (`timi_gorm_test.go`):
  - MySQL with microsecond precision
  - PostgreSQL with timezone support
  - SQLite with nanosecond precision
  - CRUD operations, queries, edge cases

- **MongoDB** (`timi_mongo_test.go`):
  - BSON marshaling/unmarshaling
  - MongoDB queries and operations
  - Precision handling (milliseconds)
  - Edge cases and timezone handling

#### **MongoDB Workspace Tests** (`bson_helpers_test.go`)
- BSON helper function testing
- Round-trip conversion validation
- Null value handling in BSON context

## 🏗️ **Architecture: Go Workspaces**

This project uses Go workspaces to solve the dependency management problem:

### **The Problem**
```bash
# Traditional approach - forces dependencies on all users
go mod tidy  # ❌ Adds MongoDB, GORM, database drivers to main go.mod
```

### **The Solution**
```bash
# Workspace approach - clean separation
go mod tidy                    # ✅ Main package: zero dependencies
cd integration-tests && go mod tidy  # ✅ Heavy deps isolated here
cd mongodb && go mod tidy      # ✅ MongoDB deps isolated here
```

### **Workspace Configuration (`go.work`)**

```go
go 1.24

use (
    .                    # Main timi package
    ./integration-tests  # Integration tests module
    ./mongodb           # MongoDB utilities module
)
```

### **Benefits**

- ✅ **Zero forced dependencies**: Main package stays clean
- ✅ **Complete testing**: Full database integration coverage
- ✅ **Developer friendly**: Work on all modules simultaneously  
- ✅ **CI/CD ready**: Test modules independently
- ✅ **Optional features**: Use MongoDB helpers when needed

## 📊 **Dependency Breakdown**

| Module | Dependencies | Purpose |
|--------|-------------|---------|
| **Main Package** | `go.mod` → Zero external deps | Core time functionality |
| **Integration Tests** | `go.mod` → MongoDB, GORM, DB drivers | Comprehensive database testing |
| **MongoDB Workspace** | `go.mod` → MongoDB driver only | BSON utilities and tests |
| **Your Project** | Only what you choose | Clean imports |

## 🔧 **For Package Consumers**

### **Scenario 1: Basic Usage (Most Common)**
```go
import "github.com/ieshan/timi"
// ✅ No external dependencies added to your go.mod
// ✅ JSON, SQL, time operations work perfectly
```

### **Scenario 2: With MongoDB Support**
```bash
# You explicitly add MongoDB to YOUR project
go get go.mongodb.org/mongo-driver/v2

# Copy BSON helper pattern from mongodb/bson_helpers.go
# ✅ You control your dependencies
```

### **Scenario 3: Contributing/Testing**
```bash
# Clone repository
git clone https://github.com/ieshan/timi
cd timi

# Work with workspace (everything in Docker)
docker-compose run --rm go-workspace go work sync  # Sync modules
docker-compose run --rm all-tests                  # Test everything

# Or work locally
go work sync               # Sync modules
go test -v                 # Test main package
cd integration-tests && go test -v  # Test integrations
cd ../mongodb && go test -v         # Test MongoDB workspace
```

## 🚦 **Development Workflow**

### **Docker-First Development (Recommended)**
```bash
# Quick development cycle
docker-compose run --rm unit-test           # Fast unit tests

# Comprehensive testing
docker-compose run --rm all-tests           # Test everything

# Targeted testing
docker-compose run --rm integration-test    # Database integration only
docker-compose run --rm mongodb-test        # MongoDB utilities only

# Interactive development
docker-compose run --rm go-workspace bash   # Get shell in container
```

### **Local Development (Alternative)**
```bash
# Quick unit tests
go test -v                    # Fast unit tests (local)

# Full integration testing (requires local databases)
cd integration-tests && go test -v
cd ../mongodb && go test -v

# Workspace management
go work sync                  # Keep modules aligned
```

## 📋 **API Reference**

### **Core Types**

```go
type Time sql.NullTime
var NilTime = Time{Time: time.Time{}, Valid: false}
```

### **Creation Functions**

```go
func Now() Time
func Date(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) Time
```

### **Key Methods**

```go
func (t Time) String() string
func (t *Time) IsNull() bool
func (t Time) IsZero() bool
func (t Time) After(u Time) bool
func (t Time) Before(u Time) bool
func (t Time) Equal(u Time) bool
func (t Time) Add(d time.Duration) Time
func (t Time) Sub(u Time) time.Duration

// JSON support
func (t Time) MarshalJSON() ([]byte, error)
func (t *Time) UnmarshalJSON(data []byte) error

// SQL support  
func (t *Time) Scan(value interface{}) error
func (t Time) Value() (driver.Value, error)
```

### **MongoDB BSON Helpers** (`mongodb/` workspace)

```go
func ToTimiBSON(t timi.Time) (bson.Raw, error)
func FromTimiBSON(raw bson.Raw) timi.Time
```

## 🤝 **Contributing**

1. **Core changes**: Work in main directory, test with `docker-compose run --rm unit-test`
2. **Integration changes**: Work in `integration-tests/`, test with `docker-compose run --rm integration-test`
3. **MongoDB changes**: Work in `mongodb/`, test with `docker-compose run --rm mongodb-test`
4. **All tests**: Always run `docker-compose run --rm all-tests` before submitting

**All development should use Docker to ensure consistency across environments.**

## 📄 **License**

[Add your license here]

---

**🎉 Timi provides clean, nullable time handling with zero dependency pollution and Docker-first development - the best of both worlds!**
