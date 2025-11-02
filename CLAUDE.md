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

## Release Process

**⚠️ CRITICAL: Version is manually maintained in version.txt**

Version is stored in `cmd/streamv3/version/version.txt` and MUST be updated before creating tags.

**Correct Release Workflow (CRITICAL - Follow Exact Order):**

```bash
# 1. Make all code changes and commit them
git add .
git commit -m "Description of changes"

# 2. Update version.txt (WITHOUT "v" prefix)
echo "X.Y.Z" > cmd/streamv3/version/version.txt

# 3. Commit the version change
git add cmd/streamv3/version/version.txt
git commit -m "Bump version to vX.Y.Z"

# 4. Create annotated tag (WITH "v" prefix)
git tag -a vX.Y.Z -m "Release notes..."

# 5. Push everything
git push && git push --tags

# 6. CRITICAL: Verify go.mod has NO replace directive
cat go.mod  # Should NOT contain "replace" line

# 7. Verify install works from GitHub
GOPROXY=direct go install github.com/rosscartlidge/streamv3/cmd/streamv3@vX.Y.Z
streamv3 version  # Should show: streamv3 vX.Y.Z
```

**⚠️ CRITICAL:**
- **version.txt format**: Store WITHOUT "v" prefix (e.g., `1.2.0` not `v1.2.0`)
- **git tag format**: Use WITH "v" prefix (e.g., `v1.2.0`)
- **completionflags adds "v"**: `.Version()` automatically adds "v" prefix to display
- **No replace directive**: `go.mod` must NOT contain `replace` line (breaks `go install`)
- **Annotated tags only**: Use `git tag -a vX.Y.Z -m "..."` not `git tag vX.Y.Z`
- **Test install**: Always verify with `GOPROXY=direct go install` before announcing release

**How It Works:**
- Version stored in `cmd/streamv3/version/version.txt` (plain text, without "v")
- Embedded in binary via `//go:embed version.txt` in `cmd/streamv3/version/version.go`
- completionflags `.Version()` method adds "v" prefix automatically
- `streamv3 version` subcommand shows: "streamv3 vX.Y.Z"
- `streamv3 -help` header shows: "streamv3 vX.Y.Z - Unix-style data processing tools"

**Common Mistakes:**
- ❌ Including "v" in version.txt → Results in "vvX.Y.Z" display
- ❌ Having `replace` directive in go.mod → `go install` fails with error
- ❌ Using lightweight tags → Use annotated tags with `-a` flag
- ❌ Not testing install → Release may be broken for users

**Testing a Release:**
```bash
# After pushing tag, test from a different directory:
cd /tmp
GOPROXY=direct go install github.com/rosscartlidge/streamv3/cmd/streamv3@latest
streamv3 version  # Should show correct version
streamv3 -help    # Should work without errors
```

## Architecture Overview

StreamV3 is a modern Go library built on three core abstractions:

**Core Types:**
- `iter.Seq[T]` and `iter.Seq2[T,error]` - Go 1.23+ iterators (lazy sequences)
- `Record` - Encapsulated struct with private fields map (`struct { fields map[string]any }`)
- `MutableRecord` - Efficient record builder with in-place mutation
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

## Record Design - Encapsulated Struct (v1.0+)

**⚠️ BREAKING CHANGE in v1.0:** Record is now an encapsulated struct, not a bare `map[string]any`.

### Record vs MutableRecord

**Record (Immutable):**
- Struct with private `fields map[string]any`
- Immutable - methods return new copies
- Use for function parameters, return values, pipeline data
- Access via `Get()`, `GetOr()`, `.All()` iterator

**MutableRecord (Mutable Builder):**
- Struct with private `fields map[string]any`
- Mutable - methods modify in-place and return self for chaining
- Use for efficient record construction
- Convert to Record via `.Freeze()` (creates copy)

### Creating Records

