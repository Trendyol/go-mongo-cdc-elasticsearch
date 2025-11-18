package integration

import (
	"context"
	"os"
	"testing"
	"time"

	cdcelasticsearch "github.com/Trendyol/go-mongo-cdc-elasticsearch"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch/document"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestIntegration_BasicInsertOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	helper := NewTestHelper(t)

	// Setup connections
	if err := helper.SetupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to setup MongoDB: %v", err)
	}
	defer helper.Close(ctx)

	if err := helper.SetupElasticsearch(); err != nil {
		t.Fatalf("Failed to setup Elasticsearch: %v", err)
	}

	// Cleanup before test
	if err := helper.CleanupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to cleanup MongoDB: %v", err)
	}
	if err := helper.CleanupElasticsearch(); err != nil {
		t.Fatalf("Failed to cleanup Elasticsearch: %v", err)
	}

	// Wait for services to be ready
	if err := WaitForService(t, ElasticsearchURL, DefaultTimeout); err != nil {
		t.Fatalf("Elasticsearch not ready: %v", err)
	}

	// Create and start connector with unique metrics port
	cfg, err := LoadConfigWithUniqueMetricsPort("config/test-config.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	connector, err := cdcelasticsearch.NewConnectorBuilder(cfg).Build()
	if err != nil {
		t.Fatalf("Failed to create connector: %v", err)
	}

	// Start connector in background
	connectorCtx, connectorCancel := context.WithCancel(ctx)
	defer connectorCancel()

	go func() {
		connector.Start(connectorCtx)
	}()

	// Give connector time to initialize
	time.Sleep(5 * time.Second)

	// Insert test document
	testDoc := CreateTestDocument("Test Document 1", 1)
	docID, err := helper.InsertMongoDocument(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	t.Logf("Inserted document with ID: %s", docID)

	// Wait for document to appear in Elasticsearch
	if err := helper.WaitForElasticsearchDocument(docID, 30*time.Second); err != nil {
		t.Fatalf("Document not found in Elasticsearch: %v", err)
	}

	// Verify document in Elasticsearch
	esDoc, err := helper.GetElasticsearchDocument(docID)
	if err != nil {
		t.Fatalf("Failed to get document from Elasticsearch: %v", err)
	}

	t.Logf("Document found in Elasticsearch: %+v", esDoc)

	// Verify document content
	source, ok := esDoc["_source"].(map[string]interface{})
	if !ok {
		t.Fatalf("Invalid document source format")
	}

	if name, ok := source["name"].(string); !ok || name != "Test Document 1" {
		t.Errorf("Expected name 'Test Document 1', got: %v", source["name"])
	}

	// Cleanup
	connectorCancel()
	connector.Close()
	time.Sleep(2 * time.Second)
}

func TestIntegration_MultipleInserts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	helper := NewTestHelper(t)

	if err := helper.SetupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to setup MongoDB: %v", err)
	}
	defer helper.Close(ctx)

	if err := helper.SetupElasticsearch(); err != nil {
		t.Fatalf("Failed to setup Elasticsearch: %v", err)
	}

	// Cleanup before test
	if err := helper.CleanupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to cleanup MongoDB: %v", err)
	}
	if err := helper.CleanupElasticsearch(); err != nil {
		t.Fatalf("Failed to cleanup Elasticsearch: %v", err)
	}

	// Create and start connector with unique metrics port
	cfg, err := LoadConfigWithUniqueMetricsPort("config/test-config.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	connector, err := cdcelasticsearch.NewConnectorBuilder(cfg).Build()
	if err != nil {
		t.Fatalf("Failed to create connector: %v", err)
	}

	connectorCtx, connectorCancel := context.WithCancel(ctx)
	defer connectorCancel()

	go func() {
		connector.Start(connectorCtx)
	}()

	time.Sleep(5 * time.Second)

	numDocs := 10
	for i := 1; i <= numDocs; i++ {
		testDoc := CreateTestDocument("Test Document "+string(rune(i)), i)
		_, err := helper.InsertMongoDocument(ctx, testDoc)
		if err != nil {
			t.Fatalf("Failed to insert document %d: %v", i, err)
		}
	}

	t.Logf("Inserted %d documents", numDocs)

	if err := helper.WaitForElasticsearchDocumentCount(numDocs, 60*time.Second); err != nil {
		t.Fatalf("Not all documents found in Elasticsearch: %v", err)
	}

	count, err := helper.GetElasticsearchDocumentCount()
	if err != nil {
		t.Fatalf("Failed to get document count: %v", err)
	}

	if count < numDocs {
		t.Errorf("Expected at least %d documents, got: %d", numDocs, count)
	}

	t.Logf("Successfully synced %d documents to Elasticsearch", count)

	connectorCancel()
	connector.Close()
	time.Sleep(2 * time.Second)
}

