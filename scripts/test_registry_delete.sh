#!/bin/bash

REGISTRY_URL="http://localhost:5000"

echo "1. Listing repositories..."
curl -s "$REGISTRY_URL/v2/_catalog" | python3 -m json.tool

echo -e "\nEnter repository name to test (e.g. nginx):"
read REPO

echo "2. Listing tags for $REPO..."
curl -s "$REGISTRY_URL/v2/$REPO/tags/list" | python3 -m json.tool

echo -e "\nEnter tag to delete (e.g. v1):"
read TAG

echo "3. Getting manifest digest..."
# Debug: print full headers without Accept constraint, force IPv4
curl -v -4 "$REGISTRY_URL/v2/$REPO/manifests/$TAG" > /dev/null

DIGEST=$(curl -s -4 -I -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "$REGISTRY_URL/v2/$REPO/manifests/$TAG" | grep -i Docker-Content-Digest | awk '{print $2}' | tr -d '\r')

echo "Digest: $DIGEST"

if [ -z "$DIGEST" ]; then
    echo "❌ Failed to get digest"
    exit 1
fi

echo "4. Attempting to delete manifest..."
curl -v -X DELETE "$REGISTRY_URL/v2/$REPO/manifests/$DIGEST"
