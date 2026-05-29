#!/bin/bash

set -e

echo "Waiting for PostgreSQL to be ready..."
TIMEOUT=30
ELAPSED=0
until pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1; do
  if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
    echo "Error: PostgreSQL failed to start within $TIMEOUT seconds. Exiting."
    exit 1
  fi
  echo "PostgreSQL is not ready yet. Sleeping..."
  sleep 2
  ELAPSED=$((ELAPSED + 2))
done
echo "PostgreSQL is ready!"

echo "Waiting for Redis to be ready..."
ELAPSED=0
until redis-cli ping > /dev/null 2>&1; do
  if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
    echo "Error: Redis failed to start within $TIMEOUT seconds. Exiting."
    exit 1
  fi
  echo "Redis is not ready yet. Sleeping..."
  sleep 1
  ELAPSED=$((ELAPSED + 1))
done
echo "Redis is ready!"

echo "Starting Go application..."
exec /app/main
