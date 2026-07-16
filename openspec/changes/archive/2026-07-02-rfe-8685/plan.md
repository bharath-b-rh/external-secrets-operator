# Technical Implementation Plan

**Feature:** RFE-8685 — Custom Enterprise CA Bundle Integration

## 0. Inputs acknowledged

| Input | Status |
|-------|--------|
| Spec source | RFE-8685 — `openspec/changes/rfe-8685/specs.md` (approved) |
| Repo assessment pin | `/workspaces/external-secrets-operator`, branch `release-1.1`, commit `1f54b74a6fbe2c327c9aa2d1f4f7662cce1871ea` (tooling_status: OK) |
| `agents.md` | PROVIDED — `/workspaces/external-secrets-operator/AGENTS.md` |
| `spec_validator_results.json` | PROVIDED — `openspec/changes/rfe-8685/validation.json` (NEEDS_REVISION, no blockers) |
| `constitution.md` | PROVIDED — `/workspaces/external-secrets-operator/constitution.md` |
| **AgentRoutingMode** | **PROVIDED** (per constitution.md) |
| Scope constraints | `openspec/changes/rfe-8685/inputs/scope-constraints.md` — working-folder only, no Jira MCP |

**Clarifications resolved in this plan (from specs A-007, A-011):** see §1 and §8.

---

## 1. Architectural strategy

### Repo-grounded reality check

Per `repo-assessment.md` §0 and §11.2 on `release-1.1`:

| Capability | Branch state | Plan stance |
|------------|--------------|-------------|
| Operator API field for enterprise CA ConfigMap | **Absent** | **Greenfield** on `ExternalSecretsConfig` |
| Operand controller/webhook CA volume reconcile | **Partial** — proxy-gated CNO injection only | **Delta** — extend `configmap.go` / `deployments.go` |
| Upstream `SecretStore` / `ClusterSecretStore` `caProvider` | **Present** in shipped CRD v0.20.4 | **Configuration + docs** — no operator CRD/bindata schema work |
| E2E `Feature:TrustedCABundle` | **Label documented, no test file** | **Greenfield** e2e in later phase |

This is primarily **greenfield operator API + controller reconcile** layered on existing proxy/CNO trusted-CA machinery. It is **not** an upstream CRD or bindata tier-4 change.

### Strategy summary

1. **Operator-owned path (P1):** Add an optional enterprise CA ConfigMap reference on `ExternalSecretsConfig.spec.appConfig` (embedded via existing `ApplicationConfig` → `CommonConfigs` pattern). Reconcile mounts onto **core controller** and **webhook** deployments only (`deployment_external-secrets.yml`, `deployment_external-secrets-webhook.yml` assets). Do not mount on cert-controller unless a later SME expands scope.

2. **Store-owned path (P2 — no operator code):** Document that administrators configure existing upstream `caProvider` (`type: ConfigMap`, `name`, `key`, `namespace`) on `SecretStore` / `ClusterSecretStore` for provider TLS. The upstream operand controller consumes that field; this operator does not propagate ESC settings into store specs.

3. **Constitution alignment:** No upstream sync logic fork (Principle I). Create-or-update deployment mutation only (Principle IV). No SSA-first operand changes (Principle IV). No edits to upstream `external-secrets.io` CRD YAML (Principle VIII — tier-4 comes from helm bump only).

4. **Decouple enterprise CA from proxy gate:** Today `ensureTrustedCABundleConfigMap` and `updateTrustedCABundleVolumes` require `getProxyConfiguration() != nil`. Enterprise CA reconcile runs when the new API field is set, independent of proxy.

### Decision: A-007 (shared vs per-store ConfigMap)

| Layer | ConfigMap binding | Coupling |
|-------|-------------------|----------|
| **Operator (ESC)** | One optional `configMapRef` on `ExternalSecretsConfig.spec.appConfig` — cluster admin chooses name + namespace (+ optional key) | Operator reads/writes **operand pod volumes only** |
| **Store (ClusterSecretStore / SecretStore)** | Per-store `caProvider` in upstream CRD — app team chooses ConfigMap per provider | **Independent** of ESC field |
| **Same object allowed?** | Yes — admin may reference the **same** ConfigMap in both ESC and store `caProvider`, but the operator **does not** auto-sync or enforce sameness | Documented convention only |

