# Listener Security Guide

## Overview

GODNSLOG provides four OAST (Out-of-Band Application Security Testing) protocol
listeners: SMTP, LDAP, SMB, and FTP. These listeners accept inbound connections
from the public Internet to capture callback evidence during security assessments.

This document describes the security posture of each listener, explains the
default-disable policy, and provides operational guidance for secure deployment.

---

## Risk Level Summary

| Listener | Risk Level | Rationale |
|----------|-----------|-----------|
| SMTP     | Medium     | Email protocol; acceptable surface, bounded parsing |
| LDAP     | Medium     | Simplified ASN.1/BER parser, minimal feature surface |
| SMB      | High       | Complex protocol, raw binary parsing, privileged port |
| FTP      | Medium     | Plain-text command protocol, clear-text credentials |
| RMI      | Medium     | Java RMI protocol; used for JNDI injection detection |

---

## Default-Disable Policy

**All protocol listeners are DISABLED by default.** A newly created listener will
not be started by the Manager until its `IsEnabled` field is explicitly set to
`true`.

### Enforcement Points

1. **Database schema** (`internal/models/listener.go`):
   `IsEnabled bool xorm:"bool notnull default false"`

2. **Manager startup** (`internal/listener/manager.go`, line 62-64):
   ```go
   for _, l := range listeners {
       if !l.IsEnabled {
           continue
       }
   ```
   Only listeners whose `IsEnabled` field is `true` in the database are started.

3. **Test coverage** (`internal/listener/manager_test.go`, `TestManager_StartStop`):
   A listener with `IsEnabled: false` results in 0 active listeners.

### Activation

To enable a listener, set `"is_enabled": true` in the create request body, or
call `PUT /listeners/{id}` with `{"is_enabled": true}` after creation. The
Manager's `RestartListener` method must be invoked after updating `IsEnabled`
for the change to take effect at runtime.

---

## Security Infrastructure (Shared)

All four listeners share the same security infrastructure:

### Rate Limiting (`internal/listener/ratelimit.go`)

- Sliding-window per-IP rate limiter using `RateLimiter`
- Window: 1 minute
- Max connections per window: configured per protocol (default: 500-1000)
- Thread-safe via `sync.Mutex`; cleanup runs every 5 minutes

### Connection Limiting (`internal/listener/ratelimit.go`)

- Max concurrent connections enforced via `ConnLimiter`
- Default limits: SMTP 500, LDAP 500, SMB 200, FTP 200
- Callers must call `ReleaseConnection()` when done (guaranteed via `defer`)

### Security Context (`internal/listener/security_context.go`)

- Every listener's accept loop calls `GetSecurityContext(listenerID)`
- `CheckConnection()` evaluates in order:
  1. Nil-context guard (returns true for safety)
  2. IP whitelist bypass (CIDR-based, skips all limits)
  3. Rate limit check
  4. Concurrent connection limit check
- Rejected connections are recorded as `ListenerInteraction` records with
  `reject_reason` metadata for analyst visibility
- `ReleaseConnection()` decrements the connection counter; safe to call on nil

### Connection Timeout

- All listeners apply `conn.SetDeadline()` using the configured timeout
- Default: 30 seconds for all protocols
- Prevents resource exhaustion from slow/broken clients

---

## Listener-Specific Security Analysis

### SMTP Listener (`internal/listener/smtp.go`)

**Risk Level: Medium**

**Attack Surface:**
- Binds TCP (default port 25 or user-configured)
- Accepts SMTP commands: HELO/EHLO, MAIL FROM, RCPT TO, DATA, QUIT, RSET
- Stores email body, headers, sender, and recipients in the database

**Security Mechanisms:**
- SecurityContext rate/conn limiting on accept
- Timeout on connection read/write
- Command whitelist: unknown commands return `500 Command not recognized`
- No SMTP AUTH implemented (reduces authentication attack surface)
- No relay functionality (DATA is accepted but never forwarded)

**Concerns:**
- Email body size is unbounded; only the 30-second timeout limits input. A
  client sending data slowly could keep a connection open for the full timeout.
  The body `strings.Builder` grows in memory during the connection.

### LDAP Listener (`internal/listener/ldap.go`)

**Risk Level: Medium**

**Attack Surface:**
- Binds TCP (default port 389 or user-configured)
- Accepts raw LDAP (ASN.1/BER) messages
- Simplified parser extracts BaseDN, Filter, and BindDN

**Security Mechanisms:**
- SecurityContext rate/conn limiting on accept
- Timeout on connection read/write
- Simplified parser limits parsing scope; complex LDAP operations are ignored
- Response is a hardcoded success message; no sensitive data leaked

**Concerns:**
- The LDAP parser uses raw byte indexing and string searching on the input
  buffer rather than a full ASN.1/BER decoder. Malformed or unexpected input
  may produce incorrect extractions but the bounded read buffer limits the
  blast radius to the connection goroutine only.
- No LDAP StartTLS or SASL authentication paths implemented (reduces surface).

### SMB Listener (`internal/listener/smb.go`)

**Risk Level: High**

**Attack Surface:**
- Binds TCP (default port 445 or user-configured)
- Port 445 requires root on Linux, elevating the process privilege profile
- Accepts raw SMB protocol messages (complex binary format)

**Security Mechanisms:**
- SecurityContext rate/conn limiting on accept
- Timeout on connection read/write
- Minimal parsing: only the SMB magic bytes (`\xFFSMB`) and command byte are
  interpreted; the raw packet bytes are JSON-encoded before storage
- No SMB dialect negotiation or session setup (no authentication material
  accepted beyond what is logged)

