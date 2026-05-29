#!/bin/bash

set -e

TIMEOUT=30

echo "Waiting for PostgreSQL to be ready..."
POSTGRES_START_TIME=$(date +%s)
until pg_isready -h localhost -p 5432 -U postgres; do
  CURRENT_TIME=$(date +%s)
  ELAPSED=$((CURRENT_TIME - POSTGRES_START_TIME))
  if [ $ELAPSED -ge $TIMEOUT ]; then
    echo "ERROR: PostgreSQL failed to become ready within $TIMEOUT seconds"
    exit 1
  fi
  echo "PostgreSQL is not ready yet. Sleeping... (elapsed: ${ELAPSED}s)"
  sleep 2
done
echo "PostgreSQL is ready!"

echo "Waiting for Redis to be ready..."
REDIS_START_TIME=$(date +%s)
until redis-cli ping; do
  CURRENT_TIME=$(date +%s)
  ELAPSED=$((CURRENT_TIME - REDIS_START_TIME))
  if [ $ELAPSED -ge $TIMEOUT ]; then
    echo "ERROR: Redis failed to become ready within $TIMEOUT seconds"
    exit 1
  fi
  echo "Redis is not ready yet. Sleeping... (elapsed: ${ELAPSED}s)"
  sleep 1
done
echo "Redis is ready!"

echo "Starting Go application..."
exec /app/main
