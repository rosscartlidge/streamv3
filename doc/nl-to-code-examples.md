# Natural Language to StreamV3 Code Examples

*Training examples for LLMs to learn StreamV3 code generation patterns*

## Overview

This document provides comprehensive examples of natural language descriptions paired with their corresponding StreamV3 implementations. Use these examples to train LLMs on StreamV3 patterns and idioms.

---

## Data Analysis Examples

### Example 1: Basic Filtering and Aggregation

**Natural Language**: "Read employee data from employees.csv, filter for employees with salary over $80,000, and count how many employees there are by department"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read employee data
    employees := streamv3.ReadCSV("employees.csv")

    // Filter for high salary employees
    highSalaryEmployees := streamv3.Where(func(r streamv3.Record) bool {
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
        dept := streamv3.GetOr(result, "GroupValue", "")
        count := streamv3.GetOr(result, "employee_count", 0)
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
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read sales data
    sales := streamv3.ReadCSV("sales.csv")

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
        name := streamv3.GetOr(product, "GroupValue", "")
        revenue := streamv3.GetOr(product, "total_revenue", 0.0)
        fmt.Printf("%d. %s: $%.2f\n", rank, name, revenue)
        rank++
    }
}
```

### Example 3: Time Series Analysis

**Natural Language**: "Analyze daily website traffic, calculate 7-day moving averages, and identify days with traffic 50% above average"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read traffic data
    traffic := streamv3.ReadCSV("traffic.csv")

    // Sort by date to ensure proper time series
    sortedTraffic := streamv3.SortBy(func(r streamv3.Record) string {
        return streamv3.GetOr(r, "date", "")
    })(traffic)

    // Calculate 7-day moving average using sliding window
    movingAvgTraffic := streamv3.Select(func(window []streamv3.Record) streamv3.Record {
        if len(window) == 0 {
            return streamv3.NewRecord().Build()
        }

        // Calculate average for this window
        var total float64
        var currentRecord streamv3.Record
        for i, record := range window {
            visits := streamv3.GetOr(record, "daily_visits", 0.0)
            total += visits
            if i == len(window)-1 { // Last record in window
                currentRecord = record
            }
        }
        avg := total / float64(len(window))

        // Add moving average to the current day's record
        return streamv3.SetField(currentRecord, "moving_avg_7day", avg)
    })(streamv3.SlidingCountWindow[streamv3.Record](7, 1)(sortedTraffic))

    // Filter for days with traffic 50% above moving average
    highTrafficDays := streamv3.Where(func(r streamv3.Record) bool {
        visits := streamv3.GetOr(r, "daily_visits", 0.0)
        avg := streamv3.GetOr(r, "moving_avg_7day", 0.0)
        return avg > 0 && visits > avg*1.5
    })(movingAvgTraffic)

    fmt.Println("Days with traffic 50% above 7-day average:")
    for day := range highTrafficDays {
        date := streamv3.GetOr(day, "date", "")
        visits := streamv3.GetOr(day, "daily_visits", 0.0)
        avg := streamv3.GetOr(day, "moving_avg_7day", 0.0)
        increase := ((visits - avg) / avg) * 100
        fmt.Printf("%s: %.0f visits (%.1f%% above average of %.0f)\n",
            date, visits, increase, avg)
    }
}
```

---

## Data Transformation Examples

### Example 4: Data Enrichment

