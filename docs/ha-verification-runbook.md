# GoDNSLog HA Verification Runbook

This runbook documents the step-by-step procedure for verifying High Availability (HA)
functionality in GoDNSLog, including leader election, node registration, and failover.

## Prerequisites

- Go 1.14+ (for building the binary)
- Docker/Podman with compose support
- MySQL 8.0 and Redis images available
- Ports 3306, 6379, 8081, 8082 available

## Architecture

The HA test environment consists of:

| Component    | Role                              |
|-------------|-----------------------------------|
| godnslog-1  | GoDNSLog instance 1 (port 8081)   |
| godnslog-2  | GoDNSLog instance 2 (port 8082)   |
| MySQL 8.0   | Shared database for cluster state |
| Redis 7     | Leader election and session cache |

Both godnslog instances share the same MySQL database and Redis instance. Leader
election uses the database with an optimistic locking mechanism (term-based). The
leader holds a 30-second lease and must renew it periodically.

## Fixes Applied

### 1. Entrypoint Redis passthrough (`deploy/docker/entrypoint.sh`)

The entrypoint script was modified to pass the `-redis` flag when the `REDIS_URL`
environment variable is set. This is required because leader election is only
enabled when a Redis client is configured.

Before:
```bash
/app/godnslog serve -domain "${DOMAIN:-example.com}" -4 "${DNS_IP:-0.0.0.0}" &
```

After:
```bash
if [ -n "$REDIS_URL" ]; then
  /app/godnslog serve -domain "${DOMAIN:-example.com}" -4 "${DNS_IP:-0.0.0.0}" -redis "$REDIS_URL" &
else
  /app/godnslog serve -domain "${DOMAIN:-example.com}" -4 "${DNS_IP:-0.0.0.0}" &
fi
```

### 2. Shared database (`docker-compose.ha.yml`)

The compose file was updated to include MySQL as a shared database backend, since
both instances must use the same database for leader election coordination. The
default SQLite setup gives each instance its own isolated database, preventing
cluster-wide leader election.

## Verification Procedure

### Step 0: Build the binary

```bash
# Build the GoDNSLog binary
go build -o /tmp/godnslog .
```

### Step 1: Start shared infrastructure

Start MySQL and Redis containers:

```bash
# Start MySQL
podman run -d --name godnslog-mysql --rm \
  -e MYSQL_ROOT_PASSWORD=godnslog \
  -e MYSQL_DATABASE=godnslog \
  -e MYSQL_USER=godnslog \
  -e MYSQL_PASSWORD=godnslog \
  -p 3306:3306 mysql:8.0 \
  --default-authentication-plugin=mysql_native_password \
  --max-connections=200

# Start Redis
podman run -d --name godnslog-redis --rm \
  -p 6379:6379 redis:7-alpine \
  redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru

# Wait for MySQL to be ready
for i in $(seq 1 30); do
  if podman exec godnslog-mysql mysqladmin ping -h localhost -u root -pgodnslog 2>/dev/null; then
    echo "MySQL is ready!"
    break
  fi
  sleep 2
done
```

### Step 2: Start both godnslog instances

```bash
# Instance 1 (port 8081)
/tmp/godnslog serve \
  -domain example.com \
  -4 127.0.0.1 \
  -http :8081 \
  -redis 127.0.0.1:6379 \
  -driver mysql \
  -dsn "godnslog:godnslog@tcp(127.0.0.1:3306)/godnslog?charset=utf8mb4&parseTime=True&loc=Local" \
  -test &

# Instance 2 (port 8082)
/tmp/godnslog serve \
  -domain example.com \
  -4 127.0.0.1 \
  -http :8082 \
  -redis 127.0.0.1:6379 \
  -driver mysql \
  -dsn "godnslog:godnslog@tcp(127.0.0.1:3306)/godnslog?charset=utf8mb4&parseTime=True&loc=Local" \
  -test &

# Verify both are healthy
curl -s http://localhost:8081/api/v2/health
# Expected: {"code":0,"data":{"status":"alive"},...}

curl -s http://localhost:8082/api/v2/health
# Expected: {"code":0,"data":{"status":"alive"},...}
```

### Step 3: Log in and get auth token

```bash
TOKEN1=$(curl -s -X POST http://localhost:8081/api/v2/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"test123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

TOKEN2=$(curl -s -X POST http://localhost:8082/api/v2/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"test123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
```

Note: Each instance generates its own JWT signing key, so tokens are not
interchangeable between instances. Login separately to each instance.

### Step 4: Verify leader election

```bash
# Check leader from instance 1
curl -s http://localhost:8081/api/v2/cluster/leader \
  -H "Authorization: Bearer $TOKEN1"
# Expected: {"code":0,"data":{"leader_id":"HOSTNAME-:8081"},"message":"success"}

# Check leader from instance 2
curl -s http://localhost:8082/api/v2/cluster/leader \
  -H "Authorization: Bearer $TOKEN2"
# Expected: {"code":0,"data":{"leader_id":"HOSTNAME-:8081"},"message":"success"}
```

