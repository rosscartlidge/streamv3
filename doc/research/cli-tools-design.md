# StreamV3 CLI Tools Design Document

## Executive Summary

This document outlines the design and implementation of command-line tools for StreamV3, enabling Unix-style pipelines for data processing. The CLI will leverage the existing gs (gogstools) framework for declarative command definition, intelligent tab completion, and clause-based filtering.

**Example Pipeline:**
```bash
cat data.csv |
  streamv3 read-csv |
  streamv3 where -match age gt 18 -match status eq active |
  streamv3 group-by -fields age + -aggregate 'count=count()' + -aggregate 'avg_score=avg(score)' |
  streamv3 sort -field count -desc |
  streamv3 write-csv > output.csv
```

## Goals

1. **Unix Philosophy**: Small, composable tools that do one thing well
2. **Streaming Performance**: Process data in constant memory regardless of size
3. **Developer Experience**: Intelligent tab completion for fields, operators, and values
4. **Type Safety**: Maintain StreamV3's type system in CLI tools
5. **Discoverability**: Self-documenting with -help, -man, -bash-completion

## Architecture

### Binary Structure

**Single binary with subcommands**: `streamv3 <command> [options]`

- Single `cmd/streamv3/main.go` entry point
- Each command in separate package: `cmd/streamv3/commands/<command>/`
- Shared utilities in `cmd/streamv3/lib/`

**Benefits:**
- Single installation: `go install github.com/rosscartlidge/streamv3/cmd/streamv3@latest`
- Consistent help system: `streamv3 -help` shows all commands
- Shared completion script: `streamv3 -bash-completion`

### Data Interchange Format: JSONL (JSON Lines)

**Format:** One JSON object per line, newline-delimited
```jsonl
{"name":"Alice","age":30,"email":"alice@example.com"}
{"name":"Bob","age":25,"email":"bob@example.com"}
{"name":"Carol","age":35,"email":"carol@example.com"}
```

**Advantages:**
1. **Streaming-friendly**: Parse line-by-line, no need to load entire file
2. **Type preservation**: int64, float64, bool, null preserved (unlike CSV)
3. **Nested structures**: Supports Records and arrays
4. **Error handling**: Malformed line doesn't corrupt entire stream
5. **Human-readable**: Easy to debug with `head`, `tail`, `grep`
6. **Standard**: Well-supported format with libraries in all languages

**Canonical Types in JSONL:**
- Integers: Always encode as JSON numbers representing int64
- Floats: Always encode as JSON numbers representing float64
- Strings: JSON strings
- Booleans: JSON true/false
- Null: JSON null for missing values
- Arrays: JSON arrays for iter.Seq values
- Objects: JSON objects for nested Records

## Core Commands (Phase 1)

### 1. read-csv
**Purpose:** Convert CSV to JSONL stream

**Usage:**
```bash
streamv3 read-csv data.csv
cat data.csv | streamv3 read-csv
```

**Config:**
```go
type ReadCSVConfig struct {
    HasHeader bool   `gs:"flag,global,last,help=CSV has header row,default=true"`
    Separator string `gs:"string,global,last,help=Field separator,default=,"`
    Argv      string `gs:"file,global,last,help=Input CSV file,suffix=.csv"`
}
```

**Features:**
- Auto-detect numeric types (consistent with StreamV3.ReadCSV)
- Read from file or stdin
- Output JSONL to stdout

### 2. write-csv
**Purpose:** Convert JSONL stream to CSV

**Usage:**
```bash
streamv3 write-csv output.csv
cat data.jsonl | streamv3 write-csv
```

**Config:**
```go
type WriteCSVConfig struct {
    Header    bool   `gs:"flag,global,last,help=Include header row,default=true"`
    Separator string `gs:"string,global,last,help=Field separator,default=,"`
    Argv      string `gs:"file,global,last,help=Output CSV file,suffix=.csv"`
}
```

**Features:**
- Flatten Records to CSV rows
- Write to file or stdout
- Auto-detect field order from first record

### 3. where
**Purpose:** Filter records with boolean conditions

**Usage:**
```bash
# Single condition
streamv3 where -match age gt 18

# Multiple AND conditions (multiple -match in same command)
streamv3 where -match age gt 18 -match status eq active

# OR conditions (+ separator starts new clause)
streamv3 where -match age gt 65 + -match age lt 18

# Complex: (A AND B) OR C
streamv3 where -match age gt 18 -match status eq active + -match department eq Engineering
```

