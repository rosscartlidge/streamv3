# StreamV3 Monthly Sales Analysis Demo

## Overview

This is a complete, self-contained Go program that demonstrates StreamV3's data processing and visualization capabilities.

## What It Does

1. **Generates Sample Data**: Creates a CSV file (`/tmp/sales.csv`) with 60 sales records across 12 months
2. **Reads CSV Data**: Uses `streamv3.ReadCSV()` with automatic type parsing (strings → int64/float64)
3. **Groups by Month**: Uses `streamv3.GroupByFields()` to group sales by month
4. **Aggregates Revenue**: Uses `streamv3.Aggregate()` with multiple aggregation functions:
   - `Sum("revenue")` - Total revenue per month
   - `Count()` - Number of sales per month
   - `Avg("revenue")` - Average revenue per sale per month
5. **Creates Interactive Chart**: Generates an HTML bar chart at `/tmp/monthly_sales_detailed.html`

## Running the Program

### From Source
```bash
go run test_case_5_detailed_agent.go
```

### Build and Run
```bash
go build -o monthly_sales test_case_5_detailed_agent.go
./monthly_sales
```

## Output

### Console Output
Displays a formatted table of monthly sales statistics:
```
Monthly Sales Summary:
Month          | Total Revenue | Number of Sales | Average Revenue
---------------|---------------|-----------------|------------------
January        | $     6532.00 |               5 | $        1306.40
February       | $     7541.75 |               5 | $        1508.35
...
```

### Generated Files
- `/tmp/sales.csv` - Sample sales data with 60 records
- `/tmp/monthly_sales_detailed.html` - Interactive Chart.js visualization

## Interactive Chart Features

The generated HTML chart includes:
- **Interactive Field Selection**: Choose X-axis (month) and Y-axis fields (revenue metrics)
- **Chart Type Selection**: Switch between bar, line, scatter, pie, and other chart types
- **Zoom and Pan**: Explore data ranges interactively
- **Statistical Overlays**: Optional trend lines and moving averages
- **Export Options**: Download chart as PNG or data as CSV
- **Responsive Design**: Bootstrap 5 UI with light/dark theme support

## Dependencies

- StreamV3 library (`github.com/rosscartlidge/ssql`)
- Go 1.23+ (for iter.Seq support)

## Key StreamV3 Concepts Demonstrated

### CSV Reading with Auto-Parsing
```go
salesData, err := streamv3.ReadCSV(csvPath)
// Automatically parses: numbers → int64/float64, booleans → bool
```

### Canonical Type Usage
```go
// Always use int64 and float64 with GetOr for CSV data
totalRevenue := streamv3.GetOr(record, "total_revenue", float64(0))
numSales := streamv3.GetOr(record, "num_sales", int64(0))
```

### Functional Composition with Pipe
```go
// GroupByFields creates records with a sequence field
// Aggregate applies aggregation functions to the sequence
monthlySummary := streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
    "total_revenue": streamv3.Sum("revenue"),
    "num_sales":     streamv3.Count(),
})(streamv3.GroupByFields("sales", "month")(salesData))
```

### Interactive Visualization
```go
config := streamv3.DefaultChartConfig()
config.Title = "Monthly Sales Revenue Analysis"
config.ChartType = "bar"
streamv3.InteractiveChart(chartData, chartPath, config)
```

## Sample Data Structure

The generated CSV contains:
- **month**: January through December (string)
- **product**: Widget A, B, C, or D (string)
- **revenue**: Sales amount in dollars (float64)
- **quantity**: Number of units sold (int64)

## Code Architecture

The program is organized into clear sections:
1. **Main function**: Orchestrates the complete workflow
2. **Data generation**: Creates realistic sample sales data
3. **Data processing**: Groups and aggregates using StreamV3
4. **Visualization**: Creates interactive chart with custom configuration
5. **Error handling**: Comprehensive error checking throughout

## Learning Resources

- StreamV3 Documentation: See CLAUDE.md in the repository root
- Interactive Chart Demo: `doc/examples/chart_demo.go`
- SQL-Style Operations: Check `sql.go` for GroupBy and Aggregate details
- Chart Configuration: See `chart.go` for all ChartConfig options
