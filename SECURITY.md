# Security Policy

## Supported versions

agentreg is pre-1.0. Security fixes land on the latest release only.

| Version | Supported |
|---------|-----------|
| 0.1.x   | ✅        |
| < 0.1   | ❌        |

## Reporting a vulnerability

Please **do not open a public issue** for security problems.

Report privately using GitHub's [Report a vulnerability](https://github.com/mkk2026/agentreg/security/advisories/new)
(Security → Advisories), or email **info@corebrimtech.com**.

Include what you found, how to reproduce it, and the impact. You'll get an
acknowledgement within a few days. Once a fix is out, we're happy to credit you
unless you'd prefer to stay anonymous.

## Scope notes

agentreg is a self-hosted registry. In v0.x it has **no authentication** on its
HTTP API by design — run it on a trusted network or behind your own access
controls, not exposed to the public internet. Authentication and trust
verification are on the [roadmap](README.md#roadmap).
