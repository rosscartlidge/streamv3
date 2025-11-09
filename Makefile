# ssql Makefile

.PHONY: help test build clean doc-check doc-test doc-verify doc-update fmt vet all ci install-hooks

# Default target
help:
	@echo "ssql Makefile Targets:"
	@echo ""
	@echo "Quality Checks:"
	@echo "  make test         - Run all tests"
	@echo "  make build        - Build the project"
	@echo "  make fmt          - Format all Go code"
	@echo "  make vet          - Run go vet"
	@echo ""
	@echo "Documentation Validation (3 levels):"
	@echo "  make doc-check    - Level 1: Fast checks (syntax, links, patterns)"
	@echo "  make doc-test     - Level 2: Medium checks (godoc, exports, run examples)"
	@echo "  make doc-verify   - Level 3: Deep verification (all API refs, consistency)"
	@echo "  make doc-update   - Update godoc and run validation"
	@echo ""
	@echo "Workflows:"
	@echo "  make all          - Run fmt, vet, test, doc-check (pre-push)"
	@echo "  make ci           - Full CI pipeline (all + doc-test)"
	@echo "  make release      - Release validation (ci + doc-verify)"
	@echo ""
	@echo "Setup:"
	@echo "  make install-hooks - Install git pre-commit hook"
	@echo "  make clean         - Clean build artifacts"
	@echo ""

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Build the project
build:
	@echo "Building ssql..."
	go build ./...
	@echo "Building CLI tool..."
	cd cmd/ssql && go build

# Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Level 1: Fast documentation validation
doc-check:
	@echo "Level 1: Fast Documentation Validation"
	@echo "======================================="
	@./scripts/validate-docs.sh

# Level 2: Medium documentation testing
doc-test:
	@echo "Level 2: Documentation Testing"
	@echo "==============================="
	@./scripts/doc-test.sh

# Level 3: Deep documentation verification
doc-verify:
	@echo "Level 3: Deep Documentation Verification"
	@echo "========================================="
	@./scripts/doc-verify.sh

# Update and validate documentation
doc-update: fmt
	@echo "Regenerating godoc..."
	@echo "(godoc is generated automatically from source code)"
	@echo ""
	@$(MAKE) doc-check

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean ./...
	rm -f cmd/ssql/ssql

# Run all quality checks (pre-push)
all: fmt vet test doc-check
	@echo ""
	@echo "✓ All pre-push checks passed!"

# CI target - for continuous integration
ci: all doc-test
	@echo ""
	@echo "✓ CI pipeline complete!"

# Release target - comprehensive validation
release: ci doc-verify
	@echo ""
	@echo "✓ Release validation complete!"
	@echo "✓ Ready for release!"

# Install git hooks
install-hooks:
	@echo "Installing git pre-commit hook..."
	@mkdir -p .git/hooks
	@echo '#!/bin/bash' > .git/hooks/pre-commit
	@echo 'make doc-check' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✓ Pre-commit hook installed (runs doc-check before each commit)"
