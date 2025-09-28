# StreamV3 Introduction Codelab

*A gentle introduction to modern Go stream processing with interactive visualizations*

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
    data := streamv3.From(months).
        Map(func(month string) streamv3.Record {
            return streamv3.NewRecord().
                String("month", month).
                Float("revenue", 50000+rand.Float64()*100000).
                Int("deals", int64(20+rand.Intn(30))).
                Build()
        }).
        Collect()

    // Create an interactive chart with one line of code!
    streamv3.QuickChart(data, "month", "revenue", "sales_chart.html")

    fmt.Println("📊 Interactive chart created: sales_chart.html")
    fmt.Println("Open it in your browser and try:")
    fmt.Println("  • Switching between chart types")
    fmt.Println("  • Changing X/Y fields")
    fmt.Println("  • Hovering for details")
}
```

Run it:
```bash
go run demo.go
```

Open `sales_chart.html` in your browser. You just created an interactive, responsive chart with field selection, multiple chart types, and export capabilities - in about 10 lines of code!

## What is StreamV3?

StreamV3 brings the elegance of functional programming to Go data processing, built on Go 1.23's new iterators. It lets you:

- **Process data** with clean, composable operations
- **Visualize results** as interactive charts instantly
- **Handle any format** - CSV, JSON, command output
- **Chain operations** naturally with method chaining or functional composition

Think of it as the Unix pipeline philosophy applied to structured data, with built-in visualization superpowers.

## Your First Stream

Let's process some simple data. Create `first_stream.go`:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Start with some numbers
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    // Process them with a fluent chain
    result := streamv3.From(numbers).
        Where(func(x int) bool { return x%2 == 0 }). // Keep even numbers
        Map(func(x int) int { return x * x }).        // Square them
        Limit(3).                                     // Take first 3
        Collect()                                     // Get results

    fmt.Printf("Even squares (first 3): %v\n", result)
    // Output: [4, 16, 36]
}
```

This reads like English: "From numbers, where even, map to squares, limit to 3, collect results."

## Working with Records

Real data has multiple fields. StreamV3 uses `Record` for structured data:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Create structured data
    people := []streamv3.Record{
        streamv3.NewRecord().String("name", "Alice").Int("age", 30).Float("score", 95.5).Build(),
        streamv3.NewRecord().String("name", "Bob").Int("age", 25).Float("score", 87.2).Build(),
        streamv3.NewRecord().String("name", "Carol").Int("age", 35).Float("score", 92.1).Build(),
    }

    // Find high scorers
    highScorers := streamv3.From(people).
        Where(func(r streamv3.Record) bool {
            return r["score"].(float64) > 90
        }).
        Collect()

    fmt.Printf("High scorers (%d):\n", len(highScorers))
    for _, person := range highScorers {
        fmt.Printf("  %s: %.1f\n", person["name"], person["score"])
    }
}
```

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
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read CSV data
    data := streamv3.ReadCSV("people.csv")

    // Find average salary by department
    avgSalaries := streamv3.GroupRecordsByFields(data, "department")
    results := streamv3.AggregateGroups(avgSalaries, map[string]streamv3.AggregateFunc{
        "avg_salary": streamv3.Avg("salary"),
        "count":      streamv3.Count(),
    })

    fmt.Println("Department Averages:")
    for result := range results {
        fmt.Printf("  %s: $%.0f (%d people)\n",
            result["department"],
            result["avg_salary"],
            result["count"])
    }

    // Visualize it!
    streamv3.QuickChart(streamv3.Collect(results), "department", "avg_salary", "salaries.html")
    fmt.Println("\n📊 Chart created: salaries.html")
}
```

## Command Output Processing

StreamV3 can parse command output automatically:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Parse 'ps' command output (Linux/Mac)
    processes := streamv3.ExecCommand("ps", "aux")

    // Find top memory users
    topMemory := processes.
        SortByKey(func(r streamv3.Record) float64 {
            // MEM field is automatically parsed as float
            return r["MEM"].(float64)
        }, false). // false = descending
        Limit(5).
        Collect()

    fmt.Println("Top 5 Memory Users:")
    for _, proc := range topMemory {
        fmt.Printf("  %s: %.1f%% (%s)\n",
            proc["USER"],
            proc["MEM"],
            proc["COMMAND"])
    }
}
```

## Two Ways to Compose

StreamV3 offers two styles - pick what feels natural:

### Fluent Style (Method Chaining)
```go
result := streamv3.From(data).
    Where(predicate).
    Map(transform).
    Limit(10).
    Collect()
```

### Functional Style (Explicit Composition)
```go
pipeline := streamv3.Pipe(
    streamv3.Where(predicate),
    streamv3.Map(transform),
    streamv3.Limit(10),
)
result := streamv3.Collect(pipeline(streamv3.From(data)))
```

Both compile to the same efficient code using Go 1.23 iterators.

## Interactive Charts Made Easy

StreamV3's visualization is designed for exploration:

```go
// Quick chart - one line
streamv3.QuickChart(data, "x_field", "y_field", "chart.html")

// Customized chart
config := streamv3.DefaultChartConfig()
config.Title = "My Analysis"
config.ChartType = "scatter"
config.EnableCalculations = true  // Add trend lines, moving averages

streamv3.InteractiveChart(data, "detailed_chart.html", config)

// Time series with multiple metrics
streamv3.TimeSeriesChart(data, "timestamp", []string{"cpu", "memory"}, "metrics.html", config)
```

Every chart includes:
- Interactive field selection (change X/Y axes)
- Multiple chart types (line, bar, scatter, pie)
- Zoom and pan
- Statistical overlays (trend lines, moving averages)
- Export capabilities (PNG, CSV)

## Error Handling

StreamV3 supports both simple and error-aware processing:

```go
// Simple - panics on errors
data := streamv3.ReadCSV("data.csv").Collect()

// Error-aware
dataWithErrors := streamv3.ReadCSVSafe("data.csv")
for item, err := range dataWithErrors {
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        continue
    }
    // Process item...
}
```

## What's Next?

You now know the basics! StreamV3 can:

- Process any tabular data format
- Handle infinite streams and real-time data
- Perform complex aggregations and joins
- Create publication-ready visualizations
- Scale from simple scripts to production systems

Ready to dive deeper? Check out:

- **Advanced Tutorial** - Complex aggregations, joins, and real-time processing
- **API Reference** - Complete function documentation
- **Chart Gallery** - All visualization options with examples

## Try It Yourself

Install StreamV3 and run these examples:

```bash
go mod init my-stream-project
go get github.com/rosscartlidge/streamv3
```

Start with the first example and experiment. The best way to learn is by doing!

**Happy streaming!** 🚀