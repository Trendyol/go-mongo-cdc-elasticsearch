# Integration Tests

This directory (`test/integration/`) contains integration tests for the `go-mongo-cdc-elasticsearch` project. These tests verify the end-to-end functionality of the CDC (Change Data Capture) connector between MongoDB and Elasticsearch.

## Overview

The integration tests use Docker Compose to set up a complete test environment including:
- MongoDB Sharded Cluster (Config Server, 2 Shards, and Router)
- Elasticsearch 7.17.10
- Test data and configurations

## Prerequisites

Before running the integration tests, ensure you have the following installed:
- Docker (20.10+)
- Docker Compose (1.29+)
- Go (1.25+)
- Make

## Test Environment

### Services

The test environment includes the following services:

1. **MongoDB Config Server** (`mongodb-config-test`)
   - Port: 27019
   - Replica Set: configReplSet

2. **MongoDB Shard 1** (`mongodb-shard1-test`)
   - Port: 27018
   - Replica Set: shard1ReplSet

3. **MongoDB Shard 2** (`mongodb-shard2-test`)
   - Port: 27028
   - Replica Set: shard2ReplSet

4. **MongoDB Router** (`mongodb-router-test`)
   - Port: 27017
   - Connects to all shards

5. **Elasticsearch** (`elasticsearch-test`)
   - Port: 9200 (HTTP)
   - Port: 9300 (Transport)
   - Single node cluster

### Test Database and Collection

- Database: `testdb`
- Collection: `testcollection`
- Elasticsearch Index: `test-index`

## Running Tests

### Quick Start - Run All Tests

To run the complete integration test suite (start environment, run tests, cleanup):

```bash
make test-integration
```

This command will:
1. Start all Docker containers
2. Wait for services to be ready (45 seconds)
3. Run all integration tests sequentially (to avoid Prometheus metrics collision)
4. Stop and remove all containers

**Note**: Tests are run sequentially to avoid duplicate Prometheus metrics registration errors.

### Manual Test Execution

For more control over the test environment, you can use the following commands:

#### 1. Start Test Environment

```bash
make test-integration-up
```

This starts all Docker containers and waits for them to be ready.

#### 2. Run Tests

Once the environment is running, execute the tests:

```bash
make test-integration-run
```

Or run specific tests:

```bash
cd test/integration && go test -v -run TestIntegration_BasicInsertOperation -timeout 5m
```

#### 3. View Logs

To monitor the test environment logs:

```bash
make test-integration-logs
```

Or view specific service logs:

```bash
docker-compose -f test/integration/docker-compose.yml logs -f mongodb-router-test
docker-compose -f test/integration/docker-compose.yml logs -f elasticsearch-test
```

#### 4. Stop Test Environment

```bash
make test-integration-down
```

#### 5. Clean Up

To completely clean up the test environment including volumes:

```bash
make test-integration-clean
```

## Test Cases

The integration test suite includes the following test cases:

### 1. TestIntegration_BasicInsertOperation
Tests basic insert operation from MongoDB to Elasticsearch.
- Inserts a single document into MongoDB
- Verifies the document appears in Elasticsearch
- Validates document content

### 2. TestIntegration_MultipleInserts
Tests multiple insert operations.
- Inserts 10 documents into MongoDB
- Verifies all documents are synced to Elasticsearch
- Validates document count

### 3. TestIntegration_UpdateOperation
Tests update operation from MongoDB to Elasticsearch.
- Inserts a document into MongoDB
- Updates the document
- Verifies the update is reflected in Elasticsearch

### 4. TestIntegration_DeleteOperation
Tests delete operation from MongoDB to Elasticsearch.
- Inserts a document into MongoDB
- Deletes the document
- Verifies the document is removed from Elasticsearch

### 5. TestIntegration_CustomMapper
Tests custom mapper functionality.
- Uses a custom mapper function
- Verifies custom mapping logic is applied

## Test Configuration

The test configuration is located in `test/integration/config/test-config.yml`. Key settings include:

- **MongoDB URI**: localhost:27017 (without `mongodb://` prefix, added automatically by go-mongo-cdc)
- **Elasticsearch URL**: http://localhost:9200
- **Batch Size**: 100 documents
- **Batch Ticker Duration**: 2 seconds
- **Total Partitions**: 1
- **Checkpoint Save Count**: 10
- **Log Level**: debug

You can modify these settings to test different scenarios.

## Helper Functions

The `helpers.go` file provides utility functions for integration tests:

### TestHelper Methods

