# StreamV3 AI Code Generation - Comprehensive Guide

*Complete reference for LLMs with large context windows - includes full examples and patterns*

---

## System Prompt

```
You are an expert Go developer specializing in StreamV3, a modern Go stream processing library built on Go 1.23+ iterators. Generate high-quality, idiomatic StreamV3 code from natural language descriptions.

## 🎯 PRIMARY GOAL: Human-Readable, Verifiable Code

Write code that humans can quickly read and verify - no clever tricks. Always prioritize clarity over cleverness.
```

---

## API Documentation Source

**For complete, always-current API documentation:**

```bash
# View all exported functions
go doc github.com/rosscartlidge/streamv3

# View specific function documentation
go doc github.com/rosscartlidge/streamv3.Select
go doc github.com/rosscartlidge/streamv3.Where
go doc github.com/rosscartlidge/streamv3.GroupByFields
```

The godoc is generated directly from the source code and is always in sync with the actual implementation. When in doubt, consult `go doc`.

---

## Complete Examples Library

### Example 1: Basic Filtering and Aggregation

**Natural Language**: "Read employee data from employees.csv, filter for employees with salary over $80,000, and count how many employees there are by department"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read employee data - always handle errors
    employees, err := streamv3.ReadCSV("employees.csv")
    if err != nil {
        log.Fatalf("Failed to read CSV: %v", err)
    }

    // Filter for high salary employees
    highSalaryEmployees := streamv3.Where(func(r streamv3.Record) bool {
        // CSV auto-parses "80000" → int64(80000)
        salary := streamv3.GetOr(r, "salary", 0.0)
        return salary > 80000
    })(employees)

    // Group by department
    grouped := streamv3.GroupByFields("dept_analysis", "department")(highSalaryEmployees)

    // Count employees per department
    results := streamv3.Aggregate("dept_analysis", map[string]streamv3.AggregateFunc{
        "employee_count": streamv3.Count(),
    })(grouped)

    // Display results
    fmt.Println("High-salary employees by department:")
    for result := range results {
        dept := streamv3.GetOr(result, "department", "")
        count := streamv3.GetOr(result, "employee_count", int64(0))
        fmt.Printf("%s: %d employees\n", dept, count)
    }
}
```

### Example 2: Top N Analysis

**Natural Language**: "Find the top 5 products by revenue from sales data, showing product name and total revenue"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read sales data
    sales, err := streamv3.ReadCSV("sales.csv")
    if err != nil {
        log.Fatalf("Failed to read CSV: %v", err)
    }

    // Group by product
    grouped := streamv3.GroupByFields("product_analysis", "product_name")(sales)

    // Calculate total revenue per product
    productRevenue := streamv3.Aggregate("product_analysis", map[string]streamv3.AggregateFunc{
        "total_revenue": streamv3.Sum("revenue"),
    })(grouped)

    // Sort by revenue (descending) and take top 5
    top5 := streamv3.Limit[streamv3.Record](5)(
        streamv3.SortBy(func(r streamv3.Record) float64 {
            return -streamv3.GetOr(r, "total_revenue", 0.0) // Negative for descending
        })(productRevenue),
    )

    fmt.Println("Top 5 products by revenue:")
    rank := 1
    for product := range top5 {
        name := streamv3.GetOr(product, "product_name", "")
        revenue := streamv3.GetOr(product, "total_revenue", 0.0)
        fmt.Printf("%d. %s: $%.2f\n", rank, name, revenue)
        rank++
    }
}
```

### Example 3: Data Enrichment

