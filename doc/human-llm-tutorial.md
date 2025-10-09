# How to Use LLMs to Generate StreamV3 Solutions

*A complete guide for humans on getting the best StreamV3 code from AI assistants*

## Overview

This tutorial teaches you how to effectively communicate with Large Language Models (LLMs) like Claude, ChatGPT, Gemini, and others to generate high-quality StreamV3 code from your natural language descriptions.

## Table of Contents

- [Quick Start](#quick-start)
- [Setting Up Your LLM](#setting-up-your-llm)
- [Interactive Development with CLI Tools](#interactive-development-with-cli-tools)
- [Writing Effective Requests](#writing-effective-requests)
- [Common Request Patterns](#common-request-patterns)
- [Verifying Generated Code](#verifying-generated-code)
- [Iterating and Refining](#iterating-and-refining)
- [Advanced Techniques](#advanced-techniques)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## Quick Start

### Step 1: Prime Your LLM

**Recommended:**
1. Open the file **[streamv3-ai-prompt.md](streamv3-ai-prompt.md)**
2. Copy the entire contents of that file (everything from "You are an expert Go developer..." to the end)
3. Paste it as your first message in your LLM conversation (Claude, ChatGPT, etc.)

This teaches the LLM how to generate proper StreamV3 code following the project's conventions.

**Alternative:** If you need a custom prompt, here's the essential content:

```
You are an expert Go developer specializing in StreamV3, a modern Go stream processing library. Generate high-quality, idiomatic StreamV3 code from natural language descriptions.

🎯 PRIMARY GOAL: Human-Readable, Verifiable Code

StreamV3 Quick Reference:

Imports - CRITICAL RULE:
ONLY import packages that are actually used in your code.

Common imports:
import (
    "fmt"                                    // When using fmt.Printf, fmt.Println
    "github.com/rosscartlidge/streamv3"     // Always needed
)

Additional imports - ONLY when actually used:
- "slices"  - ONLY if using slices.Values()
- "time"    - ONLY if using time.Duration, time.Time
- "strings" - ONLY if using strings.Fields, etc.

DO NOT import packages that aren't referenced in the code.

Core Operations (SQL-style naming):
- Transform: Select(func(T) U), SelectMany(func(T) iter.Seq[U])
- Filter: Where(func(T) bool), Distinct(), DistinctBy(func(T) K)
- Limit: Limit(n), Offset(n)
- Sort: Sort(), SortBy(func(T) K), SortDesc(), Reverse()
- Group: GroupByFields("groupName", "field1", "field2")
- Aggregate: Aggregate("groupName", map[string]AggregateFunc{...})
- Join: InnerJoin(rightSeq, predicate), LeftJoin(), RightJoin(), FullJoin()
- Window: CountWindow[T](size), TimeWindow[T](duration, "timeField"), SlidingCountWindow[T](size, step)

Record Access:
- streamv3.GetOr(record, "key", defaultValue) → T
- streamv3.Get[T](record, "key") → (T, bool)

Code Generation Rules:
1. Keep it simple: Write code humans can quickly read and verify - no clever tricks
2. One step at a time: Break complex operations into clear, logical steps
3. Descriptive variables: Use names like filteredSales, groupedData
4. Always handle errors from file operations
5. Use SQL-style names: Select not Map, Where not Filter, Limit not Take
6. Complete examples: Include main function and imports

Generate clean, working Go code that follows these patterns.
```

---

### Step 2: Make Your Request

Now describe what you want in plain English:

```
"Read sales data from sales.csv, filter for amounts over $500, group by region, and show total sales per region"
```

### Step 3: Review and Use

The LLM will generate clean, step-by-step code that you can easily read and verify.

---

## Setting Up Your LLM

### Popular LLMs and Setup

#### Claude (Anthropic)
- Visit: https://claude.ai
- Start new conversation
- Copy and paste the complete contents of **[streamv3-ai-prompt.md](streamv3-ai-prompt.md)** as your first message
- Begin asking for StreamV3 solutions

#### ChatGPT (OpenAI)
- Visit: https://chat.openai.com
- Start new chat
- Copy and paste the complete contents of **[streamv3-ai-prompt.md](streamv3-ai-prompt.md)** as your first message
- Ask for StreamV3 code generation

#### Gemini (Google)
- Visit: https://gemini.google.com
- New conversation
- Copy and paste the complete contents of **[streamv3-ai-prompt.md](streamv3-ai-prompt.md)** as your first message
- Request StreamV3 solutions

#### Local LLMs (Ollama, etc.)
- Install your preferred local LLM
- Copy and paste the complete contents of **[streamv3-ai-prompt.md](streamv3-ai-prompt.md)** as your first message
- Ensure the model has sufficient context window

### Prompt Template Variations

For simpler tasks, use this shorter version:

```
Generate StreamV3 Go code from natural language.
Use: Select() not Map(), Where() not Filter(), Limit() not Take().
Always include imports and error handling.
Keep code simple and human-readable - no complex chaining.
```

---

## Interactive Development with CLI Tools

For the most powerful StreamV3 development experience, use LLM CLI tools that can **run and iterate on code** for you. These tools provide an interactive coding assistant that can generate, execute, debug, and refine StreamV3 code in real-time.

### Claude Code (Recommended)

**What it is**: An official CLI from Anthropic that combines Claude's intelligence with direct access to your filesystem and the ability to run commands.

**Installation**:
```bash
npm install -g @anthropic-ai/claude-code
```

**Setting Up StreamV3 Knowledge**:

Before using Claude Code with StreamV3, you need to give it knowledge about the library. You have two options:

**Option 1: Quick Setup (Copy prompt each session)**
```bash
cd my-streamv3-project
claude-code

# First message: Copy and paste the entire contents of
# https://github.com/rosscartlidge/streamv3/blob/main/doc/streamv3-ai-prompt.md
```

**Option 2: Project Setup (Persistent knowledge - Recommended)**
```bash
# In your project directory, create a .claude directory
mkdir -p .claude

# Download the StreamV3 prompt
curl -o .claude/streamv3-reference.md \
  https://raw.githubusercontent.com/rosscartlidge/streamv3/main/doc/streamv3-ai-prompt.md

# Create a project-specific CLAUDE.md
cat > .claude/CLAUDE.md << 'EOF'
# StreamV3 Project

This project uses StreamV3 for stream processing and data analysis.

## StreamV3 Reference

See streamv3-reference.md in this directory for complete API documentation.

## Project Guidelines

- Use `streamv3.ReadCSV(filename)` to read CSV files (panics on error)
  - **⚠️ CSV Auto-Parsing**: Numeric strings become `int64`/`float64`!
  - Use `GetOr(r, "age", int64(0))` not `GetOr(r, "age", "")` for numeric CSV fields
- Use SQL-style names: `Select`, `Where`, `Limit` (not Map, Filter, Take)
- Access records safely with `streamv3.GetOr(record, "field", defaultValue)`
- Group with `GroupByFields("groupName", "field1", "field2")`
- Aggregate with `Aggregate("groupName", map[string]AggregateFunc{...})`
- Create charts with `QuickChart(data, "output.html")`

## Important
- Always include `package main` and `func main()`
- Only import packages actually used (don't import "slices" unless using slices.Values())
- Write human-readable code with descriptive variable names
- Break complex operations into clear steps
EOF

# Now start Claude Code - it will automatically read .claude/CLAUDE.md
claude-code
```

Now Claude Code will know about StreamV3 every time you work in this directory!

**Example Interactive Session**:
```
You: "Read sales.csv, filter for sales over $500, group by region, show top 5"

Claude Code:
- Generates the code
- Creates main.go file
- Runs it with `go run main.go`
- Shows you the output
- If there are errors, automatically fixes them

You: "Great! Now add a chart visualization"

Claude Code:
- Updates the code with chart generation
- Runs it again
- Opens the chart in your browser
- Iterates until it's perfect
```

**Why it's powerful**:
- ✅ **Automatic iteration**: Fixes compilation errors and runtime issues
- ✅ **File awareness**: Knows your CSV structure by reading the file
- ✅ **Runs code**: Actually executes and tests the code
- ✅ **Context aware**: Remembers your project structure and previous conversations
- ✅ **StreamV3 knowledge**: Has access to StreamV3 documentation via CLAUDE.md

**Best practices with Claude Code**:
```
✅ "Read my sales.csv file and tell me what fields are available"
   → Claude Code will read the file and show you the structure

✅ "Create a StreamV3 pipeline to analyze this data"
   → Generates, runs, and iterates until it works

✅ "The chart isn't showing correctly - can you fix it?"
   → Claude Code will read the generated HTML, identify issues, and fix them

✅ "Now do the same analysis but for products instead of regions"
   → Quick iteration on working code
```

### Gemini CLI (Google)

**Installation**:
```bash
# Install Gemini CLI (check current installation method)
npm install -g @google/generative-ai-cli
```

**Setting Up StreamV3 Knowledge**:
```bash
gemini-cli

# First message: Copy and paste the entire contents of
# https://github.com/rosscartlidge/streamv3/blob/main/doc/streamv3-ai-prompt.md

# Then start requesting StreamV3 code
```

**Advantages**:
- Good for quick iterations
- Fast response times
- Multi-modal capabilities (can read images if you have chart screenshots)
- Tested and works well with StreamV3

### GitHub Copilot CLI

**Installation**:
```bash
gh extension install github/gh-copilot
```

**Usage**:
```bash
gh copilot suggest "create streamv3 pipeline to analyze sales data"
```

**Advantages**:
- Integrates with GitHub workflow
- Good for command-line operations
- Can suggest both code and terminal commands

### Aider (Local Development)

**What it is**: An AI pair programmer that works with your local editor and can use various LLM backends.

**Installation**:
```bash
pip install aider-chat
```

**Setting Up StreamV3 Knowledge**:
```bash
cd my-streamv3-project

# Download the StreamV3 reference
curl -o streamv3-reference.md \
  https://raw.githubusercontent.com/rosscartlidge/streamv3/main/doc/streamv3-ai-prompt.md

# Start aider and add the reference
aider

# In aider, first message:
/add streamv3-reference.md
"Read this StreamV3 reference. I'll be asking you to generate StreamV3 code."

# Then start coding:
/add main.go
"Create a StreamV3 pipeline to process sales data"
```

**Advantages**:
- Works with multiple LLM backends (GPT-4, Claude, local models)
- Git integration - automatically commits working code
- Can edit existing files intelligently
- Great for iterative development

### Interactive Development Workflow

Here's a complete workflow using Claude Code or similar tools:

#### Step 1: Project Setup
```bash
mkdir my-analysis
cd my-analysis
go mod init my-analysis
go get github.com/rosscartlidge/streamv3

# Start Claude Code
claude-code
```

#### Step 2: Explore Your Data
```
You: "I have a file called sales.csv. Can you read it and show me what fields are in it?"

Claude Code:
- Reads sales.csv
- Shows you the CSV structure
- Suggests relevant StreamV3 operations
```

#### Step 3: Generate Initial Pipeline
```
You: "Create a StreamV3 pipeline that:
1. Reads sales.csv
2. Filters for amounts over $500
3. Groups by region and product
4. Calculates total revenue and count
5. Shows the top 10 by revenue"

Claude Code:
- Generates complete Go program
- Saves it to main.go
- Runs `go run main.go`
- Shows you the output
```

#### Step 4: Iterate and Refine
```
You: "The output looks good but I also want to see the average sale amount per region"

Claude Code:
- Updates the aggregation
- Reruns the code
- Shows updated results

You: "Perfect! Now create a bar chart of the top 10 regions"

Claude Code:
- Adds chart generation
- Runs the code
- Chart opens in your browser
- You see the visualization

You: "Can you change it to a horizontal bar chart and use a dark theme?"

Claude Code:
- Updates chart configuration
- Regenerates the chart
- Shows the improved visualization
```

#### Step 5: Handle Edge Cases
```
You: "What if the CSV file doesn't exist?"

Claude Code:
- Adds error handling
- Shows you the updated code
- Explains the error handling approach
```

### Comparison: Web LLMs vs CLI Tools

| Feature | Web LLMs (Claude.ai, ChatGPT) | CLI Tools (Claude Code, Aider) |
|---------|-------------------------------|--------------------------------|
| **Code Generation** | ✅ Excellent | ✅ Excellent |
| **File Access** | ❌ None | ✅ Can read your files |
| **Code Execution** | ❌ Manual | ✅ Automatic |
| **Error Fixing** | 🔶 Manual iteration | ✅ Automatic iteration |
| **Context Awareness** | 🔶 Limited | ✅ Full project context |
| **Chart Viewing** | ❌ Manual | ✅ Opens automatically |
| **Learning Curve** | ✅ Easy | 🔶 Moderate |
| **Best For** | Learning, planning | Development, production |

### Tips for Interactive CLI Development

**1. Start with exploration**:
```
"Read my data files and tell me what you see"
"What fields are available in customers.csv?"
"Show me the first few records"
```

**2. Build incrementally**:
```
"First, just read and filter the data"
→ Works? Good!
"Now add grouping and aggregation"
→ Works? Great!
"Now add the chart"
→ Perfect!
```

**3. Let the tool iterate**:
```
❌ Don't: "Generate code" → copy → paste → run → find errors → ask for fix
✅ Do: "Create a pipeline to analyze sales" → tool generates, runs, and fixes automatically
```

**4. Use specific file references**:
```
✅ "Read data from sales.csv in the data/ directory"
✅ "The CSV has columns: date, amount, region, product"
❌ "Read some sales data"
```

**5. Ask for improvements**:
```
"The code works but it's a bit slow with large files - can you optimize it?"
"Can you add progress indicators?"
"Add logging so I can see what's happening at each step"
```

### Example: Complete Interactive Session

```bash
$ claude-code

You: "I want to analyze monthly sales trends. I have sales_2024.csv with columns:
date, product_id, quantity, unit_price, total_amount, region"

Claude Code:
✓ Read sales_2024.csv
✓ Detected 15,234 records
✓ Confirmed column structure

"I can help you analyze this data. What specific insights are you looking for?"

You: "Show me total sales by month and region, with a line chart showing the trend"

Claude Code:
✓ Created main.go with StreamV3 pipeline
✓ Running: go run main.go
✓ Generated analysis:
  - Grouped by month and region
  - Calculated total sales per group
  - Created chart: sales_trends.html
✓ Opening chart in browser...

"The chart shows sales trends across 4 regions over 12 months.
I notice the East region has strong growth. Would you like me to add any filters or drill-downs?"

You: "Yes - filter for just Q4 (October-December) and show which products are driving the East region growth"

Claude Code:
✓ Updated pipeline with Q4 filter
✓ Added product breakdown for East region
✓ Running updated analysis...
✓ Generated new chart: q4_east_products.html

"Here are the top products in East region for Q4. Product SKU-1234 shows
the strongest growth. Would you like to export this data to CSV?"

You: "Yes please, and also create a summary report"

Claude Code:
✓ Added CSV export
✓ Created summary report in markdown
✓ Files generated:
  - q4_analysis.csv (detailed data)
  - summary_report.md (executive summary)
  - q4_east_products.html (interactive chart)

"Analysis complete! All files are in the current directory."
```

---

## Writing Effective Requests

### ✅ Good Request Patterns

#### 1. Be Specific About Data
```
❌ "Analyze some data"
✅ "Read customer data from customers.csv, filter for customers with orders > $1000"
```

#### 2. Describe the Steps
```
❌ "Do sales analysis"
✅ "Read sales.csv, group by product category, calculate total revenue per category, show top 5"
```

#### 3. Mention Output Format
```
❌ "Show results"
✅ "Display results in console and create a bar chart saved as sales_chart.html"
```

#### 4. Specify Data Fields
```
❌ "Filter the data"
✅ "Filter for records where the 'amount' field is greater than 500"
```

### Request Templates

#### Data Analysis Template
```
"Read [filename] data, filter for [condition], group by [field], calculate [aggregation], and [output format]"

Example: "Read sales.csv data, filter for amounts > $100, group by region, calculate total and average sales, and display top 10 regions"
```

#### Real-Time Processing Template
```
"Process [data source] in [time window] windows, calculate [metric], alert when [condition]"

Example: "Process sensor data in 5-minute windows, calculate average temperature, alert when temperature exceeds 35°C"
```

#### Visualization Template
```
"Create [chart type] showing [metric] by [dimension] from [data source]"

Example: "Create line chart showing monthly revenue trends by product category from sales.csv"
```

#### Join Analysis Template
```
"Join [dataset1] with [dataset2] on [key], find [pattern], calculate [metric]"

Example: "Join customers.csv with orders.csv on customer_id, find high-value customers, calculate lifetime value"
```

---

## Common Request Patterns

### 1. Basic Data Analysis

**Request:**
```
"Read employee data from employees.csv, group by department and location, count employees in each group, then filter to show only groups with more than 5 employees"
```

**Expected Output:**
```go
package main

import (
    "fmt"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read employee data
    employees := streamv3.ReadCSV("employees.csv")

    // Group by department and location
    grouped := streamv3.GroupByFields("dept_location_analysis", "department", "location")(employees)

    // Count employees in each group
    employeeCounts := streamv3.Aggregate("dept_location_analysis", map[string]streamv3.AggregateFunc{
        "employee_count": streamv3.Count(),
    })(grouped)

    // Filter to show only groups with more than 5 employees
    largeGroups := streamv3.Where(func(r streamv3.Record) bool {
        count := streamv3.GetOr(r, "employee_count", int64(0))
        return count > 5
    })(employeeCounts)

    // Display results
    fmt.Println("Department/Location groups with more than 5 employees:")
    for result := range largeGroups {
        dept := streamv3.GetOr(result, "department", "")
        location := streamv3.GetOr(result, "location", "")
        count := streamv3.GetOr(result, "employee_count", int64(0))
        fmt.Printf("%s - %s: %d employees\n", dept, location, count)
    }
}
```

### 2. Time Series Analysis

**Request:**
```
"Analyze website traffic data by day, calculate 7-day moving average, identify days with traffic 20% above average"
```

### 3. Real-Time Monitoring

**Request:**
```
"Monitor API response times in 1-minute windows, calculate average response time, alert if average exceeds 500ms"
```

### 4. Complex Joins

**Request:**
```
"Join product catalog with sales data and inventory, find products that are selling well but have low stock levels"
```

---

## Verifying Generated Code

### Quick Verification Checklist

When you receive generated code, check these items:

#### ✅ **Structure Check**
- [ ] Has `package main` and `func main()`
- [ ] Includes all necessary imports
- [ ] Handles errors from file operations
- [ ] Uses descriptive variable names

#### ✅ **StreamV3 API Check**
- [ ] Uses current function names (`Select`, `Where`, `Limit`)
- [ ] Uses `streamv3.GetOr()` for record access
- [ ] Includes group names in `GroupByFields()` and `Aggregate()`
- [ ] Proper type parameters where needed

#### ✅ **Logic Check**
- [ ] Processing steps are in logical order
- [ ] Each step is clearly understandable
- [ ] Variable names match their purpose
- [ ] Comments explain complex logic

#### ✅ **Completeness Check**
- [ ] Addresses all parts of your request
- [ ] Includes output/display logic
- [ ] Has proper error handling
- [ ] Code looks complete and runnable

### Common Issues to Watch For

#### ❌ **API Mistakes**
```go
// Wrong - old API
result := streamv3.Map(fn)(data)      // Should be Select
result := streamv3.Filter(fn)(data)   // Should be Where
result := streamv3.Take(10)(data)     // Should be Limit
```

#### ❌ **Unsafe Record Access**
```go
// Wrong - direct map access could panic
name := record["name"].(string)

// Correct - use GetOr for safe access
name := streamv3.GetOr(record, "name", "")
```

#### ❌ **Unsafe Record Access**
```go
// Wrong - could panic
name := record["name"].(string)

// Correct
name := streamv3.GetOr(record, "name", "")
```

---

## Iterating and Refining

### When the First Result Isn't Perfect

#### Strategy 1: Ask for Clarification
```
"The code looks good, but can you add comments explaining the aggregation step?"
```

#### Strategy 2: Request Modifications
```
"Can you modify this to also calculate the standard deviation of sales amounts?"
```

#### Strategy 3: Ask for Simplification
```
"This code is a bit complex. Can you break it into simpler steps with more descriptive variable names?"
```

#### Strategy 4: Request Different Approach
```
"Instead of grouping by region first, can you filter for high-value sales first, then group?"
```

### Example Iteration Session

**Initial Request:**
```
"Analyze sales data to find top-performing products"
```

**LLM Response:** *[generates basic code]*

**Your Follow-up:**
```
"Good start! Can you modify this to:
1. Only include sales from the last 30 days
2. Calculate both revenue and unit sales
3. Show the top 10 products by revenue
4. Create a bar chart of the results"
```

**LLM Response:** *[generates enhanced code with all requirements]*

---

## Advanced Techniques

### 1. Multi-Step Requests

Break complex analyses into phases:

```
"First, help me understand this data structure. Then I'll ask you to build the analysis pipeline."

Phase 1: "Read my sales.csv and show me what fields are available and some sample records"
Phase 2: "Now create analysis to find seasonal sales patterns by product category"
Phase 3: "Add visualization and export the results to Excel format"
```

### 2. Template Creation

Ask the LLM to create reusable templates:

```
"Create a template function that can analyze any CSV file to find top N records by any numeric field, with flexible grouping"
```

### 3. Domain-Specific Language

Teach the LLM your domain terminology:

```
"In our retail business:
- 'SKU' means product identifier
- 'COGS' means cost of goods sold
- 'LTV' means customer lifetime value

Now analyze our product performance using these terms."
```

### 4. Error Handling Strategies

Request robust error handling:

```
"Generate the sales analysis code, but make it robust - handle missing files, empty data, and invalid field values gracefully"
```

### 5. Performance Optimization

Ask for performance considerations:

```
"The previous code works but might be slow with large files. Can you optimize it for processing 1M+ records efficiently?"
```

---

## Troubleshooting

### Common Problems and Solutions

#### Problem: Code Uses Wrong Function Names
**Symptoms:** `Map()`, `Filter()`, `Take()` instead of `Select()`, `Where()`, `Limit()`

**Solution:**
```
"Please update this code to use the current StreamV3 API: Select instead of Map, Where instead of Filter, Limit instead of Take"
```

#### Problem: Complex, Hard-to-Read Code
**Symptoms:** Deeply nested function calls, unclear variable names

**Solution:**
```
"Can you rewrite this with simpler, step-by-step processing? Use descriptive variable names and break complex operations into separate steps."
```

#### Problem: Missing Error Handling
**Symptoms:** File operations without error checks

**Solution:**
```
"Please add proper error handling for all file operations and CSV reading"
```

#### Problem: Incorrect Aggregation Syntax
**Symptoms:** Missing group names in `Aggregate()` calls

**Solution:**
```
"The Aggregate function needs a group name as the first parameter. Please fix the aggregation calls."
```

#### Problem: Type Issues
**Symptoms:** Compiler errors about types

**Solution:**
```
"There seem to be some type issues. Can you add explicit type parameters where needed, like CountWindow[streamv3.Record](10)?"
```

### When Your LLM Gets Confused

#### Reset and Restart
If the LLM starts generating incorrect code consistently:

1. Start a new conversation
2. Reapply the prompt template
3. Rephrase your request more clearly
4. Provide a simple example first

#### Provide Examples
Show the LLM what good code looks like:

**Example request:**

> "Here's an example of the coding style I want:
>
> ```go
> // Read and filter data
> salesData := streamv3.ReadCSV("sales.csv")
>
> highValueSales := streamv3.Where(func(r streamv3.Record) bool {
>     amount := streamv3.GetOr(r, "amount", 0.0)
>     return amount > 1000
> })(salesData)
> ```
>
> Now create similar code for analyzing customer data."

---

## Best Practices

### 1. Start Simple, Build Complexity

```
✅ Good Progression:
1. "Read sales.csv and show first 10 records"
2. "Filter for sales > $100"
3. "Group by region and calculate totals"
4. "Create chart of results"

❌ Avoid:
"Build a comprehensive sales analytics dashboard with real-time monitoring, predictive analytics, and machine learning"
```

### 2. Be Specific About Data

```
✅ Good:
"The CSV has columns: customer_id, order_date, product_name, quantity, unit_price, total_amount"

❌ Vague:
"Analyze the sales data"
```

### 3. Request Human-Readable Code

```
✅ Always ask for:
"Make the code easy to read and verify, with descriptive variable names and clear steps"

❌ Don't accept:
Complex nested function calls that are hard to understand
```

### 4. Verify Each Step

```
✅ Good practice:
"Before adding the chart, let me verify the data processing logic first"

❌ Risky:
Accepting large, complex code without reviewing each component
```

### 5. Build a Library of Patterns

Save successful request patterns:

```
📁 Your Pattern Library:
- "Read [file], filter for [condition], group by [field], calculate [metric]"
- "Process [data] in [time] windows, detect [pattern], alert when [condition]"
- "Join [dataset1] with [dataset2], find [insight], create [visualization]"
```

### 6. Test with Small Data First

```
✅ Smart approach:
"Generate code to test with a small sample file first, then we'll apply it to the full dataset"

❌ Risky:
Running generated code on large production datasets without testing
```

---

## Example Session Walkthrough

Here's a complete example of an effective LLM session:

### Session Start
```
Human: [Pastes StreamV3 prompt template]

LLM: I'm ready to help you generate StreamV3 code! What data processing task would you like to accomplish?
```

### Request 1: Start Simple
```
Human: "I have a CSV file called customer_orders.csv with columns: customer_id, order_date, product_category, order_amount. Can you show me how to read it and display the first few records?"

LLM: [Generates simple CSV reading code with error handling]

Human: "Perfect! That works. Now can you filter this data to show only orders with amount > $100 and group them by product category?"

LLM: [Generates filtering and grouping code]

Human: "Great! The logic is clear. Now add aggregation to calculate total revenue and order count per category, and sort by total revenue."

LLM: [Generates aggregation and sorting code]

Human: "Excellent! Finally, can you create a bar chart showing these results and save it as category_analysis.html?"

LLM: [Generates complete solution with visualization]
```

### Why This Session Worked

1. **Started simple** - Basic CSV reading first
2. **Built incrementally** - Added features step by step
3. **Verified each step** - Confirmed logic before adding complexity
4. **Clear requirements** - Specific about data fields and desired outputs
5. **Human-readable results** - Each step was easy to understand and verify

---

## Quick Reference Cards

### Request Templates You Can Copy

#### Data Analysis Requests
```
"Read [filename], filter for [condition], group by [field], calculate [metrics], show top [N]"

"Analyze [dataset] to find [pattern], calculate [statistics], display results as [format]"

"Compare [metric] across [dimensions] in [dataset], highlight [insights]"
```

#### Real-Time Processing Requests
```
"Monitor [data source] in [time window] windows, calculate [metrics], alert when [condition]"

"Process [streaming data] to detect [patterns], maintain [running statistics], report [anomalies]"

"Track [events] over time, identify [trends], predict [outcomes]"
```

#### Visualization Requests
```
"Create [chart type] showing [metric] by [dimension], save as [filename]"

"Build dashboard with [multiple charts] analyzing [different aspects] of [dataset]"

"Generate interactive chart where users can filter by [criteria] and drill down into [details]"
```

#### Join and Complex Analysis
```
"Join [dataset1] with [dataset2] on [key field], analyze [combined insights], find [patterns]"

"Combine multiple data sources to create [comprehensive analysis], focusing on [business question]"

"Merge [data types] to build [360-degree view] of [business entity], calculate [derived metrics]"
```

### Verification Checklist You Can Use

Print this out and check each generated code:

```
□ Package main and func main() present
□ ONLY imports packages that are actually used in the code (no unused imports like "slices" or "time")
□ Error handling for file operations
□ Uses current API (Select, Where, Limit, not Map, Filter, Take)
□ Safe record access (GetOr, Get[T])
□ Descriptive variable names
□ Clear, step-by-step processing
□ Comments explain complex logic
□ Addresses all parts of request
□ Output/display logic included
□ Code looks complete and runnable
```

### Common Fixes You Can Request

```
"Please add error handling for file operations"
"Can you use more descriptive variable names?"
"Break this into simpler steps with intermediate variables"
"Add comments explaining the aggregation logic"
"Use the current StreamV3 API (Select not Map, Where not Filter)"
"Include the missing imports"
"Add type parameters where needed"
"Make the code easier to read and verify"
```

---

## Success Stories

### Story 1: Marketing Campaign Analysis

**User**: Sarah, Marketing Analyst
**Challenge**: Analyze email campaign performance across different customer segments

**Approach**:
1. Started with: "Read campaign_data.csv and show me the data structure"
2. Built up: "Filter for campaigns from last quarter, group by customer segment"
3. Added complexity: "Calculate open rates, click rates, and conversion rates"
4. Finished with: "Create dashboard showing performance comparison across segments"

**Result**: Complete analysis pipeline in 15 minutes, would have taken hours to code manually.

### Story 2: IoT Sensor Monitoring

**User**: Mike, Operations Engineer
**Challenge**: Monitor temperature sensors in real-time and detect equipment failures

**Approach**:
1. "Process sensor readings in 5-minute windows"
2. "Calculate average and detect readings outside normal range"
3. "Group alerts by equipment location"
4. "Generate real-time monitoring dashboard"

**Result**: Production monitoring system deployed in one afternoon.

### Story 3: Financial Risk Analysis

**User**: Lisa, Risk Analyst
**Challenge**: Analyze trading patterns to identify potential compliance issues

**Approach**:
1. "Join trade data with account information"
2. "Calculate trading volumes and frequencies by account"
3. "Detect unusual patterns and outliers"
4. "Generate risk report with visualization"

**Result**: Comprehensive risk analysis tool that runs daily reports automatically.

---

## Going Further

### Learning More About StreamV3

Once you're comfortable generating code with LLMs, explore these resources:

- **[Getting Started Guide](codelab-intro.md)** - Learn StreamV3 fundamentals
- **[API Reference](api-reference.md)** - Complete function documentation
- **[Advanced Tutorial](advanced-tutorial.md)** - Complex patterns and optimization

### Building Your Own Templates

Create domain-specific prompt templates:

```
"You're a financial data specialist. Generate StreamV3 code for:
- Portfolio analysis and risk calculations
- Trading pattern detection
- Regulatory compliance reporting
- Performance attribution analysis

Always include proper financial calculations and risk metrics."
```

### Automating Your Workflow

1. **Save successful prompts** - Build a library of working templates
2. **Document patterns** - Keep notes on what works well
3. **Share with team** - Spread effective techniques across your organization
4. **Iterate and improve** - Refine your requests based on results

### Contributing Back

If you discover particularly effective prompt patterns:
- Share them with the StreamV3 community
- Contribute examples to the documentation
- Help improve the AI generation system

---

## Conclusion

With the right approach, LLMs can become incredibly powerful tools for generating StreamV3 solutions. The key principles are:

🎯 **Start with clear, specific requests**
📋 **Build complexity incrementally**
✅ **Verify each step before proceeding**
🔍 **Prioritize human-readable, verifiable code**
🔄 **Iterate and refine as needed**

Remember: The goal isn't just working code - it's code you can understand, verify, and maintain. Always prioritize clarity over cleverness.

Happy stream processing! 🚀

---

## 📚 Related Resources

### StreamV3 AI System
- **[StreamV3 AI System](streamv3-ai-system.md)** - Complete overview of AI code generation
- **[StreamV3 AI Prompt](streamv3-ai-prompt.md)** - Ready-to-use prompt for any LLM
- **[Training Examples](nl-to-code-examples.md)** - Natural language → code patterns

### StreamV3 Documentation
- **[Getting Started Guide](codelab-intro.md)** - Learn StreamV3 basics
- **[API Reference](api-reference.md)** - Complete function documentation
- **[Advanced Tutorial](advanced-tutorial.md)** - Complex patterns and optimization
- **[Examples Directory](../examples/)** - Working code samples

---

*For the complete AI generation system, see [streamv3-ai-system.md](streamv3-ai-system.md)*