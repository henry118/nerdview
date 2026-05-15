#!/bin/bash
# Manage test images for nerdview.
# Usage: sudo ./scripts/images.sh pull   — pull 20 popular images
#        sudo ./scripts/images.sh clean  — remove those images

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
    hashicorp/vault:latest
    hashicorp/consul:latest
    registry:latest
    hello-world:latest
)

pull_images() {
    echo "Pulling ${#images[@]} images..."
    for img in "${images[@]}"; do
        echo "  pulling $img"
        nerdctl pull "$img" >/dev/null 2>&1 || echo "    FAILED: $img"
    done
    echo "Done."
}

clean_images() {
    echo "Removing ${#images[@]} images..."
    for img in "${images[@]}"; do
        echo "  removing $img"
        nerdctl rmi "$img" >/dev/null 2>&1 || true
    done
    echo "Done."
}

case "${1:-}" in
    pull)
        pull_images
        ;;
    clean)
        clean_images
        ;;
    *)
        echo "Usage: $0 {pull|clean}" >&2
        exit 1
        ;;
esac
