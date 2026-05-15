#!/bin/bash
# Pull 20 popular Docker Hub images for testing nerdview's image view.
# Usage: sudo ./scripts/pull-images.sh

set -euo pipefail

export PATH="/usr/local/bin:/usr/bin:/opt/containerd/bin:$PATH"

if ! command -v nerdctl &>/dev/null; then
    echo "Error: nerdctl not found" >&2
    exit 1
fi

images=(
    alpine:latest
    busybox:latest
    nginx:latest
    redis:latest
    postgres:latest
    mysql:latest
    node:lts-alpine
    python:3-alpine
    golang:alpine
    ubuntu:latest
    debian:latest
    httpd:latest
    memcached:latest
    mongo:latest
    rabbitmq:latest
    traefik:latest
    vault:latest
    consul:latest
    registry:latest
    hello-world:latest
)

echo "Pulling ${#images[@]} images..."

for img in "${images[@]}"; do
    echo "  pulling $img"
    nerdctl pull "$img" >/dev/null 2>&1 || echo "    FAILED: $img"
done

echo "Done."
