#!/bin/sh
# Entrypoint script for GoDNSLog container
# Starts backend and frontend with proper signal handling and crash detection
# Supports dual deployment modes: standalone (TLS) and nginx reverse proxy

set -e

# Trap signals and forward to child processes
trap 'kill -TERM $BACKEND_PID $FRONTEND_PID 2>/dev/null; wait; exit 0' TERM INT

# Build backend command arguments
BACKEND_ARGS="-domain ${DOMAIN:-example.com} -4 ${DNS_IP:-0.0.0.0}"

# Redis
if [ -n "$REDIS_URL" ]; then
  BACKEND_ARGS="$BACKEND_ARGS -redis $REDIS_URL"
fi

# TLS mode (env: GODNSLOG_TLS_MODE)
TLS_MODE="${GODNSLOG_TLS_MODE:-disabled}"
if [ "$TLS_MODE" != "disabled" ] && [ -n "$TLS_MODE" ]; then
  BACKEND_ARGS="$BACKEND_ARGS -tls-mode $TLS_MODE"
  # ACME email
  if [ -n "$GODNSLOG_ACME_EMAIL" ]; then
    BACKEND_ARGS="$BACKEND_ARGS -acme-email $GODNSLOG_ACME_EMAIL"
  fi
  # Certificate directory
  CERT_DIR="${GODNSLOG_CERT_DIR:-/data/certs}"
  BACKEND_ARGS="$BACKEND_ARGS -cert-dir $CERT_DIR"
  # Static cert files
  if [ -n "$GODNSLOG_TLS_CERT" ]; then
    BACKEND_ARGS="$BACKEND_ARGS -tls-cert $GODNSLOG_TLS_CERT"
  fi
  if [ -n "$GODNSLOG_TLS_KEY" ]; then
    BACKEND_ARGS="$BACKEND_ARGS -tls-key $GODNSLOG_TLS_KEY"
  fi
  # In standalone TLS mode, listen on 443
  if [ "$TLS_MODE" = "acme" ] || [ "$TLS_MODE" = "self-signed" ] || [ "$TLS_MODE" = "static" ]; then
    BACKEND_ARGS="$BACKEND_ARGS -http :443"
  fi
fi

# Demo mode (env: GODNSLOG_DEMO)
if [ "$GODNSLOG_DEMO" = "true" ] || [ "$GODNSLOG_DEMO" = "1" ]; then
  BACKEND_ARGS="$BACKEND_ARGS -demo"
fi

# HTTP listen port (default 8080, can be overridden)
if [ -z "$GODNSLOG_HTTP_LISTEN" ]; then
  GODNSLOG_HTTP_LISTEN=":8080"
fi
# Only add -http if not already set by TLS mode
case "$BACKEND_ARGS" in
  *-http\ *) ;; # already set
  *) BACKEND_ARGS="$BACKEND_ARGS -http $GODNSLOG_HTTP_LISTEN" ;;
esac

echo "[entrypoint] Starting backend with args: $BACKEND_ARGS"
/app/godnslog serve $BACKEND_ARGS &
BACKEND_PID=$!

# Start Next.js frontend
FRONTEND_PORT="${FRONTEND_PORT:-3000}"
cd /app/frontend && npx next start -p $FRONTEND_PORT &
FRONTEND_PID=$!

# Wait for either process to exit; if one dies, kill the other
wait -n $BACKEND_PID $FRONTEND_PID 2>/dev/null
EXIT_CODE=$?

# If one process exited, kill the other
kill -TERM $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
wait $BACKEND_PID $FRONTEND_PID 2>/dev/null || true

exit $EXIT_CODE
