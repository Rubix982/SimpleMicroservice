#!/bin/bash

# Deploy the application to the dev environment
echo "Deploying to dev environment..."

kubectl apply -f main-deployment.yaml

echo "Done."
