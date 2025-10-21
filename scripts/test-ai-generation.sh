#!/bin/bash
# Test harness for AI code generation validation
# Tests that the AI prompt generates correct StreamV3 code

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_DIR="$PROJECT_ROOT/test-output/ai-generation-tests"
RESULTS_FILE="$TEST_DIR/results.txt"

# Create test directory
mkdir -p "$TEST_DIR"
rm -f "$RESULTS_FILE"

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo -e "${BLUE}StreamV3 AI Code Generation Tests${NC}"
echo "===================================="
echo ""

# Test case format:
# test_case "Test Name" "Natural language prompt" "Expected patterns to find"
test_case() {
    local test_name="$1"
    local prompt="$2"
    local expected_file="$3"

    ((TOTAL_TESTS++))

    echo -e "${BLUE}Test $TOTAL_TESTS: $test_name${NC}"
    echo "Prompt: $prompt"

    # The generated code will be in $expected_file
    # We'll validate it after the agent runs

    echo "  Waiting for agent to generate code..."
    echo "  Expected output: $expected_file"
    echo ""
}

# Validation function
validate_code() {
    local test_name="$1"
    local code_file="$2"
    local should_have=("${@:3}")

    if [[ ! -f "$code_file" ]]; then
        echo -e "${RED}✗ FAILED${NC}: No code file generated"
        echo "$test_name: FAILED - No code file" >> "$RESULTS_FILE"
        ((FAILED_TESTS++))
        return 1
    fi

    # Check if it compiles
    local test_program="$TEST_DIR/test_$(basename "$code_file")"
    if ! go build -o "$test_program" "$code_file" 2>/dev/null; then
        echo -e "${RED}✗ FAILED${NC}: Code doesn't compile"
        echo "$test_name: FAILED - Compilation error" >> "$RESULTS_FILE"
        ((FAILED_TESTS++))
        return 1
    fi
    rm -f "$test_program"

    # Check for expected patterns
    local missing_patterns=()
    for pattern in "${should_have[@]}"; do
        if ! grep -q "$pattern" "$code_file"; then
            missing_patterns+=("$pattern")
        fi
    done

    if [[ ${#missing_patterns[@]} -gt 0 ]]; then
        echo -e "${RED}✗ FAILED${NC}: Missing expected patterns:"
        for pattern in "${missing_patterns[@]}"; do
            echo "    - $pattern"
        done
        echo "$test_name: FAILED - Missing patterns" >> "$RESULTS_FILE"
        ((FAILED_TESTS++))
        return 1
    fi

    echo -e "${GREEN}✓ PASSED${NC}"
    echo "$test_name: PASSED" >> "$RESULTS_FILE"
    ((PASSED_TESTS++))
    return 0
}

# Summary
show_summary() {
    echo ""
    echo -e "${BLUE}Test Summary${NC}"
    echo "============="
    echo "Total tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo ""
        echo -e "${GREEN}✓ All AI generation tests passed!${NC}"
        exit 0
    else
        echo ""
        echo -e "${RED}✗ Some tests failed. See results above.${NC}"
        exit 1
    fi
}

# Export functions for use by the agent coordination script
export -f test_case
export -f validate_code
export -f show_summary
export TEST_DIR
export RESULTS_FILE
export TOTAL_TESTS
export PASSED_TESTS
export FAILED_TESTS

echo "Test harness ready. Test directory: $TEST_DIR"
echo ""
