# Testing Guide

This document explains how to run tests for the `go-mongo-cdc-elasticsearch` project.

## Test Types

### 1. Unit Tests
Tests individual functions and modules.

### 2. Integration Tests
Real integration tests with MongoDB and Elasticsearch.

## Running Integration Tests

### Method 1: Complete Integration Test (Recommended) ⭐

Run all integration tests with a single command:

```bash
# From project root directory
make test-integration
```

This command:
1. Starts all Docker containers
2. Waits for services to be healthy (~10-20 seconds)
3. Runs all integration tests sequentially
4. Stops and cleans up all containers

**Advantages:**
- ✅ Everything works with a single command
- ✅ Ideal for CI/CD
- ✅ Clean and isolated environment
- ✅ Automatic cleanup

### Method 2: Step-by-Step with Makefile

```bash
# Start test environment
make test-integration-up

# Run tests
make test-integration-run

# View logs (optional)
make test-integration-logs

# Stop environment
make test-integration-down
```

**Advantages:**
- ✅ More control
- ✅ Suitable for debugging
- ✅ Step-by-step execution
- ✅ Manual access to services

### Method 3: Manual Docker Compose

```bash
# Start services
docker-compose -f test/integration/docker-compose.yml up -d

# Wait for services to be ready
# (The services will automatically become healthy)

# Run tests
cd test/integration
go test -v -timeout 10m

# Stop services
cd ../..
docker-compose -f test/integration/docker-compose.yml down
```

## Test Scenarios

### Run All Tests

```bash
# Method 1: Complete (Recommended)
make test-integration

# Method 2: Manual
cd test/integration
go test -v -timeout 10m
```

### Run Specific Test

```bash
# First, start test environment
make test-integration-up

# Run specific test
cd test/integration
go test -v -run TestIntegration_BasicInsertOperation -timeout 5m

# Stop environment
cd ../..
make test-integration-down
```

### Available Tests

1. **TestIntegration_BasicInsertOperation**
   - Insert to MongoDB → Sync to Elasticsearch

2. **TestIntegration_MultipleInserts**
   - Bulk insert of 10 documents

3. **TestIntegration_UpdateOperation**
   - Update in MongoDB → Update in Elasticsearch

4. **TestIntegration_DeleteOperation**
   - Delete from MongoDB → Delete from Elasticsearch

5. **TestIntegration_CustomMapper**
   - Custom mapper function test

## Service Control

### MongoDB

```bash
# Connect to container
docker exec -it mongodb-router-test mongosh

# In MongoDB shell
show dbs
use testdb
db.testcollection.find()
```

### Elasticsearch

```bash
# Health check
curl http://localhost:9200/_cluster/health?pretty

# List indices
curl http://localhost:9200/_cat/indices?v

# Count documents
curl http://localhost:9200/test-index/_count

# List documents
curl http://localhost:9200/test-index/_search?pretty
```

## Viewing Logs

### All Service Logs

```bash
# Using docker-compose
docker-compose -f test/integration/docker-compose.yml logs -f

# OR using Makefile
make test-integration-logs
```

### Specific Service Logs

```bash
# MongoDB
docker-compose -f test/integration/docker-compose.yml logs -f mongodb-router-test

# Elasticsearch
docker-compose -f test/integration/docker-compose.yml logs -f elasticsearch-test
```

## Troubleshooting

### Containers Won't Start

```bash
# Check container status
docker ps -a

# Check logs
docker-compose -f test/integration/docker-compose.yml logs

# Restart
docker-compose -f test/integration/docker-compose.yml down -v
docker-compose -f test/integration/docker-compose.yml up -d
```

### Port Conflicts

```bash
# Check running containers
docker ps

# Check ports
lsof -i :27017  # MongoDB
lsof -i :9200   # Elasticsearch

# Clean up old containers
docker-compose -f test/integration/docker-compose.yml down -v
```

### Tests Timeout

```bash
# Use longer timeout
go test -v -timeout 20m

# Or increase healthcheck retries in docker-compose.yml
```

### Image Build Errors

```bash
# Clear cache and rebuild
docker-compose -f test/integration/docker-compose.yml build --no-cache

# Or
docker system prune -a
docker-compose -f test/integration/docker-compose.yml up --build
```

## Cleanup

### Clean All Containers and Volumes

```bash
# Using Makefile
make test-integration-clean

# OR using docker-compose
docker-compose -f test/integration/docker-compose.yml down -v
```

### Docker System Cleanup

```bash
# Clean unused images
docker image prune -a

# Clean entire system (CAUTION!)
docker system prune -a --volumes
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Run Integration Tests
        run: make test-integration
```

### GitLab CI Example

```yaml
integration-test:
  stage: test
  image: docker:latest
  services:
    - docker:dind
  script:
    - make test-integration
  after_script:
    - docker-compose -f test/integration/docker-compose.yml down -v
```

## Performance Tips

1. **First Build**: May take 5-10 minutes (images are built)
2. **Subsequent Builds**: 1-2 minutes thanks to cache
3. **Test Duration**: ~5-10 minutes (all tests)
4. **RAM Usage**: ~2-3 GB
5. **Disk Usage**: ~2 GB

## Recommended Workflow

### During Development

```bash
# 1. Start test environment (once)
make test-integration-up

# 2. Make code changes

# 3. Run tests (repeatedly)
cd test/integration
go test -v -run TestIntegration_YourTest -timeout 5m

# 4. Stop environment when done
cd ../..
make test-integration-down
```

### In CI/CD

```bash
# Single command - does everything
make test-integration
```

## Additional Resources

- [Integration Test README](test/integration/README.md)
- [MongoDB Test Infrastructure](test/mongodb/README.md)
- [Elasticsearch Test Infrastructure](test/elasticsearch/README.md)

## Help

```bash
# View all make commands
make help

# Or read the Makefile
cat Makefile
```

## Project Structure

```
go-mongo-cdc-elasticsearch/
├── Makefile                    # Build and test commands
├── test/
│   ├── mongodb/
│   │   ├── Dockerfile         # Custom MongoDB image
│   │   ├── configure.sh       # MongoDB setup script
│   │   └── README.md
│   ├── elasticsearch/
│   │   ├── Dockerfile         # Custom Elasticsearch image
│   │   ├── config/
│   │   │   └── elasticsearch.yml
│   │   └── README.md
│   └── integration/
│       ├── docker-compose.yml # Test environment setup
│       ├── config/
│       │   └── test-config.yml
│       ├── scripts/
│       │   └── setup-mongodb.sh
│       ├── helpers.go         # Test helper functions
│       ├── integration_test.go # Integration tests
│       └── README.md
```

## Notes

- Tests run **sequentially** to avoid Prometheus metrics collision
- Each test uses a **unique metrics port** automatically
- Services use **health checks** for smart waiting
- **No fixed sleep times** - tests wait for actual conditions
- All tests are **isolated** and **independent**
