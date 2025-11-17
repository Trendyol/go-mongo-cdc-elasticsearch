# go-mongo-cdc-elasticsearch

[![Build](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/actions/workflows/build.yml/badge.svg)](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/actions/workflows/build.yml)
[![Scorecard](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/actions/workflows/scorecard.yml/badge.svg)](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/actions/workflows/scorecard.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Trendyol/go-mongo-cdc-elasticsearch)](https://goreportcard.com/report/github.com/Trendyol/go-mongo-cdc-elasticsearch)
[![codecov](https://codecov.io/gh/Trendyol/go-mongo-cdc-elasticsearch/branch/main/graph/badge.svg)](https://codecov.io/gh/Trendyol/go-mongo-cdc-elasticsearch)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

MongoDB Change Data Capture (CDC) connector for Elasticsearch. This library provides real-time data synchronization from MongoDB to Elasticsearch using MongoDB Change Streams.

## Features

- 🚀 **Real-time Sync**: Automatically syncs MongoDB changes to Elasticsearch
- 📊 **Change Streams**: Uses MongoDB Change Streams for reliable CDC
- 🔄 **Full Operations**: Supports insert, update, and delete operations
- 🎯 **Custom Mapping**: Flexible document transformation with custom mappers
- 📦 **Bulk Processing**: Efficient bulk indexing to Elasticsearch
- 🔧 **Configurable**: Extensive configuration options via YAML
- 🏗️ **Sharding Support**: Works with MongoDB sharded clusters
- 📈 **Metrics**: Prometheus metrics for monitoring
- 🔍 **Partitioning**: Distributed processing with partition management
- ⚡ **High Performance**: Optimized for high-throughput scenarios

## Installation

```bash
go get github.com/Trendyol/go-mongo-cdc-elasticsearch
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    cdcelasticsearch "github.com/Trendyol/go-mongo-cdc-elasticsearch"
)

func main() {
    connector, err := cdcelasticsearch.NewConnectorBuilder("config.yml").Build()
    if err != nil {
        panic(err)
    }
    defer connector.Close()
    
    connector.Start(context.Background())
}
```

### Configuration

Create a `config.yml` file:

```yaml
mongodb:
  connection:
    uri: "localhost:27017"
    database: "mydb"
    collection: "mycollection"
  connectionPool:
    maxPoolSize: 100
    minPoolSize: 5

elasticsearch:
  urls:
    - "http://localhost:9200"
  collectionIndexMapping:
    mycollection: "my-index"
  batchSizeLimit: 1000
  batchTickerDuration: 5s

metric:
  port: 8080

partition:
  totalPartition: 1
  heartbeatInterval: 5s

logger:
  logLevel: "info"
```

### Custom Mapper

```go
func customMapper(event mongodb.Event) []document.ESActionDocument {
    if event.IsMutated {
        // Transform your data
        transformedData := transformData(event.Value)
        doc := document.NewIndexAction(event.Key, transformedData, nil)
        return []document.ESActionDocument{doc}
    }
    doc := document.NewDeleteAction(event.Key, nil)
    return []document.ESActionDocument{doc}
}

connector, err := cdcelasticsearch.NewConnectorBuilder("config.yml").
    SetMapper(customMapper).
    Build()
```

## Architecture

```
MongoDB → Change Streams → go-mongo-cdc → Mapper → Bulk Processor → Elasticsearch
```

The connector:
1. Listens to MongoDB Change Streams
2. Processes change events through the mapper
3. Batches documents for efficient indexing
4. Bulk indexes to Elasticsearch
5. Manages checkpoints for reliability

## Testing

### Run All Tests

```bash
# Run integration tests
make test-integration

# Run unit tests
make test-unit

# Test all CI checks locally (lint + build + tests)
make test-ci-local
```

### Testing Locally

You can test the entire CI pipeline locally before pushing:

```bash
# Run all CI checks (lint, build, unit tests, integration tests)
make test-ci-local
```

This will:
1. ✅ Check Go version
2. ✅ Install dependencies
3. ✅ Run linter (if installed)
4. ✅ Run unit tests
5. ✅ Build project
6. ✅ Run integration tests

### CI/CD

The project includes GitHub Actions workflows for:
- **Build**: Linting, unit tests, build verification, and security gates
- **Integration**: Full end-to-end testing with MongoDB and Elasticsearch
- **Release**: Automated releases with GoReleaser
- **Scorecard**: Supply-chain security scanning

See [TESTING.md](TESTING.md) for detailed testing guide.

## Configuration Options

### MongoDB Configuration

- `uri`: MongoDB connection URI (without `mongodb://` prefix)
- `database`: Target database name
- `collection`: Target collection name
- `connectionPool`: Connection pool settings
- `timeouts`: Connection timeout settings

### Elasticsearch Configuration

- `urls`: List of Elasticsearch URLs
- `collectionIndexMapping`: MongoDB collection to Elasticsearch index mapping
- `batchSizeLimit`: Maximum documents per batch
- `batchTickerDuration`: Batch flush interval
- `compressionEnabled`: Enable/disable compression
- `maxRetries`: Maximum retry attempts

### Partition Configuration

- `totalPartition`: Number of partitions for distributed processing
- `heartbeatInterval`: Worker heartbeat interval
- `workerTimeout`: Worker timeout duration

See [example configurations](example/) for more details.

## Examples

- [Simple Example](example/simple/) - Basic usage
- [Custom Mapper](example/custom-mapper/) - Custom data transformation
- [Default Mapper](example/default-mapper/) - Using default mapper
- [Struct Config](example/struct-config/) - Programmatic configuration

## Monitoring

The connector exposes Prometheus metrics on the configured port (default: 8080):

```bash
curl http://localhost:8080/metrics
```

Available metrics:
- MongoDB change stream events
- Elasticsearch bulk operations
- Processing latency
- Error rates
- Partition assignments

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [go-mongo-cdc](https://github.com/Trendyol/go-mongo-cdc) - MongoDB CDC library
- [go-dcp-elasticsearch](https://github.com/Trendyol/go-dcp-elasticsearch) - Couchbase DCP to Elasticsearch connector

## Support

For issues and questions:
- Open an [issue](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/issues)
- Check [documentation](test/integration/README.md)
- Review [examples](example/)