**Rationale:** FR-001/002 are operand-pod trust; FR-003 is provider TLS inside upstream controller. Different reconciliation owners; coupling would violate Principle I and expand scope.

### Decision: A-011 (proxy absent — platform CA behavior)

| Cluster state | Platform CNO CA (`external-secrets-trusted-ca-bundle`) | Enterprise user ConfigMap |
|---------------|--------------------------------------------------------|---------------------------|
| **Proxy absent, enterprise unset** | Not created, not mounted (unchanged) | Not mounted |
| **Proxy absent, enterprise set** | Not created, not mounted | **Sole** supplemental trust volume mounted to operand controller + webhook |
| **Proxy present, enterprise unset** | Created (inject label) + mounted (unchanged) | Not mounted |
| **Proxy present, enterprise set** | Created + mounted **and** enterprise mounted — use **projected volume** merging both ConfigMap sources into one read-only mount at `/etc/pki/tls/certs` on affected containers | Same mount path via projection |

**Rationale:** Matches code truth (`configmap.go:24-29`) and specs FR-008 (preserve proxy behavior) + A-002 (supplement, not replace). Projected volume avoids two mounts on the same path and keeps Go system-cert-directory semantics.

When enterprise config is removed, reconcile removes enterprise source from projection (or entire volume if proxy-only / enterprise-only).

---

## 2. Persistence & state

| Object | Role | Notes |
|--------|------|-------|
| `ExternalSecretsConfig` (`cluster`) | Source of truth for enterprise CA reference | New optional field; singleton CEL unchanged |
| User enterprise `ConfigMap` | User-managed PEM data | Operator **reads** only; never writes CA bytes |
| `external-secrets-trusted-ca-bundle` ConfigMap | Platform CNO-injected (proxy only) | Existing; operator ensures metadata/labels only |
| Operand `Deployment` (controller, webhook) | Derived — volume/volumeMount mutations | `HasObjectChanged` drift detection |
| `ExternalSecretsConfig.status.conditions` | Derived — Ready/Degraded | Extend messages when enterprise ConfigMap missing/invalid |

No new CRD kinds. No bindata content change for base deployments (volumes applied at reconcile). No new operator feature flag on ESM for this feature unless a follow-on requires global default.

---

## 3. Interfaces & contracts (operator-native)

### 3.1 Kubernetes APIs (CRDs/CRs)

**New field (proposed shape — derived from release-1.1 patterns, not other branches):**

Add to `CommonConfigs` in `api/v1alpha1/meta.go` (inherited by `ApplicationConfig` on ESC and optionally later ESM `GlobalConfig`):

```text
spec.appConfig.additionalTrustedCAConfigMapRef   (optional)
  name      string   (required when ref present; MaxLength 253)
  namespace string   (required when ref present; DNS-1123 subdomain)
  key       string   (optional; default "ca-bundle.crt" or single-key ConfigMap semantics)
```

**Design notes (release-1.1 patterns):**

- Follow `SecretReference` / `ProxyConfig` kubebuilder validation style in `meta.go`.
- Field lives under `appConfig` (not `controllerConfig`) because it configures **operand application trust**, consistent with `proxy` on `CommonConfigs`.
- CEL: if `additionalTrustedCAConfigMapRef` is present, `name` and `namespace` must be non-empty.
- **Out of scope:** ESM mirror field in initial delivery unless plan phase explicitly adds parity with `proxy` precedence (ESC > ESM); defer to Phase 6 optional.
- **Out of scope:** Upstream `SecretStore` / `ClusterSecretStore` schema changes — `caProvider` already exists in `config/crd/bases/customresourcedefinition_clustersecretstores.external-secrets.io.yml`.

**Immutability:** No immutability marker on first introduction; standard reconcile updates allowed.

### 3.2 Controller/runtime interfaces (internal)

