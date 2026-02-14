[![CI](https://github.com/0xsj/firewatch/actions/workflows/ci.yml/badge.svg)](https://github.com/0xsj/firewatch/actions/workflows/ci.yml)

# Firewatch

Honeypot server that masquerades as vulnerable web applications to detect, fingerprint, and analyze attackers. Deploys fake endpoints across 7 modules, captures scanner behavior, and generates exportable threat intelligence.

## Features

- **7 Honeypot Modules** — Next.js, WordPress, API, Exposure, Admin Panels, Cloud Metadata, CVE Emulation
- **CVE Emulation** — Log4Shell, Spring4Shell, MOVEit, PAN-OS, Struts2, Confluence
- **Request Fingerprinting** — JA3/JA4 TLS fingerprints, header ordering analysis, anomaly detection
- **Detection Engine** — Signature and pattern matching with severity ranking, custom YAML signatures
- **IP Filtering** — Allowlist/blocklist with CIDR support, file-based lists
- **Rate Limiting** — Per-IP token bucket rate limiting to detect and slow down scanners
- **Attacker Profiling** — Automatic per-IP attacker records with severity escalation and auto-tagging
- **Behavioral Analysis** — Per-IP temporal patterns: scan sweeps, brute force, module hopping, recon escalation
- **Campaign Correlation** — Background detection of coordinated multi-IP attacks via signature clustering
- **Threat Intelligence** — IOC extraction, campaign correlation, STIX/MISP/CSV export
- **Real-time Alerts** — Slack, Discord, generic webhook with per-channel severity thresholds
- **Deception Responses** — Realistic fake login pages, API responses, config files, and honey tokens

## Requirements

- Go 1.21+
- SQLite (bundled via `modernc.org/sqlite`, no CGO required)
- MaxMind GeoIP database (optional, for geolocation enrichment)

## Installation

Build from source:

```bash
git clone https://github.com/0xsj/firewatch
cd firewatch
make build
```

Or directly with Go:

```bash
go build -o firewatch ./cmd/firewatch
```

## Quick Start

```bash
# Start with default config (creates firewatch.yaml if missing)
./firewatch

# Start with a specific config file
./firewatch -config /path/to/firewatch.yaml

# Print version
./firewatch -version
```

The server reads `firewatch.yaml` for all configuration. Enable or disable modules, configure alerts, and tune fingerprinting from that file.

## Configuration

```yaml
server:
  domain: "localhost"
  port: 8080
  tls:
    enabled: false
    # cert: "/etc/firewatch/tls/cert.pem"
    # key:  "/etc/firewatch/tls/key.pem"

rate_limit:
  enabled: true
  requests_per_second: 10  # Sustained rate (600/min)
  burst: 20                # Burst allowance
  cleanup_minutes: 5       # Cleanup interval

modules:
  nextjs:
    enabled: true
    endpoints: ["/", "/_next/server/pages", "/_rsc"]
  wordpress:
    enabled: true
    fake_version: "6.4.2"
  exposure:
    enabled: true
  api:
    enabled: true
  cloud:
    enabled: true
  admin:
    enabled: false
  cve:
    enabled: false
    # cves:                    # Empty = all 6 CVEs enabled
    #   - "CVE-2021-44228"    # Log4Shell
    #   - "CVE-2022-22965"    # Spring4Shell
    #   - "CVE-2023-34362"    # MOVEit Transfer
    #   - "CVE-2024-3400"     # PAN-OS GlobalProtect
    #   - "CVE-2017-5638"     # Apache Struts2
    #   - "CVE-2023-22515"    # Confluence

alerts:
  slack:
    # webhook_url: "https://hooks.slack.com/services/T00/B00/xxx"
    min_severity: "medium"
  discord:
    # webhook_url: "https://discord.com/api/webhooks/000/xxx"
    min_severity: "medium"
  webhook:
    # url: "https://your-siem.example.com/api/ingest"
    # headers:
    #   Authorization: "Bearer token"
    min_severity: "medium"

fingerprinting:
  ja3: true
  ja4: true
  geoip: false
  reverse_dns: false

detection:
  # signatures_file: "/etc/firewatch/custom-signatures.yaml"
  # signatures_dir: "/etc/firewatch/signatures.d/"
  behavior:
    enabled: false
    window_minutes: 5
  campaign:
    enabled: false
    window_minutes: 30
    tick_seconds: 60

storage:
  type: "sqlite"
  path: "./firewatch.db"

deception:
  honey_tokens: true
  breadcrumbs: true
  fake_errors: true

logging:
  level: "info"     # debug, info, warn, error
  format: "json"    # json, text
```

## Honeypot Modules

| Module     | Endpoints                                          | Detects                                  |
|------------|----------------------------------------------------|------------------------------------------|
| `nextjs`   | `/_next/*`, RSC headers, server actions            | Next.js scanners, RSC probes             |
| `wordpress`| `/wp-login.php`, `/wp-admin/`, `/xmlrpc.php`       | WordPress scanners, brute force, XML-RPC |
| `api`      | `/api/*`, `/graphql`, `/swagger/`                  | API enumeration, auth probes, GraphQL    |
| `exposure` | `/.env`, `/.git/`, config files                    | Sensitive file scanners                  |
| `admin`    | `/phpmyadmin/`, `/adminer.php`, `/cpanel`, `/admin`| Admin panel scanners, brute force        |
| `cloud`    | `/latest/meta-data/`, `/metadata/v1/`              | SSRF, cloud credential theft             |
| `cve`      | Solr, Actuator, MOVEit, GlobalProtect, Struts, Confluence | CVE exploit scanners              |

## Deployment

### Docker

```bash
docker build -t firewatch .
docker run -p 8080:8080 -v ./firewatch.yaml:/etc/firewatch/firewatch.yaml firewatch
```

### Binary

```bash
make build
./firewatch -config firewatch.yaml
```

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for system design and component diagrams.

See [docs/TREE.md](docs/TREE.md) for the full project structure.

## License

MIT
