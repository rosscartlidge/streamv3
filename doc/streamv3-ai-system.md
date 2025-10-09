# StreamV3 AI Code Generation System

*Complete guide for using AI to generate StreamV3 code - from simple prompts to advanced customization*

## 🎯 Overview

StreamV3 includes a comprehensive AI code generation system that enables any LLM to generate high-quality, human-readable StreamV3 code from natural language descriptions.

## 📚 **Documentation Map**

### For Humans Using LLMs
- **[Human LLM Tutorial](human-llm-tutorial.md)** - Step-by-step guide for using LLMs effectively
- **This document** - Complete system overview and setup

### For LLMs (Ready-to-Use)
- **[StreamV3 AI Prompt](streamv3-ai-prompt.md)** - Copy this into any LLM session
- **[Training Examples](nl-to-code-examples.md)** - 12+ natural language → code examples

---

## 🚀 **Quick Start: Use AI Right Now**

### Step 1: Copy the Prompt
Copy the complete prompt from **[streamv3-ai-prompt.md](streamv3-ai-prompt.md)** and paste it into your LLM session (Claude, ChatGPT, Gemini, etc.).

### Step 2: Ask for Code
Describe what you want in plain English:

```
"Read sales data from sales.csv, filter for amounts over $500, group by region, create a chart"
```

### Step 3: Get Working Code
The LLM generates clean, readable Go code:

```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read sales data
    salesData := streamv3.ReadCSV("sales.csv")

    // Filter for high-value sales
    highValueSales := streamv3.Where(func(r streamv3.Record) bool {
        amount := streamv3.GetOr(r, "amount", 0.0)
        return amount > 500
    })(salesData)

    // Group by region
    groupedByRegion := streamv3.GroupByFields("sales", "region")(highValueSales)

    // Calculate totals per region
    regionTotals := streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
        "total_revenue": streamv3.Sum("amount"),
        "sale_count":    streamv3.Count(),
    })(groupedByRegion)

    // Create chart
    streamv3.QuickChart(regionTotals, "sales_by_region.html")

    // Display results
    fmt.Println("Sales analysis complete. Chart saved as sales_by_region.html")
}
```

---

## 🔧 **System Components**

### 1. **Ready-to-Use Prompt** ([streamv3-ai-prompt.md](streamv3-ai-prompt.md))

**What it is:** A complete, optimized prompt that teaches any LLM to generate StreamV3 code.

**Contains:**
- ✅ Complete API reference (optimized for AI)
- ✅ Code generation rules (human-readable focus)
- ✅ Common patterns and examples
- ✅ Error prevention guidance
- ✅ Natural language → code mapping

**How to use:** Copy and paste into any LLM session.

### 2. **Training Examples** ([nl-to-code-examples.md](nl-to-code-examples.md))

**What it is:** 12+ comprehensive examples of natural language descriptions paired with working StreamV3 code.

**Use cases:**
- Few-shot learning with LLMs
- Fine-tuning training data
- Understanding patterns and capabilities

**Example pattern:**
```
Human Request: "Analyze employee data to find high-salary engineers by location"
Generated Code: [Complete working Go program]
```

### 3. **Human Tutorial** ([human-llm-tutorial.md](human-llm-tutorial.md))

**What it is:** Complete guide teaching humans how to effectively use LLMs for StreamV3 development.

**Covers:**
- Writing effective requests
- Verifying generated code
- Iterating and refining
- Troubleshooting common issues
- Best practices and patterns

---

## 🎯 **For Different Use Cases**

### Individual Developers (Most Common)

**Goal:** Generate StreamV3 code quickly and accurately.

**Setup:**
1. Copy [streamv3-ai-prompt.md](streamv3-ai-prompt.md) into your LLM
2. Follow the [human tutorial](human-llm-tutorial.md) for best practices
3. Start generating code from natural language

**Example workflow:**
```
"I need to process customer data to find purchasing patterns"
→ Gets working StreamV3 pipeline in seconds
```

