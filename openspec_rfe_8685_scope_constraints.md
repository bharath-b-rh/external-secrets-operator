# RFE-8685 scope constraints
_Exported on 7/3/2026 at 13:13:18 GMT+5:30 from Cursor (3.8.11)_

---

**User**

/opsx-continue RFE-8685

Scope constraints (mandatory for this change):
- Read and honor openspec/changes/RFE-8685/inputs/scope-constraints.md
- Jira: use ONLY the text in openspec/changes/RFE-8685/inputs/jira-spec.md
- Do NOT use Jira MCP, linked issues, subtasks, epic, comments, or attachments
- Repo: use ONLY this working directory at the current HEAD (release-1.1)
- Do NOT git fetch, checkout main, compare to other branches, or browse GitHub
- Do NOT clone or read remote repos
- Working-folder mode: use this as the working directory

---

**Cursor**

Continuing the RFE-8685 OpenSpec change: I'll read the workflow skill and check which artifact is ready next.


Renaming the change folder to kebab-case so the OpenSpec CLI works, then creating the first artifact (`validation`).


Creating `validation.json` from the Jira spec and AGENTS.md validation hints, then writing the eval results and evaluation report.


## Artifact created: `validation`

**Change:** `rfe-8685` (folder renamed from `RFE-8685` → `rfe-8685` so the OpenSpec CLI accepts kebab-case)  
**Output:** `openspec/changes/rfe-8685/validation.json`  
**Eval results:** `openspec/changes/rfe-8685/eval-results/validation.yaml`  
**Evaluation report:** `openspec/changes/rfe-8685/validation_evaluation_report.md`  
**Progress:** 1/9 artifacts (pending your approval)

### Validation summary

Evaluated **only** `inputs/jira-spec.md` per scope constraints. Working-folder mode honored (`use_working_folder_as_repo: true`).

| Metric | Value |
|--------|-------|
| **Overall score** | **56%** (0/2 rubric dimensions pass threshold) |
| **Status** | **NEEDS_REVISION** (no blockers) |
| Completeness | 58/100 |
| Quality | 52/100 |

**Strengths in the spec**
- Clear motivation: enterprise PKI / `x509: unknown authority` blocking secret sync
- Security intent: avoid unsupported manual volume mounts
- Affected components identified (controller, webhook, CRDs, OCP CA injection)

**Top gaps (non-blocking)**
1. No testable Given/When/Then acceptance criteria
2. ConfigMap reference contract undefined (name, namespace, key, mount targets)
3. Unclear operator API (`ExternalSecretsConfig`) vs upstream `SecretStore`/`ClusterSecretStore` ownership
4. Interaction between proxy-triggered OCP CA injection and custom enterprise CA not defined
5. Missing operational semantics (ConfigMap delete/update, RBAC, upgrade path)

Per workflow rules, **NEEDS_REVISION with no blockers** can proceed to `specs.md` after approval; gaps will be captured as assumptions or `[NEEDS CLARIFICATION]` markers there.

---

Eval score: **56%** (0/2 rubric cases pass).  
Evaluation report: `openspec/changes/rfe-8685/validation_evaluation_report.md`  
Gaps: **1 critical**, **3 moderate**, **2 minor**

**Approve this artifact? (Approve / Reject with feedback)**

---

**User**

Approve validation for rfe-8685

/opsx-continue RFE-8685

Scope constraints (mandatory for this change):
- Read and honor openspec/changes/RFE-8685/inputs/scope-constraints.md
- Jira: use ONLY the text in openspec/changes/RFE-8685/inputs/jira-spec.md
- Do NOT use Jira MCP, linked issues, subtasks, epic, comments, or attachments
- Repo: use ONLY this working directory at the current HEAD (release-1.1)
- Do NOT git fetch, checkout main, compare to other branches, or browse GitHub
- Do NOT clone or read remote repos
- Working-folder mode: use this as the working directory

---

**Cursor**

Validation is approved. Creating the next artifact (`specs.md`) for rfe-8685.


Creating `specs.md` from the approved validation and Jira spec, addressing validation gaps as assumptions.


## Artifact created: `specs`

Validation was treated as approved. The next artifact **`specs.md`** is ready for your review.

**Output:** `openspec/changes/rfe-8685/specs.md`  
**Evaluation report:** `openspec/changes/rfe-8685/specs_evaluation_report.md`  
**Eval gate:** Skipped (no stage eval for specs)  
**Progress:** 2/9 artifacts written (pending your approval on specs)

---

### Spec summary

**Feature:** Custom Enterprise CA Bundle Integration (RFE-8685)

