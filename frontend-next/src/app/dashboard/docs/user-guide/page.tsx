'use client'

import { DocPageLayout } from '../doc-page-layout'

const content = `## Cases

A Case represents a testing task, vulnerability verification, or exercise project. Each Case can contain multiple Payloads and captures all related Interactions.

- **Create**: Navigate to Cases → New Case, fill in title, description, and tags.
- **Detail View**: Click a case to see stats, payloads, interactions timeline.
- **Edit/Close**: Update status (active/closed/archived) from the detail page.

## Payloads

Payloads are trackable probes with unique tokens. Each payload is associated with a Case and supports template-based generation with variable substitution.

- **Templates**: Built-in templates for SSRF, XXE, RCE, Blind SQLi, and more.
- **Variables**: \`{{token}}\`, \`{{case_id}}\`, \`{{domain}}\` are auto-replaced.
- **Preview**: Use the preview feature to see the rendered payload before injection.
- **Revoke**: Revoke a payload to stop tracking new interactions for it.

## Interactions

Interactions are captured out-of-band events (DNS queries, HTTP requests, SMTP connections, etc.). They are automatically attributed to Cases and Payloads via token matching.

- **Timeline View**: Visual chronological display of all interactions.
- **Filters**: Filter by protocol type, token, source IP, time range.
- **Real-time**: Use the SSE stream endpoint for live updates.
- **Export**: Export interactions as JSON, Markdown, or CSV.

## Canary Tokens

Canary tokens are long-lived decoy probes for detecting unauthorized access, data exfiltration, or lateral movement. Supported types: DNS, HTTP, web bug, and more.

- **Create**: Choose a type, set optional expiry and description.
- **Monitor**: View hits on the Canary page; each hit shows source IP and timestamp.
- **Revoke**: Revoke a canary token when no longer needed.

## DNS Rebinding

Configure DNS rebinding rules to return different IP addresses on first vs subsequent queries. Useful for testing SSRF, bypassing access controls, and attacking internal services.

- **Scenarios**: Pre-defined rebinding scenarios (e.g., browser rebinding).
- **Rules**: Create custom rules with specific IP stages and TTL.
- **Sessions**: View active rebinding sessions and their current stage.

## Protocol Listeners

Start SMTP, LDAP, SMB, or FTP listeners to capture out-of-band interactions beyond DNS and HTTP. Each listener has its own token for attribution. All listeners are **disabled by default** for security — enable them explicitly via the Listener configuration.

## Workflow Rules

Define rules that trigger actions when interactions match specific criteria. Supported actions: HTTP request, DNS lookup, SMTP email, Webhook, and multi-channel notifications.

## Scanner Hub

Integrate with scanners like Nuclei. Create scanner runs, associate them with cases, and backfill results to automatically correlate scan findings with OAST interactions.

## Agent Runs

Track AI Agent testing sessions. Each agent run records the target, steps, payloads used, interactions captured, and supports a full review workflow with evidence export.`

export default function UserGuideDoc() {
  return <DocPageLayout titleKey="docs.user_guide" md={content} />
}
