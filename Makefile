.PHONY: default

default: init

init:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@v0.39.0

linter:
	fieldalignment -fix ./...
	golangci-lint run -c .golangci.yml --timeout=5m -v --fix

lint:
	golangci-lint run -c .golangci.yml --timeout=5m -v

test:
	go test ./... -bench . -benchmem

test-unit:
	go test -v -short ./...

test-integration:
	@echo "Starting integration test environment..."
	docker compose -f test/integration/docker-compose.yml up -d
	@echo "Waiting for services to be healthy..."
	@max_attempts=60; \
	attempt=0; \
	while [ $$attempt -lt $$max_attempts ]; do \
		healthy_count=$$(docker compose -f test/integration/docker-compose.yml ps | grep -c "(healthy)" || echo "0"); \
		if [ "$$healthy_count" -ge "5" ]; then \
			echo "All services are healthy!"; \
			break; \
		fi; \
		echo "Healthy services: $$healthy_count/5 - waiting..."; \
		sleep 2; \
		attempt=$$((attempt + 1)); \
	done
	@echo "All services are ready!"
	@echo "Waiting for MongoDB sharded cluster to be fully provisioned..."
	@max_attempts=120; \
	attempt=0; \
	while [ $$attempt -lt $$max_attempts ]; do \
		shard_count=$$(docker exec mongodb-router-test mongosh --quiet --eval 'db.adminCommand({listShards:1}).shards.length' 2>/dev/null || echo 0); \
		sharded=$$(docker exec mongodb-router-test mongosh --quiet --eval 'printjson(db.getSiblingDB("config").collections.findOne({_id:"testdb.testcollection"}) != null)' 2>/dev/null || echo false); \
		if [ "$$shard_count" -ge "2" ] && [ "$$sharded" = "true" ]; then \
			echo "Sharded cluster ready (shards=$$shard_count, testdb.testcollection sharded)."; \
			break; \
		fi; \
		echo "Cluster not ready yet (shards=$$shard_count/2, testdb.testcollection sharded=$$sharded) - waiting..."; \
		sleep 2; \
		attempt=$$((attempt + 1)); \
	done; \
	if [ $$attempt -ge $$max_attempts ]; then \
		echo "ERROR: MongoDB sharded cluster was not provisioned in time."; \
		echo "----- mongodb-setup logs -----"; \
		docker compose -f test/integration/docker-compose.yml logs mongodb-setup || true; \
		exit 1; \
	fi
	@echo "Running integration tests sequentially to avoid metrics collision..."
	@cd test/integration && \
	for test in BasicInsertOperation MultipleInserts UpdateOperation DeleteOperation CustomMapper; do \
		echo ""; \
		echo "=== Running TestIntegration_$$test ==="; \
		go test -v -run "TestIntegration_$$test$$" -timeout 5m || exit 1; \
	done
	@echo ""
	@echo "All integration tests passed!"
	@echo "Stopping integration test environment..."
	docker compose -f test/integration/docker-compose.yml down

test-integration-up:
	@echo "Starting integration test environment..."
	docker compose -f test/integration/docker-compose.yml up -d
	@echo "Waiting for services to be healthy..."
	@max_attempts=60; \
	attempt=0; \
	while [ $$attempt -lt $$max_attempts ]; do \
		healthy_count=$$(docker compose -f test/integration/docker-compose.yml ps | grep -c "(healthy)" || echo "0"); \
		if [ "$$healthy_count" -ge "5" ]; then \
			echo "All services are healthy!"; \
			break; \
		fi; \
		echo "Healthy services: $$healthy_count/5 - waiting..."; \
		sleep 2; \
		attempt=$$((attempt + 1)); \
	done
	@echo "Integration test environment is ready!"

test-integration-down:
	@echo "Stopping integration test environment..."
	docker compose -f test/integration/docker-compose.yml down -v

test-integration-logs:
	docker compose -f test/integration/docker-compose.yml logs -f

test-integration-run:
	@echo "Running integration tests..."
	go test -v ./test/integration/... -timeout 10m

test-integration-clean:
	@echo "Cleaning up integration test environment..."
	docker compose -f test/integration/docker-compose.yml down -v
	rm -rf test/integration/logs/*.log

test-ci-local:
	@echo "🚀 Running all CI checks locally..."
	./scripts/test-ci-locally.sh

tidy:
	go mod tidy

