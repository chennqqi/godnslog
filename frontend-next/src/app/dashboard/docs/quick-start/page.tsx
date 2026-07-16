'use client'

import { DocPageLayout } from '../doc-page-layout'

const content = `## 1. Start the Server

Run GODNSLOG with your domain and IP address:

\`\`\`bash
go run . serve -domain example.com -4 127.0.0.1
\`\`\`

## 2. Log In

Navigate to \`http://localhost:80\` and log in with the auto-generated admin credentials printed in the server console. Change the password immediately via the **Settings** page.

## 3. Create a Case

Go to **Cases** and click **New Case**. Enter a title and description for your testing task.

## 4. Generate a Payload

Navigate to **Payloads** and click **New Payload**. Select a template (SSRF, XXE, RCE, etc.), choose your Case, and the system generates a unique token-based payload URL.

## 5. Inject and Monitor

Copy the generated payload URL and inject it into your target application. Watch the **Interactions** page for incoming DNS/HTTP callbacks. Each interaction is automatically attributed to your Case and Payload.

## 6. Export Evidence

Go to **Evidence** and select your Case or Payload. Choose a format (JSON, CSV, Markdown) to download the interaction log with attribution metadata.

## 7. Evidence Summary

For a higher-level view, use the **Evidence Summary** page to aggregate interactions by technique, protocol, and confidence level. The summary generates AI-ready evidence bundles.`

export default function QuickStartDoc() {
  return <DocPageLayout titleKey="docs.quick_start" md={content} />
}
