# ---- Stage 1: Build ----
FROM golang:1.22-alpine AS builder

# Enable Go modules
ENV GO111MODULE=on

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first for better caching of dependencies
COPY go.mod go.sum ./

# Download and cache dependencies
RUN go mod download

# Copy the rest of the application source code (including main.go)
COPY . .

# Build the Go application into a binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/microservice main.go

# ---- Stage 2: Final minimal image ----
FROM alpine:3.18

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory in the final container
WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=builder /app/microservice /app/microservice

# Use the non-root user
USER appuser

# Expose the port that the service runs on
EXPOSE 8080

# Start the application
ENTRYPOINT ["/app/microservice"]
