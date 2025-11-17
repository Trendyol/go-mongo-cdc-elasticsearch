#!/bin/bash
set -e

echo "Waiting for MongoDB services to start..."
sleep 15

# Initialize Config Server Replica Set
echo "Initializing Config Server Replica Set..."
mongosh --host mongodb-config:27019 <<EOF
rs.initiate(
  {
    _id: "configReplSet",
    configsvr: true,
    members: [
      { _id: 0, host: "mongodb-config:27019" }
    ]
  }
)
EOF

echo "Waiting for Config Server replication to initialize..."
sleep 15

# Initialize Shard 1 Replica Set
echo "Initializing Shard 1 Replica Set..."
mongosh --host mongodb-shard1:27018 <<EOF
rs.initiate(
  {
    _id: "shard1ReplSet",
    members: [
      { _id: 0, host: "mongodb-shard1:27018" }
    ]
  }
)
EOF

# Initialize Shard 2 Replica Set
echo "Initializing Shard 2 Replica Set..."
mongosh --host mongodb-shard2:27018 <<EOF
rs.initiate(
  {
    _id: "shard2ReplSet",
    members: [
      { _id: 0, host: "mongodb-shard2:27018" }
    ]
  }
)
EOF

echo "Waiting for Shard replication to initialize..."
sleep 30

# Make sure mongos is ready
echo "Checking if mongos is ready..."
until mongosh --host mongodb-router:27017 --eval "db.adminCommand('ping')" >/dev/null 2>&1; do
  echo "Waiting for mongos to be ready..."
  sleep 5
done

# Add shards to the cluster
echo "Adding shards to the cluster..."
mongosh --host mongodb-router:27017 <<EOF
sh.addShard("shard1ReplSet/mongodb-shard1:27018")
sh.addShard("shard2ReplSet/mongodb-shard2:27018")
EOF

# Enable sharding on test database
echo "Enabling sharding on test database..."
mongosh --host mongodb-router:27017 <<EOF
sh.enableSharding("testdb")
sh.shardCollection("testdb.testcollection", { "_id": "hashed" })
EOF

echo "MongoDB Sharded Cluster setup completed!"

