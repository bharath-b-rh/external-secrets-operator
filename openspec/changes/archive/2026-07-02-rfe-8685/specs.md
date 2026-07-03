# Feature Specification: Custom Enterprise CA Bundle Integration

**Feature Branch**: `rfe-8685-custom-ca-bundle`

**Created**: 2026-07-02

**Status**: Draft

**Input**: User description: "Enable External Secrets Operator to trust Enterprise PKI via ConfigMap CA injection, resolving x509 unknown authority errors for internal services, without unsupported manual volume mounts."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Trust internal vault from operator workloads (Priority: P1)

As a cluster administrator, I need the External Secrets Operator controller and webhook to trust my organization's enterprise PKI so that secret synchronization succeeds against internal certificate authorities (such as IBM/Thycotic Secret Server) without TLS handshake failures.

**Why this priority**: The customer cannot synchronize secrets today because the operator rejects certificates signed by unknown authorities. This is the core business blocker.

**Independent Test**: Configure an enterprise CA bundle reference, create a secret store targeting an internal vault signed by that CA, and verify an external secret sync completes without x509 trust errors while operator-managed workloads remain healthy.

**Acceptance Scenarios**:

1. **Given** a cluster running External Secrets Operator with no custom enterprise CA configured, **When** a secret store targets an internal vault whose server certificate chains to an enterprise CA not in the default trust store, **Then** synchronization fails with a certificate trust error and the failure is observable to the administrator.
2. **Given** a valid ConfigMap containing the enterprise CA certificate bundle referenced through the supported operator configuration, **When** the operator reconciles and the secret store targets the same internal vault, **Then** synchronization succeeds without x509 unknown authority errors.
3. **Given** enterprise CA configuration is enabled, **When** an administrator inspects operator-managed controller and webhook workloads, **Then** the enterprise CA trust material is applied through supported operator mechanisms and not through ad-hoc manual volume mount edits.

---

### User Story 2 - Declare provider CA trust at the store level (Priority: P2)

As a platform or application administrator, I need to reference a trusted CA bundle when configuring a secret store or cluster-scoped secret store so that provider connections verify server certificates using enterprise PKI rather than only default public trust.

**Why this priority**: The request explicitly includes store-level CA references for server verification, enabling per-provider trust without cluster-wide workarounds.

**Independent Test**: Configure a store with a CA reference pointing at enterprise trust material, then verify provider connectivity and secret retrieval against an internally signed endpoint.

**Acceptance Scenarios**:

1. **Given** a secret store configuration that includes a CA reference for server verification, **When** the store connects to a provider endpoint signed by the referenced enterprise CA, **Then** the store becomes ready and secret synchronization proceeds successfully.
2. **Given** a cluster-scoped secret store with a CA reference, **When** external secrets in multiple namespaces consume that store, **Then** all dependent synchronizations use the declared trust for provider TLS verification.

---

### User Story 3 - Manage CA trust safely through supported configuration (Priority: P2)

As a cluster administrator responsible for security compliance, I need to add, update, or remove enterprise CA trust through declarative operator and store configuration so that I never rely on unsupported manual pod volume changes.

**Why this priority**: Manual volume mounts are explicitly discouraged for security reasons and must be replaced by a first-class, auditable configuration path.

**Independent Test**: Enable enterprise CA trust via supported API fields, confirm workloads receive updated trust, then disable or change the reference and observe predictable reconciliation behavior.

**Acceptance Scenarios**:

1. **Given** enterprise CA trust is configured through supported fields, **When** an administrator removes or clears the CA reference, **Then** the operator reverts to prior trust behavior on the next reconciliation without leaving orphaned manual mounts.
2. **Given** the enterprise CA ConfigMap content is updated with additional CA certificates, **When** reconciliation runs, **Then** operator workloads and affected store connections trust the updated bundle without requiring unsupported pod spec edits.

---

### User Story 4 - Coexist with platform-injected trust when proxy is configured (Priority: P3)

As a cluster administrator on OpenShift with cluster-wide proxy settings, I need custom enterprise CA trust to work alongside existing platform trusted CA injection so that both proxy-related and enterprise PKI endpoints remain reachable.

**Why this priority**: The ticket states platform trusted CA injection currently applies only when a cluster-wide proxy is configured; enterprise PKI must not break that existing behavior.

**Independent Test**: Enable cluster proxy with platform CA injection, add enterprise CA configuration, and verify connectivity to both proxy-dependent and enterprise-signed endpoints.

**Acceptance Scenarios**:

1. **Given** cluster-wide proxy is configured and platform trusted CA material is injected into operator workloads, **When** enterprise CA configuration is also enabled, **Then** both platform and enterprise trust material are available for outbound TLS verification without manual intervention.
2. **Given** enterprise CA configuration is disabled while proxy remains configured, **When** reconciliation completes, **Then** platform-injected trust continues to function as before.

---

### Edge Cases

