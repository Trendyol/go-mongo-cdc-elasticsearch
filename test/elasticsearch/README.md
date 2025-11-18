# Elasticsearch Test Infrastructure

This directory contains Elasticsearch-specific test infrastructure including Docker configuration and custom settings.

## Contents

### Dockerfile
Custom Elasticsearch Docker image based on official Elasticsearch 7.17.10 with:
- Custom configuration for testing
- Security disabled for easier testing
- Optimized settings for test environment
- Single-node discovery mode

### config/elasticsearch.yml
Elasticsearch configuration optimized for testing:
- Single-node cluster mode
- Security features disabled
- Network accessible from all interfaces
- Debug logging for discovery

### config/log4j2.properties
Logging configuration with:
- Console and file appenders
- Deprecation logging
- Configurable log levels
- Rolling file policies

## Usage

### Building the Image

```bash
cd test/elasticsearch
docker build -t elasticsearch-test .
```

### Running the Container

```bash
docker run -d \
  --name es-test \
  -p 9200:9200 \
  -p 9300:9300 \
  -e "discovery.type=single-node" \
  -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
  elasticsearch-test
```

### Testing the Connection

```bash
# Health check
curl http://localhost:9200/_cluster/health?pretty

# Cluster info
curl http://localhost:9200

# Create an index
curl -X PUT http://localhost:9200/test-index

# Index a document
curl -X POST http://localhost:9200/test-index/_doc \
  -H 'Content-Type: application/json' \
  -d '{"field": "value"}'
```

## Configuration

### Environment Variables

- `discovery.type`: Set to `single-node` for testing
- `ES_JAVA_OPTS`: JVM options (default: `-Xms512m -Xmx512m`)
- `xpack.security.enabled`: Security features (default: `false`)
- `cluster.name`: Cluster name (default: `test-cluster`)
- `node.name`: Node name (default: `test-node`)

### Ports

- **9200**: HTTP REST API
- **9300**: Transport protocol (node-to-node communication)

### Memory Settings

Default memory allocation:
- Heap size: 512MB (min and max)
- Suitable for testing with limited resources
- Can be adjusted via `ES_JAVA_OPTS`

## Integration with docker-compose

This image is used in the integration test docker-compose setup:

```yaml
elasticsearch:
  build:
    context: ../elasticsearch
    dockerfile: Dockerfile
  environment:
    - discovery.type=single-node
    - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
  ports:
    - "9200:9200"
    - "9300:9300"
```

## Features

- ✅ Single-node cluster for testing
- ✅ Security disabled for easier access
- ✅ Custom logging configuration
- ✅ Health check support
- ✅ Optimized for test environment
- ✅ Based on official Elasticsearch image

## Testing

### Basic Health Check

```bash
# Wait for Elasticsearch to be ready
until curl -s http://localhost:9200/_cluster/health | grep -q '"status":"green"'; do
  echo "Waiting for Elasticsearch..."
  sleep 2
done
echo "Elasticsearch is ready!"
```

### Index Operations

```bash
# Create index with mapping
curl -X PUT http://localhost:9200/test-index \
  -H 'Content-Type: application/json' \
  -d '{
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0
    },
    "mappings": {
      "properties": {
        "name": { "type": "text" },
        "value": { "type": "integer" }
      }
    }
  }'

# Bulk index documents
curl -X POST http://localhost:9200/_bulk \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @test-data.ndjson
```

### Search Operations

```bash
# Search all documents
curl http://localhost:9200/test-index/_search?pretty

# Count documents
curl http://localhost:9200/test-index/_count

# Get specific document
curl http://localhost:9200/test-index/_doc/1
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker logs elasticsearch-test

# Common issues:
# - Insufficient memory
# - Port already in use
# - Disk space full
```

### Out of Memory
```bash
# Increase heap size
docker run -e "ES_JAVA_OPTS=-Xms1g -Xmx1g" elasticsearch-test
```

### Connection Refused
```bash
# Check if Elasticsearch is running
curl http://localhost:9200

# Check container status
docker ps | grep elasticsearch

# Check network
docker network inspect test-network
```

### Slow Startup
```bash
# Monitor startup progress
docker logs -f elasticsearch-test

# Elasticsearch typically takes 20-30 seconds to start
```

## Performance Tuning

### For Faster Tests

```yaml
# Reduce refresh interval
PUT /test-index/_settings
{
  "index": {
    "refresh_interval": "1s"
  }
}
```

### For More Documents

```bash
# Increase heap size
ES_JAVA_OPTS="-Xms1g -Xmx1g"

# Increase thread pool
thread_pool.write.queue_size: 1000
```

## Future Tests

Planned Elasticsearch-specific tests:
- `bulk_test.go`: Test bulk processor logic
- `document_test.go`: Test document action creation
- `client_test.go`: Test client initialization
- `response_handler_test.go`: Test response handling
- `metrics_test.go`: Test metrics collection
- `index_test.go`: Test index operations
- `search_test.go`: Test search functionality

## Monitoring

### Cluster Stats

```bash
# Cluster health
curl http://localhost:9200/_cluster/health?pretty

# Node stats
curl http://localhost:9200/_nodes/stats?pretty

# Index stats
curl http://localhost:9200/_stats?pretty
```

### Performance Metrics

```bash
# Thread pool stats
curl http://localhost:9200/_cat/thread_pool?v

# Pending tasks
curl http://localhost:9200/_cat/pending_tasks?v

# JVM stats
curl http://localhost:9200/_nodes/stats/jvm?pretty
```

## Contributing

When modifying the Elasticsearch test infrastructure:
1. Test with different data volumes
2. Verify health checks work correctly
3. Ensure proper cleanup on shutdown
4. Update documentation for configuration changes
5. Test with different Elasticsearch versions
6. Verify compatibility with the Go client library
