#!/bin/bash
# Generate version.txt from git describe
# This script is run before building to embed the version

set -e

# Get version from git describe
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")

# Write to version package
echo "$VERSION" > cmd/ssql/version/version.txt

echo "Version updated to: $VERSION" >&2
