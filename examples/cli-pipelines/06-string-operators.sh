#!/bin/bash
# Example 6: String Operators
# Demonstrate contains, startswith, endswith operators

set -e

echo "Example 6: String Matching Operators"
echo "====================================="
echo

# Create sample data with email addresses
cat > /tmp/users.csv <<EOF
name,email,department
Alice,alice@engineering.com,Engineering
Bob,bob@sales.org,Sales
Carol,carol@engineering.com,Engineering
David,david@marketing.net,Marketing
Eve,eve.smith@engineering.com,Engineering
Frank,frank@sales.org,Sales
EOF

echo "Input data (users.csv):"
cat /tmp/users.csv
echo
echo "---"
echo

# Example 1: contains operator
echo "Example 6a: Find emails containing 'engineering'"
echo "Command:"
echo "  streamv3 read-csv /tmp/users.csv | \\"
echo "    streamv3 where -match email contains engineering | \\"
echo "    streamv3 write-csv"
echo
echo "Results:"

streamv3 read-csv /tmp/users.csv | \
  streamv3 where -match email contains engineering | \
  streamv3 write-csv

echo
echo "---"
echo

# Example 2: endswith operator
echo "Example 6b: Find emails ending with '.org'"
echo "Command:"
echo "  streamv3 read-csv /tmp/users.csv | \\"
echo "    streamv3 where -match email endswith .org | \\"
echo "    streamv3 write-csv"
echo
echo "Results:"

streamv3 read-csv /tmp/users.csv | \
  streamv3 where -match email endswith .org | \
  streamv3 write-csv

echo
echo "---"
echo

# Example 3: startswith operator
echo "Example 6c: Find names starting with 'C'"
echo "Command:"
echo "  streamv3 read-csv /tmp/users.csv | \\"
echo "    streamv3 where -match name startswith C | \\"
echo "    streamv3 write-csv"
echo
echo "Results:"

streamv3 read-csv /tmp/users.csv | \
  streamv3 where -match name startswith C | \
  streamv3 write-csv

echo
echo "String operators available:"
echo "  - contains: String contains substring"
echo "  - startswith: String starts with prefix"
echo "  - endswith: String ends with suffix"
