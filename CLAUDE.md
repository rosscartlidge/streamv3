# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Building and Running:**
- `go build` - Build the module
- `go run doc/examples/chart_demo.go` - Run the comprehensive chart demo
- `go test` - Run all tests
- `go test -v` - Run tests with verbose output
- `go test -run TestSpecificFunction` - Run specific test
- `go fmt ./...` - Format all Go code
- `go vet ./...` - Run Go vet for static analysis
- `go mod tidy` - Clean up module dependencies

**Testing:**
- Tests are in `*_test.go` files using standard Go testing
- Main test files: `example_test.go`, `chart_demo_test.go`, `benchmark_test.go`
- No custom test runners or frameworks - use standard `go test`

## Architecture Overview

StreamV3 is a modern Go library built on three core abstractions:

**Core Types:**
- `iter.Seq[T]` and `iter.Seq2[T,error]` - Go 1.23+ iterators (lazy sequences)
- `Record` - Map-based flexible data structure (`map[string]any`)
- `Filter[T,U]` - Composable transformations (`func(iter.Seq[T]) iter.Seq[U]`)

**Key Architecture Files:**
- `core.go` - Core types, Filter functions, Record system, composition functions
- `operations.go` - Stream operations (Map, Where, Reduce, etc.)
- `chart.go` - Interactive Chart.js visualization with Bootstrap 5 UI
- `io.go` - CSV/JSON I/O, command parsing, file operations
- `sql.go` - GROUP BY aggregations and SQL-style operations

**API Design - Functional Composition:**
- **Functional API** - Explicit Filter composition: `Pipe(Where(...), GroupByFields(...), Aggregate(...))`
  - Handles all operations including type-changing operations (GroupBy, Aggregate)
  - Flexible and composable for complex pipelines
  - One clear way to compose operations

**Error Handling:**
- Simple iterators: `iter.Seq[T]`
- Error-aware iterators: `iter.Seq2[T, error]`
- Conversion utilities: `Safe()`, `Unsafe()`, `IgnoreErrors()`

**Data Visualization:**
- Chart.js integration with interactive HTML output
- Field selection UI, zoom/pan, statistical overlays
- Multiple chart types: line, bar, scatter, pie, radar
- Export formats: PNG, CSV

**Entry Points:**
- `slices.Values(slice)` - Create iterator from slice
- `ReadCSV(filename)` - Parse CSV files returning `iter.Seq[Record]`
- `ExecCommand(cmd, args...)` - Parse command output returning `iter.Seq[Record]`
- `QuickChart(data, x, y, filename)` - Generate interactive charts

## API Naming Conventions (SQL-Style)

StreamV3 uses SQL-like naming instead of functional programming conventions. **Always use these canonical names:**

**Stream Operations (operations.go):**
- **`SelectMany`** - Flattens nested sequences (NOT FlatMap)
  - `SelectMany[T, U any](fn func(T) iter.Seq[U]) Filter[T, U]`
  - Use for one-to-many transformations (e.g., splitting records)
- **`Where`** - Filters records based on predicate (NOT Filter)
  - Note: `Filter[T,U]` is the type name for transformations
- **`Select`** - Projects/transforms fields (similar to Map, but SQL-style)
- **`Reduce`** - Aggregates sequence to single value
- **`Take`** - Limits number of records (like SQL LIMIT)
- **`Skip`** - Skips first N records (like SQL OFFSET)

**Aggregation Operations (sql.go):**
- **`GroupByFields`** - Groups and aggregates (SQL GROUP BY)
- **`Aggregate`** - Applies aggregation functions (Count, Sum, Avg, etc.)

**Common Mistakes:**
- ❌ Looking for `FlatMap` → ✅ Use `SelectMany`
- ❌ Using `Filter` as function → ✅ Use `Where` (Filter is a type)
- ❌ Looking for LINQ-style names → ✅ Check operations.go for SQL-style names

When in doubt, check `operations.go` for the canonical API - don't assume LINQ or functional programming naming conventions.

## Canonical Numeric Types (Hybrid Approach)

StreamV3 enforces a **hybrid type system** for clarity and consistency:

**Scalar Values - Canonical Types Only:**
- **Integers**: Always use `int64`, never `int`, `int32`, `uint`, etc.
- **Floats**: Always use `float64`, never `float32`
- **Reason**: Eliminates type conversion ambiguity, consistent with CSV auto-parsing

**Sequence Values - Flexible Types:**
- **Sequences**: Allow all numeric types (`iter.Seq[int]`, `iter.Seq[int32]`, `iter.Seq[float32]`, etc.)
- **Reason**: Works naturally with Go's standard library (`slices.Values([]int{...})`)

