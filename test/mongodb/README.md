# MongoDB Test Infrastructure

This directory contains MongoDB-specific test infrastructure including Docker configuration and setup scripts.

## Contents

### Dockerfile
Custom MongoDB Docker image with configuration capabilities for different roles:
- **Config Server**: MongoDB config server for sharded cluster
- **Shard Server**: MongoDB shard server
- **Router (Mongos)**: MongoDB router for sharded cluster
- **Standalone**: Single MongoDB instance with replica set

### configure.sh
Automated configuration script that:
- Starts MongoDB in the appropriate role
- Initializes replica sets
- Waits for MongoDB to be ready
- Configures sharding (if applicable)

## Usage

### Building the Image

```bash
cd test/mongodb
docker build -t mongo-cdc-test .
```

### Running Different Roles

#### Config Server
```bash
docker run -e MONGO_ROLE=config -p 27019:27019 mongo-cdc-test
```

#### Shard Server
```bash
docker run -e MONGO_ROLE=shard1 -p 27018:27018 mongo-cdc-test
```

#### Router (Mongos)
```bash
docker run -e MONGO_ROLE=router -p 27017:27017 mongo-cdc-test
```

#### Standalone
```bash
docker run -e MONGO_ROLE=standalone -p 27017:27017 mongo-cdc-test
```

## Environment Variables

- `MONGO_ROLE`: Determines the MongoDB role
  - `config`: Config server for sharded cluster
  - `shard1`: First shard server
  - `shard2`: Second shard server
  - `router`: Mongos router
  - `standalone`: Single instance (default)

## Ports

- **27017**: Mongos router / Standalone
- **27018**: Shard servers
- **27019**: Config server

## Integration with docker-compose

This image is used in the integration test docker-compose setup:

```yaml
mongodb-config:
  build:
    context: ../mongodb
    dockerfile: Dockerfile
  environment:
    - MONGO_ROLE=config
```

## Features

- ✅ Automatic replica set initialization
- ✅ Health check support
- ✅ Configurable roles via environment variables
- ✅ Proper signal handling
- ✅ Logging with timestamps
- ✅ Error propagation

## Testing

To test the MongoDB setup:

```bash
# Start the container
docker run -d --name mongo-test -e MONGO_ROLE=standalone mongo-cdc-test

# Check logs
docker logs -f mongo-test

# Connect to MongoDB
docker exec -it mongo-test mongosh

# Stop and remove
docker stop mongo-test && docker rm mongo-test
```

## Troubleshooting

### Container won't start
- Check logs: `docker logs <container-name>`
- Verify port availability
- Ensure sufficient disk space

### Replica set initialization fails
- Wait longer for MongoDB to start
- Check network connectivity between containers
- Verify hostnames are resolvable

### Connection refused
- Ensure MongoDB is fully initialized
- Check if the correct port is exposed
- Verify network configuration

## Future Tests

Planned MongoDB-specific tests:
- `event_test.go`: Test Event struct and factory functions
- `listener_test.go`: Test change stream message processing
- `document_id_test.go`: Test various document ID type conversions
- `replica_set_test.go`: Test replica set operations
- `sharding_test.go`: Test sharded cluster operations

## Contributing

When modifying the MongoDB test infrastructure:
1. Test all roles (config, shard, router, standalone)
2. Verify health checks work correctly
3. Ensure proper cleanup on shutdown
4. Update documentation for any new features
5. Test with different MongoDB versions
