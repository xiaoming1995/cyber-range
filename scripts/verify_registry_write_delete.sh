#!/bin/bash

REGISTRY_URL="http://localhost:5000"
TEST_IMAGE="test-delete"
TEST_TAG="v1"

echo "1. Pulling busybox to use as test image..."
docker pull busybox:latest

echo "2. Tagging and pushing to local registry..."
docker tag busybox:latest localhost:5000/$TEST_IMAGE:$TEST_TAG
docker push localhost:5000/$TEST_IMAGE:$TEST_TAG

echo "3. Verifying it exists in catalog..."
curl -s "$REGISTRY_URL/v2/_catalog" | grep $TEST_IMAGE

echo "4. Getting manifest digest..."
# Debug: print full headers
curl -v -I -H "Accept: application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json" "$REGISTRY_URL/v2/$TEST_IMAGE/manifests/$TEST_TAG"

# Try with correct Accept header for v2 and OCI
DIGEST=$(curl -v -I -H "Accept: application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json" "$REGISTRY_URL/v2/$TEST_IMAGE/manifests/$TEST_TAG" 2>&1 | grep -i Docker-Content-Digest | awk '{print $3}' | tr -d '\r')

echo "Digest: $DIGEST"

if [ -z "$DIGEST" ]; then
    echo "❌ Failed to get digest for fresh image"
    exit 1
fi

echo "5. Attempting to delete manifest..."
curl -v -X DELETE "$REGISTRY_URL/v2/$TEST_IMAGE/manifests/$DIGEST"

echo "6. Verifying deletion..."
# Should return 404
curl -v -I -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "$REGISTRY_URL/v2/$TEST_IMAGE/manifests/$TEST_TAG"
