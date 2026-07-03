# EP / ARD — RFE-8685

**Source:** `openspec/changes/rfe-8685/inputs/jira-spec.md` (scope-locked)

## Summary

Enable External Secrets Operator to trust Enterprise PKI via ConfigMap CA injection, resolving `x509: unknown authority` errors for internal services. Replaces unsupported manual volume mounts.

## Functional requirements (from Jira)

1. Extend operator API with CA reference for server verification on ESC `appConfig`.
2. Reference a ConfigMap containing trusted CA certificates for controller/webhook operand workloads.
3. Enable connectivity to internal vaults (IBM/Thycotic) signed by enterprise PKI.
4. Avoid ad-hoc manual volume mounts — declarative operator configuration only.

## Affected components

- External Secrets Operator controller + webhook deployments
- `ExternalSecretsConfig` operator CRD (greenfield field)
- Store-level `caProvider` on upstream SecretStore/ClusterSecretStore (docs/config only — already in operand CRD v0.20.4)
- OpenShift platform trusted CA injection (proxy-gated CNO path — preserve behavior)

## Non-goals

- Upstream `external-secrets.io` CRD schema changes
- Bindata operand CRD refresh
- cert-controller / bitwarden enterprise CA mounts
- ESM globalConfig parity (deferred)
