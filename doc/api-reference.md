# StreamV3 API Reference

*Complete reference for all StreamV3 types, functions, and methods*

## Installation

```bash
go get github.com/rosscartlidge/streamv3
```

**Requires:** Go 1.23+ (for iterator support)

## Package Import

```go
import "github.com/rosscartlidge/streamv3"
```

---

## Core Types

### Stream[T]

The fundamental type representing a lazy sequence of values.

```go
type Stream[T] interface {
    iter.Seq[T]
}
```

Created by:
- `From[T]([]T) Stream[T]`
- `FromIter[T](iter.Seq[T]) Stream[T]`
- `FromChannel[T](<-chan T) Stream[T]`

### Record

A flexible map-based data structure for heterogeneous data.

```go
type Record map[string]any
```

**Supported value types:** `int`, `int64`, `float64`, `string`, `bool`, `time.Time`, nested `Record`, and `Stream` types.

### Filter[T, U]

A function that transforms one iterator to another.

```go
type Filter[T, U any] func(iter.Seq[T]) iter.Seq[U]
```

Used for functional composition with `Pipe`, `Chain`, etc.

---

## Stream Creation

### From
```go
func From[T any](data []T) *StreamBuilder[T]
```
Creates a stream from a slice.

**Example:**
```go
numbers := streamv3.From([]int{1, 2, 3, 4, 5})
```

### FromIter
```go
func FromIter[T any](seq iter.Seq[T]) *StreamBuilder[T]
```
Creates a stream from an existing iterator.

### FromChannel
```go
func FromChannel[T any](ch <-chan T) *StreamBuilder[T]
```
Creates a stream from a channel.

### ReadCSV
```go
func ReadCSV(filename string) *StreamBuilder[Record]
```
Reads CSV file and returns stream of Records.

**Error-aware version:**
```go
func ReadCSVSafe(filename string) iter.Seq2[Record, error]
```

### ReadJSON
```go
func ReadJSON(filename string) *StreamBuilder[Record]
```
Reads JSON file (array of objects) and returns stream of Records.

### ExecCommand
```go
func ExecCommand(command string, args ...string) *StreamBuilder[Record]
```
Executes command and parses output as Records with automatic field detection.

**Example:**
```go
processes := streamv3.ExecCommand("ps", "aux")
```

---

## Stream Operations (Fluent API)

### Filtering

#### Where
```go
func (s *StreamBuilder[T]) Where(predicate func(T) bool) *StreamBuilder[T]
```
Filters elements based on predicate.

**Example:**
```go
evens := numbers.Where(func(x int) bool { return x%2 == 0 })
```

#### Limit
```go
func (s *StreamBuilder[T]) Limit(n int) *StreamBuilder[T]
```
Takes first n elements.

#### Skip
```go
func (s *StreamBuilder[T]) Skip(n int) *StreamBuilder[T]
```
Skips first n elements.

#### Distinct
```go
func (s *StreamBuilder[T]) Distinct() *StreamBuilder[T]
```
Removes duplicate elements.

### Transformation

#### Map
```go
func (s *StreamBuilder[T]) Map(fn func(T) U) *StreamBuilder[U]
```
Transforms each element.

**Example:**
```go
squares := numbers.Map(func(x int) int { return x * x })
```

#### FlatMap
```go
func (s *StreamBuilder[T]) FlatMap(fn func(T) Stream[U]) *StreamBuilder[U]
```
Maps each element to a stream and flattens results.

### Sorting

#### Sort
```go
func (s *StreamBuilder[T]) Sort(less func(T, T) bool) *StreamBuilder[T]
```
Sorts elements using comparison function.

#### SortByKey
```go
func (s *StreamBuilder[T]) SortByKey(keyFn func(T) K, ascending bool) *StreamBuilder[T]
```
Sorts by extracted key.

**Example:**
```go
sorted := people.SortByKey(func(p Record) string {
    return p["name"].(string)
}, true)
```

### Aggregation

#### Collect
```go
func (s *StreamBuilder[T]) Collect() []T
```
Materializes stream into slice.

#### Reduce
```go
func (s *StreamBuilder[T]) Reduce(initial U, fn func(U, T) U) U
```
Reduces stream to single value.

#### Count
```go
func (s *StreamBuilder[T]) Count() int
```
Counts elements in stream.

#### First
```go
func (s *StreamBuilder[T]) First() (T, bool)
```
Returns first element and whether it exists.

#### Last
```go
func (s *StreamBuilder[T]) Last() (T, bool)
```
Returns last element and whether it exists.

---

## Functional API

### Core Operations

#### Map
```go
func Map[T, U any](fn func(T) U) Filter[T, U]
```

#### Where
```go
func Where[T any](predicate func(T) bool) FilterSameType[T]
```

#### Limit
```go
func Limit[T any](n int) FilterSameType[T]
```

#### Skip
```go
func Skip[T any](n int) FilterSameType[T]
```

### Composition

#### Pipe
```go
func Pipe[T, U, V any](f1 Filter[T, U], f2 Filter[U, V]) Filter[T, V]
```
Composes two filters sequentially.

