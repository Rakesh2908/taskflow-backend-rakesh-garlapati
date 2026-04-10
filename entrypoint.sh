#!/bin/sh
set -e

if [ -z "$POSTGRES_DB" ]; then POSTGRES_DB="taskflow"; fi
if [ -z "$POSTGRES_USER" ]; then POSTGRES_USER="postgres"; fi
if [ -z "$POSTGRES_PASSWORD" ]; then POSTGRES_PASSWORD="postgres"; fi
if [ -z "$PORT" ]; then PORT="8080"; fi
if [ -z "$BCRYPT_COST" ]; then BCRYPT_COST="12"; fi

# If DB_URL isn't provided (e.g. no .env), construct it for docker-compose.
if [ -z "$DB_URL" ]; then
  DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable"
  export DB_URL
fi

# If JWT_SECRET isn't provided, generate a random one at startup.
if [ -z "$JWT_SECRET" ]; then
  JWT_SECRET="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  export JWT_SECRET
fi

migrate -path /migrations -database "$DB_URL" up
exec /server