**Natural Language**: "Read customer data, add a customer_tier field based on total purchases (Bronze < $1000, Silver $1000-$5000, Gold > $5000)"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read customer data
    customers, err := streamv3.ReadCSV("customers.csv")
    if err != nil {
        log.Fatalf("Failed to read CSV: %v", err)
    }

    // Add customer tier based on total purchases
    enrichedCustomers := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        totalPurchases := streamv3.GetOr(r, "total_purchases", 0.0)

        var tier string
        switch {
        case totalPurchases > 5000:
            tier = "Gold"
        case totalPurchases >= 1000:
            tier = "Silver"
        default:
            tier = "Bronze"
        }

        return streamv3.SetImmutable(r, "customer_tier", tier)
    })(customers)

    // Group by tier and count customers
    grouped := streamv3.GroupByFields("tier_analysis", "customer_tier")(enrichedCustomers)

    tierCounts := streamv3.Aggregate("tier_analysis", map[string]streamv3.AggregateFunc{
        "customer_count": streamv3.Count(),
        "avg_purchases":  streamv3.Avg("total_purchases"),
    })(grouped)

    fmt.Println("Customer tier distribution:")
    for result := range tierCounts {
        tier := streamv3.GetOr(result, "customer_tier", "")
        count := streamv3.GetOr(result, "customer_count", int64(0))
        avgPurchases := streamv3.GetOr(result, "avg_purchases", 0.0)
        fmt.Printf("%s: %d customers (avg: $%.2f)\n", tier, count, avgPurchases)
    }
}
```

### Example 4: Data Cleaning and Validation

**Natural Language**: "Clean product data by removing items with missing names or negative prices, and standardize category names to title case"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "log"
    "strings"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read product data
    products, err := streamv3.ReadCSV("products.csv")
    if err != nil {
        log.Fatalf("Failed to read CSV: %v", err)
    }

    // Filter out invalid products
    validProducts := streamv3.Where(func(r streamv3.Record) bool {
        name := streamv3.GetOr(r, "name", "")
        price := streamv3.GetOr(r, "price", -1.0)

        // Reject if name is empty or price is negative
        return strings.TrimSpace(name) != "" && price >= 0
    })(products)

    // Standardize category names to title case
    cleanedProducts := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        category := streamv3.GetOr(r, "category", "")
        standardizedCategory := strings.Title(strings.ToLower(strings.TrimSpace(category)))
        return streamv3.SetImmutable(r, "category", standardizedCategory)
    })(validProducts)

    // Count valid products by category
    grouped := streamv3.GroupByFields("category_analysis", "category")(cleanedProducts)

    categoryStats := streamv3.Aggregate("category_analysis", map[string]streamv3.AggregateFunc{
        "product_count": streamv3.Count(),
        "avg_price":     streamv3.Avg("price"),
        "min_price":     streamv3.Min[float64]("price"),
        "max_price":     streamv3.Max[float64]("price"),
    })(grouped)

    fmt.Println("Cleaned product data by category:")
    for result := range categoryStats {
        category := streamv3.GetOr(result, "category", "")
        count := streamv3.GetOr(result, "product_count", int64(0))
        avgPrice := streamv3.GetOr(result, "avg_price", 0.0)
        minPrice := streamv3.GetOr(result, "min_price", 0.0)
        maxPrice := streamv3.GetOr(result, "max_price", 0.0)

        fmt.Printf("%s: %d products, price range $%.2f-$%.2f (avg: $%.2f)\n",
            category, count, minPrice, maxPrice, avgPrice)
    }
}
```

### Example 5: Real-Time Monitoring

