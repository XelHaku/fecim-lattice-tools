#!/bin/bash

# commit-push.sh - Schedule a git commit and push after N hours
# Usage: ./commit-push.sh -12  (commits and pushes in 12 hours)
#        ./commit-push.sh --periodically  (commits and pushes every hour)

cd "$(dirname "$0")" || exit 1

# Periodic mode: commit and push every hour (runs in background)
if [[ "$1" == "--periodically" ]]; then
    LOGFILE="$(dirname "$0")/commit-push.log"
    PIDFILE="$(dirname "$0")/.commit-push.pid"

    # Check if already running
    if [[ -f "$PIDFILE" ]] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
        echo "Periodic commit-push is already running (PID: $(cat "$PIDFILE"))"
        echo "Use '$0 --stop' to stop it."
        exit 1
    fi

    echo "Starting periodic commit-push mode (every hour) in background..."

    nohup bash -c "
        cd \"$(dirname "$0")\" || exit 1
        echo \$\$ > \"$PIDFILE\"
        while true; do
            git add -A
            if ! git diff --cached --quiet; then
                git commit -m \"Auto-commit (periodic) at \$(date '+%Y-%m-%d %H:%M:%S')\"
                git push
                echo \"Commit and push completed at \$(date)\"
            else
                echo \"No changes to commit at \$(date)\"
            fi
            echo \"Next commit-push in 1 hour...\"
            sleep 3600
        done
    " > "$LOGFILE" 2>&1 &

    sleep 0.5  # Wait for PID file to be written
    echo "Running in background (PID: $(cat "$PIDFILE"))"
    echo "Log file: $LOGFILE"
    echo "Use '$0 --stop' to stop it."
    exit 0
fi

# Stop periodic mode
if [[ "$1" == "--stop" ]]; then
    PIDFILE="$(dirname "$0")/.commit-push.pid"
    if [[ -f "$PIDFILE" ]]; then
        PID=$(cat "$PIDFILE")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID"
            rm -f "$PIDFILE"
            echo "Stopped periodic commit-push (PID: $PID)"
        else
            rm -f "$PIDFILE"
            echo "Process was not running (stale PID file removed)"
        fi
    else
        echo "No periodic commit-push is running"
    fi
    exit 0
fi

# Delayed mode: commit and push after N hours
if [[ $# -ne 1 ]] || [[ ! "$1" =~ ^-[0-9]+$ ]]; then
    echo "Usage: $0 -<hours>"
    echo "       $0 --periodically"
    echo "       $0 --stop"
    echo "Examples:"
    echo "  $0 -12           (commit and push in 12 hours)"
    echo "  $0 --periodically (commit and push every hour in background)"
    echo "  $0 --stop        (stop periodic mode)"
    exit 1
fi

HOURS="${1#-}"
SECONDS_DELAY=$((HOURS * 3600))

echo "Scheduling commit and push in $HOURS hour(s)..."

# Run in background
(
    sleep "$SECONDS_DELAY"
    cd "$(dirname "$0")" || exit 1
    git add -A
    if ! git diff --cached --quiet; then
        git commit -m "Auto-commit after $HOURS hour delay at $(date '+%Y-%m-%d %H:%M:%S')"
        git push
        echo "Commit and push completed at $(date)"
    else
        echo "No changes to commit at $(date)"
    fi
) &

echo "Scheduled! Process running in background (PID: $!)"
echo "The commit and push will happen at approximately: $(date -d "+$HOURS hours")"
