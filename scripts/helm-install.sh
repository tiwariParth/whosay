#!/bin/bash

set -e

echo "Whosay Helm Chart Installation Helper"
echo "===================================="
echo

# Parse command line arguments
environment="dev"
namespace_flag="--set createNamespace=true"

while [[ $# -gt 0 ]]; do
  case $1 in
    --prod)
      environment="prod"
      shift
      ;;
    --no-namespace)
      namespace_flag=""
      shift
      ;;
    --help)
      echo "Usage: $0 [options]"
      echo 
      echo "Options:"
      echo "  --prod           Use production values"
      echo "  --no-namespace   Don't create namespace"
      echo "  --help           Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

echo "Environment: $environment"
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

values_file="./helm/whosay/values-${environment}.yaml"
echo "Installing with Helm using values from $values_file..."

# Set createNamespace to true for the first installation
helm upgrade --install whosay ./helm/whosay -f "$values_file" $namespace_flag

echo
echo "Installation complete. Checking status..."
kubectl get all -n whosay

echo
echo "Done! Your Whosay application should be running now."
echo
echo "To access the logs, run:"
echo "  kubectl logs -n whosay -l app=whosay"
echo
echo "For more information, run:"
echo "  helm status whosay"
