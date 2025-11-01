#!/bin/bash
# Generate version.txt files from git describe
# This script is run before building to embed the version

set -e

# Get version from git describe
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")

# Write to both locations (cmd/streamv3 and internal/version)
echo "$VERSION" > cmd/streamv3/version.txt
echo "$VERSION" > internal/version/version.txt

echo "Version updated to: $VERSION" >&2
