# ---- Stage 1: Build ----
FROM golang:1.22-alpine AS builder

# Enable Go modules
ENV GO111MODULE=on

# Tag the image with the name of the service
LABEL service=order

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first for better caching of dependencies
COPY ./services/user/go.mod ./services/user/go.sum ./

# Download and cache dependencies
RUN go mod download

# Install gosec for security scanning
RUN go install github.com/securego/gosec/v2/cmd/gosec@latest

# Copy the rest of the application source code (including main.go)
COPY . .

# Run the security scanner
RUN gosec ./...

# Copy the built binary from the build folder
COPY ./build/user-service-linux-amd64.bin /app/user-service

# ---- Stage 2: Final minimal image ----
FROM alpine:3.18

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory in the final container
WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=builder /app/user-service /app/microservice

# Use the non-root user
USER appuser

# Expose the port that the service runs on
EXPOSE 8080

# Start the application
ENTRYPOINT ["/app/microservice"]