**Config:**
```go
type WhereConfig struct {
    // Match uses multi-argument pattern: -match field op value
    // Multiple -match in same clause are ANDed, separate clauses (+) are ORed
    Match []map[string]interface{} `gs:"multi,local,list,args=field:op:value,help=Filter condition: field operator value"`
    Argv  string                   `gs:"file,global,last,help=Input JSONL file (or stdin if not specified),suffix=.jsonl"`
}
```

**Operators:**
- `eq`, `ne`: Equality for all types
- `gt`, `ge`, `lt`, `le`: Comparison for numbers and strings
- `contains`, `startswith`, `endswith`: String matching

**Implementation:**
- Each clause = one Where() filter
- Multiple clauses = Pipe(where1, where2, ...)
- `+` separator = OR logic (separate pipelines merged)

### 4. select
**Purpose:** Project and transform fields

**Usage:**
```bash
# Select specific fields (use + to separate multiple fields)
streamv3 select -field name + -field age

# Rename fields
streamv3 select -field name -as fullname + -field age

# Multiple field selection
streamv3 select -field name + -field age + -field department
```

**Config:**
```go
type SelectConfig struct {
    Field string `gs:"field,local,last,help=Field to select"`
    As    string `gs:"string,local,last,help=Rename field to"`
    Expr  string `gs:"string,local,last,help=Expression to compute"`
    Argv  string `gs:"file,global,last,help=Input JSONL file,suffix=.jsonl"`
}
```

**Implementation:**
- Maps to StreamV3.Select() for field projection
- Each clause specifies one output field

### 5. group-by
**Purpose:** GROUP BY with aggregations

**Usage:**
```bash
# Group by single field with aggregation
streamv3 group-by - fields department - agg 'count=count()' - agg 'avg_salary=avg(salary)'

# Group by multiple fields
streamv3 group-by - fields department - fields location - agg 'total=sum(sales)'
```

**Config:**
```go
type GroupByConfig struct {
    Fields []string `gs:"field,local,list,help=Fields to group by"`
    Agg    string   `gs:"string,local,last,help=Aggregation expression (name=func(field))"`
    Argv   string   `gs:"file,global,last,help=Input JSONL file,suffix=.jsonl"`
}
```

**Aggregation Functions:**
- `count()`: Count records
- `sum(field)`: Sum numeric field
- `avg(field)`: Average numeric field
- `min(field)`: Minimum value
- `max(field)`: Maximum value
- `first(field)`: First value
- `last(field)`: Last value

**Implementation:**
- Parse aggregation expressions: `name=func(field)`
- Maps to StreamV3.GroupByFields() + Aggregate()

### 6. sort
**Purpose:** Sort records by fields

**Usage:**
```bash
# Sort ascending
streamv3 sort -field age

# Sort descending
streamv3 sort -field age -desc

# Multi-field sort (future enhancement)
streamv3 sort -field department + -field age -desc
```

**Config:**
```go
type SortConfig struct {
    Field string `gs:"field,local,last,help=Field to sort by"`
    Desc  bool   `gs:"flag,local,last,help=Sort descending"`
    Argv  string `gs:"file,global,last,help=Input JSONL file,suffix=.jsonl"`
}
```

**Implementation:**
- Maps to StreamV3.SortBy() with field extraction
- Descending = negate numeric values

### 7. limit
**Purpose:** Take first N records

**Usage:**
```bash
streamv3 limit -n 100
```

**Config:**
```go
type LimitConfig struct {
    N    int    `gs:"number,global,last,help=Number of records to take"`
    Argv string `gs:"file,global,last,help=Input JSONL file,suffix=.jsonl"`
}
```

### 8. distinct
**Purpose:** Remove duplicate records

**Usage:**
```bash
# Distinct by all fields
streamv3 distinct

# Distinct by specific fields
streamv3 distinct - field name - field email
```

**Config:**
```go
type DistinctConfig struct {
    Field []string `gs:"field,local,list,help=Fields to check for uniqueness (empty = all fields)"`
    Argv  string   `gs:"file,global,last,help=Input JSONL file,suffix=.jsonl"`
}
```

**Implementation:**
- No fields specified = StreamV3.Distinct() on entire record
- Fields specified = Distinct by field subset

## gs Framework Enhancements

### 1. JSONL Field Completion

**Current:** gs reads TSV files for field completion
**Enhancement:** Add JSONL support

