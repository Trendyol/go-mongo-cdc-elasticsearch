package config

import (
	"math"
	"time"

	cdcConfig "github.com/Trendyol/go-mongo-cdc/config"
)

type Elasticsearch struct {
	BatchByteSizeLimit          any               `yaml:"batchByteSizeLimit"`
	BatchCommitTickerDuration   *time.Duration    `yaml:"batchCommitTickerDuration"`
	MaxConnsPerHost             *int              `yaml:"maxConnsPerHost"`
	MaxIdleConnDuration         *time.Duration    `yaml:"maxIdleConnDuration"`
	DiscoverNodesInterval       *time.Duration    `yaml:"discoverNodesInterval"`
	CollectionIndexMapping      map[string]string `yaml:"collectionIndexMapping"`
	TypeName                    string            `yaml:"typeName"`
	Password                    string            `yaml:"password"`
	Username                    string            `yaml:"username"`
	RejectionLog                RejectionLog      `yaml:"rejectionLog"`
	Urls                        []string          `yaml:"urls"`
	BatchSizeLimit              int               `yaml:"batchSizeLimit"`
	BatchTickerDuration         time.Duration     `yaml:"batchTickerDuration"`
	ConcurrentRequest           int               `yaml:"concurrentRequest"`
	CompressionEnabled          bool              `yaml:"compressionEnabled"`
	DisableDiscoverNodesOnStart bool              `yaml:"disableDiscoverNodesOnStart"`
	MaxRetries                  int               `yaml:"maxRetries"`
}

type RejectionLog struct {
	Index         string `yaml:"index"`
	IncludeSource bool   `yaml:"includeSource"`
}

type Config struct {
	Elasticsearch Elasticsearch    `yaml:"elasticsearch"`
	CDC           cdcConfig.Config `yaml:",inline" mapstructure:",squash"`
}

func (c *Config) ApplyDefaults() {
	if c.Elasticsearch.BatchTickerDuration == 0 {
		c.Elasticsearch.BatchTickerDuration = 10 * time.Second
	}

	if c.Elasticsearch.BatchSizeLimit == 0 {
		c.Elasticsearch.BatchSizeLimit = 1000
	}

	if c.Elasticsearch.BatchByteSizeLimit == nil {
		c.Elasticsearch.BatchByteSizeLimit = "10mb"
	}

	if c.Elasticsearch.ConcurrentRequest == 0 {
		c.Elasticsearch.ConcurrentRequest = 1
	}

	if c.Elasticsearch.DiscoverNodesInterval == nil {
		duration := 5 * time.Minute
		c.Elasticsearch.DiscoverNodesInterval = &duration
	}

	if c.Elasticsearch.MaxRetries == 0 {
		c.Elasticsearch.MaxRetries = math.MaxInt
	}

	if c.CDC.MongoDB.ConnectionPool.MaxPoolSize == 0 {
		c.CDC.MongoDB.ConnectionPool.MaxPoolSize = 100
	}

	if c.CDC.MongoDB.ConnectionPool.MinPoolSize == 0 {
		c.CDC.MongoDB.ConnectionPool.MinPoolSize = 5
	}

	if c.CDC.MongoDB.ConnectionPool.MaxIdleTimeMS == 0 {
		c.CDC.MongoDB.ConnectionPool.MaxIdleTimeMS = 300000
	}

	if c.CDC.MongoDB.Timeouts.ConnectTimeoutMS == 0 {
		c.CDC.MongoDB.Timeouts.ConnectTimeoutMS = 30000
	}

	if c.CDC.MongoDB.Timeouts.ServerSelectionTimeoutMS == 0 {
		c.CDC.MongoDB.Timeouts.ServerSelectionTimeoutMS = 60000
	}

	if c.CDC.MongoDB.Timeouts.SocketTimeoutMS == 0 {
		c.CDC.MongoDB.Timeouts.SocketTimeoutMS = 120000
	}

	if c.CDC.Metric.Port == 0 {
		c.CDC.Metric.Port = 8080
	}

	if c.CDC.Metric.CollectionInterval == 0 {
		c.CDC.Metric.CollectionInterval = 30 * time.Second
	}

	if c.CDC.Checkpoint.TokenSaveInterval == 0 {
		c.CDC.Checkpoint.TokenSaveInterval = 10 * time.Second
	}

	if c.CDC.Checkpoint.BootstrapSaveCount == 0 {
		c.CDC.Checkpoint.BootstrapSaveCount = 1000
	}

	if c.CDC.Checkpoint.TokenSaveTimeout == 0 {
		c.CDC.Checkpoint.TokenSaveTimeout = 10 * time.Second
	}

	if c.CDC.Checkpoint.BootstrapSaveInterval == 0 {
		c.CDC.Checkpoint.BootstrapSaveInterval = 5 * time.Second
	}

	if c.CDC.Checkpoint.BootstrapBatchSize == 0 {
		c.CDC.Checkpoint.BootstrapBatchSize = 500
	}

	if c.CDC.Checkpoint.IdleHeartbeatInterval == 0 {
		c.CDC.Checkpoint.IdleHeartbeatInterval = 3 * time.Minute
	}

	if c.CDC.Checkpoint.MaxIdleTime == 0 {
		c.CDC.Checkpoint.MaxIdleTime = 15 * time.Minute
	}

	if c.CDC.Checkpoint.ChangeStreamBatchSize == 0 {
		c.CDC.Checkpoint.ChangeStreamBatchSize = 100
	}

	if c.CDC.Partition.HeartbeatInterval == 0 {
		c.CDC.Partition.HeartbeatInterval = 10 * time.Second
	}

	if c.CDC.Partition.WorkerTimeout == 0 {
		c.CDC.Partition.WorkerTimeout = 90 * time.Second
	}

	if c.CDC.Partition.WorkersCollection == "" {
		c.CDC.Partition.WorkersCollection = "workers"
	}

	if c.CDC.Partition.PartitionsCollection == "" {
		c.CDC.Partition.PartitionsCollection = "partition_assignments"
	}

	if c.CDC.Partition.RebalanceCheckInterval == 0 {
		c.CDC.Partition.RebalanceCheckInterval = 15 * time.Second
	}

	if c.CDC.Partition.TotalPartition == 0 {
		c.CDC.Partition.TotalPartition = 15
	}

	if c.CDC.GracefulShutdownTimeout == 0 {
		c.CDC.GracefulShutdownTimeout = 5 * time.Second
	}
}
