#!/bin/bash
# Stress test: rapidly create and terminate containers to generate high event volume.
# Usage: sudo ./scripts/stress-events.sh [count] [parallelism]
#   count       — total containers to create (default: 100)
#   parallelism — concurrent container lifecycles (default: 10)

set -euo pipefail

NERDCTL="/usr/local/bin/nerdctl"
COUNT="${1:-100}"
PARALLEL="${2:-10}"
IMAGE="alpine:latest"

echo "Pulling image (one-time)..."
$NERDCTL pull "$IMAGE" >/dev/null 2>&1 || true

echo "Starting stress test: $COUNT containers, $PARALLEL concurrent"
echo "Watch nerdview in another terminal to observe debounce behavior."
echo ""

run_lifecycle() {
    local id
    id=$($NERDCTL run -d "$IMAGE" sleep 30 2>/dev/null)
    if [ -n "$id" ]; then
        sleep 5
        $NERDCTL rm -f "$id" >/dev/null 2>&1 || true
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
