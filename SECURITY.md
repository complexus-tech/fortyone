# Internal security process

FortyOne is a managed cloud service. This document defines the repository's
internal security expectations; it is not a public vulnerability-disclosure or
version-support policy.

## Report immediately

Escalate suspected vulnerabilities, leaked credentials, unauthorized access,
customer-data exposure, and suspicious production activity through the private
company incident channel and `security@fortyone.app`.

Include only the information needed to begin triage:

- affected service, workspace, integration, or deployment;
- first observed time and whether the issue is still active;
- a concise reproduction or sequence of events;
- likely impact and affected data classes;
- relevant request, trace, deployment, or audit identifiers.

Do not include secrets, bearer URLs, raw access tokens, private keys, complete
customer records, or unnecessary personal data. Share sensitive evidence only
through the approved encrypted channel.

## Initial response

The incident owner should:

1. Preserve relevant audit, deployment, and application evidence.
2. Contain active access without destroying evidence.
3. Rotate or revoke exposed credentials and integration grants.
4. Identify affected tenants and verify tenant isolation.
5. Record decisions, owners, and timestamps in the private incident record.
6. Coordinate customer or regulatory communication with company leadership.

Never announce an incident publicly, contact customers, or disclose technical
details without the incident owner's approval.

## Engineering requirements

- Keep credentials in the managed secret store; never commit them to Git.
- Use parameterized, tenant-scoped database access and explicit authorization.
- Verify webhook signatures, timestamps, replay protection, and idempotency.
- Redact secrets and sensitive payloads from logs and traces.
- Apply least-privilege permissions to cloud roles, API keys, and integrations.
- Add regression tests for every confirmed security defect.
- Treat generated artifacts, fixtures, and screenshots as potential data leaks.
- Review dependency and container findings before production release.

## Production changes

Security-sensitive changes require review proportional to their impact. Changes
to authentication, authorization, tenant boundaries, cryptography, credentials,
webhooks, billing, data export, or production infrastructure must include an
explicit threat and rollback assessment.

Production releases use the reviewed CI/CD path. Do not publish source archives,
installers, environment templates, or application images as public artifacts.