**Example:**
```go
pipeline := streamv3.Pipe(
    streamv3.Where(func(x int) bool { return x > 5 }),
    streamv3.Map(func(x int) int { return x * 2 }),
)
result := streamv3.Collect(pipeline(streamv3.From(numbers)))
```

#### Chain
```go
func Chain[T any](filters ...FilterSameType[T]) FilterSameType[T]
```
Chains multiple same-type filters.

---

## Record Operations

### Record Builder

#### NewRecord
```go
func NewRecord() *TypedRecord
```
Creates a new type-safe Record builder.

**Methods:**
```go
func (tr *TypedRecord) String(key, value string) *TypedRecord
func (tr *TypedRecord) Int(key string, value int64) *TypedRecord
func (tr *TypedRecord) Float(key string, value float64) *TypedRecord
func (tr *TypedRecord) Bool(key string, value bool) *TypedRecord
func (tr *TypedRecord) Time(key string, value time.Time) *TypedRecord
func (tr *TypedRecord) Set(key string, value any) *TypedRecord
func (tr *TypedRecord) Build() Record
```

**Example:**
```go
record := streamv3.NewRecord().
    String("name", "Alice").
    Int("age", 30).
    Float("score", 95.5).
    Build()
```

### Record Stream Operations

#### GroupRecordsByFields
```go
func GroupRecordsByFields(stream Stream[Record], fields ...string) Stream[RecordGroup]
```
Groups records by specified field values.

**RecordGroup:**
```go
type RecordGroup struct {
    Key     Record    // Group key fields
    Records []Record  // Records in group
}
```

#### AggregateGroups
```go
func AggregateGroups(groups Stream[RecordGroup],
                    aggregates map[string]AggregateFunc) Stream[Record]
```
Applies aggregation functions to grouped records.

**Example:**
```go
groups := streamv3.GroupRecordsByFields(sales, "region")
results := streamv3.AggregateGroups(groups, map[string]streamv3.AggregateFunc{
    "total_revenue": streamv3.Sum("amount"),
    "avg_deal":      streamv3.Avg("deal_size"),
    "count":         streamv3.Count(),
})
```

---

## Aggregation Functions

### Numeric Aggregations

#### Sum
```go
func Sum(field string) AggregateFunc
```

#### Avg
```go
func Avg(field string) AggregateFunc
```

#### Min
```go
func Min(field string) AggregateFunc
```

#### Max
```go
func Max(field string) AggregateFunc
```

#### StdDev
```go
func StdDev(field string) AggregateFunc
```

### Other Aggregations

#### Count
```go
func Count() AggregateFunc
```

#### First
```go
func First(field string) AggregateFunc
```

#### Last
```go
func Last(field string) AggregateFunc
```

#### Collect
```go
func Collect(field string) AggregateFunc
```
Collects all values into a slice.

#### Custom
```go
func Custom(fn func([]Record) interface{}) AggregateFunc
```
Creates custom aggregation function.

---

## Visualization

### Quick Charts

#### QuickChart
```go
func QuickChart(data Stream[Record], xField, yField, filename string) error
```
Creates interactive chart with default settings.

**Example:**
```go
streamv3.QuickChart(salesData, "month", "revenue", "sales.html")
```

### Advanced Charts

#### InteractiveChart
```go
func InteractiveChart(data Stream[Record], filename string, config ChartConfig) error
```

#### TimeSeriesChart
```go
func TimeSeriesChart(data Stream[Record], timeField string,
                    valueFields []string, filename string, config ChartConfig) error
```

**Example:**
```go
streamv3.TimeSeriesChart(metrics, "timestamp",
    []string{"cpu_usage", "memory_usage"}, "metrics.html", config)
```

### Chart Configuration

#### ChartConfig
```go
type ChartConfig struct {
    Title              string            // Chart title
    Width              int               // Chart width (default: 1200)
    Height             int               // Chart height (default: 600)
    ChartType          string            // line, bar, scatter, pie, etc.
    TimeFormat         string            // Time format for time axes
    XAxisType          string            // linear, logarithmic, time, category
    YAxisType          string            // linear, logarithmic
    ShowLegend         bool              // Show legend
    ShowTooltips       bool              // Show tooltips
    EnableZoom         bool              // Enable zoom/pan
    EnableAnimations   bool              // Enable animations
    ShowDataLabels     bool              // Show data labels
    EnableInteractive  bool              // Enable field selection UI
    EnableCalculations bool              // Enable trend lines, moving averages
    ColorScheme        string            // default, vibrant, pastel, monochrome
    Theme              string            // light, dark
    ExportFormats      []string          // png, svg, pdf, csv
    CustomCSS          string            // Custom CSS
    Fields             map[string]string // Field type hints
}
```

#### DefaultChartConfig
```go
func DefaultChartConfig() ChartConfig
```
Returns configuration with sensible defaults.

**Chart Types:**
- `"line"` - Line chart
- `"bar"` - Bar chart
- `"scatter"` - Scatter plot
- `"pie"` - Pie chart
- `"doughnut"` - Doughnut chart
- `"radar"` - Radar chart
- `"polarArea"` - Polar area chart

---

