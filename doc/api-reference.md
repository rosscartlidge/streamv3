# StreamV3 API Reference

*Complete reference for all StreamV3 types, functions, and methods*

> 📖 **Documentation Note**: This is a learning-focused API reference with examples and best practices. For raw API documentation directly from source code, use `go doc github.com/rosscartlidge/streamv3` or browse specific functions with `go doc github.com/rosscartlidge/streamv3.FunctionName`

## Table of Contents

### Documentation Navigation
- [Getting Started Guide](codelab-intro.md) - Learn StreamV3 basics step-by-step
- [Advanced Tutorial](advanced-tutorial.md) - Complex patterns and real-world examples

### API Reference Sections
- [Installation & Setup](#installation--setup)
- [Core Types](#core-types)
- [Creating Iterators](#creating-iterators)
- [Transform Operations](#transform-operations)
- [Filter Operations](#filter-operations)
- [Limiting & Pagination](#limiting--pagination)
- [Ordering Operations](#ordering-operations)
- [Aggregation & Analysis](#aggregation--analysis)
- [Window Operations](#window-operations)
- [Early Termination](#early-termination)
- [SQL-Style Operations](#sql-style-operations)
  - [Join Operations](#join-operations)
  - [GroupBy Operations](#groupby-operations)
  - [Aggregation Functions](#aggregation-functions)
- [Composition Operations](#composition-operations)
- [Flattening Operations](#flattening-operations)
- [Utility Operations](#utility-operations)
- [I/O Operations](#io-operations)
  - [CSV Operations](#csv-operations)
  - [JSON Operations](#json-operations)
  - [Line Operations](#line-operations)
  - [Command Output Operations](#command-output-operations)
- [Chart & Visualization](#chart--visualization)
- [Helper Functions](#helper-functions)
  - [Record Access](#record-access)
- [Error Handling](#error-handling)

---

## Installation & Setup

### Requirements
- **Go 1.23+** (required for iterator support)

### Step 1: Install Go

If you don't have Go installed:

**macOS:**
```bash
brew install go
```

**Linux:**
```bash
# Download and install Go 1.23+
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**Windows:**
Download the installer from [https://go.dev/dl/](https://go.dev/dl/)

**Verify installation:**
```bash
go version  # Should show 1.23 or higher
```

### Step 2: Create a New Project

```bash
# Create project directory
mkdir my-streamv3-project
cd my-streamv3-project

# Initialize Go module (required for dependency management)
go mod init myproject

# Or use your GitHub path for a real project:
# go mod init github.com/yourusername/myproject
```

### Step 3: Install StreamV3

```bash
go get github.com/rosscartlidge/streamv3
```

This will:
- Download StreamV3 and its dependencies
- Update your `go.mod` file with the dependency
- Create/update `go.sum` with checksums

### Step 4: Import and Use

Create a file `main.go`:
```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    data, err := streamv3.ReadCSV("data.csv")
    if err != nil {
        log.Fatalf("Failed to read CSV: %v", err)
    }

    for record := range data {
        name := streamv3.GetOr(record, "name", "")
        fmt.Println(name)
    }
}
```

### Step 5: Run Your Program

```bash
go run main.go
```

Or build an executable:
```bash
go build
./my-streamv3-project  # or my-streamv3-project.exe on Windows
```

---

## Core Types

### Iterator Types
StreamV3 uses Go 1.23+ iterators as its core abstraction:

```go
iter.Seq[T]           // Simple iterator
iter.Seq2[T, error]   // Iterator with error handling
```

### Record
A flexible data structure for heterogeneous data.

```go
type Record struct {
    // fields are private - use the provided API to access
}
```

**🚨 CRITICAL (v1.0+): Record is an encapsulated struct, NOT a map**

Record fields are **NOT directly accessible**. You MUST use the provided API:

**❌ WRONG - Direct field access (will not compile):**
```go
record["name"] = "Alice"        // ❌ Compile error!
value := record["age"]          // ❌ Compile error!
for k, v := range record {      // ❌ Compile error!
```

**✅ CORRECT - Use the builder pattern and accessor functions:**
```go
// Creating - Use MakeMutableRecord builder
record := streamv3.MakeMutableRecord().
    String("name", "Alice").
    Int("age", int64(30)).
    Float("score", 95.5).
    Freeze()

// Reading - Use Get/GetOr
name := streamv3.GetOr(record, "name", "")
age, exists := streamv3.Get[int64](record, "age")

// Modifying - Use SetImmutable (creates new record)
updated := streamv3.SetImmutable(record, "score", 98.0)

// Iterating - Use .All() method
for key, value := range record.All() {
    fmt.Printf("%s: %v\n", key, value)
}
```

**Supported value types:** `int64`, `float64`, `string`, `bool`, `time.Time`, nested `Record`, `iter.Seq[T]`, and slices.

### MutableRecord
A mutable record type optimized for efficient building.

```go
type MutableRecord struct {
    // fields are private - use the provided methods
}
```

MutableRecord is the recommended way to build new records efficiently. Unlike Record methods which create copies, MutableRecord methods modify the same underlying map, avoiding unnecessary allocations.

**Building with MutableRecord:**
```go
// Efficient building with mutation
record := streamv3.MakeMutableRecord().
    String("name", "Alice").
    Int("age", int64(30)).
    Float("salary", 95000.50).
    Bool("active", true).
    Freeze()  // Convert to immutable Record

// For use in slices, always call .Freeze()
records := []streamv3.Record{
    streamv3.MakeMutableRecord().
        String("id", "001").
        Int("count", 42).
        Freeze(),
}
```

**Available Methods:**
- `.String(field, value)` - Set string field
- `.Int(field, value)` - Set int64 field
- `.Float(field, value)` - Set float64 field
- `.Bool(field, value)` - Set boolean field
- `.Time(field, value)` - Set time.Time field
- `.Nested(field, value)` - Set nested Record field
- `.JSONString(field, value)` - Set JSONString field
- `.IntSeq(field, value)` - Set int sequence field (also Int8Seq, Int16Seq, Int32Seq, Int64Seq, UintSeq, etc.)
- `.FloatSeq(field, value)` - Set float sequence (Float32Seq, Float64Seq)
- `.StringSeq(field, value)` - Set string sequence
- `.RecordSeq(field, value)` - Set Record sequence
- `.SetAny(field, value)` - Set any value type
- `.Delete(field)` - Remove a field
- `.Freeze()` - Convert to immutable Record
- `.Len()` - Get number of fields

### JSONString
A string type containing valid JSON data.

```go
type JSONString string
```

JSONString provides type safety and rich methods for working with JSON-structured data that needs to be embedded in Records.

**Example:**
```go
// Create JSONString from Go value
jsonStr, err := streamv3.NewJSONString(map[string]any{
    "status": "active",
    "count": 42,
})
if err != nil {
    log.Fatal(err)
}

// Use in Record
record := streamv3.MakeMutableRecord().
    String("id", "user123").
    JSONString("metadata", jsonStr).
    Freeze()

// Parse back to Go value
metadata := streamv3.GetOr(record, "metadata", streamv3.JSONString(""))
value, err := metadata.Parse()

// Pretty print
fmt.Println(metadata.Pretty())
```

**Methods:**
- `NewJSONString(value any) (JSONString, error)` - Create from Go value
- `.Parse() (any, error)` - Parse to Go value
- `.MustParse() any` - Parse or panic
- `.IsValid() bool` - Check if valid JSON
- `.Pretty() string` - Pretty-printed JSON
- `.String() string` - Raw JSON string

### Value Interface
Type constraint for Record values.

```go
type Value interface {
    ~int64 | ~float64 |
    ~bool | string | time.Time |
    JSONString | Record |
    iter.Seq[int] | iter.Seq[int8] | iter.Seq[int16] | iter.Seq[int32] | iter.Seq[int64] |
    iter.Seq[uint] | iter.Seq[uint8] | iter.Seq[uint16] | iter.Seq[uint32] | iter.Seq[uint64] |
    iter.Seq[float32] | iter.Seq[float64] |
    iter.Seq[bool] | iter.Seq[string] | iter.Seq[time.Time] |
    iter.Seq[Record]
}
```

This interface defines all valid types that can be stored in a Record. Uses a **hybrid approach**:
- **Canonical scalars**: `int64` and `float64` only (not `int`, `int32`, `float32`, etc.)
- **Flexible sequences**: Any numeric iterator type allowed (e.g., `iter.Seq[int]`, `iter.Seq[int32]`, etc.)

This design eliminates type ambiguity for scalar values while maintaining compatibility with Go's standard library for sequences.

### Filter Types
Function types for stream transformations:

```go
type Filter[T, U any] func(iter.Seq[T]) iter.Seq[U]
type FilterWithErrors[T, U any] func(iter.Seq2[T, error]) iter.Seq2[U, error]
```

---

## Creating Iterators

### From[T]
```go
func From[T any](slice []T) iter.Seq[T]
```
Creates an iterator from a slice - convenience wrapper providing a more discoverable API.

**Example:**
```go
numbers := streamv3.From([]int{1, 2, 3, 4, 5})
records := streamv3.From([]streamv3.Record{...})
```

**Note:** This is equivalent to `slices.Values()` from the standard library, but provides a more intuitive name for users familiar with other streaming libraries.

### From Slices (Standard Library)
```go
slices.Values([]T) iter.Seq[T]
```
Creates an iterator from a slice (standard library function). You can use either this or `streamv3.From()`.

**Example:**
```go
numbers := slices.Values([]int{1, 2, 3, 4, 5})
```

### ToChannel[T]
```go
func ToChannel[T any](sb iter.Seq[T]) <-chan T
```
Converts an iterator to a channel.

### FromChannelSafe[T]
```go
func FromChannelSafe[T any](itemCh <-chan T, errCh <-chan error) iter.Seq2[T, error]
```
Creates an iterator from separate item and error channels.

### ToChannelWithErrors[T]
```go
func ToChannelWithErrors[T any](sb iter.Seq2[T, error]) (<-chan T, <-chan error)
```
Converts an error-aware iterator to separate item and error channels.

**Example:**
```go
data, err := streamv3.ReadCSVSafe("data.csv")
if err != nil {
    log.Fatal(err)
}
itemCh, errCh := streamv3.ToChannelWithErrors(data)

go func() {
    for err := range errCh {
        log.Printf("Error: %v", err)
    }
}()

for record := range itemCh {
    // Process record
}
```

### MakeMutableRecord
```go
func MakeMutableRecord() MutableRecord
```
Creates a new mutable record for efficient building. Use `.Freeze()` to convert to a regular `Record` when done building.

**Example:**
```go
// Efficient building with mutation
record := streamv3.MakeMutableRecord().
    String("name", "Alice").
    Int("age", 30).
    Float("score", 95.5).
    Freeze()  // Convert to frozen Record

// For use in slices, always call .Freeze()
records := []streamv3.Record{
    streamv3.MakeMutableRecord().
        String("id", "001").
        Int("count", 42).
        Freeze(),
}
```

**Note:** `MutableRecord` methods mutate in place for efficiency during construction. Call `.Freeze()` to get a regular `Record` for use in pipelines or data structures.

---

## Transform Operations

*Functions that transform elements from one type to another*

> 💡 **Learn by Example**: See these operations in action in the [Getting Started Guide](codelab-intro.md#basic-transformations) and [Advanced Tutorial](advanced-tutorial.md#complex-aggregations).

### Select[T, U]
```go
func Select[T, U any](fn func(T) U) Filter[T, U]
```
Transforms each element using the provided function (SQL SELECT equivalent).

**Example:**
```go
doubled := streamv3.Select(func(x int) int { return x * 2 })(numbers)
```

### Update
```go
func Update(fn func(MutableRecord) MutableRecord) Filter[Record, Record]
```
Convenience wrapper around Select for updating record fields. Automatically handles `ToMutable()` and `Freeze()` boilerplate, making field updates more concise.

**Example - Update single field:**
```go
updated := streamv3.Update(func(mut streamv3.MutableRecord) streamv3.MutableRecord {
    return mut.String("status", "processed")
})(records)
```

**Example - Update multiple fields:**
```go
updated := streamv3.Update(func(mut streamv3.MutableRecord) streamv3.MutableRecord {
    return mut.
        String("status", "active").
        Time("updated_at", time.Now())
})(records)
```

**Example - Computed field:**
```go
updated := streamv3.Update(func(mut streamv3.MutableRecord) streamv3.MutableRecord {
    frozen := mut.Freeze()
    price := streamv3.GetOr(frozen, "price", float64(0))
    qty := streamv3.GetOr(frozen, "quantity", int64(0))
    return mut.Float("total", price * float64(qty))
})(records)
```

**Equivalent without Update:**
```go
// More verbose - need explicit ToMutable() and Freeze()
updated := streamv3.Select(func(r streamv3.Record) streamv3.Record {
    return r.ToMutable().String("status", "processed").Freeze()
})(records)
```

### SelectSafe[T, U]
```go
func SelectSafe[T, U any](fn func(T) (U, error)) FilterWithErrors[T, U]
```
Safe version of Select that handles errors.

### SelectMany[T, U]
```go
func SelectMany[T, U any](fn func(T) iter.Seq[U]) Filter[T, U]
```
Flattens nested sequences (FlatMap equivalent).

**Example:**
```go
words := streamv3.SelectMany(func(line string) iter.Seq[string] {
    return slices.Values(strings.Fields(line))
})(lines)
```

---

## Filter Operations

*Functions that filter elements based on conditions*

### Where[T]
```go
func Where[T any](predicate func(T) bool) Filter[T, T]
```
Filters elements based on a predicate (SQL WHERE equivalent).

**Example:**
```go
evens := streamv3.Where(func(x int) bool { return x%2 == 0 })(numbers)
```

### WhereSafe[T]
```go
func WhereSafe[T any](predicate func(T) (bool, error)) FilterWithErrors[T, T]
```
Safe version of Where that handles errors.

### Distinct[T]
```go
func Distinct[T comparable]() Filter[T, T]
```
Removes duplicate elements.

### DistinctBy[T, K]
```go
func DistinctBy[T any, K comparable](keyFn func(T) K) Filter[T, T]
```
Removes duplicates based on a key function.

---

## Limiting & Pagination

*Functions for limiting and paginating streams*

### Limit[T]
```go
func Limit[T any](n int) Filter[T, T]
```
Takes only the first n elements (SQL LIMIT equivalent).

**Example:**
```go
first5 := streamv3.Limit[int](5)(numbers)
```

### LimitSafe[T]
```go
func LimitSafe[T any](n int) FilterWithErrors[T, T]
```
Safe version of Limit that handles errors.

### Offset[T]
```go
func Offset[T any](n int) Filter[T, T]
```
Skips the first n elements (SQL OFFSET equivalent).

### OffsetSafe[T]
```go
func OffsetSafe[T any](n int) FilterWithErrors[T, T]
```
Safe version of Offset that handles errors.

---

## Ordering Operations

*Functions for sorting and ordering streams*

### Sort[T]
```go
func Sort[T cmp.Ordered]() Filter[T, T]
```
Sorts elements in ascending order.

### SortBy[T, K]
```go
func SortBy[T any, K cmp.Ordered](keyFn func(T) K) Filter[T, T]
```
Sorts elements by a key function.

### SortDesc[T]
```go
func SortDesc[T cmp.Ordered]() Filter[T, T]
```
Sorts elements in descending order.

### Reverse[T]
```go
func Reverse[T any]() Filter[T, T]
```
Reverses the order of elements.

---

## Aggregation & Analysis

*Functions for running aggregations and statistical analysis*

### RunningSum
```go
func RunningSum(fieldName string) Filter[Record, Record]
```
Calculates running sum for a numeric field.

### RunningAverage
```go
func RunningAverage(fieldName string, windowSize int) Filter[Record, Record]
```
Calculates running average over a sliding window.

### ExponentialMovingAverage
```go
func ExponentialMovingAverage(fieldName string, alpha float64) Filter[Record, Record]
```
Calculates exponential moving average.

### RunningMinMax
```go
func RunningMinMax(fieldName string) Filter[Record, Record]
```
Tracks running minimum and maximum values.

### RunningCount
```go
func RunningCount(fieldName string) Filter[Record, Record]
```
Maintains running count statistics.

---

## Window Operations

*Functions for windowing and batching streams*

> 🔄 **Infinite Stream Patterns**: Learn advanced windowing for real-time processing in the [Advanced Tutorial](advanced-tutorial.md#windowing-for-infinite-streams).

### CountWindow[T]
```go
func CountWindow[T any](size int) Filter[T, []T]
```
Groups elements into fixed-size windows.

**Example:**
```go
batches := streamv3.CountWindow[int](3)(numbers) // [1,2,3], [4,5,6], ...
```

### SlidingCountWindow[T]
```go
func SlidingCountWindow[T any](windowSize, stepSize int) Filter[T, []T]
```
Creates sliding windows with configurable step size.

### TimeWindow[T]
```go
func TimeWindow[T any](duration time.Duration, timeField string) Filter[T, []T]
```
Groups elements by time intervals.

### SlidingTimeWindow[T]
```go
func SlidingTimeWindow[T any](windowDuration, slideDuration time.Duration, timeField string) Filter[T, []T]
```
Creates sliding time-based windows.

---

## Early Termination

*Functions for controlled stream termination*

### TakeWhile[T]
```go
func TakeWhile[T any](predicate func(T) bool) Filter[T, T]
```
Takes elements while condition is true.

### TakeUntil[T]
```go
func TakeUntil[T any](predicate func(T) bool) Filter[T, T]
```
Takes elements until condition becomes true.

### SkipWhile[T]
```go
func SkipWhile[T any](predicate func(T) bool) Filter[T, T]
```
Skips elements while condition is true.

### SkipUntil[T]
```go
func SkipUntil[T any](predicate func(T) bool) Filter[T, T]
```
Skips elements until condition becomes true.

### Timeout[T]
```go
func Timeout[T any](duration time.Duration) Filter[T, T]
```
Terminates stream after specified duration.

### TimeBasedTimeout
```go
func TimeBasedTimeout(timeField string, duration time.Duration) Filter[Record, Record]
```
Terminates based on time field values in records.

---

## SQL-Style Operations

*Database-like operations for Record streams*

> 🎯 **Real-World Examples**: See comprehensive join and aggregation patterns in the [Advanced Tutorial](advanced-tutorial.md#stream-joins) section.

### Join Operations

#### JoinPredicate Type

```go
type JoinPredicate func(left, right Record) bool
```

JoinPredicate defines the condition for joining two records. Returns true if the left and right records should be joined together.

**Example:**
```go
// Using OnFields helper
predicate := streamv3.OnFields("user_id")

// Custom predicate
customPredicate := streamv3.OnCondition(func(left, right streamv3.Record) bool {
    leftID := streamv3.GetOr(left, "user_id", "")
    rightID := streamv3.GetOr(right, "customer_id", "")
    return leftID == rightID
})
```

#### InnerJoin
```go
func InnerJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) Filter[Record, Record]
```
Performs inner join between two record streams.

#### LeftJoin
```go
func LeftJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) Filter[Record, Record]
```
Performs left outer join.

#### RightJoin
```go
func RightJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) Filter[Record, Record]
```
Performs right outer join.

#### FullJoin
```go
func FullJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) Filter[Record, Record]
```
Performs full outer join.

#### Join Predicates

##### OnFields
```go
func OnFields(fields ...string) JoinPredicate
```
Creates join predicate based on field equality.

##### OnCondition
```go
func OnCondition(condition func(left, right Record) bool) JoinPredicate
```
Creates custom join predicate.

**Example:**
```go
joined := streamv3.InnerJoin(
    rightStream,
    streamv3.OnFields("user_id")
)(leftStream)
```

### GroupBy Operations

#### GroupBy[K]
```go
func GroupBy[K comparable](sequenceField string, keyField string, keyFn func(Record) K) Filter[Record, Record]
```
Groups records by a key function.

#### GroupByFields
```go
func GroupByFields(sequenceField string, fields ...string) Filter[Record, Record]
```
Groups records by field values.

**Example:**
```go
grouped := streamv3.GroupByFields("sales_data", "region", "product")(records)
```

### Aggregation Functions

#### AggregateFunc Type

```go
type AggregateFunc func([]Record) any
```

AggregateFunc defines an aggregation function over a group of records. Takes a slice of records and returns an aggregated value.

**Built-in Aggregation Functions:**

#### Count
```go
func Count() AggregateFunc
```
Counts records in each group.

#### Sum
```go
func Sum(field string) AggregateFunc
```
Sums numeric field values.

#### Avg
```go
func Avg(field string) AggregateFunc
```
Calculates average of numeric field.

#### Min[T] / Max[T]
```go
func Min[T cmp.Ordered](field string) AggregateFunc
func Max[T cmp.Ordered](field string) AggregateFunc
```
Finds minimum/maximum field values.

#### First / Last
```go
func First(field string) AggregateFunc
func Last(field string) AggregateFunc
```
Gets first/last field value in group.

#### Collect
```go
func Collect(field string) AggregateFunc
```
Collects all field values into an array.

#### Aggregate
```go
func Aggregate(sequenceField string, aggregations map[string]AggregateFunc) Filter[Record, Record]
```
Applies multiple aggregations to grouped data.

**Example:**
```go
results := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
    "total_sales": streamv3.Sum("amount"),
    "avg_sale":    streamv3.Avg("amount"),
    "count":       streamv3.Count(),
})(groupedRecords)
```

---

## Composition Operations

### Pipe[T, U, V]
```go
func Pipe[T, U, V any](f1 Filter[T, U], f2 Filter[U, V]) Filter[T, V]
```
Composes two filters into a single filter.

**Example:**
```go
// Compose "double" and "add 1" into single filter
double := streamv3.Select(func(x int) int { return x * 2 })
addOne := streamv3.Select(func(x int) int { return x + 1 })
composed := streamv3.Pipe(double, addOne)

result := composed(numbers) // Doubles then adds 1
```

### Pipe3[T, U, V, W]
```go
func Pipe3[T, U, V, W any](f1 Filter[T, U], f2 Filter[U, V], f3 Filter[V, W]) Filter[T, W]
```
Composes three filters into a single filter.

### PipeWithErrors[T, U, V]
```go
func PipeWithErrors[T, U, V any](f1 FilterWithErrors[T, U], f2 FilterWithErrors[U, V]) FilterWithErrors[T, V]
```
Composes two error-handling filters.

### ChainWithErrors[T]
```go
func ChainWithErrors[T any](filters ...FilterWithErrors[T, T]) FilterWithErrors[T, T]
```
Chains multiple same-type error-handling filters together.

---

## Flattening Operations

### DotFlatten
```go
func DotFlatten(separator string, fields ...string) Filter[Record, Record]
```
Flattens multiple sequence fields using dot product (parallel iteration). If you have sequences of equal length and want to pair up elements at matching positions, use this.

**Example:**
```go
// Input: {names: ["Alice", "Bob"], ages: [30, 25]}
// Output: [{name: "Alice", age: 30}, {name: "Bob", age: 25}]
flattened := streamv3.DotFlatten(",", "names", "ages")(records)
```

### CrossFlatten
```go
func CrossFlatten(separator string, fields ...string) Filter[Record, Record]
```
Flattens multiple sequence fields using Cartesian product. Each element from the first sequence is paired with every element from the second sequence.

**Example:**
```go
// Input: {colors: ["red", "blue"], sizes: ["S", "M"]}
// Output: [{color: "red", size: "S"}, {color: "red", size: "M"},
//          {color: "blue", size: "S"}, {color: "blue", size: "M"}]
flattened := streamv3.CrossFlatten(",", "colors", "sizes")(records)
```

---

## Utility Operations

### Hash
```go
func Hash(sourceField, targetField string) Filter[Record, Record]
```
Creates a SHA256 hash of a string field for efficient grouping. Useful for grouping on long strings or when you need fixed-length grouping keys. The hash is hex-encoded (64 characters) for readability and compatibility.

**Example:**
```go
// Hash long URLs for efficient grouping
hashed := streamv3.Hash("url", "url_hash")(records)

// Now group by the hash instead of the full URL
grouped := streamv3.GroupByFields("data", "url_hash")(hashed)
```

**Use cases:**
- Grouping on very long strings (URLs, file paths, etc.)
- Creating fixed-length keys for external systems
- Deduplication based on content

### Materialize
```go
func Materialize(field string) Filter[Record, Record]
```
Converts `iter.Seq` fields to `[]string` for better readability when inspecting data.

### MaterializeJSON
```go
func MaterializeJSON(field string) Filter[Record, Record]
```
Converts `iter.Seq` and `Record` fields to JSON-compatible types while preserving type information.

### Tee[T]
```go
func Tee[T any](input iter.Seq[T], n int) []iter.Seq[T]
```
Splits stream into multiple independent streams.

### LazyTee[T]
```go
func LazyTee[T any](input iter.Seq[T], n int) []iter.Seq[T]
```
Lazy version of Tee for memory efficiency.

### Chain[T]
```go
func Chain[T any](filters ...Filter[T, T]) Filter[T, T]
```
Chains multiple same-type filters together.

**Example:**
```go
pipeline := streamv3.Chain(
    streamv3.Where(func(x int) bool { return x > 0 }),
    streamv3.Where(func(x int) bool { return x < 100 }),
    streamv3.Sort[int](),
)
result := pipeline(numbers)
```

---

## I/O Operations

> 📁 **Practical Examples**: See file processing patterns in the [Getting Started Guide](codelab-intro.md#working-with-data) and production I/O strategies in the [Advanced Tutorial](advanced-tutorial.md#performance-optimization).

### CSV Operations

**⚠️ Important: CSV Auto-Parsing Behavior**

CSV operations automatically parse string values into appropriate Go types:
- Numeric strings → `int64` or `float64` (e.g., `"25"` becomes `int64(25)`)
- Boolean strings → `bool` (e.g., `"true"` becomes `true`)
- Other values → `string`

This means when reading CSV data, you must use the correct type when accessing fields:

```go
// CSV file: name,age,score
//           Alice,30,95.5

data := streamv3.ReadCSV("data.csv")
for record := range data {
    // ❌ WRONG - age is int64, not string
    age := streamv3.GetOr(record, "age", "")

    // ✅ CORRECT - use int64 for numeric CSV values
    age := streamv3.GetOr(record, "age", int64(0))

    // ✅ CORRECT - use float64 for decimal CSV values
    score := streamv3.GetOr(record, "score", 0.0)

    // ✅ CORRECT - strings remain strings
    name := streamv3.GetOr(record, "name", "")
}
```

When filtering CSV data, use the parsed types:
```go
// Filter for ages greater than 25
filtered := streamv3.Where(func(r streamv3.Record) bool {
    age := streamv3.GetOr(r, "age", int64(0))
    return age > 25  // Compare as int64
})(data)
```

#### ReadCSV
```go
func ReadCSV(filename string, config ...CSVConfig) (iter.Seq[Record], error)
```
Reads CSV file into Record iterator. Returns error if file cannot be opened. **Values are auto-parsed** to appropriate types.

**Example:**
```go
data, err := streamv3.ReadCSV("data.csv")
if err != nil {
    log.Fatalf("Failed to read CSV: %v", err)
}
for record := range data {
    // Process record
}
```

#### ReadCSVFromReader
```go
func ReadCSVFromReader(reader io.Reader, config ...CSVConfig) iter.Seq[Record]
```
Reads CSV from any io.Reader. **Values are auto-parsed** to appropriate types.

#### ReadCSVSafe
```go
func ReadCSVSafe(filename string, config ...CSVConfig) iter.Seq2[Record, error]
```
Error-aware version of ReadCSV. Returns iterator that yields both records and errors encountered during reading.

**Example:**
```go
data, err := streamv3.ReadCSVSafe("data.csv")
if err != nil {
    log.Fatalf("Failed to open CSV: %v", err)
}
for record, err := range data {
    if err != nil {
        log.Printf("Error reading record: %v", err)
        continue
    }
    // Process record
}
```

#### ReadCSVSafeFromReader
```go
func ReadCSVSafeFromReader(reader io.Reader, config ...CSVConfig) iter.Seq2[Record, error]
```
Error-aware version of ReadCSVFromReader.

#### WriteCSV
```go
func WriteCSV(stream iter.Seq[Record], filename string, config ...CSVConfig) error
```
Writes Record iterator to CSV file. Fields are auto-detected (all non-underscore, non-complex fields in alphabetical order) unless explicitly specified via config.Fields.

#### WriteCSVToWriter
```go
func WriteCSVToWriter(stream iter.Seq[Record], writer io.Writer, config ...CSVConfig) error
```
Writes Record iterator to any io.Writer as CSV.

**Example:**
```go
var buf bytes.Buffer
err := streamv3.WriteCSVToWriter(records, &buf)
if err != nil {
    log.Fatalf("Failed to write CSV: %v", err)
}
csvString := buf.String()
```

#### DefaultCSVConfig
```go
func DefaultCSVConfig() CSVConfig
```
Returns default CSV configuration.

#### CSVConfig
```go
type CSVConfig struct {
    Delimiter      rune
    Comment        rune
    FieldsPerRecord int
    LazyQuotes     bool
    TrimLeadingSpace bool
    Fields         []string  // Explicit field order for writing
}
```

### JSON Operations

#### ReadJSON
```go
func ReadJSON(filename string) (iter.Seq[Record], error)
```
Reads JSONL (JSON Lines) file into Record iterator. Returns error if file cannot be opened.

**Example:**
```go
data, err := streamv3.ReadJSON("data.jsonl")
if err != nil {
    log.Fatalf("Failed to read JSON: %v", err)
}
for record := range data {
    // Process record
}
```

#### ReadJSONFromReader
```go
func ReadJSONFromReader(reader io.Reader) iter.Seq[Record]
```
Reads JSONL from any io.Reader (stdin, network, etc).

#### ReadJSONSafe
```go
func ReadJSONSafe(filename string) iter.Seq2[Record, error]
```
Error-aware version of ReadJSON. Returns iterator that yields both records and parse errors.

**Example:**
```go
data, err := streamv3.ReadJSONSafe("data.jsonl")
if err != nil {
    log.Fatalf("Failed to open JSON file: %v", err)
}
for record, err := range data {
    if err != nil {
        log.Printf("Error parsing JSON: %v", err)
        continue
    }
    // Process record
}
```

#### ReadJSONSafeFromReader
```go
func ReadJSONSafeFromReader(reader io.Reader) iter.Seq2[Record, error]
```
Error-aware version of ReadJSONFromReader.

#### WriteJSON
```go
func WriteJSON(stream iter.Seq[Record], filename string) error
```
Writes Record iterator to JSON file.

#### WriteJSONToWriter
```go
func WriteJSONToWriter(stream iter.Seq[Record], writer io.Writer) error
```
Writes Record iterator to any io.Writer as JSONL.

### Line Operations

#### ReadLines
```go
func ReadLines(filename string) (iter.Seq[Record], error)
```
Reads text file line by line into Records with a "line" field. Returns error if file cannot be opened.

**Example:**
```go
lines, err := streamv3.ReadLines("logfile.txt")
if err != nil {
    log.Fatalf("Failed to read file: %v", err)
}
for record := range lines {
    line := streamv3.GetOr(record, "line", "")
    fmt.Println(line)
}
```

#### ReadLinesSafe
```go
func ReadLinesSafe(filename string) iter.Seq2[Record, error]
```
Error-aware version of ReadLines. Returns iterator that yields both records and read errors.

#### WriteLines
```go
func WriteLines(stream iter.Seq[Record], filename string) error
```
Writes Records to text file, one line per record (uses "line" field).

### Command Output Operations

#### ExecCommand
```go
func ExecCommand(command string, args []string, config ...CommandConfig) (iter.Seq[Record], error)
```
Executes a command and parses its column-aligned output into Records. Returns error if command fails to start.

**Example:**
```go
processes, err := streamv3.ExecCommand("ps", []string{"-efl"})
if err != nil {
    log.Fatalf("Failed to execute command: %v", err)
}
for process := range processes {
    cmd := streamv3.GetOr(process, "CMD", "")
    fmt.Println(cmd)
}
```

#### ReadCommandOutput
```go
func ReadCommandOutput(filename string, config ...CommandConfig) (iter.Seq[Record], error)
```
Reads previously captured command output from a file and parses column-aligned data. Returns error if file cannot be opened.

#### ReadCommandOutputSafe
```go
func ReadCommandOutputSafe(filename string, config ...CommandConfig) iter.Seq2[Record, error]
```
Error-aware version of ReadCommandOutput. Returns iterator that yields both records and parse errors.

#### ExecCommandSafe
```go
func ExecCommandSafe(command string, args []string, config ...CommandConfig) iter.Seq2[Record, error]
```
Error-aware version of ExecCommand. Returns iterator that yields both records and execution errors.

**Example:**
```go
processes, err := streamv3.ExecCommandSafe("ps", []string{"-efl"})
if err != nil {
    log.Fatalf("Failed to start command: %v", err)
}
for process, err := range processes {
    if err != nil {
        log.Printf("Error parsing output: %v", err)
        continue
    }
    cmd := streamv3.GetOr(process, "CMD", "")
    fmt.Println(cmd)
}
```

#### DefaultCommandConfig
```go
func DefaultCommandConfig() CommandConfig
```
Returns default command parsing configuration.

#### CommandConfig
```go
type CommandConfig struct {
    SkipLines      int
    TrimSpaces     bool
    MinColumnWidth int
}
```

---

## Chart & Visualization

> 📊 **Interactive Examples**: See chart creation in action in the [Getting Started Guide](codelab-intro.md#visualizing-data) and advanced dashboard patterns in the [Advanced Tutorial](advanced-tutorial.md#advanced-visualizations).

### Chart Configuration

#### DefaultChartConfig
```go
func DefaultChartConfig() ChartConfig
```
Creates default chart configuration.

#### ChartConfig
```go
type ChartConfig struct {
    Title              string            // Chart title
    Width              int               // Chart width in pixels
    Height             int               // Chart height in pixels
    ChartType          string            // "line", "bar", "scatter", "pie", "doughnut", "radar", "polarArea"
    TimeFormat         string            // Time format for time-based X axis
    XAxisType          string            // "linear", "logarithmic", "time", "category"
    YAxisType          string            // "linear", "logarithmic"
    ShowLegend         bool              // Show chart legend
    ShowTooltips       bool              // Enable hover tooltips
    EnableZoom         bool              // Enable zoom functionality
    EnablePan          bool              // Enable pan functionality
    EnableAnimations   bool              // Enable chart animations
    ShowDataLabels     bool              // Show data value labels
    EnableInteractive  bool              // Enable field selection UI
    EnableCalculations bool              // Enable running averages, etc.
    ColorScheme        string            // "default", "vibrant", "pastel", "monochrome"
    Theme              string            // "light", "dark"
    ExportFormats      []string          // Export options: "png", "svg", "pdf", "csv"
    CustomCSS          string            // Custom CSS for chart styling
    Fields             map[string]string // Field name -> data type hints
}
```

ChartConfig provides comprehensive control over chart appearance, behavior, and export options. Use `DefaultChartConfig()` to get sensible defaults, then customize as needed.

### Chart Creation

#### InteractiveChart
```go
func InteractiveChart(data iter.Seq[Record], filename string, config ...ChartConfig) error
```
Creates interactive HTML chart.

#### TimeSeriesChart
```go
func TimeSeriesChart(data iter.Seq[Record], timeField string, valueFields []string, filename string, config ...ChartConfig) error
```
Creates time series chart.

#### QuickChart
```go
func QuickChart(data iter.Seq[Record], xField, yField, filename string) error
```
Creates chart with default settings using specified X and Y fields.

**Example:**
```go
config := streamv3.DefaultChartConfig()
config.Title = "Sales Analysis"
config.ChartType = "bar"

err := streamv3.InteractiveChart(
    salesData,
    "sales_chart.html",
    config,
)
```

---

## Helper Functions

### Record Access

#### Get[T]
```go
func Get[T any](record Record, key string) (T, bool)
```
Safely gets typed value from Record. Includes automatic type conversion for numeric and string types.

**Example:**
```go
name, exists := streamv3.Get[string](record, "name")
if !exists {
    log.Println("name field not found")
}

// Type conversion: string "42" → int64(42)
age, ok := streamv3.Get[int64](record, "age")
```

#### GetOr[T]
```go
func GetOr[T any](record Record, key string, defaultValue T) T
```
Gets value with default fallback. Includes automatic type conversion.

**Example:**
```go
age := streamv3.GetOr(record, "age", int64(0))
name := streamv3.GetOr(record, "name", "Unknown")
```

#### Set[V]
```go
func Set[V Value](m MutableRecord, field string, value V) MutableRecord
```
Sets field value in a MutableRecord (mutates in place).

**Example:**
```go
mut := streamv3.MakeMutableRecord()
streamv3.Set(mut, "name", "Alice")
streamv3.Set(mut, "age", int64(30))
record := mut.Freeze()
```

#### SetImmutable[V]
```go
func SetImmutable[V Value](r Record, field string, value V) Record
```
Creates a new Record with the field value set (immutable operation).

**Example:**
```go
updated := streamv3.SetImmutable(record, "processed", true)
// original record is unchanged
```

#### Field[V]
```go
func Field[V Value](key string, value V) Record
```
Creates a single-field Record.

**Example:**
```go
nameField := streamv3.Field("name", "Alice")
// Returns: Record{"name": "Alice"}
```

#### ValidateRecord
```go
func ValidateRecord(r Record) error
```
Validates that a Record contains only supported value types.

**Example:**
```go
err := streamv3.ValidateRecord(record)
if err != nil {
    log.Printf("Invalid record: %v", err)
}
```

---

## Error Handling

> 🛡️ **Production Patterns**: Learn robust error handling strategies in the [Advanced Tutorial](advanced-tutorial.md#error-handling-and-resilience).

StreamV3 provides multiple error handling approaches:

### Source Functions (I/O Operations)
Source functions that read from files or execute commands return errors explicitly:

```go
// Source functions return (iter.Seq[Record], error)
data, err := streamv3.ReadCSV("data.csv")
if err != nil {
    log.Fatalf("Failed to open file: %v", err)
}

records, err := streamv3.ReadJSON("data.jsonl")
if err != nil {
    log.Fatalf("Failed to open file: %v", err)
}

lines, err := streamv3.ReadLines("logfile.txt")
if err != nil {
    log.Fatalf("Failed to open file: %v", err)
}

processes, err := streamv3.ExecCommand("ps", []string{"-efl"})
if err != nil {
    log.Fatalf("Failed to execute command: %v", err)
}
```

### Sink Functions (Write Operations)
Sink functions that write to files also return errors:

```go
err := streamv3.WriteCSV(records, "output.csv")
if err != nil {
    log.Fatalf("Failed to write CSV: %v", err)
}

err = streamv3.WriteJSON(records, "output.jsonl")
if err != nil {
    log.Fatalf("Failed to write JSON: %v", err)
}
```

### Filter Operations
Filter functions have two versions for transformation error handling:

- **Regular functions**: Designed for transformations that don't fail
- **Safe functions**: Return errors via `iter.Seq2[T, error]` for error-prone transformations

**Example:**
```go
// Regular filter - for transformations that don't fail
result := streamv3.Select(func(x int) int {
    return x * 2
})(data)

// Safe filter - for transformations that can fail
safeResult := streamv3.SelectSafe(func(x int) (int, error) {
    if x < 0 {
        return 0, fmt.Errorf("negative value: %d", x)
    }
    return x * 2, nil
})(dataWithErrors)

for value, err := range safeResult {
    if err != nil {
        log.Printf("Error: %v", err)
        continue
    }
    // Process value
}
```

### Best Practices

1. **Always check errors from Source and Sink functions** - These involve I/O and can fail
2. **Use Safe filters for user input or external data** - Where validation is needed
3. **Use regular filters for pure transformations** - Cleaner and more efficient
4. **Fail fast at the source** - Catch file/command errors before processing begins

---

## Best Practices

1. **Chain Operations**: Use functional composition for readable pipelines
2. **Use Type Safety**: Leverage generics for compile-time safety
3. **Handle Errors**: Use Safe versions for error-prone operations
4. **Memory Efficiency**: Use lazy evaluation and avoid materializing large datasets
5. **Performance**: Use appropriate window sizes and batch operations

## Related Documentation

- **[Getting Started Guide](codelab-intro.md)** - Learn StreamV3 basics with hands-on examples
- **[Advanced Tutorial](advanced-tutorial.md)** - Complex patterns, performance optimization, and real-world use cases

---

*Generated for StreamV3 - Modern Go stream processing library*