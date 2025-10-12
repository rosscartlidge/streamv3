#!/bin/bash
# Example 5: Full Pipeline - End-to-End Data Processing
# Filter, select, sort, limit, and export to new CSV

set -e

echo "Example 5: Complete Pipeline - Engineering Top Performers Report"
echo "================================================================="
echo

# Create sample data
cat > /tmp/employees.csv <<EOF
name,age,department,salary
Alice,30,Engineering,95000
Bob,25,Marketing,65000
Carol,35,Engineering,105000
David,28,Sales,70000
Eve,32,Engineering,98000
Frank,45,Marketing,80000
Grace,29,Sales,72000
Henry,38,Engineering,110000
Isabel,27,Engineering,88000
Jack,42,Marketing,92000
EOF

echo "Input data (employees.csv):"
cat /tmp/employees.csv
echo
echo "---"
echo

# Run the complete pipeline
echo "Pipeline Steps:"
echo "  1. Read CSV"
echo "  2. Filter: Engineering department only"
echo "  3. Select: name, age, salary fields"
echo "  4. Sort: by salary descending"
echo "  5. Limit: top 3"
echo "  6. Write: to output CSV"
echo
echo "Command:"
echo "  streamv3 read-csv /tmp/employees.csv | \\"
echo "    streamv3 where -match department eq Engineering | \\"
echo "    streamv3 select -field name + -field age + -field salary | \\"
echo "    streamv3 sort -field salary -desc | \\"
echo "    streamv3 limit -n 3 | \\"
echo "    streamv3 write-csv > /tmp/top_engineers.csv"
echo

streamv3 read-csv /tmp/employees.csv | \
  streamv3 where -match department eq Engineering | \
  streamv3 select -field name + -field age + -field salary | \
  streamv3 sort -field salary -desc | \
  streamv3 limit -n 3 | \
  streamv3 write-csv > /tmp/top_engineers.csv

echo "Output saved to: /tmp/top_engineers.csv"
echo
echo "Results:"
cat /tmp/top_engineers.csv
echo
echo "Top 3 Engineering performers:"
echo "  1. Henry (38) - 110,000"
echo "  2. Carol (35) - 105,000"
echo "  3. Eve (32) - 98,000"
