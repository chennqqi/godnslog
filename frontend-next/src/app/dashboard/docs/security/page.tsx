'use client'

import { DocPageLayout } from '../doc-page-layout'

export default function SecurityDoc() {
  return (
    <DocPageLayout titleKey="docs.security">
      <h2>Authentication</h2>
      <ul>
        <li><strong>JWT-based</strong>: All API requests require a valid JWT token in the <code>Access-Token</code> header.</li>
        <li><strong>Token expiry</strong>: JWT tokens expire after 24 hours by default.</li>
        <li><strong>4-role system</strong>: Super Admin (0), Admin (1), Normal (2), Guest (3).</li>
        <li><strong>Password storage</strong>: Passwords are hashed with bcrypt.</li>
      </ul>

      <h2>API Key Security</h2>
      <ul>
        <li><strong>Scoped access</strong>: Each API Key has defined scopes limiting accessible endpoints.</li>
        <li><strong>Expiration</strong>: API Keys support configurable expiration times.</li>
        <li><strong>Risk levels</strong>: Tools and operations are classified by risk level (low/medium/high/critical).</li>
        <li><strong>Audit trail</strong>: All API Key usage is logged with timestamp, IP, and action.</li>
        <li><strong>Revocation</strong>: Keys can be revoked at any time.</li>
      </ul>

      <h2>Agent &amp; MCP Security</h2>
      <ul>
        <li><strong>Permission checks</strong>: Every MCP tool call validates scope and risk level.</li>
        <li><strong>High-risk restrictions</strong>: Destructive operations (delete, revoke) require explicit authorization.</li>
        <li><strong>Audit logging</strong>: All Agent and MCP operations are recorded in the audit log.</li>
        <li><strong>Permission denied alerts</strong>: Denied operations are logged with reason (missing scope or risk level exceeded).</li>
      </ul>

      <h2>Outbound Request Security</h2>
      <p>
        Workflow action executors (HTTP, Webhook, Notify) enforce outbound security:
      </p>
      <ul>
        <li><strong>SSRF protection</strong>: Allowlist-based URL filtering blocks internal IPs (169.254.169.254, 127.0.0.0/8, 10.0.0.0/8, etc.).</li>
        <li><strong>Timeout limits</strong>: All outbound requests have configurable timeouts.</li>
        <li><strong>Response size limits</strong>: Prevents memory exhaustion from large responses.</li>
        <li><strong>Template sandbox</strong>: Template rendering uses safe variable substitution only.</li>
      </ul>

      <h2>Listener Security</h2>
      <ul>
        <li><strong>Rate limiting</strong>: Per-IP connection rate limits with configurable thresholds.</li>
        <li><strong>Max concurrent connections</strong>: Prevents resource exhaustion.</li>
        <li><strong>CIDR allowlist</strong>: Bypass rate limits for trusted IPs.</li>
        <li><strong>Connection timeout</strong>: Idle connections are automatically closed.</li>
      </ul>

      <h2>Data Security</h2>
      <ul>
        <li><strong>Self-hosted</strong>: All interaction data stays within your organization.</li>
        <li><strong>Data retention</strong>: Configure automatic cleanup and archival policies.</li>
        <li><strong>Evidence export</strong>: Support for data redaction in exported reports.</li>
        <li><strong>Audit logs</strong>: All critical operations are logged with user, action, resource, and IP.</li>
      </ul>

      <h2>Deployment Security</h2>
      <ul>
        <li><strong>Run as non-root</strong>: Docker image runs as unprivileged user by default.</li>
        <li><strong>Capability binding</strong>: Only <code>cap_net_bind_service</code> for port 53.</li>
        <li><strong>TLS support</strong>: Configure TLS for HTTP listeners.</li>
        <li><strong>Reverse proxy</strong>: Use Nginx (config provided in <code>deploy/nginx/</code>) for TLS termination.</li>
      </ul>

      <h2>Best Practices</h2>
      <ul>
        <li>Change the default admin password immediately after first login.</li>
        <li>Use API Keys with minimal scopes for integrations.</li>
        <li>Regularly review audit logs for suspicious activity.</li>
        <li>Configure data retention policies to limit stored data.</li>
        <li>Use HTTPS in production with a reverse proxy.</li>
        <li>Restrict network access to the GODNSLOG server using firewalls.</li>
        <li>Keep the software updated to the latest version.</li>
      </ul>
    </DocPageLayout>
  )
}
