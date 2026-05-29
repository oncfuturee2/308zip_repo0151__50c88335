#!/bin/bash

set -e

TIMEOUT=30

echo "Waiting for PostgreSQL to be ready (timeout: ${TIMEOUT}s)..."
START_TIME=$(date +%s)
until pg_isready -h localhost -p 5432 -U postgres; do
  ELAPSED=$(($(date +%s) - START_TIME))
  if [ $ELAPSED -ge $TIMEOUT ]; then
    echo "ERROR: PostgreSQL did not become ready within ${TIMEOUT}s. Exiting." >&2
    exit 1
  fi
  echo "PostgreSQL is not ready yet (elapsed: ${ELAPSED}s). Sleeping..."
  sleep 2
done
echo "PostgreSQL is ready!"

echo "Waiting for Redis to be ready (timeout: ${TIMEOUT}s)..."
START_TIME=$(date +%s)
until redis-cli ping; do
  ELAPSED=$(($(date +%s) - START_TIME))
  if [ $ELAPSED -ge $TIMEOUT ]; then
    echo "ERROR: Redis did not become ready within ${TIMEOUT}s. Exiting." >&2
    exit 1
  fi
  echo "Redis is not ready yet (elapsed: ${ELAPSED}s). Sleeping..."
  sleep 1
done
echo "Redis is ready!"

echo "Starting Go application..."
exec /app/main