```go
// ✅ CORRECT - Use MutableRecord builder
record := streamv3.MakeMutableRecord().
    String("name", "Alice").
    Int("age", int64(30)).
    Float("salary", 95000.50).
    Bool("active", true).
    Freeze()  // Convert to immutable Record

// ✅ CORRECT - From map (for compatibility)
record := streamv3.NewRecord(map[string]any{
    "name": "Alice",
    "age": int64(30),
})

// ❌ WRONG - Can't use struct literal
record := streamv3.Record{"name": "Alice"}  // Won't compile!

// ❌ WRONG - Can't use make()
record := make(streamv3.Record)  // Won't compile!
```

### Accessing Record Fields

**Within streamv3 package:**
```go
// ✅ Can access .fields directly (private field)
for k, v := range record.All() {
    record.fields[k] = v
}

// ✅ Direct field access for internal operations
value := record.fields["name"]
```

**Outside streamv3 package (CLI commands, tests, user code):**
```go
// ✅ CORRECT - Use Get/GetOr
name := streamv3.GetOr(record, "name", "")
age := streamv3.GetOr(record, "age", int64(0))

// ✅ CORRECT - Iterate with .All()
for k, v := range record.All() {
    fmt.Printf("%s: %v\n", k, v)
}

// ✅ CORRECT - Build with MutableRecord
mut := streamv3.MakeMutableRecord()
mut = mut.String("city", "NYC")           // Chainable
mut = mut.SetAny("field", anyValue)       // For unknown types
frozen := mut.Freeze()                    // Convert to Record

// ❌ WRONG - Can't access .fields (private!)
value := record.fields["name"]            // Compile error!

// ❌ WRONG - Can't index directly
name := record["name"]                    // Compile error!

// ❌ WRONG - Can't iterate directly
for k, v := range record {                // Compile error!
    ...
}
```

### Iterating Over Records

```go
// ✅ CORRECT - Use .All() iterator (maps.All pattern)
for k, v := range record.All() {
    fmt.Printf("%s: %v\n", k, v)
}

// ✅ CORRECT - Use .KeysIter() for keys only
for k := range record.KeysIter() {
    fmt.Println(k)
}

// ✅ CORRECT - Use .Values() for values only
for v := range record.Values() {
    fmt.Println(v)
}

// ❌ WRONG - Can't iterate Record directly
for k, v := range record {                // Compile error!
    ...
}
```

### Migration Patterns

**Converting old code to v1.0:**

```go
// OLD (v0.x):
record := make(streamv3.Record)
record["name"] = "Alice"
value := record["age"]
for k, v := range record {
    ...
}

// NEW (v1.0+):
record := streamv3.MakeMutableRecord()
record = record.String("name", "Alice")
value := streamv3.GetOr(record.Freeze(), "age", int64(0))
for k, v := range record.Freeze().All() {
    ...
}
```

**Test code migration:**

```go
// OLD (v0.x):
testData := []streamv3.Record{
    {"name": "Alice", "age": int64(30)},
    {"name": "Bob", "age": int64(25)},
}

// NEW (v1.0+):
r1 := streamv3.MakeMutableRecord()
r1.fields["name"] = "Alice"    // Within streamv3 package
r1.fields["age"] = int64(30)

r2 := streamv3.MakeMutableRecord()
r2.fields["name"] = "Bob"
r2.fields["age"] = int64(25)

testData := []streamv3.Record{r1.Freeze(), r2.Freeze()}
```

## Record Field Access (CRITICAL)

**⚠️ ALWAYS use `Get()` or `GetOr()` methods to read fields from Records. NEVER use direct map access or type assertions.**

**Why:**
- Direct access `r["field"]` requires type assertions: `r["field"].(string)` → **panics if field missing or wrong type**
- Type assertions `r["field"].(string)` are unsafe and fragile
- `Get()` and `GetOr()` handle type conversion, missing fields, and type mismatches gracefully

