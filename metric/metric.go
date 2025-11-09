package metric

import (
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const Name = "go_mongo_cdc_elasticsearch"

var (
	actionCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(Name, "elasticsearch_connector_action_total", "current"),
			Help: "Elasticsearch connector action counter",
		},
		[]string{"action_type", "result", "index_name"},
	)

	processLatencyGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(Name, "elasticsearch_connector_latency_ms", "current"),
			Help: "Elasticsearch connector latency ms",
		},
	)

	bulkRequestProcessLatencyGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(Name, "elasticsearch_connector_bulk_request_process_latency_ms", "current"),
			Help: "Elasticsearch connector bulk request process latency ms",
		},
	)
)

type PrometheusMetricsRecorder struct{}

func NewMetricsRecorder() elasticsearch.MetricsRecorder {
	return &PrometheusMetricsRecorder{}
}

func (m *PrometheusMetricsRecorder) RecordIndexSuccess(indexName string, count int64) {
	actionCounter.WithLabelValues("index", "success", indexName).Add(float64(count))
}

func (m *PrometheusMetricsRecorder) RecordIndexError(indexName string, count int64) {
	actionCounter.WithLabelValues("index", "error", indexName).Add(float64(count))
}

func (m *PrometheusMetricsRecorder) RecordDeleteSuccess(indexName string, count int64) {
	actionCounter.WithLabelValues("delete", "success", indexName).Add(float64(count))
}

func (m *PrometheusMetricsRecorder) RecordDeleteError(indexName string, count int64) {
	actionCounter.WithLabelValues("delete", "error", indexName).Add(float64(count))
}

func (m *PrometheusMetricsRecorder) RecordProcessLatency(latencyMs int64) {
	processLatencyGauge.Set(float64(latencyMs))
}

func (m *PrometheusMetricsRecorder) RecordBulkRequestProcessLatency(latencyMs int64) {
	bulkRequestProcessLatencyGauge.Set(float64(latencyMs))
}
