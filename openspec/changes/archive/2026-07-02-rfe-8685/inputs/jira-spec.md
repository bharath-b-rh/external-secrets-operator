# RFE-8685 — scope-locked spec source

## Summary

Enable External Secrets Operator to trust Enterprise PKI via ConfigMap CA injection, resolving "x509: unknown authority" errors for internal services. This replaces unsupported manual volume mounts.

## Description

1. Proposed Title
Feature Request: Support for Custom CA Bundle Integration via ConfigMap for External Secrets Operator (ESO)

2. Nature and Description
The request is for a downstream enhancement to the External Secrets Operator (ESO) on Red Hat OpenShift Container Platform (OCP).
Currently, the operator only injects the OpenShift trusted CA bundle into the well-known path (/etc/pki/tls/certs) if a cluster-wide proxy is configured. The customer requires a mechanism to provide a custom CA bundle (from their enterprise PKI) to the ESO controller/webhook without manually adding volume mounts, which is currently unsupported and discouraged for security reasons.
The proposed technical solution includes:
a. Extending the API to include a CA reference for server verification.
b. Adding an option to reference a ConfigMap containing the trusted CA certificates.

3. Business Requirements
The customer is unable to synchronize secrets because the ESO cannot establish a secure TLS handshake with their internal vault.
a. Enterprise Connectivity: Enable ESO to communicate with customer's internal, self-signed services (specifically IBM/Thycotic Secret Server).
b. TLS Verification: Resolve the x509: certificate signed by unknown authority error by allowing the controller to trust the Enterprise PKI.
c. Security Compliance: Avoid "custom volume mounts" in the operator spec to prevent arbitrary data injection and maintain the integrity of the pod filesystem.

4. Affected Packages or Components
External Secrets Operator (ESO): Specifically the controller and webhook deployments.
Custom Resource Definitions (CRDs): SecretStore and ClusterSecretStore (to include CA references).
cert-manager Operator for Red Hat OpenShift: Used for managing the External Issuer.
OCP Trusted CA Bundle Injection: The logic governing the automated injection of /etc/pki/tls/certs.

---
**Out of scope for this change:** linked issues, epic, subtasks, comments, acceptance criteria from other tickets, Confluence, PRs.