### Teams and Organizations

**Goal:** Democratize data processing across the organization.

**Setup:**
1. Train team members using [human tutorial](human-llm-tutorial.md)
2. Share the [AI prompt](streamv3-ai-prompt.md) as standard setup
3. Create domain-specific examples based on your data

**Benefits:**
- Non-programmers can create data analysis pipelines
- Consistent code patterns across the team
- Rapid prototyping and iteration

### Local LLM and Custom Setups

**Goal:** Use your own LLM infrastructure for StreamV3 code generation.

#### What Should Your LLM Read?

**✅ LLM Should Read (Optimized for AI):**
1. **[streamv3-ai-prompt.md](streamv3-ai-prompt.md)** - The essential prompt (always include)
2. **[nl-to-code-examples.md](nl-to-code-examples.md)** - Training examples (for fine-tuning or few-shot)
3. **Selected examples from `examples/` directory** - Working code patterns

**❌ LLM Should NOT Read (Not optimized for AI):**
- ❌ Raw source code from the repo (too verbose, not instructional)
- ❌ Full API documentation (already condensed in the prompt)
- ❌ Tutorial markdown (designed for human learning, not AI training)

**Why this matters:**
- **Curated content** generates better code than raw source
- **Token efficiency** - every token should teach patterns, not implementation details
- **Human-readable focus** - we want the AI to prioritize clarity over source-code-style complexity

**Token Budget Guidelines:**
- **Small models (7B):** Just the essential API from the prompt (~2K tokens)
- **Medium models (13B-30B):** Full prompt + 2-3 examples (~4K-8K tokens)
- **Large models (70B+):** Complete prompt + training examples (~8K-16K tokens)
- **Fine-tuning:** All examples + API prompt + working code samples

### Fine-Tuning and Training

**Goal:** Create specialized models for StreamV3 code generation.

**Training data:**
1. **Base knowledge:** Complete [API reference](api-reference.md)
2. **Patterns:** All examples from [nl-to-code-examples.md](nl-to-code-examples.md)
3. **Real code:** All files from `examples/` directory
4. **Best practices:** Code generation rules from the AI prompt

---

## 🌟 **Key Innovation: Human-Readable Code**

Unlike other AI code generators, StreamV3's system prioritizes **human-readable, verifiable code**:

### ❌ **Typical AI Code (Complex)**
```go
// Hard to read and verify
result := streamv3.Limit(10)(streamv3.SortBy(func(r streamv3.Record) float64 {
    return -streamv3.GetOr(r, "revenue", 0.0)
})(streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
    "total": streamv3.Sum("amount"), "count": streamv3.Count(),
})(streamv3.GroupByFields("sales", "region")(data))))
```

### ✅ **StreamV3 AI Code (Clear)**
```go
// Easy to read and verify step by step
highValueSales := streamv3.Where(func(r streamv3.Record) bool {
    amount := streamv3.GetOr(r, "amount", 0.0)
    return amount > 1000
})(salesData)

groupedByRegion := streamv3.GroupByFields("sales", "region")(highValueSales)

regionTotals := streamv3.Aggregate("sales", map[string]streamv3.AggregateFunc{
    "total_revenue": streamv3.Sum("amount"),
    "sale_count":    streamv3.Count(),
})(groupedByRegion)

topRegions := streamv3.Limit[streamv3.Record](10)(
    streamv3.SortBy(func(r streamv3.Record) float64 {
        return -streamv3.GetOr(r, "total_revenue", 0.0)
    })(regionTotals),
)
```

**Why this matters:**
- Users can quickly understand what the code does
- Easy to modify and debug
- Builds trust in AI-generated solutions
- Enables learning and code review

---

## 🔍 **How It Works**

### 1. **Optimized Prompt Design**
The AI prompt includes:
- **SQL-style operation names** (Select, Where, Limit) - familiar to most users
- **Complete API coverage** - but condensed for token efficiency
- **Human-readable rules** - emphasizing clarity over cleverness
- **Common patterns** - proven solutions for typical tasks
- **Error prevention** - guidance to avoid common mistakes