| Function / area | Change |
|-----------------|--------|
| `getEnterpriseCAConfigMapRef(esc)` | New helper — read ESC `appConfig.additionalTrustedCAConfigMapRef` |
| `validateEnterpriseCAConfigMap(ref)` | Fetch ConfigMap; verify key exists and PEM non-empty; return user-facing error |
| `ensureTrustedCABundleConfigMap` | Unchanged proxy gate for CNO ConfigMap |
| `updateTrustedCABundleVolumes` | Refactor to `updateOperandTrustedCAVolumes` — handle proxy-only, enterprise-only, both (projected), neither |
| `processReconcileRequest` / status | Set `Degraded=True` when enterprise ref set but ConfigMap invalid/missing (FR-009) |
| ConfigMap watch | Existing ConfigMap watch in `controller.go` — ensure user enterprise CM changes enqueue ESC reconcile |

Use `CtrlClient` interface; wrap errors with `common.FromClientError`. Missing ConfigMap → retryable reconcile error + degraded status (not irrecoverable unless malformed ref in spec caught by API validation).

### 3.3 Webhooks / admission (if applicable)

N/A — no new admission webhook. Operand **serving** TLS remains cert-manager or in-tree cert-controller (constitution Principle VII). This feature affects **outbound** trust from controller/webhook containers only.

### 3.4 RBAC / security boundaries (if applicable)

- Operator may need **get/list/watch** on ConfigMaps in user-specified namespaces (cross-namespace read). Evaluate existing operator ClusterRole in `config/rbac/role.yaml`; add `configmaps` get/list/watch if not sufficient for referenced namespaces.
- Blast radius: read-only access to user ConfigMaps explicitly referenced in ESC; no cluster-wide ConfigMap write.
- Users must hold RBAC to create ESC with namespace reference (standard Kubernetes authorization).

### 3.5 Packaging / OLM (if applicable)

- Regenerate CRD YAML and bundle after API markers: `make manifests && make bundle`.
- CSV `spec.customresourcedefinitions` descriptors for new field (human-readable x-descriptors if used elsewhere).
- No new `RELATED_IMAGE_*` env vars.

---

## 4. Dependencies & sequencing graph

**Critical path:**

```text
Phase 1 (API types + validation tests)
  → Phase 2 (codegen/manifests/docs)
  → Phase 3 (controller volume + ConfigMap validation)
  → Phase 4 (status conditions + unit tests)
  → Phase 5 (docs: store caProvider configuration guide)
  → Phase 6 (e2e Feature:TrustedCABundle)
  → Phase 7 (OLM bundle verify + final verify)
```

**Parallelizable after Phase 2:**

- Phase 5 (docs) can start after Phase 1 API shape is stable.
- RBAC review (Phase 3) can proceed in parallel with volume logic.

**Explicit blockers:**

- None external to repo. A-007 and A-011 resolved in this plan.

---

## 5. Implementation phases (logical sequence; NOT tasks)

### Phase 1: Operator API — enterprise CA ConfigMap reference

- **Goal:** Introduce typed, validated `additionalTrustedCAConfigMapRef` on ESC `appConfig` per §3.1.
- **Dependencies:** None.
- **Target files:**
  - `api/v1alpha1/meta.go`
  - `api/v1alpha1/external_secrets_config_types.go` (if doc comments only)
  - `api/v1alpha1/tests/externalsecretsconfig.operator.openshift.io/externalsecretsconfig.testsuite.yaml`
- **Required capabilities:** **API_Agent**
- **Verification hooks:** `make test-apis`; `go build` / `go vet` on `api/v1alpha1/`

### Phase 2: Code generation and manifest cascade

- **Goal:** Regenerate deepcopy, CRD bases, API reference, OLM bundle inputs.
- **Dependencies:** Phase 1 merged/stable markers.
- **Target files:**
  - `api/v1alpha1/zz_generated.deepcopy.go` (generated)
  - `config/crd/bases/operator.openshift.io_externalsecretsconfigs.yaml` (generated)
  - `docs/api_reference.md` (generated)
  - `bundle/manifests/operator.openshift.io_externalsecretsconfigs.yaml` (via `make bundle`)
