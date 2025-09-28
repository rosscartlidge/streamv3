# StreamV3 🚀

A modern, type-safe Go library for functional stream processing with interactive data visualization. Built on Go 1.23+ with first-class support for iterators, generics, and functional composition.

## ✨ Features

### 🔄 Stream Processing
- **Functional composition** with type-safe generic operations
- **Go 1.23+ iterators** (`iter.Seq[T]` and `iter.Seq2[T,error]`)
- **Fluent API** for intuitive data transformations
- **Error-aware processing** with safe error propagation
- **Lazy evaluation** for memory-efficient operations

### 📊 Interactive Visualizations
- **Chart.js integration** with modern, responsive charts
- **Interactive field selection** - change X/Y axes dynamically
- **Multiple chart types** - line, bar, scatter, pie charts
- **Time series support** with zoom and pan capabilities
- **Statistical overlays** - trend lines, moving averages, min/max/mean
- **Modern UI** with Bootstrap 5 responsive design
- **Export capabilities** - PNG, CSV formats

### 🗂️ Data I/O
- **CSV/TSV/JSON** reading and writing
- **Command output parsing** (ps, top, etc.) with auto-column detection
- **Record-based processing** with flexible field types
- **SQL-style operations** (GROUP BY, aggregations)

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Create a stream of data
    numbers := streamv3.From([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

    // Functional composition - filter evens and take first 3
    result := numbers.
        Where(func(x int) bool { return x%2 == 0 }).
        Limit(3).
        Collect()

    fmt.Println(result) // [2, 4, 6]
}
```

## 📈 Interactive Charts

Create stunning interactive visualizations with a single function call:

```go
// Create sample data
data := streamv3.From([]streamv3.Record{
    streamv3.NewRecord().String("month", "Jan").Float("revenue", 120000).Build(),
    streamv3.NewRecord().String("month", "Feb").Float("revenue", 135000).Build(),
    streamv3.NewRecord().String("month", "Mar").Float("revenue", 118000).Build(),
})

// Generate interactive chart
streamv3.QuickChart(data, "month", "revenue", "revenue_chart.html")
```

### 🎨 Chart Demo

Run the comprehensive chart demo:

```bash
go run doc/examples/chart_demo.go
```

This creates 5 interactive HTML charts showcasing:
- **Sales Dashboard** - Business analytics with seasonal trends
- **System Metrics** - Time series monitoring with dark theme
- **Process Analysis** - Scatter plot visualization
- **Network Traffic** - Real-time network monitoring
- **Quick Example** - Simple revenue chart

## 🛠️ Advanced Features

### Functional Composition
```go
// Traditional filter composition
filter1 := func(s streamv3.Stream[int]) streamv3.Stream[int] {
    return streamv3.Where(s, func(x int) bool { return x > 5 })
}

filter2 := func(s streamv3.Stream[int]) streamv3.Stream[int] {
    return streamv3.Limit(s, 3)
}

// Compose filters
composedFilter := streamv3.Pipe(filter1, filter2)
result := composedFilter(streamv3.From([]int{1,2,3,4,5,6,7,8,9,10}))
```

### SQL-Style Operations
```go
// Group sales by region and calculate totals
sales := streamv3.ReadCSV("sales.csv")
grouped := streamv3.GroupRecordsByFields(sales, "region")
aggregated := streamv3.AggregateGroups(grouped, map[string]streamv3.AggregateFunc{
    "total_sales": streamv3.Sum("amount"),
    "avg_sales":   streamv3.Avg("amount"),
    "count":       streamv3.Count(),
})
```

### Command Output Processing
```go
// Parse ps command output automatically
processes := streamv3.ExecCommand("ps", "-eflww")
topMemory := processes.
    SortByKey(func(r streamv3.Record) float64 {
        return r["MEM"].(float64)
    }, false).
    Limit(10)
```

### Time Series Charts
```go
config := streamv3.DefaultChartConfig()
config.Title = "System Metrics Over Time"
config.EnableCalculations = true

streamv3.TimeSeriesChart(
    metricsData,
    "timestamp",
    []string{"cpu_usage", "memory_usage"},
    "metrics.html",
    config
)
```

## 📦 Installation

```bash
go get github.com/rosscartlidge/streamv3
```

Requires Go 1.23+ for iterator support.

## 🎯 Use Cases

- **Data Analysis** - Process CSV/JSON files with functional operations
- **System Monitoring** - Parse command output and create dashboards
- **Business Intelligence** - Generate interactive charts from sales data
- **Log Processing** - Stream and analyze log files efficiently
- **ETL Pipelines** - Transform data between formats with type safety

## 🏗️ Architecture

StreamV3 is built on three core abstractions:

- **`Stream[T]`** - Lazy sequence of typed values using Go 1.23 iterators
- **`Record`** - Flexible map-based data structure for heterogeneous data
- **`Filter[T,U]`** - Composable transformations with full type safety

The library emphasizes functional composition while providing modern Go idioms and comprehensive visualization capabilities.

## 📚 Examples

Check out the `doc/examples/` directory for comprehensive usage demonstrations, including the interactive chart showcase.

## 🤝 Contributing

StreamV3 is ready for production use. Issues and contributions are welcome!

## 📄 License

MIT License - see LICENSE file for details.
