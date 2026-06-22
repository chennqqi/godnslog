'use client'

import { DocPageLayout } from '../doc-page-layout'

export default function QuickStartDoc() {
  return (
    <DocPageLayout titleKey="docs.quick_start">
      <h2>1. Start the Server</h2>
      <p>Run GODNSLOG with your domain and IP address:</p>
      <pre><code>{`go run . serve -domain example.com -4 127.0.0.1`}</code></pre>

      <h2>2. Log In</h2>
      <p>
        Navigate to <code>http://localhost:80</code> and log in with the
        auto-generated admin credentials printed in the server console.
        Change the password immediately via the Settings page.
      </p>

      <h2>3. Create a Case</h2>
      <p>
        Go to <strong>Cases</strong> and click <strong>New Case</strong>.
        Enter a title and description for your testing task.
      </p>

      <h2>4. Generate a Payload</h2>
      <p>
        Navigate to <strong>Payloads</strong> and click <strong>New Payload</strong>.
        Select a template (SSRF, XXE, RCE, etc.), choose your Case,
        and the system will generate a unique token-based payload URL.
      </p>

      <h2>5. Inject and Monitor</h2>
      <p>
        Copy the generated payload URL and inject it into your target application.
        Watch the <strong>Interactions</strong> page for incoming DNS/HTTP callbacks.
        Each interaction is automatically attributed to your Case and Payload.
      </p>

      <h2>6. Export Evidence</h2>
      <p>
        Once you have captured interactions, go to <strong>Evidence</strong>
        to generate and export a report in Markdown or JSON format.
      </p>

      <h2>CLI Quick Start</h2>
      <p>Use the CLI tool for automation:</p>
      <pre><code>{`# Create a case
godnslog-cli case create --title "Test XSS" --description "XSS in search"

# Generate a payload
godnslog-cli payload create --case-id <CASE_ID> --template ssrf

# Poll for interactions
godnslog-cli interaction list --case-id <CASE_ID> --watch

# Export report
godnslog-cli report export --case-id <CASE_ID> --format markdown`}</code></pre>
    </DocPageLayout>
  )
}
