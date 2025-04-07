#!/bin/bash

set -e

echo "Checking for existing namespace..."
if kubectl get namespace whosay &> /dev/null; then
  echo "Existing namespace found. Deleting resources..."
  kubectl delete deployment whosay -n whosay --ignore-not-found
  kubectl delete service whosay -n whosay --ignore-not-found
  kubectl delete namespace whosay
  echo "Resources deleted."
else
  echo "No existing namespace found."
fi

# Wait a moment for resources to be fully deleted
sleep 3

echo "Installing with Helm..."
# Set createNamespace to true for the first installation
helm upgrade --install whosay ./helm/whosay -f ./helm/whosay/values-dev.yaml --set createNamespace=true

echo "Installation complete. Checking status..."
kubectl get all -n whosay

echo "Done! Your Whosay application should be running now."