```go
// In gs/command.go
func (cmd *GSCommand) readJSONLFields(filename string) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Read first record to extract field names
    scanner := bufio.NewScanner(file)
    if !scanner.Scan() {
        return nil, fmt.Errorf("empty JSONL file")
    }

    var record map[string]interface{}
    if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
        return nil, err
    }

    fields := make([]string, 0, len(record))
    for k := range record {
        fields = append(fields, k)
    }
    sort.Strings(fields)
    return fields, nil
}
```

### 2. Stdin Field Completion

**Problem:** When piping data, completion needs to read from stdin without consuming it

**Solution:** Support `-` as filename to read from stdin, cache field names

```go
// In gs/command.go
type CompletionContext struct {
    InputFile   string   // File or "-" for stdin
    CachedFields []string // Cached from stdin peek
}

func (cmd *GSCommand) detectFields(filename string) ([]string, error) {
    if filename == "-" || filename == "" {
        // Peek stdin for field names (read first line, then rewind)
        return cmd.peekStdinFields()
    }

    // Auto-detect format: .csv, .jsonl, .tsv
    if strings.HasSuffix(filename, ".jsonl") {
        return cmd.readJSONLFields(filename)
    }
    // ... existing TSV/CSV logic
}
```

### 3. Dynamic Operator Completion

**Enhancement:** Context-aware operator suggestions based on field type

```go
// In gs/types.go
type FieldMeta struct {
    // ... existing fields ...
    OperatorEnum func(fieldName string) []string `json:"-"` // Dynamic operator list
}

// In command.go
func (cfg *WhereConfig) GetOperators(fieldName string) []string {
    // Could inspect field type in data to suggest relevant operators
    // For now, return all operators
    return []string{"eq", "ne", "gt", "ge", "lt", "le", "contains", "startswith", "endswith"}
}
```

### 4. Generalized Input Detection

**Enhancement:** Auto-detect stdin vs file input

```go
// In cmd/streamv3/lib/input.go
func OpenInput(filename string) (iter.Seq[streamv3.Record], error) {
    var reader io.Reader

    if filename == "" || filename == "-" {
        // Check if stdin has data
        stat, _ := os.Stdin.Stat()
        if (stat.Mode() & os.ModeCharDevice) == 0 {
            reader = os.Stdin
        } else {
            return nil, fmt.Errorf("no input provided (use file or pipe data to stdin)")
        }
    } else {
        file, err := os.Open(filename)
        if err != nil {
            return nil, err
        }
        reader = file
    }

    return ReadJSONL(reader)
}
```

### 5. Clause-to-Filter Composition

**Enhancement:** Helper to convert gs.ClauseSet to StreamV3 Filter

```go
// In cmd/streamv3/lib/compose.go
type ClauseFilter func(clause gs.ClauseSet) streamv3.Filter[streamv3.Record, streamv3.Record]

func ComposeFilters(clauses []gs.ClauseSet, makeFilter ClauseFilter) streamv3.Filter[streamv3.Record, streamv3.Record] {
    if len(clauses) == 0 {
        return func(seq iter.Seq[streamv3.Record]) iter.Seq[streamv3.Record] {
            return seq // Identity filter
        }
    }

    // Each clause = one filter, compose with Pipe()
    filters := make([]streamv3.Filter[streamv3.Record, streamv3.Record], len(clauses))
    for i, clause := range clauses {
        filters[i] = makeFilter(clause)
    }

    return streamv3.Pipe(filters...)
}
```

## Implementation Plan

### Phase 1: Foundation (Week 1)
1. **gs Framework Enhancements**
   - Add JSONL field reading
   - Add stdin detection and peeking
   - Add dynamic operator completion
   - Test with existing gogstools examples

2. **CLI Infrastructure**
   - Create `cmd/streamv3/main.go` with command router
   - Create `cmd/streamv3/lib/` with shared utilities:
     - `input.go`: JSONL reading from stdin/file
     - `output.go`: JSONL writing to stdout/file
     - `compose.go`: Clause-to-Filter helpers
   - Implement subcommand registration system

3. **Basic I/O Commands**
   - Implement `read-csv` command
   - Implement `write-csv` command
   - Test: `cat data.csv | streamv3 read-csv | streamv3 write-csv`

### Phase 2: Core Filtering (Week 2)
1. **where command**
   - Implement all operators (eq, ne, gt, lt, ...)
   - Support AND logic (multiple clauses with -)
   - Support OR logic (+ separator)
   - Test complex conditions