| Section | Content |
|---------|---------|
| **User stories** | 4 stories — P1: operator trusts enterprise PKI for internal vault sync; P2: store-level CA references + safe declarative management; P3: coexist with platform proxy CA injection |
| **Functional requirements** | FR-001–FR-010 — ConfigMap reference, declarative trust (no manual mounts), store CA references, error handling, lifecycle, proxy coexistence, degraded status |
| **Success criteria** | SC-001–SC-006 — sync without x509 errors, observable failures, safe removal, proxy coexistence, zero manual volume mount workflows |
| **Assumptions** | A-001–A-012 — personas, scope (operator + store CA), cert-manager out of direct scope, migration behavior |
| **NEEDS CLARIFICATION** | 2 markers (within max 3): shared vs per-store ConfigMap (A-007); platform CA behavior when proxy is absent (A-011) |

Validation gaps were addressed: personas, Given/When/Then scenarios, edge cases (missing/invalid/deleted ConfigMap), and operational semantics. No implementation details (no CRD names, file paths, or API groups).

---

Eval score: **N/A** (specs stage — eval skipped)  
Evaluation report: `openspec/changes/rfe-8685/specs_evaluation_report.md`  
Gaps: **0 critical**, **2 moderate** (NEEDS CLARIFICATION items), **2 minor**

**Approve this artifact? (Approve / Reject with feedback)**

Note: Rejecting `specs.md` exits the workflow per schema rules — it will not be regenerated in a feedback loop. Approval unlocks **repo-assessment** on the next `/opsx-continue`.

---

**User**

Approve specs for rfe-8685

/opsx-continue rfe-8685

Scope constraints (mandatory for this change):
- Read and honor openspec/changes/rfe-8685/inputs/scope-constraints.md
- Jira: use ONLY openspec/changes/rfe-8685/inputs/jira-spec.md
- Do NOT use Jira MCP, linked issues, subtasks, epic, comments, or attachments
- Repo: use ONLY this working directory at the current HEAD (release-1.1)
- Do NOT git fetch, checkout main, compare to other branches, or browse GitHub
- Do NOT clone or read remote repos
- Working-folder mode: use this as the working directory

Repo-assessment focus:
- Ground scope in release-1.1 only — cite existing files; do not assume APIs or fields that are not in this checkout
- Resolve operator-level vs SecretStore/ClusterSecretStore ownership (downstream operator vs upstream operand CRDs)
- Note existing proxy-driven trusted CA injection in the codebase vs the new enterprise CA ConfigMap requirement
- Carry forward NEEDS CLARIFICATION from specs.md (A-007, A-011) without expanding Jira scope

---

**Cursor**

Specs approved. Creating `repo-assessment.md` grounded in release-1.1 at HEAD — exploring the codebase for CA injection, proxy trust, and operator vs operand CRD ownership.


Writing the repo-assessment grounded in release-1.1 evidence.


## Artifact created: `repo-assessment`

**Output:** `openspec/changes/rfe-8685/repo-assessment.md`  
**Evaluation report:** `openspec/changes/rfe-8685/repo-assessment_evaluation_report.md`  
**Eval results:** `openspec/changes/rfe-8685/eval-results/repo-assessment.yaml`  
**Progress:** 3/9 artifacts (pending your approval)

Grounded in **`release-1.1` @ `1f54b74a`** — no remote fetch, no assumed APIs.

---

### Findings summary

**Operator vs SecretStore ownership (release-1.1)**

| Layer | Owner | On branch today | Work type |
|-------|-------|-----------------|-----------|
| Enterprise CA ConfigMap on operator API | Downstream operator | **Not present** | Greenfield API + reconcile |
| CA mounts on controller/webhook pods | `pkg/controller/external_secrets` | **Proxy-gated only** | Extend existing volume logic |
| `caProvider` / `caBundle` on SecretStore | Upstream `external-secrets.io` | **Already in CRD v0.20.4** | Configure/document — no operator schema change |

**Existing proxy-driven trusted CA injection**

- `ensureTrustedCABundleConfigMap` and `updateTrustedCABundleVolumes` run **only when** `getProxyConfiguration()` is non-nil (`configmap.go`, `deployments.go`).
- Fixed CNO-managed ConfigMap: `external-secrets-trusted-ca-bundle`, label `config.openshift.io/inject-trusted-cabundle`, mount `/etc/pki/tls/certs`.
- **Without proxy:** no platform CA ConfigMap and no CA volume mounts on operand pods — matches jira pain for enterprise PKI-only clusters.

**Primary target files:** `api/v1alpha1/` (new field), `configmap.go`, `deployments.go`, `constants.go`, unit/API tests; **not** upstream SecretStore CRD YAML.

**Carried forward:** A-007 (shared vs per-store ConfigMap), A-011 (platform CA when proxy absent) — flagged in §11.1, not expanded beyond specs.

**Gap:** `Feature:TrustedCABundle` e2e label is documented but no `test/e2e/*.go` on this branch.

---

Eval score: **100%** (0/0 automated cases — stage eval list empty; manual template check pass)  
Evaluation report: `openspec/changes/rfe-8685/repo-assessment_evaluation_report.md`  
Gaps for planning: **0 critical**, **3 moderate** (mount path strategy, missing e2e, store-only vs operand trust)

