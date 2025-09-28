# StreamV3 Advanced Tutorial

*Master complex stream processing, aggregations, and real-time analytics*

## Prerequisites

This tutorial assumes you've completed the [Introduction Codelab](codelab-intro.md) and understand:
- Basic stream operations (Map, Where, Limit)
- Working with Records
- Creating simple charts

## Complex Aggregations

### Multi-Level Grouping

Process sales data across multiple dimensions:

```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Create complex sales data
    data := generateSalesData() // Implementation below

    // Group by region, then by product category
    regionGroups := streamv3.GroupRecordsByFields(data, "region")

    results := []streamv3.Record{}
    for regionGroup := range regionGroups {
        region := regionGroup.Key["region"].(string)

        // Sub-group by category within each region
        categoryGroups := streamv3.GroupRecordsByFields(
            streamv3.From(regionGroup.Records), "category")

        for categoryGroup := range categoryGroups {
            category := categoryGroup.Key["category"].(string)

            // Aggregate within each region-category combination
            agg := streamv3.AggregateGroups(
                streamv3.From([]streamv3.RecordGroup{categoryGroup}),
                map[string]streamv3.AggregateFunc{
                    "total_revenue": streamv3.Sum("revenue"),
                    "avg_deal_size": streamv3.Avg("deal_size"),
                    "deal_count":    streamv3.Count(),
                })

            for result := range agg {
                result["region"] = region
                result["category"] = category
                results = append(results, result)
            }
        }
    }

    // Create multi-dimensional visualization
    config := streamv3.DefaultChartConfig()
    config.Title = "Revenue by Region and Category"
    config.ChartType = "bar"
    config.EnableInteractive = true

    streamv3.InteractiveChart(streamv3.From(results), "regional_analysis.html", config)
    fmt.Println("📊 Multi-dimensional analysis: regional_analysis.html")
}

func generateSalesData() streamv3.Stream[streamv3.Record] {
    regions := []string{"North", "South", "East", "West"}
    categories := []string{"Software", "Hardware", "Services"}

    var records []streamv3.Record
    for i := 0; i < 1000; i++ {
        records = append(records, streamv3.NewRecord().
            String("region", regions[i%len(regions)]).
            String("category", categories[i%len(categories)]).
            Float("revenue", 1000+float64(i%50)*1000).
            Float("deal_size", 5000+float64(i%20)*2000).
            Time("date", time.Now().AddDate(0, -(i%12), 0)).
            Build())
    }
    return streamv3.From(records)
}
```

### Rolling Window Analytics

Calculate moving averages and trending metrics:

```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Generate time-series data
    timeSeries := generateTimeSeriesData()

    // Calculate 7-day rolling averages
    rollingAvg := calculateRollingAverage(timeSeries, 7)

    // Detect trends (increasing/decreasing)
    withTrends := addTrendAnalysis(rollingAvg)

    // Create time series visualization with trend analysis
    config := streamv3.DefaultChartConfig()
    config.Title = "Metrics with 7-Day Rolling Average and Trends"
    config.EnableCalculations = true
    config.ChartType = "line"

    streamv3.TimeSeriesChart(
        withTrends,
        "timestamp",
        []string{"value", "rolling_avg", "trend_strength"},
        "trending_analysis.html",
        config)

    fmt.Println("📊 Trending analysis: trending_analysis.html")
}

func calculateRollingAverage(data streamv3.Stream[streamv3.Record], windowSize int) streamv3.Stream[streamv3.Record] {
    buffer := make([]streamv3.Record, 0, windowSize)

    return streamv3.FlatMap(data, func(record streamv3.Record) streamv3.Stream[streamv3.Record] {
        buffer = append(buffer, record)
        if len(buffer) > windowSize {
            buffer = buffer[1:]
        }

        if len(buffer) == windowSize {
            sum := 0.0
            for _, r := range buffer {
                sum += r["value"].(float64)
            }
            avg := sum / float64(windowSize)

            result := streamv3.NewRecord()
            for k, v := range record {
                result.Set(k, v)
            }
            result.Float("rolling_avg", avg)

            return streamv3.From([]streamv3.Record{result.Build()})
        }
        return streamv3.From([]streamv3.Record{}) // Empty until window is full
    })
}

func addTrendAnalysis(data streamv3.Stream[streamv3.Record]) streamv3.Stream[streamv3.Record] {
    var prevAvg *float64

    return streamv3.Map(data, func(record streamv3.Record) streamv3.Record {
        currentAvg := record["rolling_avg"].(float64)

        result := streamv3.NewRecord()
        for k, v := range record {
            result.Set(k, v)
        }

        if prevAvg != nil {
            trendStrength := (currentAvg - *prevAvg) / *prevAvg * 100
            result.Float("trend_strength", trendStrength)

            trend := "stable"
            if trendStrength > 5 {
                trend = "increasing"
            } else if trendStrength < -5 {
                trend = "decreasing"
            }
            result.String("trend_direction", trend)
        } else {
            result.Float("trend_strength", 0)
            result.String("trend_direction", "stable")
        }

        prevAvg = &currentAvg
        return result.Build()
    })
}

func generateTimeSeriesData() streamv3.Stream[streamv3.Record] {
    baseTime := time.Now().AddDate(0, -1, 0) // Start 1 month ago
    var records []streamv3.Record

    for i := 0; i < 30; i++ {
        timestamp := baseTime.AddDate(0, 0, i)
        // Simulate some trending data with noise
        baseValue := 100.0 + float64(i)*2.5 // Upward trend
        noise := (rand.Float64() - 0.5) * 20  // ±10 noise

        records = append(records, streamv3.NewRecord().
            Time("timestamp", timestamp).
            Float("value", baseValue+noise).
            Build())
    }

    return streamv3.From(records)
}
```

