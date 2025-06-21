#!/bin/bash
set -e

# Start the compose stack
docker compose up -d

# Give the services a moment to initialize
sleep 5

# Run the existing rate limit test against the running stack
./test-rate-limiting.sh

# Tear down the stack after the test
docker compose down
