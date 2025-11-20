# Go Mongo CDC Elasticsearch [![Go Reference](https://pkg.go.dev/badge/github.com/Trendyol/go-mongo-cdc-elasticsearch.svg)](https://pkg.go.dev/github.com/Trendyol/go-mongo-cdc-elasticsearch) [![Build](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/actions/workflows/build.yml/badge.svg)](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/actions/workflows/build.yml) [![Scorecard](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/actions/workflows/scorecard.yml/badge.svg)](https://github.com/Trendyol/go-mongo-cdc-elasticsearch/actions/workflows/scorecard.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/Trendyol/go-mongo-cdc-elasticsearch)](https://goreportcard.com/report/github.com/Trendyol/go-mongo-cdc-elasticsearch) [![codecov](https://codecov.io/gh/Trendyol/go-mongo-cdc-elasticsearch/branch/main/graph/badge.svg)](https://codecov.io/gh/Trendyol/go-mongo-cdc-elasticsearch) [![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE) [![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/Trendyol/go-mongo-cdc-elasticsearch/badge)](https://scorecard.dev/viewer/?uri=github.com/Trendyol/go-mongo-cdc-elasticsearch)

MongoDB Change Data Capture (CDC) connector for Elasticsearch. This library provides real-time data synchronization from MongoDB to Elasticsearch using MongoDB Change Streams.

## Features

- **Near real-time synchronization** from MongoDB to Elasticsearch using MongoDB Change Streams.
- **Custom routing and mapping** via pluggable mapper functions that can emit one or many Elasticsearch actions per change event.
- **Batch processing controls** such as maximum batch size, batch bytes and batch ticker durations for efficient bulk indexing.
- **Request body compression** support for Elasticsearch bulk requests.
- **Easily manageable configurations** via YAML files or Go structs.

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
  consumerGroup: myConsumerGroup

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

## Configuration Options

### MongoDB Configuration

Check out on [go-mongo-cdc](https://github.com/Trendyol/go-mongo-cdc#configuration)

### Elasticsearch Configuration

| Variable                                    | Type              | Required | Default      | Description                                                                                                |
|---------------------------------------------|-------------------|----------|--------------|------------------------------------------------------------------------------------------------------------|
| `elasticsearch.collectionIndexMapping`      | map[string]string | yes      |              | Defines which MongoDB collection events will be written to which index                                     |
| `elasticsearch.urls`                        | []string          | yes      |              | Elasticsearch connection URLs                                                                              |
| `elasticsearch.username`                    | string            | no       |              | The username of Elasticsearch                                                                              |
| `elasticsearch.password`                    | string            | no       |              | The password of Elasticsearch                                                                              |
| `elasticsearch.typeName`                    | string            | no       |              | Defines Elasticsearch index type name                                                                      |
| `elasticsearch.batchSizeLimit`              | int               | no       | 1000         | Maximum message count for a batch; if exceeded, a flush is triggered                                       |
| `elasticsearch.batchTickerDuration`         | time.Duration     | no       | 10s          | Batch is flushed automatically at specific time intervals for long-waiting messages in the batch           |
| `elasticsearch.batchCommitTickerDuration`   | time.Duration     | no       | 0s           | Configures checkpoint offset save time; by default, offsets are updated immediately after each batch flush |
| `elasticsearch.batchByteSizeLimit`          | int, string       | no       | 10mb         | Maximum size (bytes) for a batch; if exceeded, a flush is triggered. `10mb` is the default                 |
| `elasticsearch.maxConnsPerHost`             | int               | no       | 512          | Maximum number of connections per host which may be established                                            |
| `elasticsearch.maxIdleConnDuration`         | time.Duration     | no       | 10s          | Idle keep-alive connections are closed after this duration                                                 |
| `elasticsearch.compressionEnabled`          | boolean           | no       | false        | Enables compression for large messages; CPU usage may increase                                             |
| `elasticsearch.concurrentRequest`           | int               | no       | 1            | Concurrent Elasticsearch bulk request count                                                                |
| `elasticsearch.disableDiscoverNodesOnStart` | boolean           | no       | false        | Disables node discovery during client initialization                                                       |
| `elasticsearch.discoverNodesInterval`       | time.Duration     | no       | 5m           | Discovers cluster nodes periodically                                                                       |
| `elasticsearch.rejectionLog.index`          | string            | no       | cbes-rejects | Rejection log index name. `cbes-rejects` is the default                                                    |
| `elasticsearch.rejectionLog.includeSource`  | boolean           | no       | false        | Includes source information in rejection logs. `false` is the default                                      |
| `elasticsearch.maxRetries`                  | int               | no       | math.MaxInt  | Maximum retry count for the Elasticsearch client                                                           |

See [example configurations](example/) for more details.

## Examples

- [Simple Example](example/simple/) - Basic usage with file-based configuration
- [Default Mapper](example/default-mapper/) - Using the built-in default mapper
- [Struct Config](example/struct-config/) - Programmatic configuration

### Exposed metrics

| Metric Name                                                                                  | Description                      | Labels                                                                                                                                                                           | Value Type |
|----------------------------------------------------------------------------------------------|----------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------|
| `go_mongo_cdc_elasticsearch_elasticsearch_connector_latency_ms_current`                      | Time to add events to the batch. | N/A                                                                                                                                                                              | Gauge      |
| `go_mongo_cdc_elasticsearch_elasticsearch_connector_bulk_request_process_latency_ms_current` | Time to process bulk requests.   | N/A                                                                                                                                                                              | Gauge      |
| `go_mongo_cdc_elasticsearch_elasticsearch_connector_action_total_current`                    | Count of Elasticsearch actions   | `action_type`: Type of action (for example `delete`, `index`) `result`: Result of the action (for example `success`, `error`) `index_name`: The name of the index for the action | Counter    |

CDC-related metrics are also exposed by the underlying change data capture layer and are available on the same Prometheus endpoint.

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

See [TESTING.md](TESTING.md) for detailed testing guide.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

Released under the [MIT License](LICENSE).
