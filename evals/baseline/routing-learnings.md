# Routing Learnings — Round 1 (RFE-8685)

## Agent routing

| Task type | Agent | Notes |
|-----------|-------|-------|
| ESC API + testsuite | API_Agent | Run manifests before test-apis for new fields |
| Volume reconcile / A-011 | OperatorController_Agent | Gate mounts by deployment asset name |
| Store caProvider docs | Docs_Agent | No bindata/CRD tasks |
| Code-gen eval gate | Scores by `oape_command` filter | manual for manifests-only subset |

## Guardrails learned

1. Do not expand to upstream `external-secrets.io` CRD or bindata for enterprise CA on ESC.
2. Cross-namespace ConfigMap refs require operand-namespace sync for volume mounts.
3. `CommonConfigs` changes may regenerate ESM + ESC CRDs — expected, not a bug.
4. CEL validation negative tests use bracket multi-error format in this repo.
5. `ConfigurationError` → Degraded=True with requeue for missing user ConfigMaps.

## Makefile verification pairing

| Task | Verify |
|------|--------|
| T1_1 API types | go build/vet; generate for compile |
| T1_2 API tests | make test-apis (requires manifests) |
| T2_1 codegen | make generate manifests docs bundle |
| T3_1 controller | go test ./pkg/controller/external_secrets/... |
| T4_2 unit gate | make test-unit |