**Examples:**
```go
// ✅ CORRECT - Canonical scalar types
record := streamv3.NewRecord().
    Int("count", int64(42)).           // int64 required
    Float("price", 99.99).             // float64 required
    IntSeq("scores", slices.Values([]int{1, 2, 3})).  // iter.Seq[int] allowed
    Build()

// ✅ CORRECT - Type conversion when needed
age := int(streamv3.GetOr(record, "age", int64(0)))

// ❌ WRONG - Non-canonical scalar types
record := streamv3.NewRecord().
    Int("count", 42).                  // Won't compile - int not allowed
    Float("price", float32(99.99)).    // Won't compile - float32 not allowed
    Build()
```

**CSV Auto-Parsing:**
- CSV reader produces `int64` for integers, `float64` for decimals
- Always use `int64(0)` and `float64(0)` as default values with `GetOr()`
- Example: `age := streamv3.GetOr(record, "age", int64(0))`

**Type Conversion:**
- `Get[int64]()` works for string → int64 parsing
- `Get[float64]()` works for string → float64 parsing
- `Get[int]()` will NOT convert from strings (no automatic parsing)
- Users must explicitly convert: `age := int(GetOr(r, "age", int64(0)))`

This hybrid approach balances ergonomics (flexible sequences) with consistency (canonical scalars).

This library emphasizes functional composition with Go 1.23+ iterators while providing comprehensive data visualization capabilities.

## CLI Tools Architecture

StreamV3 includes a CLI tool (`cmd/streamv3`) with a self-generating code architecture.

**Self-Generating Commands:**
- Each CLI command has a `-generate` flag that emits Go code instead of executing
- Commands communicate via JSONL code fragments on stdin/stdout
- The `generate-go` command assembles all fragments into a complete Go program
- This eliminates the need for a separate parser and keeps generated code in sync with implementations

**Code Fragment System (`cmd/streamv3/lib/codefragment.go`):**
- `CodeFragment` struct with Type/Var/Input/Code/Imports fields
- `ReadAllCodeFragments()` - Read all previous fragments from stdin
- `WriteCodeFragment()` - Write fragment to stdout
- Commands pass through all previous fragments, then append their own

**Adding New Commands - REQUIRED STEPS:**
1. Add `Generate bool` flag to config struct with `gs:"flag,global,last,help=Generate Go code instead of executing"`
2. Add branch in Execute() method: `if c.Generate { return c.generateCode(...) }`
3. Implement `generateCode()` method:
   - Call `lib.ReadAllCodeFragments()` to read all previous fragments from stdin
   - Pass through all previous fragments with `lib.WriteCodeFragment(frag)`
   - Get input variable from last fragment: `fragments[len(fragments)-1].Var`
   - Generate your command's Go code
   - Create fragment with `lib.NewStmtFragment()` (or NewInitFragment/NewFinalFragment)
   - Write your fragment with `lib.WriteCodeFragment()`

**Fragment Types:**
- `init` - First command in pipeline (e.g., read-csv), creates initial variable
- `stmt` - Middle command (e.g., where, select), has input and output variable
- `final` - Last command (e.g., write-csv), has input but no output variable

**Example Implementation Pattern:**
```go
type MyCommandConfig struct {
    Generate bool   `gs:"flag,global,last,help=Generate Go code instead of executing"`
    // ... other fields
}

func (c *MyCommandConfig) Execute(ctx context.Context, clauses []gs.ClauseSet) error {
    if c.Generate {
        return c.generateCode(clauses)
    }
    // ... normal execution
}

func (c *MyCommandConfig) generateCode(clauses []gs.ClauseSet) error {
    // Read all previous fragments
    fragments, err := lib.ReadAllCodeFragments()
    if err != nil {
        return fmt.Errorf("reading code fragments: %w", err)
    }

    // Pass through all previous fragments
    for _, frag := range fragments {
        if err := lib.WriteCodeFragment(frag); err != nil {
            return fmt.Errorf("writing previous fragment: %w", err)
        }
    }

    // Get input variable from last fragment
    var inputVar string
    if len(fragments) > 0 {
        inputVar = fragments[len(fragments)-1].Var
    } else {
        inputVar = "records"
    }

    // Generate your code
    code := fmt.Sprintf("output := myFunc(%s)", inputVar)

    // Create and write your fragment
    frag := lib.NewStmtFragment("output", inputVar, code, []string{"import1", "import2"})
    return lib.WriteCodeFragment(frag)
}
```

**Testing Code Generation:**
```bash
# Test your command's generation
streamv3 read-csv -generate data.csv | streamv3 mycommand -generate | streamv3 generate-go

# The output should be valid Go code that compiles
```

**⚠️ CRITICAL: Every new CLI command MUST implement -generate support, or code generation pipelines will break!**

- ai_generation
- doc_improvement
- llm_test
- llm_eval
- gemini_works
- llm_cli
- cli-gs-tools
- code_generation
- pattern
- source_sink_consistent