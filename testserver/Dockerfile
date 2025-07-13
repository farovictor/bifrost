# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o testserver

# Final image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /src/testserver /app/testserver
EXPOSE 8081
CMD ["/app/testserver"]
