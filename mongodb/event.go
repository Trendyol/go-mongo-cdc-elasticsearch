package mongodb

import (
	"time"

	"github.com/elastic/go-elasticsearch/v7"
)

type Event struct {
	ElasticsearchClient *elasticsearch.Client
	EventTime           time.Time
	CollectionName      string
	Key                 []byte
	Value               []byte
	IsDeleted           bool
	IsMutated           bool
	PartitionID         int
}

func NewDeleteEvent(
	esClient *elasticsearch.Client,
	key []byte,
	collectionName string,
	eventTime time.Time,
	partitionID int,
) Event {
	return Event{
		ElasticsearchClient: esClient,
		Key:                 key,
		IsDeleted:           true,
		CollectionName:      collectionName,
		EventTime:           eventTime,
		PartitionID:         partitionID,
	}
}

func NewMutateEvent(
	esClient *elasticsearch.Client,
	key []byte,
	value []byte,
	collectionName string,
	eventTime time.Time,
	partitionID int,
) Event {
	return Event{
		ElasticsearchClient: esClient,
		Key:                 key,
		Value:               value,
		IsMutated:           true,
		CollectionName:      collectionName,
		EventTime:           eventTime,
		PartitionID:         partitionID,
	}
}
