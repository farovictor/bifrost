# Makefile for Bifrost

# setup: Install Go 1.23.8 and project dependencies
setup:
	go install golang.org/dl/go1.23.8@latest
	go1.23.8 download
	go mod download

# build: Compile all Go packages using Go 1.23.8
build:
	go build ./...

# run: Execute the main application using Go 1.23.8
run:
	go run main.go

test:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

# compose-up: start Docker Compose environment
compose-up:
		@docker compose up -d --build
	@token=$$(docker compose logs setup-job | tail -n 1 | awk '{print $$3}'); \
	echo "Your token is: $$token"

# compose-down: stop Docker Compose environment
compose-down:
	docker compose down --remove-orphans --volumes

# compose-attach: Attach to web-server
compose-attach:
	docker exec -it bifrost-bifrost-1 sh