**Correct Field Access:**
```go
// ✅ CORRECT - Use GetOr with appropriate default
name := streamv3.GetOr(r, "name", "")                    // String field
age := streamv3.GetOr(r, "age", int64(0))                // Numeric field
price := streamv3.GetOr(r, "price", float64(0.0))        // Float field

// ✅ CORRECT - Use in generated code
strings.Contains(streamv3.GetOr(r, "email", ""), "@")
regexp.MustCompile("pattern").MatchString(streamv3.GetOr(r, "name", ""))
streamv3.GetOr(r, "salary", float64(0)) > 50000
```

**Wrong Field Access:**
```go
// ❌ WRONG - Direct map access with type assertion (WILL PANIC!)
name := r["name"].(string)                               // Panic if field missing or wrong type
r["email"].(string)                                      // Panic if field missing
asFloat64(r["price"])                                    // Don't create helper functions - use GetOr!

// ❌ WRONG - Direct map access in comparisons
r["status"] == "active"                                  // May work, but inconsistent
```

**Code Generation Rules:**
- **String operations**: Always use `streamv3.GetOr(r, field, "")` with empty string default
- **Numeric operations**: Always use `streamv3.GetOr(r, field, float64(0))` or `int64(0)` default
- **Never generate**: Type assertions like `r[field].(string)`
- **Never generate**: Custom helper functions like `asFloat64()`

**Examples in Generated Code:**
```go
// String operators (contains, startswith, endswith, regexp)
strings.Contains(streamv3.GetOr(r, "name", ""), "test")
strings.HasPrefix(streamv3.GetOr(r, "email", ""), "admin")
regexp.MustCompile("^[A-Z]").MatchString(streamv3.GetOr(r, "code", ""))

// Numeric operators (eq, ne, gt, ge, lt, le)
streamv3.GetOr(r, "age", float64(0)) > 18
streamv3.GetOr(r, "salary", float64(0)) >= 50000
streamv3.GetOr(r, "count", float64(0)) == 42
```

This approach eliminates runtime panics and makes generated code robust and maintainable.

This library emphasizes functional composition with Go 1.23+ iterators while providing comprehensive data visualization capabilities.

## CLI Tools Architecture (completionflags v0.2.1+)

StreamV3 CLI uses **completionflags v0.2.1+** for native subcommand support with auto-generated help and tab completion. All 14 commands migrated as of v1.2.0.

**Architecture Overview:**
- `cmd/streamv3/main.go` - All subcommands defined using completionflags builder API
- `cmd/streamv3/helpers.go` - Shared utilities (comparison operators, aggregation, extractNumeric, chainRecords)
- `cmd/streamv3/version/version.txt` - Version string (manually maintained)
- All commands use context-based flag access: `ctx.GlobalFlags` and `ctx.Clauses`