**Natural Language**: "Read customer data, add a customer_tier field based on total purchases (Bronze < $1000, Silver $1000-$5000, Gold > $5000)"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read customer data
    customers := streamv3.ReadCSV("customers.csv")

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

        return streamv3.SetField(r, "customer_tier", tier)
    })(customers)

    // Group by tier and count customers
    grouped := streamv3.GroupByFields("tier_analysis", "customer_tier")(enrichedCustomers)

    tierCounts := streamv3.Aggregate("tier_analysis", map[string]streamv3.AggregateFunc{
        "customer_count": streamv3.Count(),
        "avg_purchases": streamv3.Avg("total_purchases"),
    })(grouped)

    fmt.Println("Customer tier distribution:")
    for result := range tierCounts {
        tier := streamv3.GetOr(result, "GroupValue", "")
        count := streamv3.GetOr(result, "customer_count", 0)
        avgPurchases := streamv3.GetOr(result, "avg_purchases", 0.0)
        fmt.Printf("%s: %d customers (avg: $%.2f)\n", tier, count, avgPurchases)
    }
}
```

### Example 5: Data Cleaning and Validation

**Natural Language**: "Clean product data by removing items with missing names or negative prices, and standardize category names to title case"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "strings"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read product data
    products := streamv3.ReadCSV("products.csv")

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
        return streamv3.SetField(r, "category", standardizedCategory)
    })(validProducts)

    // Count valid products by category
    grouped := streamv3.GroupByFields("category_analysis", "category")(cleanedProducts)

    categoryStats := streamv3.Aggregate("category_analysis", map[string]streamv3.AggregateFunc{
        "product_count": streamv3.Count(),
        "avg_price": streamv3.Avg("price"),
        "min_price": streamv3.Min[float64]("price"),
        "max_price": streamv3.Max[float64]("price"),
    })(grouped)

    fmt.Println("Cleaned product data by category:")
    for result := range categoryStats {
        category := streamv3.GetOr(result, "GroupValue", "")
        count := streamv3.GetOr(result, "product_count", 0)
        avgPrice := streamv3.GetOr(result, "avg_price", 0.0)
        minPrice := streamv3.GetOr(result, "min_price", 0.0)
        maxPrice := streamv3.GetOr(result, "max_price", 0.0)

        fmt.Printf("%s: %d products, price range $%.2f-$%.2f (avg: $%.2f)\n",
            category, count, minPrice, maxPrice, avgPrice)
    }
}
```

---

## Real-Time Processing Examples

### Example 6: Streaming Alerts

