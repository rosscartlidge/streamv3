# ssql CLI Tutorial

*Command-line data processing with code generation - Still in active development*

## Table of Contents

### Documentation Navigation
- **[API Reference](api-reference.md)** - Complete function reference
- **[Getting Started Guide](codelab-intro.md)** - Learn the library fundamentals
- **[Advanced Tutorial](advanced-tutorial.md)** - Complex patterns and optimization
- **[Debugging Pipelines](debugging_pipelines.md)** - Debug with jq, inspect data, profile performance
- **[Troubleshooting Guide](troubleshooting.md)** - Common issues and quick solutions

### Learning Path
- [Quick Start](#quick-start)
- [What is the ssql CLI?](#what-is-the-streamv3-cli)
- [Basic Pipeline Operations](#basic-pipeline-operations)
- [Working with Real Data](#working-with-real-data)
- [Grouping and Aggregations](#grouping-and-aggregations)
- [SQL-Like Operations](#sql-like-operations)
- [Creating Visualizations](#creating-visualizations)
- [Code Generation](#code-generation)
- [Complete Example](#complete-example)
- [Available Commands](#available-commands)
- [What's Next?](#whats-next)

---

## Quick Start

### Installation

```bash
# Install the CLI tool
go install github.com/rosscartlidge/ssql/cmd/ssql@latest

# Verify installation
ssql -version
```

### Your First Pipeline

```bash
# Create a sample CSV file
cat > employees.csv << 'EOF'
name,age,department,salary
Alice,30,Engineering,95000
Bob,25,Marketing,65000
Carol,35,Engineering,105000
David,28,Sales,70000
EOF

# Process it with a pipeline
ssql read-csv employees.csv | \
  ssql where -match department eq Engineering | \
  ssql include name salary
```

Output:
```json
{"name":"Alice","salary":95000}
{"name":"Carol","salary":105000}
```

---

## What is the ssql CLI?

The ssql CLI brings Unix pipeline philosophy to structured data processing. It provides:

- **🔗 Pipeline Operations** - Chain commands with Unix pipes
- **📊 Built-in Visualization** - Create charts directly from pipelines
- **🤖 Code Generation** - Convert CLI commands to Go code
- **⚡ Interactive Development** - Prototype fast, then generate production code

### Key Features

**Command Chaining**
```bash
ssql read-csv data.csv | ssql where ... | ssql group ... | ssql chart ...
```

**Self-Generating Commands**
Every command supports `-generate` flag to emit Go code instead of executing:
```bash
ssql read-csv -generate data.csv | ssql generate-go
```

**Universal Data Format**
All commands use JSONL (JSON Lines) for inter-command communication, enabling complex pipelines.

**Debugging with jq**
Since all commands communicate via JSONL, you can inspect data at any stage with `jq`:
```bash
ssql read-csv data.csv | jq '.' | head -5          # Pretty-print data
ssql read-csv data.csv | jq '.age | type' | head   # Check field types
ssql ... | ssql where ... | jq -s 'length'      # Count results
```
[**See full debugging guide →**](debugging_pipelines.md)

> ⚠️ **Development Status**: The CLI is under active development. Commands and flags may change. Use `-help` on any command to see current options.

---

## Basic Pipeline Operations

### Reading Data

Read CSV files and output as JSONL:

```bash
ssql read-csv employees.csv
```

Output (JSONL):
```json
{"_row_number":0,"age":30,"department":"Engineering","name":"Alice","salary":95000}
{"_row_number":1,"age":25,"department":"Marketing","name":"Bob","salary":65000}
...
```

### Filtering Data

Filter records based on conditions:

```bash
# Single condition
ssql read-csv employees.csv | \
  ssql where -match salary gt 70000

# Multiple conditions (AND)
ssql read-csv employees.csv | \
  ssql where -match age gt 25 -match department eq Engineering

# Multiple conditions (OR) - use + separator
ssql read-csv employees.csv | \
  ssql where -match department eq Engineering + -match department eq Sales
```

**Available Operators:**
- `eq` - Equal
- `ne` - Not equal
- `gt` - Greater than
- `lt` - Less than
- `ge` - Greater than or equal
- `le` - Less than or equal
- `contains` - String contains
- `starts` - String starts with
- `ends` - String ends with

### Selecting Fields

Select specific fields or rename them:

```bash
# Select fields
ssql read-csv employees.csv | \
  ssql include name salary

# Rename fields
ssql read-csv employees.csv | \
  ssql include name salary | \
  ssql rename -as name employee_name -as salary annual_salary
```

### Updating Fields

Update record fields conditionally using if-elseif-else logic:

```bash
# Unconditional update - all records
ssql read-csv employees.csv | \
  ssql update -set status "active"

# Conditional update - only matching records
ssql read-csv employees.csv | \
  ssql update -match salary gt 100000 -set bracket "high"

# Multiple conditions (AND logic)
ssql read-csv employees.csv | \
  ssql update -match status eq pending -match priority eq urgent -set assignee "alice"

# If-elseif-else with + separator (first match wins)
ssql read-csv customers.csv | \
  ssql update \
    -match purchases gt 5000 -set tier "Gold" -set discount 0.2 + \
    -match purchases gt 1000 -set tier "Silver" -set discount 0.1 + \
    -set tier "Bronze" -set discount 0.0
```

**How It Works:**
- **Without `-match`**: Updates all records
- **With `-match`**: Only updates records matching conditions, others pass through unchanged
- **Multiple `-match` flags**: AND logic (all must match)
- **`+` separator**: Creates clauses for if-elseif-else logic (first matching clause wins)
- **Default clause**: Clause with no `-match` acts as else (catches all remaining records)

**Type Inference:**
The `update` command automatically infers types from string values:
- `"123"` → integer (`int64`)
- `"99.99"` → float (`float64`)
- `"true"` / `"false"` → boolean
- `"2025-11-04"` → time.Time (if valid date format)
- Everything else → string

**Complex Example:**
```bash
# Set priority based on multiple conditions
ssql read-csv orders.csv | \
  ssql update \
    -match status eq pending -match amount gt 10000 -set priority "critical" -set sla 24 + \
    -match status eq pending -match amount gt 1000 -set priority "high" -set sla 48 + \
    -match status eq pending -set priority "normal" -set sla 72 + \
    -set priority "low" -set sla 168
```

This keeps ALL records while selectively updating fields based on conditions.

### Writing Output

Write results to CSV:

```bash
ssql read-csv employees.csv | \
  ssql where -match department eq Engineering | \
  ssql write-csv engineers.csv
```

### Displaying Data as Tables

Display records in a formatted table on the terminal:

```bash
# Simple table display
ssql read-csv employees.csv | ssql table

# With filtering
ssql read-csv employees.csv | \
  ssql where -match department eq Engineering | \
  ssql table

# Limit column width to prevent wrapping
ssql read-csv employees.csv | \
  ssql table -max-width 30

# Complex pipeline with updates and filtering
ssql read-csv customers.csv | \
  ssql update \
    -match purchases gt 5000 -set tier "Gold" + \
    -match purchases gt 1000 -set tier "Silver" + \
    -set tier "Bronze" | \
  ssql where -match tier eq Gold | \
  ssql table
```

**Features:**
- Automatically calculates column widths
- Sorts columns alphabetically for consistent output
- Truncates long values with `...` when exceeding `-max-width`
- Works with all field types (strings, numbers, dates, etc.)
- Supports code generation with `-generate` flag

**Example output:**
```
_row_number   age   city      name      salary
----------------------------------------------
0             30    NYC       Alice     95000
1             25    LA        Bob       75000
2             35    Chicago   Charlie   120000
```

---

## Working with Real Data

### Processing Command Output

Execute shell commands and parse their output:

```bash
# Analyze process information
ssql exec -- ps -efl | \
  ssql where -match CMD contains chrome | \
  ssql include PID USER CMD
```

**Note:** The `--` separator is required to prevent ssql from interpreting command flags like `-efl` as its own flags.

### Example: System Monitoring

Find memory-intensive processes:

```bash
# Get top memory users
ssql exec -- ps aux | \
  ssql where -match USER eq root | \
  ssql include PID MEM CMD | \
  ssql write-csv system_processes.csv
```

---

## Grouping and Aggregations

Group data and calculate statistics:

### Basic Aggregation

```bash
# Count records by department
ssql read-csv employees.csv | \
  ssql group-by department -function count -result total
```

Output:
```json
{"department":"Engineering","total":3}
{"department":"Marketing","total":2}
{"department":"Sales","total":1}
```

### Multiple Aggregations

Use `+` to separate multiple aggregation functions:

```bash
ssql read-csv employees.csv | \
  ssql group-by department \
    -function count -result employee_count + \
    -function avg -field salary -result avg_salary + \
    -function max -field salary -result max_salary
```

Output:
```json
{"avg_salary":98333,"department":"Engineering","employee_count":3,"max_salary":105000}
{"avg_salary":65000,"department":"Marketing","employee_count":2,"max_salary":65000}
{"avg_salary":70000,"department":"Sales","employee_count":1,"max_salary":70000}
```

**Available Aggregation Functions:**
- `count` - Count records
- `sum` - Sum values
- `avg` - Average values
- `min` - Minimum value
- `max` - Maximum value

---

## SQL-Like Operations

ssql supports common SQL operations for multi-table queries and data manipulation.

### Pagination with OFFSET and LIMIT

Skip and take records for pagination:

```bash
# Skip first 20 records, take next 10 (records 21-30)
ssql read-csv data.csv | \
  ssql offset 20 | \
  ssql limit 10
```

Equivalent SQL:
```sql
SELECT * FROM data LIMIT 10 OFFSET 20
```

### Remove Duplicates with DISTINCT

Remove duplicate records:

```bash
# Distinct on all fields
ssql read-csv data.csv | ssql distinct

# Distinct by specific fields
ssql read-csv employees.csv | \
  ssql distinct -by department -by location
```

Equivalent SQL:
```sql
SELECT DISTINCT department, location FROM employees
```

### JOIN Operations

Join two data sources on common fields:

```bash
# Inner join on same field name
ssql read-csv employees.csv | \
  ssql join -type inner -right departments.csv -on dept_id

# Left join with different field names
ssql read-csv orders.csv | \
  ssql join -type left -right customers.csv \
    -left-field customer_id -right-field id

# Join on multiple fields (composite key)
ssql read-csv sales.csv | \
  ssql join -right products.csv \
    -on product_id -on region
```

**Join Types:**
- `inner` - Only matching records (default)
- `left` - All left records, matched right records
- `right` - All right records, matched left records
- `full` - All records from both sides

Equivalent SQL:
```sql
SELECT * FROM employees e
INNER JOIN departments d ON e.dept_id = d.dept_id
```

### UNION Operations

Combine multiple data sources:

```bash
# UNION (remove duplicates)
ssql read-csv customers.csv | \
  ssql union -file suppliers.csv

# UNION ALL (keep duplicates)
ssql read-csv file1.csv | \
  ssql union -all -file file2.csv -file file3.csv
```

Equivalent SQL:
```sql
SELECT * FROM customers
UNION
SELECT * FROM suppliers
```

---

## Creating Visualizations

Generate interactive HTML charts with Chart.js:

### Simple Chart

```bash
ssql read-csv employees.csv | \
  ssql chart -x department -y salary -output salary_chart.html
```

Opens `salary_chart.html` with an interactive chart featuring:
- Multiple chart types (line, bar, scatter, pie, radar)
- Zoom and pan controls
- Field selection UI
- Data export to PNG/CSV

### Chart with Aggregations

```bash
ssql read-csv employees.csv | \
  ssql group-by department \
    -function avg -field salary -result avg_salary | \
  ssql chart -x department -y avg_salary -output dept_salaries.html
```

---

## Code Generation

Every command supports the `-generate` flag to output Go code instead of executing:

### Generate Code from Pipeline

```bash
ssql read-csv -generate employees.csv | \
  ssql where -generate -match department eq Engineering | \
  ssql include name salary | \
  ssql write-csv -generate output.csv | \
  ssql generate-go
```

Output:
```go
package main

import (
	"github.com/rosscartlidge/ssql"
)

func main() {
	records := ssql.ReadCSV("employees.csv")
	filtered := ssql.Where(func(r ssql.Record) bool {
		return r["department"].(string) == "Engineering"
	})(records)
	selected := ssql.Select(func(r ssql.Record) ssql.Record {
		result := make(ssql.Record)
		result["name"] = r["name"]
		result["salary"] = r["salary"]
		return result
	})(filtered)
	ssql.WriteCSV(selected, "output.csv")
}
```

### Compile and Run Generated Code

```bash
# Generate code to file
ssql read-csv -generate data.csv | \
  ssql group -generate -by region -function sum -field sales -result total | \
  ssql generate-go > analysis.go

# Add package initialization
cat > go.mod << 'EOF'
module analysis
go 1.23
require github.com/rosscartlidge/ssql latest
EOF

# Build and run
go mod tidy
go run analysis.go
```

### Advanced Example: Complex Pipeline with Chain()

When you use multiple transformation commands, the generated code automatically uses `ssql.Chain()` for clean, readable code:

```bash
# Complex pipeline: filter, select, sort, limit
ssql read-csv -generate sales.csv | \
  ssql where -match revenue gt 1000 -generate | \
  ssql include salesperson revenue | \
  ssql sort revenue -desc -generate | \
  ssql limit 10 -generate | \
  ssql write-csv -generate top_performers.csv | \
  ssql generate-go > report.go
```

Generated code (`report.go`):
```go
package main

import (
	"fmt"
	"os"
	"github.com/rosscartlidge/ssql"
)

// asFloat64 converts Record values to float64 for numeric comparisons
// Handles both int64 (from CSV parsing integers) and float64
func asFloat64(v any) float64 {
	switch val := v.(type) {
	case int64:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

func main() {
	records, err := ssql.ReadCSV("sales.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Errorf("reading CSV: %w", err))
		os.Exit(1)
	}

	// Multiple operations composed with Chain()
	result := ssql.Chain(
		ssql.Where(func(r ssql.Record) bool {
			return asFloat64(r["revenue"]) > 1000
		}),
		ssql.Select(func(r ssql.Record) ssql.Record {
			result := make(ssql.Record)
			if val, exists := r["salesperson"]; exists {
				result["salesperson"] = val
			}
			if val, exists := r["revenue"]; exists {
				result["revenue"] = val
			}
			return result
		}),
		ssql.SortBy(func(r ssql.Record) float64 {
			val, _ := r["revenue"]
			switch v := val.(type) {
			case int64:
				return -float64(v)  // Negative for descending
			case float64:
				return -v
			default:
				return 0
			}
		}),
		ssql.Limit[ssql.Record](10),
	)(records)

	ssql.WriteCSV(result, "top_performers.csv")
}
```

Compile and run:
```bash
# Setup and run
go mod init report
go mod tidy
go build -o report report.go
./report

# View results
cat top_performers.csv
```

**Key Features of Generated Code:**
- ✅ **Clean Chain() pattern** - Multiple operations composed functionally
- ✅ **Type-safe helpers** - `asFloat64()` handles both int64 and float64
- ✅ **Proper error handling** - Exit codes and stderr for errors
- ✅ **Production-ready** - Compiles and runs immediately
- ✅ **Readable** - Easy to understand and modify

### Example with Aggregations

Generate code for GROUP BY with multiple aggregations:

```bash
ssql read-csv -generate sales.csv | \
  ssql group-by region \
    -function count -result num_sales + \
    -function sum -field revenue -result total_revenue + \
    -function avg -field revenue -result avg_revenue -generate | \
  ssql sort total_revenue -desc -generate | \
  ssql write-csv -generate region_report.csv | \
  ssql generate-go > region_analysis.go
```

Generated code:
```go
package main

import (
	"fmt"
	"os"
	"github.com/rosscartlidge/ssql"
)

func main() {
	records, err := ssql.ReadCSV("sales.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	grouped := ssql.GroupByFields("_group", "region")(records)

	aggregated := ssql.Aggregate("_group", map[string]ssql.AggregateFunc{
		"num_sales": ssql.Count(),
		"total_revenue": ssql.Sum("revenue"),
		"avg_revenue": ssql.Avg("revenue"),
	})(grouped)

	sorted := ssql.SortBy(func(r ssql.Record) float64 {
		val, _ := r["total_revenue"]
		switch v := val.(type) {
		case int64:
			return -float64(v)
		case float64:
			return -v
		default:
			return 0
		}
	})(aggregated)

	ssql.WriteCSV(sorted, "region_report.csv")
}
```

This workflow enables **rapid prototyping** with the CLI, then **instant production deployment** with generated, type-safe Go code!

---

## Complete Example

Let's build a comprehensive data analysis pipeline:

### Scenario: Analyze Process Counts by User

```bash
# Execute the pipeline
ssql exec -- ps -efl | \
  ssql group-by UID -function count -result process_count | \
  ssql chart -x UID -y process_count -output /tmp/processes_by_user.html
```

This will:
1. Execute `ps -efl` and parse the output
2. Group processes by UID (user)
3. Count processes per user
4. Create an interactive chart

Output: `Chart created: /tmp/processes_by_user.html`

### Generate Production Code

Now convert the same pipeline to Go code:

```bash
ssql exec -generate -- ps -efl | \
  ssql group -generate -by UID -function count -result process_count | \
  ssql chart -generate -x UID -y process_count -output processes.html | \
  ssql generate-go > monitor.go
```

Generated code in `monitor.go`:
```go
package main

import (
	"github.com/rosscartlidge/ssql"
)

func main() {
	records := ssql.ExecCommand("ps", []string{"-efl"})
	grouped := ssql.GroupByFields("_group", "UID")(records)
	aggregated := ssql.Aggregate("_group", map[string]ssql.AggregateFunc{
		"process_count": ssql.Count(),
	})(grouped)
	ssql.QuickChart(aggregated, "UID", "process_count", "processes.html")
}
```

Compile and run:
```bash
# Setup module
go mod init monitor
go get github.com/rosscartlidge/ssql

# Build and run
go build -o monitor monitor.go
./monitor
```

---

## Available Commands

### Data Sources
- `read-csv [file]` - Read CSV file (or stdin)
- `exec -- [command] [args...]` - Execute command and parse output

### Transformations
- `where` - Filter records by conditions
- `select` - Select/rename fields
- `update` - Conditionally update field values (if-elseif-else logic)
- `group` - Group and aggregate data
- `sort` - Sort records by field
- `limit` - Take first N records
- `offset` - Skip first N records (SQL OFFSET)
- `distinct` - Remove duplicate records (SQL DISTINCT)

### Multi-Table Operations
- `join` - Join two data sources (SQL JOIN - inner/left/right/full)
- `union` - Combine multiple data sources (SQL UNION/UNION ALL)

### Outputs
- `write-csv [file]` - Write CSV file (or stdout)
- `table` - Display records as formatted table
- `chart` - Create interactive HTML chart

### Code Generation
- `generate-go` - Assemble code fragments into Go program

### Getting Help

```bash
# Show all commands
ssql -help

# Show command-specific help
ssql read-csv -help
ssql where -help
ssql group -help
ssql chart -help
```

### Bash Completion

The CLI supports intelligent tab completion for commands, flags, and even field names:

```bash
# Install bash completion (for current session)
eval "$(ssql -bash-completion)"

# Install permanently
ssql -bash-completion > ~/.local/share/bash-completion/completions/streamv3

# Or add to ~/.bashrc
echo 'eval "$(ssql -bash-completion)"' >> ~/.bashrc
source ~/.bashrc
```

Now you can tab-complete:
```bash
ssql <TAB>          # Shows all commands
ssql where <TAB>    # Shows flags like -match, -help
ssql read-csv <TAB> # Completes .csv files
```

### Understanding Command Structure (Advanced)

ssql CLI uses the **completionflags** framework for declarative command definitions. This enables powerful features:

**Clause Pattern:**
Commands that support multiple items use `+` as a separator to create "clauses". Each clause can have its own set of flags:

```bash
# Multiple WHERE conditions (OR logic) - each clause after + is independent
ssql where -match age gt 30 + -match salary gt 100000

# Multiple aggregations - each + starts a new aggregation
ssql group-by department \
  -function count -result total + \
  -function avg -field salary -result avg_salary + \
  -function max -field salary -result max_salary
```

**How it works:**
- **Before `+`**: First clause with its flags
- **After `+`**: New clause starts, can have different flags
- **Framework**: Automatically parses and validates all flags
- **Benefits**: Type safety, auto-completion, consistent error messages

**Example breakdown:**
```bash
ssql group-by department \
  -function count -result total + \
  #     └─ Clause 1 ─────────┘   │
  -function avg -field salary -result avg_salary
  #     └────────── Clause 2 ───────────────────┘
```

Each clause is independently validated, so you get clear error messages if flags are missing or incorrect.

**Global vs Local Flags:**
- **Global flags** apply to the entire command: `-by department`, `-generate`
- **Local flags** belong to specific clauses: `-function`, `-field`, `-result`

This pattern makes complex commands readable while maintaining type safety and completion support.

---

## What's Next?

### Workflow: CLI → Code → Production

1. **Prototype with CLI** - Quickly explore your data
   ```bash
   ssql read-csv data.csv | ssql where ... | ssql chart ...
   ```

2. **Generate Code** - Convert to Go when satisfied
   ```bash
   ssql read-csv -generate data.csv | ... | ssql generate-go > app.go
   ```

3. **Refine and Deploy** - Edit generated code, add error handling, deploy
   ```go
   // Add your business logic, error handling, logging, etc.
   ```

### Advanced Topics

- **[API Reference](api-reference.md)** - Full library documentation for refining generated code
- **[Getting Started Guide](codelab-intro.md)** - Learn the ssql library directly
- **[Advanced Tutorial](advanced-tutorial.md)** - Production patterns and optimization

### Recently Added (v0.7.0)

ssql now supports essential SQL operations:
- ✅ `join` - All JOIN types (INNER, LEFT, RIGHT, FULL)
- ✅ `distinct` - Remove duplicates
- ✅ `offset` - Skip N records for pagination
- ✅ `union` - Combine datasets with UNION/UNION ALL
- ✅ `sort` - Sort by single field
- ✅ `limit` - Take first N records

### Coming Soon (Phase 2)

The CLI is actively being developed. Upcoming features:
- Multi-field sorting with mixed ASC/DESC order
- `having` - Post-aggregation filtering (like SQL HAVING)
- More aggregation functions
- Better error messages
- Tab completion improvements

### Need Help?

- **[Debugging Guide](debugging_pipelines.md)** - Learn to debug pipelines with jq
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions
- **[GitHub Issues](https://github.com/rosscartlidge/ssql/issues)** - Report bugs
- **Examples** - Check `examples/` directory
- **API Reference** - Full library documentation

---

*Prototype fast with the CLI, deploy with confidence using generated Go code!* ⚡
