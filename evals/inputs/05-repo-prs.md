# Implementation evidence — RFE-8685 (working-folder; no GitHub PR)

**Change:** `openspec/changes/rfe-8685`  
**Mode:** Working-folder at release-1.1 HEAD (no remote PR)

## Summary of code changes (T1_1–T4_2)

### API (`api-generate`)

- Added `AdditionalTrustedCAConfigMapRef` + field on `CommonConfigs` in `api/v1alpha1/meta.go`
- Testsuite cases in `externalsecretsconfig.testsuite.yaml` (valid/invalid; CEL multi-error format)

### Codegen (`manual` / api-generate follow-up)

- `make generate`, `manifests`, `docs`, `bundle` — CRD, deepcopy, api_reference, OLM bundle
- ESM CRD also updated (shared `CommonConfigs` embedding)

### Controller (`api-implement`)

- `ensureEnterpriseTrustedCAConfigMap` — sync user CM to `external-secrets-additional-trusted-ca` in operand namespace (cross-ns mount workaround)
- `updateOperandTrustedCAVolumes` — A-011 matrix: none / platform / enterprise / projected
- `ConfigurationError` + Degraded status for missing enterprise CM
- Unit tests: `configmap_test.go`, `deployments_enterprise_ca_test.go`

## Implementation learnings (retrospective)

1. **T1_2 blocked** until CRD manifests regenerated (`unknown field` in envtest).
2. **Testsuite `expectedError`** must match CEL bracket format when multiple validation errors fire.
3. **Cross-namespace ConfigMap ref** cannot mount directly — sync to operand-namespace CM required.
4. **Enterprise mounts** limited to controller + webhook assets only (not cert-controller/bitwarden).
5. **T2_2** `make verify` failed on baseline `govulncheck` (unrelated to feature).

## Task reports

`openspec/changes/rfe-8685/implementation/task-reports/T1_1.md` through `T4_2.md`