## Stream Joins

### Inner Join

Combine data from multiple sources:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Customer data
    customers := streamv3.From([]streamv3.Record{
        streamv3.NewRecord().Int("id", 1).String("name", "Acme Corp").String("tier", "Enterprise").Build(),
        streamv3.NewRecord().Int("id", 2).String("name", "Beta LLC").String("tier", "Professional").Build(),
        streamv3.NewRecord().Int("id", 3).String("name", "Gamma Inc").String("tier", "Standard").Build(),
    })

    // Orders data
    orders := streamv3.From([]streamv3.Record{
        streamv3.NewRecord().Int("customer_id", 1).Float("amount", 15000).String("product", "Software").Build(),
        streamv3.NewRecord().Int("customer_id", 2).Float("amount", 8500).String("product", "Consulting").Build(),
        streamv3.NewRecord().Int("customer_id", 1).Float("amount", 22000).String("product", "Hardware").Build(),
        streamv3.NewRecord().Int("customer_id", 3).Float("amount", 3500).String("product", "Support").Build(),
    })

    // Join customers with their orders
    joined := joinStreams(customers, orders, "id", "customer_id")

    // Calculate total revenue by customer tier
    tierGroups := streamv3.GroupRecordsByFields(joined, "tier")
    tierRevenue := streamv3.AggregateGroups(tierGroups, map[string]streamv3.AggregateFunc{
        "total_revenue": streamv3.Sum("amount"),
        "order_count":   streamv3.Count(),
        "avg_order":     streamv3.Avg("amount"),
    })

    fmt.Println("Revenue by Customer Tier:")
    for result := range tierRevenue {
        fmt.Printf("  %s: $%.0f (%d orders, avg $%.0f)\n",
            result["tier"],
            result["total_revenue"],
            result["order_count"],
            result["avg_order"])
    }

    // Visualize the join results
    streamv3.QuickChart(streamv3.Collect(tierRevenue), "tier", "total_revenue", "tier_revenue.html")
}

func joinStreams(left, right streamv3.Stream[streamv3.Record], leftKey, rightKey string) streamv3.Stream[streamv3.Record] {
    // Build index from right stream
    rightIndex := make(map[interface{}][]streamv3.Record)
    for record := range right {
        key := record[rightKey]
        rightIndex[key] = append(rightIndex[key], record)
    }

    // Join with left stream
    return streamv3.FlatMap(left, func(leftRecord streamv3.Record) streamv3.Stream[streamv3.Record] {
        leftKeyValue := leftRecord[leftKey]
        matches, exists := rightIndex[leftKeyValue]

        if !exists {
            return streamv3.From([]streamv3.Record{}) // No matches (inner join)
        }

        var joined []streamv3.Record
        for _, rightRecord := range matches {
            // Merge records
            result := streamv3.NewRecord()
            for k, v := range leftRecord {
                result.Set(k, v)
            }
            for k, v := range rightRecord {
                if k != rightKey { // Avoid duplicate join key
                    result.Set(k, v)
                }
            }
            joined = append(joined, result.Build())
        }

        return streamv3.From(joined)
    })
}
```

## Real-Time Processing

### Live Data Monitoring

Process streaming data with real-time updates:

```go
package main

