#!/bin/bash

# Spin up a Redis before executing requests
# docker run -d --name redis-dev -p 6379:6379 redis:7-alpine

for i in {1..30}; do
  curl -X POST http://localhost:3333/v1/rate \
    -H "X-API-Key: your-api-key-here" \
    -H "Content-Type: application/json" \
    -d '{"message": "hello"}'

  echo # optional: newline for clarity
done