**Natural Language**: "Monitor server logs in real-time, process in 1-minute windows, and alert if error rate exceeds 5%"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "strings"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Simulate real-time log stream
    logStream := func(yield func(streamv3.Record) bool) {
        statuses := []string{"200", "200", "200", "200", "404", "500", "200"}

        for i := 0; ; i++ {
            status := statuses[i%len(statuses)]

            record := streamv3.MakeMutableRecord().
                String("status", status).
                String("endpoint", fmt.Sprintf("/api/endpoint%d", i%3)).
                Time("timestamp", time.Now()).
                Freeze()

            if !yield(record) {
                return
            }
            time.Sleep(100 * time.Millisecond) // Simulate real-time
        }
    }

    // Process in 1-minute windows
    windowed := streamv3.TimeWindow[streamv3.Record](
        1*time.Minute,
        "timestamp",
    )(streamv3.Limit[streamv3.Record](100)(logStream)) // Limit for demo

    fmt.Println("Real-time Error Rate Monitoring:")
    for window := range windowed {
        if len(window) == 0 {
            continue
        }

        // Calculate error rate for this window
        totalRequests := len(window)
        errorCount := 0

        for _, record := range window {
            status := streamv3.GetOr(record, "status", "")
            if strings.HasPrefix(status, "4") || strings.HasPrefix(status, "5") {
                errorCount++
            }
        }

        errorRate := float64(errorCount) / float64(totalRequests) * 100

        // Alert if error rate exceeds 5%
        if errorRate > 5 {
            fmt.Printf("🚨 ALERT: Error rate %.1f%% (%d errors in %d requests)\n",
                errorRate, errorCount, totalRequests)
        } else {
            fmt.Printf("✅ Normal: Error rate %.1f%% (%d errors in %d requests)\n",
                errorRate, errorCount, totalRequests)
        }
    }
}
```

### Example 6: Join Analysis

**Natural Language**: "Join customer data with order data to find customers who have placed orders totaling more than $1000"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Create sample customer data
    customers := []streamv3.Record{
        streamv3.MakeMutableRecord().String("customer_id", "C001").String("name", "Alice Johnson").String("city", "New York").Freeze(),
        streamv3.MakeMutableRecord().String("customer_id", "C002").String("name", "Bob Smith").String("city", "Los Angeles").Freeze(),
        streamv3.MakeMutableRecord().String("customer_id", "C003").String("name", "Carol Davis").String("city", "Chicago").Freeze(),
    }

    // Create sample order data
    orders := []streamv3.Record{
        streamv3.MakeMutableRecord().String("customer_id", "C001").Float("amount", 500.0).String("product", "Laptop").Freeze(),
        streamv3.MakeMutableRecord().String("customer_id", "C001").Float("amount", 800.0).String("product", "Monitor").Freeze(),
        streamv3.MakeMutableRecord().String("customer_id", "C002").Float("amount", 200.0).String("product", "Mouse").Freeze(),
        streamv3.MakeMutableRecord().String("customer_id", "C003").Float("amount", 1200.0).String("product", "Workstation").Freeze(),
    }

    // Join customers with their orders
    customerOrders := streamv3.InnerJoin(
        slices.Values(orders),
        streamv3.OnFields("customer_id"),
    )(slices.Values(customers))

    // Group by customer to calculate total spending
    grouped := streamv3.GroupByFields("customer_spending", "customer_id")(customerOrders)

    customerTotals := streamv3.Aggregate("customer_spending", map[string]streamv3.AggregateFunc{
        "total_spent":   streamv3.Sum("amount"),
        "order_count":   streamv3.Count(),
        "customer_name": streamv3.First("name"),
        "city":          streamv3.First("city"),
    })(grouped)

    // Filter for customers with total orders > $1000
    highValueCustomers := streamv3.Where(func(r streamv3.Record) bool {
        total := streamv3.GetOr(r, "total_spent", 0.0)
        return total > 1000
    })(customerTotals)

    fmt.Println("High-value customers (>$1000 total orders):")
    for customer := range highValueCustomers {
        name := streamv3.GetOr(customer, "customer_name", "")
        city := streamv3.GetOr(customer, "city", "")
        total := streamv3.GetOr(customer, "total_spent", 0.0)
        orders := streamv3.GetOr(customer, "order_count", int64(0))

        fmt.Printf("%s (%s): $%.2f across %d orders\n", name, city, total, orders)
    }
}
```

### Example 7: Chart Creation

**Natural Language**: "Create an interactive chart showing monthly sales trends with the ability to filter by product category"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read sales data
    sales, err := streamv3.ReadCSV("monthly_sales.csv")
    if err != nil {
        log.Fatalf("Failed to read CSV: %v", err)
    }

    // Group by month and category
    grouped := streamv3.GroupByFields("monthly_analysis", "month", "category")(sales)

    // Calculate monthly totals by category
    monthlyTrends := streamv3.Aggregate("monthly_analysis", map[string]streamv3.AggregateFunc{
        "total_sales":     streamv3.Sum("sales_amount"),
        "avg_order_value": streamv3.Avg("order_value"),
        "order_count":     streamv3.Count(),
    })(grouped)

    // Create interactive chart configuration
    config := streamv3.DefaultChartConfig()
    config.Title = "Monthly Sales Trends by Category"
    config.ChartType = "line"
    config.Width = 1200
    config.Height = 600
    config.EnableZoom = true
    config.EnablePan = true
    config.ShowLegend = true

    // Generate the interactive chart
    err = streamv3.InteractiveChart(
        monthlyTrends,
        "monthly_sales_dashboard.html",
        config,
    )
    if err != nil {
        log.Fatalf("Failed to create chart: %v", err)
    }

    fmt.Println("Interactive dashboard created: monthly_sales_dashboard.html")
    fmt.Println("Features:")
    fmt.Println("- Zoom and pan controls")
    fmt.Println("- Category filtering")
    fmt.Println("- Hover tooltips with detailed data")
    fmt.Println("- Export capabilities")
}
```

### Example 8: Performance Comparison Chart

**Natural Language**: "Create a bar chart comparing average response times across different API endpoints, highlighting endpoints with response times over 500ms"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read API performance data
    apiLogs, err := streamv3.ReadCSV("api_performance.csv")
    if err != nil {
        log.Fatalf("Failed to read CSV: %v", err)
    }

    // Group by endpoint
    grouped := streamv3.GroupByFields("endpoint_analysis", "endpoint")(apiLogs)

    // Calculate performance metrics per endpoint
    endpointStats := streamv3.Aggregate("endpoint_analysis", map[string]streamv3.AggregateFunc{
        "avg_response_time": streamv3.Avg("response_time_ms"),
        "max_response_time": streamv3.Max[float64]("response_time_ms"),
        "min_response_time": streamv3.Min[float64]("response_time_ms"),
        "request_count":     streamv3.Count(),
    })(grouped)

    // Add performance status (slow if > 500ms)
    enrichedStats := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        avgTime := streamv3.GetOr(r, "avg_response_time", 0.0)
        status := "Normal"
        if avgTime > 500 {
            status = "Slow"
        }
        return streamv3.SetImmutable(r, "performance_status", status)
    })(endpointStats)

    // Sort by average response time (descending)
    sortedStats := streamv3.SortBy(func(r streamv3.Record) float64 {
        return -streamv3.GetOr(r, "avg_response_time", 0.0)
    })(enrichedStats)

    // Create bar chart
    config := streamv3.DefaultChartConfig()
    config.Title = "API Endpoint Performance Analysis"
    config.ChartType = "bar"
    config.Width = 1000
    config.Height = 600
    config.ShowDataLabels = true

    err = streamv3.InteractiveChart(
        sortedStats,
        "api_performance_chart.html",
        config,
    )
    if err != nil {
        log.Fatalf("Failed to create chart: %v", err)
    }

    // Also display summary to console
    fmt.Println("API Performance Summary:")
    for stat := range streamv3.Limit[streamv3.Record](10)(sortedStats) {
        endpoint := streamv3.GetOr(stat, "endpoint", "")
        avgTime := streamv3.GetOr(stat, "avg_response_time", 0.0)
        status := streamv3.GetOr(stat, "performance_status", "")
        requests := streamv3.GetOr(stat, "request_count", int64(0))

        statusIcon := "✅"
        if status == "Slow" {
            statusIcon = "🚨"
        }

        fmt.Printf("%s %s: %.1fms avg (%d requests)\n",
            statusIcon, endpoint, avgTime, requests)
    }

    fmt.Println("\nChart created: api_performance_chart.html")
}
```

