# Security Policy

## Supported versions

Only the latest release published to the
[Terraform Registry](https://registry.terraform.io/providers/amrutp24/fireworks)
receives security fixes.

## Reporting a vulnerability

**Please do not open a public issue for security vulnerabilities.**

Report privately via
[GitHub Security Advisories](https://github.com/amrutp24/terraform-provider-fireworks/security/advisories/new)
("Report a vulnerability"). You can expect an acknowledgement within 72
hours and a fix or mitigation plan within 14 days for confirmed issues.

## Scope

In scope:

- API key or account credentials leaking into state, logs, plan output, or
  error messages beyond Terraform's normal sensitive-value handling
- Requests sent to any host other than the configured `base_url`
- TLS or certificate-validation weaknesses in the API client

Out of scope:

- Vulnerabilities in the Fireworks AI service itself (report to Fireworks)
- Issues requiring a malicious `base_url` set by the practitioner themselves
