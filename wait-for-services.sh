#!/bin/bash

set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h localhost -p 5432 -U postgres; do
  echo "PostgreSQL is not ready yet. Sleeping..."
  sleep 2
done
echo "PostgreSQL is ready!"

echo "Waiting for Redis to be ready..."
until redis-cli ping; do
  echo "Redis is not ready yet. Sleeping..."
  sleep 1
done
echo "Redis is ready!"

echo "Starting Go application..."
exec /app/main