## I/O Operations

### CSV Operations

#### ReadCSV
```go
func ReadCSV(filename string) Stream[Record]
func ReadCSVSafe(filename string) iter.Seq2[Record, error]
```

#### WriteCSV
```go
func WriteCSV(data Stream[Record], filename string) error
```

### JSON Operations

#### ReadJSON
```go
func ReadJSON(filename string) Stream[Record]
func ReadJSONSafe(filename string) iter.Seq2[Record, error]
```

#### WriteJSON
```go
func WriteJSON(data Stream[Record], filename string) error
```

### Command Execution

#### ExecCommand
```go
func ExecCommand(command string, args ...string) Stream[Record]
```
Executes command and parses output with automatic field detection.

**Supported commands:**
- `ps` - Process listings
- `top` - System monitoring
- `df` - Disk usage
- `netstat` - Network connections
- Custom commands with tabular output

---

## Error Handling

### Error-Aware Streams

Error-aware operations return `iter.Seq2[T, error]` instead of `iter.Seq[T]`.

#### Safe Operations
```go
func ReadCSVSafe(filename string) iter.Seq2[Record, error]
func ReadJSONSafe(filename string) iter.Seq2[Record, error]
func MapSafe[T, U any](fn func(T) (U, error)) FilterWithErrors[T, U]
```

### Conversion Utilities

#### Safe
```go
func Safe[T any](seq iter.Seq[T]) iter.Seq2[T, error]
```
Converts simple iterator to error-aware (never errors).

#### Unsafe
```go
func Unsafe[T any](seq iter.Seq2[T, error]) iter.Seq[T]
```
Converts error-aware iterator to simple (panics on errors).

#### IgnoreErrors
```go
func IgnoreErrors[T any](seq iter.Seq2[T, error]) iter.Seq[T]
```
Converts error-aware iterator to simple (ignores errors).

---

## Utility Functions

### Collection Utilities

#### Collect
```go
func Collect[T any](stream Stream[T]) []T
```

#### ToSlice
```go
func ToSlice[T any](seq iter.Seq[T]) []T
```

#### Length
```go
func Length[T any](seq iter.Seq[T]) int
```

### Comparison Utilities

#### Equal
```go
func Equal[T comparable](a, b Stream[T]) bool
```

#### Contains
```go
func Contains[T comparable](stream Stream[T], value T) bool
```

---

## Performance Notes

### Memory Efficiency

- **Lazy evaluation:** Operations are not executed until terminal operation (`Collect`, `Reduce`, etc.)
- **Streaming:** Large datasets processed without loading into memory
- **Iterator-based:** Built on Go 1.23 iterators for optimal performance

### Best Practices

1. **Use streaming operations** for large datasets
2. **Filter early** to reduce processing overhead
3. **Prefer functional composition** for reusable pipelines
4. **Use error-aware streams** for production code
5. **Batch operations** when processing very large streams

### Benchmarks

Performance characteristics (typical workloads):

- **CSV parsing:** 100K+ records/second
- **Aggregations:** 500K+ records/second
- **Filtering:** 1M+ records/second
- **Memory usage:** O(1) for streaming operations

---

## Examples

### Basic Processing
```go
result := streamv3.From(data).
    Where(condition).
    Map(transform).
    SortByKey(keyFunc, true).
    Limit(10).
    Collect()
```

### Functional Composition
```go
pipeline := streamv3.Pipe(
    streamv3.Where(condition),
    streamv3.Map(transform),
    streamv3.Limit(10),
)
result := streamv3.Collect(pipeline(streamv3.From(data)))
```

### Aggregation
```go
groups := streamv3.GroupRecordsByFields(data, "category")
results := streamv3.AggregateGroups(groups, map[string]streamv3.AggregateFunc{
    "total":   streamv3.Sum("amount"),
    "average": streamv3.Avg("amount"),
    "count":   streamv3.Count(),
})
```

### Visualization
```go
config := streamv3.DefaultChartConfig()
config.Title = "Sales Analysis"
config.ChartType = "line"
config.EnableCalculations = true

streamv3.InteractiveChart(data, "analysis.html", config)
```

---

## Migration from StreamV2

StreamV3 is designed to be familiar to StreamV2 users:

### Key Changes
- Uses Go 1.23 `iter.Seq[T]` instead of custom iterator interface
- Enhanced type safety with generics
- Built-in visualization capabilities
- Improved error handling with `iter.Seq2[T, error]`

### Migration Examples

**StreamV2:**
```go
stream := stream.From(data).Filter(pred).Map(fn).Collect()
```

**StreamV3:**
```go
result := streamv3.From(data).Where(pred).Map(fn).Collect()
```

Most operations have direct equivalents with similar names.

---

## See Also

- [Introduction Codelab](codelab-intro.md) - Getting started
- [Advanced Tutorial](advanced-tutorial.md) - Complex use cases
- [Chart Gallery](doc/chart_examples/) - Interactive visualization examples
- [GitHub Repository](https://github.com/rosscartlidge/streamv3) - Source code and issues

---

*This reference covers StreamV3 v1.0. For the latest updates, see the GitHub repository.*