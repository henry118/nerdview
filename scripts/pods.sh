#!/bin/bash
# Manage test pods via crictl.
# Usage: sudo ./scripts/pods.sh run    — create 5 pods with 2 containers each
#        sudo ./scripts/pods.sh clean  — remove all test pods

set -euo pipefail

export PATH="/usr/local/bin:/usr/bin:/opt/containerd/bin:$PATH"

if ! command -v crictl &>/dev/null; then
    echo "Error: crictl not found" >&2
    exit 1
fi

POD_PREFIX="nerdview-test"
POD_COUNT=5
CONTAINERS_PER_POD=2
PAUSE_IMAGE="registry.k8s.io/pause:3.10"
APP_IMAGE="docker.io/library/alpine:latest"

run_pods() {
    echo "Pulling images..."
    crictl pull "$PAUSE_IMAGE" >/dev/null 2>&1 || true
    crictl pull "$APP_IMAGE" >/dev/null 2>&1 || true

    echo "Creating $POD_COUNT pods with $CONTAINERS_PER_POD containers each..."

    for i in $(seq 1 $POD_COUNT); do
        pod_name="${POD_PREFIX}-pod-${i}"

        pod_id=$(crictl runp <(cat <<EOF
{
  "metadata": {
    "name": "${pod_name}",
    "namespace": "default",
    "uid": "nerdview-${i}"
  },
  "log_directory": "/tmp/${pod_name}"
}
EOF
        ))
        echo "  pod ${pod_name}: ${pod_id}"

        for j in $(seq 1 $CONTAINERS_PER_POD); do
            ctr_name="${pod_name}-ctr-${j}"
            mkdir -p "/tmp/${pod_name}"

            ctr_id=$(crictl create "$pod_id" <(cat <<EOF
{
  "metadata": { "name": "${ctr_name}" },
  "image": { "image": "${APP_IMAGE}" },
  "command": ["sleep", "3600"],
  "log_path": "${ctr_name}.log"
}
EOF
            ) <(cat <<EOF
{
  "metadata": {
    "name": "${pod_name}",
    "namespace": "default",
    "uid": "nerdview-${i}"
  },
  "log_directory": "/tmp/${pod_name}"
}
EOF
            ))
            crictl start "$ctr_id" >/dev/null
            echo "    container ${ctr_name}: ${ctr_id}"
        done
    done

    echo "Done. $POD_COUNT pods running."
}

clean_pods() {
    echo "Cleaning up test pods..."

    pods=$(crictl pods --name "${POD_PREFIX}" -q 2>/dev/null || true)
    if [ -z "$pods" ]; then
        echo "  no test pods found."
        return
    fi

    for pod_id in $pods; do
        crictl stopp "$pod_id" >/dev/null 2>&1 || true
        crictl rmp "$pod_id" >/dev/null 2>&1 || true
        echo "  removed pod ${pod_id}"
    done

    echo "Done."
}

case "${1:-}" in
    run)
        run_pods
        ;;
    clean)
        clean_pods
        ;;
    *)
        echo "Usage: $0 {run|clean}" >&2
        exit 1
        ;;
esac