2. **select command**
   - Field projection
   - Field renaming
   - Test: `streamv3 read-csv data.csv | streamv3 select - field name - field age`

3. **limit command**
   - Simple record limiting
   - Test with large datasets

### Phase 3: Aggregation (Week 3)
1. **group-by command**
   - Parse aggregation expressions
   - Implement all aggregation functions
   - Support multi-field grouping
   - Test: `streamv3 group-by - fields dept - agg 'count=count()' - agg 'avg=avg(salary)'`

2. **sort command**
   - Single and multi-field sorting
   - Ascending and descending
   - Test with numeric and string fields

3. **distinct command**
   - Full record deduplication
   - Field-specific deduplication

### Phase 4: Polish (Week 4)
1. **Documentation**
   - Command reference docs
   - Tutorial with real-world examples
   - Performance benchmarks

2. **Testing**
   - Integration tests for each command
   - End-to-end pipeline tests
   - Performance tests with large datasets

3. **Completion Enhancement**
   - Ensure field completion works in all commands
   - Test operator completion in where
   - Test value completion in where

## Example Use Cases

### Use Case 1: Data Cleaning
```bash
# Remove invalid records, select relevant fields, deduplicate
cat users.csv |
  streamv3 read-csv |
  streamv3 where -match age gt 0 -match email contains '@' |
  streamv3 select -field name + -field email + -field age |
  streamv3 distinct -field email |
  streamv3 write-csv > clean_users.csv
```

### Use Case 2: Analytics
```bash
# Find top Engineering employees by salary
cat employees.csv |
  streamv3 read-csv |
  streamv3 where -match department eq Engineering |
  streamv3 select -field name + -field salary + -field age |
  streamv3 sort -field salary -desc |
  streamv3 limit -n 10 |
  streamv3 write-csv > top_engineers.csv
```

### Use Case 3: Complex Filtering
```bash
# Find employees who are either senior (age > 40) or high earners (salary > 100k)
streamv3 read-csv employees.csv |
  streamv3 where -match age gt 40 + -match salary gt 100000 |
  streamv3 sort -field salary -desc |
  streamv3 write-csv > senior_or_high_earners.csv
```

### Use Case 4: Multi-criteria Filtering
```bash
# Find young Engineering employees OR any Sales employees
streamv3 read-csv employees.csv |
  streamv3 where -match age lt 30 -match department eq Engineering + -match department eq Sales |
  streamv3 select -field name + -field department + -field age |
  streamv3 write-csv > filtered_employees.csv
```

## Design Decisions

### Why JSONL instead of CSV for interchange?
- **Type preservation**: CSV loses type information (everything is a string)
- **Streaming**: Both are line-delimited and stream-friendly
- **Nested data**: JSONL supports Records and arrays
- **Error handling**: Malformed line in JSONL doesn't corrupt stream

### Why single binary with subcommands?
- **Installation**: Single `go install` command
- **Discoverability**: `streamv3 -help` shows all commands
- **Consistency**: Shared flags (-help, -bash-completion, -man)
- **Completion**: Single completion script for all commands

### Why clause-based syntax with - and +?
- **Composability**: Matches StreamV3's Pipe() mental model
- **Readability**: Clear separation of arguments
- **Power**: OR logic (+) and AND logic (-) in same command
- **Completion**: gs framework handles parsing automatically

### Why enhance gs framework instead of starting fresh?
- **Proven design**: Already works well in gogstools
- **Tab completion**: Field and content completion already implemented
- **Clause parsing**: Handles - and + separators correctly
- **Extensibility**: Easy to add new field types and completion modes
- **No users yet**: Free to modify without breaking changes

## Code Generation: CLI to Production Go

### Overview

One of the most powerful features of the StreamV3 CLI is the ability to **prototype pipelines interactively**, then convert them to **production-quality Go code** for deployment. This workflow combines the rapid iteration of CLI tools with the performance and type safety of compiled Go programs.

**Workflow:**
1. **Prototype**: Use CLI to explore data and build pipeline interactively
2. **Test**: Verify pipeline produces correct results with real data
3. **Convert**: Generate equivalent Go code using StreamV3 library
4. **Optimize**: Add type safety, error handling, custom logic
5. **Deploy**: Compile single binary, 10-100x faster than CLI

### Why Code Generation?

**CLI Advantages (Prototyping):**
- Fast iteration: No compile step
- Interactive: Tab completion, immediate feedback
- Exploratory: Easy to try different operations
- Low overhead: Perfect for one-off data tasks