- **Required capabilities:** **API_Agent** + **OLMRelease_Agent** (bundle)
- **Verification hooks:** `make generate && make manifests && make verify-generated`; `make verify`

### Phase 3: Controller reconcile — volume projection and ConfigMap validation

- **Goal:** Mount enterprise CA on controller + webhook deployments; implement A-011 matrix (proxy absent/present × enterprise on/off); validate user ConfigMap on reconcile.
- **Dependencies:** Phases 1–2.
- **Target files:**
  - `pkg/controller/external_secrets/constants.go` (enterprise volume name constant, default key name)
  - `pkg/controller/external_secrets/utils.go` (getter for enterprise ref)
  - `pkg/controller/external_secrets/configmap.go` (keep CNO path; add validation helper or separate file)
  - `pkg/controller/external_secrets/deployments.go` (`updateTrustedCABundleVolumes` refactor)
  - `pkg/controller/external_secrets/install_external_secrets.go` (ordering if validation before deployments)
  - `pkg/controller/common/utils.go` (extend `HasObjectChanged` if projection shape differs)
  - `config/rbac/role.yaml` (if cross-namespace ConfigMap get needed)
- **Required capabilities:** **OperatorController_Agent**; **RBACSecurity_Agent** if RBAC change
- **Verification hooks:** `go test ./pkg/controller/external_secrets/...`

### Phase 4: Status, events, and unit test coverage

- **Goal:** FR-005/FR-009 — degraded Ready/Degraded when ConfigMap missing/invalid; events on successful mount updates; table tests for A-011 scenarios.
- **Dependencies:** Phase 3.
- **Target files:**
  - `pkg/controller/external_secrets/controller.go` (condition messages)
  - `pkg/controller/external_secrets/deployments_test.go` (extend proxy/CA tests: enterprise-only, proxy+enterprise projected, removal)
  - New `pkg/controller/external_secrets/configmap_test.go` or tests in existing `_test.go` if preferred
- **Required capabilities:** **OperatorController_Agent**
- **Verification hooks:** `make test-unit`; `make test`

### Phase 5: Configuration documentation — store-level `caProvider` (no CRD changes)

- **Goal:** FR-003 via documentation — how to configure existing upstream `caProvider` for internal vaults (IBM/Thycotic pattern); clarify independence from ESC field (A-007); example ClusterSecretStore YAML using ConfigMap CA.
- **Dependencies:** Phase 1 API shape stable (for cross-reference to ESC field).
- **Target files:**
  - `docs/` (new or extend integration/trusted-CA doc — single focused doc)
  - Optional: `config/console/` sample snippet if console samples exist for SecretStore
- **Required capabilities:** **Docs_Agent**
- **Verification hooks:** Manual review; no codegen

**Explicitly excluded:** `config/crd/bases/customresourcedefinition_*secretstores*`, `hack/update-external-secrets-manifests.sh`, bindata operand CRD refresh.

### Phase 6: End-to-end verification

- **Goal:** SC-001–SC-005 — cluster test for enterprise CA mount with/without proxy; label `Feature:TrustedCABundle` per `docs/testing-guidelines.md`.
- **Dependencies:** Phases 3–4.
- **Target files:**
  - `test/e2e/` (new spec file)
  - `test/utils/` (helpers if needed)
- **Required capabilities:** **Testing_Agent**
- **Verification hooks:** `make test-e2e E2E_GINKGO_LABEL_FILTER="Feature:TrustedCABundle"` (cluster required)

### Phase 7: Release hardening — bundle, verify, lint

- **Goal:** Constitution IX — full verify/lint; CSV complete.
- **Dependencies:** Phases 1–6.
- **Target files:**
  - `bundle/` (regenerated)
  - `config/manifests/bases/openshift-external-secrets-operator.clusterserviceversion.yaml` (if x-descriptors needed)
- **Required capabilities:** **OLMRelease_Agent**
- **Verification hooks:** `make update && make verify && make lint`

---

## 6. Verification matrix (maps to spec acceptance)

