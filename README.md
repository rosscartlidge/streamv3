# StreamV3 🚀

**Modern Go stream processing made simple** - Transform data with intuitive operations, create interactive visualizations, and even generate code from natural language descriptions.

Built on Go 1.23+ with first-class support for iterators, generics, and functional composition.

## ✨ What Makes StreamV3 Special

### 🎯 **Simple Yet Powerful**

**Go Library:**
```go
// Read data, filter, group, and visualize - all type-safe
sales, err := streamv3.ReadCSV("sales.csv")
if err != nil {
    log.Fatal(err)
}

topRegions := streamv3.Limit[streamv3.Record](5)(
    streamv3.SortBy(func(r streamv3.Record) float64 {
        return -streamv3.GetOr(r, "total_revenue", 0.0) // Descending
    })(streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
        "total_revenue": streamv3.Sum("amount"),
    })(streamv3.GroupByFields("sales", "region")(sales)))
)

streamv3.QuickChart(topRegions, "region", "total_revenue", "top_regions.html")
```

**Or use the CLI:**
```bash
# Prototype with Unix-style pipelines, then generate production Go code
streamv3 exec -- ps -efl | \
  streamv3 group-by -by UID -function count -result process_count | \
  streamv3 chart -x UID -y process_count -output chart.html

# Debug pipelines with jq (JSONL streaming format)
streamv3 read-csv data.csv | jq '.' | head -5  # Inspect data
streamv3 read-csv data.csv | streamv3 where -match age gt 30 | jq -s 'length'  # Count results
```

[**Try the CLI →**](doc/cli/codelab-cli.md) | [**Debug with jq →**](doc/cli/debugging_pipelines.md)

### 🤖 **AI-Powered Code Generation**
Describe what you want in plain English, get working StreamV3 code:

> *"Read customer data, find high-value customers, group by region, create a chart"*

→ **Generates clean, readable Go code automatically**

[**Try the AI Assistant →**](doc/ai-human-guide.md)

### 📊 **Interactive Visualizations**
Create modern, responsive charts with zoom, pan, and filtering capabilities:

```go
streamv3.QuickChart(data, "month", "revenue", "chart.html")  // One line = full dashboard
```

[**See Chart Demo →**](examples/chart_demo.go)

## 🚀 Quick Start

### Prerequisites
- **Go 1.23+** required for iterator support

