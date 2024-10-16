#!/bin/bash

# Variables for OS and architecture
TARGET_OS="linux"
TARGET_ARCH="amd64"

# Array of services with their paths
services=(
    "item-service:./services/item/src/"
    "order-service:./services/order/src/"
    "payment-service:./services/payment/src/"
    "user-service:./services/user/src/"
)

# Loop through the services array and build each service
for service in "${services[@]}"; do
    # Split the service name and source path
    serviceName=${service%%:*}    # Get the part before the colon
    srcPath=${service#*:}         # Get the part after the colon

    echo "Building $serviceName..."
    GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -o "./build/$serviceName-$TARGET_OS-$TARGET_ARCH.bin" "$srcPath"

    if [ $? -ne 0 ]; then
        echo "Failed to build $serviceName"
        exit 1
    fi
done

echo "All services built successfully."
