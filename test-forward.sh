#!/bin/bash
set -e

docker compose up -d --build

# Allow services time to start
sleep 5

TOKEN=$(docker compose logs setup-job | tail -n 1 | awk '{print $3}')

# Create root key
curl -s -X POST http://localhost:3333/v1/rootkeys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: testadmin" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"testroot","api_key":"secret"}'

# Register service using the root key
curl -s -X POST http://localhost:3333/v1/services \
  -H "Content-Type: application/json" \
  -H "X-API-Key: testadmin" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"testsvc","endpoint":"http://testserver:8081","root_key_id":"testroot"}'

# Issue virtual key
curl -s -X POST http://localhost:3333/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-API-Key: testadmin" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"testkey","scope":"read","target":"testsvc","expires_at":"2099-01-01T00:00:00Z","rate_limit":10}'

# Forward a request through Bifrost
curl -s \
  -H "X-Virtual-Key: testkey" \
  -H "X-API-Key: testadmin" \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:3333/v1/proxy/check

docker compose down
