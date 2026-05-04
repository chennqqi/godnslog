# Protocol Listeners

Multi-protocol listener system for OAST (Out-of-Band Application Security Testing).

## Features

- **Multiple Protocols**: SMTP, LDAP, SMB, FTP
- **Interaction Tracking**: Record all protocol interactions
- **Token-based**: Associate interactions with specific tokens
- **Configurable**: Customizable timeout, buffer size, TLS settings
- **Storage**: Persistent storage for all interactions

## Supported Protocols

### SMTP Listener
- Full SMTP server implementation
- Captures MAIL FROM, RCPT TO, DATA
- Records headers and body
- Source IP tracking

### LDAP Listener
- LDAP server implementation
- Captures bind requests, searches
- Records base DN, filters, attributes
- Source IP tracking

### SMB Listener
- SMB server implementation
- Captures connection attempts and basic SMB commands
- Records share names, file paths, and operations
- Source IP and port tracking
- Simplified SMB protocol parsing

### FTP Listener
- FTP server implementation
- Captures login, commands, and file operations
- Records USER, PASS, LIST, RETR, STOR commands
- Source IP and port tracking
- Basic FTP protocol support

## Usage

### Create SMTP Listener

```bash
curl -X POST http://localhost:8080/api/v2/listeners \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "protocol": "smtp",
    "host": "0.0.0.0",
    "port": 2525,
    "token": "smtp-token-abc123",
    "is_enabled": true
  }'
```

### Create LDAP Listener

```bash
curl -X POST http://localhost:8080/api/v2/listeners \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "protocol": "ldap",
    "host": "0.0.0.0",
    "port": 389,
    "token": "ldap-token-abc123",
    "is_enabled": true
  }'
```

### Create SMB Listener

```bash
curl -X POST http://localhost:8080/api/v2/listeners \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "protocol": "smb",
    "host": "0.0.0.0",
    "port": 445,
    "token": "smb-token-abc123",
    "is_enabled": true
  }'
```

### Create FTP Listener

```bash
curl -X POST http://localhost:8080/api/v2/listeners \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "protocol": "ftp",
    "host": "0.0.0.0",
    "port": 21,
    "token": "ftp-token-abc123",
    "is_enabled": true
  }'
```

### List Listeners

```bash
curl http://localhost:8080/api/v2/listeners \
  -H "Authorization: Bearer $API_KEY"
```

### Get SMTP Messages

```bash
curl http://localhost:8080/api/v2/listeners/{id}/smtp \
  -H "Authorization: Bearer $API_KEY"
```

### Get LDAP Queries

```bash
curl http://localhost:8080/api/v2/listeners/{id}/ldap \
  -H "Authorization: Bearer $API_KEY"
```

### Get SMB Requests

```bash
curl http://localhost:8080/api/v2/listeners/{id}/smb \
  -H "Authorization: Bearer $API_KEY"
```

### Get FTP Commands

```bash
curl http://localhost:8080/api/v2/listeners/{id}/ftp \
  -H "Authorization: Bearer $API_KEY"
```

## Programmatic Usage

### Start SMTP Listener

```go
listener := &listener.Listener{
    ID:        "listener-1",
    Protocol:  listener.ProtocolSMTP,
    Host:      "0.0.0.0",
    Port:      2525,
    Token:     "smtp-token-abc123",
    IsEnabled: true,
}

config := listener.DefaultSMTPConfig()
smtpListener := listener.NewSMTPListener(listener, config, store, logger)

ctx := context.Background()
if err := smtpListener.Start(ctx); err != nil {
    log.Fatal(err)
}
```

### Start LDAP Listener

```go
listener := &listener.Listener{
    ID:        "listener-2",
    Protocol:  listener.ProtocolLDAP,
    Host:      "0.0.0.0",
    Port:      389,
    Token:     "ldap-token-abc123",
    IsEnabled: true,
}

config := listener.DefaultLDAPConfig()
ldapListener := listener.NewLDAPListener(listener, config, store, logger)

ctx := context.Background()
if err := ldapListener.Start(ctx); err != nil {
    log.Fatal(err)
}
```

### Start SMB Listener

```go
listener := &listener.Listener{
    ID:        "listener-3",
    Protocol:  listener.ProtocolSMB,
    Host:      "0.0.0.0",
    Port:      445,
    Token:     "smb-token-abc123",
    IsEnabled: true,
}

config := &listener.ListenerConfig{
    MaxConnections: 10,
    Timeout:        30 * time.Second,
    BufferSize:     4096,
}
smbListener := listener.NewSMBListener(config, store, listener)

ctx := context.Background()
if err := smbListener.Start(ctx); err != nil {
    log.Fatal(err)
}
```