**Natural Language**: "Monitor server logs in real-time, process in 1-minute windows, and alert if error rate exceeds 5%"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "time"
    "strings"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Simulate real-time log stream
    logStream := func(yield func(streamv3.Record) bool) {
        statuses := []string{"200", "200", "200", "200", "404", "500", "200"}

        for i := 0; ; i++ {
            status := statuses[i%len(statuses)]

            record := streamv3.NewRecord().
                String("status", status).
                String("endpoint", fmt.Sprintf("/api/endpoint%d", i%3)).
                Time("timestamp", time.Now()).
                Build()

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

### Example 7: IoT Sensor Processing

**Natural Language**: "Process IoT temperature sensors, detect anomalies (readings outside 2 standard deviations), and group alerts by sensor location"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "time"
    "math"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Simulate IoT sensor data
    sensorStream := func(yield func(streamv3.Record) bool) {
        locations := []string{"warehouse_a", "warehouse_b", "office"}
        baseTemp := 22.0

        for i := 0; i < 200; i++ {
            location := locations[i%len(locations)]

            // Simulate normal variation with occasional anomalies
            var temp float64
            if i%50 == 0 { // Inject anomaly every 50 readings
                temp = baseTemp + float64(i%2*20-10) // Extreme values
            } else {
                temp = baseTemp + float64(i%10-5) // Normal variation
            }

            record := streamv3.NewRecord().
                String("sensor_id", fmt.Sprintf("temp_%s_%d", location, i%3)).
                String("location", location).
                Float("temperature", temp).
                Time("timestamp", time.Now().Add(time.Duration(i)*time.Second)).
                Build()

            if !yield(record) {
                return
            }
        }
    }

    // Process in windows for statistical analysis
    windowed := streamv3.CountWindow[streamv3.Record](20)(sensorStream)

    fmt.Println("IoT Anomaly Detection:")
    for window := range windowed {
        if len(window) < 10 { // Need sufficient data for statistics
            continue
        }

        // Calculate mean and standard deviation
        var temps []float64
        for _, record := range window {
            temp := streamv3.GetOr(record, "temperature", 0.0)
            temps = append(temps, temp)
        }

        // Calculate stats
        var sum float64
        for _, temp := range temps {
            sum += temp
        }
        mean := sum / float64(len(temps))

        var squareSum float64
        for _, temp := range temps {
            squareSum += (temp - mean) * (temp - mean)
        }
        stdDev := math.Sqrt(squareSum / float64(len(temps)))

        // Detect anomalies (outside 2 standard deviations)
        anomalies := []streamv3.Record{}
        for _, record := range window {
            temp := streamv3.GetOr(record, "temperature", 0.0)
            if math.Abs(temp-mean) > 2*stdDev {
                anomalies = append(anomalies, record)
            }
        }

        if len(anomalies) > 0 {
            // Group anomalies by location
            anomalyStream := streamv3.From(anomalies)
            grouped := streamv3.GroupByFields("anomaly_analysis", "location")(anomalyStream)

            fmt.Printf("Window Stats: Mean=%.1f°C, StdDev=%.1f°C\n", mean, stdDev)

            for group := range grouped {
                location := streamv3.GetOr(group, "GroupValue", "")
                sensors, _ := streamv3.Get[[]streamv3.Record](group, "anomaly_analysis")

                fmt.Printf("🚨 Anomalies in %s (%d sensors):\n", location, len(sensors))
                for _, sensor := range sensors {
                    sensorId := streamv3.GetOr(sensor, "sensor_id", "")
                    temp := streamv3.GetOr(sensor, "temperature", 0.0)
                    deviation := math.Abs(temp - mean) / stdDev
                    fmt.Printf("  %s: %.1f°C (%.1fσ from mean)\n", sensorId, temp, deviation)
                }
            }
        }
    }
}
```

---

## Join Examples

### Example 8: Customer Order Analysis

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
        streamv3.NewRecord().String("customer_id", "C001").String("name", "Alice Johnson").String("city", "New York").Build(),
        streamv3.NewRecord().String("customer_id", "C002").String("name", "Bob Smith").String("city", "Los Angeles").Build(),
        streamv3.NewRecord().String("customer_id", "C003").String("name", "Carol Davis").String("city", "Chicago").Build(),
    }

    // Create sample order data
    orders := []streamv3.Record{
        streamv3.NewRecord().String("customer_id", "C001").Float("amount", 500).String("product", "Laptop").Build(),
        streamv3.NewRecord().String("customer_id", "C001").Float("amount", 800).String("product", "Monitor").Build(),
        streamv3.NewRecord().String("customer_id", "C002").Float("amount", 200).String("product", "Mouse").Build(),
        streamv3.NewRecord().String("customer_id", "C003").Float("amount", 1200).String("product", "Workstation").Build(),
    }

    // Join customers with their orders
    customerOrders := streamv3.InnerJoin(
        slices.Values(orders),
        streamv3.OnFields("customer_id"),
    )(slices.Values(customers))

    // Group by customer to calculate total spending
    grouped := streamv3.GroupByFields("customer_spending", "customer_id")(customerOrders)

    customerTotals := streamv3.Aggregate("customer_spending", map[string]streamv3.AggregateFunc{
        "total_spent": streamv3.Sum("amount"),
        "order_count": streamv3.Count(),
        "customer_name": streamv3.First("name"),
        "city": streamv3.First("city"),
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
        orders := streamv3.GetOr(customer, "order_count", 0)

        fmt.Printf("%s (%s): $%.2f across %d orders\n", name, city, total, orders)
    }
}
```

### Example 9: Product Performance Analysis

**Natural Language**: "Join product catalog with sales data and inventory data to find products that are selling well but have low stock"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Product catalog
    products := []streamv3.Record{
        streamv3.NewRecord().String("product_id", "P001").String("name", "Gaming Laptop").String("category", "Electronics").Build(),
        streamv3.NewRecord().String("product_id", "P002").String("name", "Office Chair").String("category", "Furniture").Build(),
        streamv3.NewRecord().String("product_id", "P003").String("name", "Coffee Maker").String("category", "Appliances").Build(),
    }

    // Sales data (last 30 days)
    sales := []streamv3.Record{
        streamv3.NewRecord().String("product_id", "P001").Int("quantity_sold", 25).Float("revenue", 25000).Build(),
        streamv3.NewRecord().String("product_id", "P001").Int("quantity_sold", 15).Float("revenue", 15000).Build(),
        streamv3.NewRecord().String("product_id", "P002").Int("quantity_sold", 8).Float("revenue", 2400).Build(),
        streamv3.NewRecord().String("product_id", "P003").Int("quantity_sold", 12).Float("revenue", 1800).Build(),
    }

    // Inventory data
    inventory := []streamv3.Record{
        streamv3.NewRecord().String("product_id", "P001").Int("stock_level", 5).Int("reorder_point", 10).Build(),
        streamv3.NewRecord().String("product_id", "P002").Int("stock_level", 25).Int("reorder_point", 15).Build(),
        streamv3.NewRecord().String("product_id", "P003").Int("stock_level", 3).Int("reorder_point", 8).Build(),
    }

    // First, aggregate sales by product
    salesGrouped := streamv3.GroupByFields("sales_summary", "product_id")(slices.Values(sales))

    salesSummary := streamv3.Aggregate("sales_summary", map[string]streamv3.AggregateFunc{
        "total_quantity": streamv3.Sum("quantity_sold"),
        "total_revenue": streamv3.Sum("revenue"),
    })(salesGrouped)

    // Join products with sales summary
    productSales := streamv3.InnerJoin(
        salesSummary,
        streamv3.OnFields("product_id"),
    )(slices.Values(products))

    // Join with inventory data
    fullData := streamv3.InnerJoin(
        slices.Values(inventory),
        streamv3.OnFields("product_id"),
    )(productSales)

    // Find products selling well but with low stock
    lowStockHighSales := streamv3.Where(func(r streamv3.Record) bool {
        totalSold := streamv3.GetOr(r, "total_quantity", 0.0)
        stockLevel := streamv3.GetOr(r, "stock_level", 0)
        reorderPoint := streamv3.GetOr(r, "reorder_point", 0)

        // High sales (>10 units) and below reorder point
        return totalSold > 10 && stockLevel <= reorderPoint
    })(fullData)

    fmt.Println("Products with high sales but low stock (reorder needed):")
    for product := range lowStockHighSales {
        name := streamv3.GetOr(product, "name", "")
        category := streamv3.GetOr(product, "category", "")
        sold := streamv3.GetOr(product, "total_quantity", 0.0)
        revenue := streamv3.GetOr(product, "total_revenue", 0.0)
        stock := streamv3.GetOr(product, "stock_level", 0)
        reorderPoint := streamv3.GetOr(product, "reorder_point", 0)

        fmt.Printf("⚠️  %s (%s)\n", name, category)
        fmt.Printf("   Sales: %.0f units, $%.2f revenue\n", sold, revenue)
        fmt.Printf("   Stock: %d units (reorder at %d)\n", stock, reorderPoint)
        fmt.Println()
    }
}
```

---

## Visualization Examples

### Example 10: Dashboard Creation

**Natural Language**: "Create an interactive chart showing monthly sales trends with the ability to filter by product category"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read sales data
    sales := streamv3.ReadCSV("monthly_sales.csv")

    // Group by month and category
    grouped := streamv3.GroupByFields("monthly_analysis", "month", "category")(sales)

    // Calculate monthly totals by category
    monthlyTrends := streamv3.Aggregate("monthly_analysis", map[string]streamv3.AggregateFunc{
        "total_sales": streamv3.Sum("sales_amount"),
        "avg_order_value": streamv3.Avg("order_value"),
        "order_count": streamv3.Count(),
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
        panic(err)
    }

    fmt.Println("Interactive dashboard created: monthly_sales_dashboard.html")
    fmt.Println("Features:")
    fmt.Println("- Zoom and pan controls")
    fmt.Println("- Category filtering")
    fmt.Println("- Hover tooltips with detailed data")
    fmt.Println("- Export capabilities")
}
```

