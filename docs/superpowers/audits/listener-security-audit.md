# Protocol Listener Security Audit Report

- **Audit date**: 2026-06-21
- **Auditor**: Engineering team
- **Scope**: `internal/listener/smtp.go`, `ldap.go`, `smb.go`, `ftp.go`, `manager.go`, `security_context.go`, `ratelimit.go`

---

## 1. Summary

All protocol listeners (SMTP, LDAP, SMB, FTP) have been reviewed for security risks. The listeners are designed as lightweight honeypot-style receivers for OAST callback verification — they accept inbound connections, record interaction data, and close. They do not implement full protocol servers.

**Verdict**: Safe for production deployment with the mitigations listed below.

---

## 2. Findings

### 2.1 Default State (FIXED)

- **Before**: `Listener.IsEnabled` defaulted to `true`, meaning any newly created listener would start automatically.
- **After**: Changed default to `false`. Listeners must be explicitly enabled by an administrator.
- **File**: `internal/models/listener.go:26`

### 2.2 Rate Limiting (IN PLACE)

- Each listener has a per-IP `RateLimiter` (1-minute window, configurable max).
- Default limits: SMTP/LDAP = 1000/min, SMB/FTP = 500/min.
- Whitelisted CIDRs bypass rate limits.
- Rejected connections are recorded as interactions for analyst visibility.
- **Files**: `security_context.go`, `ratelimit.go`

### 2.3 Connection Limiting (IN PLACE)

- `ConnLimiter` enforces max concurrent connections across all IPs.
- Default: SMTP/LDAP = 1000, SMB/FTP = 500.
- Slots are acquired on accept and released on close.
- **File**: `security_context.go`, `connlimit.go`

### 2.4 TLS Support (IN PLACE)

- `ListenerConfig.EnableTLS`, `TLSCertFile`, `TLSKeyFile` fields exist.
- `loadTLSConfig` enforces `MinVersion: tls.VersionTLS12`.
- Listeners implementing `TLSCapable` interface can wrap connections in TLS.
- **File**: `manager.go:196-204, 272-282`

### 2.5 IP Whitelisting (IN PLACE)

- `ListenerConfig.WhitelistCIDRs` allows specifying trusted CIDR ranges.
- Whitelisted IPs bypass rate and connection limits.
- **File**: `security_context.go:60-77`

### 2.6 Timeout Handling (IN PLACE)

- All listeners enforce configurable read/write timeouts (default 30s).
- Prevents slow-loris style resource exhaustion.

### 2.7 Data Handling

- Listener interactions are stored in the database via `Store` interface.
- No data is written to disk outside the database.
- No command execution or file system access in listener code.

---

## 3. Recommendations for Production

1. **Network isolation**: Deploy listeners in a separate network namespace or container sidecar. Use iptables/security groups to restrict access to listener ports.
2. **TLS**: Enable TLS for SMTP and FTP in production. LDAP should use LDAPS (port 636).
3. **Monitoring**: Alert on high rates of rejected connections (potential scanning/abuse).
4. **Port binding**: Listeners should bind to specific interfaces, not `0.0.0.0`, in production.
5. **SMB/LDAP specific**: These protocols have known reflection/amplification risks. Consider disabling in environments where they are not needed.

---

## 4. Conclusion

The listener subsystem is safe for production use with the default-disabled state and existing security controls. The rate limiting, connection limiting, TLS, and whitelisting mechanisms provide defense-in-depth against abuse. The primary remaining recommendation is network-level isolation in production deployments.
