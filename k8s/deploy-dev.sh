#!/bin/bash

# Deploy the application to the dev environment
echo "Deploying to dev environment..."

# Create the namespace
kubectl apply -f namespace.yaml

# Create the config maps
cd configmaps || return
kubectl apply -f otel-cm.yaml
kubectl apply -f postgres-cm.yaml
kubectl apply -f fluentd-cm.yaml

# Create the daemon-sets
cd ../daemons || return
kubectl apply -f fluentd-daemonset.yaml

# Create the PVCs
cd ../pvc || return
kubectl apply -f es-pvc.yaml

# Create the deployments
cd ../deployments || return
kubectl apply -f main-deployment.yaml
kubectl apply -f jaeger-deployment.yaml
kubectl apply -f es-deployment.yaml
kubectl apply -f kibana-deployment.yaml
kubectl apply -f postgres-deployment.yaml

# Create the services
cd ../services || return
kubectl apply -f main-service.yaml
kubectl apply -f jaeger-service.yaml
kubectl apply -f postgres-service.yaml
kubectl apply -f kibana-service.yaml

echo "Done."
