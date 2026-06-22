'use client'

import { DocPageLayout } from '../doc-page-layout'

export default function UserGuideDoc() {
  return (
    <DocPageLayout titleKey="docs.user_guide">
      <h2>Cases</h2>
      <p>
        A Case represents a testing task, vulnerability verification, or exercise project.
        Each Case can contain multiple Payloads and captures all related Interactions.
      </p>
      <ul>
        <li><strong>Create</strong>: Navigate to Cases → New Case, fill in title, description, and tags.</li>
        <li><strong>Detail View</strong>: Click a case to see stats, payloads, interactions timeline.</li>
        <li><strong>Edit/Close</strong>: Update status (active/closed/archived) from the detail page.</li>
      </ul>

      <h2>Payloads</h2>
      <p>
        Payloads are trackable probes with unique tokens. Each payload is associated with a Case
        and supports template-based generation with variable substitution.
      </p>
      <ul>
        <li><strong>Templates</strong>: Built-in templates for SSRF, XXE, RCE, Blind SQLi, and more.</li>
        <li><strong>Variables</strong>: <code>{`{{token}}`}</code>, <code>{`{{case_id}}`}</code>, <code>{`{{domain}}`}</code> are auto-replaced.</li>
        <li><strong>Preview</strong>: Use the preview feature to see the rendered payload before injection.</li>
        <li><strong>Revoke</strong>: Revoke a payload to stop tracking new interactions for it.</li>
      </ul>

      <h2>Interactions</h2>
      <p>
        Interactions are captured out-of-band events (DNS queries, HTTP requests, SMTP connections, etc.).
        They are automatically attributed to Cases and Payloads via token matching.
      </p>
      <ul>
        <li><strong>Timeline View</strong>: Visual chronological display of all interactions.</li>
        <li><strong>Filters</strong>: Filter by protocol type, token, source IP, time range.</li>
        <li><strong>Real-time</strong>: Use the SSE stream endpoint for live updates.</li>
        <li><strong>Export</strong>: Export interactions as JSON, Markdown, or CSV.</li>
      </ul>

      <h2>Canary Tokens</h2>
      <p>
        Canary tokens are long-lived decoy probes for detecting unauthorized access, data exfiltration,
        or lateral movement. Supported types: DNS, HTTP, web bug, and more.
      </p>
      <ul>
        <li><strong>Create</strong>: Choose a type, set optional expiry and description.</li>
        <li><strong>Monitor</strong>: View hits on the Canary page; each hit shows source IP and timestamp.</li>
        <li><strong>Revoke</strong>: Revoke a canary token when no longer needed.</li>
      </ul>

      <h2>DNS Rebinding</h2>
      <p>
        Configure DNS rebinding rules to return different IP addresses on first vs subsequent queries.
        Useful for testing SSRF, bypassing access controls, and attacking internal services.
      </p>
      <ul>
        <li><strong>Scenarios</strong>: Pre-defined rebinding scenarios (e.g., browser rebinding).</li>
        <li><strong>Rules</strong>: Create custom rules with specific IP stages and TTL.</li>
        <li><strong>Sessions</strong>: View active rebinding sessions and their current stage.</li>
      </ul>

      <h2>Protocol Listeners</h2>
      <p>
        Start SMTP, LDAP, SMB, or FTP listeners to capture out-of-band interactions
        beyond DNS and HTTP. Each listener has its own token for attribution.
      </p>

      <h2>Workflow Rules</h2>
      <p>
        Define rules that trigger actions when interactions match specific criteria.
        Supported actions: HTTP request, DNS lookup, SMTP email, Webhook, and multi-channel notifications.
      </p>

      <h2>Scanner Hub</h2>
      <p>
        Integrate with scanners like Nuclei. Create scanner runs, associate them with cases,
        and backfill results to automatically correlate scan findings with OAST interactions.
      </p>

      <h2>Agent Runs</h2>
      <p>
        Track AI Agent testing sessions. Each agent run records the target, steps, payloads used,
        interactions captured, and supports a full review workflow with evidence export.
      </p>
    </DocPageLayout>
  )
}