**Generated Code Advantages (Production):**
- **Performance**: 10-100x faster (no process spawning, JSON parsing per command)
- **Type safety**: Compile-time checks, custom types
- **Single binary**: No shell dependencies, easy deployment
- **Customization**: Add business logic, error handling, monitoring
- **Maintainability**: Code review, version control, testing

### CLI-to-Go Mapping

The StreamV3 CLI has a **1:1 mapping** with the library API, making code generation straightforward:

#### Example 1: Simple Filter Pipeline

**CLI Pipeline:**
```bash
streamv3 read-csv data.csv | \
  streamv3 where -field age -op gt -value 18 | \
  streamv3 select -field name -field email | \
  streamv3 limit -n 100 | \
  streamv3 write-csv output.csv
```

**Generated Go Code:**
```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read CSV
    records := streamv3.ReadCSV("data.csv")

    // Where: age > 18
    filtered := streamv3.Where(func(r streamv3.Record) bool {
        if age, ok := r["age"].(float64); ok {
            return age > 18
        }
        return false
    })(records)

    // Select: name, email
    selected := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        return streamv3.Record{
            "name": r["name"],
            "email": r["email"],
        }
    })(filtered)

    // Limit: 100
    limited := streamv3.Limit[streamv3.Record](100)(selected)

    // Write CSV
    if err := streamv3.WriteCSV("output.csv", limited); err != nil {
        log.Fatalf("Error writing CSV: %v", err)
    }
}
```

#### Example 2: Group By with Aggregation

**CLI Pipeline:**
```bash
streamv3 read-csv sales.csv | \
  streamv3 group-by -fields region -fields product \
    -agg 'total=sum(amount)' -agg 'count=count()' | \
  streamv3 sort -field total -desc | \
  streamv3 limit -n 10 | \
  streamv3 write-csv top_sales.csv
```

**Generated Go Code:**
```go
package main

import (
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Read CSV
    records := streamv3.ReadCSV("sales.csv")

    // Group by: region, product
    grouped := streamv3.GroupByFields("region", "product")(records)

    // Aggregate: total=sum(amount), count=count()
    aggregated := streamv3.Aggregate(
        streamv3.Sum("amount", "total"),
        streamv3.Count("count"),
    )(grouped)

    // Flatten groups to records
    flattened := streamv3.FlattenGroups(aggregated)

    // Sort by: total descending
    sorted := streamv3.SortBy(func(r streamv3.Record) float64 {
        if total, ok := r["total"].(float64); ok {
            return -total // Negative for descending
        }
        return 0
    })(flattened)

    // Limit: 10
    limited := streamv3.Limit[streamv3.Record](10)(sorted)

    // Write CSV
    if err := streamv3.WriteCSV("top_sales.csv", limited); err != nil {
        log.Fatalf("Error writing CSV: %v", err)
    }
}
```

#### Example 3: Complex Where Conditions

**CLI Pipeline:**
```bash
streamv3 read-csv users.csv | \
  streamv3 where -field age -op gt -value 18 -field status -op eq -value active | \
  streamv3 write-csv adult_active.csv
```

**Generated Go Code:**
```go
package main

import (
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    records := streamv3.ReadCSV("users.csv")

    // Where: age > 18 AND status = active
    filtered := streamv3.Where(func(r streamv3.Record) bool {
        age, ageOk := r["age"].(float64)
        status, statusOk := r["status"].(string)

        return ageOk && age > 18 &&
               statusOk && status == "active"
    })(records)

    if err := streamv3.WriteCSV("adult_active.csv", filtered); err != nil {
        log.Fatalf("Error writing CSV: %v", err)
    }
}
```

### Implementation Architecture

There are three approaches to implementing code generation:

#### Approach 1: Simple Code Generator (Recommended for MVP)

**Command:** `streamv3 generate-go`

**Usage:**
```bash
# Generate from shell history
streamv3 generate-go < pipeline.sh > main.go

# Or inline
echo 'streamv3 read-csv data.csv | streamv3 where -field age -op gt -value 18' | \
  streamv3 generate-go > main.go

# Then compile and run
go mod init myproject
go get github.com/rosscartlidge/streamv3
go build -o myproject main.go
./myproject
```

