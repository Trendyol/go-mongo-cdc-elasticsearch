package elasticsearch

type MetricsRecorder interface {
	RecordIndexSuccess(indexName string, count int64)
	RecordIndexError(indexName string, count int64)
	RecordDeleteSuccess(indexName string, count int64)
	RecordDeleteError(indexName string, count int64)
	RecordProcessLatency(latencyMs int64)
	RecordBulkRequestProcessLatency(latencyMs int64)
}