**Don't have Go installed?**
- macOS: `brew install go`
- Linux/Windows: [Download from go.dev](https://go.dev/dl/)
- Verify: `go version` (should show 1.23+)

### Installation

#### Option 1: CLI Tool (for rapid prototyping)

```bash
# Install the command-line tool
go install github.com/rosscartlidge/streamv3/cmd/streamv3@latest

# Verify installation
streamv3 -version

# Try it out
echo "name,age,salary
Alice,30,95000
Bob,25,65000" | streamv3 read-csv | streamv3 where -match age gt 28
```

[**See CLI Tutorial →**](doc/cli/codelab-cli.md)

#### Option 2: Go Library (for application development)

**Step 1: Create a new project**
```bash
mkdir my-project
cd my-project
go mod init myproject  # Initialize Go module (required!)
```

**Step 2: Install StreamV3**
```bash
go get github.com/rosscartlidge/streamv3
```

### Hello StreamV3
```go
package main

import (
    "fmt"
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    numbers := slices.Values([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

    evenNumbers := streamv3.Where(func(x int) bool {
        return x%2 == 0
    })(numbers)

    first3 := streamv3.Limit[int](3)(evenNumbers)

    fmt.Println("First 3 even numbers:")
    for num := range first3 {
        fmt.Println(num) // 2, 4, 6
    }
}
```

### Your First Chart
```go
package main

import (
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Create sample data
    monthlyRevenue := []streamv3.Record{
        streamv3.MakeMutableRecord().String("month", "Jan").Float("revenue", 120000).Freeze(),
        streamv3.MakeMutableRecord().String("month", "Feb").Float("revenue", 135000).Freeze(),
        streamv3.MakeMutableRecord().String("month", "Mar").Float("revenue", 118000).Freeze(),
    }

    data := slices.Values(monthlyRevenue)

    // Generate interactive chart
    streamv3.QuickChart(data, "month", "revenue", "revenue_chart.html")
    // Opens in browser with zoom, pan, and export features
}
```

## 🎓 Learning Path

**New to StreamV3?** We've got you covered with step-by-step guides:

### 1. ⚡ **[CLI Tutorial](doc/cli/codelab-cli.md)** *(In Development)*
*Prototype fast with Unix-style pipelines, generate production code*
- Quick data exploration with command-line tools
- Process system commands (ps, df, etc.)
- Create visualizations with one command
- Generate Go code from CLI pipelines
- **Debug pipelines with jq** - [See debugging guide →](doc/cli/debugging_pipelines.md)
- **Perfect for rapid prototyping!**

### 2. 📚 **[Getting Started Guide](doc/codelab-intro.md)**
*Learn the Go library fundamentals with hands-on examples*
- Basic operations (Select, Where, Limit)
- Working with CSV/JSON data
  - **⚠️ Note**: CSV auto-parses `"25"` → `int64(25)`, use correct types with `GetOr()`
- Creating your first visualizations
- Real-world examples

### 3. 📖 **[API Reference](doc/api-reference.md)**
*Complete function documentation with examples*
- All operations organized by category
- Transform, Filter, Aggregate, Join operations
- Window processing for real-time data
- Chart and visualization options

### 4. 🎯 **[Advanced Tutorial](doc/advanced-tutorial.md)**
*Master complex patterns and production techniques*
- Stream joins and complex aggregations
- Real-time processing with windowing
- Infinite stream handling
- Performance optimization

### 5. 🤖 **[AI Code Generation](doc/ai-human-guide.md)**
*Generate StreamV3 code from natural language*
- Use any AI assistant (Claude, ChatGPT, Gemini)
- Describe what you want, get working code
- Human-readable, verifiable results
- Perfect for rapid prototyping
- **For LLMs**: Copy [ai-code-generation.md](doc/ai-code-generation.md) into your LLM

## 🔧 Core Capabilities

### **SQL-Style Data Processing**
```go
// Group sales by region, calculate totals, get top 5
topRegions := streamv3.Limit[streamv3.Record](5)(
    streamv3.SortBy(keyFunc)(
        streamv3.Aggregate("sales", aggregations)(
            streamv3.GroupByFields("sales", "region")(salesData))))
```

### **Real-Time Stream Processing**
```go
// Process sensor data in 5-minute windows
windowed := streamv3.TimeWindow[streamv3.Record](5*time.Minute, "timestamp")(sensorStream)
for window := range windowed {
    // Analyze each time window
}
```

### **Interactive Dashboards**
```go
config := streamv3.DefaultChartConfig()
config.Title = "Sales Dashboard"
config.ChartType = "line"
streamv3.InteractiveChart(data, "dashboard.html", config)
```

### **Data Integration**
```go
// Join customer and order data
customerOrders := streamv3.InnerJoin(
    orderStream,
    streamv3.OnFields("customer_id")
)(customerStream)
```

## 🎨 Try the Examples

Run these to see StreamV3 in action:

```bash
# Interactive chart showcase
go run examples/chart_demo.go

# Data analysis pipeline
go run examples/functional_example.go

# Real-time processing
go run examples/early_termination_example.go
```

## 🌟 Why Choose StreamV3?

- **🎯 Simple API** - If you know SQL, you know StreamV3
- **🔒 Type Safe** - Go generics catch errors at compile time
- **📊 Visual** - Create charts as easily as processing data
- **🤖 AI Ready** - Generate code from descriptions
- **⚡ Performance** - Lazy evaluation and memory efficiency
- **🔄 Composable** - Build complex pipelines from simple operations
- **🔍 Debuggable** - JSONL streaming works with jq and Unix tools

## 🎯 Perfect For

- **Data Scientists** - Analyze CSV/JSON files with ease
- **DevOps Engineers** - Monitor systems and create dashboards
- **Business Analysts** - Generate reports and visualizations
- **Developers** - Build ETL pipelines and data processing tools
- **Anyone** - Who wants to turn data descriptions into working code

## 🚀 What's Next?

1. **[Install StreamV3](#installation)** and try the quick start
2. **[Try the CLI](doc/cli/codelab-cli.md)** for rapid prototyping *(in development)*
3. **[Follow the Getting Started Guide](doc/codelab-intro.md)** for library fundamentals
4. **[Try the AI Assistant](doc/ai-human-guide.md)** for code generation
5. **[Explore Advanced Patterns](doc/advanced-tutorial.md)** for production use

## 📚 Documentation

- **[Debugging Pipelines](doc/cli/debugging_pipelines.md)** - Debug with jq, inspect data, profile performance
- **[Troubleshooting Guide](doc/cli/troubleshooting.md)** - Common issues and quick solutions
- **[API Reference](doc/api-reference.md)** - Complete function documentation
- **[CLI Tutorial](doc/cli/codelab-cli.md)** - Command-line tool guide
- **[AI Code Generation](doc/ai-human-guide.md)** - Natural language to code
  - **[For LLMs](doc/ai-code-generation.md)** - Copy this prompt into your LLM
  - **[Detailed Examples](doc/ai-code-generation-detailed.md)** - Comprehensive example library

## 🤝 Community

StreamV3 is production-ready and actively maintained. Questions, issues, and contributions are welcome!

- 📖 **Documentation**: Complete guides and API reference
- 🤖 **AI Integration**: Generate code from natural language
- 📊 **Visualization**: Interactive charts and dashboards
- 🔧 **Examples**: Real-world usage patterns
- 🔍 **Debugging**: jq integration for pipeline inspection

---

**Ready to transform how you process data?** [Get started now →](doc/codelab-intro.md)

*StreamV3: Where data processing meets AI-powered development* ✨