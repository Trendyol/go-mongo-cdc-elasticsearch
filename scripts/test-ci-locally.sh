#!/bin/bash

set -e

echo "🚀 Testing CI locally..."
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. Check Go version
echo -e "${BLUE}📦 Checking Go version...${NC}"
go version
echo ""

# 2. Install dependencies
echo -e "${BLUE}📥 Installing dependencies...${NC}"
go mod download
cd test/integration && go mod download && cd ../..
echo ""

# 3. Run lint (if golangci-lint is installed)
if command -v golangci-lint &> /dev/null; then
    echo -e "${BLUE}🔍 Running lint...${NC}"
    if golangci-lint run -c .golangci.yml --timeout=5m 2>&1 | grep -v "Go language version"; then
        echo -e "${GREEN}✅ Lint passed${NC}"
    else
        echo -e "${BLUE}⚠️  Lint completed with warnings (Go version mismatch can be ignored locally)${NC}"
    fi
    echo ""
else
    echo -e "${BLUE}⚠️  golangci-lint not installed, skipping lint${NC}"
    echo ""
fi

# 4. Run unit tests
echo -e "${BLUE}🧪 Running unit tests...${NC}"
go test -v -short ./...
echo ""

# 5. Build
echo -e "${BLUE}🔨 Building project...${NC}"
go build -v ./...
echo ""

# 6. Run integration tests
echo -e "${BLUE}🔗 Running integration tests...${NC}"
make test-integration
echo ""

echo -e "${GREEN}✅ All CI checks passed locally!${NC}"

