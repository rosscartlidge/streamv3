# StreamV3 AI Code Generation Guide

*Enabling LLMs to generate high-quality StreamV3 code from natural language descriptions*

## Overview

This document provides a comprehensive prompt and guidance system that enables any Large Language Model (LLM) to generate accurate, idiomatic StreamV3 code from natural language descriptions. Whether you're using Claude, GPT, Gemini, or other models, this guide will help them understand StreamV3 patterns and produce working solutions.

## Table of Contents

- [Master Prompt Template](#master-prompt-template)
- [Core Concepts for LLMs](#core-concepts-for-llms)
- [Pattern Recognition Examples](#pattern-recognition-examples)
- [Common Use Case Templates](#common-use-case-templates)
- [Best Practices for Code Generation](#best-practices-for-code-generation)
- [Error Patterns to Avoid](#error-patterns-to-avoid)
- [Testing Generated Code](#testing-generated-code)

---

## Master Prompt Template

Use this prompt with your LLM of choice to enable StreamV3 code generation:

```
You are an expert Go developer specializing in StreamV3, a modern Go stream processing library. Your task is to generate high-quality, idiomatic StreamV3 code from natural language descriptions.

## StreamV3 Core Knowledge

### Essential Imports
```go
import (
    "fmt"
    "slices"
    "iter"
    "time"
    "github.com/rosscartlidge/streamv3"
)
```

### Core Types
- `Stream[T]` - Lazy sequence of values implementing `iter.Seq[T]`
- `Record` - Map-based data structure: `map[string]any`
- `Filter[T, U any]` - Function type: `func(iter.Seq[T]) iter.Seq[U]`
- `FilterSameType[T any]` - Function type: `func(iter.Seq[T]) iter.Seq[T]`

### Stream Creation
- `streamv3.From([]T)` - Create from slice
- `streamv3.ReadCSV(filename)` - Read CSV file
- `streamv3.ReadJSON[T](filename)` - Read JSON file
- `streamv3.NewRecord().String("key", "value").Int("num", 42).Build()` - Build records

### Core Operations (SQL-style naming)
- **Transform**: `Select(fn)`, `SelectMany(fn)` (FlatMap)
- **Filter**: `Where(predicate)`, `Distinct()`, `DistinctBy(keyFn)`
- **Limit**: `Limit(n)`, `Offset(n)`
- **Sort**: `Sort()`, `SortBy(keyFn)`, `SortDesc()`, `Reverse()`
- **Aggregate**: `GroupByFields(groupName, ...fields)`, `Aggregate(groupName, aggregations)`
- **Join**: `InnerJoin(rightSeq, predicate)`, `LeftJoin()`, etc.
- **Window**: `CountWindow(size)`, `TimeWindow(duration, timeField)`, `SlidingCountWindow(size, step)`
- **Early Termination**: `TakeWhile(predicate)`, `TakeUntil(predicate)`, `Timeout(duration)`

### Record Access
- `streamv3.Get[T](record, "key")` returns `(T, bool)`
- `streamv3.GetOr(record, "key", defaultValue)` returns `T`
- `streamv3.SetField(record, "key", value)` returns modified record

### Aggregation Functions
- `streamv3.Count()`, `streamv3.Sum("field")`, `streamv3.Avg("field")`
- `streamv3.Min[T]("field")`, `streamv3.Max[T]("field")`
- `streamv3.First("field")`, `streamv3.Last("field")`, `streamv3.Collect("field")`

### Visualization
- `streamv3.QuickChart(data, "filename.html")` - Default chart
- `streamv3.InteractiveChart(data, filename, config)` - Custom chart

## Code Generation Rules

1. **Always use functional composition** - chain operations using `()()`
2. **Use SQL-style naming** - `Select` not `Map`, `Where` not `Filter`
3. **Handle types carefully** - Go generics require explicit type parameters when ambiguous
4. **Follow Go conventions** - proper error handling, clear variable names
5. **Include necessary imports** - add all required packages
6. **Create complete examples** - include main function and execution
7. **Use Record builder pattern** - `NewRecord().String(...).Int(...).Build()`
8. **Process in logical steps** - break complex pipelines into readable chunks

## Pattern Matching Guide

When you see these phrases, use these patterns:

- "filter/where/only" → `streamv3.Where(predicate)`
- "transform/convert/map" → `streamv3.Select(transformFn)`
- "group by" → `streamv3.GroupByFields(groupName, ...fields)`
- "count/sum/average" → `streamv3.Aggregate(groupName, map[string]streamv3.AggregateFunc{...})`
- "first N/top N/limit" → `streamv3.Limit(n)`
- "sort by/order by" → `streamv3.SortBy(keyFn)`
- "join/combine" → `streamv3.InnerJoin(rightSeq, predicate)`
- "in batches/windows" → `streamv3.CountWindow(size)` or `streamv3.TimeWindow(duration, timeField)`
- "chart/visualize/plot" → `streamv3.QuickChart()` or `streamv3.InteractiveChart()`

Generate clean, working Go code that follows these patterns and includes all necessary imports and error handling.
```

---

## Core Concepts for LLMs

### Stream Processing Mental Model

Help LLMs understand that StreamV3 uses:

1. **Lazy Evaluation** - Operations are not executed until iteration begins
2. **Functional Composition** - Chain operations using `operation(params)(previousStream)`
3. **Type Safety** - Go generics provide compile-time type checking
4. **SQL Semantics** - Operations mirror SQL concepts (SELECT, WHERE, GROUP BY, etc.)

### Key Patterns to Teach

```go
// Basic pattern: Create → Transform → Process
data := streamv3.From([]int{1, 2, 3, 4, 5})
result := streamv3.Select(func(x int) int { return x * 2 })(
    streamv3.Where(func(x int) bool { return x > 2 })(data)
)

// Record processing pattern
records := streamv3.ReadCSV("data.csv")
processed := streamv3.Select(func(r streamv3.Record) streamv3.Record {
    value := streamv3.GetOr(r, "amount", 0.0)
    return streamv3.SetField(r, "doubled", value * 2)
})(records)

// Aggregation pattern
grouped := streamv3.GroupByFields("sales", "region")(records)
aggregated := streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
    "total": streamv3.Sum("amount"),
    "count": streamv3.Count(),
    "avg":   streamv3.Avg("amount"),
})(grouped)
```

---

## Pattern Recognition Examples

### Example 1: Data Analysis Request

**Natural Language**: "Read sales data from sales.csv, filter for amounts over $100, group by region, and calculate total sales per region"

**Generated Code**:
```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read sales data
    salesData, err := streamv3.ReadCSV("sales.csv")
    if err != nil {
        panic(err)
    }

    // Filter for amounts over $100
    filtered := streamv3.Where(func(r streamv3.Record) bool {
        amount := streamv3.GetOr(r, "amount", 0.0)
        return amount > 100
    })(salesData)

    // Group by region
    grouped := streamv3.GroupByFields("sales", "region")(filtered)

    // Calculate total sales per region
    results := streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
        "total_sales": streamv3.Sum("amount"),
    })(grouped)

    // Display results
    for result := range results {
        region := streamv3.GetOr(result, "GroupValue", "")
        total := streamv3.GetOr(result, "total_sales", 0.0)
        fmt.Printf("Region %s: $%.2f\n", region, total)
    }
}
```

### Example 2: Real-Time Processing

**Natural Language**: "Process sensor data in 10-second windows, calculate average temperature, and alert if temperature exceeds 35°C"

**Generated Code**:
```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Simulate real-time sensor data
    sensorData := func(yield func(streamv3.Record) bool) {
        for i := 0; ; i++ {
            record := streamv3.NewRecord().
                String("sensor_id", fmt.Sprintf("temp_%d", i%3)).
                Float("temperature", 20 + float64(i%30)).
                Time("timestamp", time.Now()).
                Build()

            if !yield(record) {
                return
            }
            time.Sleep(100 * time.Millisecond)
        }
    }

    // Process in 10-second windows
    windowed := streamv3.TimeWindow[streamv3.Record](
        10*time.Second,
        "timestamp",
    )(streamv3.Limit[streamv3.Record](100)(sensorData)) // Limit for demo

    fmt.Println("Temperature Monitoring:")
    for window := range windowed {
        if len(window) == 0 {
            continue
        }

        // Calculate average temperature
        var total float64
        for _, record := range window {
            temp := streamv3.GetOr(record, "temperature", 0.0)
            total += temp
        }
        avg := total / float64(len(window))

        // Alert if temperature exceeds 35°C
        if avg > 35 {
            fmt.Printf("🚨 ALERT: Average temperature %.1f°C exceeds threshold!\n", avg)
        } else {
            fmt.Printf("✅ Normal: Average temperature %.1f°C\n", avg)
        }
    }
}
```

### Example 3: Data Visualization

**Natural Language**: "Create a chart showing monthly revenue trends from the sales data"

**Generated Code**:
```go
package main

import (
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read sales data
    salesData, err := streamv3.ReadCSV("sales.csv")
    if err != nil {
        panic(err)
    }

    // Group by month and calculate revenue
    grouped := streamv3.GroupByFields("monthly_sales", "month")(salesData)

    monthlyRevenue := streamv3.Aggregate("monthly_sales", map[string]streamv3.AggregateFunc{
        "revenue": streamv3.Sum("amount"),
    })(grouped)

    // Create interactive chart
    config := streamv3.DefaultChartConfig()
    config.Title = "Monthly Revenue Trends"
    config.ChartType = "line"

    err = streamv3.InteractiveChart(
        monthlyRevenue,
        "monthly_revenue_chart.html",
        config,
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("Chart created: monthly_revenue_chart.html")
}
```

---

## Common Use Case Templates

### Template 1: CSV Analysis Pipeline

```go
// For: "Analyze CSV data with grouping and aggregation"
func analyzeCSVData(filename, groupField string, aggregations map[string]streamv3.AggregateFunc) {
    data, err := streamv3.ReadCSV(filename)
    if err != nil {
        panic(err)
    }

    grouped := streamv3.GroupByFields("analysis", groupField)(data)
    results := streamv3.Aggregate("analysis", aggregations)(grouped)

    for result := range results {
        // Process results...
    }
}
```

### Template 2: Real-Time Stream Processing

```go
// For: "Process infinite data streams with windowing"
func processRealTimeStream(windowSize time.Duration) {
    dataStream := createInfiniteStream() // User defines this

    windowed := streamv3.TimeWindow[streamv3.Record](windowSize, "timestamp")(dataStream)

    for window := range windowed {
        // Process each window...
    }
}
```

### Template 3: Data Transformation Pipeline

```go
// For: "Transform and filter data"
func transformData(input streamv3.Stream[streamv3.Record]) streamv3.Stream[streamv3.Record] {
    return streamv3.Select(func(r streamv3.Record) streamv3.Record {
        // Transform logic here
        return r
    })(streamv3.Where(func(r streamv3.Record) bool {
        // Filter logic here
        return true
    })(input))
}
```

---

## Best Practices for Code Generation

### 1. Always Include Error Handling

```go
// GOOD
data, err := streamv3.ReadCSV("file.csv")
if err != nil {
    return err
}

// BAD - ignoring errors
data := streamv3.ReadCSV("file.csv") // This won't compile
```

### 2. Use Type-Safe Record Access

```go
// GOOD
amount := streamv3.GetOr(record, "amount", 0.0)
name, exists := streamv3.Get[string](record, "name")

// BAD - direct map access
amount := record["amount"].(float64) // Potential panic
```

### 3. Prioritize Human Readability

```go
// ❌ TOO COMPLEX - Hard to verify
result := streamv3.Limit(10)(streamv3.SortBy(func(r streamv3.Record) float64 {
    return -streamv3.GetOr(r, "revenue", 0.0)
})(streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
    "total": streamv3.Sum("amount"), "count": streamv3.Count(),
})(streamv3.GroupByFields("sales", "region")(data))))

// ✅ CLEAR AND SIMPLE - Easy to verify step by step
groupedSales := streamv3.GroupByFields("sales", "region")(data)

regionTotals := streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
    "total_revenue": streamv3.Sum("amount"),
    "sale_count": streamv3.Count(),
})(groupedSales)

sortedByRevenue := streamv3.SortBy(func(r streamv3.Record) float64 {
    return -streamv3.GetOr(r, "total_revenue", 0.0) // Negative for descending
})(regionTotals)

top10Regions := streamv3.Limit[streamv3.Record](10)(sortedByRevenue)
```

### 4. Use Descriptive Variable Names

```go
// GOOD
salesData := streamv3.ReadCSV("sales.csv")
highValueSales := streamv3.Where(func(r streamv3.Record) bool {
    return streamv3.GetOr(r, "amount", 0.0) > 1000
})(salesData)

// BAD
d := streamv3.ReadCSV("sales.csv")
f := streamv3.Where(func(r streamv3.Record) bool {
    return streamv3.GetOr(r, "amount", 0.0) > 1000
})(d)
```

---

## Error Patterns to Avoid

### 1. Incorrect Function Names (Use SQL-style)

```go
// WRONG - Old API
result := streamv3.Map(fn)(data)          // Use Select instead
result := streamv3.Filter(predicate)(data) // Use Where instead
result := streamv3.Take(10)(data)         // Use Limit instead

// CORRECT - Current API
result := streamv3.Select(fn)(data)
result := streamv3.Where(predicate)(data)
result := streamv3.Limit(10)(data)
```

### 2. Missing Type Parameters

```go
// WRONG - Ambiguous types
windowed := streamv3.CountWindow(10)(data)

// CORRECT - Explicit types when needed
windowed := streamv3.CountWindow[streamv3.Record](10)(data)
```

### 3. Incorrect Aggregation Syntax

```go
// WRONG - Missing group field name
aggregated := streamv3.Aggregate(map[string]streamv3.AggregateFunc{
    "total": streamv3.Sum("amount"),
})(grouped)

// CORRECT - Include group field name
aggregated := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
    "total": streamv3.Sum("amount"),
})(grouped)
```

### 4. Improper Record Building

```go
// WRONG - Manual map construction
record := streamv3.Record{
    "name": "Alice",
    "age": 30,
}

// CORRECT - Use builder pattern
record := streamv3.NewRecord().
    String("name", "Alice").
    Int("age", 30).
    Build()
```

---

## Testing Generated Code

### Quick Validation Checklist

1. **Imports**: All necessary packages included
2. **Error Handling**: File operations check for errors
3. **Type Safety**: Explicit type parameters where needed
4. **Function Names**: Using current API (Select, Where, Limit, etc.)
5. **Record Access**: Using Get/GetOr instead of direct map access
6. **Complete Example**: Includes main function and execution flow

### Example Test Template

```go
package main

import (
    "fmt"
    "testing"
    "github.com/rosscartlidge/streamv3"
)

func TestGeneratedCode(t *testing.T) {
    // Create test data
    testData := []streamv3.Record{
        streamv3.NewRecord().String("name", "Alice").Int("age", 30).Build(),
        streamv3.NewRecord().String("name", "Bob").Int("age", 25).Build(),
    }

    stream := streamv3.From(testData)

    // Test the generated pipeline
    result := streamv3.Where(func(r streamv3.Record) bool {
        age := streamv3.GetOr(r, "age", 0)
        return age > 26
    })(stream)

    count := 0
    for range result {
        count++
    }

    if count != 1 {
        t.Errorf("Expected 1 result, got %d", count)
    }
}
```

---

## Usage Instructions

1. **For Users**: Copy the master prompt template and paste it into your LLM session before asking for StreamV3 code generation.

2. **For LLM Integration**: Include this documentation in your knowledge base to improve StreamV3 code generation accuracy.

3. **For Custom Applications**: Use the pattern recognition examples to train domain-specific code generators.

4. **For Validation**: Use the error patterns and testing sections to verify generated code quality.

---

## Example LLM Session

**User**: "I want to analyze log files to find error patterns. Read access.log, filter for 5xx status codes, group by hour, and count errors per hour."

**LLM Response** (using this guide):
```go
package main

import (
    "fmt"
    "time"
    "strings"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read log file
    logData, err := streamv3.ReadCSV("access.log")
    if err != nil {
        panic(err)
    }

    // Filter for 5xx status codes
    errorLogs := streamv3.Where(func(r streamv3.Record) bool {
        status := streamv3.GetOr(r, "status", "")
        return strings.HasPrefix(status, "5")
    })(logData)

    // Add hour field for grouping
    withHour := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        timestamp := streamv3.GetOr(r, "timestamp", "")
        // Parse timestamp and extract hour
        if t, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
            hour := t.Format("2006-01-02 15")
            return streamv3.SetField(r, "hour", hour)
        }
        return r
    })(errorLogs)

    // Group by hour
    grouped := streamv3.GroupByFields("error_analysis", "hour")(withHour)

    // Count errors per hour
    results := streamv3.Aggregate("error_analysis", map[string]streamv3.AggregateFunc{
        "error_count": streamv3.Count(),
    })(grouped)

    // Display results
    fmt.Println("Error Patterns by Hour:")
    for result := range results {
        hour := streamv3.GetOr(result, "GroupValue", "")
        count := streamv3.GetOr(result, "error_count", 0)
        fmt.Printf("%s: %d errors\n", hour, count)
    }
}
```

This comprehensive guide enables any LLM to become proficient at generating StreamV3 code from natural language descriptions!