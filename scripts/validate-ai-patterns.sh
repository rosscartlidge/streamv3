#!/bin/bash
# Validates that code follows StreamV3 best practices and API patterns

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

if [ $# -ne 1 ]; then
    echo "Usage: $0 <go-file>"
    exit 1
fi

FILE="$1"

if [ ! -f "$FILE" ]; then
    echo -e "${RED}Error: File $FILE does not exist${NC}"
    exit 1
fi

echo -e "${BLUE}Validating StreamV3 patterns in: $FILE${NC}"
echo ""

ERRORS=0
WARNINGS=0

# Check 1: Correct import path
echo -n "Checking import path... "
if grep -q '"github.com/rosscartlidge/streamv3"' "$FILE"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo "  Error: Must import github.com/rosscartlidge/streamv3"
    ((ERRORS++))
fi

# Check 2: No wrong import paths
echo -n "Checking for wrong imports... "
if grep -q '"github.com/rocketlaunchr/streamv3"' "$FILE"; then
    echo -e "${RED}✗${NC}"
    echo "  Error: Found wrong import path (rocketlaunchr instead of rosscartlidge)"
    ((ERRORS++))
else
    echo -e "${GREEN}✓${NC}"
fi

# Check 3: SQL-style naming (Where not Filter)
echo -n "Checking SQL-style API usage... "
FILTER_USAGE=$(grep -c "streamv3\.Filter(" "$FILE" || true)
if [ "$FILTER_USAGE" -gt 0 ]; then
    echo -e "${YELLOW}⚠${NC}"
    echo "  Warning: Found streamv3.Filter() - should use streamv3.Where() for filtering"
    ((WARNINGS++))
else
    echo -e "${GREEN}✓${NC}"
fi

# Check 4: Error handling for ReadCSV
echo -n "Checking error handling... "
if grep -q "ReadCSV(" "$FILE"; then
    if grep -q "if err != nil" "$FILE"; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        echo "  Error: ReadCSV used but no error handling found"
        ((ERRORS++))
    fi
else
    echo -e "${BLUE}N/A${NC} (no CSV reading)"
fi

# Check 5: Proper GroupByFields usage
echo -n "Checking GroupByFields usage... "
if grep -q "GroupByFields(" "$FILE"; then
    # Check for wrong API: GroupByFields([]string{...}, []Aggregation{...})
    if grep -q 'GroupByFields(\s*\[\]string{' "$FILE"; then
        echo -e "${RED}✗${NC}"
        echo "  Error: Wrong GroupByFields API - should be GroupByFields(namespace, field1, field2, ...)"
        ((ERRORS++))
    else
        echo -e "${GREEN}✓${NC}"
    fi
else
    echo -e "${BLUE}N/A${NC} (no grouping)"
fi

# Check 6: Proper Aggregate usage
echo -n "Checking Aggregate usage... "
if grep -q "Aggregate(" "$FILE"; then
    # Check for wrong Count syntax: Count("field_name")
    if grep -q 'Count("' "$FILE" && ! grep -q 'Count()' "$FILE"; then
        echo -e "${RED}✗${NC}"
        echo "  Error: Wrong Count() syntax - should be Count() without field name in map"
        ((ERRORS++))
    else
        # Check for proper map syntax
        if grep -q 'map\[string\]streamv3\.AggregateFunc{' "$FILE"; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${YELLOW}⚠${NC}"
            echo "  Warning: Should use map[string]streamv3.AggregateFunc{...}"
            ((WARNINGS++))
        fi
    fi
else
    echo -e "${BLUE}N/A${NC} (no aggregation)"
fi

# Check 7: Proper use of Chain or Pipe
echo -n "Checking composition style... "
if grep -q "Chain(" "$FILE" || grep -q "Pipe(" "$FILE"; then
    echo -e "${GREEN}✓${NC}"
elif grep -q "streamv3\.Where\|streamv3\.Select\|streamv3\.GroupByFields" "$FILE"; then
    echo -e "${YELLOW}⚠${NC}"
    echo "  Warning: Consider using Chain() for better readability"
    ((WARNINGS++))
else
    echo -e "${BLUE}N/A${NC} (simple pipeline)"
fi

# Check 8: Compilation test
echo -n "Checking compilation... "
if go build -o /tmp/validate_test "$FILE" 2>/dev/null; then
    echo -e "${GREEN}✓${NC}"
    rm -f /tmp/validate_test
else
    echo -e "${RED}✗${NC}"
    echo "  Error: Code does not compile"
    ((ERRORS++))
fi

echo ""
echo -e "${BLUE}Validation Summary:${NC}"
echo "  Errors: $ERRORS"
echo "  Warnings: $WARNINGS"

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ Passed with warnings${NC}"
    exit 0
else
    echo -e "${RED}✗ Validation failed${NC}"
    exit 1
fi
