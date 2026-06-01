#!/bin/bash

set -e

wait_for_postgres() {
  local timeout_seconds=30
  local start_time=$SECONDS

  echo "Waiting for PostgreSQL to be ready..."
  until pg_isready -h localhost -p 5432 -U postgres >/dev/null 2>&1; do
    if (( SECONDS - start_time >= timeout_seconds )); then
      echo "ERROR: PostgreSQL did not become ready within ${timeout_seconds} seconds." >&2
      exit 1
    fi

    echo "PostgreSQL is not ready yet. Sleeping..."
    sleep 2
  done
  echo "PostgreSQL is ready!"
}

wait_for_redis() {
  local timeout_seconds=30
  local start_time=$SECONDS

  echo "Waiting for Redis to be ready..."
  until redis-cli ping >/dev/null 2>&1; do
    if (( SECONDS - start_time >= timeout_seconds )); then
      echo "ERROR: Redis did not become ready within ${timeout_seconds} seconds." >&2
      exit 1
    fi

    echo "Redis is not ready yet. Sleeping..."
    sleep 1
  done
  echo "Redis is ready!"
}

wait_for_postgres
wait_for_redis

echo "Starting Go application..."
exec /app/main
