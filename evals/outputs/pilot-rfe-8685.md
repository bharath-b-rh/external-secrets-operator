# OpenSpec SDLC Pilot — RFE-8685 (Track B Greenfield)

**Feature:** Custom Enterprise CA Bundle Integration (RFE-8685)  
**Branch:** `release-1.1` @ `1f54b74a` (pre-feature base)  
**Known solution reference:** shipped PR chain below (merged on `main`)  
**Change:** `openspec/changes/archive/2026-07-02-rfe-8685`  
**Eval-loop:** Round 1 complete (`evals/baseline/rounds/round-1/`)  
**Date:** 2026-07-02

**Stakeholder summary:** [`pilot-rfe-8685-executive-summary.md`](pilot-rfe-8685-executive-summary.md)

### Official delivery chain (production)

| Layer | PR | Jira | LOC (merge commit) |
|-------|-----|------|-------------------|
| Enhancement proposal | [openshift/enhancements#2011](https://github.com/openshift/enhancements/pull/2011) | ESO-417 | EP doc only |
| API + CRD + testsuite | [external-secrets-operator#145](https://github.com/openshift/external-secrets-operator/pull/145) | ESO-435 | **+324 / −24** (9 files) |
| Controller + e2e + NP refactor | [external-secrets-operator#149](https://github.com/openshift/external-secrets-operator/pull/149) | ESO-396 | **+3,261 / −968** (43 files; includes network-policy refactor beyond CA) |

**EP + PR design (authoritative):**

- `spec.controllerConfig.trustedCABundle` → `ConfigMapKeyReference` (`name`, `key` default `ca-bundle.crt`)
- ConfigMap must exist in **operand namespace** (no cross-namespace ref in API)
- Mount **core controller only** — webhook and cert-controller **not** affected
- Separate volume `user-ca-bundle` at `/etc/pki/tls/user-certs` + `SSL_CERT_DIR` env
- Full PEM CA validation (`trusted_ca_bundle.go`), warning events, ConfigMap watch label
- Skip user mount when CNO inject label + proxy active
- Reserve `SSL_CERT_DIR` / `SSL_CERT_FILE` in `componentConfigs.overrideEnv` CEL
- E2E: `test/e2e/trusted_ca_bundle_test.go`

---

## Executive summary

OpenSpec produced a **coherent end-to-end greenfield solution** from scope-locked Jira text without spoilers or scope violations. The implementation **solves the customer problem directionally** (operator-level ConfigMap CA → operand TLS trust) but **diverges materially** from the known `main` solution in API shape, mount semantics, validation depth, and workload scope.

| Dimension | Verdict |
|-----------|---------|
| Spec pipeline | Strong — thin Jira → grounded plan/tasks |
| Implementation (T1_1–T4_2) | Functional prototype — `make test-unit` passed |
| vs known solution | ~**60–70% concept alignment**; ~**400–700 LOC rework** to reach production parity (excl. e2e) |
| Expected production bugs if shipped as-is | **4–6 likely** (see §6) |
| Adopt OpenSpec for next real story? | **4/5** — with eval-loop template fixes applied |

---

## Spec pipeline metrics

| Metric | Value |
|--------|-------|
| Validation `overall_score` | **56%** |
| Validation status | **NEEDS_REVISION** (0 blockers) |
| `[NEEDS CLARIFICATION]` in specs.md | **2** (A-007, A-011) |
| Feedback rounds per artifact | **0** (batch approvals) |
| Stage eval pass rate | validation rubric 0/2; repo/plan/tasks manual 100% |
| Scope violations (Jira MCP / main / linked Jira) | **0** |
| Artifacts approved | validation → specs → repo-assessment → plan → tasks |
| `/opsx-apply` tasks completed | **8 / 13** (61%) — stopped at T4_2 by design |
| `make test-apis` | Passed (after manifests regen — WFR-001) |
| `make verify` | Passed (T2_2) |
| `make test-unit` | Passed (T4_2) |
| Eval-loop cases added | **19** |
| Templates patched | **4** (validation, repo-assessment, plan, tasks) |

### Repo grounding

| Metric | Value |
|--------|-------|
| Valid file citations in repo-assessment | High — line refs to `deployments.go`, `configmap.go`, CRD YAML |
| False-positive features claimed | **0** |
| Operator vs upstream CRD split | Correctly identified in repo-assessment §11.2 |

### Planning / tasks

| Metric | Value |
|--------|-------|
| Tasks with Assigned Agent | **100%** (13/13) |
| Tasks with verification pairing | **100%** (impl + verify pairs) |
| High-risk task flagged | T3_1 (complexity 8) |

---

## Implementation size (greenfield)

| Category | Lines (approx.) |
|----------|-----------------|
| Hand-written production code | **~380 net** (+405 / −25) |
| New unit tests | **~422** (`configmap_test.go`, `deployments_enterprise_ca_test.go`) |
| Generated (CRD, bundle, deepcopy, api_reference) | **~193** insertions |
| OpenSpec / eval artifacts | separate (not feature LOC) |

**Known solution on `main` vs `release-1.1` (core, no e2e):** **~1,344 net** lines across API, controller, testsuite.  
**Known solution e2e:** **~490** additional lines (`test/e2e/trusted_ca_bundle_test.go`).

Greenfield delivered **~60% of production LOC** for core logic, **0%** of e2e, **0%** of user docs (T5 skipped).

---

## Compare to known solution (EP #2011 + PR #145 + PR #149)

### Greenfield vs each production artifact

| Production artifact | What it specifies | Greenfield match? |
|---------------------|-------------------|-------------------|
| **EP #2011** | `controllerConfig.trustedCABundle`, operand-ns ConfigMap, core controller TLS | **Partial** — right feature area, wrong API path |
| **PR #145** | `ConfigMapKeyReference`, testsuite, `SSL_CERT_DIR` reserved in overrideEnv CEL | **Miss** — different type/field; no policy field; ESM inherits via `CommonConfigs` |
| **PR #149** | `trusted_ca_bundle.go`, controller-only mount, watch label, CNO skip, e2e | **Partial** — similar reconcile ideas; wrong mount path, webhook included, no e2e |

### Concept alignment (Jira intent)

| Jira / business intent | Greenfield | Production (PR chain) | Match |
|------------------------|------------|----------------------|-------|
| ConfigMap reference for enterprise CA | `appConfig.additionalTrustedCAConfigMapRef` (name, **namespace**, key) | `controllerConfig.trustedCABundle` (`ConfigMapKeyReference`, operand-ns only) | Partial |
| Controller trusts enterprise PKI | Yes — sync CM + mount | Yes — direct mount + `SSL_CERT_DIR` (PR #149) | Partial |
| Webhook trusts enterprise PKI | **Yes** — mounted on webhook | **No** — EP + release notes: core controller only | **Mismatch** |
| No manual volume mounts | Yes | Yes | Yes |
| Coexist with proxy / platform CA | Projected volume at `/etc/pki/tls/certs` | Separate volumes + CNO skip (PR #149) | Different design |
| Store-level CA (`caProvider`) | Plan/docs path (not implemented) | Upstream CRD only — same conclusion | Yes |
| Degraded on missing/invalid CM | `ConfigurationError` → Degraded | `UserTrustedCABundleError` + events + fail-closed PEM (PR #149) | Partial |
| E2E `Feature:TrustedCABundle` | Not done | PR #149 e2e suite | No |

### Technical divergence matrix

| Area | Greenfield | Known (`main`) | Severity |
|------|------------|----------------|----------|
| **API location** | `CommonConfigs` / `appConfig` | `ControllerConfig` | High — breaking API shape |
| **API type** | `AdditionalTrustedCAConfigMapRef` (+ namespace) | `ConfigMapKeyReference` (operand namespace implied) | High |
| **ESM CRD** | Field appears via `CommonConfigs` embed | N/A on `ControllerConfig` path | Medium — API surface noise |
| **Cross-namespace ref** | Allowed; sync to operand CM | Operand namespace only | Medium — different ops model |
| **Mount targets** | Controller + webhook | Core controller container only | Medium — scope creep vs production |
| **Mount path** | `/etc/pki/tls/certs` (shared with platform) | `/etc/pki/tls/user-certs` + `SSL_CERT_DIR` | High — TLS behavior risk |
| **PEM validation** | Substring `BEGIN CERTIFICATE` check | Full x509 CA parse; reject keys, leaf certs | High |
| **CNO inject CM + proxy** | No skip — may double-mount / merge | Skip user mount + warning event | Medium |
| **Invalid PEM on update** | `IrrecoverableError` for missing key | Fail-closed keep mount + Degraded | Medium |
| **Watch label on user CM** | Not implemented | `updateWatchLabel` before validation | Low |
| **Code structure** | Logic in `deployments.go` / `configmap.go` | Dedicated `trusted_ca_bundle.go` | Low |
| **Events** | Minimal | Rich validation event reasons | Low |

---

## Deviation from concept

**Overall concept alignment: ~65%**

What aligned well:

- Operator-owned feature on `ExternalSecretsConfig` (not upstream SecretStore CRD edits).
- ConfigMap-based enterprise CA injection without `componentConfigs` volume hacks.
- Proxy × enterprise matrix thought through (A-011).
- Operand-namespace sync pattern correctly handles K8s cross-namespace mount limitation (discovered at implementation — not in Jira).

What diverged from production concept:

1. **Wrong API home** — production places trust on `controllerConfig`; greenfield used `appConfig`/`CommonConfigs` (followed `proxy` pattern, not production field).
2. **Wrong field name** — `additionalTrustedCAConfigMapRef` vs `trustedCABundle`.
3. **Broader workload scope** — webhook mounts not in known solution.
4. **Different TLS integration** — no `SSL_CERT_DIR`; merged mount path instead of separate user cert directory.
5. **Weaker validation and events** — production-grade PEM parsing and CNO-skip logic missing.

The pipeline **correctly captured business intent** but **could not infer production API conventions** without branch spoilers — expected for fair greenfield eval.

---

## What was gotten wrong

| # | Issue | Impact |
|---|-------|--------|
| 1 | API on `appConfig` instead of `controllerConfig.trustedCABundle` | Breaking vs shipping API; ESM inherits unused field |
| 2 | No `SSL_CERT_DIR` for Go TLS cert store | Controller may not load enterprise CAs reliably |
| 3 | Enterprise + platform CAs at same mount path via projection | Subtle TLS ordering / overwrite risk vs separate `user-certs` path |
| 4 | Shallow PEM validation | Accepts malformed bundles containing substring match |
| 5 | Webhook enterprise CA mount | Unnecessary scope; production targets core controller only |
| 6 | Missing CNO-managed CM skip when proxy enabled | Duplicate trust paths / confusing behavior |
| 7 | Cross-namespace API without product doc | Sync CM helps but adds drift if user CM updates unwatched |
| 8 | `IrrecoverableError` for missing ConfigMap key | Harsh reconcile behavior vs production `UserConfigurationError` |
| 9 | No e2e or user documentation | Not production-ready |
| 10 | T1_2 before T2_1 ordering | Workflow friction (caught by eval-loop WFR-001) |

**Not wrong (good decisions):**

- Projected volume for proxy+enterprise coexistence (reasonable alternative to dual-path + env).
- Operand-namespace ConfigMap sync for cross-ns references.
- Gating mounts with `isOperandTrustedCAMountSupported` (cert-controller excluded).
- Store `caProvider` scoped to docs, not CRD work.

---

## Solution effectiveness

| Criterion | Score (1–5) | Notes |
|-----------|-------------|-------|
| Solves customer x509 problem (theory) | **3** | Directionally yes; TLS path/env semantics unproven without e2e |
| Matches production architecture | **2** | Different API, mount, validation models |
| Test coverage confidence | **3** | Unit tests for A-011 matrix; no cluster proof |
| Security / validation rigor | **2** | PEM check too weak; missing private-key rejection |
| Operability (events, status) | **3** | Degraded on config errors; fewer events than production |
| Merge readiness | **2** | Needs API rename, controller semantics, e2e, docs |

**Effective as a pilot artifact:** **5/5** — demonstrates OpenSpec can drive greenfield implementation.  
**Effective as production code:** **2.5/5** — prototype requiring substantial convergence.

---

## LOC to reach production-quality solution

Estimate to converge greenfield → `main` parity:

| Work item | Est. LOC change |
|-----------|-----------------|
| Move/rename API to `controllerConfig.trustedCABundle` + `ConfigMapKeyReference` | **80–120** |
| Replace mount logic with controller-only + `user-certs` path + `SSL_CERT_DIR` | **150–250** |
| Port `trusted_ca_bundle.go` PEM validation + events | **+160** (new file) |
| Remove webhook enterprise mounts; simplify projected volume | **−80 to −120** |
| Add CNO skip + watch label behavior | **40–60** |
| Align error types (`UserConfigurationError`, fail-closed) | **30–50** |
| Port/adapt `trusted_ca_bundle_test.go` patterns | **+150–200** |
| E2E `trusted_ca_bundle_test.go` | **+490** |
| User docs (T5) | **+80–150** |

| Scope | Est. net LOC to converge |
|-------|--------------------------|
| **Core production parity (no e2e)** | **~400–700** changed/added |
| **Full parity incl. e2e + docs** | **~900–1,200** |

Alternative: **discard greenfield controller** and port `main` implementation wholesale (~1,344 net from `release-1.1`) — faster if API name is fixed.

---

## Expected bugs if greenfield shipped as-is

| ID | Bug / risk | Likelihood | Severity |
|----|------------|------------|----------|
| B1 | Go TLS ignores enterprise CAs (no `SSL_CERT_DIR`, crowded `/etc/pki/tls/certs`) | Medium | **High** |
| B2 | Invalid PEM accepted; sync proceeds with bad trust material | Medium | **High** |
| B3 | User refs CNO-inject-labeled CM + proxy → confusing duplicate/merged CAs | Medium | Medium |
| B4 | Synced operand CM stale after user CM update (watch gap) | Medium | Medium |
| B5 | Webhook gets unnecessary CA mount — unexpected trust surface | Low | Low |
| B6 | `IrrecoverableError` on bad key — ESC stuck vs recoverable Degraded | Low | Medium |
| B7 | Projected platform CM `optional: true` hides empty CNO bundle | Low | Medium |
| B8 | API consumers use ESM `additionalTrustedCAConfigMapRef` — no runtime effect | Medium | Low |

**Expected bug count (first 6 months):** **4–6** customer-visible or CI issues without e2e and production convergence.

---

## Eval-loop learnings (fed back)

| ID | Learning | Template fix |
|----|----------|--------------|
| WFR-001 | `make test-apis` before `make manifests` fails | tasks-template |
| WFR-002 | CEL multi-error `expectedError` format | code-generation eval |
| WFR-003 | Cross-ns CM → operand sync required | plan + repo-assessment |
| WFR-004 | `CommonConfigs` updates multiple CRDs | repo-assessment |

---

## Overall verdict (1–5)

| Criterion | Score |
|-----------|-------|
| Spec understanding (thin Jira) | **4** |
| Repo assessment accuracy | **5** |
| Plan/tasks actionability | **5** |
| Implementation correctness vs known solution | **2.5** |
| Implementation correctness vs Jira intent | **3.5** |
| Workflow efficiency (with compact prompts) | **4** |
| **Would adopt for next real story?** | **4** |

---

## Recommendations

1. **For production RFE-8685:** Converge to `main` `trustedCABundle` API and `trusted_ca_bundle.go` patterns rather than evolving greenfield field names.
2. **For next OpenSpec eval:** Add repo-assessment hint: “check `main` API naming conventions” only when **not** doing blind greenfield; or accept divergence as measured outcome.
3. **For template hardening:** Apply round-1 patches; add code-gen eval for PEM validation depth and `SSL_CERT_DIR` when CA mount tasks are generated.
4. **Archive:** Change already at `openspec/changes/archive/2026-07-02-rfe-8685` — pilot complete.

---

## What OpenSpec would have gotten from EP/PRs (hypothetical Track A)

If `inputs/jira-spec.md` had been replaced with EP #2011 + PR #145/149 summaries (not allowed in this fair eval):

| Item | Likely outcome |
|------|----------------|
| API field name/path | **Correct** (`trustedCABundle` on `controllerConfig`) |
| Controller-only scope | **Correct** |
| `SSL_CERT_DIR` + `user-certs` path | **Correct** |
| PEM validation depth | **Correct** |
| Cross-ns ConfigMap | **Avoided** (operand-ns only) |
| Implementation LOC estimate | Closer to **PR #145 + subset of #149** (~800–1,200 core) vs greenfield ~800 |

Fair greenfield eval **correctly measured** discovery cost when EP/PRs are withheld.

---

## References

- Greenfield implementation: working tree on `release-1.1` (+T1_1–T4_2)
- Production: [enhancements#2011](https://github.com/openshift/enhancements/pull/2011), [ESO#145](https://github.com/openshift/external-secrets-operator/pull/145), [ESO#149](https://github.com/openshift/external-secrets-operator/pull/149)
- Merge commits: `3c9e90af` (API), `59a41dd3` (controller)
- Eval-loop: `evals/baseline/rounds/round-1/`
- Pilot template: `docs/openspec-eval-pilot-jira.md`
