#!/bin/bash

# commit-push.sh - Schedule a git commit and push after N hours
# Usage: ./commit-push.sh -12  (commits and pushes in 12 hours)
#        ./commit-push.sh --periodically  (commits and pushes every 30 minutes)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$SCRIPT_DIR"
REPO_NAME="$(basename "$REPO_DIR")"
LOGFILE="/tmp/commit-push-${REPO_NAME}.log"
PIDFILE="$REPO_DIR/.commit-push.pid"
UNIT="commit-push-${REPO_NAME}"

cd "$REPO_DIR" || exit 1

systemd_available() {
    command -v systemd-run >/dev/null 2>&1 && command -v systemctl >/dev/null 2>&1
}

systemd_unit_active() {
    systemd_available && systemctl --user is-active --quiet "$UNIT.service" 2>/dev/null
}

pid_running() {
    [[ -f "$PIDFILE" ]] || return 1
    local pid
    pid="$(cat "$PIDFILE")"
    [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

periodic_worker() {
    exec >> "$LOGFILE" 2>&1
    echo "$$" > "$PIDFILE"
    trap 'rm -f "$PIDFILE"' EXIT INT TERM

    while true; do
        cd "$REPO_DIR" || exit 1
        # Stage all changes except PID file (log is in /tmp, not repo)
        git add -A -- ':!.commit-push.pid'
        # Only commit and push if there are real changes
        if ! git diff --cached --quiet -- ':!.commit-push.pid'; then
            git commit -m "Auto-commit (periodic) at $(date '+%Y-%m-%d %H:%M:%S')"
            git push
            echo "Commit and push completed at $(date)"
        else
            # Reset any accidentally staged PID file
            git reset HEAD -- .commit-push.pid 2>/dev/null || true
            echo "No real changes to commit at $(date)"
        fi
        echo "Next commit-push in 30 minutes..."
        sleep 1800
    done
}

# Periodic mode: commit and push every 30 minutes (runs in background)
if [[ "$1" == "--periodic-worker" ]]; then
    periodic_worker
    exit 0
fi

if [[ "$1" == "--periodically" ]]; then
    # Check if already running
    if systemd_unit_active; then
        echo "Periodic commit-push is already running (systemd unit: $UNIT)"
        echo "Use '$0 --stop' to stop it."
        exit 1
    fi
    if pid_running; then
        echo "Periodic commit-push is already running (PID: $(cat "$PIDFILE"))"
        echo "Use '$0 --stop' to stop it."
        exit 1
    fi

    echo "Starting periodic commit-push mode (every 30 minutes) in background..."

    # Prefer systemd-run so the process survives terminal closure
    if systemd_available; then
        systemctl --user stop "$UNIT.service" >/dev/null 2>&1 || true
        systemctl --user reset-failed "$UNIT.service" >/dev/null 2>&1 || true
        if systemd-run --user --unit="$UNIT" --description="Periodic git commit/push for $REPO_NAME" \
            --property=WorkingDirectory="$REPO_DIR" "$SCRIPT_DIR/commit-push.sh" --periodic-worker; then
            echo "Running as systemd user unit: $UNIT"
            echo "Log file: $LOGFILE"
            echo "Use '$0 --stop' to stop it."
            exit 0
        else
            echo "systemd-run failed; falling back to detached mode."
        fi
    fi

    # Fallback: detach via nohup (+ setsid when available)
    if command -v setsid >/dev/null 2>&1; then
        nohup setsid "$SCRIPT_DIR/commit-push.sh" --periodic-worker >/dev/null 2>&1 &
    else
        nohup "$SCRIPT_DIR/commit-push.sh" --periodic-worker >/dev/null 2>&1 &
    fi
    sleep 0.5  # Wait for PID file to be written
    if pid_running; then
        echo "Running in background (PID: $(cat "$PIDFILE"))"
    else
        echo "Started in background (PID file pending)"
    fi
    echo "Log file: $LOGFILE"
    echo "Use '$0 --stop' to stop it."
    exit 0
fi

# Stop periodic mode
if [[ "$1" == "--stop" ]]; then
    stopped=0
    if systemd_available && systemctl --user status "$UNIT.service" >/dev/null 2>&1; then
        if systemctl --user is-active --quiet "$UNIT.service"; then
            systemctl --user stop "$UNIT.service"
            echo "Stopped systemd unit ($UNIT)"
            stopped=1
        else
            systemctl --user stop "$UNIT.service" >/dev/null 2>&1 || true
        fi
    fi

    if [[ -f "$PIDFILE" ]]; then
        PID=$(cat "$PIDFILE")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID"
            echo "Stopped periodic commit-push (PID: $PID)"
            stopped=1
        else
            echo "Process was not running (stale PID file removed)"
        fi
        rm -f "$PIDFILE"
    fi

    if [[ "$stopped" -eq 0 ]]; then
        echo "No periodic commit-push is running"
    fi
    exit 0
fi

if [[ "$1" == "--status" ]]; then
    if systemd_unit_active; then
        PID=$(systemctl --user show -p MainPID --value "$UNIT.service" 2>/dev/null || true)
        if [[ -n "$PID" ]] && [[ "$PID" != "0" ]]; then
            echo "Periodic commit-push is running (systemd unit: $UNIT, PID: $PID)"
        else
            echo "Periodic commit-push is running (systemd unit: $UNIT)"
        fi
        echo "Log file: $LOGFILE"
        exit 0
    fi
    if pid_running; then
        echo "Periodic commit-push is running (PID: $(cat "$PIDFILE"))"
        echo "Log file: $LOGFILE"
        exit 0
    fi
    echo "No periodic commit-push is running"
    exit 1
fi

# Delayed mode: commit and push after N hours
if [[ $# -ne 1 ]] || [[ ! "$1" =~ ^-[0-9]+$ ]]; then
    echo "Usage: $0 -<hours>"
    echo "       $0 --periodically"
    echo "       $0 --status"
    echo "       $0 --stop"
    echo "Examples:"
    echo "  $0 -12           (commit and push in 12 hours)"
    echo "  $0 --periodically (commit and push every 30 minutes in background)"
    echo "  $0 --status      (show periodic status)"
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
    # Stage all changes except PID file (log is in /tmp, not repo)
    git add -A -- ':!.commit-push.pid'
    # Only commit and push if there are real changes
    if ! git diff --cached --quiet -- ':!.commit-push.pid'; then
        git commit -m "Auto-commit after $HOURS hour delay at $(date '+%Y-%m-%d %H:%M:%S')"
        git push
        echo "Commit and push completed at $(date)"
    else
        # Reset any accidentally staged PID file
        git reset HEAD -- .commit-push.pid 2>/dev/null || true
        echo "No real changes to commit at $(date)"
    fi
) &

echo "Scheduled! Process running in background (PID: $!)"
echo "The commit and push will happen at approximately: $(date -d "+$HOURS hours")"
