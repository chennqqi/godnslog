# GODNSLOG Burp Suite Extension

A Burp Suite extension for GODNSLOG 2.0 OAST integration.

## Features

- Generate OAST payloads directly from Burp Suite
- Insert payloads into requests
- Monitor for interactions in real-time
- Add notes to interactions
- Export evidence reports

## Installation

1. Build the extension:
```bash
cd extensions/burp
./build.sh
```

2. Load the JAR file in Burp Suite:
   - Extender -> Extensions -> Add
   - Select the generated JAR file

## Usage

### Generate Payload

1. Select text in Repeater/Proxy
2. Right-click -> GODNSLOG -> Generate Payload
3. Configure payload options (type, case, expiration)
4. Click Generate

### Monitor Interactions

1. Open GODNSLOG tab
2. View real-time interactions
3. Click interaction to view details
4. Add notes for evidence

### Export Report

1. Select interactions
2. Click Export
3. Choose format (JSON, Markdown, CSV)

## Configuration

Configure API endpoint and API key in extension settings:
- API URL: http://localhost:8080/api/v2
- API Key: Your GODNSLOG API key

## Development

Requirements:
- Java 11+
- Burp Suite API
- Maven

Build:
```bash
mvn clean package
```

## License

MIT
