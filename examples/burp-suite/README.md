# GODNSLOG OAST Burp Suite Extension

A Burp Suite extension that integrates GODNSLOG OAST functionality directly into Burp Suite.

## Features

- **Right-click context menu**: Generate OAST payload from any request
- **OAST Tab**: View interactions in real-time within Burp Suite
- **Case Management**: Create and manage OAST cases from Burp
- **Evidence Export**: Generate evidence reports without leaving Burp
- **Auto-polling**: Background polling for interactions on active payloads

## Requirements

- Burp Suite Professional or Community Edition (2023.12+)
- Java 17+
- GODNSLOG 2.0 server running and accessible

## Installation

### From Source

1. Build the extension JAR:
   ```bash
   cd examples/burp-suite
   mvn clean package
   ```

2. In Burp Suite:
   - Go to Extensions > Installed > Add
   - Extension type: Java
   - Select the built JAR file at `target/godnslog-burp-extension.jar`
   - Click Next

### Configuration

1. Go to Extensions > GODNSLOG OAST
2. Configure:
   - **API URL**: Your GODNSLOG server URL (e.g., `http://localhost:8080`)
   - **API Key**: Your Agent API Key with `agent:create_probe` and `agent:read_interactions` scopes
3. Click "Test Connection" to verify

## Usage

### Generate Payload

1. Right-click on any request in Proxy/Repeater/Intruder
2. Select "GODNSLOG OAST" > "Generate Payload"
3. Choose a template (SSRF, XXE, RCE, etc.)
4. The payload token is copied to clipboard and added to the OAST tab

### Insert Payload

1. Right-click on a parameter value in Request view
2. Select "GODNSLOG OAST" > "Insert Payload Here"
3. The selected parameter value is replaced with the OAST payload

### View Interactions

1. Open the "GODNSLOG OAST" tab in Burp Suite
2. Active payloads are listed with their interaction counts
3. Click on a payload to see detailed interactions
4. Interactions auto-refresh every 30 seconds

### Export Evidence

1. Right-click on a case in the OAST tab
2. Select "Export Evidence"
3. Choose format: Markdown, JSON, or CSV
4. The report is saved to your specified path

## API Scopes Required

| Scope | Purpose |
|-------|---------|
| `agent:create_probe` | Create cases and payloads |
| `agent:read_interactions` | Poll for interactions |
| `agent:summarize_evidence` | Generate evidence summaries |
| `agent:export_report` | Export evidence reports |

## Architecture

The extension uses the Montoya API (Burp Suite's modern extension API) and communicates with GODNSLOG via REST API calls. No local state is stored — all data is managed by the GODNSLOG server.

```
Burp Suite <-> Extension (Java) <-> HTTP REST API <-> GODNSLOG Server
```

## Development

### Project Structure

```
examples/burp-suite/
├── pom.xml                    # Maven build configuration
├── src/main/java/com/godnslog/burp/
│   ├── GodnslogExtension.java # Main extension entry point
│   ├── GodnslogApiClient.java # REST API client
│   ├── OastTabProvider.java   # OAST tab UI
│   ├── ContextMenuProvider.java # Right-click menu actions
│   └── models/                # Data models
└── README.md
```

### Build

```bash
mvn clean package
```

The built JAR is at `target/godnslog-burp-extension.jar`.
