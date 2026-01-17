#!/bin/bash

# commit-push.sh - Schedule a git commit and push after N hours
# Usage: ./commit-push.sh -12  (commits and pushes in 12 hours)

if [[ $# -ne 1 ]] || [[ ! "$1" =~ ^-[0-9]+$ ]]; then
    echo "Usage: $0 -<hours>"
    echo "Example: $0 -12  (commit and push in 12 hours)"
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
