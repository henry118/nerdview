#!/bin/bash
# Stress test: rapidly create and terminate containers to generate high event volume.
# Usage: sudo ./scripts/stress-events.sh [count] [parallelism]
#   count       — total containers to create (default: 100)
#   parallelism — concurrent container lifecycles (default: 10)

set -euo pipefail

export PATH="/usr/local/bin:/usr/bin:/opt/containerd/bin:$PATH"

if ! command -v nerdctl &>/dev/null; then
    echo "Error: nerdctl not found" >&2
    exit 1
fi
COUNT="${1:-100}"
PARALLEL="${2:-10}"
IMAGE="alpine:latest"

echo "Pulling image (one-time)..."
nerdctl pull "$IMAGE" >/dev/null 2>&1 || true

echo "Starting stress test: $COUNT containers, $PARALLEL concurrent"
echo "Watch nerdview in another terminal to observe debounce behavior."
echo ""

run_lifecycle() {
    local id
    id=$(nerdctl run -d "$IMAGE" sleep 30 2>/dev/null)
    if [ -n "$id" ]; then
        sleep 5
        nerdctl rm -f "$id" >/dev/null 2>&1 || true
    fi
}

started=0
active=0

for ((i = 1; i <= COUNT; i++)); do
    run_lifecycle &
    active=$((active + 1))
    started=$((started + 1))

    if [ "$active" -ge "$PARALLEL" ]; then
        wait -n 2>/dev/null || true
        active=$((active - 1))
    fi

    if ((started % 10 == 0)); then
        echo "  launched $started/$COUNT containers..."
    fi
done

echo "Waiting for remaining containers to finish..."
wait

echo "Done. $COUNT container lifecycles completed."
