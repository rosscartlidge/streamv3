#!/bin/bash
# Main test runner for AI code generation validation
# Run this regularly to validate that reference implementations still work

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}StreamV3 AI Code Generation Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

cd "$PROJECT_ROOT"

echo -e "${BLUE}Testing Reference Implementations...${NC}"
echo ""

# Test each file
for test_file in test-output/test_case_*.go; do
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    test_name=$(basename "$test_file")

    echo -e "${BLUE}Test $TOTAL_TESTS: $test_name${NC}"

    # Validate patterns
    if ./scripts/validate-ai-patterns.sh "$test_file" >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓ Validation passed${NC}"
    else
        echo -e "  ${RED}✗ Validation failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        continue
    fi

    # Run the code
    if go run "$test_file" >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓ Execution passed${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "  ${RED}✗ Execution failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    echo ""
done

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo "Total tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo ""

if [ "$FAILED_TESTS" -eq 0 ]; then
    echo -e "${GREEN}✓ All AI code generation tests passed!${NC}"
    echo ""
    echo -e "${BLUE}Next Steps:${NC}"
    echo "1. Copy the AI prompt:"
    echo "   sed -n '8,401p' doc/ai-code-generation.md"
    echo ""
    echo "2. Test with your LLM using prompts from:"
    echo "   test-ai-generation-cases.md"
    echo ""
    echo "3. Validate generated code:"
    echo "   ./scripts/validate-ai-patterns.sh <generated-file.go>"
    echo ""
    exit 0
else
    echo -e "${RED}✗ Some tests failed.${NC}"
    exit 1
fi