---

## Pattern Recognition Rules

When processing natural language requests, map common phrases to StreamV3 operations:

1. **"Filter/Where/Only"** → `streamv3.Where(predicate)`
2. **"Transform/Convert/Calculate"** → `streamv3.Select(transformFn)`
3. **"Group by X"** → `streamv3.GroupByFields("groupName", "X")`
4. **"Count/Sum/Average"** → `streamv3.Aggregate("groupName", aggregations)`
5. **"Top N/First N/Limit"** → `streamv3.Limit(n)`
6. **"Sort by/Order by"** → `streamv3.SortBy(keyFn)`
7. **"Join/Combine"** → `streamv3.InnerJoin(rightSeq, predicate)`
8. **"In batches/Windows"** → `streamv3.CountWindow[T](size)` or `streamv3.TimeWindow[T](duration, field)`
9. **"Real-time/Streaming"** → Use generators with time-based operations
10. **"Chart/Visualize"** → `streamv3.QuickChart()` or `streamv3.InteractiveChart()`

---

## Code Generation Principles

### 1. Human-Readable Code Structure

```go
// ✅ CORRECT - Easy to read and verify
// Step 1: Filter data
filteredData := streamv3.Where(predicate)(data)

// Step 2: Group by region
grouped := streamv3.GroupByFields("analysis", "region")(filteredData)

// Step 3: Calculate totals
results := streamv3.Aggregate("analysis", aggregations)(grouped)
```

### 2. Error Handling (Always!)

```go
// ✅ CORRECT - Always handle I/O errors
data, err := streamv3.ReadCSV("file.csv")
if err != nil {
    log.Fatalf("Failed to read CSV: %v", err)
}

// ✅ CORRECT - Check chart creation errors
err = streamv3.QuickChart(data, "x", "y", "chart.html")
if err != nil {
    log.Fatalf("Failed to create chart: %v", err)
}
```

### 3. Type Safety

```go
// ✅ CORRECT - Use correct types for CSV data
age := streamv3.GetOr(record, "age", int64(0))      // CSV number → int64
price := streamv3.GetOr(record, "price", 0.0)       // CSV decimal → float64
name := streamv3.GetOr(record, "name", "")          // CSV text → string
active := streamv3.GetOr(record, "active", false)   // CSV bool → bool
```

### 4. Clear Variable Names

```go
// ✅ CORRECT - Descriptive names
highValueSales := streamv3.Where(...)(sales)
groupedByRegion := streamv3.GroupByFields(...)(highValueSales)
regionTotals := streamv3.Aggregate(...)(groupedByRegion)

// ❌ WRONG - Unclear abbreviations
hvs := streamv3.Where(...)(s)
gbr := streamv3.GroupByFields(...)(hvs)
rt := streamv3.Aggregate(...)(gbr)
```

