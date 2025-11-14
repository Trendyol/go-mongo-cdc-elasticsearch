package main

import (
	"context"
	"time"

	cdcelasticsearch "github.com/Trendyol/go-mongo-cdc-elasticsearch"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/config"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch/document"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/mongodb"
	cdcConfig "github.com/Trendyol/go-mongo-cdc/config"
)

func mapper(event mongodb.Event) []document.ESActionDocument {
	if event.IsMutated {
		e := document.NewIndexAction(event.Key, event.Value, nil)
		return []document.ESActionDocument{e}
	}
	e := document.NewDeleteAction(event.Key, nil)
	return []document.ESActionDocument{e}
}

func main() {
	cfg := config.Config{
		Elasticsearch: config.Elasticsearch{
			CollectionIndexMapping: map[string]string{
				"exampleCollection": "example-collection",
			},
			Urls:                []string{"http://localhost:9200"},
			BatchSizeLimit:      3000,
			BatchTickerDuration: 10 * time.Second,
			BatchByteSizeLimit:  "10mb",
			ConcurrentRequest:   1,
		},
		CDC: cdcConfig.Config{
			MongoDB: cdcConfig.MongoDB{
				Connection: cdcConfig.Connection{
					URI:        "localhost:27017",
					Database:   "exampleDB",
					Collection: "exampleCollection",
				},
				ConnectionPool: cdcConfig.ConnectionPool{
					MaxPoolSize:   100,
					MinPoolSize:   5,
					MaxIdleTimeMS: 300000,
				},
				Timeouts: cdcConfig.Timeouts{
					ConnectTimeoutMS:         30000,
					ServerSelectionTimeoutMS: 60000,
					SocketTimeoutMS:          120000,
				},
			},
			Metric: cdcConfig.MetricConfig{
				Port: 8080,
			},
			Checkpoint: cdcConfig.CheckpointConfig{
				BootstrapSaveCount:      3000,
				BootstrapQueryBatchSize: 3000,
				ChangeStreamSaveCount:   3000,
			},
			Partition: cdcConfig.PartitionConfig{
				HeartbeatInterval:      10 * time.Second,
				WorkerTimeout:          90 * time.Second,
				RebalanceCheckInterval: 10 * time.Second,
				TotalPartition:         5,
			},
			Logger: cdcConfig.LoggerConfig{LogLevel: "debug"},
		},
	}

	connector, err := cdcelasticsearch.NewConnectorBuilder(cfg).
		SetMapper(mapper).
		Build()
	if err != nil {
		panic(err)
	}

	defer connector.Close()
	connector.Start(context.Background())
}
