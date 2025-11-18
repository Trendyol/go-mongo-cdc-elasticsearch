package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Trendyol/go-mongo-cdc-elasticsearch/config"
	"github.com/elastic/go-elasticsearch/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
)

const (
	MongoDBURI       = "mongodb://localhost:27017"
	ElasticsearchURL = "http://localhost:9200"
	TestDatabase     = "testdb"
	TestCollection   = "testcollection"
	TestIndex        = "test-index"
	DefaultTimeout   = 30 * time.Second
	RetryInterval    = 1 * time.Second
	MaxRetries       = 30
	BaseMetricsPort  = 8080
)

// Global counter for unique metrics ports
var metricsPortCounter int32 = 0

// TestHelper provides utility functions for integration tests
type TestHelper struct {
	t              *testing.T
	mongoClient    *mongo.Client
	esClient       *elasticsearch.Client
	testCollection *mongo.Collection
}

// NewTestHelper creates a new test helper instance
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// SetupMongoDB connects to MongoDB and returns the test collection
func (h *TestHelper) SetupMongoDB(ctx context.Context) error {
	clientOptions := options.Client().ApplyURI(MongoDBURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	h.mongoClient = client
	h.testCollection = client.Database(TestDatabase).Collection(TestCollection)
	h.t.Logf("Connected to MongoDB successfully")
	return nil
}

func (h *TestHelper) SetupElasticsearch() error {
	cfg := elasticsearch.Config{
		Addresses: []string{ElasticsearchURL},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned error: %s", res.Status())
	}

	h.esClient = client
	h.t.Logf("Connected to Elasticsearch successfully")
	return nil
}

func (h *TestHelper) CleanupMongoDB(ctx context.Context) error {
	if h.testCollection == nil {
		return nil
	}

	_, err := h.testCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to cleanup MongoDB: %w", err)
	}

	h.t.Logf("Cleaned up MongoDB collection")
	return nil
}

func (h *TestHelper) CleanupElasticsearch() error {
	if h.esClient == nil {
		return nil
	}

	res, err := h.esClient.Indices.Delete([]string{TestIndex})
	if err != nil {
		return fmt.Errorf("failed to delete Elasticsearch index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch delete index error: %s, body: %s", res.Status(), string(body))
	}

	h.t.Logf("Cleaned up Elasticsearch index")
	return nil
}

func (h *TestHelper) Close(ctx context.Context) {
	if h.mongoClient != nil {
		if err := h.mongoClient.Disconnect(ctx); err != nil {
			h.t.Logf("Error disconnecting MongoDB: %v", err)
		}
	}
}

func (h *TestHelper) InsertMongoDocument(ctx context.Context, doc interface{}) (string, error) {
	result, err := h.testCollection.InsertOne(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	var docID string
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		docID = oid.Hex()
	} else {
		docID = fmt.Sprintf("%v", result.InsertedID)
	}

	h.t.Logf("Inserted document with ID: %s", docID)
	return docID, nil
}

func (h *TestHelper) UpdateMongoDocument(ctx context.Context, filter, update interface{}) error {
	result, err := h.testCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	h.t.Logf("Updated %d document(s)", result.ModifiedCount)
	return nil
}

func (h *TestHelper) DeleteMongoDocument(ctx context.Context, filter interface{}) error {
	result, err := h.testCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	h.t.Logf("Deleted %d document(s)", result.DeletedCount)
	return nil
}

func (h *TestHelper) GetElasticsearchDocument(docID string) (map[string]interface{}, error) {
	res, err := h.esClient.Get(TestIndex, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document from Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("elasticsearch get error: %s, body: %s", res.Status(), string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Elasticsearch response: %w", err)
	}

	return result, nil
}

func (h *TestHelper) SearchElasticsearch(query map[string]interface{}) ([]map[string]interface{}, error) {
	var buf strings.Builder
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := h.esClient.Search(
		h.esClient.Search.WithIndex(TestIndex),
		h.esClient.Search.WithBody(strings.NewReader(buf.String())),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("elasticsearch search error: %s, body: %s", res.Status(), string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid search response format")
	}

	hitsArray, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid hits format")
	}

	documents := make([]map[string]interface{}, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		documents = append(documents, hitMap)
	}

	return documents, nil
}

func (h *TestHelper) WaitForElasticsearchDocument(docID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for document %s in Elasticsearch", docID)
		case <-ticker.C:
			_, err := h.GetElasticsearchDocument(docID)
			if err == nil {
				h.t.Logf("Document %s found in Elasticsearch", docID)
				return nil
			}
		}
	}
}