**Concerns:**
- The SMB protocol is historically complex and has been the vector for
  critical vulnerabilities (e.g., EternalBlue). While this implementation is
  minimal, running any SMB-facing service carries inherent risk.
- Uses `log.Printf` instead of the structured `*logrus.Logger` (inconsistency
  with SMTP/LDAP listeners).
- Share name and file path extraction uses simple string operations on raw
  binary data, which may produce garbled output.

**Mitigation:** Deploy SMB listener only on isolated test networks. Use
iptables to restrict source IPs when possible.

### FTP Listener (`internal/listener/ftp.go`)

**Risk Level: Medium**

**Attack Surface:**
- Binds TCP (default port 21 or user-configured)
- Accepts FTP commands: USER, PASS, QUIT, SYST, TYPE, PASV, LIST, RETR, STOR
- Stores command arguments (including passwords) in plain text

**Security Mechanisms:**
- SecurityContext rate/conn limiting on accept
- Timeout on connection read/write
- Command whitelist: unknown commands return `502 Command not implemented`
- No anonymous access enforcement (all passwords are accepted)
- RETR/STOR return `550 File not found` -- no actual file access
- PASV returns a fake address -- no data channel is opened

**Concerns:**
- FTP passwords are logged and stored in plain text. While this is expected for
  an OAST testing tool (you want to capture the credentials being tested),
  the database should be treated as sensitive.
- The FTP listener has no actual file storage; RETR and STOR are stubs.
- Uses `log.Printf` instead of the structured `*logrus.Logger` (inconsistency
  with SMTP/LDAP listeners).

### RMI Listener (`internal/listener/rmi.go`)

**Risk Level: Medium**

**Attack Surface:**
- Binds TCP (default port 1099 or user-configured)
- Accepts Java RMI protocol messages
- Used for JNDI injection detection in Java applications

**Security Mechanisms:**
- SecurityContext rate/conn limiting on accept
- Timeout on connection read/write
- Minimal RMI protocol parsing; only captures callbacks without
  performing Java deserialization
- Returns fixed response; no actual RMI registry operations

**Concerns:**
- RMI protocol has historically been associated with deserialization
  vulnerabilities. This implementation does not perform deserialization
  beyond basic protocol identification.
- Default RMI port (1099) is a common scan target.

---

## Production Deployment Recommendations

### Network Isolation

Restrict inbound access to listener ports using iptables:

```bash
# Allow only specific test target IPs
iptables -A INPUT -p tcp --dport 2525 -s <TESTER_IP> -j ACCEPT  # SMTP
iptables -A INPUT -p tcp --dport 389  -s <TESTER_IP> -j ACCEPT  # LDAP
iptables -A INPUT -p tcp --dport 445  -s <TESTER_IP> -j ACCEPT  # SMB
iptables -A INPUT -p tcp --dport 21   -s <TESTER_IP> -j ACCEPT  # FTP

# Drop everything else
iptables -A INPUT -p tcp --dport 2525 -j DROP
iptables -A INPUT -p tcp --dport 389  -j DROP
iptables -A INPUT -p tcp --dport 445  -j DROP
iptables -A INPUT -p tcp --dport 21   -j DROP
```

Or with a default-drop policy:

```bash
# Set default policy to DROP for INPUT
iptables -P INPUT DROP

# Allow established connections
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Allow specific services
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT  # Web UI / API
# ... add listener ports as needed
```

### Minimal Port List

For production, only open ports that are actively in use. If a listener is
enabled, its port must be open on the firewall. Below is the full set of ports
used by GODNSLOG:

| Port  | Protocol | Service       | Required |
|-------|----------|---------------|----------|
| 53    | UDP/TCP  | DNS           | Yes      |
| 80    | TCP      | HTTP          | Optional |
| 8080  | TCP      | Web API/UI    | Yes      |
| 25    | TCP      | SMTP          | If used  |
| 389   | TCP      | LDAP          | If used  |
| 445   | TCP      | SMB           | If used  |
| 21    | TCP      | FTP           | If used  |
| 1099  | TCP      | RMI           | If used  |

### TLS Configuration

Listeners that support TLS (SMTP, LDAP) should be configured with valid
certificates when deployed on untrusted networks. TLS is disabled by default.

```go
config.EnableTLS = true
config.TLSCertFile = "/path/to/cert.pem"
config.TLSKeyFile  = "/path/to/key.pem"
```

When TLS is enabled, the manager applies `tls.Config{MinVersion: tls.VersionTLS12}`
to all TLSCapable listeners (`internal/listener/manager.go`, line 280).

### Monitoring

- Rate limit rejections are logged and stored as `ListenerInteraction` records
  with `reject_reason` metadata. Monitor for spikes, which may indicate
  scanning or abuse.
- All listener bind and accept errors are logged.
- Use the `GET /listeners/{id}/interactions` API to review all interactions,
  including rejected connections.

---

## Testing Checklist

Before deploying listener changes to production:

- [ ] All listeners are disabled by default (`IsEnabled: false`)
- [ ] Manager only starts listeners with `IsEnabled == true`
- [ ] Each accept loop calls `GetSecurityContext(id)` and `CheckConnection()`
- [ ] Each connection handler calls `defer sc.ReleaseConnection()`
- [ ] Connection timeout is configured and non-zero
- [ ] Rate limiter and connection limiter values are appropriate for expected load
- [ ] Whitelist CIDRs are validated (invalid CIDRs logged but not blocking)
- [ ] Unit tests pass: `go test ./internal/listener/... -v`
- [ ] Ports not in use are closed at the firewall
