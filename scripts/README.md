# Scripts

This directory contains utility scripts for development and testing.

## test-ci-locally.sh

Runs all CI checks locally before pushing to GitHub.

### Usage

```bash
# From project root
./scripts/test-ci-locally.sh

# Or using make
make test-ci-local
```

### What it does

1. **Go Version Check**: Verifies Go installation
2. **Dependencies**: Downloads all required modules
3. **Lint**: Runs golangci-lint (if installed)
4. **Unit Tests**: Runs all unit tests
5. **Build**: Builds the project
6. **Integration Tests**: Runs full integration test suite

### Requirements

- Go 1.25+
- Docker (for integration tests)
- golangci-lint (optional, for linting)

### Example Output

```bash
🚀 Testing CI locally...

📦 Checking Go version...
go version go1.25.1 darwin/arm64

📥 Installing dependencies...
go: downloading ...

🔍 Running lint...
✅ Lint passed

🧪 Running unit tests...
✅ All unit tests passed

🔨 Building project...
✅ Build successful

🔗 Running integration tests...
✅ All integration tests passed

✅ All CI checks passed locally!
```

### Troubleshooting

If the script fails:

1. **Lint errors**: Run `make linter` to auto-fix issues
2. **Unit test failures**: Check test output for details
3. **Integration test failures**: Ensure Docker is running and ports are available
4. **Build errors**: Run `go mod tidy` to clean up dependencies

### CI/CD Integration

This script mimics the GitHub Actions workflows:
- `.github/workflows/build.yml` - Build and test
- `.github/workflows/integration.yml` - Integration tests

By running this locally, you can catch issues before pushing to GitHub.

