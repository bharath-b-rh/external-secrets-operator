# Validation Refinements — Round 1

**Feature:** RFE-8685

Added validation rubric check for operator CA/trust features:

- Spec must separate operator ESC operand trust from upstream store `caProvider` configuration.
- Flag upstream CRD/bindata work when repo assessment shows fields already exist on branch.
- Require edge cases for missing/invalid ConfigMap and proxy+enterprise coexistence when TLS trust is in scope.

Patched: `evals/refined-templates/validation-template.md`
