# GODNSLOG 2.9 Deployment Guide

This guide covers the two deployment modes introduced in version 2.9.

## Mode 1: Let's Encrypt Standalone Mode

In this mode, GODNSLOG directly binds ports 443/80/53 and manages TLS certificates
internally via Let's Encrypt ACME, with self-signed certificate fallback.

### Non-container deployment

```bash
# Build
go build -o godnslog

# Run with ACME TLS
./godnslog serve \
  -domain oast.example.com \
  -4 YOUR_PUBLIC_IP \
  -http :443 \
  -tls-mode acme \
  -acme-email admin@example.com \
  -cert-dir /var/lib/godnslog/certs
```

### Container deployment

```bash
# Using docker-compose-standalone.yml
export DOMAIN=oast.example.com
export ACME_EMAIL=admin@example.com
export DNS_IP=YOUR_PUBLIC_IP

docker compose -f docker-compose-standalone.yml up -d
```

### How it works

1. On startup, the TLS manager attempts to load or obtain a Let's Encrypt certificate.
2. If ACME issuance fails (e.g., no network access), a self-signed certificate is generated as fallback.
3. The web server listens on :443 with HTTPS.
4. DNS server listens on :53 (TCP/UDP).
5. DNSLog HTTP verification works on both 443 and 80.

### Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GODNSLOG_TLS_MODE` | TLS mode: `acme`, `static`, `self-signed`, `disabled` | `disabled` |
| `GODNSLOG_ACME_EMAIL` | Email for Let's Encrypt registration | (required for ACME) |
| `GODNSLOG_CERT_DIR` | Directory for storing certificates | `/data/certs` |
| `GODNSLOG_TLS_CERT` | Path to static cert file (for `static` mode) | |
| `GODNSLOG_TLS_KEY` | Path to static key file (for `static` mode) | |

## Mode 2: Nginx Reverse Proxy Mode

In this mode, GODNSLOG listens on high ports (e.g., 18080) and Nginx handles
TLS termination and port 80/443 binding.

### Non-container deployment

```bash
# Start GODNSLOG on high port
./godnslog serve \
  -domain oast.example.com \
  -4 127.0.0.1 \
  -http :18080 \
  -tls-mode disabled

# Configure Nginx (see deploy/nginx/godnslog-https.conf)
# Reload Nginx
nginx -s reload
```

### Container deployment

```bash
# Using docker-compose.nginx.yml
export DOMAIN=oast.example.com
export DNS_IP=YOUR_PUBLIC_IP

docker compose -f docker-compose.nginx.yml up -d
```

### Nginx configuration

A reference configuration is provided at `deploy/nginx/godnslog-https.conf`.
Key features:
- HTTP (port 80) redirects to HTTPS
- HTTPS (port 443) proxies to GODNSLOG backend
- WebSocket support for real-time interaction push
- TLS 1.2/1.3 with modern cipher suites
- HSTS header

## Demo Mode

Demo mode creates pre-configured demo accounts with sample data for evaluation.

### Enable demo mode

```bash
# Via command line flag
./godnslog serve -domain demo.example.com -4 0.0.0.0 -demo

# Via environment variable
GODNSLOG_DEMO=true ./godnslog serve -domain demo.example.com -4 0.0.0.0
```

### Demo accounts

Demo mode creates 3 demo users:
- `demo_user1` / `demo1123`
- `demo_user2` / `demo2123`
- `demo_user3` / `demo3123`

Each demo user has:
- 1 sample Case (SSRF test)
- 1 sample Payload (ssrf-basic template)
- 2 sample Interactions (1 DNS + 1 HTTP)

### Demo user restrictions

Demo users cannot:
- Access admin user management (`/api/admin/`)
- Modify system settings (`/api/setting/`, `/api/v2/settings`)
- Create or manage API keys (`/api/v2/apikeys`)

### Data reset

Demo data is automatically reset every 6 hours (configurable via `--demo-reset-hours`).