func TestIntegration_UpdateOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	helper := NewTestHelper(t)

	// Setup connections
	if err := helper.SetupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to setup MongoDB: %v", err)
	}
	defer helper.Close(ctx)

	if err := helper.SetupElasticsearch(); err != nil {
		t.Fatalf("Failed to setup Elasticsearch: %v", err)
	}

	// Cleanup before test
	if err := helper.CleanupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to cleanup MongoDB: %v", err)
	}
	if err := helper.CleanupElasticsearch(); err != nil {
		t.Fatalf("Failed to cleanup Elasticsearch: %v", err)
	}

	// Create and start connector with unique metrics port
	cfg, err := LoadConfigWithUniqueMetricsPort("config/test-config.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	connector, err := cdcelasticsearch.NewConnectorBuilder(cfg).Build()
	if err != nil {
		t.Fatalf("Failed to create connector: %v", err)
	}

	// Start connector in background
	connectorCtx, connectorCancel := context.WithCancel(ctx)
	defer connectorCancel()

	go func() {
		connector.Start(connectorCtx)
	}()

	time.Sleep(5 * time.Second)

	testDoc := CreateTestDocument("Original Name", 100)
	docID, err := helper.InsertMongoDocument(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	if err := helper.WaitForElasticsearchDocument(docID, 30*time.Second); err != nil {
		t.Fatalf("Document not found in Elasticsearch: %v", err)
	}

	objID, _ := primitive.ObjectIDFromHex(docID)
	update := bson.M{"$set": bson.M{"name": "Updated Name", "value": 200}}
	if err := helper.UpdateMongoDocument(ctx, bson.M{"_id": objID}, update); err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	if err := helper.WaitForElasticsearchDocumentUpdate(docID, "name", "Updated Name", 30*time.Second); err != nil {
		t.Fatalf("Update not reflected in Elasticsearch: %v", err)
	}

	if err := helper.WaitForElasticsearchDocumentUpdate(docID, "value", 200, 30*time.Second); err != nil {
		t.Fatalf("Update not reflected in Elasticsearch: %v", err)
	}

	t.Logf("Document successfully updated in Elasticsearch")

	connectorCancel()
	connector.Close()
	time.Sleep(2 * time.Second)
}

func TestIntegration_DeleteOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	helper := NewTestHelper(t)

	if err := helper.SetupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to setup MongoDB: %v", err)
	}
	defer helper.Close(ctx)

	if err := helper.SetupElasticsearch(); err != nil {
		t.Fatalf("Failed to setup Elasticsearch: %v", err)
	}

	if err := helper.CleanupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to cleanup MongoDB: %v", err)
	}
	if err := helper.CleanupElasticsearch(); err != nil {
		t.Fatalf("Failed to cleanup Elasticsearch: %v", err)
	}

	cfg, err := LoadConfigWithUniqueMetricsPort("config/test-config.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	connector, err := cdcelasticsearch.NewConnectorBuilder(cfg).Build()
	if err != nil {
		t.Fatalf("Failed to create connector: %v", err)
	}

	connectorCtx, connectorCancel := context.WithCancel(ctx)
	defer connectorCancel()

	go func() {
		connector.Start(connectorCtx)
	}()

	time.Sleep(5 * time.Second)

	testDoc := CreateTestDocument("To Be Deleted", 300)
	docID, err := helper.InsertMongoDocument(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	if err := helper.WaitForElasticsearchDocument(docID, 30*time.Second); err != nil {
		t.Fatalf("Document not found in Elasticsearch: %v", err)
	}

	objID, _ := primitive.ObjectIDFromHex(docID)
	if err := helper.DeleteMongoDocument(ctx, bson.M{"_id": objID}); err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	if err := helper.WaitForElasticsearchDocumentDeletion(docID, 30*time.Second); err != nil {
		t.Fatalf("Deletion not reflected in Elasticsearch: %v", err)
	}

	t.Logf("Document successfully deleted from Elasticsearch")

	connectorCancel()
	connector.Close()
	time.Sleep(2 * time.Second)
}

func TestIntegration_CustomMapper(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	helper := NewTestHelper(t)

	if err := helper.SetupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to setup MongoDB: %v", err)
	}
	defer helper.Close(ctx)

	if err := helper.SetupElasticsearch(); err != nil {
		t.Fatalf("Failed to setup Elasticsearch: %v", err)
	}

	if err := helper.CleanupMongoDB(ctx); err != nil {
		t.Fatalf("Failed to cleanup MongoDB: %v", err)
	}
	if err := helper.CleanupElasticsearch(); err != nil {
		t.Fatalf("Failed to cleanup Elasticsearch: %v", err)
	}

	customMapper := func(event mongodb.Event) []document.ESActionDocument {
		if event.IsMutated {
			e := document.NewIndexAction(event.Key, event.Value, nil)
			return []document.ESActionDocument{e}
		}
		e := document.NewDeleteAction(event.Key, nil)
		return []document.ESActionDocument{e}
	}

	cfg, err := LoadConfigWithUniqueMetricsPort("config/test-config.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	connector, err := cdcelasticsearch.NewConnectorBuilder(cfg).
		SetMapper(customMapper).
		Build()
	if err != nil {
		t.Fatalf("Failed to create connector: %v", err)
	}

	connectorCtx, connectorCancel := context.WithCancel(ctx)
	defer connectorCancel()

	go func() {
		connector.Start(connectorCtx)
	}()

	time.Sleep(5 * time.Second)

	testDoc := CreateTestDocument("Custom Mapper Test", 400)
	docID, err := helper.InsertMongoDocument(ctx, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	if err := helper.WaitForElasticsearchDocument(docID, 30*time.Second); err != nil {
		t.Fatalf("Document not found in Elasticsearch: %v", err)
	}

	esDoc, err := helper.GetElasticsearchDocument(docID)
	if err != nil {
		t.Fatalf("Failed to get document from Elasticsearch: %v", err)
	}

	t.Logf("Custom mapper document found in Elasticsearch: %+v", esDoc)

	connectorCancel()
	connector.Close()
	time.Sleep(2 * time.Second)
}
