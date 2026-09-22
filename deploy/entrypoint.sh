#!/bin/sh
set -eu
export PORT=${PORT:-8080}

db=${SQLITE_PATH:-/data/one-api.db}
if [ -z "${SQL_DSN:-}" ] && [ ! -s "$db" ]; then
    echo "Existing database is missing; refusing to initialize an empty production database." >&2
    exit 1
fi

app_pid=
guard_pid=
stop() {
    trap - TERM INT
    [ -z "$guard_pid" ] || kill "$guard_pid" 2>/dev/null || true
    [ -z "$app_pid" ] || kill -TERM "$app_pid" 2>/dev/null || true
    [ -z "$app_pid" ] || wait "$app_pid" 2>/dev/null || true
    exit 0
}
trap stop TERM INT

while :; do
    /chat-api "$@" &
    app_pid=$!
    (
        failures=0
        while kill -0 "$app_pid" 2>/dev/null; do
            sleep 30
            if wget -q -T 5 -O /dev/null "http://127.0.0.1:${PORT:-8080}/healthz"; then
                failures=0
            else
                failures=$((failures + 1))
                if [ "$failures" -ge 3 ]; then
                    echo "Health check failed three times; restarting application process." >&2
                    kill -TERM "$app_pid" 2>/dev/null || true
                    sleep 10
                    kill -KILL "$app_pid" 2>/dev/null || true
                    exit
                fi
            fi
        done
    ) &
    guard_pid=$!
    code=0
    wait "$app_pid" || code=$?
    kill "$guard_pid" 2>/dev/null || true
    wait "$guard_pid" 2>/dev/null || true
    echo "Application exited ($code); restarting in five seconds." >&2
    sleep 5
done
