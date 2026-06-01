#!/bin/bash

set -e

TIMEOUT=30

echo "Waiting for PostgreSQL to be ready..."
elapsed=0
until pg_isready -h localhost -p 5432 -U postgres; do
  if [ "$elapsed" -ge "$TIMEOUT" ]; then
    echo "ERROR: PostgreSQL did not become ready within ${TIMEOUT} seconds. Exiting." >&2
    exit 1
  fi
  echo "PostgreSQL is not ready yet. Sleeping..."
  sleep 2
  elapsed=$((elapsed + 2))
done
echo "PostgreSQL is ready!"

echo "Waiting for Redis to be ready..."
elapsed=0
until redis-cli ping; do
  if [ "$elapsed" -ge "$TIMEOUT" ]; then
    echo "ERROR: Redis did not become ready within ${TIMEOUT} seconds. Exiting." >&2
    exit 1
  fi
  echo "Redis is not ready yet. Sleeping..."
  sleep 1
  elapsed=$((elapsed + 1))
done
echo "Redis is ready!"

echo "Starting Go application..."
exec /app/main
