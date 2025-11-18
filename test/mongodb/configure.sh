#!/bin/bash

# Enables job control
set -m

# Enables error propagation
set -e

echo "Starting MongoDB configuration..."

# Determine the MongoDB role based on environment variable
MONGO_ROLE=${MONGO_ROLE:-standalone}

case "$MONGO_ROLE" in
  "config")
    echo "Starting as Config Server..."
    mongod --configsvr --replSet configReplSet --port 27019 --dbpath /data/db --bind_ip_all &
    ;;
  "shard1")
    echo "Starting as Shard 1..."
    mongod --shardsvr --replSet shard1ReplSet --port 27018 --dbpath /data/db --bind_ip_all &
    ;;
  "shard2")
    echo "Starting as Shard 2..."
    mongod --shardsvr --replSet shard2ReplSet --port 27018 --dbpath /data/db --bind_ip_all &
    ;;
  "router")
    echo "Starting as Mongos Router..."
    # Wait for config server to be ready
    sleep 15
    mongos --configdb configReplSet/mongodb-config:27019 --port 27017 --bind_ip_all &
    ;;
  *)
    echo "Starting as Standalone..."
    mongod --replSet rs0 --port 27017 --dbpath /data/db --bind_ip_all &
    ;;
esac

# Variable used in echo
i=1
# Echo with counter
log() {
  echo "[$i] [$(date +"%T")] $@"
  i=`expr $i + 1`
}

# Check if MongoDB is up
check_mongo() {
  if [ "$MONGO_ROLE" = "router" ]; then
    mongosh --host localhost:27017 --eval "db.adminCommand('ping')" > /dev/null 2>&1
  elif [ "$MONGO_ROLE" = "config" ]; then
    mongosh --host localhost:27019 --eval "db.adminCommand('ping')" > /dev/null 2>&1
  else
    mongosh --host localhost:27018 --eval "db.adminCommand('ping')" > /dev/null 2>&1
  fi
  echo $?
}

# Wait until MongoDB is ready
until [[ $(check_mongo) = 0 ]]; do
  log "Waiting for MongoDB to be available..."
  sleep 2
done

log "MongoDB is ready!"

# Initialize replica set based on role
case "$MONGO_ROLE" in
  "config")
    log "Initializing Config Server Replica Set..."
    mongosh --host localhost:27019 <<EOF
rs.initiate({
  _id: "configReplSet",
  configsvr: true,
  members: [{ _id: 0, host: "mongodb-config:27019" }]
})
EOF
    ;;
  "shard1")
    log "Initializing Shard 1 Replica Set..."
    mongosh --host localhost:27018 <<EOF
rs.initiate({
  _id: "shard1ReplSet",
  members: [{ _id: 0, host: "mongodb-shard1:27018" }]
})
EOF
    ;;
  "shard2")
    log "Initializing Shard 2 Replica Set..."
    mongosh --host localhost:27018 <<EOF
rs.initiate({
  _id: "shard2ReplSet",
  members: [{ _id: 0, host: "mongodb-shard2:27018" }]
})
EOF
    ;;
  "standalone")
    log "Initializing Standalone Replica Set..."
    mongosh --host localhost:27017 <<EOF
rs.initiate({
  _id: "rs0",
  members: [{ _id: 0, host: "localhost:27017" }]
})
EOF
    ;;
esac

log "MongoDB $MONGO_ROLE configuration completed!"

# Bring the MongoDB process to foreground
fg 1

