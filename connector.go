package cdcelasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	jsoniter "github.com/json-iterator/go"

	cdc "gitlab.trendyol.com/order/coex/poc/go-mongo-cdc-poc"
	cdcConfig "gitlab.trendyol.com/order/coex/poc/go-mongo-cdc-poc/config"
	"gitlab.trendyol.com/order/coex/poc/go-mongo-cdc-poc/logger"
	"gitlab.trendyol.com/order/coex/poc/go-mongo-cdc-poc/mongo/message"
	"gitlab.trendyol.com/order/coex/poc/go-mongo-cdc-poc/stream"

	esClient "github.com/elastic/go-elasticsearch/v7"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"

	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/config"
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/elasticsearch"
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/elasticsearch/bulk"
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/elasticsearch/client"
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/metric"
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/mongodb"
)

type Connector interface {
	Start(ctx context.Context)
	Close()
}

type connector struct {
	cdc                 cdc.Connector
	mapper              Mapper
	config              *config.Config
	bulk                *bulk.Bulk
	esClient            *esClient.Client
	sinkResponseHandler elasticsearch.SinkResponseHandler
}

func (c *connector) Start(ctx context.Context) {
	go func() {
		c.bulk.StartBulk()
	}()
	c.cdc.Start(ctx)
}

func (c *connector) Close() {
	c.cdc.Close()
	c.bulk.Close()
}

func (c *connector) listener(ctx *stream.ListenerContext) error {
	select {
	case <-ctx.Context.Done():
		return ctx.Context.Err()
	default:
	}

	var e mongodb.Event
	switch ctx.Message.OperationType {
	case message.OperationInsert, message.OperationUpdate, message.OperationReplace:
		if ctx.Message.FullDocument == nil {
			return ctx.Ack()
		}

		doc := ctx.Message.FullDocument
		delete(doc, "_id")

		docBytes, err := json.Marshal(doc)
		if err != nil {
			logger.Log.Error("Failed to marshal document to JSON: %v", err)
			return ctx.Ack()
		}

		var docID string
		switch id := ctx.Message.DocumentID.(type) {
		case primitive.ObjectID:
			docID = id.Hex()
		case int:
			docID = strconv.Itoa(id)
		case int32:
			docID = strconv.FormatInt(int64(id), 10)
		case int64:
			docID = strconv.FormatInt(id, 10)
		case string:
			docID = id
		default:
			docID = fmt.Sprintf("%v", id)
			logger.Log.Warn("Unexpected document ID type: %T, value: %v", id, id)
		}

		e = mongodb.NewMutateEvent(
			c.esClient,
			[]byte(docID),
			docBytes,
			ctx.Message.Collection,
			ctx.Message.EventTime,
			ctx.PartitionID,
		)
	case message.OperationDelete:
		var docID string
		switch id := ctx.Message.DocumentID.(type) {
		case primitive.ObjectID:
			docID = id.Hex()
		case int:
			docID = strconv.Itoa(id)
		case int32:
			docID = strconv.FormatInt(int64(id), 10)
		case int64:
			docID = strconv.FormatInt(id, 10)
		case string:
			docID = id
		default:
			docID = fmt.Sprintf("%v", id)
			logger.Log.Warn("Unexpected document ID type (delete): %T, value: %v", id, id)
		}

		e = mongodb.NewDeleteEvent(
			c.esClient,
			[]byte(docID),
			ctx.Message.Collection,
			ctx.Message.EventTime,
			ctx.PartitionID,
		)
	default:
		return ctx.Ack()
	}

	actions := c.mapper(e)

	if len(actions) == 0 {
		return ctx.Ack()
	}

	ackFunc := func() {
		if err := ctx.Ack(); err != nil {
			logger.Log.Error("Failed to ack event: %v", err)
		}
	}

	c.bulk.AddActions(e.EventTime, actions, e.CollectionName, ackFunc, ctx.PartitionID, ctx.IsBootstrap)

	return nil
}

