# AI Code Generation Test Cases

These test cases validate that the AI prompt generates correct ssql code.

## Test Case 1: Basic Filtering and Grouping
**Natural Language Prompt:**
"Read employee data from employees.csv, filter for employees with salary over 80000, group by department, and count how many employees are in each department"

**Expected Code Patterns:**
- `ReadCSV("employees.csv")`
- `if err != nil`
- `Where(` or `ssql.Where(`
- `salary > 80000` or `> 80000`
- `GroupByFields(` and `"department"`
- `Count()`

**Should Use:**
- Error handling after ReadCSV
- SQL-style naming (Where, not Filter)
- Clear variable names

## Test Case 2: Top N with Chain
**Natural Language Prompt:**
"Find the top 5 products by revenue from sales data. Group by product name and show the total revenue for each"

**Expected Code Patterns:**
- `ReadCSV("sales` or `ReadCSV("`
- `GroupByFields(` and `"product`
- `Sum("revenue")`
- `SortBy(` and descending (negative value)
- `Limit[` and `](5)`
- `Chain(` (preferred) or sequential steps

**Should Use:**
- Chain() for pipeline composition (or clear sequential steps)
- Descending sort (negative value)
- Error handling

## Test Case 3: Join Operation
**Natural Language Prompt:**
"Join customer data with order data on customer_id, then calculate total spending per customer"

**Expected Code Patterns:**
- `ReadCSV` (two calls)
- `InnerJoin(`
- `OnFields("customer_id")`
- `GroupByFields(` and `"customer_id"`
- `Sum(`

**Should Use:**
- Proper join predicate
- Error handling for both CSV reads
- Clear variable names (not just `left`, `right`)

## Test Case 4: Transformation with Select
**Natural Language Prompt:**
"Read product data, add a 'price_tier' field that is 'Budget' if price < 100, 'Mid' if 100-500, 'Premium' if > 500"

**Expected Code Patterns:**
- `ReadCSV("product`
- `Select(` or `ssql.Select(`
- `switch` or multiple `if` statements
- `SetImmutable(` or `SetField(`
- `"price_tier"`

**Should Use:**
- Select for transformation
- Clear tier logic
- Immutable record updates

## Test Case 5: Chart Creation
**Natural Language Prompt:**
"Read monthly sales from sales.csv, group by month, sum the revenue, and create a bar chart"

**Expected Code Patterns:**
- `ReadCSV("sales.csv")`
- `GroupByFields(` and `"month"`
- `Sum("revenue")`
- `QuickChart(` or `InteractiveChart(`
- `.html"`

**Should Use:**
- Chart generation at the end
- Proper aggregation before charting