**Approve this artifact? (Approve / Reject with feedback)**

Approval unlocks **plan** (and constitution input resolution per schema).

---

**User**

Approve repo-assessment for rfe-8685

/opsx-continue rfe-8685

Scope constraints (mandatory for this change):
- Read and honor openspec/changes/rfe-8685/inputs/scope-constraints.md
- Jira: use ONLY openspec/changes/rfe-8685/inputs/jira-spec.md
- Do NOT use Jira MCP, linked issues, subtasks, epic, comments, or attachments
- Repo: use ONLY this working directory at the current HEAD (release-1.1)
- Do NOT git fetch, checkout main, compare to other branches, or browse GitHub
- Do NOT clone or read remote repos
- Working-folder mode: use this as the working directory

Plan focus:
- Build on approved repo-assessment.md and specs.md only
- Prioritize operator-owned work: ExternalSecretsConfig enterprise CA API + controller/webhook volume reconcile
- Treat upstream SecretStore/ClusterSecretStore caProvider as configuration/docs path, not operator CRD changes
- Resolve A-007 and A-011 explicitly in plan (shared vs per-store ConfigMap; proxy-absent platform CA behavior)
- Do not assume API field names from other branches; derive from release-1.1 patterns

---

**Cursor**

Repo-assessment approved. Creating `plan.md` from the approved specs and repo-assessment, resolving constitution input, and running the plan eval gate.


Writing `plan.md` with explicit resolutions for A-007 and A-011, grounded in release-1.1 patterns.


## Artifact created: `plan`

**Output:** `openspec/changes/rfe-8685/plan.md`  
**Evaluation report:** `openspec/changes/rfe-8685/plan_evaluation_report.md`  
**Progress:** 4/9 artifacts (pending your approval)

---

### Plan summary

**7 phases** — operator-first, upstream store CA as docs only:

| Phase | Focus | Agent |
|-------|--------|-------|
| 1 | API: `additionalTrustedCAConfigMapRef` on `CommonConfigs` / ESC `appConfig` | API_Agent |
| 2 | Codegen, CRD, bundle | API_Agent + OLMRelease_Agent |
| 3 | Controller volumes + ConfigMap validation (controller + webhook) | OperatorController_Agent |
| 4 | Degraded status, unit tests (A-011 scenarios) | OperatorController_Agent |
| 5 | Docs for existing `caProvider` — **no CRD/bindata changes** | Docs_Agent |
| 6 | E2E `Feature:TrustedCABundle` | Testing_Agent |
| 7 | `make verify && make lint`, bundle | OLMRelease_Agent |

### A-007 resolved (shared vs per-store ConfigMap)

- **ESC field:** one optional ConfigMap ref for operand pod trust (cluster admin).
- **Store `caProvider`:** independent per upstream CRD; operator does not sync from ESC.
- **Same ConfigMap object:** allowed if admin configures both — convention only, not enforced.

### A-011 resolved (proxy absent platform CA)

| Proxy | Enterprise CA | Behavior |
|-------|---------------|----------|
| Off | Off | No CA volumes (unchanged) |
| Off | On | Enterprise ConfigMap only — sole supplemental trust |
| On | Off | CNO platform CA only (unchanged) |
| On | On | **Projected volume** merging CNO + enterprise ConfigMaps at `/etc/pki/tls/certs` |

### API shape (release-1.1 pattern)

Proposed on `CommonConfigs` in `meta.go`:

`spec.appConfig.additionalTrustedCAConfigMapRef` → `{ name, namespace, key? }`

Not copied from other branches; aligned with existing `ProxyConfig` / `SecretReference` style.

---

Eval score: **100%** (0/0 automated cases)  
Evaluation report: `openspec/changes/rfe-8685/plan_evaluation_report.md`  
Open questions: **None** (A-007 and A-011 closed in plan)

**Approve this artifact? (Approve / Reject with feedback)**

Approval unlocks **tasks** on the next `/opsx-continue`.

---

**User**

Approve plan for rfe-8685

/opsx-continue rfe-8685

Scope constraints (mandatory for this change):
- Read and honor openspec/changes/rfe-8685/inputs/scope-constraints.md
- Jira: use ONLY openspec/changes/rfe-8685/inputs/jira-spec.md
- Do NOT use Jira MCP, linked issues, subtasks, epic, comments, or attachments
- Repo: use ONLY this working directory at the current HEAD (release-1.1)
- Do NOT git fetch, checkout main, compare to other branches, or browse GitHub
- Do NOT clone or read remote repos
- Working-folder mode: use this as the working directory

Tasks focus:
- Follow approved plan.md Phases 1–7 in order
- One task per plan phase capability; pair implementation with verification per constitution
- Use plan field name additionalTrustedCAConfigMapRef unless API task explicitly revises naming
- Phase 6 e2e (Feature:TrustedCABundle) as separate optional task if cluster unavailable
- Do not add upstream SecretStore/ClusterSecretStore CRD or bindata tasks

