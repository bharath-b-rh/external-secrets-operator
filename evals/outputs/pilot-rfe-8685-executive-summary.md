# Executive Summary — OpenSpec SDLC Pilot (RFE-8685)

**Audience:** Engineering leadership, product, platform teams  
**Date:** July 2026  
**Feature:** Custom enterprise CA bundle for External Secrets Operator (RFE-8685)  
**Evaluation type:** Track B greenfield — scope-locked Jira only, no EP/PR spoilers, `release-1.1` base  
**Full report:** [`pilot-rfe-8685.md`](pilot-rfe-8685.md)

---

## Purpose

We evaluated whether the **openspec-agile-workflow** toolchain can take a real customer RFE from specification through working code on a branch that does **not** yet contain the feature — without leaking design from enhancement proposals or merged PRs.

Production reference (shipped separately):

| Artifact | Link |
|----------|------|
| Enhancement proposal | [openshift/enhancements#2011](https://github.com/openshift/enhancements/pull/2011) |
| API (ESO-435) | [external-secrets-operator#145](https://github.com/openshift/external-secrets-operator/pull/145) |
| Controller + e2e (ESO-396) | [external-secrets-operator#149](https://github.com/openshift/external-secrets-operator/pull/149) |

---

## Bottom line

| Question | Answer |
|----------|--------|
| **Did OpenSpec produce a usable delivery path?** | **Yes** — validation → specs → repo assessment → plan → tasks → implementation, with zero scope violations. |
| **Is the generated code production-ready?** | **No** — it is a credible prototype (~800 LOC) that diverges from the shipped design in API shape, TLS behavior, and validation depth. |
| **Should we adopt OpenSpec for the next story?** | **Yes, with guardrails** — recommended **4/5**; pair with EP or SME review before `/opsx-apply` on customer-facing APIs. |

---

## What we ran

- **Input:** Thin Jira text only (`jira-spec.md`); no Jira MCP, no linked tickets, no reads from `main` or GitHub during the run.
- **Base:** `release-1.1` (feature absent).
- **Completed:** Full spec pipeline + implementation tasks T1_1–T4_2 (API, codegen, controller, unit tests).
- **Stopped intentionally:** Docs, e2e, and release tasks (T5–T7).
- **Follow-up:** `/eval-loop` round 1 — 19 eval cases, 4 workflow templates improved.

---

## Results at a glance

| Area | Outcome |
|------|---------|
| Spec quality (validation score) | **56%** — honestly reflects thin Jira; no blockers |
| Repo grounding | **Strong** — correct operator vs upstream CRD split; no false “feature exists” claims |
| Implementation gate | **`make test-unit` passed** after 8/13 tasks |
| Scope discipline | **0 violations** (no MCP, no cross-branch reads) |
| vs production concept | **~55–60% aligned** with EP + PR chain |
| Rework to match shipped solution | **~900–1,200 LOC** (incl. e2e/docs) |
| Risk if prototype merged as-is | **4–6 likely defects** (TLS path, PEM validation, proxy/CNO edge cases) |

---

## Greenfield vs what we actually shipped

| Topic | OpenSpec greenfield | Production (EP #2011, PR #145, #149) |
|-------|---------------------|----------------------------------------|
| API field | `appConfig.additionalTrustedCAConfigMapRef` (+ namespace) | `controllerConfig.trustedCABundle` (operand namespace only) |
| Workloads | Controller **and** webhook | **Core controller only** |
| TLS integration | Shared mount at `/etc/pki/tls/certs` | Separate `user-certs` path + `SSL_CERT_DIR` |
| PEM validation | Basic string check | Full x509 CA parsing, events, private-key rejection |
| Cluster proof | None (e2e skipped) | Full e2e suite in PR #149 |
| Size | ~800 hand-written LOC | ~3,600+ LOC across API + controller PRs |

**What greenfield got right:** operator-owned feature (not upstream SecretStore CRDs), no manual volume-mount workarounds, proxy coexistence considered, cross-namespace ConfigMap handled via operand-namespace sync (valid Kubernetes pattern, but not the chosen product design).

**What it missed without EP/PR context:** exact API placement and naming, controller-only scope, `SSL_CERT_DIR` semantics, CNO-inject skip when proxy is on, and production-grade validation and e2e.

---

## Value demonstrated

1. **Structured progression** from ambiguous RFE to repo-grounded plan and executable tasks without ad-hoc prompting.
2. **Fair evaluation** — measurable gap between “inferred from Jira + repo” and “designed with EP + review.”
3. **Institutional learning** — eval-loop captured workflow friction (e.g. run `make manifests` before `make test-apis`) and patched templates for the next feature.
4. **Time efficiency** — compact `/opsx-apply — continue next task` after scope lock; spec pipeline completable without full implementation.

---

## Recommendations

| For… | Recommendation |
|------|----------------|
| **Next greenfield story** | Use OpenSpec with scope constraints; stop at T4_2 for a planning/implementation pilot or run through e2e for full proof. |
| **Customer-facing API features** | Add **EP or API sketch** to inputs before `/opsx-apply` — do not rely on Jira alone for field names and mount semantics. |
| **RFE-8685 production** | **Do not merge** greenfield API/controller; continue with shipped PR chain (#145 + #149). |
| **Tooling investment** | Apply round-1 template patches; add code-generation evals for PEM validation and `SSL_CERT_DIR` patterns. |

---

## Suggested adoption model

```text
Thin Jira  →  OpenSpec spec pipeline  →  EP/SME gate on API + plan  →  /opsx-apply  →  /eval-loop
```

OpenSpec is **strong at discovery and decomposition**; **human/EP alignment** remains required for API contracts and security-sensitive controller behavior.

---

**Contact / artifacts:** `openspec/changes/archive/2026-07-02-rfe-8685`, `evals/baseline/rounds/round-1/`, `evals/outputs/pilot-rfe-8685.md`