Expected: Both instances report the same leader (the first instance that acquired
the lock). The leader_id follows the pattern `{hostname}-:{port}`.

### Step 5: Verify cluster nodes

```bash
# List cluster nodes
curl -s http://localhost:8081/api/v2/cluster/nodes \
  -H "Authorization: Bearer $TOKEN1"
# Expected: {"code":0,"data":{"items":[
#   {"id":"HOSTNAME-:8081","status":"online",...},
#   {"id":"HOSTNAME-:8082","status":"online",...}
# ],"total":2},"message":"success"}

# Check cluster status
curl -s http://localhost:8081/api/v2/cluster/status \
  -H "Authorization: Bearer $TOKEN1"
# Expected: {"code":0,"data":{"cluster_status":"healthy","online_nodes":2,...}}
```

Expected: 2 nodes online, cluster status "healthy".

### Step 6: Verify failover

```bash
# Kill the leader (instance 1)
kill $(pgrep -f "/tmp/godnslog.*:8081")

# Wait for leader lease to expire (30 seconds)
sleep 35

# Check new leader on instance 2
curl -s http://localhost:8082/api/v2/cluster/leader \
  -H "Authorization: Bearer $TOKEN2"
# Expected: {"code":0,"data":{"leader_id":"HOSTNAME-:8082"},"message":"success"}

# Check cluster status after failover
curl -s http://localhost:8082/api/v2/cluster/nodes \
  -H "Authorization: Bearer $TOKEN2"
# Expected: instance 1 status "offline", instance 2 status "online"
```

Expected: Instance 2 becomes the new leader. Instance 1 is marked "offline".

### Step 7: Verify recovery after restart

```bash
# Restart instance 1
/tmp/godnslog serve \
  -domain example.com \
  -4 127.0.0.1 \
  -http :8081 \
  -redis 127.0.0.1:6379 \
  -driver mysql \
  -dsn "godnslog:godnslog@tcp(127.0.0.1:3306)/godnslog?charset=utf8mb4&parseTime=True&loc=Local" \
  -test &

# Wait for startup
sleep 10

# Check both nodes
curl -s http://localhost:8082/api/v2/cluster/nodes \
  -H "Authorization: Bearer $TOKEN2"
# Expected: {"code":0,"data":{"items":[
#   {"id":"HOSTNAME-:8081","status":"online",...},
#   {"id":"HOSTNAME-:8082","status":"online",...}
# ],"total":2},"message":"success"}
```

Expected: Both nodes back online. Instance 2 remains the leader (no preemption).

### Step 8: Cleanup

```bash
# Kill godnslog instances
pkill -f "/tmp/godnslog" 2>/dev/null || true

# Stop MySQL and Redis containers
podman stop godnslog-mysql godnslog-redis 2>/dev/null || true
```

## Verification Results Summary

| Step | Test                          | Expected Result                                      | Status |
|------|-------------------------------|------------------------------------------------------|--------|
| 1    | Start infrastructure          | MySQL and Redis containers running and healthy       | PASS   |
| 2    | Start godnslog instances      | Both instances return 200 on /api/v2/health          | PASS   |
| 3    | Authentication                | JWT tokens obtained from both instances              | PASS   |
| 4    | Leader election               | Both instances report same leader node               | PASS   |
| 5    | Cluster nodes                 | 2 nodes online, cluster status "healthy"             | PASS   |
| 6    | Failover                      | Instance 2 becomes leader after instance 1 is killed | PASS   |
| 7    | Recovery                      | Both nodes online after instance 1 restart           | PASS   |

## Troubleshooting

### "no space left on device" during Docker build

The Docker build copies 576 MB of `node_modules` to the final stage, which can
exhaust the overlay filesystem quota in Podman. Workaround: build the binary
directly on the host with `go build` and run it outside containers.

### JWT tokens are not interchangeable

Each godnslog instance generates its own random JWT signing key (`verifyKey`)
at startup. A token from instance 1 will be rejected by instance 2 with a 401
error. Always login separately to each instance.

### Leader lease duration

The leader lease is hard-coded to 30 seconds in `internal/ha/election.go`.
After a leader fails, it takes up to 30 seconds for another node to acquire
the lock. During this window, there is no active leader.

### Node ID format

Node IDs follow the pattern `{hostname}-:{port}`, e.g., `myhost-:8081`. The
hostname is obtained from `os.Hostname()` at startup.

## Docker Compose Alternative

For a fully containerized setup, use `docker-compose.ha.yml`:

```bash
# Build and start
docker compose -f docker-compose.ha.yml up -d --build

# Check health
curl -s http://localhost:8081/api/v2/health
curl -s http://localhost:8082/api/v2/health
```

Note: The Docker build may fail in Podman environments due to overlay filesystem
size limits. If this occurs, use the host-based procedure described above.
