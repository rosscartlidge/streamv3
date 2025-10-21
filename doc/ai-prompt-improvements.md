# AI Prompt Self-Improvement Analysis

Based on the automated validation suite, here are insights and improvements we can make to the AI prompt.

## What the Validation Suite Revealed

### ✅ Strong Areas (Already Well-Covered)

1. **SQL-Style Naming** - Multiple warnings, good examples ✅
2. **Error Handling** - Emphasized repeatedly ✅
3. **Chain() Usage** - Well documented with examples ✅
4. **Import Guidelines** - Clear "ONLY import used packages" ✅

### ⚠️ Areas That Could Be Stronger

Based on validation checks that detect errors, here's what LLMs might get wrong:

## 1. GroupByFields + Aggregate Pattern (CRITICAL)

**What validation detects:**
```go
❌ GroupByFields([]string{"field"}, []Aggregation{Count("count")})
✅ GroupByFields("namespace", "field")
   Aggregate("namespace", map[string]AggregateFunc{...})
```

**Current prompt coverage:** Lines 83-84, 261
**Issue:** Not prominent enough! This is the #1 API confusion pattern.

**Improvement:** Add a dedicated "ANTI-PATTERNS" section:

```markdown
## ⛔ ANTI-PATTERNS - DO NOT USE THESE

### Wrong GroupByFields + Aggregate (Common LLM Mistake)

❌ **WRONG - This API doesn't exist:**
```go
// LLMs often hallucinate this combined API
result := streamv3.GroupByFields(
    []string{"department"},
    []streamv3.Aggregation{
        streamv3.Count("employee_count"),
        streamv3.Sum("salary", "total_salary"),
    },
)
```

✅ **CORRECT - Two separate steps:**
```go
// Step 1: Group by fields
grouped := streamv3.GroupByFields("analysis", "department")(data)

// Step 2: Aggregate with map
results := streamv3.Aggregate("analysis", map[string]streamv3.AggregateFunc{
    "employee_count": streamv3.Count(),
    "total_salary":   streamv3.Sum("salary"),
})(grouped)
```

**Why this matters:** The namespace ("analysis") must match between GroupByFields and Aggregate!
```

## 2. Count() Takes No Parameters

**What validation detects:**
```go
❌ Count("field_name")
✅ "field_name": Count()  // Field name is the map key!
```

**Current prompt coverage:** Line 105 shows `Count()` but doesn't emphasize NO parameters
**Issue:** Not explicit enough about this being parameterless

**Improvement:** Add to Aggregation Functions section:

```markdown
### Aggregation Functions

```go
streamv3.Count()                    // ⚠️ NO PARAMETERS! Field name is map key
streamv3.Sum("field")               // DOES take field parameter
streamv3.Avg("field")
streamv3.Min[T]("field")
streamv3.Max[T]("field")
```

**Common mistake:**
❌ `Count("employee_count")` - Won't compile! Count takes no parameters
✅ `"employee_count": Count()` - Field name is the map key, not a parameter
```

## 3. Import Path Disambiguation

**What validation detects:**
```go
❌ "github.com/rocketlaunchr/streamv3"
✅ "github.com/rosscartlidge/streamv3"
```

**Current prompt coverage:** Line 38 shows correct import
**Issue:** Doesn't explain WHY this specific import path

**Improvement:** Add context to import section:

```markdown
### Imports - CRITICAL RULE

**⚠️ IMPORTANT: Use the correct import path!**

```go
import "github.com/rosscartlidge/streamv3"  // ✅ CORRECT
```

**Common mistakes:**
- ❌ `github.com/rocketlaunchr/streamv3` - Old/different project
- ❌ `github.com/streamv3/v3` - Doesn't exist
- ❌ `streamv3` - Not in stdlib

**Why this matters:** LLMs trained on public code might confuse this with other stream libraries. Always use `github.com/rosscartlidge/streamv3`.
```

## 4. Descending Sort Pattern

**What our reference implementations show:**
```go
// All 3 top-N examples use negative values for descending sort
streamv3.SortBy(func(r streamv3.Record) float64 {
    return -streamv3.GetOr(r, "total", 0.0)  // Negative = descending
})
```

**Current prompt coverage:** Line 174 shows this in an example
**Issue:** Not explained WHY negative values = descending

**Improvement:** Add to Sort section:

```markdown
### Sorting

- `Sort()` - Sort with natural ordering
- `SortBy(func(T) K)` - Sort by key function
  - **Descending order:** Use negative values: `return -value`
  - **Ascending order:** Use positive values: `return value`
- `SortDesc()` - Reverse sort order
- `Reverse()` - Reverse sequence

**Top N pattern (very common):**
```go
// Get top 10 by revenue (descending)
topItems := streamv3.Chain(
    streamv3.SortBy(func(r streamv3.Record) float64 {
        return -streamv3.GetOr(r, "revenue", 0.0)  // Negative for DESC
    }),
    streamv3.Limit[streamv3.Record](10),
)(data)
```
```

## 5. Type Parameter Guidance

**What validation shows:** `Limit[streamv3.Record](10)` appears in all examples

**Current prompt coverage:** Line 209 mentions type parameters briefly
**Issue:** Doesn't explain WHEN you need `[T]` vs when you don't

**Improvement:** Add dedicated section:

