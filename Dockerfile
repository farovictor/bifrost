# Stage 1: build the Go binary
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o bifrost-server main.go
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o bifrost-cli ./cmd/bifrost

# Stage 2: create a minimal runtime image
FROM alpine:latest
RUN addgroup -S bifrost && adduser -S bifrost -G bifrost
WORKDIR /app
COPY --from=builder /app/bifrost-server /app/
COPY --from=builder /app/bifrost-cli /app/
USER bifrost
EXPOSE 3333
CMD ["/app/bifrost-server"]