### 2. **Natural Language Processing**
The system maps common phrases to StreamV3 operations:
- "filter/where/only" → `streamv3.Where(predicate)`
- "transform/convert/calculate" → `streamv3.Select(transformFn)`
- "group by X" → `streamv3.GroupByFields("group", "X")`
- "count/sum/average" → `streamv3.Aggregate("group", aggregations)`
- "chart/visualize" → `streamv3.QuickChart()` or `streamv3.InteractiveChart()`

### 3. **Code Generation Rules**
Every generated program follows these principles:
1. **Keep it simple** - no clever tricks or complex nesting
2. **One step at a time** - clear, logical progression
3. **Descriptive names** - `filteredSales` not `fs`
4. **Always handle errors** - especially file operations
5. **Complete examples** - include main function and imports

---

## 📊 **Capabilities**

### Data Processing
- CSV/JSON file reading and writing
- Filtering, transforming, and aggregating data
- SQL-style operations (GROUP BY, JOIN, etc.)
- Statistical analysis and calculations

### Real-Time Processing
- Stream windowing (time and count-based)
- Early termination patterns
- Infinite stream handling
- Real-time monitoring and alerting

### Visualization
- Interactive chart generation
- Dashboard creation
- Time series analysis
- Export capabilities (PNG, CSV)

### Advanced Patterns
- Stream joins and complex aggregations
- Performance optimization
- Error handling and resilience
- Production-ready pipelines

---

## 🚀 **Getting Started**

### For Immediate Use
1. **Go to [streamv3-ai-prompt.md](streamv3-ai-prompt.md)**
2. **Copy the entire prompt**
3. **Paste into your favorite LLM**
4. **Start asking for StreamV3 code**

### For Learning
1. **Read the [human tutorial](human-llm-tutorial.md)**
2. **Try the examples from [nl-to-code-examples.md](nl-to-code-examples.md)**
3. **Experiment with different request patterns**

### For Advanced Use
1. **Study the training examples** for pattern understanding
2. **Customize the prompt** for your domain-specific needs
3. **Build your own example library** for your use cases

---

## 🎓 **Best Practices**

### Writing Effective Requests
- **Be specific about data structure** and field names
- **Describe the steps** you want, not just the end goal
- **Mention output format** (console, chart, file)
- **Start simple** and build complexity incrementally

### Verifying Generated Code
- **Check imports** - should include necessary packages
- **Verify error handling** - especially for file operations
- **Review function names** - should use current API (Select, Where, Limit)
- **Test logic flow** - ensure steps make sense

### Iterating and Improving
- **Ask for clarifications** when code is unclear
- **Request simplifications** when code is too complex
- **Add domain context** for specialized use cases
- **Build a pattern library** of successful requests

---

## 🌟 **Success Stories**

### Marketing Campaign Analysis
**Request:** "Analyze email campaign performance across customer segments"
**Result:** Complete pipeline analyzing open rates, click rates, and conversions - delivered in minutes instead of hours.

### IoT Sensor Monitoring
**Request:** "Monitor temperature sensors and detect equipment failures"
**Result:** Real-time monitoring system with alerting and visualization - deployed in one afternoon.

### Financial Risk Analysis
**Request:** "Analyze trading patterns for compliance issues"
**Result:** Comprehensive risk analysis tool generating daily reports automatically.

---

## 📚 **Related Resources**

- **[Getting Started Guide](codelab-intro.md)** - Learn StreamV3 fundamentals
- **[API Reference](api-reference.md)** - Complete function documentation
- **[Advanced Tutorial](advanced-tutorial.md)** - Complex patterns and optimization
- **[Examples Directory](../examples/)** - Working code samples

---

*StreamV3 AI System: Where natural language meets production-ready stream processing code* ✨