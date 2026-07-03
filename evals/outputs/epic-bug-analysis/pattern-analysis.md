# Pattern Analysis — RFE-8685

**Round:** 1  
**Feature:** RFE-8685 — Custom Enterprise CA Bundle Integration  
**Evidence:** `evals/inputs/01-ep-ard.md`, `04-user-stories.md`, `openspec/changes/rfe-8685/specs.md`, `plan.md`, `tasks.md`

## Requirement → spec → plan → tasks mapping

| EP / Jira intent | Spec coverage | Plan phase | Tasks |
|------------------|---------------|------------|-------|
| API CA reference on ESC | FR-001, FR-002, US-1 | Phase 1 API | T1_1, T1_2 |
| Controller/webhook mounts | FR-004, FR-008, A-011 | Phase 3 Controller | T3_1, T3_2 |
| Store caProvider (no CRD work) | FR-003, US-2, A-007 | Phase 5 Docs | T5_1 |
| Degraded on missing CM | FR-005, FR-009, edge cases | Phase 4 Status | T4_1 |
| No manual volume mounts | FR-010, SC-006 | Phase 3 + 5 | T3_1, T5_1 |
| Proxy coexistence | FR-008, US-4, A-011 | Phase 3 | T3_1, T4_1 |

## ARD / spec structure strengths

- Clear **operator vs upstream** split (A-007): ESC field vs store `caProvider` documented separately.
- **A-011 matrix** resolved in plan before implementation (proxy × enterprise four states).
- **Repo-grounded** assessment on release-1.1 prevented library-go / SSA-first anti-patterns.
- Edge cases explicitly call out missing CM, invalid PEM, and removal paths.

## Story carving gaps

| Gap | Impact | Stage to catch |
|-----|--------|----------------|
| T1_2 (`make test-apis`) listed before T2_1 (manifests) but envtest requires current CRD | First test-apis run failed with `unknown field` | **tasks** — verification pairing order |
| Cross-namespace ConfigMap ref in API/testsuite implies direct mount | Implementation needed sync CM pattern (not in EP) | **plan**, **repo-assessment** |
| Jira mentions SecretStore CRD changes | Correctly scoped to docs-only in plan; risk of agent over-scoping | **validation**, **repo-assessment** |
| E2E optional but no cluster in apply session | T6_1 skipped by user directive | **tasks** (acceptable) |

## EP content vs stories

- **In EP, fully covered:** operator API + ConfigMap ref, controller/webhook, security (no manual mounts).
- **In EP, deferred to docs:** upstream store `caProvider` configuration path.
- **Beyond EP (justified):** `ConfigurationError` reconcile reason, projected volume merge, operand-namespace sync ConfigMap.

## Recurrence (round 1)

All patterns are **new** for baseline round 1.
