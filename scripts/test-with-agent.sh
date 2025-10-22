#!/bin/bash
# Test AI code generation using Claude Code agents
# Sends natural language prompts to an agent and validates the generated code

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_DIR="$PROJECT_ROOT/test-output/agent-tests"
PROMPT_FILE="$PROJECT_ROOT/doc/ai-code-generation.md"

# Create test directory
mkdir -p "$TEST_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}AI Agent Code Generation Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo "This script will use Claude Code agents to generate code from natural language."
echo "Generated code will be validated against our test suite."
echo ""
echo "Test directory: $TEST_DIR"
echo "AI Prompt: $PROMPT_FILE"
echo ""

# Extract the AI prompt (lines 8-401)
PROMPT_TEXT=$(sed -n '8,401p' "$PROMPT_FILE")

echo -e "${YELLOW}Note: Agent testing requires manual coordination with Claude Code${NC}"
echo -e "${YELLOW}This script prepares test cases and validates results.${NC}"
echo ""
echo "Press Enter to continue..."
read

# Test Case 1: Basic Filtering and Grouping
echo -e "${BLUE}Test Case 1: Basic Filtering and Grouping${NC}"
echo ""
echo "Natural Language Prompt:"
echo "  'Read employee data from employees.csv, filter for employees with"
echo "   salary over 80000, group by department, and count how many"
echo "   employees are in each department'"
echo ""
echo "Expected patterns:"
echo "  - ReadCSV with error handling"
echo "  - Where clause for salary > 80000"
echo "  - GroupByFields for department"
echo "  - Aggregate with Count()"
echo ""
echo "Output file: $TEST_DIR/test_case_1_agent.go"
echo ""
echo -e "${YELLOW}Action needed: Run this agent task in Claude Code${NC}"
echo ""

# Create a marker file with the prompt
cat > "$TEST_DIR/test_case_1_prompt.txt" <<'EOF'
Read employee data from employees.csv, filter for employees with salary over 80000,
group by department, and count how many employees are in each department.

Generate complete, runnable Go code with sample data.
EOF

echo "Waiting for generated code at: $TEST_DIR/test_case_1_agent.go"
echo "Press Enter when you have saved the generated code there..."
read

if [ -f "$TEST_DIR/test_case_1_agent.go" ]; then
    echo ""
    echo -e "${BLUE}Validating generated code...${NC}"

    # Run validation
    if "$SCRIPT_DIR/validate-ai-patterns.sh" "$TEST_DIR/test_case_1_agent.go"; then
        echo ""
        echo -e "${GREEN}✓ Validation passed!${NC}"

        # Try to run it
        echo ""
        echo -e "${BLUE}Running generated code...${NC}"
        if go run "$TEST_DIR/test_case_1_agent.go" 2>&1; then
            echo ""
            echo -e "${GREEN}✓ Code executed successfully!${NC}"
        else
            echo ""
            echo -e "${RED}✗ Code failed to execute${NC}"
        fi
    else
        echo ""
        echo -e "${RED}✗ Validation failed${NC}"
        echo "See validation output above for details"
    fi
else
    echo ""
    echo -e "${YELLOW}⚠ No code file found, skipping validation${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Agent Test Complete${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Results saved in: $TEST_DIR/"
echo ""
echo "To compare with reference implementation:"
echo "  diff -u test-output/test_case_1_manual.go $TEST_DIR/test_case_1_agent.go"
echo ""
