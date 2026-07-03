# Root Cause Analysis — RFE-8685

**Round:** 1  
**Bugs in bundle:** none (`evals/inputs/bugs/index.yaml` empty)

No formal bug tickets were supplied. Analysis derives **workflow friction** and **design gaps** from implementation evidence in `evals/inputs/05-repo-prs.md` and OpenSpec task reports.

---

## Workflow friction items (treated as pre-bug learnings)

### WFR-001 — API tests before CRD regeneration

| Field | Detail |
|-------|--------|
| Symptom | T1_2 `make test-apis` failed: `unknown field spec.appConfig.additionalTrustedCAConfigMapRef` |
| Fix | Ran `make manifests` (partial T2_1 overlap), then tests passed |
| Root cause category | **story_formation** |
| Type | non_functional (CI/agent workflow) |
| Workflow stage | **tasks** |
| Would eval catch? | New tasks eval: require manifests before test-apis for new API fields |

### WFR-002 — Testsuite expectedError vs CEL multi-error format

| Field | Detail |
|-------|--------|
| Symptom | Negative API tests failed after CRD regen — expected single-field error, got bracket list |
| Fix | Updated `expectedError` to include CEL follow-up message |
| Root cause category | **coding** |
| Type | non_functional |
| Workflow stage | **implementation** / **code-generation** (api-generate-tests) |
| Would eval catch? | Code-gen eval for testsuite: match repo pattern for CEL errors |

### WFR-003 — Cross-namespace ConfigMap reference

| Field | Detail |
|-------|--------|
| Symptom | API allows `namespace` on ref; Kubernetes volumes cannot mount cross-ns ConfigMaps |
| Fix | Sync user CM data to `external-secrets-additional-trusted-ca` in operand namespace |
| Root cause category | **design** |
| Type | functional |
| Workflow stage | **plan**, **repo-assessment** |
| Would eval catch? | Plan eval: must document mount strategy for cross-ns refs |

### WFR-004 — Shared CommonConfigs embeds field on ESM CRD

| Field | Detail |
|-------|--------|
| Symptom | `make bundle` updated ESM CRD/bundle with enterprise CA field |
| Fix | Accepted — schema sharing is correct; field only used on ESC at runtime |
| Root cause category | **design** |
| Type | non_functional |
| Workflow stage | **repo-assessment** |
| Would eval catch? | Repo-assessment: note CommonConfigs consumers |

---

## Code approach (from implementation)

| Approach | Outcome |
|----------|---------|
| Refactor `updateTrustedCABundleVolumes` → `updateOperandTrustedCAVolumes` with A-011 matrix | Correct; unit tests added |
| `ConfigurationError` for missing CM + Degraded in `processReconcileRequest` | Meets FR-005/FR-009 |
| Projected volume for proxy+enterprise | Matches plan A-011 |
| Gate enterprise mounts with `isOperandTrustedCAMountSupported(assetName)` | Meets FR-010 scope |
| Uncached client for user CM fetch | Correct for cross-ns read |

**Functional issues avoided:** mounting enterprise CA on cert-controller; writing to user CM; SSA-first volumes.

**Non-functional:** baseline `govulncheck` failure on release-1.1 unrelated to feature.

---

## Bug RCA (none)

No entries in `bugs/index.yaml`. Formal bug RCA deferred until next bundle includes bug files.
