#!/bin/sh
# Entrypoint script for GoDNSLog container
# Go backend: TLS, DNS, API, and reverse proxy to Next.js frontend
# Next.js: internal-only SSR server (port 3000, not exposed to host)

set -e

trap 'kill -TERM $BACKEND_PID $FRONTEND_PID 2>/dev/null; wait; exit 0' TERM INT

# Build backend command arguments
BACKEND_ARGS="-domain ${DOMAIN:-example.com} -4 ${DNS_IP:-0.0.0.0}"

# Redis
if [ -n "$REDIS_URL" ]; then
  BACKEND_ARGS="$BACKEND_ARGS -redis $REDIS_URL"
fi

# SQLite database path. Defaults to /data (persistent volume) so data survives
# container re-creates. Override with GODNSLOG_DB_PATH if needed.
if [ -z "$GODNSLOG_MYSQL_DSN" ]; then
  DB_PATH="${GODNSLOG_DB_PATH:-/data/godnslog.db}"
  BACKEND_ARGS="$BACKEND_ARGS -dsn file:${DB_PATH}?cache=shared&mode=rwc"
else
  BACKEND_ARGS="$BACKEND_ARGS -driver mysql -dsn $GODNSLOG_MYSQL_DSN"
fi

# TLS mode (env: GODNSLOG_TLS_MODE)
TLS_MODE="${GODNSLOG_TLS_MODE:-disabled}"
if [ "$TLS_MODE" != "disabled" ] && [ -n "$TLS_MODE" ]; then
  BACKEND_ARGS="$BACKEND_ARGS -tls-mode $TLS_MODE"
  if [ -n "$GODNSLOG_ACME_EMAIL" ]; then
    BACKEND_ARGS="$BACKEND_ARGS -acme-email $GODNSLOG_ACME_EMAIL"
  fi
  CERT_DIR="${GODNSLOG_CERT_DIR:-/data/certs}"
  BACKEND_ARGS="$BACKEND_ARGS -cert-dir $CERT_DIR"
  if [ -n "$GODNSLOG_TLS_CERT" ]; then
    BACKEND_ARGS="$BACKEND_ARGS -tls-cert $GODNSLOG_TLS_CERT"
  fi
  if [ -n "$GODNSLOG_TLS_KEY" ]; then
    BACKEND_ARGS="$BACKEND_ARGS -tls-key $GODNSLOG_TLS_KEY"
  fi
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
case "$BACKEND_ARGS" in
  *-http\ *) ;;
  *) BACKEND_ARGS="$BACKEND_ARGS -http $GODNSLOG_HTTP_LISTEN" ;;
esac

echo "[entrypoint] Starting backend with args: $BACKEND_ARGS"
/app/godnslog serve $BACKEND_ARGS &
BACKEND_PID=$!

# Start Next.js standalone server on loopback only (Go reverse-proxies to it)
cd /app/frontend && HOSTNAME=127.0.0.1 node server.js -p 3000 &
FRONTEND_PID=$!

wait -n $BACKEND_PID $FRONTEND_PID 2>/dev/null
EXIT_CODE=$?

kill -TERM $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
wait $BACKEND_PID $FRONTEND_PID 2>/dev/null || true

exit $EXIT_CODE
