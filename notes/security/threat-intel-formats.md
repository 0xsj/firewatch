# Threat Intelligence Formats

## Overview

Threat intelligence sharing uses standardized formats so different tools and organizations can exchange IOCs and campaign data. Firewatch exports in three formats.

## STIX 2.1 (Structured Threat Information Expression)

**Standard**: OASIS open standard for cyber threat intelligence.

### Key concepts

| Concept    | Description                                    |
|------------|------------------------------------------------|
| **Bundle** | Container for STIX objects                     |
| **SDO**    | STIX Domain Object (Indicator, Campaign, etc.) |
| **SRO**    | STIX Relationship Object (links SDOs)          |
| **SCO**    | STIX Cyber-observable Object (IP, file, etc.)  |

### Bundle structure

```json
{
  "type": "bundle",
  "id": "bundle--<uuid>",
  "objects": [
    { "type": "indicator", ... },
    { "type": "campaign", ... }
  ]
}
```

### STIX Pattern syntax

Indicators use a pattern language to express what to look for:

```
[ipv4-addr:value = '192.168.1.1']
[domain-name:value = 'evil.com']
[url:value = 'http://evil.com/shell.php']
[file:hashes.SHA-256 = 'abc123...']
[email-addr:value = 'attacker@evil.com']
```

The pattern type is always `"stix"` — other types like `"sigma"` or `"snort"` are also valid.

### IDs are namespaced

```
indicator--<uuid>
campaign--<uuid>
bundle--<uuid>
```

The prefix identifies the object type. UUIDs ensure global uniqueness.

## MISP (Malware Information Sharing Platform)

**Standard**: De facto standard for threat intelligence sharing among CERTs and SOCs.

### Key concepts

| Concept      | Description                              |
|--------------|------------------------------------------|
| **Event**    | Container for related attributes         |
| **Attribute**| Single IOC (IP, domain, hash, etc.)      |
| **Galaxy**   | Taxonomy cluster (threat actor, tool)    |
| **Tag**      | Free-form classification label           |

### Event structure

```json
{
  "info": "Firewatch honeypot IOCs - 2024-01-15",
  "threat_level_id": "1",
  "date": "2024-01-15",
  "published": false,
  "Attribute": [
    {
      "uuid": "<uuid>",
      "type": "ip-src",
      "category": "Network activity",
      "value": "192.168.1.1",
      "to_ids": true,
      "comment": "Source: firewatch"
    }
  ]
}
```

### Threat levels

| Level | Meaning   | Maps from            |
|-------|-----------|----------------------|
| 1     | High      | critical, high       |
| 2     | Medium    | medium               |
| 3     | Low       | low                  |
| 4     | Undefined | info                 |

### Attribute types

| IOC Type | MISP Type  | Category          |
|----------|------------|-------------------|
| IP       | `ip-src`   | Network activity  |
| Domain   | `domain`   | Network activity  |
| URL      | `url`      | Network activity  |
| Hash     | `sha256`   | Payload delivery  |
| Email    | `email-src`| Payload delivery  |

### to_ids flag

`"to_ids": true` means the attribute should trigger IDS rules. Only set for high/critical severity IOCs — low severity generates too much noise.

## CSV

Simplest format — universal import compatibility.

### IOC CSV columns

```
id,type,value,severity,first_seen,last_seen,hostname,country,asn,tags
```

### Design choices

- **Semicolons for lists**: Tags use `;` as delimiter within the CSV field since `,` is the field separator
- **GeoIP as separate columns**: Flattened rather than nested for spreadsheet compatibility
- **ASN format**: Prefixed with `AS` (e.g., `AS13335`) per convention

## Which format to use when

| Use case                           | Format  |
|------------------------------------|---------|
| Share with SIEM/SOAR platform      | STIX    |
| Share with other CERTs/SOCs        | MISP    |
| Import into spreadsheet/dashboard  | CSV     |
| Automated threat feed              | STIX    |
| Quick manual review                | CSV     |
| Integrate with MISP instance       | MISP    |
