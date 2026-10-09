#!/bin/bash
set -eu

target_dir=$1
node_pid=

cleanup() {
    exit_code=$?
    trap - EXIT
    trap '' INT TERM
    if [ -n "$node_pid" ]; then
        kill -TERM "$node_pid" 2>/dev/null || true
        stop_deadline=$((SECONDS + 10))
        while kill -0 "$node_pid" 2>/dev/null; do
            if [ "$SECONDS" -ge "$stop_deadline" ]; then
                echo "iserver did not stop within 10 seconds; terminating PID $node_pid" >&2
                kill -KILL "$node_pid" 2>/dev/null || true
                break
            fi
            sleep 1
        done
        wait "$node_pid" 2>/dev/null || true
    fi
    exit "$exit_code"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

"$target_dir/iserver" -f config/iserver.yml &
node_pid=$!
deadline=$((SECONDS + 60))
while :; do
    if ! kill -0 "$node_pid" 2>/dev/null; then
        echo "iserver exited before becoming ready" >&2
        exit 1
    fi
    remaining=$((deadline - SECONDS))
    if [ "$remaining" -le 0 ]; then
        echo "iserver was not ready within 60 seconds" >&2
        exit 1
    fi
    if [ "$remaining" -gt 2 ]; then
        remaining=2
    fi
    if curl -fsS --noproxy '*' --max-time "$remaining" http://127.0.0.1:30001/getChainInfo >/dev/null 2>&1; then
        break
    fi
    sleep 1
done

"$target_dir/itest" run a_case
"$target_dir/itest" run t_case
"$target_dir/itest" run c_case
