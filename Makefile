# Makefile for Bifrost

# setup: Install Go 1.23.8 and project dependencies
setup:
	go install golang.org/dl/go1.23.8@latest
	go1.23.8 download
	go mod download

# build: Compile all Go packages using Go 1.23.8
build:
	go1.23.8 build ./...

# run: Execute the main application using Go 1.23.8
run:
	go1.23.8 run main.go

# compose-up: start Docker Compose environment
compose-up:
	docker compose up -d --build

# compose-down: stop Docker Compose environment
compose-down:
	docker compose down