```markdown
### Type Parameters - When to Add [T]

**Rule of thumb:** Add `[T]` when the compiler can't infer the type.

**Common cases needing type parameters:**
```go
streamv3.Limit[streamv3.Record](10)           // ✅ Usually needed
streamv3.Offset[streamv3.Record](5)           // ✅ Usually needed
streamv3.CountWindow[streamv3.Record](100)    // ✅ Usually needed
streamv3.TimeWindow[streamv3.Record](duration, field)  // ✅ Usually needed
```

**Cases that DON'T need type parameters:**
```go
streamv3.Where(predicate)       // ✅ Type inferred from predicate
streamv3.Select(transform)      // ✅ Type inferred from transform
streamv3.SortBy(keyFunc)        // ✅ Type inferred from keyFunc
```

**When in doubt:** Try without `[T]` first. If compiler complains, add it.
```

## 6. Namespace Matching (GroupBy + Aggregate)

**What reference implementations show:** Namespace must match!

```go
grouped := streamv3.GroupByFields("sales", "region")(data)
totals := streamv3.Aggregate("sales", aggregations)(grouped)
                                 ↑↑↑↑↑ Must match!
```

**Current prompt coverage:** Not explicitly stated
**Issue:** LLMs might use different namespaces

**Improvement:** Add to GroupByFields section:

```markdown
### GroupBy + Aggregate Pattern

**CRITICAL: Namespace must match between GroupByFields and Aggregate!**

```go
// Step 1: Group by fields with namespace "analysis"
grouped := streamv3.GroupByFields("analysis", "department")(data)

// Step 2: Aggregate using SAME namespace "analysis"
results := streamv3.Aggregate("analysis", map[string]streamv3.AggregateFunc{
    "count": streamv3.Count(),
})(grouped)
```

❌ **Wrong - Different namespaces:**
```go
grouped := streamv3.GroupByFields("sales", "region")(data)
results := streamv3.Aggregate("analysis", aggs)(grouped)  // Won't work!
                                  ↑↑↑↑↑↑↑↑↑
                                  Different namespace!
```

**Tip:** Use descriptive namespaces that explain what you're analyzing: `"sales_by_region"`, `"customer_spending"`, etc.
```

## 7. Complete Examples Pattern

**What ALL reference implementations do:**
- Generate sample data in /tmp
- Include full package main
- Create runnable, self-contained code

**Current prompt coverage:** Rule #10 says "Complete examples"
**Issue:** Doesn't show the sample data generation pattern

**Improvement:** Add to Code Generation Rules:

```markdown
10. **Complete, runnable examples**:
    - Include `package main` and `func main()`
    - Generate sample data in /tmp when helpful
    - Make examples copy-paste runnable
    - Example pattern:
    ```go
    func main() {
        // Create sample data
        csvData := `name,age
    Alice,30
    Bob,25`
        os.WriteFile("/tmp/data.csv", []byte(csvData), 0644)

        // Process the data
        data, err := streamv3.ReadCSV("/tmp/data.csv")
        // ... rest of code
    }
    ```
```

## 8. Chain vs Sequential Steps

**What reference implementations show:** Mix of both styles

**Current prompt coverage:** Lines 110-146 show both options
**Issue:** Doesn't give clear guidance on when to use which

**Improvement:** Add decision guide:

```markdown
### Chain() vs Sequential Steps - Decision Guide

**Use Chain() when:**
- ✅ Pipeline has 3+ operations
- ✅ All operations are same type (Record → Record)
- ✅ Reading top-to-bottom is natural
- ✅ You want concise code

```go
result := streamv3.Chain(
    streamv3.Where(pred1),
    streamv3.Where(pred2),
    streamv3.Select(transform),
    streamv3.Limit[streamv3.Record](10),
)(data)
```

**Use sequential steps when:**
- ✅ You need to inspect intermediate results
- ✅ Each step has complex logic
- ✅ Operations change types
- ✅ Debugging/learning

```go
filtered := streamv3.Where(complexPredicate)(data)
// Can print/inspect filtered here
grouped := streamv3.GroupByFields("analysis", "field")(filtered)
// Can print/inspect grouped here
final := streamv3.Aggregate("analysis", aggs)(grouped)
```
```

## Summary of Improvements

### High Priority (Prevent Common Errors):
1. ⚠️ **Add ANTI-PATTERNS section** - Show wrong GroupByFields API explicitly
2. ⚠️ **Emphasize Count() has NO parameters** - This is very confusing
3. ⚠️ **Explain namespace matching** - GroupBy and Aggregate must match
4. ⚠️ **Disambiguate import path** - Explain why this specific path

### Medium Priority (Improve Code Quality):
5. 📊 **Explain negative values for descending sort** - Common pattern
6. 📊 **Type parameter decision guide** - When to add [T]
7. 📊 **Chain vs Sequential decision guide** - When to use which

### Low Priority (Nice to Have):
8. ✨ **Sample data generation pattern** - Make examples runnable
9. ✨ **Common pipeline patterns** - Real-world templates

## Recommended Prompt Structure Update

```markdown
1. Quick Reference (existing)
2. **ANTI-PATTERNS** ← NEW! Show what NOT to do
3. Core Operations (existing)
4. **Type Parameters Guide** ← NEW! When to use [T]
5. **Pattern Decision Guides** ← NEW! Chain vs Sequential, etc.
6. Code Examples (existing)
7. Common Patterns (existing)
```

## Implementation Plan

Should we:
1. Create an improved prompt in `doc/ai-code-generation-v2.md`?
2. A/B test with LLMs to see if fewer errors occur?
3. Update validation report with "before/after" comparisons?
