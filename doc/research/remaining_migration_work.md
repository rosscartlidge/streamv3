# Remaining Migration Work

## Status: POC Complete ✅

Successfully migrated 3 commands (limit, offset, distinct) to native subcommand support.
- Auto-generated help works perfectly
- Completion works for subcommands and flags
- Commands execute correctly
- Code is much cleaner

## Remaining Commands (11)

### Simple Commands (2)
These have straightforward flags, no clause support:

1. **sort** - `-field`, `-desc` flags
   - Extract field and desc from GlobalFlags
   - Use extractNumeric() helper (already created in helpers.go)
   - Business logic unchanged

### I/O Commands (2)
File reading/writing commands:

2. **read-csv** - `-file`, `-delimiter`, `-header` flags
   - Read CSV file or stdin
   - Output JSONL

3. **write-csv** - `-file`, `-delimiter`, `-header` flags
   - Read JSONL from stdin
   - Write CSV file

### Clause-Based Commands (3)
Use Local() flags with clause support:

4. **select** - `-field`, `-as` flags (Local)
   - Iterate ctx.Clauses
   - Build field mapping
   - Apply selection

5. **where** - `-match` flag with 3 args (field, operator, value) (Local)
   - Iterate ctx.Clauses
   - Build AND conditions within clause, OR between clauses
   - Apply filtering

6. **group-by** - `-field`, `-aggregate` flags (Local)
   - Extract grouping fields and aggregations
   - Apply GROUP BY logic

### Complex Commands (4)

7. **join** - `-type`, `-right`, `-on`, `-left-field`, `-right-field`
   - Already uses OnFields() and OnCondition()
   - Should work with minimal changes
   - Most complex command

8. **chart** - Many flags for chart configuration
   - Extract all chart config flags
   - Apply charting logic

9. **exec** - `-cmd` flag, command execution
   - Execute external command
   - Parse output as records

10. **union** - `-file` flags (Accumulate)
    - Combine multiple data sources
    - Simple concatenation

11. **generate-go** - Special command for code generation
    - Reads code fragments
    - Generates complete Go program

## Migration Pattern

For each command, follow this pattern:

```go
Subcommand("command-name").
    Description("...").

    Handler(func(ctx *cf.Context) error {
        // 1. Extract flags from ctx.GlobalFlags or ctx.Clauses
        // 2. Validate inputs
        // 3. Read input data
        // 4. Apply operation
        // 5. Write output
        return nil
    }).

    // 2. Define flags
    Flag("-flag").
        Type().
        Bind() or use ctx.GlobalFlags.
        Global() or Local().
        Help("...").
        Done().

    Done().
```

## Key Decisions Made

1. **Skip -generate flag in POC** - Focus on execution first, add code generation later
2. **Use helpers.go** - Shared utility functions like extractNumeric()
3. **Disable old main** - Using build tag `//go:build old`
4. **Keep command files** - Can reference for business logic

## Testing Strategy

For each migrated command:
1. Build: `go build ./cmd/streamv3`
2. Test help: `./streamv3 command-name -help`
3. Test execution: `echo '{}' | ./streamv3 command-name [flags]`
4. Test completion: `./streamv3 -complete 1 comm`

## Time Estimates

- Simple commands (sort): 15 min
- I/O commands (read-csv, write-csv): 30 min each
- Clause commands (select, where, group-by): 30 min each
- Complex commands: 20-30 min each
- Testing: 30 min

**Total remaining: 4-5 hours**

## Next Session Plan

1. Start with **sort** (simplest)
2. Then **read-csv** and **write-csv** (I/O pattern)
3. Then **select** (clause pattern)
4. Then **where** (most important clause command)
5. Then **join** (most complex, but mostly done)
6. Then remaining commands
7. Full testing
8. Cleanup and release

## Current State

**Files:**
- `cmd/streamv3/main.go` - Has 3 commands (limit, offset, distinct)
- `cmd/streamv3/main_old.go` - Disabled with build tag
- `cmd/streamv3/helpers.go` - Has extractNumeric()
- `cmd/streamv3/commands/*.go` - Original implementations (reference)

**Branch:** `feature/native-subcommands`
**Last Commit:** POC with 3 commands working

**Ready to continue!**
