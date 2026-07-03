# Template Gaps — Round 1 (RFE-8685)

**Feature:** RFE-8685  
**Taxonomy:** `evals/outputs/epic-bug-analysis/issue-taxonomy.json`

| Template | Gap | Pattern | Resolution | Fixed |
|----------|-----|---------|------------|-------|
| repo-assessment-template.md | Missing trusted CA / proxy-gate baseline inventory | PAT-002, PAT-006 | patchable | Yes |
| plan-template.md | No guidance for A-011 matrix or cross-ns CM mount strategy | PAT-002, PAT-003 | patchable | Yes |
| tasks-template.md | test-apis can run before manifests regen | PAT-004, ISSUE-001 | patchable | Yes |
| validation-template.md | Operator vs upstream CRD scope not checked for CA features | PAT-001 | patchable | Yes |
| constitution-template.md | ConfigurationError / Degraded pairing for user config | ISSUE-003 | eval-only | N/A |
| implementation-template.md | Projected volume / sync ConfigMap patterns | PAT-003 | eval-only | N/A |
| spec-template.md | Cross-ns mount semantics implicit in API | PAT-003 | deferred | SME |

## Code generation evals

| Gap | Resolution |
|-----|------------|
| No api-generate-tests eval for CEL multi-error format | eval-r001-codegen-002 |
| No api-implement eval for A-011 volume matrix | eval-r001-codegen-003 |
| No api-generate eval for CommonConfigs multi-CRD | eval-r001-codegen-001 |

Minimum 2 cases per represented oape_command: **met** (api-generate, api-generate-tests, api-implement).

## Validation refinements

See `evals/outputs/eval-generation/validation-refinements.md`.