---

**Cursor**

Plan approved. Creating `tasks.md` from Phases 1–7, pairing each implementation task with verification per the constitution.


Creating the execution backlog with implementation and verification tasks paired per phase.


## Artifact created: `tasks`

**Output:** `openspec/changes/rfe-8685/tasks.md`  
**Evaluation report:** `openspec/changes/rfe-8685/tasks_evaluation_report.md`  
**Progress:** 5/9 artifacts (pending your approval)

---

### Backlog summary

**13 tasks** across plan Phases 1–7 — implementation paired with verification per constitution:

| Phase | Implement | Verify |
|-------|-----------|--------|
| 1 API | T1_1 `additionalTrustedCAConfigMapRef` types + testsuite | T1_2 `make test-apis` |
| 2 Codegen | T2_1 generate/manifests/docs/bundle | T2_2 `make verify` |
| 3 Controller | T3_1 volume projection + ConfigMap validation (**High risk**) | T3_2 `go test` smoke |
| 4 Status/tests | T4_1 Degraded + A-011 unit matrix | T4_2 `make test-unit` |
| 5 Docs | T5_1 ESC + store `caProvider` doc | T5_2 manual checklist |
| 6 E2E | **T6_1 OPTIONAL** `Feature:TrustedCABundle` | (cluster required; skippable) |
| 7 Release | T7_1 bundle + verify | T7_2 `make lint` + `make test` |

**Complexity:** 40 points (35 required, excluding optional T6_1)  
**Agents:** API_Agent (4), OperatorController_Agent (3), Testing_Agent (2), Docs_Agent (2), OLMRelease_Agent (2)

**Scope honored:**
- Field name: `additionalTrustedCAConfigMapRef`
- No upstream SecretStore/ClusterSecretStore CRD or bindata tasks
- T6_1 optional when cluster unavailable — T7_1 can proceed after T4_2 + T5_2

**Linear order starts:** T1_1 → T1_2 → T2_1 → … → T7_2

---

Eval score: **100%** (0/0 automated cases)  
Evaluation report: `openspec/changes/rfe-8685/tasks_evaluation_report.md`

**Approve this artifact? (Approve / Reject with feedback)**

Approval unlocks the **implementation** stage (`/opsx-apply` with T1_1).

---

**User**

Approve tasks for rfe-8685

/opsx-apply rfe-8685

Scope constraints (mandatory for this change):
- Read and honor openspec/changes/rfe-8685/inputs/scope-constraints.md
- Jira: use ONLY openspec/changes/rfe-8685/inputs/jira-spec.md
- Do NOT use Jira MCP, linked issues, subtasks, epic, comments, or attachments
- Repo: use ONLY this working directory at the current HEAD (release-1.1)
- Do NOT git fetch, checkout main, compare to other branches, or browse GitHub
- Do NOT clone or read remote repos
- Working-folder mode: use this as the working directory

Apply focus:
- Execute tasks in linear order from tasks.md (T1_1 first)
- One task per /opsx-apply invocation; wait for my approval before the next task
- Use field name additionalTrustedCAConfigMapRef per tasks.md
- Skip T6_1 unless I explicitly request e2e (no cluster available)
- Run verification commands paired with each task (make test-apis, make verify, make test-unit, etc.)

---

**Cursor**

Tasks approved. Starting `/opsx-apply` with T1_1 — reading the apply workflow and current API patterns.


Implementing T1_1: adding API types and testsuite cases.


**Using change:** `rfe-8685` (working-folder mode at `release-1.1` HEAD; override with `/opsx-apply <other-change>`)

---

## Task 1/13: T1_1 — Add `additionalTrustedCAConfigMapRef` to `CommonConfigs`

**Phase:** 1 — Operator API

### OAPE Commands Executed
Manual API_Agent implementation (equivalent to `/oape:api-generate` scope for types + testsuite only; CRD manifests deferred to T2_1).

### Files Touched
| File | Change |
|------|--------|
| `api/v1alpha1/meta.go` | Added `AdditionalTrustedCAConfigMapRef` struct; field on `CommonConfigs`; CEL validation |
| `api/v1alpha1/zz_generated.deepcopy.go` | Regenerated via `make generate` |
| `api/v1alpha1/tests/.../externalsecretsconfig.testsuite.yaml` | 5 new onCreate cases (2 valid, 3 invalid) |

### API shape
- **Field:** `spec.appConfig.additionalTrustedCAConfigMapRef`
- **Struct:** `{ name, namespace, key? }` — `key` defaults to `ca-bundle.crt` (documented; OpenAPI default applied at CRD generation in T2_1)
- **CEL:** When ref is set, `name` and `namespace` must be non-empty

### Test Results
| Command | Result |
|---------|--------|
| `make generate` | PASSED |
| `go build ./api/v1alpha1/...` | PASSED |
| `go vet ./api/v1alpha1/...` | PASSED |

