#!/bin/sh
# Entrypoint script for GoDNSLog container
# Starts backend and frontend with proper signal handling and crash detection

set -e

# Trap signals and forward to child processes
trap 'kill -TERM $BACKEND_PID $FRONTEND_PID 2>/dev/null; wait; exit 0' TERM INT

# Start Go backend
/app/godnslog serve -domain "${DOMAIN:-example.com}" -4 "${DNS_IP:-0.0.0.0}" &
BACKEND_PID=$!

# Start Next.js frontend
cd /app/frontend && npx next start -p 3000 &
FRONTEND_PID=$!

# Wait for either process to exit; if one dies, kill the other
wait -n $BACKEND_PID $FRONTEND_PID 2>/dev/null
EXIT_CODE=$?

# If one process exited, kill the other
kill -TERM $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
wait $BACKEND_PID $FRONTEND_PID 2>/dev/null || true

exit $EXIT_CODE