| Spec ID | Scenario | Phase | Category | Files / Suites |
|---------|----------|-------|----------|----------------|
| FR-001, FR-002, US-1 | ESC references enterprise ConfigMap; controller/webhook get mounts | 3, 4 | Unit | `deployments_test.go` |
| FR-003, US-2 | Store `caProvider` documented | 5 | Manual / Docs | new doc + sample YAML |
| FR-004, SC-001, US-1 | Sync without x509 against internal CA | 6 | E2E | `test/e2e/*` `Feature:TrustedCABundle` |
| FR-005, FR-009, SC-003 | Missing/invalid ConfigMap → Degraded | 4 | Unit | `controller.go` tests / deployment tests |
| FR-006 | ConfigMap data update → rolling reconcile | 4, 6 | Unit + E2E | deployment drift tests |
| FR-007, SC-004 | Remove enterprise ref → volumes revert | 4 | Unit | `deployments_test.go` |
| FR-008, SC-005, US-4, A-011 | Proxy + enterprise coexist via projection | 3, 4, 6 | Unit + E2E | `deployments_test.go`, e2e |
| FR-010 | No componentConfigs volume escape hatch | 3 | Review | constitution + code review |
| SC-006 | No manual volume mounts in supported workflow | 5 | Docs | operator doc states declarative-only |
| A-007 | Independent ESC vs store ConfigMap refs | 5 | Docs | explicit section in doc |

| Category | Coverage |
|----------|----------|
| Unit | Phases 3–4 — deployment volumes, ConfigMap validation, status |
| Integration (API) | Phase 1–2 — `make test-apis` |
| E2E | Phase 6 — TrustedCABundle (new) |
| Manual / Cluster | Phase 5 — store caProvider runbook; Phase 6 proxy cluster for FR-008 |
| N/A | Upstream CRD schema — already on branch; verified by inspection not new tests |

---

## 7. Risks, migrations, and operational follow-ups

### Upgrade / migration

- **Greenfield API field** — optional; existing clusters unchanged until ESC updated.
- Clusters using **unsupported manual volume mounts** (A-008): admin must remove manual edits and apply ESC field; operator reconcile converges on next loop.
- Removing proxy config after upgrade: existing TODO in `configmap.go` on CNO ConfigMap removal remains; enterprise path must not regress.

### Compatibility

- **OpenShift without cluster proxy:** Enterprise CA becomes only operand trust injection (A-011) — primary customer scenario.
- **OpenShift with proxy:** Projected volume must be tested on real cluster (CNO populates platform CM asynchronously).
- **Hypershift / MicroShift:** Not in spec scope (A-010); no special handling planned.

### Upstream API drift

- Store `caProvider` schema tied to `EXTERNAL_SECRETS_VERSION` (`v0.20.4` on branch). Tier-3 bump may change provider fields; docs should reference version pin.

### Operational risks

| Risk | Mitigation |
|------|------------|
| Projected volume merge ordering | Unit tests for volume spec shape; e2e with real ConfigMaps |
| Go TLS not loading mounted CAs | Mount at `/etc/pki/tls/certs` (existing constant); e2e sync test |
| Cross-namespace ConfigMap RBAC denied | Degraded status + clear message; doc RBAC prerequisites |
| cert-controller receives unintended mounts | Limit deployment asset list to controller + webhook only |

---

## 8. Open questions / SME decisions

**None — all decisions resolved in this plan.**

Resolved items (for task creation traceability):

| Former marker | Resolution | Section |
|---------------|------------|---------|
| **A-007** | ESC and store `caProvider` are independent; same ConfigMap object allowed by admin choice only | §1 |
| **A-011** | Proxy absent → enterprise only; proxy present → CNO + enterprise via projected volume at `/etc/pki/tls/certs` | §1, Phase 3 |

**Deferred (not blocking tasks):**

- ESM `globalConfig` parity for enterprise CA (mirror `proxy` precedence) — defer unless product requests cluster-wide default without ESC duplication.
- cert-controller deployment mounts — out of scope unless SME expands jira controller list.
- Thycotic-specific provider field verification — use generic `caProvider` documentation.