### Example 11: Performance Comparison Chart

**Natural Language**: "Create a bar chart comparing average response times across different API endpoints, highlighting endpoints with response times over 500ms"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read API performance data
    apiLogs := streamv3.ReadCSV("api_performance.csv")

    // Group by endpoint
    grouped := streamv3.GroupByFields("endpoint_analysis", "endpoint")(apiLogs)

    // Calculate performance metrics per endpoint
    endpointStats := streamv3.Aggregate("endpoint_analysis", map[string]streamv3.AggregateFunc{
        "avg_response_time": streamv3.Avg("response_time_ms"),
        "max_response_time": streamv3.Max[float64]("response_time_ms"),
        "min_response_time": streamv3.Min[float64]("response_time_ms"),
        "request_count": streamv3.Count(),
    })(grouped)

    // Add performance status (slow if > 500ms)
    enrichedStats := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        avgTime := streamv3.GetOr(r, "avg_response_time", 0.0)
        status := "Normal"
        if avgTime > 500 {
            status = "Slow"
        }
        return streamv3.SetField(r, "performance_status", status)
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
        panic(err)
    }

    // Also display summary to console
    fmt.Println("API Performance Summary:")
    for stat := range streamv3.Limit[streamv3.Record](10)(sortedStats) {
        endpoint := streamv3.GetOr(stat, "GroupValue", "")
        avgTime := streamv3.GetOr(stat, "avg_response_time", 0.0)
        status := streamv3.GetOr(stat, "performance_status", "")
        requests := streamv3.GetOr(stat, "request_count", 0)

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

## Advanced Pattern Examples

### Example 12: Complex Data Pipeline

**Natural Language**: "Build a data pipeline that reads transaction data, detects fraudulent patterns (multiple high-value transactions from same IP within 1 hour), enriches with customer risk scores, and generates alerts"

**StreamV3 Code**:
```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read transaction data
    transactions := streamv3.ReadCSV("transactions.csv")

    // Step 1: Filter for high-value transactions (>$1000)
    highValueTxns := streamv3.Where(func(r streamv3.Record) bool {
        amount := streamv3.GetOr(r, "amount", 0.0)
        return amount > 1000
    })(transactions)

    // Step 2: Add hour field for time-based grouping
    enrichedTxns := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        timestamp := streamv3.GetOr(r, "timestamp", "")
        if t, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
            hour := t.Format("2006-01-02 15")
            return streamv3.SetField(r, "hour_window", hour)
        }
        return r
    })(highValueTxns)

    // Step 3: Group by IP and hour window
    grouped := streamv3.GroupByFields("fraud_analysis", "ip_address", "hour_window")(enrichedTxns)

    // Step 4: Detect potential fraud (multiple high-value txns from same IP within 1 hour)
    fraudCandidates := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        transactions, _ := streamv3.Get[[]streamv3.Record](r, "fraud_analysis")

        if len(transactions) < 2 {
            return streamv3.NewRecord().Build() // Skip if not multiple transactions
        }

        // Calculate fraud indicators
        var totalAmount float64
        var customerIds []string
        var countries []string

        for _, txn := range transactions {
            amount := streamv3.GetOr(txn, "amount", 0.0)
            customerId := streamv3.GetOr(txn, "customer_id", "")
            country := streamv3.GetOr(txn, "country", "")

            totalAmount += amount
            customerIds = append(customerIds, customerId)
            countries = append(countries, country)
        }

        // Build fraud alert record
        return streamv3.NewRecord().
            String("ip_address", streamv3.GetOr(r, "GroupValue", "")).
            String("time_window", streamv3.GetOr(r, "hour_window", "")).
            Int("transaction_count", len(transactions)).
            Float("total_amount", totalAmount).
            Int("unique_customers", len(uniqueStrings(customerIds))).
            Int("unique_countries", len(uniqueStrings(countries))).
            Float("fraud_score", calculateFraudScore(len(transactions), totalAmount, len(uniqueStrings(customerIds)))).
            Build()
    })(grouped)

    // Step 5: Filter for high fraud scores
    fraudAlerts := streamv3.Where(func(r streamv3.Record) bool {
        score := streamv3.GetOr(r, "fraud_score", 0.0)
        return score > 7.0 // High fraud score threshold
    })(fraudCandidates)

    // Step 6: Sort by fraud score (highest first)
    sortedAlerts := streamv3.SortBy(func(r streamv3.Record) float64 {
        return -streamv3.GetOr(r, "fraud_score", 0.0)
    })(fraudAlerts)

    // Generate fraud alerts
    fmt.Println("🚨 FRAUD DETECTION ALERTS:")
    fmt.Println("="*50)

    alertCount := 0
    for alert := range sortedAlerts {
        ip := streamv3.GetOr(alert, "ip_address", "")
        timeWindow := streamv3.GetOr(alert, "time_window", "")
        txnCount := streamv3.GetOr(alert, "transaction_count", 0)
        totalAmount := streamv3.GetOr(alert, "total_amount", 0.0)
        uniqueCustomers := streamv3.GetOr(alert, "unique_customers", 0)
        uniqueCountries := streamv3.GetOr(alert, "unique_countries", 0)
        fraudScore := streamv3.GetOr(alert, "fraud_score", 0.0)

        alertCount++
        fmt.Printf("Alert #%d - Fraud Score: %.1f\n", alertCount, fraudScore)
        fmt.Printf("  IP Address: %s\n", ip)
        fmt.Printf("  Time Window: %s\n", timeWindow)
        fmt.Printf("  Transactions: %d high-value transactions\n", txnCount)
        fmt.Printf("  Total Amount: $%.2f\n", totalAmount)
        fmt.Printf("  Customers: %d unique\n", uniqueCustomers)
        fmt.Printf("  Countries: %d unique\n", uniqueCountries)
        fmt.Println()
    }

    if alertCount == 0 {
        fmt.Println("✅ No fraud alerts detected")
    }
}

// Helper function to calculate fraud score
func calculateFraudScore(txnCount int, totalAmount float64, uniqueCustomers int) float64 {
    score := 0.0

    // Multiple transactions weight
    if txnCount >= 5 {
        score += 3.0
    } else if txnCount >= 3 {
        score += 2.0
    } else {
        score += 1.0
    }

    // High amount weight
    if totalAmount > 10000 {
        score += 3.0
    } else if totalAmount > 5000 {
        score += 2.0
    } else {
        score += 1.0
    }

    // Multiple customers from same IP weight
    if uniqueCustomers > 1 {
        score += 4.0
    }

    return score
}

// Helper function to get unique strings
func uniqueStrings(strings []string) []string {
    keys := make(map[string]bool)
    var unique []string

    for _, str := range strings {
        if !keys[str] {
            keys[str] = true
            unique = append(unique, str)
        }
    }

    return unique
}
```

---

## Key Learning Patterns

### Pattern Recognition Rules for LLMs:

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

### Common Code Structures:

```go
// Basic Analysis Pipeline
data := streamv3.ReadCSV("file.csv")
filtered := streamv3.Where(predicate)(data)
grouped := streamv3.GroupByFields("analysis", "field")(filtered)
results := streamv3.Aggregate("analysis", aggregations)(grouped)

// Real-time Processing
windowed := streamv3.TimeWindow[streamv3.Record](duration, "timestamp")(stream)
for window := range windowed { /* process window */ }

// Join Analysis
joined := streamv3.InnerJoin(rightStream, streamv3.OnFields("key"))(leftStream)
processed := streamv3.Select(enrichmentFn)(joined)

// Chart Creation
config := streamv3.DefaultChartConfig()
config.Title = "Chart Title"
streamv3.InteractiveChart(data, "output.html", config)
```

These examples provide comprehensive training data for LLMs to learn StreamV3 patterns and generate accurate code from natural language descriptions.