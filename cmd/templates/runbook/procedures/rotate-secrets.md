---
title: Rotate Secrets
tags: [security, secrets, credentials]
status: draft
last-reviewed: 2026-01-01
---

# Rotate Secrets

Rotate credentials, API keys, or certificates when they expire, are
compromised, or as part of regular rotation policy.

## When to Use

- A credential has been exposed (leaked in logs, committed to repo)
- Scheduled rotation (quarterly, annually)
- An employee with access leaves the team
- Vendor requires key rotation

## Prerequisites

- [ ] Access to secrets manager (e.g., AWS Secrets Manager, Vault, 1Password)
- [ ] Permission to update secrets in production
- [ ] Deployment pipeline access (secrets may require a redeploy)

## Steps

### 1. Generate New Credential

```bash
# Example: generate a new API key
openssl rand -base64 32

# Example: generate a new JWT secret
openssl rand -hex 64
```

### 2. Update Secrets Manager

```bash
# Example: AWS Secrets Manager
aws secretsmanager update-secret \
  --secret-id "my-service/api-key" \
  --secret-string '{"API_KEY":"<new-value>"}'
```

### 3. Deploy / Restart Services

```bash
# Services need to pick up the new secret
# Option A: Redeploy
# Option B: Restart
# Option C: Service reads secrets dynamically (no action needed)
```

### 4. Verify

- [ ] Service starts successfully with new credential
- [ ] API calls using old credential are rejected (if applicable)
- [ ] Health checks pass

### 5. Revoke Old Credential

```bash
# Only after verifying the new one works
# Revoke/delete the old key in the secrets manager or vendor dashboard
```

## Rollback

If the new credential doesn't work:

1. Re-set the old credential in secrets manager
2. Restart/redeploy the service
3. Investigate why the new credential failed before trying again

## Related

- [[procedures/deploy-rollback]] — if rotation requires a deploy
