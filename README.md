# Firewatch

Honeypot server for detecting and analyzing web application scanners. Deploys fake vulnerable endpoints, captures attacker behavior, and generates threat intelligence.

## Features

- **Multi-Framework Honeypots** — Next.js, WordPress, APIs, cloud metadata
- **CVE Emulation** — Fake vulnerable endpoints for specific CVEs
- **Request Fingerprinting** — JA3/JA4, User-Agent, header analysis
- **Threat Intelligence** — IOC extraction, attacker profiling, campaign detection
- **Deception Tokens** — Embedded honey tokens that track further access
- **Real-time Alerts** — Slack, Discord, webhook, email notifications
- **Rich Dashboard** — Visualize attacks, patterns, and attacker profiles

## Installation

```bash
go install github.com/yourusername/firewatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/firewatch
cd firewatch
make build
```

## Quick Start

```bash
# Start with default honeypots
firewatch serve --domain honeypot.example.com

# Enable specific honeypot modules
firewatch serve --modules nextjs,wordpress,api,exposure

# Start with alerting
firewatch serve --alert-slack https://hooks.slack.com/xxx

# View captured events
firewatch events --last 1h

# Export threat intelligence
firewatch export --format stix -o iocs.json
```

## Usage

```bash
firewatch <command> [options]

Commands:
  serve       Start the honeypot server
  events      View captured events
  attackers   View attacker profiles
  campaigns   View detected campaigns
  export      Export threat intelligence
  stats       Show statistics

Serve Options:
  -d, --domain <domain>     Domain name for the honeypot
  -p, --port <port>         Listen port (default: 8080)
  --tls                     Enable TLS
  --cert <file>             TLS certificate
  --key <file>              TLS private key
  --modules <list>          Honeypot modules to enable
  --alert-slack <url>       Slack webhook for alerts
  --alert-discord <url>     Discord webhook for alerts
  --alert-webhook <url>     Generic webhook for alerts
  --db <path>               Database path (default: ./firewatch.db)
  -v, --verbose             Verbose logging

Events Options:
  --last <duration>         Show events from last duration
  --attacker <ip>           Filter by attacker IP
  --module <name>           Filter by honeypot module
  --severity <level>        Filter by severity
  -n, --limit <n>           Limit results

Export Options:
  --format <fmt>            Export format: stix, csv, json, misp
  --last <duration>         Export events from last duration
  -o, --output <file>       Output file
```

## Honeypot Modules

| Module      | Endpoints                                   | Detects                             |
| ----------- | ------------------------------------------- | ----------------------------------- |
| `nextjs`    | `/_next/*`, `next-action` header            | CVE-2025-55182 scanners, RSC probes |
| `wordpress` | `/wp-admin`, `/wp-login.php`, `/xmlrpc.php` | WordPress scanners, brute force     |
| `api`       | `/api/*`, `/graphql`, `/swagger`            | API enumeration, auth probes        |
| `exposure`  | `/.env`, `/.git`, `/config.php`             | Sensitive file scanners             |
| `admin`     | `/admin`, `/phpmyadmin`, `/adminer`         | Admin panel scanners                |
| `cloud`     | `/latest/meta-data`                         | SSRF, cloud metadata attacks        |
| `cve`       | Various CVE-specific endpoints              | Exploit scanners                    |

## Example Output

```
$ firewatch serve --modules nextjs,wordpress,exposure

  🔥 Firewatch v1.0.0
  ─────────────────────────────────────────────────────
  Honeypot:  honeypot.example.com:8080
  Modules:   nextjs, wordpress, exposure
  Database:  ./firewatch.db
  ─────────────────────────────────────────────────────

  [14:32:01] [nextjs] POST / next-action:"x"
            IP: 45.33.32.156 (Linode, US)
            JA3: e7d705a3286e19ea42f587b344ee6865
            → Known scanner: next-action-probe-v1

  [14:32:15] [wordpress] POST /wp-login.php
            IP: 192.168.1.50 (Internal)
            Credentials: admin:password123
            → Brute force attempt (47th try)

  [14:33:02] [exposure] GET /.env
            IP: 103.21.244.0 (Cloudflare)
            → Sensitive file probe

  [14:33:45] [ALERT] Campaign detected
            IPs: 45.33.32.0/24 (156 hosts)
            Pattern: Coordinated Next.js scanning
            First seen: 2 hours ago
```

## Configuration

```yaml
# ~/.firewatch/config.yaml
server:
  domain: honeypot.example.com
  port: 8080
  tls:
    enabled: false
    cert: /path/to/cert.pem
    key: /path/to/key.pem

modules:
  nextjs:
    enabled: true
    endpoints:
      - "/"
      - "/_next/server/pages"
      - "/_rsc"
  wordpress:
    enabled: true
    fake_version: "6.4.2"
  exposure:
    enabled: true
    fake_env: |
      DB_HOST=localhost
      DB_PASS=fake_password_123
  api:
    enabled: true
  cloud:
    enabled: false

fingerprinting:
  ja3: true
  ja4: true
  geoip: true
  reverse_dns: true

alerts:
  slack:
    webhook_url: https://hooks.slack.com/xxx
    min_severity: medium
  discord:
    webhook_url: https://discord.com/api/webhooks/xxx
  webhook:
    url: https://siem.example.com/api/alerts

storage:
  type: sqlite
  path: ./firewatch.db

deception:
  honey_tokens: true
  breadcrumbs: true
  fake_errors: true
```

## Deployment

```bash
# Docker
docker run -p 8080:8080 yourusername/firewatch

# Docker Compose
docker-compose up -d

# Kubernetes
kubectl apply -f deployments/kubernetes/

# Alongside real applications
# Run on a subdomain or separate port to catch scanners
```

## Requirements

- Go 1.21+
- MaxMind GeoIP database (optional, for geolocation)

## License

MIT