`make test-apis` is intentionally **not** run yet — that is T1_2 (requires CRD regeneration in T2_1 first).

### Code Generation Eval Scorecard
No applicable eval cases in `evals/code-generation_eval.yaml` (empty registry). Verification-only gate passed.

### Deviations
None.

---

**Code eval score: N/A (0 eval cases; build/vet verification passed).**

Approve the code changes for task **T1_1** and proceed to **T1_2** (`make test-apis` after manifests in T2_1)?

**(Approve / Reject with feedback)**

---

**User**

Approve T1_1 for rfe-8685

/opsx-apply rfe-8685 T1_2

Scope constraints (mandatory for this change):
- Read and honor openspec/changes/rfe-8685/inputs/scope-constraints.md
- Jira: use ONLY openspec/changes/rfe-8685/inputs/jira-spec.md
- Do NOT use Jira MCP, linked issues, subtasks, epic, comments, or attachments
- Repo: use ONLY this working directory at the current HEAD (release-1.1)
- Do NOT git fetch, checkout main, compare to other branches, or browse GitHub
- Do NOT clone or read remote repos
- Working-folder mode: use this as the working directory

Task: T1_2 — Verify API testsuite for enterprise CA ref
- Run make test-apis
- Fix T1_1 only if validation failures occur
- Do not start T2_1 yet

---

**Cursor**

Recording T1_1 approval, then running T1_2 verification with `make test-apis`.


CRD manifests are stale — the new field isn't in the schema yet. Running `make manifests` (minimal prerequisite), then re-running `make test-apis`.


Updating testsuite error messages to match the CRD's multi-error format (T1_1 fix).


**Using change:** `rfe-8685` (working-folder mode; override with `/opsx-apply <other-change>`)

T1_1 is recorded as approved (`implementation/task-reports/T1_1.md`, phase log updated).

---

## Task 2/13: T1_2 — Verify API testsuite for enterprise CA ref

**Phase:** 1 — Operator API

### OAPE Commands Executed
Verification-only (`api-generate-tests` equivalent): run `make test-apis`; fix testsuite if validation messages mismatch.

### Files Touched
| File | Change |
|------|--------|
| `api/v1alpha1/tests/.../externalsecretsconfig.testsuite.yaml` | Fixed 2 `expectedError` strings for missing `name`/`namespace` (CEL multi-error format) |
| `config/crd/bases/operator.openshift.io_externalsecretsconfigs.yaml` | Regenerated via `make manifests` (required for CRD validation in envtest) |

### Test Results
| Step | Result |
|------|--------|
| `make test-apis` (initial, stale CRD) | FAILED — 5 failures (`unknown field`) |
| `make manifests` | PASSED — CRD includes `additionalTrustedCAConfigMapRef` |
| `make test-apis` (after manifest regen) | FAILED — 2 failures (error message format) |
| Testsuite `expectedError` fix | Applied |
| `make test-apis` (final) | **PASSED** — 112/112 (all 5 new cases green) |

### New tests verified
- Valid ref (default `key`)
- Valid ref (custom `key`)
- Missing `namespace` → validation error
- Missing `name` → validation error
- Invalid `namespace` → pattern error

### Code Generation Eval Scorecard
No applicable eval cases (empty registry). Verification gate passed.

### Deviations
- **`make manifests` was run** as a prerequisite for `make test-apis` (CRD was stale after T1_1). This overlaps part of T2_1 but **T2_1 was not started** — no docs, bundle, or full `make verify` run.

---

**Code eval score: N/A (0 eval cases; `make test-apis` 112/112 pass).**

Approve the code changes for task **T1_2** and proceed to **T2_1**?

**(Approve / Reject with feedback)**

---

**User**

Approve T1_2 for rfe-8685

/opsx-apply rfe-8685 T2_1

Scope constraints (mandatory for this change):
- Read and honor openspec/changes/rfe-8685/inputs/scope-constraints.md
- Jira: use ONLY openspec/changes/rfe-8685/inputs/jira-spec.md
- Do NOT use Jira MCP, linked issues, subtasks, epic, comments, or attachments
- Repo: use ONLY this working directory at the current HEAD (release-1.1)
- Do NOT git fetch, checkout main, compare to other branches, or browse GitHub
- Do NOT clone or read remote repos
- Working-folder mode: use this as the working directory

Task: T2_1 — Run codegen, manifests, docs, bundle
- Run make generate && make manifests && make docs && make bundle (or make update subset)
- Regenerate: zz_generated.deepcopy.go, CRD bases, docs/api_reference.md, bundle/manifests
- Do not hand-edit generated files
- No bindata or upstream external-secrets.io CRD refresh
- Do not start T2_2 or T3_1 yet

---

**Cursor**

Recording T1_2 approval, then running T2_1 codegen and regeneration.


