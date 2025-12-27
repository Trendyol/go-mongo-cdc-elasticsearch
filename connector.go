package cdcelasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"

	"go.mongodb.org/mongo-driver/bson/primitive"

	jsoniter "github.com/json-iterator/go"

	cdc "github.com/Trendyol/go-mongo-cdc"
	cdcConfig "github.com/Trendyol/go-mongo-cdc/config"
	"github.com/Trendyol/go-mongo-cdc/logger"
	"github.com/Trendyol/go-mongo-cdc/mongo/message"
	"github.com/Trendyol/go-mongo-cdc/stream"
	"go.uber.org/zap"

	"github.com/elastic/go-elasticsearch/v7"
	"gopkg.in/yaml.v3"

	"github.com/Trendyol/go-mongo-cdc-elasticsearch/config"
	cdcElasticsearch "github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch/bulk"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch/client"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch/document"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/helper"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/metric"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/mongodb"
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
	esClient            *elasticsearch.Client
	sinkResponseHandler cdcElasticsearch.SinkResponseHandler
	closing             int32
}

func (c *connector) Start(ctx context.Context) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT, syscall.SIGQUIT)
	defer stop()

	go func() {
		c.bulk.StartBulk()
	}()

	cdcCtx, cdcCancel := context.WithCancel(context.Background())
	defer cdcCancel()

	go func() {
		select {
		case <-ctx.Done():
			logger.Log.Debug("Shutdown signal received: interrupt, initiating graceful shutdown...")
			atomic.StoreInt32(&c.closing, 1)

			logger.Log.Debug("Stopped accepting new events, waiting for in-flight events to complete...")

			c.bulk.Close()

			logger.Log.Debug("Bulk closed, committing final checkpoints before closing CDC")
			cdcCancel()
		case <-cdcCtx.Done():
		}
	}()

	c.cdc.Start(cdcCtx)
}

func (c *connector) Close() {
	atomic.StoreInt32(&c.closing, 1)
	c.bulk.Close()
	c.cdc.Close()
}

func (c *connector) listener(ctx *stream.ListenerContext) error {
	if atomic.LoadInt32(&c.closing) == 1 {
		logger.Log.Debug("Rejecting new event during shutdown - documentId: %v, partitionId: %d", ctx.Message.DocumentID, ctx.PartitionID)
		return nil
	}

	select {
	case <-ctx.Context.Done():
		return ctx.Context.Err()
	default:
	}

	var e mongodb.Event
	switch ctx.Message.OperationType {
	case message.OperationInsert, message.OperationUpdate, message.OperationReplace:
		if ctx.Message.FullDocument == nil {
			ctx.Ack()
			return nil
		}

		doc := ctx.Message.FullDocument
		delete(doc, "_id")

		docBytes, err := jsoniter.Marshal(doc)
		if err != nil {
			logger.Log.Error("Failed to marshal document to JSON: %v", err)
			ctx.Ack()
			return err
		}

		docIDBytes := documentIDToBytes(ctx.Message.DocumentID)

		e = mongodb.NewMutateEvent(
			c.esClient,
			docIDBytes,
			docBytes,
			ctx.Message.Collection,
			ctx.Message.EventTime,
			ctx.PartitionID,
		)
	case message.OperationDelete:
		docIDBytes := documentIDToBytes(ctx.Message.DocumentID)

		e = mongodb.NewDeleteEvent(
			c.esClient,
			docIDBytes,
			ctx.Message.Collection,
			ctx.Message.EventTime,
			ctx.PartitionID,
		)
	default:
		ctx.Ack()
		return nil
	}

	actions := c.mapper(e)

	if len(actions) == 0 {
		ctx.Ack()
		return nil
	}

	batchSizeLimit := c.config.Elasticsearch.BatchSizeLimit
	if len(actions) > batchSizeLimit {
		chunks := helper.ChunkSliceWithSize[document.ESActionDocument](actions, batchSizeLimit)
		lastChunkIndex := len(chunks) - 1
		for idx, chunk := range chunks {
			c.bulk.AddActions(ctx, e.EventTime, chunk, e.CollectionName, idx == lastChunkIndex, ctx.PartitionID, ctx.IsBootstrap)
		}
	} else {
		c.bulk.AddActions(ctx, e.EventTime, actions, e.CollectionName, true, ctx.PartitionID, ctx.IsBootstrap)
	}

	return nil
}

func documentIDToBytes(id interface{}) []byte {
	var docID string
	switch id := id.(type) {
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

	return helper.Byte(docID)
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

func newConnector(cf any, mapper Mapper, sinkResponseHandler cdcElasticsearch.SinkResponseHandler) (Connector, error) {
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
		return nil, err
	}

	connector.cdc = cdc

	copyOfConfig := cfg.Elasticsearch
	printConfiguration(copyOfConfig)

	esClient, err := client.NewElasticClient(cfg)
	if err != nil {
		return nil, err
	}
	connector.esClient = esClient

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
		metric.NewMetricsRecorder(),
	)
	if err != nil {
		return nil, err
	}

	eventHandler := &CdcEventHandler{
		bulk: connector.bulk,
	}
	cdc.SetEventHandler(eventHandler)

	return connector, nil
}

type ConnectorBuilder struct {
	mapper              Mapper
	config              any
	sinkResponseHandler cdcElasticsearch.SinkResponseHandler
}

func NewConnectorBuilder(config any) *ConnectorBuilder {
	return &ConnectorBuilder{
		config: config,
		mapper: DefaultMapper,
	}
}

func (c *ConnectorBuilder) Build() (Connector, error) {
	return newConnector(c.config, c.mapper, c.sinkResponseHandler)
}

func (c *ConnectorBuilder) SetMapper(mapper Mapper) *ConnectorBuilder {
	c.mapper = mapper
	return c
}

func (c *ConnectorBuilder) SetLogger(zapLogger *zap.Logger) *ConnectorBuilder {
	logger.Log = &logger.Loggers{
		Zap: zapLogger,
	}
	return c
}

func (c *ConnectorBuilder) SetSinkResponseHandler(sinkResponseHandler cdcElasticsearch.SinkResponseHandler) *ConnectorBuilder {
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
