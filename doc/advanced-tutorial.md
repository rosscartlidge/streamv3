# StreamV3 Advanced Tutorial

*Master complex stream processing, real-time analytics, and production-ready patterns*

## Table of Contents

### Documentation Navigation
- **[Getting Started Guide](codelab-intro.md)** - Learn StreamV3 basics step-by-step
- **[API Reference](api-reference.md)** - Complete function reference and examples

### Advanced Topics
- [Prerequisites](#prerequisites)
- [Complex Aggregations](#complex-aggregations)
  - [Multi-Level Grouping](#multi-level-grouping)
  - [Rolling Window Analytics](#rolling-window-analytics)
  - [Statistical Analysis](#statistical-analysis)
- [Stream Joins](#stream-joins)
  - [Inner Join Patterns](#inner-join-patterns)
  - [Left Join with Defaults](#left-join-with-defaults)
  - [Complex Join Conditions](#complex-join-conditions)
- [Real-Time Processing](#real-time-processing)
  - [Live Data Monitoring](#live-data-monitoring)
  - [Infinite Stream Handling](#infinite-stream-handling)
  - [Event-Driven Processing](#event-driven-processing)
  - [Windowing for Infinite Streams](#windowing-for-infinite-streams)
- [Advanced Visualizations](#advanced-visualizations)
  - [Multi-Panel Dashboards](#multi-panel-dashboards)
  - [Time Series Analysis](#time-series-analysis)
  - [Custom Chart Configurations](#custom-chart-configurations)
- [Performance Optimization](#performance-optimization)
  - [Memory-Efficient Processing](#memory-efficient-processing)
  - [Lazy Evaluation Patterns](#lazy-evaluation-patterns)
  - [Parallel Processing](#parallel-processing)
- [Error Handling and Resilience](#error-handling-and-resilience)
  - [Robust Data Processing](#robust-data-processing)
  - [Fault-Tolerant Pipelines](#fault-tolerant-pipelines)
  - [Data Validation](#data-validation)
  - [Mixing Safe and Unsafe Filters](#mixing-safe-and-unsafe-filters)
- [Production Patterns](#production-patterns)
  - [Configuration Management](#configuration-management)
  - [Monitoring and Observability](#monitoring-and-observability)
  - [Testing Strategies](#testing-strategies)
- [What's Next?](#whats-next)

---

## Prerequisites

This tutorial assumes you've completed the [Getting Started Guide](codelab-intro.md) and understand:
- Basic stream operations (`Select`, `Where`, `Limit`)
- Working with Records and the builder pattern
- Reading CSV/JSON data
- Creating simple charts

> 📚 **Need a refresher?** Return to the [Getting Started Guide](codelab-intro.md) for foundational concepts.

---

## Complex Aggregations

### Multi-Level Grouping

Process sales data across multiple dimensions using StreamV3's SQL-style operations:

```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Create complex sales data
    salesData := generateSalesData()

    // Step 1: Group by region and product
    grouped := streamv3.GroupByFields("sales_data", "region", "product")(salesData)

    // Step 2: Apply multiple aggregations
    results := streamv3.Aggregate("sales_data", map[string]streamv3.AggregateFunc{
        "total_revenue": streamv3.Sum("revenue"),
        "avg_deal_size": streamv3.Avg("deal_size"),
        "max_sale":      streamv3.Max[float64]("revenue"),
        "sale_count":    streamv3.Count(),
        "all_customers": streamv3.Collect("customer"),
    })(grouped)

    fmt.Println("Sales Analysis by Region and Product:")
    for result := range results {
        region := result["region"]
        product := result["product"]
        revenue := result["total_revenue"]
        count := result["sale_count"]

        fmt.Printf("  %s %s: $%.0f (%d sales)\n",
            region, product, revenue, count)
    }
}

func generateSalesData() []streamv3.Record {
    regions := []string{"North", "South", "East", "West"}
    products := []string{"Laptop", "Phone", "Tablet"}
    customers := []string{"Alice", "Bob", "Carol", "David", "Eve"}

    var data []streamv3.Record
    for i := 0; i < 100; i++ {
        record := streamv3.MakeMutableRecord().
            String("region", regions[i%len(regions)]).
            String("product", products[i%len(products)]).
            String("customer", customers[i%len(customers)]).
            Float("revenue", 1000 + float64(i*100)).
            Float("deal_size", 500 + float64(i*50)).
            Time("date", time.Now().AddDate(0, 0, -i)).
            Build()
        data = append(data, record)
    }
    return data
}
```

> 📚 **API Reference**: See [GroupBy Operations](api-reference.md#groupby-operations) and [Aggregation Functions](api-reference.md#aggregation-functions) for all available options.

### Rolling Window Analytics

Implement time-based moving averages and trend analysis:

```go
package main

import (
    "fmt"
    "slices"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Generate time series data
    data := generateTimeSeriesData()

    // Apply rolling window analytics
    enriched := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        // This is a simplified example - real implementation would maintain state
        return r
    })(slices.Values(data))

    // Use running aggregations for real-time analysis
    withRunningStats := streamv3.RunningAverage("value", 5)(enriched)
    withRunningSum := streamv3.RunningSum("value")(withRunningStats)

    fmt.Println("Time Series Analysis with Rolling Windows:")
    count := 0
    for record := range withRunningSum {
        if count >= 10 { // Show first 10 for demo
            break
        }

        timestamp := streamv3.GetOr(record, "timestamp", "")
        value := streamv3.GetOr(record, "value", 0.0)
        avg := streamv3.GetOr(record, "running_avg", 0.0)
        sum := streamv3.GetOr(record, "running_sum", 0.0)

        fmt.Printf("  %s: %.2f (avg: %.2f, sum: %.2f)\n",
            timestamp, value, avg, sum)
        count++
    }
}

func generateTimeSeriesData() []streamv3.Record {
    var data []streamv3.Record
    for i := 0; i < 50; i++ {
        record := streamv3.MakeMutableRecord().
            Time("timestamp", time.Now().Add(-time.Duration(i)*time.Hour)).
            Float("value", 100 + float64(i%10)*5).
            String("sensor", fmt.Sprintf("sensor_%d", i%3)).
            Build()
        data = append(data, record)
    }
    return data
}
```

### Statistical Analysis

Implement advanced statistical operations:

```go
package main

import (
    "fmt"
    "math"
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    data := generateNumericData()

    // Calculate statistical metrics using aggregations
    grouped := streamv3.GroupByFields("stats", "category")(slices.Values(data))

    stats := streamv3.Aggregate("stats", map[string]streamv3.AggregateFunc{
        "count":   streamv3.Count(),
        "sum":     streamv3.Sum("value"),
        "avg":     streamv3.Avg("value"),
        "min":     streamv3.Min[float64]("value"),
        "max":     streamv3.Max[float64]("value"),
        "values":  streamv3.Collect("value"),
    })(grouped)

    fmt.Println("Statistical Analysis by Category:")
    for result := range stats {
        category := result["category"]
        count := result["count"]
        avg := result["avg"]
        min := result["min"]
        max := result["max"]
        values := result["values"].([]interface{})

        // Calculate standard deviation
        var variance float64
        avgVal := avg.(float64)
        for _, v := range values {
            val := v.(float64)
            variance += math.Pow(val-avgVal, 2)
        }
        stddev := math.Sqrt(variance / float64(len(values)))

        fmt.Printf("  %s: avg=%.2f, std=%.2f, range=[%.2f, %.2f], n=%d\\n",
            category, avg, stddev, min, max, count)
    }
}

func generateNumericData() []streamv3.Record {
    categories := []string{"A", "B", "C"}
    var data []streamv3.Record

    for i := 0; i < 30; i++ {
        record := streamv3.MakeMutableRecord().
            String("category", categories[i%len(categories)]).
            Float("value", 50 + float64(i)*1.5 + float64(i%5)*10).
            Build()
        data = append(data, record)
    }
    return data
}
```

> 📚 **Reference**: See [Aggregation & Analysis](api-reference.md#aggregation--analysis) for running statistical operations.

---

## Stream Joins

### Inner Join Patterns

Combine datasets using various join strategies:

```go
package main

import (
    "fmt"
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Create user and order datasets
    users := []streamv3.Record{
        streamv3.MakeMutableRecord().String("user_id", "1").String("name", "Alice").String("city", "NYC").Freeze(),
        streamv3.MakeMutableRecord().String("user_id", "2").String("name", "Bob").String("city", "LA").Freeze(),
        streamv3.MakeMutableRecord().String("user_id", "3").String("name", "Carol").String("city", "Chicago").Freeze(),
    }

    orders := []streamv3.Record{
        streamv3.MakeMutableRecord().String("user_id", "1").Float("amount", 100).String("product", "Laptop").Freeze(),
        streamv3.MakeMutableRecord().String("user_id", "2").Float("amount", 50).String("product", "Mouse").Freeze(),
        streamv3.MakeMutableRecord().String("user_id", "1").Float("amount", 200).String("product", "Monitor").Freeze(),
    }

    // Perform inner join on user_id
    joined := streamv3.InnerJoin(
        slices.Values(orders),
        streamv3.OnFields("user_id"),
    )(slices.Values(users))

    fmt.Println("User Orders (Inner Join):")
    for record := range joined {
        name := streamv3.GetOr(record, "name", "Unknown")
        product := streamv3.GetOr(record, "product", "Unknown")
        amount := streamv3.GetOr(record, "amount", 0.0)
        city := streamv3.GetOr(record, "city", "Unknown")

        fmt.Printf("  %s (%s): %s - $%.0f\n", name, city, product, amount)
    }
}
```

### Left Join with Defaults

Handle missing data gracefully:

```go
func demonstrateLeftJoin() {
    users := []streamv3.Record{
        streamv3.MakeMutableRecord().String("user_id", "1").String("name", "Alice").Freeze(),
        streamv3.MakeMutableRecord().String("user_id", "2").String("name", "Bob").Freeze(),
        streamv3.MakeMutableRecord().String("user_id", "3").String("name", "Carol").Freeze(),
    }

    orders := []streamv3.Record{
        streamv3.MakeMutableRecord().String("user_id", "1").Float("amount", 100).Freeze(),
        // Note: No orders for user 2 and 3
    }

    // Left join preserves all users, even those without orders
    leftJoined := streamv3.LeftJoin(
        slices.Values(orders),
        streamv3.OnFields("user_id"),
    )(slices.Values(users))

    fmt.Println("All Users with Orders (Left Join):")
    for record := range leftJoined {
        name := streamv3.GetOr(record, "name", "Unknown")
        amount := streamv3.GetOr(record, "amount", 0.0) // Defaults to 0 for users without orders

        if amount > 0 {
            fmt.Printf("  %s: $%.0f\n", name, amount)
        } else {
            fmt.Printf("  %s: No orders\n", name)
        }
    }
}
```

### Complex Join Conditions

Use custom predicates for sophisticated joins:

```go
func demonstrateComplexJoin() {
    products := []streamv3.Record{
        streamv3.MakeMutableRecord().String("product_id", "1").String("category", "Electronics").Float("price", 100).Freeze(),
        streamv3.MakeMutableRecord().String("product_id", "2").String("category", "Electronics").Float("price", 200).Freeze(),
        streamv3.MakeMutableRecord().String("product_id", "3").String("category", "Books").Float("price", 20).Freeze(),
    }

    sales := []streamv3.Record{
        streamv3.MakeMutableRecord().String("product_id", "1").Int("quantity", 5).String("region", "North").Freeze(),
        streamv3.MakeMutableRecord().String("product_id", "2").Int("quantity", 3).String("region", "South").Freeze(),
    }

    // Custom join condition: match products with sales and calculate revenue
    customJoined := streamv3.InnerJoin(
        slices.Values(sales),
        streamv3.OnCondition(func(product, sale streamv3.Record) bool {
            // Custom logic: only join electronics products
            category := streamv3.GetOr(product, "category", "")
            return category == "Electronics" &&
                   product["product_id"] == sale["product_id"]
        }),
    )(slices.Values(products))

    fmt.Println("Electronics Sales Analysis:")
    for record := range customJoined {
        productId := record["product_id"]
        price := streamv3.GetOr(record, "price", 0.0)
        quantity := streamv3.GetOr(record, "quantity", int64(0))
        region := streamv3.GetOr(record, "region", "Unknown")
        revenue := price * float64(quantity)

        fmt.Printf("  Product %s in %s: $%.0f (qty: %d)\n",
            productId, region, revenue, quantity)
    }
}
```

> 📚 **Reference**: See [Join Operations](api-reference.md#join-operations) for all join types and predicates.

---

## Real-Time Processing

### Live Data Monitoring

Process infinite streams with early termination patterns:

```go
package main

import (
    "fmt"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Simulate infinite sensor data stream
    sensorStream := func(yield func(streamv3.Record) bool) {
        for i := 0; ; i++ {
            record := streamv3.MakeMutableRecord().
                String("sensor_id", fmt.Sprintf("sensor_%d", i%3)).
                Float("temperature", 20 + float64(i%20)).
                Float("humidity", 40 + float64(i%30)).
                Time("timestamp", time.Now()).
                Build()

            if !yield(record) {
                return
            }
            time.Sleep(100 * time.Millisecond) // Simulate real-time data
        }
    }

    // Process with early termination
    limited := streamv3.Limit[streamv3.Record](50)(sensorStream)

    // Apply real-time processing
    processed := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        temp := streamv3.GetOr(r, "temperature", 0.0)

        // Add alert status
        var status string
        if temp > 35 {
            status = "HIGH"
        } else if temp < 15 {
            status = "LOW"
        } else {
            status = "NORMAL"
        }

        return streamv3.SetField(r, "status", status)
    })(limited)

    // Filter alerts
    alerts := streamv3.Where(func(r streamv3.Record) bool {
        status := streamv3.GetOr(r, "status", "")
        return status != "NORMAL"
    })(processed)

    fmt.Println("Real-time Sensor Monitoring (Alerts Only):")
    for alert := range alerts {
        sensorId := streamv3.GetOr(alert, "sensor_id", "")
        temp := streamv3.GetOr(alert, "temperature", 0.0)
        status := streamv3.GetOr(alert, "status", "")

        fmt.Printf("  ALERT %s: %.1f°C (%s)\n", sensorId, temp, status)
    }
}
```

### Infinite Stream Handling

Use timeouts and conditions to control infinite processing:

```go
func demonstrateInfiniteStreamHandling() {
    // Create infinite data generator
    infiniteData := func(yield func(streamv3.Record) bool) {
        for i := 0; ; i++ {
            record := streamv3.MakeMutableRecord().
                String("event_id", fmt.Sprintf("event_%d", i)).
                Float("value", float64(i)).
                Time("timestamp", time.Now()).
                Build()

            if !yield(record) {
                return
            }
            time.Sleep(50 * time.Millisecond)
        }
    }

    // Method 1: Time-based termination
    timedStream := streamv3.Timeout[streamv3.Record](2 * time.Second)(infiniteData)

    // Method 2: Condition-based termination
    conditionalStream := streamv3.TakeWhile(func(r streamv3.Record) bool {
        value := streamv3.GetOr(r, "value", 0.0)
        return value < 100 // Stop when value reaches 100
    })(infiniteData)

    // Method 3: Count-based termination
    countLimited := streamv3.Limit[streamv3.Record](20)(infiniteData)

    fmt.Println("Processing infinite streams:")

    // Process timed stream
    fmt.Println("  Time-limited (2 seconds):")
    start := time.Now()
    count := 0
    for record := range timedStream {
        count++
        if count <= 5 { // Show first 5
            eventId := streamv3.GetOr(record, "event_id", "")
            fmt.Printf("    %s\n", eventId)
        }
    }
    fmt.Printf("    Processed %d events in %v\n", count, time.Since(start))
}
```

### Event-Driven Processing

Implement reactive patterns for event streams:

```go
func demonstrateEventDrivenProcessing() {
    // Simulate event stream
    events := []streamv3.Record{
        streamv3.MakeMutableRecord().String("type", "user_login").String("user", "alice").Time("time", time.Now()).Freeze(),
        streamv3.MakeMutableRecord().String("type", "purchase").String("user", "alice").Float("amount", 100).Freeze(),
        streamv3.MakeMutableRecord().String("type", "user_login").String("user", "bob").Time("time", time.Now()).Freeze(),
        streamv3.MakeMutableRecord().String("type", "error").String("service", "payment").String("message", "timeout").Freeze(),
        streamv3.MakeMutableRecord().String("type", "purchase").String("user", "bob").Float("amount", 50).Freeze(),
    }

    // Process different event types
    purchases := streamv3.Where(func(r streamv3.Record) bool {
        return streamv3.GetOr(r, "type", "") == "purchase"
    })(slices.Values(events))

    errors := streamv3.Where(func(r streamv3.Record) bool {
        return streamv3.GetOr(r, "type", "") == "error"
    })(slices.Values(events))

    fmt.Println("Event-Driven Processing:")

    fmt.Println("  Purchases:")
    for purchase := range purchases {
        user := streamv3.GetOr(purchase, "user", "")
        amount := streamv3.GetOr(purchase, "amount", 0.0)
        fmt.Printf("    %s purchased $%.0f\n", user, amount)
    }

    fmt.Println("  Errors:")
    for error := range errors {
        service := streamv3.GetOr(error, "service", "")
        message := streamv3.GetOr(error, "message", "")
        fmt.Printf("    %s error: %s\n", service, message)
    }
}
```

> 📚 **Reference**: See [Early Termination](api-reference.md#early-termination) for infinite stream control patterns.

### Windowing for Infinite Streams

Handle infinite data streams with sophisticated windowing patterns that enable real-time analytics and bounded memory usage:

#### Count-Based Windows

Process data in fixed-size chunks:

```go
package main

import (
    "fmt"
    "time"
    "slices"
    "github.com/rosscartlidge/streamv3"
)

func demonstrateCountWindows() {
    // Simulate infinite sensor readings
    sensorData := func(yield func(streamv3.Record) bool) {
        for i := 0; ; i++ {
            record := streamv3.MakeMutableRecord().
                String("sensor_id", fmt.Sprintf("temp_sensor_%d", i%3)).
                Float("temperature", 20 + float64(i%15) + (float64(i)*0.1)).
                Time("timestamp", time.Now().Add(time.Duration(i)*time.Second)).
                Build()

            if !yield(record) {
                return
            }
        }
    }

    // Process in batches of 10 readings
    windowed := streamv3.CountWindow[streamv3.Record](10)(
        streamv3.Limit[streamv3.Record](50)(sensorData), // Limit for demo
    )

    fmt.Println("Count-Based Window Analysis (batches of 10):")
    batchNum := 1
    for batch := range windowed {
        if len(batch) == 0 {
            continue
        }

        // Calculate batch statistics
        var sum, min, max float64
        min = 1000 // Initialize to high value
        max = -1000 // Initialize to low value

        for _, record := range batch {
            temp := streamv3.GetOr(record, "temperature", 0.0)
            sum += temp
            if temp < min {
                min = temp
            }
            if temp > max {
                max = temp
            }
        }

        avg := sum / float64(len(batch))
        fmt.Printf("  Batch %d: Avg=%.1f°C, Min=%.1f°C, Max=%.1f°C (%d readings)\n",
            batchNum, avg, min, max, len(batch))
        batchNum++
    }
}
```

#### Time-Based Windows

Group data by time intervals for temporal analysis:

```go
func demonstrateTimeWindows() {
    // Create timestamped financial data
    priceStream := func(yield func(streamv3.Record) bool) {
        baseTime := time.Now()
        basePrice := 100.0

        for i := 0; i < 60; i++ { // 60 seconds of data
            // Simulate price fluctuations
            price := basePrice + (float64(i%10)-5)*2 + float64(i)*0.1

            record := streamv3.MakeMutableRecord().
                String("symbol", "AAPL").
                Float("price", price).
                Time("timestamp", baseTime.Add(time.Duration(i)*time.Second)).
                Build()

            if !yield(record) {
                return
            }
        }
    }

    // Group into 10-second windows
    timeWindowed := streamv3.TimeWindow[streamv3.Record](
        10*time.Second,
        "timestamp",
    )(priceStream)

    fmt.Println("Time-Based Window Analysis (10-second intervals):")
    windowNum := 1
    for window := range timeWindowed {
        if len(window) == 0 {
            continue
        }

        // Calculate OHLC (Open, High, Low, Close) for each window
        var prices []float64
        for _, record := range window {
            price := streamv3.GetOr(record, "price", 0.0)
            prices = append(prices, price)
        }

        if len(prices) > 0 {
            open := prices[0]
            close := prices[len(prices)-1]
            high := slices.Max(prices)
            low := slices.Min(prices)

            fmt.Printf("  Window %d: O=%.2f H=%.2f L=%.2f C=%.2f (%d ticks)\n",
                windowNum, open, high, low, close, len(prices))
        }
        windowNum++
    }
}
```

#### Sliding Windows for Continuous Analytics

Use sliding windows for smooth, overlapping analysis:

```go
func demonstrateSlidingWindows() {
    // Generate continuous data stream
    dataStream := func(yield func(streamv3.Record) bool) {
        for i := 0; i < 20; i++ {
            // Simulate network latency measurements
            latency := 50 + float64(i%8) + float64(i)*0.5

            record := streamv3.MakeMutableRecord().
                String("server", fmt.Sprintf("srv_%d", i%3)).
                Float("latency_ms", latency).
                Int("sequence", i).
                Time("timestamp", time.Now().Add(time.Duration(i)*100*time.Millisecond)).
                Build()

            if !yield(record) {
                return
            }
        }
    }

    // Create sliding window: size=5, step=2
    sliding := streamv3.SlidingCountWindow[streamv3.Record](5, 2)(dataStream)

    fmt.Println("Sliding Window Analysis (window=5, step=2):")
    windowNum := 1
    for window := range sliding {
        if len(window) == 0 {
            continue
        }

        // Calculate moving average latency
        var totalLatency float64
        var sequences []int

        for _, record := range window {
            latency := streamv3.GetOr(record, "latency_ms", 0.0)
            seq := streamv3.GetOr(record, "sequence", 0)
            totalLatency += latency
            sequences = append(sequences, seq)
        }

        avgLatency := totalLatency / float64(len(window))
        fmt.Printf("  Window %d [seq %d-%d]: Avg Latency=%.1fms (%d samples)\n",
            windowNum, sequences[0], sequences[len(sequences)-1], avgLatency, len(window))
        windowNum++
    }
}
```

#### Real-Time Aggregation with Windows

Combine windowing with aggregation for continuous insights:

```go
func demonstrateWindowedAggregation() {
    // Simulate IoT sensor network
    iotStream := func(yield func(streamv3.Record) bool) {
        sensors := []string{"temp_01", "temp_02", "temp_03", "humidity_01", "humidity_02"}

        for i := 0; i < 100; i++ {
            sensorId := sensors[i%len(sensors)]
            var value float64
            var unit string

            if sensorId[:4] == "temp" {
                value = 18 + float64(i%20) + float64(i)*0.05 // Temperature trend
                unit = "°C"
            } else {
                value = 30 + float64(i%40) + float64(i)*0.1 // Humidity trend
                unit = "%"
            }

            record := streamv3.MakeMutableRecord().
                String("sensor_id", sensorId).
                String("metric_type", sensorId[:4]).
                Float("value", value).
                String("unit", unit).
                Time("timestamp", time.Now().Add(time.Duration(i)*500*time.Millisecond)).
                Build()

            if !yield(record) {
                return
            }
        }
    }

    // Window by count and group by metric type
    windowed := streamv3.CountWindow[streamv3.Record](15)(iotStream)

    fmt.Println("Real-Time IoT Analytics (15-reading windows):")
    windowNum := 1
    for window := range windowed {
        if len(window) == 0 {
            continue
        }

        // Group by metric type within each window
        windowStream := slices.Values(window)
        grouped := streamv3.GroupByFields("sensors", "metric_type")(windowStream)

        fmt.Printf("  Window %d Analysis:\n", windowNum)

        for groupRecord := range grouped {
            // Extract grouped data
            groupKey := streamv3.GetOr(groupRecord, "GroupKey", "")
            groupValue := streamv3.GetOr(groupRecord, "GroupValue", "")
            sensors, _ := streamv3.Get[[]streamv3.Record](groupRecord, "sensors")

            if len(sensors) == 0 {
                continue
            }

            // Calculate statistics for this group
            var values []float64
            sensorCount := make(map[string]int)
            unit := ""

            for _, sensor := range sensors {
                value := streamv3.GetOr(sensor, "value", 0.0)
                sensorId := streamv3.GetOr(sensor, "sensor_id", "")
                sensorUnit := streamv3.GetOr(sensor, "unit", "")

                values = append(values, value)
                sensorCount[sensorId]++
                if unit == "" {
                    unit = sensorUnit
                }
            }

            if len(values) > 0 {
                sum := 0.0
                for _, v := range values {
                    sum += v
                }
                avg := sum / float64(len(values))
                min := slices.Min(values)
                max := slices.Max(values)

                fmt.Printf("    %s: Avg=%.1f%s, Range=[%.1f-%.1f]%s (%d readings from %d sensors)\n",
                    groupValue, avg, unit, min, max, unit, len(values), len(sensorCount))
            }
        }
        windowNum++
    }
}
```

#### Session Windows with Timeout

Handle irregular data streams with session-based windowing:

```go
func demonstrateSessionWindows() {
    // Simulate user activity with gaps
    userEvents := []streamv3.Record{
        streamv3.MakeMutableRecord().String("user", "alice").String("action", "login").Time("time", time.Now()).Freeze(),
        streamv3.MakeMutableRecord().String("user", "alice").String("action", "browse").Time("time", time.Now().Add(30*time.Second)).Freeze(),
        streamv3.MakeMutableRecord().String("user", "alice").String("action", "purchase").Time("time", time.Now().Add(45*time.Second)).Freeze(),
        // Gap of 10 minutes - new session
        streamv3.MakeMutableRecord().String("user", "alice").String("action", "login").Time("time", time.Now().Add(10*time.Minute)).Freeze(),
        streamv3.MakeMutableRecord().String("user", "alice").String("action", "browse").Time("time", time.Now().Add(10*time.Minute+20*time.Second)).Freeze(),
    }

    // Group events into sessions using time-based timeout
    sessionTimeout := 5 * time.Minute
    sessions := streamv3.TimeBasedTimeout("time", sessionTimeout)(slices.Values(userEvents))

    fmt.Println("Session-Based Window Analysis:")
    sessionNum := 1
    var currentSession []streamv3.Record

    for event := range sessions {
        currentSession = append(currentSession, event)

        // Check if this might be the end of a session
        // (In a real implementation, you'd have more sophisticated session detection)
        user := streamv3.GetOr(event, "user", "")
        action := streamv3.GetOr(event, "action", "")

        fmt.Printf("  Session %d - %s: %s\n", sessionNum, user, action)

        if action == "purchase" {
            // End of purchase session
            fmt.Printf("    Session completed with %d actions\n", len(currentSession))
            currentSession = nil
            sessionNum++
        }
    }
}
```

> 💡 **Pro Tip**: Combine windowing with early termination patterns (`TakeWhile`, `Timeout`) to create robust infinite stream processors that gracefully handle memory constraints and processing limits.

> 📚 **Reference**: See [Window Operations](api-reference.md#window-operations) for complete windowing API documentation.

---

## Advanced Visualizations

### Multi-Panel Dashboards

Create comprehensive dashboards with multiple chart types:

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Generate comprehensive business data
    salesData := generateBusinessData()

    // Create multiple visualizations
    createSalesDashboard(salesData)
    createTrendAnalysis(salesData)
    createRegionalComparison(salesData)

    fmt.Println("📊 Multi-panel dashboard created:")
    fmt.Println("  • sales_overview.html - Main dashboard")
    fmt.Println("  • trends_analysis.html - Time series trends")
    fmt.Println("  • regional_comparison.html - Geographic analysis")
}

func createSalesDashboard(data []streamv3.Record) {
    config := streamv3.DefaultChartConfig()
    config.Title = "Sales Overview Dashboard"
    config.ChartType = "bar"
    config.Width = 1400
    config.Height = 700
    config.Theme = "light"
    config.ColorScheme = "vibrant"

    err := streamv3.InteractiveChart(
        streamv3.From(data),
        "sales_overview.html",
        config,
    )
    if err != nil {
        fmt.Printf("Error creating sales dashboard: %v\n", err)
    }
}

func createTrendAnalysis(data []streamv3.Record) {
    config := streamv3.DefaultChartConfig()
    config.Title = "Sales Trends Over Time"
    config.ChartType = "line"
    config.Theme = "dark"
    config.EnableZoom = true
    config.ShowDataLabels = true

    err := streamv3.TimeSeriesChart(
        streamv3.From(data),
        "date",
        []string{"revenue", "units_sold"},
        "trends_analysis.html",
        config,
    )
    if err != nil {
        fmt.Printf("Error creating trends chart: %v\n", err)
    }
}

func createRegionalComparison(data []streamv3.Record) {
    config := streamv3.DefaultChartConfig()
    config.Title = "Regional Performance Comparison"
    config.ChartType = "scatter"
    config.ColorScheme = "pastel"

    err := streamv3.InteractiveChart(
        streamv3.From(data),
        "regional_comparison.html",
        config,
    )
    if err != nil {
        fmt.Printf("Error creating regional chart: %v\n", err)
    }
}

func generateBusinessData() []streamv3.Record {
    regions := []string{"North", "South", "East", "West"}
    products := []string{"Software", "Hardware", "Services"}

    var data []streamv3.Record
    for i := 0; i < 200; i++ {
        record := streamv3.MakeMutableRecord().
            String("region", regions[rand.Intn(len(regions))]).
            String("product", products[rand.Intn(len(products))]).
            Float("revenue", 10000 + rand.Float64()*90000).
            Int("units_sold", int64(10 + rand.Intn(100))).
            String("quarter", fmt.Sprintf("Q%d", (i%12)/3 + 1)).
            Time("date", time.Now().AddDate(0, 0, -i)).
            Build()
        data = append(data, record)
    }
    return data
}
```

### Time Series Analysis

Implement sophisticated time series visualizations:

```go
func createAdvancedTimeSeriesChart() {
    // Generate time series with multiple metrics
    timeSeriesData := generateTimeSeriesData()

    // Create comprehensive time series configuration
    config := streamv3.DefaultChartConfig()
    config.Title = "Multi-Metric Time Series Analysis"
    config.Width = 1600
    config.Height = 800
    config.ChartType = "line"
    config.Theme = "dark"

    // Advanced time series options
    config.ShowLegend = true
    config.EnableZoom = true
    config.EnablePan = true
    config.ShowDataLabels = false
    config.EnableAnimations = true

    err := streamv3.TimeSeriesChart(
        streamv3.From(timeSeriesData),
        "timestamp",
        []string{"cpu_usage", "memory_usage", "network_io", "disk_io"},
        "system_metrics_timeline.html",
        config,
    )
    if err != nil {
        fmt.Printf("Error creating time series chart: %v\n", err)
        return
    }

    fmt.Println("📈 Advanced time series chart created: system_metrics_timeline.html")
    fmt.Println("Features:")
    fmt.Println("  • Multi-metric overlay")
    fmt.Println("  • Interactive zoom and pan")
    fmt.Println("  • Dark theme optimized for monitoring")
    fmt.Println("  • Real-time data visualization patterns")
}
```

### Custom Chart Configurations

Leverage advanced chart customization:

```go
func demonstrateCustomCharts() {
    data := generateSampleData()

    // Highly customized chart configuration
    config := streamv3.DefaultChartConfig()
    config.Title = "Custom Business Intelligence Dashboard"
    config.Width = 1800
    config.Height = 1000
    config.ChartType = "bar"

    // Theme and colors
    config.Theme = "light"
    config.ColorScheme = "monochrome"

    // Interactivity
    config.EnableZoom = true
    config.EnablePan = true
    config.ShowTooltips = true
    config.ShowLegend = true
    config.ShowDataLabels = true

    // Advanced features
    config.EnableInteractive = true
    config.EnableCalculations = true
    config.ExportFormats = []string{"png", "csv", "json"}

    // Custom styling (if supported)
    config.CustomCSS = `
        .chart-title { font-size: 24px; font-weight: bold; }
        .legend { font-size: 14px; }
        .tooltip { background: rgba(0,0,0,0.9); color: white; }
    `

    err := streamv3.InteractiveChart(streamv3.From(data), "custom_dashboard.html", config)
    if err != nil {
        fmt.Printf("Error creating custom chart: %v\n", err)
        return
    }

    fmt.Println("🎨 Custom dashboard created with advanced features")
}
```

> 📚 **Reference**: See [Chart & Visualization](api-reference.md#chart--visualization) for all configuration options.

---

## Performance Optimization

### Memory-Efficient Processing

Handle large datasets without loading everything into memory:

```go
package main

import (
    "fmt"
    "iter"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    demonstrateMemoryEfficientProcessing()
}

func demonstrateMemoryEfficientProcessing() {
    // Process large CSV files efficiently (panics if file doesn't exist)
    // In production, you'd wrap this in a defer/recover for graceful degradation
    data := streamv3.ReadCSV("large_dataset.csv")

    // Chain operations without materializing intermediate results
    pipeline := func(stream iter.Seq[streamv3.Record]) iter.Seq[streamv3.Record] {
        // Filter early to reduce data volume
        filtered := streamv3.Where(func(r streamv3.Record) bool {
            value := streamv3.GetOr(r, "amount", 0.0)
            return value > 1000
        })(stream)

        // Transform only what's needed
        transformed := streamv3.Select(func(r streamv3.Record) streamv3.Record {
            amount := streamv3.GetOr(r, "amount", 0.0)
            return streamv3.SetField(r, "category", categorizeAmount(amount))
        })(filtered)

        // Limit results to control memory usage
        return streamv3.Limit[streamv3.Record](1000)(transformed)
    }

    // Process in chunks using windows
    windowed := streamv3.CountWindow[streamv3.Record](100)(pipeline(data))

    fmt.Println("Memory-efficient processing with windows:")
    windowCount := 0
    for window := range windowed {
        windowCount++
        fmt.Printf("  Processed window %d: %d records\n", windowCount, len(window))

        if windowCount >= 5 { // Limit for demo
            break
        }
    }
}

func categorizeAmount(amount float64) string {
    if amount > 10000 {
        return "enterprise"
    } else if amount > 5000 {
        return "business"
    } else {
        return "standard"
    }
}
```

### Lazy Evaluation Patterns

Leverage Go's iterator patterns for efficient stream processing:

```go
func demonstrateLazyEvaluation() {
    // Create a large sequence without materializing it
    largeSequence := func(yield func(int) bool) {
        for i := 0; i < 1000000; i++ {
            if !yield(i) {
                return // Early termination saves computation
            }
        }
    }

    // Chain multiple lazy operations
    evens := streamv3.Where(func(x int) bool { return x%2 == 0 })(largeSequence)
    squares := streamv3.Select(func(x int) int { return x * x })(evens)
    limited := streamv3.Limit[int](10)(squares)

    fmt.Println("Lazy evaluation - only first 10 even squares:")
    for value := range limited {
        fmt.Printf("  %d\n", value)
    }
    // Note: Only computed 20 values from the million-element sequence!
}
```

### Parallel Processing

Use Tee for parallel processing pipelines:

```go
func demonstrateParallelProcessing() {
    data := generateLargeDataset()

    // Split stream for parallel processing
    streams := streamv3.Tee(slices.Values(data), 3)

    // Process different aspects in parallel
    go processForAnalytics(streams[0])
    go processForReporting(streams[1])
    go processForArchiving(streams[2])

    fmt.Println("Parallel processing started...")
    time.Sleep(2 * time.Second) // Wait for demo completion
}

func processForAnalytics(stream iter.Seq[streamv3.Record]) {
    // Analytics-specific processing
    filtered := streamv3.Where(func(r streamv3.Record) bool {
        return streamv3.GetOr(r, "type", "") == "analytics"
    })(stream)

    count := 0
    for range filtered {
        count++
    }
    fmt.Printf("  Analytics processing completed: %d records\n", count)
}

func processForReporting(stream iter.Seq[streamv3.Record]) {
    // Reporting-specific processing
    count := 0
    for range stream {
        count++
        if count%100 == 0 {
            // Simulate batch processing
            time.Sleep(10 * time.Millisecond)
        }
    }
    fmt.Printf("  Reporting processing completed: %d records\n", count)
}

func processForArchiving(stream iter.Seq[streamv3.Record]) {
    // Archiving-specific processing
    count := 0
    for range stream {
        count++
    }
    fmt.Printf("  Archiving processing completed: %d records\n", count)
}
```

> 📚 **Reference**: See [Utility Operations](api-reference.md#utility-operations) for Tee and parallel processing patterns.

---

## Error Handling and Resilience

### Robust Data Processing

Implement comprehensive error handling strategies:

```go
package main

import (
    "fmt"
    "iter"
    "log"
    "slices"
    "strconv"
    "strings"
    "time"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    demonstrateRobustProcessing()
}

func demonstrateRobustProcessing() {
    // Data with potential errors
    rawData := []string{"123", "456", "invalid", "789", "", "101"}

    // Safe processing with error handling
    safeResults := streamv3.SelectSafe(func(s string) (int, error) {
        if s == "" {
            return 0, fmt.Errorf("empty string")
        }
        return strconv.Atoi(s)
    })(slices.Values(rawData))

    fmt.Println("Robust processing with error handling:")
    validCount := 0
    errorCount := 0

    for value, err := range safeResults {
        if err != nil {
            errorCount++
            log.Printf("  Error processing value: %v", err)
            continue
        }

        validCount++
        fmt.Printf("  Valid: %d\n", value)
    }

    fmt.Printf("Summary: %d valid, %d errors\\n", validCount, errorCount)
}
```

### Fault-Tolerant Pipelines

Build resilient processing pipelines:

```go
func demonstrateFaultTolerantPipeline() {
    // Simulate unreliable data source
    unreliableData := []streamv3.Record{
        streamv3.MakeMutableRecord().String("id", "1").String("value", "100").Freeze(),
        streamv3.MakeMutableRecord().String("id", "2").String("value", "invalid").Freeze(),
        streamv3.MakeMutableRecord().String("id", "3").String("value", "300").Freeze(),
        streamv3.MakeMutableRecord().Freeze(), // Missing fields
    }

    // Fault-tolerant processing
    processed := streamv3.SelectSafe(func(r streamv3.Record) (streamv3.Record, error) {
        // Validate required fields
        id := streamv3.GetOr(r, "id", "")
        if id == "" {
            return streamv3.Record{}, fmt.Errorf("missing id field")
        }

        valueStr := streamv3.GetOr(r, "value", "")
        value, err := strconv.ParseFloat(valueStr, 64)
        if err != nil {
            return streamv3.Record{}, fmt.Errorf("invalid value: %s", valueStr)
        }

        // Create validated record
        result := streamv3.MakeMutableRecord().
            String("id", id).
            Float("value", value).
            String("status", "processed").
            Build()

        return result, nil
    })(slices.Values(unreliableData))

    fmt.Println("Fault-tolerant pipeline results:")
    for record, err := range processed {
        if err != nil {
            fmt.Printf("  ERROR: %v\n", err)
            continue
        }

        id := streamv3.GetOr(record, "id", "")
        value := streamv3.GetOr(record, "value", 0.0)
        fmt.Printf("  SUCCESS: %s = %.0f\n", id, value)
    }
}
```

### Data Validation

Implement comprehensive data validation:

```go
func demonstrateDataValidation() {
    // Sample data with validation issues
    customerData := []streamv3.Record{
        streamv3.MakeMutableRecord().String("email", "alice@example.com").Int("age", 30).Float("score", 95.5).Freeze(),
        streamv3.MakeMutableRecord().String("email", "invalid-email").Int("age", -5).Float("score", 150.0).Freeze(),
        streamv3.MakeMutableRecord().String("email", "bob@example.com").Int("age", 25).Float("score", 87.2).Freeze(),
    }

    // Validation pipeline
    validated := streamv3.SelectSafe(func(r streamv3.Record) (streamv3.Record, error) {
        var errors []string

        // Email validation
        email := streamv3.GetOr(r, "email", "")
        if !isValidEmail(email) {
            errors = append(errors, "invalid email")
        }

        // Age validation
        age := streamv3.GetOr(r, "age", int64(0))
        if age < 0 || age > 150 {
            errors = append(errors, "invalid age")
        }

        // Score validation
        score := streamv3.GetOr(r, "score", 0.0)
        if score < 0 || score > 100 {
            errors = append(errors, "invalid score")
        }

        if len(errors) > 0 {
            return streamv3.Record{}, fmt.Errorf("validation failed: %v", errors)
        }

        // Add validation timestamp
        result := streamv3.SetField(r, "validated_at", time.Now())
        return result, nil
    })(slices.Values(customerData))

    fmt.Println("Data validation results:")
    for record, err := range validated {
        if err != nil {
            fmt.Printf("  INVALID: %v\n", err)
            continue
        }

        email := streamv3.GetOr(record, "email", "")
        age := streamv3.GetOr(record, "age", int64(0))
        score := streamv3.GetOr(record, "score", 0.0)
        fmt.Printf("  VALID: %s (age: %d, score: %.1f)\n", email, age, score)
    }
}

func isValidEmail(email string) bool {
    // Simplified email validation
    return len(email) > 0 &&
           strings.Contains(email, "@") &&
           strings.Contains(email, ".")
}
```

### Mixing Safe and Unsafe Filters

StreamV3 provides three conversion utilities to seamlessly bridge between error-aware (`iter.Seq2[T, error]`) and simple (`iter.Seq[T]`) iterators:

#### Conversion Utilities

```go
// Safe - Convert simple to error-aware (never errors)
func Safe[T any](seq iter.Seq[T]) iter.Seq2[T, error]

// Unsafe - Convert error-aware to simple (panics on errors)
func Unsafe[T any](seq iter.Seq2[T, error]) iter.Seq[T]

// IgnoreErrors - Convert error-aware to simple (skips errors)
func IgnoreErrors[T any](seq iter.Seq2[T, error]) iter.Seq[T]
```

#### Pattern 1: Start Normal, Add Error Handling, Continue Normal

Use `Safe()` to enter error-aware processing and `IgnoreErrors()` to exit gracefully:

```go
func demonstrateMixedPipeline() {
    // Start with normal data
    transactions := streamv3.From([]streamv3.Record{
        streamv3.MakeMutableRecord().String("id", "TX001").String("amount_str", "100.50").Freeze(),
        streamv3.MakeMutableRecord().String("id", "TX002").String("amount_str", "invalid").Freeze(),
        streamv3.MakeMutableRecord().String("id", "TX003").String("amount_str", "250.75").Freeze(),
    })

    // Apply normal filter
    filtered := streamv3.Where(func(r streamv3.Record) bool {
        return streamv3.GetOr(r, "id", "") != ""
    })(transactions)

    // Convert to Safe for error-prone parsing
    safeStream := streamv3.Safe(filtered)

    // Use Safe filter for parsing
    parsed := streamv3.SelectSafe(func(r streamv3.Record) (streamv3.Record, error) {
        amountStr := streamv3.GetOr(r, "amount_str", "")
        amount, err := strconv.ParseFloat(amountStr, 64)
        if err != nil {
            return streamv3.Record{}, fmt.Errorf("invalid amount: %s", amountStr)
        }

        return streamv3.MakeMutableRecord().
            String("id", streamv3.GetOr(r, "id", "")).
            Float("amount", amount).
            Build(), nil
    })(safeStream)

    // Convert back to normal, ignoring errors
    cleanData := streamv3.IgnoreErrors(parsed)

    // Continue with normal filters
    final := streamv3.Chain(
        streamv3.Where(func(r streamv3.Record) bool {
            amount := streamv3.GetOr(r, "amount", 0.0)
            return amount > 100.0
        }),
        streamv3.Limit[streamv3.Record](10),
    )(cleanData)

    fmt.Println("Mixed pipeline results:")
    for record := range final {
        id := streamv3.GetOr(record, "id", "")
        amount := streamv3.GetOr(record, "amount", 0.0)
        fmt.Printf("  %s: $%.2f\n", id, amount)
    }
}
```

#### Pattern 2: I/O with Safe, Processing with Normal

Read data with error handling, then process confidently:

```go
func demonstrateIOProcessing() {
    // Read CSV with error awareness
    safeData := streamv3.ReadCSVSafe("transactions.csv")

    // Validate data with Safe filter
    validated := streamv3.WhereSafe(func(r streamv3.Record) (bool, error) {
        amount := streamv3.GetOr(r, "amount", 0.0)
        if amount < 0 {
            return false, fmt.Errorf("negative amount: %.2f", amount)
        }
        return true, nil
    })(safeData)

    // Convert to normal - we're confident after validation
    cleanData := streamv3.IgnoreErrors(validated)

    // Process with normal filters (faster, no error checking)
    processed := streamv3.Chain(
        streamv3.Select(func(r streamv3.Record) streamv3.Record {
            amount := streamv3.GetOr(r, "amount", 0.0)
            return r.Set("tax", amount*0.08)
        }),
        streamv3.Where(func(r streamv3.Record) bool {
            return streamv3.GetOr(r, "tax", 0.0) > 10.0
        }),
    )(cleanData)

    count := 0
    for record := range processed {
        count++
        amount := streamv3.GetOr(record, "amount", 0.0)
        tax := streamv3.GetOr(record, "tax", 0.0)
        fmt.Printf("  Amount: $%.2f, Tax: $%.2f\n", amount, tax)
    }
    fmt.Printf("Processed %d valid transactions\n", count)
}
```

#### Pattern 3: Fail-Fast with Unsafe

Use `Unsafe()` when errors should halt processing:

```go
func demonstrateFailFast() {
    // Read data with error awareness
    safeData := streamv3.ReadCSVSafe("critical_data.csv")

    // Validate strictly
    validated := streamv3.SelectSafe(func(r streamv3.Record) (streamv3.Record, error) {
        // Strict validation - any error is critical
        id := streamv3.GetOr(r, "id", "")
        if id == "" {
            return streamv3.Record{}, fmt.Errorf("missing required field: id")
        }

        amount := streamv3.GetOr(r, "amount", 0.0)
        if amount <= 0 {
            return streamv3.Record{}, fmt.Errorf("invalid amount: %.2f", amount)
        }

        return r, nil
    })(safeData)

    // Convert to Unsafe - panic on any error (fail-fast)
    criticalData := streamv3.Unsafe(validated)

    // Process normally - we know data is valid or we've panicked
    final := streamv3.Limit[streamv3.Record](100)(criticalData)

    fmt.Println("Critical data processing (fail-fast mode):")
    for record := range final {
        id := streamv3.GetOr(record, "id", "")
        amount := streamv3.GetOr(record, "amount", 0.0)
        fmt.Printf("  %s: $%.2f\n", id, amount)
    }
}
```

#### Pattern 4: Best-Effort Processing

Use `IgnoreErrors()` for resilient pipelines that process what they can:

```go
func demonstrateBestEffort() {
    // Multiple data sources with varying quality
    sources := []string{"data1.csv", "data2.csv", "data3.csv"}

    var allRecords []streamv3.Record

    for _, source := range sources {
        // Read with error handling
        safeData := streamv3.ReadCSVSafe(source)

        // Parse and validate - may have errors
        processed := streamv3.SelectSafe(func(r streamv3.Record) (streamv3.Record, error) {
            // Attempt to parse and enrich
            value := streamv3.GetOr(r, "value", "")
            parsed, err := strconv.ParseFloat(value, 64)
            if err != nil {
                return streamv3.Record{}, fmt.Errorf("parse error: %v", err)
            }

            return r.Set("parsed_value", parsed), nil
        })(safeData)

        // Ignore errors - process what we can
        validData := streamv3.IgnoreErrors(processed)

        // Collect valid records
        for record := range validData {
            allRecords = append(allRecords, record)
        }
    }

    fmt.Printf("Best-effort processing: collected %d valid records from %d sources\n",
        len(allRecords), len(sources))
}
```

#### When to Use Each Conversion

**Use `Safe()`:**
- When entering error-aware processing from normal data
- Before I/O operations that might fail
- When starting validation pipelines

**Use `Unsafe()`:**
- After validation when errors shouldn't occur
- In fail-fast scenarios where errors are critical
- When you want the process to terminate on any error

**Use `IgnoreErrors()`:**
- For best-effort processing of messy data
- When you want to continue despite errors
- For resilient ETL pipelines that skip bad records
- When collecting statistics (e.g., "processed 900 of 1000 records")

> 💡 **Pro Tip**: Mix freely! Normal → Safe → Normal → Safe is perfectly valid. The conversions are zero-cost wrappers.

> 📚 **Reference**: See [Error Handling](api-reference.md#error-handling) for safe operation patterns.

---

## Production Patterns

### Configuration Management

Structure configuration for production deployments:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "github.com/rosscartlidge/streamv3"
)

type Config struct {
    DataSources struct {
        CSVPath    string `json:"csv_path"`
        APIUrl     string `json:"api_url"`
        DBConn     string `json:"db_connection"`
    } `json:"data_sources"`

    Processing struct {
        BatchSize    int     `json:"batch_size"`
        Timeout      string  `json:"timeout"`
        MaxErrors    int     `json:"max_errors"`
        RetryCount   int     `json:"retry_count"`
    } `json:"processing"`

    Output struct {
        ChartDir     string `json:"chart_directory"`
        ReportDir    string `json:"report_directory"`
        LogLevel     string `json:"log_level"`
    } `json:"output"`
}

func main() {
    config, err := loadConfig("config.json")
    if err != nil {
        fmt.Printf("Using default configuration: %v\n", err)
        config = getDefaultConfig()
    }

    fmt.Printf("Production processing with config: %+v\n", config)
    runProductionPipeline(config)
}

func loadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config Config
    err = json.Unmarshal(data, &config)
    return &config, err
}

func getDefaultConfig() *Config {
    return &Config{
        DataSources: struct {
            CSVPath string `json:"csv_path"`
            APIUrl  string `json:"api_url"`
            DBConn  string `json:"db_connection"`
        }{
            CSVPath: "data/input.csv",
            APIUrl:  "https://api.example.com/data",
            DBConn:  "postgres://localhost/db",
        },
        Processing: struct {
            BatchSize  int    `json:"batch_size"`
            Timeout    string `json:"timeout"`
            MaxErrors  int    `json:"max_errors"`
            RetryCount int    `json:"retry_count"`
        }{
            BatchSize:  1000,
            Timeout:    "30s",
            MaxErrors:  100,
            RetryCount: 3,
        },
        Output: struct {
            ChartDir  string `json:"chart_directory"`
            ReportDir string `json:"report_directory"`
            LogLevel  string `json:"log_level"`
        }{
            ChartDir:  "output/charts",
            ReportDir: "output/reports",
            LogLevel:  "info",
        },
    }
}

func runProductionPipeline(config *Config) {
    // Implement production pipeline with configuration
    fmt.Println("Running production pipeline...")

    // Example: Use config for batch processing
    if config.Processing.BatchSize > 0 {
        fmt.Printf("  Using batch size: %d\n", config.Processing.BatchSize)
    }
}
```

### Monitoring and Observability

Add metrics and monitoring to stream processing:

```go
func demonstrateMonitoring() {
    // Metrics collection
    type Metrics struct {
        RecordsProcessed int64
        ErrorsEncountered int64
        ProcessingTimeMs int64
        MemoryUsageMB   float64
    }

    metrics := &Metrics{}
    startTime := time.Now()

    // Sample data processing with metrics
    data := generateSampleData()

    processed := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        metrics.RecordsProcessed++

        // Simulate processing work
        time.Sleep(1 * time.Millisecond)

        // Add processing metadata
        result := streamv3.SetField(r, "processed_at", time.Now())
        return result
    })(slices.Values(data))

    // Collect results and update metrics
    var results []streamv3.Record
    for record := range processed {
        results = append(results, record)
    }

    metrics.ProcessingTimeMs = time.Since(startTime).Milliseconds()

    // Report metrics
    fmt.Println("Processing Metrics:")
    fmt.Printf("  Records Processed: %d\n", metrics.RecordsProcessed)
    fmt.Printf("  Processing Time: %dms\n", metrics.ProcessingTimeMs)
    fmt.Printf("  Throughput: %.2f records/sec\\n",
        float64(metrics.RecordsProcessed)/float64(metrics.ProcessingTimeMs)*1000)
}
```

### Testing Strategies

Implement comprehensive testing for stream operations:

```go
package main

import (
    "fmt"
    "slices"
    "testing"
    "github.com/rosscartlidge/streamv3"
)

func TestStreamProcessing(t *testing.T) {
    // Test data
    input := []streamv3.Record{
        streamv3.MakeMutableRecord().String("category", "A").Float("value", 100).Freeze(),
        streamv3.MakeMutableRecord().String("category", "B").Float("value", 200).Freeze(),
        streamv3.MakeMutableRecord().String("category", "A").Float("value", 150).Freeze(),
    }

    // Test pipeline
    grouped := streamv3.GroupByFields("test_data", "category")(slices.Values(input))
    results := streamv3.Aggregate("test_data", map[string]streamv3.AggregateFunc{
        "total": streamv3.Sum("value"),
        "count": streamv3.Count(),
    })(grouped)

    // Collect and verify results
    var collected []streamv3.Record
    for result := range results {
        collected = append(collected, result)
    }

    // Assertions
    if len(collected) != 2 {
        t.Errorf("Expected 2 groups, got %d", len(collected))
    }

    // Verify category A results
    for _, result := range collected {
        category := result["category"]
        if category == "A" {
            total := result["total"].(float64)
            count := result["count"].(int64)

            if total != 250 {
                t.Errorf("Category A total: expected 250, got %.0f", total)
            }
            if count != 2 {
                t.Errorf("Category A count: expected 2, got %d", count)
            }
        }
    }
}

func BenchmarkStreamProcessing(b *testing.B) {
    // Generate test data
    data := make([]streamv3.Record, 1000)
    for i := 0; i < 1000; i++ {
        data[i] = streamv3.MakeMutableRecord().
            String("id", fmt.Sprintf("item_%d", i)).
            Float("value", float64(i)).
            Build()
    }

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        // Benchmark processing pipeline
        filtered := streamv3.Where(func(r streamv3.Record) bool {
            value := streamv3.GetOr(r, "value", 0.0)
            return value > 500
        })(slices.Values(data))

        // Consume the stream
        count := 0
        for range filtered {
            count++
        }
    }
}
```

---

## What's Next?

Congratulations! You've mastered advanced StreamV3 patterns. Here's how to continue your journey:

### Production Readiness Checklist
- [ ] **Error Handling**: Implement comprehensive error recovery
- [ ] **Monitoring**: Add metrics and observability
- [ ] **Testing**: Create unit and integration tests
- [ ] **Configuration**: Externalize configuration management
- [ ] **Performance**: Profile and optimize critical paths
- [ ] **Documentation**: Document your data processing pipelines

### Advanced Exploration Areas

#### Real-World Applications
1. **Log Processing**: Build comprehensive log analysis systems
2. **IoT Data Streams**: Process sensor data in real-time
3. **Financial Analytics**: Implement trading and risk analysis
4. **Web Analytics**: Process user behavior data
5. **DevOps Monitoring**: Create system monitoring dashboards

#### Integration Patterns
1. **Database Integration**: Connect to PostgreSQL, MongoDB, etc.
2. **Message Queues**: Integrate with Kafka, RabbitMQ
3. **Cloud Services**: Process data from AWS S3, GCS, Azure
4. **APIs**: Build real-time data processing services
5. **Microservices**: Implement event-driven architectures

#### Advanced Techniques
1. **Custom Aggregation Functions**: Implement domain-specific aggregations
2. **Complex Event Processing**: Build stateful stream processors
3. **Machine Learning Integration**: Add predictive analytics
4. **Graph Processing**: Analyze relationship data
5. **Time Series Forecasting**: Implement predictive models

### Community and Resources

- **[API Reference](api-reference.md)** - Complete function documentation
- **[Getting Started Guide](codelab-intro.md)** - Foundational tutorials
- **Examples Directory** - Working code samples for common patterns
- **GitHub Issues** - Report bugs and request features

### Contributing Back

Consider contributing to StreamV3:
- Share your production patterns and use cases
- Contribute examples for common scenarios
- Report performance optimization opportunities
- Suggest new aggregation functions or operations

---

*You're now ready to build production-grade stream processing applications with StreamV3. Start with a real dataset and gradually apply these advanced patterns!*