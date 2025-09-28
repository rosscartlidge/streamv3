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
- `Stream[T]` - Lazy sequences using Go 1.23+ `iter.Seq[T]` and `iter.Seq2[T,error]`
- `Record` - Map-based flexible data structure (`map[string]any`)
- `Filter[T,U]` - Composable transformations (`func(iter.Seq[T]) iter.Seq[U]`)

**Key Architecture Files:**
- `core.go` - Core types, Filter functions, Record system, composition functions
- `operations.go` - Stream operations (Map, Where, Reduce, etc.)
- `fluent.go` - StreamBuilder fluent API for ergonomic method chaining
- `chart.go` - Interactive Chart.js visualization with Bootstrap 5 UI
- `io.go` - CSV/JSON I/O, command parsing, file operations
- `sql.go` - GROUP BY aggregations and SQL-style operations

**Dual API Design:**
1. **Functional API** - Explicit Filter composition: `Pipe(Map(...), Where(...))`
2. **Fluent API** - Method chaining: `From(data).Map(...).Where(...).Collect()`

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
- `From(slice)` - Create stream from slice
- `ReadCSV(filename)` - Parse CSV files
- `ExecCommand(cmd, args...)` - Parse command output
- `QuickChart(data, x, y, filename)` - Generate interactive charts

This library emphasizes functional composition with Go 1.23+ iterators while providing comprehensive data visualization capabilities.