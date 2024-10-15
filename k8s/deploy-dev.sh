#!/bin/bash

# Deploy the application to the dev environment
echo "Deploying to dev environment..."

# Create the namespace
kubectl apply -f namespace.yaml

# Create the config maps
cd configmaps || return
kubectl apply -f otel-cm.yaml

# Create the deployments
cd ../deployments || return
kubectl apply -f main-deployment.yaml
kubectl apply -f jaeger-deployment.yaml

# Create the services
cd ../services || return
kubectl apply -f main-service.yaml
kubectl apply -f jaeger-service.yaml

echo "Done."