import (
    "fmt"
    "time"
    "context"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Simulate real-time metrics
    metricsStream := generateLiveMetrics(ctx)

    // Process in real-time with windowing
    windowedMetrics := processWithTimeWindows(metricsStream, 5*time.Second)

    // Monitor for alerts
    alerts := detectAlerts(windowedMetrics)

    // Display live results
    fmt.Println("🔴 Real-time monitoring started (30 seconds)...")

    alertCount := 0
    for alert := range alerts {
        alertCount++
        fmt.Printf("[%s] ALERT: %s - Value: %.2f (threshold: %.2f)\n",
            alert["timestamp"].(time.Time).Format("15:04:05"),
            alert["metric_name"],
            alert["current_value"],
            alert["threshold"])
    }

    fmt.Printf("\n📊 Monitoring complete. Total alerts: %d\n", alertCount)
}

func generateLiveMetrics(ctx context.Context) streamv3.Stream[streamv3.Record] {
    ch := make(chan streamv3.Record)

    go func() {
        defer close(ch)
        ticker := time.NewTicker(500 * time.Millisecond)
        defer ticker.Stop()

        metrics := []string{"cpu_usage", "memory_usage", "disk_io", "network_latency"}

        for {
            select {
            case <-ctx.Done():
                return
            case now := <-ticker.C:
                for _, metric := range metrics {
                    // Simulate varying metrics with occasional spikes
                    baseValue := 20.0 + rand.Float64()*60 // 20-80 normal range
                    if rand.Float64() < 0.1 { // 10% chance of spike
                        baseValue += 30
                    }

                    record := streamv3.NewRecord().
                        Time("timestamp", now).
                        String("metric_name", metric).
                        Float("value", baseValue).
                        Build()

                    select {
                    case ch <- record:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }
    }()

    return streamv3.FromChannel(ch)
}

func processWithTimeWindows(data streamv3.Stream[streamv3.Record], windowSize time.Duration) streamv3.Stream[streamv3.Record] {
    windows := make(map[string][]streamv3.Record) // metric_name -> records in window

    return streamv3.FlatMap(data, func(record streamv3.Record) streamv3.Stream[streamv3.Record] {
        metricName := record["metric_name"].(string)
        timestamp := record["timestamp"].(time.Time)

        // Add to window
        windows[metricName] = append(windows[metricName], record)

        // Remove old records from window
        cutoff := timestamp.Add(-windowSize)
        filtered := windows[metricName][:0]
        for _, r := range windows[metricName] {
            if r["timestamp"].(time.Time).After(cutoff) {
                filtered = append(filtered, r)
            }
        }
        windows[metricName] = filtered

        // Calculate window statistics
        if len(windows[metricName]) >= 3 { // Need minimum samples
            values := make([]float64, len(windows[metricName]))
            for i, r := range windows[metricName] {
                values[i] = r["value"].(float64)
            }

            avg := average(values)
            stddev := standardDeviation(values, avg)

            result := streamv3.NewRecord().
                Time("timestamp", timestamp).
                String("metric_name", metricName).
                Float("current_value", record["value"].(float64)).
                Float("window_avg", avg).
                Float("window_stddev", stddev).
                Int("window_size", int64(len(values))).
                Build()

            return streamv3.From([]streamv3.Record{result})
        }

        return streamv3.From([]streamv3.Record{})
    })
}

func detectAlerts(data streamv3.Stream[streamv3.Record]) streamv3.Stream[streamv3.Record] {
    return streamv3.Where(data, func(record streamv3.Record) bool {
        current := record["current_value"].(float64)
        avg := record["window_avg"].(float64)
        stddev := record["window_stddev"].(float64)

        // Alert if value is more than 2 standard deviations from average
        threshold := avg + 2*stddev
        isAlert := current > threshold

        if isAlert {
            // Add threshold info to record for display
            record["threshold"] = threshold
            record["severity"] = "high"
            if current > avg+3*stddev {
                record["severity"] = "critical"
            }
        }

        return isAlert
    })
}

func average(values []float64) float64 {
    sum := 0.0
    for _, v := range values {
        sum += v
    }
    return sum / float64(len(values))
}

func standardDeviation(values []float64, mean float64) float64 {
    variance := 0.0
    for _, v := range values {
        variance += (v - mean) * (v - mean)
    }
    variance /= float64(len(values))
    return math.Sqrt(variance)
}
```

## Advanced Visualizations

### Dashboard Creation

Build comprehensive analytics dashboards:

```go
package main

import (
    "fmt"
    "os"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Load multiple datasets
    salesData := loadSalesData()
    customerData := loadCustomerData()
    productData := loadProductData()

    // Create comprehensive dashboard
    createAnalyticsDashboard(salesData, customerData, productData)

    fmt.Println("📊 Analytics dashboard created!")
    fmt.Println("Open these files:")
    fmt.Println("  • revenue_trends.html - Time series analysis")
    fmt.Println("  • customer_segments.html - Customer segmentation")
    fmt.Println("  • product_performance.html - Product analysis")
    fmt.Println("  • geographic_distribution.html - Geographic analysis")
}

func createAnalyticsDashboard(sales, customers, products streamv3.Stream[streamv3.Record]) {
    os.MkdirAll("dashboard", 0755)

    // 1. Revenue Trends Over Time
    createRevenueTrends(sales)

    // 2. Customer Segmentation
    createCustomerSegmentation(customers, sales)

    // 3. Product Performance Analysis
    createProductAnalysis(products, sales)

    // 4. Geographic Distribution
    createGeographicAnalysis(customers, sales)
}

func createRevenueTrends(sales streamv3.Stream[streamv3.Record]) {
    // Group by month and calculate metrics
    monthlyRevenue := streamv3.GroupRecordsByFields(sales, "month")
    monthlyMetrics := streamv3.AggregateGroups(monthlyRevenue, map[string]streamv3.AggregateFunc{
        "total_revenue":   streamv3.Sum("revenue"),
        "avg_deal_size":   streamv3.Avg("deal_size"),
        "deal_count":      streamv3.Count(),
        "revenue_growth":  streamv3.Custom(calculateGrowthRate),
    })

    config := streamv3.DefaultChartConfig()
    config.Title = "Revenue Trends and Growth"
    config.ChartType = "line"
    config.EnableCalculations = true
    config.Height = 500
    config.CustomCSS = `
        .chart-container {
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            border-radius: 8px;
        }
    `

    streamv3.TimeSeriesChart(
        monthlyMetrics,
        "month",
        []string{"total_revenue", "avg_deal_size", "revenue_growth"},
        "dashboard/revenue_trends.html",
        config)
}

func createCustomerSegmentation(customers, sales streamv3.Stream[streamv3.Record]) {
    // Join customers with sales data
    enrichedCustomers := enrichCustomersWithSales(customers, sales)

    // Segment by value
    segments := segmentCustomersByValue(enrichedCustomers)

    config := streamv3.DefaultChartConfig()
    config.Title = "Customer Segmentation by Lifetime Value"
    config.ChartType = "scatter"
    config.EnableInteractive = true

    streamv3.InteractiveChart(segments, "dashboard/customer_segments.html", config)
}

func createProductAnalysis(products, sales streamv3.Stream[streamv3.Record]) {
    // Analyze product performance
    productMetrics := analyzeProductPerformance(products, sales)

    config := streamv3.DefaultChartConfig()
    config.Title = "Product Performance Matrix"
    config.ChartType = "bar"
    config.ShowDataLabels = true

    streamv3.InteractiveChart(productMetrics, "dashboard/product_performance.html", config)
}

func createGeographicAnalysis(customers, sales streamv3.Stream[streamv3.Record]) {
    // Geographic revenue distribution
    geoRevenue := analyzeGeographicDistribution(customers, sales)

    config := streamv3.DefaultChartConfig()
    config.Title = "Revenue by Geographic Region"
    config.ChartType = "pie"
    config.EnableAnimations = true

    streamv3.InteractiveChart(geoRevenue, "dashboard/geographic_distribution.html", config)
}

// Utility functions for dashboard creation
func calculateGrowthRate(records []streamv3.Record) interface{} {
    if len(records) < 2 {
        return 0.0
    }
    // Implement growth rate calculation logic
    return 5.2 // Placeholder
}

func enrichCustomersWithSales(customers, sales streamv3.Stream[streamv3.Record]) streamv3.Stream[streamv3.Record] {
    // Implementation of customer-sales join
    return customers // Placeholder
}

// Additional utility functions...
```

## Performance Optimization

### Memory-Efficient Processing

Handle large datasets efficiently:

```go
package main

import (
    "fmt"
    "bufio"
    "os"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Process large files without loading everything into memory
    processLargeFile("huge_dataset.csv")
}

func processLargeFile(filename string) {
    // Use streaming approach - never load entire file
    data := streamv3.ReadCSVStream(filename) // Hypothetical streaming reader

    // Process in chunks with limited memory usage
    results := data.
        Filter(func(r streamv3.Record) bool {
            // Only process relevant records
            return r["status"].(string) == "active"
        }).
        Map(func(r streamv3.Record) streamv3.Record {
            // Transform data efficiently
            return enrichRecord(r)
        }).
        Buffer(1000). // Process in batches of 1000
        FlatMap(func(batch []streamv3.Record) streamv3.Stream[streamv3.Record] {
            // Batch processing for efficiency
            return processBatch(batch)
        })

    // Stream results to output without accumulating
    writeStreamToFile(results, "processed_output.csv")
}

func processBatch(batch []streamv3.Record) streamv3.Stream[streamv3.Record] {
    // Batch-optimized processing
    return streamv3.From(batch) // Placeholder
}

func writeStreamToFile(data streamv3.Stream[streamv3.Record], filename string) {
    file, err := os.Create(filename)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    writer := bufio.NewWriter(file)
    defer writer.Flush()

    // Stream write - no memory accumulation
    isFirst := true
    for record := range data {
        if isFirst {
            // Write headers
            writeCSVHeaders(writer, record)
            isFirst = false
        }
        writeCSVRecord(writer, record)
    }
}
```

## Error Handling and Resilience

### Robust Data Processing

Handle errors gracefully in production:

```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Process data with comprehensive error handling
    processDataRobustly()
}

func processDataRobustly() {
    // Use error-aware streams
    dataWithErrors := streamv3.ReadCSVSafe("potentially_corrupt_data.csv")

    validData, errors := separateErrorsFromData(dataWithErrors)

    // Process valid data
    results := validData.
        Map(func(r streamv3.Record) streamv3.Record {
            return validateAndCleanRecord(r)
        }).
        Where(func(r streamv3.Record) bool {
            return r != nil // Filter out records that couldn't be cleaned
        })

    // Log errors for monitoring
    logErrors(errors)

    // Create visualization with error reporting
    createRobustVisualization(results, errors)
}

func separateErrorsFromData(stream streamv3.Stream[streamv3.Record]) (
    data streamv3.Stream[streamv3.Record],
    errors streamv3.Stream[error]) {

    dataChannel := make(chan streamv3.Record)
    errorChannel := make(chan error)

    go func() {
        defer close(dataChannel)
        defer close(errorChannel)

        for item, err := range stream {
            if err != nil {
                errorChannel <- err
            } else {
                dataChannel <- item
            }
        }
    }()

    return streamv3.FromChannel(dataChannel), streamv3.FromChannel(errorChannel)
}

func validateAndCleanRecord(record streamv3.Record) streamv3.Record {
    // Implement data validation and cleaning
    cleaned := streamv3.NewRecord()

    for field, value := range record {
        if cleanedValue := cleanField(field, value); cleanedValue != nil {
            cleaned.Set(field, cleanedValue)
        }
    }

    if len(cleaned.data) > 0 {
        return cleaned.Build()
    }
    return nil // Invalid record
}

func cleanField(field string, value interface{}) interface{} {
    // Implement field-specific cleaning logic
    switch field {
    case "email":
        return validateEmail(value)
    case "age":
        return validateAge(value)
    default:
        return value
    }
}

func createRobustVisualization(data streamv3.Stream[streamv3.Record], errors streamv3.Stream[error]) {
    // Create chart with data quality indicators
    config := streamv3.DefaultChartConfig()
    config.Title = "Data Analysis (with Quality Metrics)"

    // Add data quality summary
    errorCount := streamv3.Count(errors)
    recordCount := streamv3.Count(data)

    fmt.Printf("Data Quality Summary:\n")
    fmt.Printf("  Valid records: %d\n", recordCount)
    fmt.Printf("  Errors: %d\n", errorCount)
    fmt.Printf("  Success rate: %.1f%%\n",
        float64(recordCount)/float64(recordCount+errorCount)*100)

    streamv3.InteractiveChart(data, "robust_analysis.html", config)
}
```

## What's Next?

You've now mastered advanced StreamV3 techniques! You can:

- Handle complex multi-dimensional aggregations
- Process real-time streaming data
- Create sophisticated visualizations
- Optimize for large-scale processing
- Build production-ready error handling

For complete reference documentation, see the [API Reference](api-reference.md).

Ready to build something amazing? StreamV3 scales from simple scripts to production data pipelines. Happy streaming! 🚀