**Version Access:**
- `streamv3 version` - Dedicated version subcommand (returns "streamv3 vX.Y.Z")
- `streamv3 -help` - Shows version in header
- ⚠️ No `-version` flag (completionflags doesn't auto-add this)

**CLI Flag Design Principles:**

When designing CLI commands with completionflags, follow these principles:

1. **Prefer Named Flags Over Positional Arguments**
   - ✅ Use: `-file data.csv` or `-input data.csv`
   - ❌ Avoid: `command data.csv` (positional)
   - Named flags are self-documenting and enable better tab completion
   - Positional arguments can consume arguments intended for other flags
   - Exception: Commands with a single, obvious positional argument (e.g., `cd directory`)

2. **Use Multi-Argument Flags Properly**
   - For flags with multiple related arguments, use `.Arg()` fluent API:
   ```go
   Flag("-match").
       Arg("field").Completer(cf.NoCompleter{Hint: "<field-name>"}).Done().
       Arg("operator").Completer(&cf.StaticCompleter{Options: operators}).Done().
       Arg("value").Completer(cf.NoCompleter{Hint: "<value>"}).Done().
   ```
   - This enables proper completion for each argument position
   - Always provide hints via `NoCompleter{Hint: "..."}` when no completion is available
   - Use `StaticCompleter{Options: [...]}` for constrained values
   - ❌ Don't use `.String()` and require quoting: `-match "field op value"`
   - ✅ Use separate arguments: `-match field op value`

3. **Use `.Accumulate()` for Repeated Flags**
   - When a flag can appear multiple times (e.g., `-match age gt 30 -match dept eq Sales`)
   - Enables building complex filters with AND/OR logic
   - The framework provides a slice of all flag occurrences

4. **Provide Completers for Constrained Arguments**
   - Use `StaticCompleter` for known options (operators, commands, etc.)
   - Use `FileCompleter` with patterns for file paths
   - Improves UX with tab completion

**Completionflags Subcommand Pattern:**

All commands follow this pattern in `main.go`:

```go
Subcommand("command-name").
    Description("Brief description").

    Handler(func(ctx *cf.Context) error {
        // 1. Extract flags from ctx.GlobalFlags (for Global flags)
        var myFlag string
        if val, ok := ctx.GlobalFlags["-myflag"]; ok {
            myFlag = val.(string)
        }

        // 2. Extract clause flags (for Local flags with + separators)
        if len(ctx.Clauses) > 0 {
            clause := ctx.Clauses[0]
            if val, ok := clause.Flags["-field"]; ok {
                // Handle accumulated flags: val.([]any)
            }
        }

        // 3. For commands with -- separator (like exec)
        if len(ctx.RemainingArgs) > 0 {
            command := ctx.RemainingArgs[0]
            args := ctx.RemainingArgs[1:]
            // ...
        }

        // 4. Perform command operation
        // 5. Return error or nil
        return nil
    }).

    Flag("-myflag").
        String().
        Global().  // Or Local() for clause-based flags
        Help("Description").
        Done().

    Done().
```

**Key Patterns:**
- **Global flags**: Use `ctx.GlobalFlags["-flagname"]` - applies to entire command
- **Local flags**: Use `ctx.Clauses[i].Flags["-flagname"]` - applies per clause (with `+` separator)
- **Accumulated flags**: Use `.Accumulate()` and access as `[]any` slice
- **-- separator**: Use `ctx.RemainingArgs` for everything after `--` (requires completionflags v0.2.1+)
- **Type assertions**: All flag values are `interface{}`, cast appropriately: `val.(string)`, `val.(int)`, `val.(bool)`

**Important Lessons Learned:**

1. **Release with replace directive fails** - `go install` fails if go.mod has `replace` directive
   - Always remove local `replace` before tagging releases
   - Test with `GOPROXY=direct go install github.com/user/repo/cmd/app@vX.Y.Z`

2. **Version display** - completionflags `.Version()` adds "v" prefix automatically
   - Store version without "v" in version.txt: `1.2.0` not `v1.2.0`
   - Display will show: "streamv3 v1.2.0"

3. **Version subcommand needed** - completionflags doesn't auto-add `-version` flag
   - Must manually add `version` subcommand if users need version access
   - Version also appears in help header automatically

4. **Context-based flag access** - Don't use `.Bind()` for complex commands
   - Use `ctx.GlobalFlags` and `ctx.Clauses` for flexibility
   - Enables dynamic flag handling and accumulation

5. **-- separator support** - Requires completionflags v0.2.1+
   - Use for commands that pass args to other programs (like `exec`)
   - Access via `ctx.RemainingArgs` slice

## Code Generation System (CRITICAL FEATURE)

**⚠️ CRITICAL: This is a core feature that enables 10-100x faster execution by generating standalone Go programs from CLI pipelines.**

### Overview

StreamV3 supports **self-generating pipelines** where commands emit Go code fragments instead of executing. This allows users to:
1. Prototype data processing pipelines using the CLI
2. Generate optimized Go code from the working pipeline
3. Compile and run standalone programs 10-100x faster than CLI execution

### Enabling Code Generation

Two ways to enable generation mode:

```bash
# Method 1: Environment variable (affects entire pipeline)
export STREAMV3_GENERATE_GO=1
streamv3 read-csv data.csv | streamv3 where -match age gt 25 | streamv3 generate-go

# Method 2: -generate flag per command
streamv3 read-csv -generate data.csv | streamv3 where -generate -match age gt 25 | streamv3 generate-go
```

The environment variable approach is preferred for full pipelines.

### Code Fragment System

**Architecture (`cmd/streamv3/lib/codefragment.go`):**
- Commands communicate via JSONL code fragments on stdin/stdout
- Each fragment has: Type, Var (variable name), Input (input var), Code, Imports, Command
- The `generate-go` command assembles all fragments into a complete Go program
- Fragments are passed through the pipeline, with each command adding its own

**Fragment Types:**
- `init` - First command (e.g., read-csv), creates initial variable, no input
- `stmt` - Middle command (e.g., where, group-by), has input and output variable
- `final` - Last command (e.g., write-csv), has input but no output variable

**Helper Functions (in `cmd/streamv3/helpers.go`):**
- `shouldGenerate(flagValue bool)` - Checks flag or STREAMV3_GENERATE_GO env var
- `getCommandString()` - Returns command line that invoked the command (filters out -generate flag)
- `shellQuote(s string)` - Quotes arguments for shell safety

### Generation Support Status (as of v1.2.4)

**✅ Commands with -generate support (8/14):**
1. `read-csv` - Generates init fragment with `streamv3.ReadCSV()`
2. `where` - Generates stmt fragment with filter predicate
3. `write-csv` - Generates final fragment with `streamv3.WriteCSV()`
4. `limit` - Generates stmt fragment with `streamv3.Limit[streamv3.Record](n)`
5. `offset` - Generates stmt fragment with `streamv3.Offset[streamv3.Record](n)`
6. `sort` - Generates stmt fragment with `streamv3.SortBy()`
7. `distinct` - Generates stmt fragment with `streamv3.DistinctBy()`
8. `group-by` - Generates TWO stmt fragments (GroupByFields + Aggregate)

**❌ Commands WITHOUT -generate support yet (6/14):**
- `select`, `join`, `union`, `exec`, `chart`, `generate-go` (doesn't need it)

**⚠️ IMPORTANT:** Commands without generation support will break pipelines in generation mode. Always add generation support when creating new commands.

### Adding Generation Support to Commands

**Step 1: Add generation function to `cmd/streamv3/helpers.go`:**

```go
// generateMyCommandCode generates Go code for the my-command command
func generateMyCommandCode(arg1 string, arg2 int) error {
    // 1. Read all previous code fragments from stdin
    fragments, err := lib.ReadAllCodeFragments()
    if err != nil {
        return fmt.Errorf("reading code fragments: %w", err)
    }

    // 2. Pass through all previous fragments
    for _, frag := range fragments {
        if err := lib.WriteCodeFragment(frag); err != nil {
            return fmt.Errorf("writing previous fragment: %w", err)
        }
    }

    // 3. Get input variable from last fragment (or default to "records")
    var inputVar string
    if len(fragments) > 0 {
        inputVar = fragments[len(fragments)-1].Var
    } else {
        inputVar = "records"
    }

    // 4. Generate your command's Go code
    outputVar := "result"
    code := fmt.Sprintf("%s := streamv3.MyCommand(%q, %d)(%s)",
        outputVar, arg1, arg2, inputVar)

    // 5. Create and write your fragment
    imports := []string{"fmt"}  // Add any needed imports
    frag := lib.NewStmtFragment(outputVar, inputVar, code, imports, getCommandString())
    return lib.WriteCodeFragment(frag)
}
```

**Step 2: Add -generate flag and check to command handler in `cmd/streamv3/main.go`:**

```go
Subcommand("my-command").
    Description("Description of my command").

    Handler(func(ctx *cf.Context) error {
        var arg1 string
        var arg2 int
        var generate bool

        // Extract flags
        if val, ok := ctx.GlobalFlags["-arg1"]; ok {
            arg1 = val.(string)
        }
        if val, ok := ctx.GlobalFlags["-arg2"]; ok {
            arg2 = val.(int)
        }
        if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
            generate = genVal.(bool)
        }

        // Check if generation is enabled (flag or env var)
        if shouldGenerate(generate) {
            return generateMyCommandCode(arg1, arg2)
        }

        // Normal execution follows...
        // ...
    }).

    Flag("-generate", "-g").
        Bool().
        Global().
        Help("Generate Go code instead of executing").
        Done().

    Flag("-arg1").
        String().
        Global().
        Help("First argument").
        Done().

    // ... other flags

    Done().
```

**Step 3: Add tests to `cmd/streamv3/generation_test.go`:**

```go
func TestMyCommandGeneration(t *testing.T) {
    buildCmd := exec.Command("go", "build", "-o", "/tmp/streamv3_test", ".")
    if err := buildCmd.Run(); err != nil {
        t.Fatalf("Failed to build streamv3: %v", err)
    }
    defer os.Remove("/tmp/streamv3_test")

    cmdLine := `echo '{"type":"init","var":"records"}' | STREAMV3_GENERATE_GO=1 /tmp/streamv3_test my-command -arg1 test -arg2 42`
    cmd := exec.Command("bash", "-c", cmdLine)
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Logf("Command output: %s", output)
    }

    outputStr := string(output)
    want := []string{`"type":"stmt"`, `"var":"result"`, `streamv3.MyCommand`}
    for _, expected := range want {
        if !strings.Contains(outputStr, expected) {
            t.Errorf("Expected output to contain %q, got: %s", expected, outputStr)
        }
    }
}
```

### Special Cases

**Commands with multiple fragments (like group-by):**

Some commands generate multiple code fragments. For example, `group-by` generates:
1. `GroupByFields` fragment (with command string)
2. `Aggregate` fragment (empty command string - part of same CLI command)

```go
// Fragment 1: GroupByFields
frag1 := lib.NewStmtFragment("grouped", inputVar, groupCode, nil, getCommandString())
lib.WriteCodeFragment(frag1)

// Fragment 2: Aggregate (note: empty command string)
frag2 := lib.NewStmtFragment("aggregated", "grouped", aggCode, nil, "")
lib.WriteCodeFragment(frag2)
```

### Testing Code Generation

**Manual testing:**
```bash
# Test individual command
export STREAMV3_GENERATE_GO=1
echo '{"type":"init","var":"records"}' | ./streamv3 my-command -arg1 test

# Test full pipeline
export STREAMV3_GENERATE_GO=1
./streamv3 read-csv data.csv | \
  ./streamv3 where -match age gt 25 | \
  ./streamv3 my-command -arg1 test | \
  ./streamv3 generate-go > program.go

# Compile and run generated code
go run program.go
```

**Automated tests:**
- All generation tests are in `cmd/streamv3/generation_test.go`
- Run with: `go test -v ./cmd/streamv3 -run TestGeneration`
- Tests ensure the feature is never lost during refactoring

### Why This Matters

**Code generation is a CRITICAL feature because:**
1. It enables 10-100x performance improvement over CLI execution
2. Generated programs can be deployed without StreamV3 CLI
3. It bridges prototyping (CLI) and production (compiled Go)
4. Breaking it silently breaks user workflows

**Always ensure:**
- New commands include -generate support
- Tests cover generation mode
- Changes to helpers.go don't break fragment system

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
- main_doco_readcsv
- self_inproving
- tags
- doing_cli_port
- cli_completion
- cli 80% SQL and join optimisation plan
- regexp added