### Start FTP Listener

```go
listener := &listener.Listener{
    ID:        "listener-4",
    Protocol:  listener.ProtocolFTP,
    Host:      "0.0.0.0",
    Port:      21,
    Token:     "ftp-token-abc123",
    IsEnabled: true,
}

config := &listener.ListenerConfig{
    MaxConnections: 10,
    Timeout:        30 * time.Second,
    BufferSize:     4096,
}
ftpListener := listener.NewFTPListener(config, store, listener)

ctx := context.Background()
if err := ftpListener.Start(ctx); err != nil {
    log.Fatal(err)
}

## Configuration

### SMTP Configuration

```go
config := &listener.ListenerConfig{
    MaxConnections: 100,
    Timeout:        30 * time.Second,
    BufferSize:     4096,
    EnableTLS:      false,
    TLSCertFile:    "",
    TLSKeyFile:     "",
}
```

### LDAP Configuration

```go
config := &listener.ListenerConfig{
    MaxConnections: 100,
    Timeout:        30 * time.Second,
    BufferSize:     4096,
    EnableTLS:      false,
}
```

## Data Models

### Listener
- ID: Unique identifier
- Protocol: SMTP, LDAP, SMB, FTP
- Host: Bind address
- Port: Bind port
- Token: Association token
- IsEnabled: Enable/disable status

### SMTPMessage
- ID: Unique identifier
- ListenerID: Associated listener
- From: Sender email
- To: Recipient emails
- Subject: Email subject
- Body: Email body
- Headers: Email headers
- SourceIP: Source IP address
- Timestamp: Capture time

### LDAPQuery
- ID: Unique identifier
- ListenerID: Associated listener
- BaseDN: Base distinguished name
- Filter: Search filter
- Attributes: Requested attributes
- BindDN: Bind distinguished name
- SourceIP: Source IP address
- Timestamp: Capture time

## Security Considerations

1. **Port Binding**: Listeners bind to specified ports, ensure proper firewall rules
2. **Token Security**: Use unique, unpredictable tokens
3. **TLS Support**: Enable TLS for production deployments
4. **Access Control**: Restrict listener management to authorized users
5. **Audit Logging**: Log all listener operations

## Use Cases

### SMTP Testing
- Email injection testing
- Mail server misconfiguration detection
- Email header analysis
- Phishing simulation

### LDAP Testing
- LDAP injection testing
- Directory service enumeration
- Authentication bypass testing
- Attribute disclosure

### SMB Testing
- SMB file access testing
- Share enumeration
- Authentication testing
- Lateral movement detection

### FTP Testing
- FTP credential testing
- File upload/download testing
- Command injection testing
- Anonymous access detection

### Integration with Payload Studio
- Generate SMTP/LDAP payloads
- Associate with specific cases
- Track interactions over time

## Limitations

- SMTP: Basic implementation, no full RFC compliance
- LDAP: Simplified ASN.1 parsing, not full LDAP protocol
- SMB/FTP: Not yet implemented
- No authentication enforcement
- No rate limiting

## Future Enhancements

- Full RFC compliance for SMTP
- Complete LDAP protocol support
- SMB and FTP listener implementation
- Authentication and authorization
- Rate limiting
- TLS support for all protocols
- WebSocket real-time updates
- Protocol-specific filtering

## Troubleshooting

### Port Already in Use
- Check if port is already bound
- Use different port
- Stop conflicting services

### No Connections
- Verify firewall rules
- Check listener is enabled
- Verify host binding
- Check network connectivity

### Data Not Saved
- Verify database connection
- Check storage implementation
- Review error logs

## API Endpoints

### POST /listeners
Create a new listener

### GET /listeners
List all listeners

### GET /listeners/{id}
Get a listener by ID

### PUT /listeners/{id}
Update a listener

### DELETE /listeners/{id}
Delete a listener

### GET /listeners/{id}/interactions
List interactions for a listener

### GET /listeners/{id}/smtp
List SMTP messages for a listener

### GET /listeners/{id}/ldap
List LDAP queries for a listener

### GET /listeners/{id}/smb
List SMB requests for a listener

### GET /listeners/{id}/ftp
List FTP commands for a listener

### GET /listeners/config
Get listener configuration
