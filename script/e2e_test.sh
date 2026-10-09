#!/bin/bash
set -eu

node_image=$1
container_id=

cleanup() {
    exit_code=$?
    trap - EXIT
    trap '' INT TERM
    if [ -n "$container_id" ]; then
        if [ "$exit_code" -ne 0 ]; then
            docker logs "$container_id" >&2 || true
        fi
        docker stop -t 10 "$container_id" >/dev/null || true
        if ! docker rm -f "$container_id" >/dev/null; then
            if [ "$exit_code" -eq 0 ]; then
                exit_code=1
            fi
        fi
    fi
    exit "$exit_code"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

container_id=$(docker create -p 127.0.0.1::30001 "$node_image")
docker start "$container_id" >/dev/null
address=$(docker port "$container_id" 30001/tcp)
deadline=$((SECONDS + 60))
while :; do
    if [ "$(docker inspect --format '{{.State.Running}}' "$container_id")" != true ]; then
        echo "iserver container exited before becoming ready" >&2
        exit 1
    fi
    remaining=$((deadline - SECONDS))
    if [ "$remaining" -le 0 ]; then
        echo "iserver container was not ready within 60 seconds" >&2
        exit 1
    fi
    if [ "$remaining" -gt 2 ]; then
        remaining=2
    fi
    if curl -fsS --noproxy '*' --max-time "$remaining" "http://$address/getChainInfo" >/dev/null 2>&1; then
        break
    fi
    sleep 1
done

docker exec "$container_id" ./itest run a_case
docker exec "$container_id" ./itest run t_case
docker exec "$container_id" ./itest run c_case