- **When** the referenced ConfigMap does not exist at reconciliation time, **then** the operator reports a degraded or not-ready status with a clear message identifying the missing trust source and secret synchronization against affected stores does not silently proceed with incorrect trust.
- **When** the ConfigMap exists but contains invalid or empty certificate data, **then** configuration is rejected or the operator surfaces a validation error without applying partial trust that could mask misconfiguration.
- **When** the ConfigMap is deleted after enterprise CA trust was active, **then** the operator detects the loss on subsequent reconciliation, surfaces degraded status, and does not continue as if enterprise trust were still present.
- **When** both platform proxy CA injection and enterprise CA configuration are active, **then** trust for outbound connections considers both sources rather than replacing one with the other unless explicitly configured otherwise.
- **When** an administrator migrates from an unsupported manual volume mount workaround to supported CA configuration, **then** supported configuration takes precedence on reconcile and no duplicate conflicting trust mounts remain on operator-managed workloads.
- **When** enterprise CA configuration is removed, **then** operator-managed workloads return to default and platform-injected trust behavior without requiring manual pod deletion beyond normal reconciliation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow cluster administrators to reference a ConfigMap containing trusted enterprise CA certificates for External Secrets Operator-managed controller and webhook workloads through supported operator configuration.
- **FR-002**: System MUST apply referenced enterprise CA trust to operator-managed workloads through reconciled, declarative configuration rather than requiring manual pod volume mount edits.
- **FR-003**: System MUST allow secret store and cluster-scoped secret store configurations to include a CA reference used for provider server certificate verification.
- **FR-004**: System MUST enable successful TLS handshakes and secret synchronization against internal services signed by enterprise PKI when the correct CA reference is configured, resolving x509 unknown authority failures described in the request.
- **FR-005**: System MUST reject or surface clear errors for invalid CA references, including missing ConfigMaps and malformed certificate bundle content, before or during reconciliation without silent fallback to incorrect trust.
- **FR-006**: System MUST update operator-managed workload trust when the referenced ConfigMap content changes, without manual pod spec intervention.
- **FR-007**: System MUST revert enterprise CA trust effects when the CA reference is removed or disabled, restoring prior default and platform-injected trust behavior on reconciliation.
- **FR-008**: System MUST preserve existing platform trusted CA injection behavior for clusters with cluster-wide proxy configured when enterprise CA trust is added, updated, or removed.
- **FR-009**: System MUST surface observable health or degraded status when enterprise CA prerequisites are unmet (missing ConfigMap, invalid data, or reconcile failure).
- **FR-010**: System MUST restrict enterprise CA configuration to supported API fields so administrators cannot inject arbitrary filesystem content through unsupported mount patterns.

### Key Entities *(include if feature involves data)*

- **Enterprise CA ConfigMap**: A Kubernetes ConfigMap holding one or more PEM-encoded CA certificates representing the customer's enterprise PKI trust anchor(s); referenced by operator or store configuration.
- **Operator CA Trust Configuration**: Cluster-level settings that bind enterprise CA trust to External Secrets Operator-managed controller and webhook workloads.
- **Store CA Reference**: A provider-facing trust declaration on a secret store or cluster-scoped secret store used when verifying remote server certificates during secret retrieval.
- **Platform Trusted CA Bundle**: Cluster-provided trust material injected into workloads when cluster-wide proxy settings apply on OpenShift.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After configuring enterprise CA trust through supported operator settings, an administrator can synchronize secrets from an internal vault signed by the enterprise CA without x509 unknown authority errors in operator logs or store status.
- **SC-002**: At least 95% of validation attempts with a correctly referenced enterprise CA ConfigMap result in successful secret store readiness and external secret sync on first reconcile cycle under normal cluster conditions.
- **SC-003**: When a referenced ConfigMap is missing or contains invalid certificates, administrators receive an observable degraded or error status within one reconciliation cycle rather than silent sync failure without explanation.
- **SC-004**: Removing enterprise CA configuration restores pre-configuration trust behavior on operator-managed workloads without manual pod edits in standard upgrade and reconfigure scenarios.
- **SC-005**: On clusters with cluster-wide proxy and platform CA injection enabled, enterprise CA configuration can be added and removed without loss of proxy-related trust for endpoints that depended on platform-injected CAs before the change.
- **SC-006**: Zero supported-configuration workflows require administrators to add custom volume mounts directly to operator operand pod templates.

## Assumptions

- **A-001**: Cluster administrators manage operator-level enterprise CA trust; platform or application administrators manage store-level CA references for their providers.
- **A-002**: Enterprise CA trust supplements existing default and platform-injected trust rather than replacing it unless an administrator explicitly clears other trust sources through supported configuration.
- **A-003**: The ConfigMap reference includes sufficient identity information (name and namespace) for the operator to locate enterprise CA material within the cluster.
- **A-004**: Enterprise CA material is provided as PEM-encoded certificates suitable for TLS server verification; mutual TLS client certificate configuration is out of scope unless later specified.
- **A-005**: Both operator workload trust and store-level CA references are in scope because the request lists controller/webhook deployments and secret store resources as affected components.
- **A-006**: cert-manager for Red Hat OpenShift is an environmental dependency for existing webhook certificate management but does not require new integration work as part of this feature unless future clarification expands scope.
- **A-007**: [NEEDS CLARIFICATION: Should store-level CA references reuse the same ConfigMap as operator-level trust, or may stores reference independent ConfigMaps per provider?]
- **A-008**: Unsupported manual volume mount workarounds are not migrated automatically; administrators must adopt supported configuration and allow reconciliation to converge operand pods.
- **A-009**: When a ConfigMap is updated, reconciliation propagates trust changes without requiring cluster downtime beyond normal rolling updates of operator-managed workloads.
- **A-010**: Minimum OpenShift and operator versions supported by the release-1.1 branch are sufficient targets; hypershift-specific or multi-cluster federation behavior is out of scope unless explicitly requested later.
- **A-011**: [NEEDS CLARIFICATION: When cluster-wide proxy is not configured, should platform trusted CA injection still be combined with enterprise CA trust, or is enterprise CA the sole supplemental trust source in that case?]
- **A-012**: RBAC for referencing enterprise CA ConfigMaps follows standard namespace access controls; only roles with ConfigMap read permission in the referenced namespace can validate configuration successfully.
