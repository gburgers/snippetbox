FROM golang:1.24.0-alpine3.21 AS builder
WORKDIR /src

# Copy only what's needed for building
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source files
COPY cmd/web/*.go ./cmd/web/
COPY internal/* ./internal/
COPY internal/models/*.go ./internal/models/
COPY internal/validator/*.go ./internal/validator/
COPY ui/* ./ui/

# Build the application
RUN go build -o /app/snippetbox ./cmd/web

# Start fresh with a smaller image
FROM alpine:3.21
WORKDIR /app

# Copy only the binary from the build stage
COPY --from=builder /app/snippetbox .
# Copy other runtime assets if needed
COPY --from=builder /src/ui ./ui

# Set the entry point
ENTRYPOINT ["./snippetbox"]
