# Jira Epic — RFE-8685

**Key:** RFE-8685  
**Title:** Support for Custom CA Bundle Integration via ConfigMap for External Secrets Operator (ESO)

## Problem

Customer cannot synchronize secrets because ESO fails TLS handshake with internal vault (`x509: certificate signed by unknown authority`). Platform CA injection at `/etc/pki/tls/certs` only occurs when cluster-wide proxy is configured.

## Proposed solution

a. Extend API with CA reference for server verification (`additionalTrustedCAConfigMapRef` on ESC `appConfig`).  
b. Reference ConfigMap containing enterprise PKI CA certificates.

## Business requirements

- Enterprise connectivity to internal self-signed services
- TLS verification without unknown authority errors
- Security compliance — no unsupported manual volume mounts

## Workflow artifacts produced

| Stage | Artifact |
|-------|----------|
| validation | validation.json |
| specs | specs.md (FR-001–FR-010, A-007, A-011) |
| repo-assessment | repo-assessment.md (release-1.1 pin) |
| plan | plan.md (7 phases) |
| tasks | tasks.md (13 tasks; T6_1 optional e2e) |
| implementation | T1_1–T4_2 completed; T5–T7 deferred |
