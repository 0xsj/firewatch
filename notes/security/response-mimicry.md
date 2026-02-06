# Response Mimicry in Honeypots

## Principle

A honeypot must return responses that are convincing enough that attackers continue interacting. Too fake and they move on. Too real and you're hosting actual vulnerabilities.

## Techniques Used in Firewatch

### 1. Realistic Headers

Set headers that match the technology being emulated:

```go
// WordPress — PHP powered
w.Header().Set("X-Powered-By", "PHP/8.1.0")
w.Header().Set("Content-Type", "text/html; charset=utf-8")

// AWS IMDSv2 — metadata token TTL
w.Header().Set("X-Aws-Ec2-Metadata-Token-Ttl-Seconds", "21600")
```

### 2. Correct Content Types

Match what the real service returns:
- WordPress pages: `text/html; charset=utf-8`
- XML-RPC: `text/xml; charset=utf-8`
- AWS metadata: `text/plain`
- REST APIs: `application/json`
- OpenAPI spec: `application/json`

### 3. Realistic Response Bodies

**WordPress login** — Render an actual login form with version-specific markup so automated scanners see expected DOM structure.

**XML-RPC** — Return valid XML listing methods (`system.multicall`, `wp.getUsersBlogs`) that attackers expect when probing for XML-RPC abuse.

**AWS metadata** — Return the actual IMDS directory listing (`ami-id`, `instance-type`, etc.) so SSRF payloads see expected responses and continue.

**IAM credentials** — Return clearly-fake-but-structured credentials:
```go
"AccessKeyId":     "AKIAIOSFODNN7EXAMPLE",
"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
"Token":           "FwoGZXIvYXdzEBAaDHoney+Token+Do+Not+Use",
```
These are honey tokens — if anyone uses them against real AWS, it triggers alerts on the AWS side too.

### 4. Correct Status Codes

| Scenario                | Code | Rationale                      |
|-------------------------|------|--------------------------------|
| Login page (GET)        | 200  | Page exists and renders        |
| Failed login (POST)     | 200  | WordPress returns 200 on failure |
| wp-admin (no auth)      | 302  | Redirect to login              |
| Config files            | 403  | Confirms existence             |
| API without auth        | 401  | Triggers auth probing          |
| GraphQL without query   | 200  | GraphQL returns errors in body |

### 5. Graduated Response

Don't reveal everything immediately. The exposure module:
- `.git/` directory → 403 (confirms existence)
- `.git/config` → Full config (with fake remote URL)
- `.git/HEAD` → Branch reference

This mimics real directory traversal where deeper paths reveal more.

### 6. Configurable Fake Content

```go
content := e.cfg.FakeEnv   // Operator-provided fake .env
if content == "" {
    content = deception.ExposedEnvFile()  // Default fake .env
}
```

Allow operators to customize honey tokens — different deployments can use different fake credentials, making it possible to trace which honeypot instance was compromised.

## What NOT to Do

- **Don't return perfect replicas** — You'll accidentally create a real vulnerability
- **Don't return empty responses** — Scanners will mark as dead and skip
- **Don't return generic 404s** — Defeats the purpose of the honeypot
- **Don't use real credentials** — Even "test" credentials. Use obviously fake but structurally valid ones
- **Don't vary responses randomly** — Inconsistency helps attackers identify honeypots
