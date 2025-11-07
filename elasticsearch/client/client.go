package client

import (
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/config"
	"gitlab.trendyol.com/order/coex/poc/go-mongo-cdc-poc/logger"

	"github.com/elastic/go-elasticsearch/v7"
)

func NewElasticClient(config *config.Config) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Username:              config.Elasticsearch.Username,
		Password:              config.Elasticsearch.Password,
		MaxRetries:            config.Elasticsearch.MaxRetries,
		Addresses:             config.Elasticsearch.Urls,
		Transport:             newTransport(config.Elasticsearch),
		CompressRequestBody:   config.Elasticsearch.CompressionEnabled,
		DiscoverNodesOnStart:  !config.Elasticsearch.DisableDiscoverNodesOnStart,
		DiscoverNodesInterval: *config.Elasticsearch.DiscoverNodesInterval,
		Logger:                &LoggerAdapter{Logger: logger.Log},
	})
	if err != nil {
		return nil, err
	}
	return es, nil
}
