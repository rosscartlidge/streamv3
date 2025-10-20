# StreamV3 Getting Started Guide

*A step-by-step introduction to modern Go stream processing with interactive visualizations*

## Table of Contents

### Documentation Navigation
- **[API Reference](api-reference.md)** - Complete function reference and examples
- **[Advanced Tutorial](advanced-tutorial.md)** - Complex patterns and optimization techniques

### Learning Path
- [Quick Demo - See the Power](#quick-demo---see-the-power)
- [What is StreamV3?](#what-is-streamv3)
- [Your First Stream](#your-first-stream)
- [Working with Records](#working-with-records)
- [Reading Real Data](#reading-real-data)
- [Command Output Processing](#command-output-processing)
- [Functional Composition](#functional-composition)
- [Interactive Charts Made Easy](#interactive-charts-made-easy)
- [Error Handling](#error-handling)
- [What's Next?](#whats-next)
- [Try It Yourself](#try-it-yourself)

---

## Quick Demo - See the Power

Let's start with something exciting. Create a file called `demo.go`:

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Generate some sample sales data
    months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}

    var data []streamv3.Record
    for _, month := range months {
        record := streamv3.MakeMutableRecord().
            String("month", month).
            Float("revenue", 50000+rand.Float64()*100000).
            Int("deals", int64(20+rand.Intn(30))).
            Freeze()
        data = append(data, record)
    }

    // Create an interactive chart with minimal code!
    config := streamv3.DefaultChartConfig()
    config.Title = "Monthly Sales Dashboard"

    err := streamv3.InteractiveChart(
        streamv3.From(data),
        "sales_chart.html",
        config,
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("📊 Interactive chart created: sales_chart.html")
    fmt.Println("Open it in your browser and try:")
    fmt.Println("  • Switching between chart types")
    fmt.Println("  • Changing X/Y fields")
    fmt.Println("  • Hovering for details")
    fmt.Println("  • Exporting as PNG")
}
```

Run it:
```bash
go run demo.go
```

Open `sales_chart.html` in your browser. You just created an interactive, responsive chart with field selection, multiple chart types, and export capabilities!

> 💡 **Next Step**: Check the [Chart & Visualization](api-reference.md#chart--visualization) section in the API reference for more chart options.

---

## What is StreamV3?

StreamV3 brings the elegance of functional programming to Go data processing, built on Go 1.23's new iterators. It lets you:

- **Process data** with clean, composable operations
- **Visualize results** as interactive charts instantly
- **Handle any format** - CSV, JSON, command output
- **Chain operations** naturally with functional composition
- **Scale efficiently** with lazy evaluation and iterators

Think of it as the Unix pipeline philosophy applied to structured data, with built-in visualization superpowers.

### Core Concepts

1. **Streams**: Lazy sequences of data that can be processed efficiently
2. **Records**: Flexible map-based structures for heterogeneous data
3. **Functional Operations**: Composable functions like `Select`, `Where`, `GroupBy`
4. **Interactive Visualization**: Built-in charting with zero configuration

> 📚 **Deep Dive**: See [Core Types](api-reference.md#core-types) for detailed type information.

---

## Your First Stream

Let's process some simple data. Create `first_stream.go`:

```go
package main

import (
    "fmt"
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Start with some numbers
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    // Process them with functional composition
    evens := streamv3.Where(func(x int) bool {
        return x%2 == 0
    })(slices.Values(numbers))

    squared := streamv3.Select(func(x int) int {
        return x * x
    })(evens)

    limited := streamv3.Limit[int](3)(squared)

    var result []int
    for num := range limited {
        result = append(result, num)
    }

    fmt.Printf("Even squares (first 3): %v\n", result)
    // Output: [4, 16, 36]
}
```

This demonstrates the functional composition approach. Each operation returns a new iterator that can be chained together.

> 📚 **Learn More**: Explore [Transform Operations](api-reference.md#transform-operations) and [Filter Operations](api-reference.md#filter-operations) in the API reference.

---

## Working with Records

Real data has multiple fields. StreamV3 uses `Record` for structured data:

```go
package main

import (
    "fmt"
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Create structured data
    people := []streamv3.Record{
        streamv3.MakeMutableRecord().String("name", "Alice").Int("age", int64(30)).Float("score", 95.5).Freeze(),
        streamv3.MakeMutableRecord().String("name", "Bob").Int("age", int64(25)).Float("score", 87.2).Freeze(),
        streamv3.MakeMutableRecord().String("name", "Carol").Int("age", int64(35)).Float("score", 92.1).Freeze(),
    }

    // Find high scorers using type-safe helpers
    highScorers := streamv3.Where(func(r streamv3.Record) bool {
        score := streamv3.GetOr(r, "score", 0.0)
        return score > 90
    })(slices.Values(people))

    fmt.Println("High scorers:")
    for person := range highScorers {
        name := streamv3.GetOr(person, "name", "Unknown")
        score := streamv3.GetOr(person, "score", 0.0)
        fmt.Printf("  %s: %.1f\n", name, score)
    }
}
```

### Record Builder Pattern

The `MakeMutableRecord()` builder provides a fluent interface for creating structured data:

```go
record := streamv3.MakeMutableRecord().
    String("name", "Alice").
    Int("age", int64(30)).
    Float("salary", 75000.0).
    Bool("active", true).
    Freeze()
```

> 📚 **Reference**: See [Helper Functions](api-reference.md#helper-functions) for Record access utilities.

---

## Reading Real Data

Let's work with CSV files. Create `people.csv`:

```csv
name,age,department,salary
Alice Johnson,32,Engineering,75000
Bob Smith,28,Marketing,55000
Carol Davis,35,Engineering,82000
David Wilson,29,Sales,48000
Eve Brown,31,Engineering,79000
```

Now process it:

```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read CSV data - returns error if file cannot be opened
    data, err := streamv3.ReadCSV("people.csv")
    if err != nil {
        log.Fatalf("Failed to read CSV: %v", err)
    }

    // Find well-paid engineers
    engineers := streamv3.Where(func(r streamv3.Record) bool {
        dept := streamv3.GetOr(r, "department", "")
        salary := streamv3.GetOr(r, "salary", 0.0)
        return dept == "Engineering" && salary > 70000
    })(data)

    fmt.Println("Well-paid engineers:")
    for person := range engineers {
        name := streamv3.GetOr(person, "name", "Unknown")
        salary := streamv3.GetOr(person, "salary", 0.0)
        fmt.Printf("  %s: $%.0f\n", name, salary)
    }
}
```

### Working with Different Formats

StreamV3 supports multiple data formats:

```go
// Read JSON (returns iterator and error)
jsonData, err := streamv3.ReadJSON("data.jsonl")
if err != nil {
    log.Fatalf("Failed to read JSON: %v", err)
}

// Read from any io.Reader (great for HTTP responses)
csvStream := streamv3.ReadCSVFromReader(httpResponse.Body)

// Write results
err = streamv3.WriteJSON(processedData, "output.json")
if err != nil {
    log.Fatalf("Failed to write JSON: %v", err)
}
```

> 📚 **Reference**: See [I/O Operations](api-reference.md#io-operations) for all supported formats.

---

## Command Output Processing

StreamV3 works great with command-line tools. Let's analyze process information:

```go
package main

import (
    "fmt"
    "os/exec"
    "slices"
    "strings"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Get process information (Linux/macOS)
    cmd := exec.Command("ps", "-eo", "pid,ppid,pcpu,pmem,comm")
    output, err := cmd.Output()
    if err != nil {
        fmt.Println("This example requires Unix-like system")
        return
    }

    // Parse lines into records
    lines := strings.Split(string(output), "\n")[1:] // Skip header

    var processes []streamv3.Record
    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        fields := strings.Fields(line)
        if len(fields) >= 5 {
            record := streamv3.MakeMutableRecord().
                String("pid", fields[0]).
                String("ppid", fields[1]).
                String("cpu", fields[2]).
                String("mem", fields[3]).
                String("command", strings.Join(fields[4:], " ")).
                Freeze()
            processes = append(processes, record)
        }
    }

    // Find memory-heavy processes
    heavyProcesses := streamv3.Where(func(r streamv3.Record) bool {
        // This is a simplified example - real parsing would convert strings to numbers
        return strings.Contains(streamv3.GetOr(r, "command", ""), "chrome") ||
               strings.Contains(streamv3.GetOr(r, "command", ""), "firefox")
    })(slices.Values(processes))

    fmt.Println("Browser processes:")
    for proc := range heavyProcesses {
        pid := streamv3.GetOr(proc, "pid", "")
        cmd := streamv3.GetOr(proc, "command", "")
        fmt.Printf("  PID %s: %s\n", pid, cmd)
    }
}
```

> 💡 **Pro Tip**: The [Advanced Tutorial](advanced-tutorial.md) shows how to build robust log processing pipelines.

---

## Functional Composition

StreamV3 operations compose beautifully. Here are two equivalent approaches:

### Chain Approach
```go
// Multiple filters of the same type can be chained
pipeline := streamv3.Chain(
    streamv3.Where(func(r streamv3.Record) bool {
        return streamv3.GetOr(r, "active", false)
    }),
    streamv3.Where(func(r streamv3.Record) bool {
        salary := streamv3.GetOr(r, "salary", 0.0)
        return salary > 50000
    }),
)
result := pipeline(data)
```

### Step-by-Step Composition
```go
// Each step is a pure function
activeUsers := streamv3.Where(func(r streamv3.Record) bool {
    return streamv3.GetOr(r, "active", false)
})(data)

wellPaid := streamv3.Where(func(r streamv3.Record) bool {
    salary := streamv3.GetOr(r, "salary", 0.0)
    return salary > 50000
})(activeUsers)
```

### Advanced Composition with Different Types
```go
// When operations change types, compose step by step
numbers := slices.Values([]int{1, 2, 3, 4, 5})

doubled := streamv3.Select(func(x int) int {
    return x * 2
})(numbers)

windows := streamv3.CountWindow[int](3)(doubled)

var results [][]int
for window := range windows {
    results = append(results, window)
}
// Results: [[2, 4, 6], [8, 10]]
```

> 📚 **Learn More**: See [Advanced Tutorial](advanced-tutorial.md) for complex composition patterns.

---

## Interactive Charts Made Easy

One of StreamV3's superpowers is instant visualization. Let's create a sales dashboard:

```go
package main

import (
    "fmt"
    "math/rand"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Generate sample sales data
    regions := []string{"North", "South", "East", "West"}
    products := []string{"Laptop", "Phone", "Tablet"}

    var salesData []streamv3.Record
    for _, region := range regions {
        for _, product := range products {
            record := streamv3.MakeMutableRecord().
                String("region", region).
                String("product", product).
                Float("sales", 10000 + rand.Float64()*50000).
                Int("units", int64(50 + rand.Intn(200))).
                Freeze()
            salesData = append(salesData, record)
        }
    }

    // Create different chart types
    config1 := streamv3.DefaultChartConfig()
    config1.Title = "Sales by Region and Product"
    config1.ChartType = "bar"
    config1.Width = 1200
    config1.Height = 600

    err := streamv3.InteractiveChart(
        streamv3.From(salesData),
        "sales_dashboard.html",
        config1,
    )
    if err != nil {
        panic(err)
    }

    // Time series example (if you have time data)
    config2 := streamv3.DefaultChartConfig()
    config2.Title = "Sales Trends"
    config2.Theme = "dark"

    // For time series, you would use:
    // streamv3.TimeSeriesChart(data, "date", []string{"sales", "units"}, "trends.html", config2)

    fmt.Println("📊 Charts created:")
    fmt.Println("  • sales_dashboard.html - Interactive bar chart")
    fmt.Println("\nFeatures to try:")
    fmt.Println("  • Switch between bar, line, scatter, pie charts")
    fmt.Println("  • Change X/Y axis fields")
    fmt.Println("  • Filter data interactively")
    fmt.Println("  • Export as PNG")
}
```

### Chart Configuration Options

```go
config := streamv3.DefaultChartConfig()
config.Title = "My Dashboard"
config.ChartType = "line"        // "bar", "line", "scatter", "pie"
config.Theme = "dark"            // "light", "dark"
config.Width = 1400
config.Height = 700
config.ShowLegend = true
config.EnableZoom = true
config.ColorScheme = "vibrant"   // "vibrant", "pastel", "monochrome"
```

> 📚 **Reference**: See [Chart & Visualization](api-reference.md#chart--visualization) for all chart options.

---

## Error Handling

StreamV3 provides both safe and unsafe versions of operations:

### Unsafe (Fast, Fail-Fast)
```go
// Panics on error - good for development and trusted data
result := streamv3.Select(func(x string) int {
    // This might panic if x is not a valid number
    return mustParseInt(x)
})(data)
```

### Safe (Error Handling)
```go
// Returns errors - good for production and untrusted data
safeResult := streamv3.SelectSafe(func(x string) (int, error) {
    return strconv.Atoi(x)
})(dataWithErrors)

for value, err := range safeResult {
    if err != nil {
        log.Printf("Error processing value: %v", err)
        continue
    }
    // Process valid value
    fmt.Printf("Parsed: %d\n", value)
}
```

### I/O Error Handling
```go
// ReadCSV returns error - always check it
data, err := streamv3.ReadCSV("data.csv")
if err != nil {
    log.Fatalf("Failed to read CSV: %v", err)
}

// WriteJSON returns error - handle it
err = streamv3.WriteJSON(processedData, "output.json")
if err != nil {
    log.Fatalf("Failed to write JSON: %v", err)
}
```

> 📚 **Best Practices**: The [Advanced Tutorial](advanced-tutorial.md) covers error handling patterns in detail.

---

## What's Next?

You've learned the fundamentals! Here's your learning path:

### Immediate Next Steps
1. **Try the examples** in this guide with your own data
2. **Explore the [API Reference](api-reference.md)** for all available functions
3. **Read the [Advanced Tutorial](advanced-tutorial.md)** for production patterns

### Advanced Topics to Explore

#### SQL-Style Operations
```go
// Join datasets
joined := streamv3.InnerJoin(rightData, streamv3.OnFields("user_id"))(leftData)

// Group and aggregate
grouped := streamv3.GroupByFields("sales_data", "region")(salesData)
results := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
    "total_sales": streamv3.Sum("amount"),
    "avg_sale":    streamv3.Avg("amount"),
    "count":       streamv3.Count(),
})(grouped)
```

#### Window Operations for Time Series
```go
// Fixed-size windows
batches := streamv3.CountWindow[streamv3.Record](10)(data)

// Time-based windows
timeWindows := streamv3.TimeWindow[streamv3.Record](
    5*time.Minute,
    "timestamp",
)(data)
```

#### Real-Time Stream Processing
```go
// Process infinite streams with early termination
limited := streamv3.Limit[streamv3.Record](1000)(infiniteStream)
timed := streamv3.Timeout[streamv3.Record](30*time.Second)(sensorData)
```

### Production Considerations
- **Performance optimization** with lazy evaluation
- **Memory management** for large datasets
- **Error recovery** strategies
- **Monitoring and observability**

> 📚 **Next Read**: Jump to the [Advanced Tutorial](advanced-tutorial.md) for production-ready patterns and performance optimization.

---

## Try It Yourself

### Exercise 1: Data Analysis Pipeline
Create a CSV file with some data and build a complete analysis pipeline:

1. Read CSV data
2. Filter and transform records
3. Group by categories
4. Calculate aggregations
5. Create an interactive chart

### Exercise 2: Log Processing
Process log files or command output:

1. Parse text into structured records
2. Filter by log level or patterns
3. Count occurrences by category
4. Visualize trends over time

### Exercise 3: API Data Processing
Fetch data from a REST API and visualize it:

1. Fetch JSON from HTTP endpoint
2. Transform and clean the data
3. Combine with local data sources
4. Create dashboard with multiple charts

### Need Help?

- **[API Reference](api-reference.md)** - Complete function documentation
- **[Advanced Tutorial](advanced-tutorial.md)** - Complex patterns and optimization
- **Examples** - Check the `examples/` directory for working code

---

*Ready to build something amazing? Head to the [Advanced Tutorial](advanced-tutorial.md) for production-ready patterns!*