**Using change:** `rfe-8685` (working-folder mode; override with `/opsx-apply <other-change>`)

T1_2 is recorded as approved (`implementation/task-reports/T1_2.md`, phase log updated).

---

## Task 3/13: T2_1 — Run codegen, manifests, docs, bundle

**Phase:** 2 — Codegen

### OAPE Commands Executed
Manual API_Agent regeneration (equivalent to post-`api-generate` codegen step):

```bash
make generate && make manifests && make docs && make bundle
```

### Files Touched (generated only)
| File | Change |
|------|--------|
| `api/v1alpha1/zz_generated.deepcopy.go` | DeepCopy for `AdditionalTrustedCAConfigMapRef` |
| `config/crd/bases/operator.openshift.io_externalsecretsconfigs.yaml` | OpenAPI schema + CEL for new field |
| `config/crd/bases/operator.openshift.io_externalsecretsmanagers.yaml` | Same field via shared `CommonConfigs` |
| `docs/api_reference.md` | API docs for new type and field |
| `bundle/manifests/operator.openshift.io_externalsecretsconfigs.yaml` | Bundle CRD sync |
| `bundle/manifests/operator.openshift.io_externalsecretsmanagers.yaml` | Bundle CRD sync |
| `bundle/manifests/external-secrets-operator.clusterserviceversion.yaml` | CSV metadata refresh |

**Not touched:** `bindata/`, `pkg/operator/assets/`, upstream `external-secrets.io` CRDs.

### Test Results
| Command | Result |
|---------|--------|
| `make generate` | PASSED |
| `make manifests` | PASSED |
| `make docs` | PASSED |
| `make bundle` | PASSED (pre-existing example-annotation warnings only) |

All target artifacts include `additionalTrustedCAConfigMapRef` with `key` default `ca-bundle.crt` and CEL validation.

### Code Generation Eval Scorecard
No applicable eval cases (empty registry). Regeneration commands completed successfully.

### Deviations
- **ESM CRD/bundle** also regenerated — expected because `CommonConfigs` is embedded in both ESC and ESM types; field is on ESC `appConfig` only at runtime, but schema is shared.

---

**Code eval score: N/A (0 eval cases; all four make targets passed).**

Approve the code changes for task **T2_1** and proceed to **T2_2** (`make verify-generated` / `make verify`)?

**(Approve / Reject with feedback)**

---

**User**

/opsx-apply rfe-8685 — continue next task (same scope as before)

---

**Cursor**

Recording T2_1 approval and running T2_2 verification (`make verify-generated`, `make verify`).


**Using change:** `rfe-8685` (working-folder mode; override with `/opsx-apply <other-change>`)

T2_1 is recorded as approved (`implementation/task-reports/T2_1.md`, phase log updated).

---

## Task 4/13: T2_2 — Verify generated CRD and deepcopy consistency

**Phase:** 2 — Codegen

### OAPE Commands Executed
Verification-only (no code changes unless verify fails).

### Files Touched
None — verification task only.

### Test Results

| Check | Result | Notes |
|-------|--------|-------|
| `make verify-generated` | **PASSED** | Bindata matches source (`verify-bindata` diff clean) |
| `make vet` | **PASSED** | |
| `make fmt` | **PASSED** | |
| `make verify-deps` | **PASSED** | |
| `make verify-bindata` | **PASSED** | |
| `make verify-bindata-assets` | **PASSED** | |
| Codegen idempotency | **PASSED** | Re-ran `generate/manifests/docs/bundle` — no additional diffs beyond T2_1 outputs |
| `make govulncheck` | **FAILED** | 2 pre-existing vulns on `release-1.1` baseline: GO-2026-5026, GO-2026-4918 (not introduced by RFE-8685) |
| `make verify` (full) | **FAILED** | Stopped at `govulncheck` (same baseline issue) |
| `check-git-diff` | **Not reached** | Would also fail in working-folder mode due to uncommitted implementation changes (expected before commit) |

**Generated artifact consistency:** DeepCopy, CRD bases, API docs, and bundle manifests for `additionalTrustedCAConfigMapRef` are in sync and idempotent. No stale generated file diffs for this change.

### Code Generation Eval Scorecard
No applicable eval cases (verification-only). Codegen consistency checks passed.

### Deviations
- **`make verify` does not fully pass** on this branch due to pre-existing `govulncheck` findings unrelated to RFE-8685. All codegen/bindata verification targets relevant to T2_1 outputs passed.
- **`check-git-diff`** not evaluated to completion — requires a committed tree; uncommitted WIP is expected at this stage.

---

**Code eval score: N/A (0 eval cases; codegen verification passed; full `make verify` blocked by baseline govulncheck).**

Approve task **T2_2** and proceed to **T3_1** (controller volume projection)?

**(Approve / Reject with feedback)**

---

**User**

Approve — continue next task

---

**Cursor**

Recording T2_2 approval and starting T3_1 — controller volume projection (high-risk).


