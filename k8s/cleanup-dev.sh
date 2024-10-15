#!/bin/bash

# Perform cleanup operations on the application to the dev environment
echo "Cleaning up the dev environment..."

# Deleting the namespace will delete all resources within it (deployments, services, etc.)
kubectl delete -f namespace.yaml

echo "Done."