func newConnectorConfigFromPath(path string) (*config.Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c config.Config
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		return nil, err
	}
	envPattern := regexp.MustCompile(`\${([^}]+)}`)
	matches := envPattern.FindAllStringSubmatch(string(file), -1)
	for _, match := range matches {
		envVar := match[1]
		if value, exists := os.LookupEnv(envVar); exists {
			updatedFile := strings.ReplaceAll(string(file), "${"+envVar+"}", value)
			file = []byte(updatedFile)
		}
	}
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func newConfig(cf any) (*config.Config, error) {
	switch v := cf.(type) {
	case *config.Config:
		return v, nil
	case config.Config:
		return &v, nil
	case string:
		return newConnectorConfigFromPath(v)
	default:
		return nil, errors.New("invalid config")
	}
}

func newConnector(cf any, mapper Mapper, sinkResponseHandler elasticsearch.SinkResponseHandler, metricCollectors ...prometheus.Collector) (Connector, error) {
	cfg, err := newConfig(cf)
	if err != nil {
		return nil, err
	}
	cfg.ApplyDefaults()

	connector := &connector{
		mapper:              mapper,
		config:              cfg,
		sinkResponseHandler: sinkResponseHandler,
	}

	cdcCfg := cdcConfig.Config{
		MongoDB:                 cfg.CDC.MongoDB,
		Metric:                  cfg.CDC.Metric,
		Logger:                  cfg.CDC.Logger,
		Checkpoint:              cfg.CDC.Checkpoint,
		Partition:               cfg.CDC.Partition,
		GracefulShutdownTimeout: cfg.CDC.GracefulShutdownTimeout,
	}

	cdcCfg.Checkpoint.Type = "manual"

	cdc, err := cdc.NewConnector(cdcCfg, connector.listener)
	if err != nil {
		logger.Log.Error("CDC error: %v", err)
		return nil, err
	}

	copyOfConfig := cfg.Elasticsearch
	printConfiguration(copyOfConfig)

	esClient, err := client.NewElasticClient(cfg)
	if err != nil {
		return nil, err
	}
	connector.esClient = esClient

	connector.cdc = cdc

	checkpointCommit := func() {
		cdc.Commit()
	}

	checkpointCommitBootstrap := func(partitionID int) {
		cdc.CommitBootstrap(partitionID)
	}

	connector.bulk, err = bulk.NewBulk(
		cfg,
		checkpointCommit,
		checkpointCommitBootstrap,
		esClient,
		sinkResponseHandler,
	)
	if err != nil {
		return nil, err
	}

	metricCollector := metric.NewMetricCollector(connector.bulk)
	_ = metricCollector

	return connector, nil
}

type ConnectorBuilder struct {
	mapper              Mapper
	config              any
	sinkResponseHandler elasticsearch.SinkResponseHandler
	metricCollectors    []prometheus.Collector
}

func NewConnectorBuilder(config any) *ConnectorBuilder {
	return &ConnectorBuilder{
		config: config,
		mapper: DefaultMapper,
	}
}

func (c *ConnectorBuilder) Build() (Connector, error) {
	return newConnector(c.config, c.mapper, c.sinkResponseHandler, c.metricCollectors...)
}

func (c *ConnectorBuilder) SetMapper(mapper Mapper) *ConnectorBuilder {
	c.mapper = mapper
	return c
}

func (c *ConnectorBuilder) SetMetricCollectors(collectors ...prometheus.Collector) *ConnectorBuilder {
	c.metricCollectors = append(c.metricCollectors, collectors...)
	return c
}

func (c *ConnectorBuilder) SetSinkResponseHandler(sinkResponseHandler elasticsearch.SinkResponseHandler) *ConnectorBuilder {
	c.sinkResponseHandler = sinkResponseHandler
	return c
}

func printConfiguration(config config.Elasticsearch) {
	config.Password = "*****"
	configJSON, _ := jsoniter.Marshal(config)

	dst := &bytes.Buffer{}
	if err := json.Compact(dst, configJSON); err != nil {
		logger.Log.Error("error while print elasticsearch configuration, err: %v", err)
		panic(err)
	}

	logger.Log.Info("using elasticsearch config: %v", dst.String())
}
