#!/bin/bash

# commit-push.sh - Schedule a git commit and push after N hours
# Usage: ./commit-push.sh -12  (commits and pushes in 12 hours)
#        ./commit-push.sh --periodically  (commits and pushes every hour)

cd "$(dirname "$0")" || exit 1

# Periodic mode: commit and push every hour
if [[ "$1" == "--periodically" ]]; then
    echo "Starting periodic commit-push mode (every hour)..."
    echo "Press Ctrl+C to stop."

    while true; do
        git add -A
        # Only commit if there are staged changes
        if ! git diff --cached --quiet; then
            git commit -m "Auto-commit (periodic) at $(date '+%Y-%m-%d %H:%M:%S')"
            git push
            echo "Commit and push completed at $(date)"
        else
            echo "No changes to commit at $(date)"
        fi
        echo "Next commit-push in 1 hour..."
        sleep 3600
    done
    exit 0
fi

# Delayed mode: commit and push after N hours
if [[ $# -ne 1 ]] || [[ ! "$1" =~ ^-[0-9]+$ ]]; then
    echo "Usage: $0 -<hours>"
    echo "       $0 --periodically"
    echo "Examples:"
    echo "  $0 -12           (commit and push in 12 hours)"
    echo "  $0 --periodically (commit and push every hour)"
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
    git commit -m "Auto-commit after $HOURS hour delay"
    git push
    echo "Commit and push completed at $(date)"
) &

echo "Scheduled! Process running in background (PID: $!)"
echo "The commit and push will happen at approximately: $(date -d "+$HOURS hours")"
