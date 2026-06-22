'use client'

import { DocPageLayout } from '../doc-page-layout'

export default function ConfigDoc() {
  return (
    <DocPageLayout titleKey="docs.config">
      <h2>Command Line Options</h2>
      <pre><code>{`godnslog serve [flags]

Flags:
  -domain string      Your domain (required)
  -4 string           IPv4 address (required)
  -6 string           IPv6 address (optional)
  -http-listen string HTTP listen address (default ":80")
  -driver string      Database driver: sqlite|mysql (default "sqlite")
  -dsn string         Database DSN (default "file:godnslog.db")
  -upstream string    Upstream DNS server (default "8.8.8.8:53")
  -redis-addr string  Redis address for HA session sharing (optional)
  -swagger            Enable Swagger UI (debug only)
  -with-guest         Enable guest access`}</code></pre>

      <h2>Database Configuration</h2>
      <h3>SQLite (default, single-node)</h3>
      <pre><code>{`-driver sqlite -dsn "file:godnslog.db"`}</code></pre>

      <h3>MySQL (production)</h3>
      <pre><code>{`-driver mysql -dsn "user:password@tcp(127.0.0.1:3306)/godnslog?charset=utf8mb4"`}</code></pre>

      <h2>Docker Compose</h2>
      <p>Use <code>docker-compose.yml</code> for a full stack with MySQL and Redis:</p>
      <pre><code>{`docker-compose up -d`}</code></pre>
      <p>For HA deployment, use <code>docker-compose.ha.yml</code>.</p>

      <h2>Environment Variables</h2>
      <p>The Docker entrypoint supports these environment variables:</p>
      <ul>
        <li><code>GODNSLOG_DOMAIN</code> — Your domain</li>
        <li><code>GODNSLOG_IP4</code> — IPv4 address</li>
        <li><code>GODNSLOG_DSN</code> — Database DSN</li>
        <li><code>GODNSLOG_DRIVER</code> — Database driver</li>
        <li><code>GODNSLOG_REDIS_ADDR</code> — Redis address (HA)</li>
        <li><code>GODNSLOG_HTTP_LISTEN</code> — HTTP listen address</li>
      </ul>

      <h2>DNS Configuration</h2>
      <p>
        Configure your domain&apos;s NS records to point to the GODNSLOG server.
        The server listens on port 53 for DNS queries and supports:
      </p>
      <ul>
        <li>A, AAAA, CNAME, TXT, MX, NS record types</li>
        <li>Wildcard subdomain resolution</li>
        <li>Upstream forwarding for non-matching queries</li>
        <li>xip-style IP encoding (dot-decimal, hex, binary)</li>
      </ul>

      <h2>Notification Channels</h2>
      <p>Configure notification channels via the Settings page or API:</p>
      <ul>
        <li><strong>Webhook</strong>: Generic HTTP webhook with custom headers and body template</li>
        <li><strong>Feishu</strong>: Feishu/Lark bot webhook</li>
        <li><strong>WeCom</strong>: WeChat Work bot webhook</li>
        <li><strong>DingTalk</strong>: DingTalk bot webhook</li>
        <li><strong>Slack</strong>: Slack incoming webhook</li>
        <li><strong>Discord</strong>: Discord webhook</li>
        <li><strong>Telegram</strong>: Telegram bot with chat ID</li>
        <li><strong>Email</strong>: SMTP with TLS support</li>
      </ul>

      <h2>System Settings</h2>
      <p>Manage system-wide settings via <code>/api/v2/settings</code>:</p>
      <ul>
        <li><code>main_domain</code> — Primary domain</li>
        <li><code>dns_domain</code> — DNS subdomain</li>
        <li><code>http_domain</code> — HTTP subdomain</li>
        <li>Notification configurations</li>
      </ul>

      <h2>High Availability</h2>
      <p>
        For HA deployment, configure Redis for session sharing and use the cluster API
        (<code>/api/v2/cluster</code>) to manage nodes. See <code>deploy/k8s/godnslog-ha.yaml</code>
        for Kubernetes deployment examples.
      </p>
    </DocPageLayout>
  )
}