**Implementation:**
```go
// cmd/streamv3/commands/generatego.go
type GenerateGoConfig struct {
    Output string `gs:"file,global,last,help=Output Go file,suffix=.go"`
}

func (c *generateGoCommand) Execute(ctx context.Context, args []string) error {
    // Parse pipeline from stdin or args
    pipeline, err := parsePipeline(os.Stdin)
    if err != nil {
        return fmt.Errorf("parsing pipeline: %w", err)
    }

    // Generate Go code
    code, err := generateGoCode(pipeline)
    if err != nil {
        return fmt.Errorf("generating code: %w", err)
    }

    // Write to output
    if c.config.Output != "" {
        return os.WriteFile(c.config.Output, []byte(code), 0644)
    }
    fmt.Println(code)
    return nil
}
```

#### Approach 2: Interactive Pipeline Recorder

**Command:** `streamv3 record-pipeline`

**Usage:**
```bash
# Start recording
streamv3 record-pipeline start my-pipeline

# Run commands normally (recorded in background)
streamv3 read-csv data.csv | streamv3 where -field age -op gt -value 18

# Stop and generate code
streamv3 record-pipeline stop
streamv3 record-pipeline generate my-pipeline.go
```

**Implementation:** Records commands to temporary file, converts on demand.

#### Approach 3: Built-in Generation Flag

**Usage:**
```bash
# Run pipeline normally, but also generate code
streamv3 --generate-go=output.go \
  read-csv data.csv | \
  streamv3 where -field age -op gt -value 18 | \
  streamv3 write-csv result.csv
```

**Implementation:** Each command logs itself to shared context, main() generates code at exit.

### Code Generation Engine

**Core Components:**

1. **Pipeline Parser**: Parse shell command into structured representation
2. **AST Builder**: Build abstract syntax tree of operations
3. **Code Generator**: Convert AST to Go code with proper imports, types, error handling

**Pipeline Parser:**
```go
// lib/pipeline.go
type Command struct {
    Name   string            // "read-csv", "where", "limit"
    Args   map[string]string // Flag values
    Stdin  bool              // True if reads from stdin
    Stdout bool              // True if writes to stdout
}

type Pipeline struct {
    Commands []Command
}

func ParsePipeline(input string) (*Pipeline, error) {
    // Split by | and parse each command
    parts := strings.Split(input, "|")
    commands := make([]Command, 0, len(parts))

    for _, part := range parts {
        cmd, err := parseCommand(strings.TrimSpace(part))
        if err != nil {
            return nil, err
        }
        commands = append(commands, cmd)
    }

    return &Pipeline{Commands: commands}, nil
}

func parseCommand(cmdLine string) (Command, error) {
    // Parse: streamv3 where -field age -op gt -value 18
    fields := strings.Fields(cmdLine)
    if len(fields) < 2 || fields[0] != "streamv3" {
        return Command{}, fmt.Errorf("invalid command: %s", cmdLine)
    }

    cmd := Command{
        Name: fields[1],
        Args: make(map[string]string),
    }

    // Parse flags
    for i := 2; i < len(fields); i += 2 {
        if !strings.HasPrefix(fields[i], "-") {
            continue
        }
        flag := strings.TrimPrefix(fields[i], "-")
        if i+1 < len(fields) {
            cmd.Args[flag] = fields[i+1]
        }
    }

    return cmd, nil
}
```

**Code Generator:**
```go
// lib/codegen.go
type CodeGenerator struct {
    buf    *bytes.Buffer
    indent int
}

func GenerateGoCode(pipeline *Pipeline) (string, error) {
    g := &CodeGenerator{buf: new(bytes.Buffer)}

    // Generate package and imports
    g.writePackage()
    g.writeImports(pipeline)
    g.writeMainFunc(pipeline)

    return g.buf.String(), nil
}

func (g *CodeGenerator) writeMainFunc(p *Pipeline) {
    g.writeln("func main() {")
    g.indent++

    varName := "records"
    for i, cmd := range p.Commands {
        switch cmd.Name {
        case "read-csv":
            g.writeln(fmt.Sprintf("%s := streamv3.ReadCSV(%q)",
                varName, cmd.Args["file"]))

        case "where":
            newVar := fmt.Sprintf("filtered%d", i)
            g.writeWhere(varName, newVar, cmd)
            varName = newVar

        case "limit":
            newVar := fmt.Sprintf("limited%d", i)
            g.writeln(fmt.Sprintf("%s := streamv3.Limit[streamv3.Record](%s)(%s)",
                newVar, cmd.Args["n"], varName))
            varName = newVar

        case "write-csv":
            g.writeWriteCSV(varName, cmd)
        }
    }

    g.indent--
    g.writeln("}")
}

func (g *CodeGenerator) writeWhere(input, output string, cmd Command) {
    field := cmd.Args["field"]
    op := cmd.Args["op"]
    value := cmd.Args["value"]

    g.writeln(fmt.Sprintf("%s := streamv3.Where(func(r streamv3.Record) bool {", output))
    g.indent++

    // Generate comparison based on operator
    switch op {
    case "gt":
        g.writeln(fmt.Sprintf("if val, ok := r[%q].(float64); ok {", field))
        g.indent++
        g.writeln(fmt.Sprintf("return val > %s", value))
        g.indent--
        g.writeln("}")
    case "eq":
        g.writeln(fmt.Sprintf("return r[%q] == %q", field, value))
    // ... other operators
    }

    g.writeln("return false")
    g.indent--
    g.writeln(fmt.Sprintf("})(%s)", input))
}
```