- `SetupMongoDB(ctx)` - Connects to MongoDB
- `SetupElasticsearch()` - Connects to Elasticsearch
- `CleanupMongoDB(ctx)` - Removes all test documents
- `CleanupElasticsearch()` - Removes test index
- `InsertMongoDocument(ctx, doc)` - Inserts a document
- `UpdateMongoDocument(ctx, filter, update)` - Updates a document
- `DeleteMongoDocument(ctx, filter)` - Deletes a document
- `GetElasticsearchDocument(docID)` - Retrieves a document from ES
- `WaitForElasticsearchDocument(docID, timeout)` - Waits for document sync
- `WaitForElasticsearchDocumentCount(count, timeout)` - Waits for specific count
- `GetElasticsearchDocumentCount()` - Returns document count
- `RefreshElasticsearchIndex()` - Refreshes ES index

### Utility Functions

- `WaitForService(t, url, timeout)` - Waits for a service to be available
- `CreateTestDocument(name, value)` - Creates a test document with standard fields
- `LoadConfigWithUniqueMetricsPort(configPath)` - Loads config with unique metrics port (prevents Prometheus collision)

## Troubleshooting

### Tests Failing Due to Timeout

If tests are failing due to timeouts, try:
1. Increase the wait time in `make test-integration` (default: 45 seconds)
2. Check if all services are healthy: `docker-compose -f integration-test/docker-compose.yml ps`
3. View service logs to identify issues

### MongoDB Connection Issues

```bash
# Check MongoDB router status
docker exec mongodb-router-test mongosh --eval "db.adminCommand('ping')"

# Check shard status
docker exec mongodb-router-test mongosh --eval "sh.status()"
```

### Elasticsearch Connection Issues

```bash
# Check Elasticsearch health
curl http://localhost:9200/_cluster/health?pretty

# Check indices
curl http://localhost:9200/_cat/indices?v
```

### Port Conflicts

If you encounter port conflicts, ensure the following ports are available:
- 27017 (MongoDB Router)
- 27018 (MongoDB Shard 1)
- 27019 (MongoDB Config)
- 27028 (MongoDB Shard 2)
- 9200 (Elasticsearch HTTP)
- 9300 (Elasticsearch Transport)

### Clean Start

For a completely clean start:

```bash
# Stop all containers and remove volumes
docker-compose -f test/integration/docker-compose.yml down -v

# Remove any orphaned containers
docker-compose -f test/integration/docker-compose.yml rm -f

# Start fresh
make test-integration-up
```

## Writing New Tests

To add new integration tests:

1. Create a new test function in `integration_test.go`:

```go
func TestIntegration_YourNewTest(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    ctx := context.Background()
    helper := NewTestHelper(t)
    
    // Setup
    if err := helper.SetupMongoDB(ctx); err != nil {
        t.Fatalf("Failed to setup MongoDB: %v", err)
    }
    defer helper.Close(ctx)
    
    // Your test logic here
    
    // Cleanup
    helper.CleanupMongoDB(ctx)
    helper.CleanupElasticsearch()
}
```

2. Use the helper functions for common operations
3. Add appropriate assertions
4. Run your test: `go test -v ./test/integration/... -run TestIntegration_YourNewTest`

## CI/CD Integration

To integrate these tests into your CI/CD pipeline:

```yaml
# Example GitHub Actions workflow
- name: Run Integration Tests
  run: |
    make test-integration
```

Or for more control:

```yaml
- name: Start Test Environment
  run: make test-integration-up

- name: Run Tests
  run: make test-integration-run

- name: Cleanup
  if: always()
  run: make test-integration-down
```

## Performance Considerations

- Tests run sequentially to avoid race conditions
- Each test starts a new connector instance
- Cleanup is performed before each test
- Default timeout for operations: 30 seconds
- Default test suite timeout: 10 minutes

## Best Practices

1. **Always cleanup**: Use `defer` to ensure cleanup happens
2. **Wait for sync**: Use `WaitFor*` functions instead of fixed sleeps
3. **Refresh index**: Call `RefreshElasticsearchIndex()` after updates
4. **Check errors**: Always check and handle errors properly
5. **Use contexts**: Pass context for cancellation support
6. **Isolated tests**: Each test should be independent

## Additional Resources

- [MongoDB Change Streams Documentation](https://docs.mongodb.com/manual/changeStreams/)
- [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/7.17/index.html)
- [go-mongo-cdc Documentation](https://github.com/Trendyol/go-mongo-cdc)

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review the logs: `make test-integration-logs`
3. Open an issue in the project repository

