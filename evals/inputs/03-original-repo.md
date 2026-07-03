# Pre-feature repo state — external-secrets-operator

**Repo:** https://github.com/openshift/external-secrets-operator  
**Branch:** release-1.1  
**Commit:** 1f54b74a6fbe2c327c9aa2d1f4f7662cce1871ea

## Relevant pre-feature behavior

| Area | State before RFE-8685 |
|------|------------------------|
| ESC API | No `additionalTrustedCAConfigMapRef`; `CommonConfigs` has `proxy`, `logLevel`, etc. |
| Trusted CA | `external-secrets-trusted-ca-bundle` ConfigMap created only when proxy configured |
| Volume mount | `updateTrustedCABundleVolumes` proxy-gated; mounts to all deployments passing through `updateProxyConfiguration` |
| Mount path | `/etc/pki/tls/certs` |
| Upstream stores | `caProvider` / `caBundle` already on SecretStore/ClusterSecretStore CRD v0.20.4 |
| E2E | `Feature:TrustedCABundle` label documented; no test file |

## Key files

- `pkg/controller/external_secrets/configmap.go` — CNO inject label ConfigMap
- `pkg/controller/external_secrets/deployments.go` — proxy + trusted CA volumes
- `api/v1alpha1/meta.go` — `CommonConfigs`