Implementing T3_1: adding helpers, ConfigMap sync, volume projection, and updating `volumesEqual`.


**Using change:** `rfe-8685` (working-folder mode)

T2_2 is recorded as approved. T3_1 implementation is complete.

---

## Task 5/13: T3_1 — Implement operand trusted CA volume reconcile

**Phase:** 3 — Controller

### OAPE Commands Executed
Manual `OperatorController_Agent` implementation (api-implement scope).

### Files Touched
| File | Change |
|------|--------|
| `pkg/controller/external_secrets/constants.go` | Enterprise CA CM name, default key, projected filename |
| `pkg/controller/external_secrets/utils.go` | Getters for ref/key; `isOperandTrustedCAMountSupported` |
| `pkg/controller/external_secrets/configmap.go` | `ensureEnterpriseTrustedCAConfigMap`, `validateEnterpriseCAConfigMap`, PEM check |
| `pkg/controller/external_secrets/deployments.go` | A-011 volume matrix: platform-only, enterprise-only, projected, neither |
| `pkg/controller/external_secrets/install_external_secrets.go` | Call `ensureEnterpriseTrustedCAConfigMap` before deployments |
| `pkg/controller/common/utils.go` | `volumesEqual` supports projected volumes |
| `pkg/controller/external_secrets/deployments_test.go` | Updated `updateProxyConfiguration` call signature |

### A-011 matrix implemented
| Proxy | Enterprise | Controller/Webhook behavior |
|-------|------------|-------------------------------|
| Off | Off | No trusted-CA volume |
| Off | On | Enterprise synced CM volume only |
| On | Off | Platform CNO CM volume (unchanged) |
| On | On | **Projected** volume merging platform + enterprise at `/etc/pki/tls/certs` |

- Enterprise mounts apply **only** to `controllerDeploymentAssetName` and `webhookDeploymentAssetName`.
- Cross-namespace refs: user CM fetched via `UncachedClient`, data synced to `external-secrets-additional-trusted-ca` in operand namespace.
- Platform CNO ConfigMap path unchanged (`ensureTrustedCABundleConfigMap` still proxy-gated).
- No cert-controller/bitwarden enterprise mounts.

### Test Results
| Command | Result |
|---------|--------|
| `go build ./pkg/controller/external_secrets/...` | PASSED |
| `go vet ./pkg/controller/external_secrets/...` | PASSED |
| `go test ./pkg/controller/external_secrets/...` | PASSED |

### Code Generation Eval Scorecard
No applicable eval cases. Build/vet/test passed.

### Deviations
- **Degraded status** for missing/invalid ConfigMap deferred to **T4_1** (validation returns reconcile errors now).
- **RBAC** unchanged — existing cluster-scoped `configmaps` get/list/watch suffices.

---

**Code eval score: N/A (0 eval cases; build/vet + unit tests pass).**

Approve task **T3_1** and proceed to **T3_2** (controller package verification)?

**(Approve / Reject with feedback)**

---

**User**

Approve — continue next task
Stop after T4_2 — skip T5–T7 unless I ask

---

**Cursor**

Recording T3_1 approval, then running T3_2–T4_2 (stopping before T5).


Implementing T3_2 verification, then T4_1 (ConfigurationError + status + tests).


**Using change:** `rfe-8685` (working-folder mode)

Completed **T3_2 → T4_2** per your batch directive. **Stopped before T5–T7** (docs, e2e, release).

---

## Tasks completed (batch)

### T3_2 — Verify controller package tests
| Check | Result |
|-------|--------|
| `go test ./pkg/controller/external_secrets/...` | **PASSED** |

### T4_1 — Degraded status + A-011 unit tests

**Status handling**
- New `ConfigurationError` in `pkg/controller/common/errors.go` for missing enterprise CA ConfigMaps
- `processReconcileRequest` sets **Degraded=True** on configuration + irrecoverable errors (still requeues configuration errors)

**Tests added**
- `configmap_test.go` — validation (missing/invalid/valid) + sync to operand namespace
- `deployments_enterprise_ca_test.go` — enterprise-only, proxy+projected, cert-controller exclusion, missing CM error, removal path
- `errors_test.go` — `ConfigurationError` classification

### T4_2 — Full unit gate
| Check | Result |
|-------|--------|
| `make test-unit` | **PASSED** |

---

## Implementation summary (T1_1 → T4_2)

| Area | Delivered |
|------|-----------|
| API | `spec.appConfig.additionalTrustedCAConfigMapRef` + testsuite |
| Codegen | CRD, deepcopy, docs, bundle |
| Controller | A-011 volume matrix, cross-ns CM sync, PEM validation |
| Status | Degraded on missing/invalid enterprise CA prerequisites |
| Tests | API (112/112), unit suite green |

---

## Deferred (per your instruction)

| Task | Status |
|------|--------|
| T5_1 / T5_2 | Docs — not started |
| T6_1 | E2e — skipped (no cluster) |
| T7_1 / T7_2 | Bundle verify + lint — not started |

