# StreamV3 API Reference

*Complete reference for all StreamV3 types, functions, and methods*

## Table of Contents

### Documentation Navigation
- [Getting Started Guide](codelab-intro.md) - Learn StreamV3 basics step-by-step
- [Advanced Tutorial](advanced-tutorial.md) - Complex patterns and real-world examples

### API Reference Sections
- [Installation & Setup](#installation--setup)
- [Core Types](#core-types)
- [Stream Creation](#stream-creation)
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
- [Utility Operations](#utility-operations)
- [I/O Operations](#io-operations)
- [Chart & Visualization](#chart--visualization)

---

## Installation & Setup

### Requirements
- **Go 1.23+** (required for iterator support)

### Installation
```bash
go get github.com/rosscartlidge/streamv3
```

### Package Import
```go
import "github.com/rosscartlidge/streamv3"
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
A flexible map-based data structure for heterogeneous data.

```go
type Record map[string]any
```

**Supported value types:** `int`, `int64`, `float64`, `string`, `bool`, `time.Time`, nested `Record`, and slices.

### Filter Types
Function types for stream transformations:

```go
type Filter[T, U any] func(iter.Seq[T]) iter.Seq[U]
type FilterSameType[T any] func(iter.Seq[T]) iter.Seq[T]
type FilterWithErrors[T, U any] func(iter.Seq2[T, error]) iter.Seq2[U, error]
```

### GroupedRecord
Result type from GroupBy operations:

```go
type GroupedRecord struct {
    GroupKey   string
    GroupValue any
    Records    []Record
}
```

---

## Creating Iterators

### From Slices
```go
slices.Values([]T) iter.Seq[T]
```
Creates an iterator from a slice (standard library function).

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

### NewRecord
```go
func NewRecord() *RecordBuilder
```
Creates a new record builder for constructing Records.

**Example:**
```go
record := streamv3.NewRecord().
    String("name", "Alice").
    Int("age", 30).
    Float("score", 95.5).
    Build()
```

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
func Where[T any](predicate func(T) bool) FilterSameType[T]
```
Filters elements based on a predicate (SQL WHERE equivalent).

**Example:**
```go
evens := streamv3.Where(func(x int) bool { return x%2 == 0 })(numbers)
```

### WhereSafe[T]
```go
func WhereSafe[T any](predicate func(T) (bool, error)) FilterWithErrorsSameType[T]
```
Safe version of Where that handles errors.

### Distinct[T]
```go
func Distinct[T comparable]() FilterSameType[T]
```
Removes duplicate elements.

### DistinctBy[T, K]
```go
func DistinctBy[T any, K comparable](keyFn func(T) K) FilterSameType[T]
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
func LimitSafe[T any](n int) FilterWithErrorsSameType[T]
```
Safe version of Limit that handles errors.

### Offset[T]
```go
func Offset[T any](n int) FilterSameType[T]
```
Skips the first n elements (SQL OFFSET equivalent).

### OffsetSafe[T]
```go
func OffsetSafe[T any](n int) FilterWithErrorsSameType[T]
```
Safe version of Offset that handles errors.

---

## Ordering Operations

*Functions for sorting and ordering streams*

### Sort[T]
```go
func Sort[T cmp.Ordered]() FilterSameType[T]
```
Sorts elements in ascending order.

### SortBy[T, K]
```go
func SortBy[T any, K cmp.Ordered](keyFn func(T) K) FilterSameType[T]
```
Sorts elements by a key function.

### SortDesc[T]
```go
func SortDesc[T cmp.Ordered]() FilterSameType[T]
```
Sorts elements in descending order.

### Reverse[T]
```go
func Reverse[T any]() FilterSameType[T]
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

#### InnerJoin
```go
func InnerJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) FilterSameType[Record]
```
Performs inner join between two record streams.

#### LeftJoin
```go
func LeftJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) FilterSameType[Record]
```
Performs left outer join.

#### RightJoin
```go
func RightJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) FilterSameType[Record]
```
Performs right outer join.

#### FullJoin
```go
func FullJoin(rightSeq iter.Seq[Record], predicate JoinPredicate) FilterSameType[Record]
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
func GroupBy[K comparable](sequenceField string, keyField string, keyFn func(Record) K) FilterSameType[Record]
```
Groups records by a key function.

#### GroupByFields
```go
func GroupByFields(sequenceField string, fields ...string) FilterSameType[Record]
```
Groups records by field values.

**Example:**
```go
grouped := streamv3.GroupByFields("sales_data", "region", "product")(records)
```

### Aggregation Functions

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
func Aggregate(sequenceField string, aggregations map[string]AggregateFunc) FilterSameType[Record]
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

## Utility Operations

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
func Chain[T any](filters ...FilterSameType[T]) FilterSameType[T]
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

#### ReadCSV
```go
func ReadCSV(filename string, config ...CSVConfig) iter.Seq[Record]
```
Reads CSV file into Record iterator. Panics on file errors.

#### ReadCSVFromReader
```go
func ReadCSVFromReader(reader io.Reader, config ...CSVConfig) iter.Seq[Record]
```
Reads CSV from any io.Reader.

#### WriteCSV
```go
func WriteCSV(stream iter.Seq[Record], filename string, fields []string, config ...CSVConfig) error
```
Writes Record iterator to CSV file.

### JSON Operations

#### ReadJSON
```go
func ReadJSON(filename string) iter.Seq[Record]
```
Reads JSON file into Record iterator. Panics on file errors.

#### WriteJSON
```go
func WriteJSON(stream iter.Seq[Record], filename string) error
```
Writes Record iterator to JSON file.

#### WriteJSONToWriter
```go
func WriteJSONToWriter(stream iter.Seq[Record], writer io.Writer) error
```
Writes Record iterator to any io.Writer as JSON.

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
    Title       string
    Width       int
    Height      int
    ChartType   string // "line", "bar", "scatter", "pie"
    Theme       string // "light", "dark"
    // ... more options
}
```

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
Safely gets typed value from Record.

#### GetOr[T]
```go
func GetOr[T any](record Record, key string, defaultValue T) T
```
Gets value with default fallback.

#### SetField
```go
func SetField[T any](record Record, key string, value T) Record
```
Sets field value in Record.

**Example:**
```go
name, exists := streamv3.Get[string](record, "name")
age := streamv3.GetOr(record, "age", 0)
updated := streamv3.SetField(record, "processed", true)
```

---

## Error Handling

> 🛡️ **Production Patterns**: Learn robust error handling strategies in the [Advanced Tutorial](advanced-tutorial.md#error-handling-and-resilience).

StreamV3 provides both safe and unsafe versions of operations:

- **Regular functions**: Panic on errors (fail-fast approach)
- **Safe functions**: Return errors via `iter.Seq2[T, error]`

**Example:**
```go
// Unsafe - panics on error
result := streamv3.Select(transform)(data)

// Safe - handles errors
safeResult := streamv3.SelectSafe(safeTransform)(dataWithErrors)
for value, err := range safeResult {
    if err != nil {
        log.Printf("Error: %v", err)
        continue
    }
    // Process value
}
```

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