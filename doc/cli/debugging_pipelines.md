# Debugging StreamV3 Pipelines

This guide shows how to debug and troubleshoot StreamV3 CLI pipelines using built-in tools and standard Unix utilities.

## Table of Contents
- [Quick Debugging Tips](#quick-debugging-tips)
- [Using jq for Pipeline Inspection](#using-jq-for-pipeline-inspection)
- [Common Debugging Patterns](#common-debugging-patterns)
- [Performance Debugging](#performance-debugging)
- [Troubleshooting Common Issues](#troubleshooting-common-issues)

---

## Quick Debugging Tips

### 1. Inspect Data at Any Stage

StreamV3 uses JSONL (JSON Lines) for inter-process communication, making it easy to inspect data flowing through your pipeline:

```bash
# See first 5 records after reading CSV
streamv3 read-csv data.csv | head -5

# Pretty-print with jq
streamv3 read-csv data.csv | jq '.' | head -5

# Check what's passing through a filter
streamv3 read-csv data.csv | streamv3 where -match age gt 30 | jq '.'
```

### 2. Count Records at Each Stage

```bash
# Count input records
streamv3 read-csv data.csv | wc -l

# Count after filtering
streamv3 read-csv data.csv | streamv3 where -match status eq active | wc -l

# Use jq for exact count (handles empty lines)
streamv3 read-csv data.csv | jq -s 'length'
```

### 3. Sample Data During Development

```bash
# Work with small dataset while developing
streamv3 read-csv large_file.csv | streamv3 limit -n 100 | \
  streamv3 where -match ... | \
  jq '.'

# Test with just first 10 records
streamv3 read-csv data.csv | head -10 | streamv3 where ...
```

---

## Using jq for Pipeline Inspection

[jq](https://jqlang.github.io/jq/) is a powerful JSON processor that pairs perfectly with StreamV3's JSONL format.

### Installation

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Other platforms: https://jqlang.github.io/jq/download/
```

### Basic jq Patterns

#### View Records (Pretty-Printed)

```bash
# Pretty-print all records
streamv3 read-csv data.csv | jq '.'

# With color, paginated
streamv3 read-csv data.csv | jq -C '.' | less -R

# Compact format (one line per record)
streamv3 read-csv data.csv | jq -c '.'
```

#### Extract Specific Fields

```bash
# Single field
streamv3 read-csv data.csv | jq '.name'

# Multiple fields
streamv3 read-csv data.csv | jq '{name, age, department}'

# Field with computed value
streamv3 read-csv data.csv | jq '{name, annual_salary: (.salary * 12)}'
```

#### Filter Records

```bash
# Filter in jq (alternative to StreamV3 where command)
streamv3 read-csv data.csv | jq 'select(.age > 30)'

# Combine with StreamV3 filters
streamv3 read-csv data.csv | \
  streamv3 where -match department eq Engineering | \
  jq 'select(.salary > 80000)'
```

#### Inspect Data Types

```bash
# See type of entire record
streamv3 read-csv data.csv | head -1 | jq 'type'

# See types of all fields
streamv3 read-csv data.csv | head -1 | jq 'to_entries | map({key, type: .value | type})'

# Check specific field type
streamv3 read-csv data.csv | jq '.age | type' | head -5
```

#### Analyze Arrays and Nested Data

```bash
# Inspect GROUP BY results
streamv3 read-csv data.csv | \
  streamv3 group -by department -function count -result total | \
  jq '.'

# Extract nested fields
streamv3 read-csv data.csv | jq '.metadata.created'

# Flatten nested structure
streamv3 read-csv data.csv | jq 'flatten'
```

#### Statistical Analysis

```bash
# Count records
streamv3 read-csv data.csv | jq -s 'length'

# Sum a field
streamv3 read-csv data.csv | jq -s 'map(.salary) | add'

# Average
streamv3 read-csv data.csv | jq -s 'map(.salary) | add / length'

# Min/Max
streamv3 read-csv data.csv | jq -s 'map(.age) | min'
streamv3 read-csv data.csv | jq -s 'map(.age) | max'

# Unique values
streamv3 read-csv data.csv | jq -r '.department' | sort -u
```

### Advanced jq Debugging

#### Compare Before/After Filter

```bash
# Count before filter
echo "Before: $(streamv3 read-csv data.csv | wc -l)"

# Count after filter
echo "After: $(streamv3 read-csv data.csv | streamv3 where -match age gt 30 | wc -l)"

# See what was filtered out
streamv3 read-csv data.csv | jq 'select(.age <= 30)'
```

#### Validate Field Presence

```bash
# Find records missing required fields
streamv3 read-csv data.csv | jq 'select(.email == null or .email == "")'

# Check for unexpected nulls
streamv3 read-csv data.csv | jq 'select(has("required_field") | not)'
```

#### Debug Type Mismatches

```bash
# Find non-numeric values in numeric field
streamv3 read-csv data.csv | jq 'select(.age | type != "number")'

# Show records where field has unexpected type
streamv3 read-csv data.csv | jq 'select(.salary | type == "string")'
```

---

## Common Debugging Patterns

### Pattern 1: "Why is my filter not matching?"

```bash
# Step 1: See what data looks like
streamv3 read-csv data.csv | jq '.' | head -3

# Step 2: Check the specific field
streamv3 read-csv data.csv | jq '.age' | head -10

# Step 3: Check field type
streamv3 read-csv data.csv | jq '.age | type' | head -5

# Step 4: Test filter manually
streamv3 read-csv data.csv | jq 'select(.age > 30)' | head -5

# Step 5: Compare with StreamV3 filter
streamv3 read-csv data.csv | streamv3 where -match age gt 30 | jq '.' | head -5
```

### Pattern 2: "My GROUP BY results look wrong"

```bash
# Step 1: Verify grouping field values
streamv3 read-csv data.csv | jq -r '.department' | sort | uniq -c

# Step 2: See raw GROUP BY output
streamv3 read-csv data.csv | \
  streamv3 group -by department -function count -result total | \
  jq '.'

# Step 3: Check specific group
streamv3 read-csv data.csv | \
  streamv3 group -by department -function count -result total | \
  jq 'select(.department == "Engineering")'

# Step 4: Verify counts manually
streamv3 read-csv data.csv | jq 'select(.department == "Engineering")' | wc -l
```

### Pattern 3: "Field values are wrong type"

```bash
# Diagnose: Check CSV parsing
streamv3 read-csv data.csv | head -1 | jq '.'

# Common issue: Numbers parsed as strings
# Solution: CSV auto-parses to int64/float64, but manual data might not

# Fix in jq if needed:
streamv3 read-csv data.csv | jq '.age = (.age | tonumber)' | \
  streamv3 where -match age gt 30
```

### Pattern 4: "Pipeline is too slow"

```bash
# Test with small sample first
time (streamv3 read-csv large.csv | head -1000 | streamv3 where ...)

# Profile each stage
time streamv3 read-csv large.csv > /dev/null
time (streamv3 read-csv large.csv | streamv3 where -match age gt 30 > /dev/null)
time (streamv3 read-csv large.csv | streamv3 where ... | streamv3 group ... > /dev/null)

# Use limit to stop early
streamv3 read-csv large.csv | streamv3 where ... | streamv3 limit -n 100
```

---

## Performance Debugging

### Measure Pipeline Stages

```bash
# Time entire pipeline
time (streamv3 read-csv data.csv | \
  streamv3 where -match age gt 30 | \
  streamv3 group -by department -function count -result total | \
  streamv3 write-csv output.csv)

# Time individual commands
time streamv3 read-csv data.csv > /tmp/stage1.jsonl
time (cat /tmp/stage1.jsonl | streamv3 where -match age gt 30 > /tmp/stage2.jsonl)
time (cat /tmp/stage2.jsonl | streamv3 group ... > /tmp/stage3.jsonl)
```

### Check Record Counts

```bash
# Ensure no data loss between stages
echo "Input: $(streamv3 read-csv data.csv | wc -l)"
echo "After filter: $(streamv3 read-csv data.csv | streamv3 where -match age gt 30 | wc -l)"
echo "After group: $(streamv3 read-csv data.csv | streamv3 where -match age gt 30 | streamv3 group -by dept -function count -result n | wc -l)"
```

### Sample Large Datasets

```bash
# Random sample (requires gnu coreutils shuf)
streamv3 read-csv huge.csv | shuf | head -1000 | streamv3 where ...

# Every Nth record
streamv3 read-csv huge.csv | awk 'NR % 100 == 0' | streamv3 where ...

# First N records
streamv3 read-csv huge.csv | streamv3 limit -n 1000 | streamv3 where ...
```

---

## Troubleshooting Common Issues

### Issue: "No output from pipeline"

**Diagnosis:**
```bash
# Check if input file exists
ls -lh data.csv

# Check if input is readable
streamv3 read-csv data.csv | head -1

# Check each stage
streamv3 read-csv data.csv | tee /tmp/stage1.jsonl | \
  streamv3 where -match age gt 30 | tee /tmp/stage2.jsonl | \
  streamv3 write-csv output.csv
```

**Common causes:**
- Empty input file
- Filter matches zero records
- Typo in field name
- Wrong operator (e.g., using `eq` instead of `gt`)

### Issue: "Wrong number of records"

**Diagnosis:**
```bash
# Count at each stage
streamv3 read-csv data.csv | tee >(wc -l >&2) | \
  streamv3 where -match age gt 30 | tee >(wc -l >&2) | \
  streamv3 write-csv output.csv

# Or step by step
echo "Input: $(streamv3 read-csv data.csv | wc -l)"
echo "Filtered: $(streamv3 read-csv data.csv | streamv3 where -match age gt 30 | wc -l)"
```

### Issue: "Field not found errors"

**Diagnosis:**
```bash
# List all field names
streamv3 read-csv data.csv | head -1 | jq 'keys'

# Check for whitespace in field names
streamv3 read-csv data.csv | head -1 | jq 'keys | map(length)'

# Look for the field you think exists
streamv3 read-csv data.csv | head -1 | jq 'keys | map(select(contains("age")))'
```

**Common causes:**
- Case sensitivity: `Age` vs `age`
- Extra spaces: `"name "` vs `"name"`
- Different separator in CSV (e.g., semicolon instead of comma)

### Issue: "Type comparison errors"

**Diagnosis:**
```bash
# Check field types
streamv3 read-csv data.csv | jq '.age | type' | sort | uniq -c

# Find mixed types
streamv3 read-csv data.csv | jq 'select(.age | type != "number")'
```

**Solution:**
```bash
# CSV auto-parsing should handle this, but if you have manual JSONL:
streamv3 read-json data.jsonl | jq '.age |= tonumber' | streamv3 where -match age gt 30
```

### Issue: "GROUP BY produces unexpected results"

**Diagnosis:**
```bash
# Check grouping key values
streamv3 read-csv data.csv | jq -r '.department' | sort | uniq -c

# Look for null/empty values
streamv3 read-csv data.csv | jq 'select(.department == null or .department == "")'

# Verify manual count matches GROUP BY count
DEPT="Engineering"
echo "Manual count: $(streamv3 read-csv data.csv | jq "select(.department == \"$DEPT\")" | wc -l)"
echo "GROUP BY count: $(streamv3 read-csv data.csv | streamv3 group -by department -function count -result n | jq "select(.department == \"$DEPT\") | .n")"
```

---

## Interactive Debugging

### Use jq's Interactive Mode

```bash
# Save intermediate results
streamv3 read-csv data.csv | streamv3 where -match age gt 30 > /tmp/filtered.jsonl

# Explore interactively with jq
jq '.' /tmp/filtered.jsonl | less

# Try different queries
jq 'select(.department == "Engineering")' /tmp/filtered.jsonl
jq 'group_by(.department) | map({dept: .[0].department, count: length})' /tmp/filtered.jsonl
```

### Build Pipeline Incrementally

```bash
# Start simple
streamv3 read-csv data.csv | jq '.'

# Add one filter
streamv3 read-csv data.csv | streamv3 where -match age gt 30 | jq '.'

# Add another filter
streamv3 read-csv data.csv | \
  streamv3 where -match age gt 30 | \
  streamv3 where -match department eq Engineering | \
  jq '.'

# Add aggregation
streamv3 read-csv data.csv | \
  streamv3 where -match age gt 30 | \
  streamv3 where -match department eq Engineering | \
  streamv3 group -by department -function avg -field salary -result avg_salary | \
  jq '.'
```

---

## Best Practices

### 1. Always Inspect Sample Data First

```bash
# Before building complex pipeline, look at the data
streamv3 read-csv data.csv | jq '.' | head -3
```

### 2. Test Filters Incrementally

```bash
# Add filters one at a time, checking counts
streamv3 read-csv data.csv | wc -l
streamv3 read-csv data.csv | streamv3 where -match age gt 30 | wc -l
streamv3 read-csv data.csv | streamv3 where -match age gt 30 | streamv3 where -match status eq active | wc -l
```

### 3. Use Limit During Development

```bash
# Work with small sample while developing
streamv3 read-csv huge.csv | streamv3 limit -n 100 | \
  streamv3 where ... | \
  streamv3 group ... | \
  jq '.'
```

### 4. Save Intermediate Results

```bash
# Save stages for debugging
streamv3 read-csv data.csv > /tmp/1-input.jsonl
cat /tmp/1-input.jsonl | streamv3 where -match age gt 30 > /tmp/2-filtered.jsonl
cat /tmp/2-filtered.jsonl | streamv3 group -by dept -function count -result n > /tmp/3-grouped.jsonl

# Inspect any stage
jq '.' /tmp/2-filtered.jsonl | less
```

### 5. Use Verbose/Debug Output

```bash
# Count records at each stage
streamv3 read-csv data.csv | tee >(wc -l >&2) | \
  streamv3 where -match age gt 30 | tee >(wc -l >&2) | \
  streamv3 group -by dept -function count -result n | tee >(wc -l >&2) | \
  streamv3 write-csv output.csv
```

---

## Quick Reference

### jq One-Liners

```bash
# Pretty-print
jq '.'

# Extract field
jq '.fieldname'

# Select multiple fields
jq '{field1, field2, field3}'

# Filter records
jq 'select(.age > 30)'

# Count records
jq -s 'length'

# List all keys
jq 'keys'

# Check type
jq '.field | type'

# Convert to number
jq '.field | tonumber'

# Sum values
jq -s 'map(.field) | add'

# Average
jq -s 'map(.field) | add / length'

# Unique values
jq -r '.field' | sort -u

# Group and count
jq -s 'group_by(.field) | map({key: .[0].field, count: length})'
```

### StreamV3 + jq Patterns

```bash
# Inspect pipeline
streamv3 read-csv data.csv | streamv3 where ... | jq '.'

# Count results
streamv3 read-csv data.csv | streamv3 where ... | jq -s 'length'

# Extract single field
streamv3 read-csv data.csv | jq -r '.name'

# Validate GROUP BY
streamv3 ... | streamv3 group ... | jq 'select(.department == "Engineering")'

# Debug type issues
streamv3 read-csv data.csv | jq '.age | type' | sort | uniq -c
```

---

## Additional Resources

- [jq Manual](https://jqlang.github.io/jq/manual/)
- [jq Cookbook](https://github.com/stedolan/jq/wiki/Cookbook)
- [StreamV3 README](../README.md)
- [StreamV3 Examples](../examples/)

---

## Getting Help

If you encounter issues not covered here:

1. Check the [troubleshooting guide](./troubleshooting.md)
2. Look at [examples](../examples/)
3. File an issue: https://github.com/rosscartlidge/streamv3/issues