### 5. SQL-Style Naming (Critical!)

```go
// ✅ CORRECT - StreamV3 uses SQL-style names
filtered := streamv3.Where(predicate)(data)
transformed := streamv3.Select(fn)(filtered)
limited := streamv3.Limit[T](10)(transformed)

// ❌ WRONG - Don't use LINQ/functional names
filtered := streamv3.Filter(predicate)(data)   // Filter is a type, not a function!
transformed := streamv3.Map(fn)(filtered)      // Map doesn't exist!
limited := streamv3.Take(10)(transformed)      // Take doesn't exist!
```

---

## Common Patterns

### CSV Analysis Pipeline

**Quick view:**
```go
// 1. Read with error handling
data, err := streamv3.ReadCSV("sales.csv")
if err != nil {
    log.Fatal(err)
}

// 2. Filter data
filtered := streamv3.Where(predicate)(data)

// 3. Group and aggregate
grouped := streamv3.GroupByFields("analysis", "region")(filtered)
results := streamv3.Aggregate("analysis", aggregations)(grouped)

// 4. Display results
for result := range results {
    // Process each result
}
```

<details>
<summary>📋 <b>Click for complete, runnable code</b></summary>

```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // 1. Read with error handling
    data, err := streamv3.ReadCSV("sales.csv")
    if err != nil {
        log.Fatal(err)
    }

    // 2. Filter data
    predicate := func(r streamv3.Record) bool {
        amount := streamv3.GetOr(r, "amount", 0.0)
        return amount > 1000
    }
    filtered := streamv3.Where(predicate)(data)

    // 3. Group and aggregate
    grouped := streamv3.GroupByFields("analysis", "region")(filtered)
    aggregations := map[string]streamv3.AggregateFunc{
        "total":   streamv3.Sum("amount"),
        "count":   streamv3.Count(),
        "average": streamv3.Avg("amount"),
    }
    results := streamv3.Aggregate("analysis", aggregations)(grouped)

    // 4. Display results
    fmt.Println("Sales Analysis by Region:")
    for result := range results {
        region := streamv3.GetOr(result, "region", "")
        total := streamv3.GetOr(result, "total", 0.0)
        count := streamv3.GetOr(result, "count", int64(0))
        avg := streamv3.GetOr(result, "average", 0.0)
        fmt.Printf("%s: $%.2f total, %d sales, $%.2f average\n", region, total, count, avg)
    }
}
```

</details>

### Real-Time Processing Pipeline

**Quick view:**
```go
// 1. Create or read stream
stream := /* data source */

// 2. Window the data
windowed := streamv3.TimeWindow[T](duration, "timestamp")(stream)

// 3. Process each window
for window := range windowed {
    // Analyze window
}
```

<details>
<parameter name="summary">📋 <b>Click for complete, runnable code</b></summary>

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // 1. Create or read stream
    stream, err := streamv3.ReadCSV("events.csv")
    if err != nil {
        log.Fatal(err)
    }

    // 2. Window the data
    duration := 5 * time.Minute
    windowed := streamv3.TimeWindow[streamv3.Record](duration, "timestamp")(stream)

    // 3. Process each window
    fmt.Println("Processing 5-minute windows:")
    for window := range windowed {
        count := len(window)
        fmt.Printf("Window: %d events\n", count)
        // Analyze window - aggregate, filter, transform, etc.
    }
}
```

</details>

### Join and Analyze Pipeline
```go
// 1. Read both datasets
left, err := streamv3.ReadCSV("customers.csv")
if err != nil {
    log.Fatal(err)
}
right, err := streamv3.ReadCSV("orders.csv")
if err != nil {
    log.Fatal(err)
}

// 2. Join on common field
joined := streamv3.InnerJoin(right, streamv3.OnFields("customer_id"))(left)

// 3. Process joined data
processed := streamv3.Select(enrichment)(joined)

// 4. Aggregate results
grouped := streamv3.GroupByFields("analysis", "customer_id")(processed)
results := streamv3.Aggregate("analysis", aggregations)(grouped)
```

---

## Best Practices Summary

1. **Always handle errors** from I/O operations
2. **Use SQL-style names** (Select, Where, Limit)
3. **Break complex pipelines** into clear steps
4. **Use descriptive variable names**
5. **Add type parameters** when needed: `Limit[streamv3.Record](10)`
6. **Only import packages** that are actually used
7. **Check godoc** when unsure: `go doc github.com/rosscartlidge/streamv3.FunctionName`

---

*For complete API documentation: `go doc github.com/rosscartlidge/streamv3`*
