# StreamV3 AI Code Generation Test Suite

This directory contains reference implementations and validation tools for testing AI code generation with StreamV3.

## Quick Start

### Run All Tests

```bash
# Run the complete test suite (validates and executes all reference implementations)
./scripts/test-ai-code-generation.sh
```

This will:
- ✅ Validate all 5 reference implementations for correct patterns
- ✅ Execute each one to ensure they run successfully
- ✅ Display a summary with pass/fail counts

### Validate Generated Code

```bash
# Validate any Go file against StreamV3 best practices
./scripts/validate-ai-patterns.sh <file.go>

# Example: Validate a reference implementation
./scripts/validate-ai-patterns.sh test-output/test_case_1_manual.go
```

### Run Reference Implementations

```bash
# Run individual test cases
go run test-output/test_case_1_manual.go
go run test-output/test_case_2_top_n.go
go run test-output/test_case_3_join.go
go run test-output/test_case_4_transform.go
go run test-output/test_case_5_chart.go

# Run all test cases
for file in test-output/test_case_*.go; do
    echo "=== Running $file ==="
    go run "$file"
    echo ""
done
```

## Files

### Reference Implementations

Each test case demonstrates a common StreamV3 pattern:

1. **test_case_1_manual.go** - Basic Filtering and Grouping
   - Filter records with `Where`
   - Group by field with `GroupByFields`
   - Count with `Aggregate`

2. **test_case_2_top_n.go** - Top N with Chain
   - Group and sum revenue
   - Sort descending (negative values)
   - Limit to top 5

3. **test_case_3_join.go** - Join Operation
   - Join two CSV files
   - Group by multiple fields
   - Aggregate totals

4. **test_case_4_transform.go** - Transformation with Select
   - Transform records with `Select`
   - Add computed fields
   - Use `SetImmutable` for record updates

5. **test_case_5_chart.go** - Chart Creation
   - Group and aggregate data
   - Create interactive HTML chart
   - Use `QuickChart` for visualization

### Test Cases (../test-ai-generation-cases.md)

Natural language prompts and expected patterns for each test case.

### Validation Script (../scripts/validate-ai-patterns.sh)

Automated validation checking for:
- ✅ Correct import path
- ✅ No wrong import paths
- ✅ SQL-style API usage (Where not Filter)
- ✅ Error handling
- ✅ Proper GroupByFields syntax
- ✅ Proper Aggregate syntax
- ✅ Proper Count() usage
- ✅ Chain/Pipe composition
- ✅ Compilation

### Validation Report (ai-validation-report.md)

Comprehensive analysis of:
- Test results summary
- AI prompt analysis
- Common error patterns
- Reference implementation patterns
- Recommendations

## Using with AI Code Generation

### Step 1: Copy the AI Prompt

Copy the prompt from `doc/ai-code-generation.md` (lines 8-401) into your LLM:

```bash
# View the AI prompt
sed -n '8,401p' doc/ai-code-generation.md
```

### Step 2: Test with Natural Language

Use prompts from `../test-ai-generation-cases.md`:

**Example Prompt:**
> "Read employee data from employees.csv, filter for employees with salary over 80000, group by department, and count how many employees are in each department"

### Step 3: Validate Generated Code

```bash
# Save the generated code to a file
vim generated_code.go

# Validate it
./scripts/validate-ai-patterns.sh generated_code.go

# If it passes validation, run it
go run generated_code.go
```

### Step 4: Compare with Reference

```bash
# Compare your generated code with the reference implementation
diff -u generated_code.go test-output/test_case_1_manual.go
```

## Expected Patterns

### All Code Should Have:

- ✅ `package main` and `func main()`
- ✅ Import: `"github.com/rosscartlidge/streamv3"`
- ✅ Error handling: `if err != nil { log.Fatalf(...) }`
- ✅ SQL-style names: `Select`, `Where`, `Limit` (not Map, Filter, Take)
- ✅ Chain composition for multi-step pipelines
- ✅ Clear variable names (`highSalaryEmployees`, not `hse`)

### Common Errors to Avoid:

- ❌ Wrong import: `github.com/rocketlaunchr/streamv3`
- ❌ Wrong API: `GroupByFields([]string{...}, []Aggregation{...})`
- ❌ Wrong Count: `Count("field")` instead of `"field": Count()`
- ❌ No error handling: `data, _ := ReadCSV(...)`
- ❌ Functional names: `Map`, `Filter`, `Take` instead of `Select`, `Where`, `Limit`

## Test Data

All reference implementations create their own sample CSV data in `/tmp`:

- `/tmp/employees.csv` - Employee salary data
- `/tmp/sales.csv` - Product sales data
- `/tmp/customers.csv` + `/tmp/orders.csv` - Customer and order data
- `/tmp/products.csv` - Product pricing data
- `/tmp/sales_monthly.csv` - Monthly sales data

Charts are generated in `/tmp`:

- `/tmp/monthly_sales.html` - Interactive Chart.js visualization

## Validation Results

All 5 reference implementations pass validation:

```
test_case_1_manual.go    ✓ 0 errors, 0 warnings
test_case_2_top_n.go     ✓ 0 errors, 0 warnings
test_case_3_join.go      ✓ 0 errors, 0 warnings
test_case_4_transform.go ✓ 0 errors, 1 warning (style preference)
test_case_5_chart.go     ✓ 0 errors, 0 warnings
```

## Next Steps

1. Copy the AI prompt from `doc/ai-code-generation.md`
2. Test with natural language queries
3. Validate generated code with `./scripts/validate-ai-patterns.sh`
4. Compare with reference implementations
5. Iterate and improve prompts based on results

## See Also

- `../doc/ai-code-generation.md` - Complete AI prompt
- `../doc/ai-code-generation-detailed.md` - Detailed examples with explanations
- `../README.md` - Main StreamV3 documentation
- `ai-validation-report.md` - Comprehensive validation analysis
