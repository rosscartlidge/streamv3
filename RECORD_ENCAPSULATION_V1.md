# Record Encapsulation Migration to v1.0.0

## Problem Statement

The current `type Record map[string]any` design breaks type safety:
- Users can bypass Get/GetOr methods with `record["field"]`
- Non-canonical types can be inserted (e.g., `int` instead of `int64`)
- Builder pattern enforcement is undermined
- Generated code can't guarantee canonical type conventions

## Solution

Encapsulate the map in a struct with private fields:
```go
type Record struct {
    fields map[string]any  // private!
}
```

## Design Decisions

### 1. Clean Break - No Backwards Compatibility
- This is v1.0.0 - worth breaking for better design
- Package is still young enough to make this change
- Benefits outweigh migration cost

### 2. Freeze() Semantics
- **Copy on freeze** - safe, matches current behavior
- Prevents mutations from leaking between mutable and immutable records

### 3. maps-style API
Model Record API on Go 1.23+ `maps` package:
- `All() iter.Seq2[string, any]` - iterate key/value pairs
- `KeysIter() iter.Seq[string]` - iterate keys
- `Values() iter.Seq[any]` - iterate values
- `Clone() Record` - shallow copy
- `Equal(Record) bool` - equality check
- `Len() int` - field count
- `Delete(string)` - remove field (mutable only)

### 4. Constructors
- `MakeMutableRecord()` - primary builder
- `MakeMutableRecordWithCapacity(int)` - preallocated capacity
- `NewRecord(map[string]any)` - compatibility constructor (copies map)

### 5. Migration Pattern
```go
// Before: for k, v := range record
// After:  for k, v := range record.All()
```

Just add `.All()` - nearly drop-in!

## Files to Update

### Core Implementation
1. ✅ `core.go` - Define new struct types and methods
2. [ ] Add JSON marshaling (MarshalJSON/UnmarshalJSON)
3. [ ] Fix range loops in core.go to use .All()

### File-by-File Migration
4. [ ] `io.go` - 26 direct assignments, CSV/JSON parsing
5. [ ] `sql.go` - Range loops, len() calls
6. [ ] `chart.go` - Range loops, len() calls
7. [ ] `operations.go` - If needed
8. [ ] CLI commands in `cmd/streamv3/` - Check for issues

### Testing & Documentation
9. [ ] Run tests, fix breakage
10. [ ] Update CLAUDE.md with new Record design
11. [ ] Update version to 1.0.0

## API Changes

### What Breaks
```go
// ❌ BREAKS: Direct map access
record["field"] = value
value := record["field"]
for k, v := range record { }
len(record)

// ✅ WORKS: Encapsulated API
Set(record, "field", value)  // or record.String("field", "value")
value, _ := Get[string](record, "field")
for k, v := range record.All() { }
record.Len()
```

### What's Added
```go
// New methods
record.All()        // iter.Seq2[string, any]
record.KeysIter()   // iter.Seq[string]
record.Values()     // iter.Seq[any]
record.Clone()      // Record
record.Equal(other) // bool
record.Len()        // int

// Mutable-specific
mutable.Delete(field)
```

## Implementation Status

- [x] Record/MutableRecord struct definitions
- [x] Constructors and conversion methods
- [x] Get/GetOr/Has/Keys updated
- [x] maps-style iterator API (All, KeysIter, Values, Clone, Equal)
- [x] Set/SetImmutable updated
- [x] Delete method for MutableRecord
- [ ] JSON marshaling
- [ ] Fix range loops in core.go
- [ ] Fix dotFlattenRecord and other core.go internals
- [ ] Update io.go (CSV/JSON parsing)
- [ ] Update sql.go
- [ ] Update chart.go
- [ ] Update CLI commands
- [ ] Test suite
- [ ] Documentation

## Testing Strategy

1. Run `go test ./...` after each major file update
2. Focus on:
   - CSV/JSON round-tripping
   - Record equality and cloning
   - Builder pattern still works
   - Generated CLI code still works
3. Check CLI with actual data pipelines

## Version Strategy

- **v1.0.0** - Breaking change, but worth it
- Update version in `cmd/streamv3/main.go` and `cmd/streamv3/lib/codefragment.go`
- Create annotated git tag
- Update CLAUDE.md to document the change
