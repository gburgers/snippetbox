FROM golang:1.24.0-alpine3.21 AS builder
WORKDIR /src

# Copy only what's needed for building
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source files
COPY cmd/web/*.go ./cmd/web/
COPY internal/ ./internal/
COPY ui/ ./ui/

# Build the application
RUN go build -o /app/snippetbox ./cmd/web

# Start fresh with a smaller image
FROM alpine:3.21
WORKDIR /app

# Copy only the binary from the build stage
COPY --from=builder /app/snippetbox .
# Copy other runtime assets if needed

# Create user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set ownership and permission
RUN chown -R appuser:appgroup /app

# Switch to user
USER appuser

# Set the entry point
ENTRYPOINT ["./snippetbox"]
