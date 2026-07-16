# User stories — RFE-8685

From `openspec/changes/rfe-8685/specs.md`

## US-1 (P1) — Trust internal vault from operator workloads

Administrator configures enterprise CA ConfigMap ref on ESC; controller/webhook trust enterprise PKI; sync succeeds without x509 errors.

## US-2 (P2) — Declare provider CA at store level

Configure upstream `caProvider` on SecretStore/ClusterSecretStore for provider TLS (no operator CRD change).

## US-3 (P2) — Manage CA trust through supported configuration

Add/update/remove enterprise CA via declarative API; no manual volume mounts; reconciliation reverts on removal.

## US-4 (P3) — Coexist with platform proxy CA (A-011)

When proxy + enterprise both enabled: projected volume merging CNO platform CA + enterprise CA at `/etc/pki/tls/certs`.

## Edge cases (spec)

- Missing ConfigMap → degraded status, clear message
- Invalid/empty PEM → validation error
- ConfigMap deleted after active → degraded on reconcile
- Migration from manual mounts → operator converges on supported config
