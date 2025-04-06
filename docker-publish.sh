#!/bin/bash

# Check if DOCKER_USERNAME and DOCKER_PASSWORD are set
if [ -z "$DOCKER_USERNAME" ] || [ -z "$DOCKER_PASSWORD" ]; then
  echo "Error: DOCKER_USERNAME and DOCKER_PASSWORD must be set as environment variables"
  echo "Example: DOCKER_USERNAME=your_username DOCKER_PASSWORD=your_token ./docker-publish.sh"
  exit 1
fi

# Set the image name
IMAGE_NAME="tiwariparth/whosay"
VERSION=$(grep -E 'Version: ".*"' ./config/config.go | sed -E 's/.*Version: "(.*)",/\1/')

echo "Building version: $VERSION"

# Build the Docker image
echo "Building Docker image..."
docker build -t "$IMAGE_NAME:latest" -t "$IMAGE_NAME:$VERSION" .

# Login to Docker Hub
echo "Logging in to Docker Hub..."
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

# Push the Docker image
echo "Pushing Docker image to Docker Hub..."
docker push "$IMAGE_NAME:latest"
docker push "$IMAGE_NAME:$VERSION"

echo "Docker image pushed successfully!"
echo "You can now pull it with: docker pull $IMAGE_NAME:latest"