**Generated Code Template:**
```go
package main

import (
    "fmt"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Generated pipeline code here
    {{range .Commands}}
    {{.Generate}}
    {{end}}
}
```

### Command-Specific Generation

**Conversion Complexity:**

| Command | Complexity | Notes |
|---------|-----------|-------|
| read-csv | Trivial | Direct function call |
| write-csv | Trivial | Direct function call |
| where | Simple | Generate comparison function |
| select | Simple | Generate projection function |
| limit | Trivial | Direct function call with type param |
| sort | Simple | Generate sort key function |
| group-by | Moderate | Parse aggregation expressions |
| distinct | Simple | Direct function call |

**Where Command Generation:**

```go
func generateWhere(cmd Command) string {
    field := cmd.Args["field"]
    op := cmd.Args["op"]
    value := cmd.Args["value"]

    // Detect value type
    valueType := detectType(value)

    switch op {
    case "eq", "ne":
        return fmt.Sprintf(
            `streamv3.Where(func(r streamv3.Record) bool {
                return r[%q] %s %s
            })`, field, opToSymbol(op), formatValue(value, valueType))

    case "gt", "ge", "lt", "le":
        return fmt.Sprintf(
            `streamv3.Where(func(r streamv3.Record) bool {
                if val, ok := r[%q].(float64); ok {
                    return val %s %s
                }
                return false
            })`, field, opToSymbol(op), value)

    case "contains":
        return fmt.Sprintf(
            `streamv3.Where(func(r streamv3.Record) bool {
                if val, ok := r[%q].(string); ok {
                    return strings.Contains(val, %q)
                }
                return false
            })`, field, value)
    }
}
```

**Group-By Command Generation:**

```go
func generateGroupBy(cmd Command) string {
    fields := strings.Split(cmd.Args["fields"], ",")
    aggs := parseAggregations(cmd.Args["agg"])

    var buf bytes.Buffer

    // Generate GroupByFields
    buf.WriteString("grouped := streamv3.GroupByFields(")
    for i, field := range fields {
        if i > 0 {
            buf.WriteString(", ")
        }
        buf.WriteString(fmt.Sprintf("%q", field))
    }
    buf.WriteString(")(records)\n")

    // Generate Aggregate
    buf.WriteString("aggregated := streamv3.Aggregate(\n")
    for i, agg := range aggs {
        if i > 0 {
            buf.WriteString(",\n")
        }
        buf.WriteString("    " + generateAggFunc(agg))
    }
    buf.WriteString("\n)(grouped)\n")

    // Flatten
    buf.WriteString("records = streamv3.FlattenGroups(aggregated)\n")

    return buf.String()
}

func generateAggFunc(agg Aggregation) string {
    switch agg.Func {
    case "count":
        return fmt.Sprintf("streamv3.Count(%q)", agg.OutputName)
    case "sum":
        return fmt.Sprintf("streamv3.Sum(%q, %q)", agg.Field, agg.OutputName)
    case "avg":
        return fmt.Sprintf("streamv3.Avg(%q, %q)", agg.Field, agg.OutputName)
    case "min":
        return fmt.Sprintf("streamv3.Min(%q, %q)", agg.Field, agg.OutputName)
    case "max":
        return fmt.Sprintf("streamv3.Max(%q, %q)", agg.Field, agg.OutputName)
    }
    return ""
}
```

### Prototype to Production Workflow