Task reports: `implementation/task-reports/T3_2.md`, `T4_1.md`, `T4_2.md`

---

**Approve the T3_2–T4_2 batch** to lock this stop point, or say if you want **T5_1** (docs) next.

---

**User**

/eval-loop rfe-8685

---

**Cursor**

Running the eval loop for rfe-8685: reading the skill, pipeline config, and inputs.


Inputs still have placeholders. Populating `evals/inputs/` from the rfe-8685 change artifacts, then running the loop.


Populating inputs and generating eval-loop outputs from the rfe-8685 implementation.


Creating eval cases and updating baseline files.


Continuing the eval-loop: syncing schema evals and updating registry, round state, and snapshots.


**Loop complete (round 1).** Review `evals/baseline/` and `evals/refined-templates/`. Replace `evals/inputs/` with the next feature bundle and run `/eval-loop` again.

## RFE-8685 eval-loop summary

### Epic Bug Analysis
- **6 patterns** (PAT-001–PAT-006) from workflow friction — no Jira bugs in bundle
- Outputs: `evals/outputs/epic-bug-analysis/{pattern-analysis,rca-summary,issue-taxonomy.json}`

### Template refinements (4 patched)
| Template | Gap addressed |
|----------|---------------|
| `repo-assessment-template.md` | Trusted CA / proxy baseline (§4.6) |
| `plan-template.md` | A-011 matrix + cross-ns sync |
| `tasks-template.md` | `make test-apis` needs manifests first |
| `validation-template.md` | Operator vs upstream CRD scope |

Patches: `evals/outputs/eval-generation/refinement-patches/*.patch`

### Consolidated eval cases (19 total)

| Stage | Cases | File |
|-------|-------|------|
| repo-assessment | 3 | `evals/baseline/evals/repo-assessment/repo-assessment_eval.yaml` |
| constitution | 3 | `evals/baseline/evals/constitution/constitution_eval.yaml` |
| plan | 3 | `evals/baseline/evals/plan/plan_eval.yaml` |
| tasks | 3 | `evals/baseline/evals/tasks/tasks_eval.yaml` |
| implementation | 3 | `evals/baseline/evals/implementation/implementation_eval.yaml` |
| code-generation | 4 | `evals/baseline/evals/code-generation/code-generation_eval.yaml` |

Code-generation evals cover `api-generate`, `api-generate-tests`, `api-implement`, and `manual` (manifests prerequisite).

### Synced & snapshotted
- Forward workflow: `openspec/schemas/openspec-agile-workflow/evals/*_eval.yaml` (template paths use `templates/…`)
- Registry: `evals/baseline/evals-registry.yaml`
- Round state: `evals/round-state.yaml` → **round 1**
- Snapshot: `evals/baseline/rounds/round-1/`
- Routing learnings: `evals/baseline/routing-learnings.md`

Future `/opsx-continue` and `/opsx-apply` runs will use these eval gates for similar CA/trust features on this operator.

---

**User**

/opsx-archive rfe-8685

---

**Cursor**

Archiving `rfe-8685`: checking artifact status, tasks, and delta specs first.


Pre-archive check: some artifacts and tasks are incomplete. Confirming before moving the change to archive.


**Archive rfe-8685 with warnings?**

> ****Change:** rfe-8685 (schema: openspec-agile-workflow)

**Incomplete artifacts (4):**
- adrs (ready — not written)
- implementation (ready — phase log exists under `implementation/` but root artifact missing)
- implementation-checklist (ready)
- implementation-report (ready)

**Incomplete tasks (5):**
- T5_1 — docs (**deferred** by user)
- T5_2 — verify docs
- T6_1 — e2e (**optional**, skipped)
- T7_1 — bundle/verify
- T7_2 — lint/test gate

**Completed:** T1_1–T4_2 (core implementation)

**Delta specs:** None — no sync needed

**Archive target:** `openspec/changes/archive/2026-07-02-rfe-8685/`

Proceed with archive?**
> - Yes — archive now (accept warnings) ✔
> - Cancel — keep change active

## Archive Complete (with warnings)

**Change:** rfe-8685  
**Schema:** openspec-agile-workflow  
**Archived to:** `openspec/changes/archive/2026-07-02-rfe-8685/`  
**Specs:** No delta specs

### Warnings
- **4 incomplete artifacts:** `adrs`, `implementation`, `implementation-checklist`, `implementation-report` (planning artifacts done; implementation closure artifacts not written at change root)
- **5 incomplete tasks:** T5_1 (deferred), T5_2, T6_1 (optional/skipped), T7_1, T7_2
- **Core implementation complete:** T1_1–T4_2 (API, codegen, controller, unit tests)

No active changes remain under `openspec/changes/` — only the archive folder. Implementation artifacts (phase log, task reports) are preserved under the archived change’s `implementation/` directory.