func (h *TestHelper) WaitForElasticsearchDocumentCount(expectedCount int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %d documents in Elasticsearch", expectedCount)
		case <-ticker.C:
			count, err := h.GetElasticsearchDocumentCount()
			if err != nil {
				continue
			}
			if count >= expectedCount {
				h.t.Logf("Found %d documents in Elasticsearch (expected: %d)", count, expectedCount)
				return nil
			}
			h.t.Logf("Current document count: %d, waiting for: %d", count, expectedCount)
		}
	}
}

func (h *TestHelper) GetElasticsearchDocumentCount() (int, error) {
	res, err := h.esClient.Count(
		h.esClient.Count.WithIndex(TestIndex),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return 0, fmt.Errorf("elasticsearch count error: %s, body: %s", res.Status(), string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode count response: %w", err)
	}

	count, ok := result["count"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid count response format")
	}

	return int(count), nil
}

func (h *TestHelper) RefreshElasticsearchIndex() error {
	res, err := h.esClient.Indices.Refresh(
		h.esClient.Indices.Refresh.WithIndex(TestIndex),
	)
	if err != nil {
		return fmt.Errorf("failed to refresh index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch refresh error: %s, body: %s", res.Status(), string(body))
	}

	return nil
}

func WaitForService(t *testing.T, url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service at %s", url)
		case <-ticker.C:
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 500 {
					t.Logf("Service at %s is available", url)
					return nil
				}
			}
		}
	}
}

func CreateTestDocument(name string, value int) map[string]interface{} {
	return map[string]interface{}{
		"name":     name,
		"value":    value,
		"category": fmt.Sprintf("category%d", value%5),
		"tags":     []string{fmt.Sprintf("tag%d", value%3), fmt.Sprintf("tag%d", value%7)},
		"metadata": map[string]interface{}{
			"created": time.Now(),
			"version": 1,
			"active":  value%2 == 0,
		},
		"timestamp": time.Now(),
	}
}

func LoadConfigWithUniqueMetricsPort(configPath string) (*config.Config, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	port := int(atomic.AddInt32(&metricsPortCounter, 1))
	cfg.CDC.Metric.Port = BaseMetricsPort + port

	return &cfg, nil
}

func (h *TestHelper) WaitForElasticsearchDocumentUpdate(docID string, expectedField string, expectedValue interface{}, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for document %s to be updated with %s=%v", docID, expectedField, expectedValue)
		case <-ticker.C:
			doc, err := h.GetElasticsearchDocument(docID)
			if err != nil {
				continue
			}

			source, ok := doc["_source"].(map[string]interface{})
			if !ok {
				continue
			}

			actualValue, exists := source[expectedField]
			if !exists {
				continue
			}

			switch expected := expectedValue.(type) {
			case string:
				if actual, ok := actualValue.(string); ok && actual == expected {
					h.t.Logf("Document %s updated with %s=%v", docID, expectedField, expectedValue)
					return nil
				}
			case int:
				if actual, ok := actualValue.(float64); ok && int(actual) == expected {
					h.t.Logf("Document %s updated with %s=%v", docID, expectedField, expectedValue)
					return nil
				}
			case float64:
				if actual, ok := actualValue.(float64); ok && actual == expected {
					h.t.Logf("Document %s updated with %s=%v", docID, expectedField, expectedValue)
					return nil
				}
			}
		}
	}
}

func (h *TestHelper) WaitForElasticsearchDocumentDeletion(docID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for document %s to be deleted from Elasticsearch", docID)
		case <-ticker.C:
			_, err := h.GetElasticsearchDocument(docID)
			if err != nil {
				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
					h.t.Logf("Document %s successfully deleted from Elasticsearch", docID)
					return nil
				}
			}
		}
	}
}