**Step 1: Interactive Prototyping**
```bash
# Explore data structure
streamv3 read-csv data.csv | head -5

# Try different filters
streamv3 read-csv data.csv | streamv3 where -field age -op gt -value 18 | head
streamv3 read-csv data.csv | streamv3 where -field status -op eq -value active | head

# Build full pipeline
streamv3 read-csv data.csv | \
  streamv3 where -field age -op gt -value 18 | \
  streamv3 select -field name -field email | \
  streamv3 write-csv > output.csv
```

**Step 2: Save Working Pipeline**
```bash
# Save to script
cat > pipeline.sh <<'EOF'
streamv3 read-csv data.csv | \
  streamv3 where -field age -op gt -value 18 | \
  streamv3 select -field name -field email | \
  streamv3 write-csv
EOF

chmod +x pipeline.sh
./pipeline.sh > output.csv
```

**Step 3: Generate Go Code**
```bash
# Generate production code
streamv3 generate-go < pipeline.sh > main.go

# Review generated code
cat main.go

# Initialize Go module
go mod init myproject
go get github.com/rosscartlidge/streamv3@latest

# Build binary
go build -o myproject main.go
```

**Step 4: Optimize and Deploy**
```go
// Customize generated code
package main

import (
    "flag"
    "log"
    "github.com/rosscartlidge/streamv3"
)

func main() {
    // Add command-line flags
    inputFile := flag.String("input", "data.csv", "Input CSV file")
    outputFile := flag.String("output", "output.csv", "Output CSV file")
    minAge := flag.Float64("min-age", 18, "Minimum age")
    flag.Parse()

    // Read CSV
    records := streamv3.ReadCSV(*inputFile)

    // Filter by age (using variable instead of hardcoded value)
    filtered := streamv3.Where(func(r streamv3.Record) bool {
        if age, ok := r["age"].(float64); ok {
            return age > *minAge
        }
        return false
    })(records)

    // Select fields
    selected := streamv3.Select(func(r streamv3.Record) streamv3.Record {
        return streamv3.Record{
            "name": r["name"],
            "email": r["email"],
        }
    })(filtered)

    // Write CSV
    if err := streamv3.WriteCSV(*outputFile, selected); err != nil {
        log.Fatalf("Error writing CSV: %v", err)
    }
}
```

**Step 5: Benchmark Performance**
```bash
# Time CLI version
time ./pipeline.sh > /dev/null
# Result: 2.5 seconds

# Time compiled version
time ./myproject
# Result: 0.08 seconds (31x faster!)
```

### Benefits Summary

**Development Speed:**
- CLI prototype in minutes
- Generated code in seconds
- Optimization as needed

**Performance Gains:**
- No process spawning overhead
- No JSONL parsing between commands
- Single-pass processing where possible
- Compiled optimizations

**Production Benefits:**
- Single binary deployment
- Type-safe custom logic
- Integrated error handling
- Version controlled source code
- Easy to test and maintain

**Typical Performance Improvements:**
- Simple pipelines (read → filter → write): **10-30x faster**
- Complex pipelines (group-by, aggregations): **50-100x faster**
- Large datasets (>1GB): **100-200x faster** (memory efficiency)

## Success Metrics

1. **Performance**: Process 1GB CSV file in < 10 seconds on standard hardware
2. **Memory**: Constant memory usage regardless of input size (streaming)
3. **Composability**: All commands work in pipelines with JSONL
4. **Usability**: Tab completion for fields, operators, and values works
5. **Correctness**: All operations match StreamV3 library behavior exactly
6. **Code Generation**: Generate working Go code from any valid pipeline
7. **Performance Gain**: Generated code 10-100x faster than CLI equivalent

## Future Enhancements (Phase 2)

1. **join command**: Merge two JSONL streams
2. **Expression evaluation**: `select - expr 'age * 2 + 10' - as computed`
3. **Window functions**: `over - partition_by dept - order_by salary`
4. **Sampling**: `sample - rate 0.1` (10% of records)
5. **Statistics**: `stats - field salary` (min, max, avg, stddev, percentiles)
6. **chart command**: Generate Chart.js visualizations from CLI
7. **Parallel processing**: `parallel - workers 8` for CPU-intensive operations
8. **SQL mode**: `streamv3 sql "SELECT name, age FROM stdin WHERE age > 18"`

## References

- StreamV3 Library: `/home/rossc/src/streamv3/`
- gs Framework: `/home/rossc/tsv/gogstools/gs/`
- JSONL Specification: https://jsonlines.org/
- Unix Philosophy: https://en.wikipedia.org/wiki/Unix_philosophy
