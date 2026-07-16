# Repository Assessment Report

**Feature:** RFE-8685 — Custom Enterprise CA Bundle Integration

## 0. Inputs & Tooling

| Field | Value |
|-------|-------|
| **repo** | `/workspaces/external-secrets-operator` (working-folder mode) |
| **branch** | `release-1.1` |
| **commit** | `1f54b74a6fbe2c327c9aa2d1f4f7662cce1871ea` |
| **tooling_status** | OK |
| **spec** | Approved `openspec/changes/rfe-8685/specs.md` |
| **jira source** | `openspec/changes/rfe-8685/inputs/jira-spec.md` only |
| **operand version pin** | `EXTERNAL_SECRETS_VERSION ?= v0.20.4` (`Makefile` line 29) |

**Feature status on pinned branch:** Operator-level enterprise CA ConfigMap reference is **NOT implemented** on `release-1.1`. Proxy-gated platform trusted CA injection **is implemented**. Upstream `SecretStore` / `ClusterSecretStore` `caProvider` / `caBundle` fields **already exist** in shipped operand CRDs (bindata/OLM); the downstream operator does not own that schema.

**Open clarifications carried from specs (not expanded here):** A-007 (shared vs per-store ConfigMap), A-011 (platform CA when proxy absent).

---

## 1. Architecture Overview

### 1.1 Project Type & Tech Stack

| Item | Evidence |
|------|----------|
| Language / Go | `go 1.25.3` (`go.mod`) |
| Framework | controller-runtime `v0.22.5`, kubebuilder-style CRDs |
| Build | GNU Make, `go.work` (4 modules), workspace vendoring |
| Operand packaging | Static YAML in `bindata/` → `go-bindata` → reconciled imperatively |
| OLM | `bundle/`, CSV `features.operators.openshift.io/proxy-aware: "true"` |

Single binary, three reconcilers (`pkg/operator/setup_manager.go`): `external_secrets_manager` → `external_secrets` → optional `crd_annotator`.

### 1.2 Component Map

| Package / area | Responsibility | Hand-written |
|----------------|----------------|--------------|
| `api/v1alpha1/` | Operator CRDs: `ExternalSecretsConfig`, `ExternalSecretsManager` | Yes |
| `pkg/controller/external_secrets/` | Operand install: deployments, RBAC, webhooks, proxy/CA volumes, ConfigMaps | Yes |
| `pkg/controller/external_secrets_manager/` | ESM singleton, global config, status aggregation | Yes |
| `pkg/controller/crd_annotator/` | cert-manager CA injection annotations on operand CRDs | Yes |
| `bindata/external-secrets/` | Upstream operand manifests + upstream CRD YAML | Static (refreshed via `hack/update-external-secrets-manifests.sh`) |
| `config/crd/bases/customresourcedefinition_*external-secrets.io*` | Upstream operand CRDs (SecretStore, etc.) | Generated from upstream helm refresh |
| `pkg/controller/client/` | `CtrlClient` interface + counterfeiter fakes | Yes + generated fakes |

### 1.3 Framework & Pattern Architecture

- **Create-or-update** reconciliation with `common.HasObjectChanged()` drift detection — not SSA-first.
- Operand controller decodes bindata YAML, mutates (namespace, labels, images, env, volumes), then Create or UpdateWithRetry.
- **Proxy + trusted CA are coupled today:** both `ensureTrustedCABundleConfigMap` and `updateTrustedCABundleVolumes` run only when `getProxyConfiguration()` returns non-nil (`configmap.go:24-29`, `deployments.go:574-580`).
- **Fixed platform CA ConfigMap:** name `external-secrets-trusted-ca-bundle`, label `config.openshift.io/inject-trusted-cabundle: "true"`, CNO fills `data` — operator never writes CA PEM content (`configmap.go:65-66`).
- **Dead-code trap:** Do not add reconcile logic to `crd_annotator` or RBAC-only paths for this feature; operand workload changes belong in `external_secrets` deployment/configmap reconcilers.

### 1.4 Runtime Data/Control Flow (RFE-8685 relevant path)

1. User creates/updates cluster singleton `ExternalSecretsConfig` (`metadata.name: cluster`).
2. `external_secrets` Reconciler fetches ESC + ESM, calls `reconcileExternalSecretsDeployment`.
3. Ordered install (`install_external_secrets.go:32-98`): namespace → network policies → service accounts → certificates → secrets → **`ensureTrustedCABundleConfigMap`** → RBAC → services → **`createOrApplyDeployments`** (includes `updateProxyConfiguration` → `updateTrustedCABundleVolumes`) → webhooks.
4. Each deployment mutation runs `updateProxyConfiguration` which sets proxy env vars and mounts/removes trusted CA volume at `/etc/pki/tls/certs` based solely on proxy presence.
5. Upstream **external-secrets controller** (operand pod) performs SecretStore sync; provider TLS uses store spec `caBundle` / `caProvider` (upstream CRD fields), plus container system trust paths.

---

## 2. Target Files (Modification & Creation)

### Operator API (enterprise CA reference — greenfield on release-1.1)

| File | Reason | Confidence |
|------|--------|------------|
| `api/v1alpha1/external_secrets_config_types.go` | Add enterprise CA ConfigMap reference field(s) on `ExternalSecretsConfigSpec` / `ApplicationConfig` or `CommonConfigs` | high |
| `api/v1alpha1/meta.go` | Possible shared type (e.g., ConfigMap reference struct) if reused by ESM | medium |
| `api/v1alpha1/externalsecretsmanager_types.go` | Optional mirror on `GlobalConfig` if cluster-wide default desired | medium |
| `api/v1alpha1/tests/externalsecretsconfig.operator.openshift.io/*.testsuite.yaml` | CEL/API validation cases for new fields | high |

### Controller reconciliation

| File | Reason | Confidence |
|------|--------|------------|
| `pkg/controller/external_secrets/configmap.go` | Today proxy-only; extend for user enterprise CA ConfigMap validation, optional separate from CNO-injected CM | high |
| `pkg/controller/external_secrets/deployments.go` | `updateTrustedCABundleVolumes` / `addTrustedCABundleVolumes` — decouple from proxy gate; mount user-referenced ConfigMap on controller + webhook deployments | high |
| `pkg/controller/external_secrets/constants.go` | New volume/mount names if enterprise CA uses distinct volume from `trusted-ca-bundle` | high |
| `pkg/controller/external_secrets/utils.go` | `getProxyConfiguration` unchanged; add `getEnterpriseCAConfiguration` or similar | high |
| `pkg/controller/external_secrets/install_external_secrets.go` | Reconcile ordering if enterprise CA ConfigMap ensure runs without proxy | high |
| `pkg/controller/external_secrets/controller.go` | May need watch on user-referenced ConfigMap (ConfigMap already in `controllerManagedResources`) | medium |
| `pkg/controller/common/utils.go` | Extend `HasObjectChanged` volume/volumeMount comparison if new volume types added | medium |

### Generated / manifest cascade (after API change)

| File | Reason | Confidence |
|------|--------|------------|
| `config/crd/bases/operator.openshift.io_externalsecretsconfigs.yaml` | Regenerated CRD | high |
| `bundle/manifests/operator.openshift.io_externalsecretsconfigs.yaml` | OLM bundle | high |
| `docs/api_reference.md` | Generated API docs | high |

### Tests

| File | Reason | Confidence |
|------|--------|------------|
| `pkg/controller/external_secrets/deployments_test.go` | Existing trusted CA + proxy table tests; add enterprise CA scenarios | high |
| `pkg/controller/external_secrets/configmap_test.go` | Create if missing; ConfigMap ensure tests | medium |
| `test/e2e/*` | New `Feature:TrustedCABundle` or enterprise CA e2e (label documented in `docs/testing-guidelines.md` but **no e2e file on branch**) | high |

### NOT target files for store-level CA schema (already upstream)

| File | Reason |
|------|--------|
| `config/crd/bases/customresourcedefinition_secretstores.external-secrets.io.yml` | `caProvider` / `caBundle` already present — tier-4 operand bump only |
| `config/crd/bases/customresourcedefinition_clustersecretstores.external-secrets.io.yml` | Same |
| `bindata/external-secrets/resources/deployment_*.yml` | Base deployments have no CA volumes; operator adds at reconcile time |

---

## 3. Reference Context (Read-Only)

### 3.1 Entry Points & Wiring

- `cmd/external-secrets-operator/main.go` — operator entry
- `pkg/operator/setup_manager.go` — controller registration order
- `pkg/controller/external_secrets/controller.go` — watches ESC, ESM, managed Deployments, ConfigMaps, etc.

### 3.2 API / Interface Patterns

- `api/v1alpha1/external_secrets_config_types.go` — ESC spec (`ApplicationConfig`, `ControllerConfig`, `Plugins`)
- `api/v1alpha1/meta.go` — `CommonConfigs` embeds `Proxy *ProxyConfig`; no CA field today
- `api/v1alpha1/conditions.go` — `Ready`, `Degraded` on ESC

### 3.3 Build, CI & Tooling

- `Makefile` — `make build`, `make test`, `make verify`, `make update`
- `.golangci.yml` — lint rules; `kubeapilinter` on `api/v1alpha1/*`

### 3.4 Manifest / Config Generation Pipelines

- `hack/update-external-secrets-manifests.sh $(EXTERNAL_SECRETS_VERSION)` — refreshes operand YAML + upstream CRDs
- `make update-bindata` — embeds bindata
- Operator CRD: `make manifests` from kubebuilder markers

### 3.5 Test Patterns & Fixtures

- Unit: table-driven tests with `FakeCtrlClient` in `deployments_test.go` (proxy + trusted CA volume expectations)
- API: declarative `api/v1alpha1/tests/**/*.testsuite.yaml`
- E2E: Ginkgo labels in `docs/testing-guidelines.md`; Bitwarden e2e uses inline `caBundle` on ClusterSecretStore (`test/e2e/bitwarden_es_test.go`)

---

## 4. Configuration Surface & Runtime Behavior

### 4.1 Current Configuration Surface (release-1.1)

**ExternalSecretsConfig** (`operator.openshift.io`, cluster singleton `cluster`):

| Field path | Type | Default / constraint | RFE-8685 relevance |
|------------|------|----------------------|-------------------|
| `spec.appConfig.proxy` | `ProxyConfig` | optional | Triggers platform CA injection today |
| `spec.appConfig.logLevel`, `resources`, `affinity`, `tolerations`, `nodeSelector` | various | via `CommonConfigs` | — |
| `spec.appConfig.operatingNamespace` | string | optional | Restricts cluster-scoped stores |
| `spec.appConfig.webhookConfig` | `WebhookConfig` | optional | Webhook receives same deployment mutations |
| `spec.controllerConfig.certProvider.certManager` | `CertManagerConfig` | webhook TLS path | Unrelated to outbound provider TLS |
| `spec.controllerConfig.componentConfigs[]` | `ComponentConfig` | `overrideEnv`, `revisionHistoryLimit` only | **No volume mount override** — supports FR-010 (no ad-hoc mounts) |
| `spec.plugins.bitwardenSecretManagerProvider` | plugin gate | conditional TLS | Separate workload |

**ExternalSecretsManager** (`spec.globalConfig.proxy`, labels, resources): proxy fallback when ESC has no proxy (`utils.go:127-155`).

**ProxyConfig** (`api/v1alpha1/meta.go:92-114`): `httpProxy`, `httpsProxy`, `noProxy` — env vars only.

**Enterprise CA ConfigMap reference:** **ABSENT** on operator API.

**Upstream SecretStore / ClusterSecretStore** (`external-secrets.io`, shipped in `config/crd/bases/`):

| Field | Present on branch | Owner |
|-------|-------------------|-------|
| Provider-specific `caBundle` (bytes) | Yes | Upstream operand |
| `caProvider.type` (`Secret` \| `ConfigMap`), `name`, `key`, `namespace` | Yes | Upstream operand |

Example `caProvider` schema: `customresourcedefinition_clustersecretstores.external-secrets.io.yml` lines 320-352.

### 4.2 Reconciliation / Processing Flow (Detailed)

| Step | Function | Behavior | Error handling |
|------|----------|----------|----------------|
| 1 | `getProxyConfiguration` | ESC → ESM → OLM env precedence | Returns nil if no proxy |
| 2 | `ensureTrustedCABundleConfigMap` | If proxy: create `external-secrets-trusted-ca-bundle` with inject label; skip if no proxy | Returns error to reconcile; TODO on removal when proxy removed |
| 3 | `createOrApplyDeployments` | Decode bindata deployment, apply image/logLevel/overrides | Irrecoverable on validation failure |
| 4 | `updateProxyConfiguration` | Set proxy env on all containers + initContainers | — |
| 5 | `updateTrustedCABundleVolumes` | If proxy: mount CNO ConfigMap at `/etc/pki/tls/certs`; else remove mounts | — |
| 6 | Operand controller runtime | Reads SecretStore, connects to provider using `caProvider` / system roots | Errors in store status (upstream) |

**Gap vs specs:** Steps 2 and 5 do not run without proxy — enterprise PKI without cluster proxy fails FR-001/FR-008 today.

### 4.3 Image / Dependency Resolution

- Operand image: `RELATED_IMAGE_EXTERNAL_SECRETS` env → `deployments.go` image override
- Version label: `OPERAND_EXTERNAL_SECRETS_IMAGE_VERSION`
- Not affected by CA feature directly

### 4.4 Status / Health Reporting

- ESC conditions: `Ready`, `Degraded` set in `controller.go:392-474` on reconcile success/failure
- No dedicated condition for missing trusted CA ConfigMap today
- FR-009 would require new validation/status paths for missing/invalid enterprise CA ConfigMap

### 4.5 Feature Gate / Feature Flag Mechanism

- ESM `spec.features[]` via `common.IsFeatureEnabled()` — not used for CA/proxy today
- OpenShift FeatureGate API not used

---

## 5. Reusable Assets (Anti-Duplication)

| Asset | Use for RFE-8685 | Evidence |
|-------|------------------|----------|
| `addTrustedCABundleVolumes` / `removeTrustedCABundleVolumes` | Extend pattern for mounting ConfigMap volumes at `/etc/pki/tls/certs` (or secondary path) | `deployments.go:584-679` |
| `getTrustedCABundleLabels` | Pattern for labels on operator-created ConfigMaps | `configmap.go:85-91` |
| `getProxyConfiguration` | Keep proxy env logic separate from enterprise CA | `utils.go:127-155` |
| `common.HasObjectChanged()` | Deployment drift including volumes | `pkg/controller/common/utils.go` |
| `common.FromClientError`, `NewIrrecoverableError` | Error wrapping for missing ConfigMap | `pkg/controller/common/errors.go` |
| Upstream `caProvider` on ClusterSecretStore | Store-level trust (FR-003) without operator CRD changes | CRD YAML on branch |
| `FakeCtrlClient` + `deployments_test.go` patterns | Unit tests for volume/env mutations | existing proxy/CA tests |

---

## 6. Architectural Guardrails

### Structural

- Operand changes only in `pkg/controller/external_secrets/`; do not embed upstream controller code.
- Create-or-update with `HasObjectChanged`; add volume field comparisons if new volume names introduced.

### API / Schema

- Operator API group `operator.openshift.io` only for new enterprise CA fields.
- Do not edit upstream `external-secrets.io` CRD YAML by hand for this feature — store CA already exists.
- CEL/kubebuilder validation on new ConfigMap reference; API tests in `.testsuite.yaml`.

### Build / Tooling

- After API markers: `make update && make verify`.
- FIPS production build via `hack/go-fips.sh` — no special CA change but do not break strict FIPS build.

### Deployment / Packaging

- Mount enterprise CA on **controller** (`deployment_external-secrets.yml`) and **webhook** (`deployment_external-secrets-webhook.yml`) per jira/spec; cert-controller only if spec/plan expands scope.
- Do not expose arbitrary volume mounts via `componentConfigs` — conflicts with FR-010.

### Code Generation

- Never hand-edit `zz_generated.deepcopy.go`, bundle CRDs, `bindata.go`.

### Security

- Validate referenced ConfigMap exists/read RBAC; read-only volume mounts (`ReadOnly: true` pattern in `deployments.go:612-614`).
- Distinguish CNO-managed platform bundle vs user-managed enterprise ConfigMap to avoid overwriting CNO `data`.

---

## 7. Change Cascade Checklist

| When you change... | You must also... | Verify with... |
|---|---|---|
| API fields in `api/v1alpha1/` | `make generate && make manifests && make docs` | `make verify-generated` |
| CRD validation rules | Add `api/v1alpha1/tests/.../*.testsuite.yaml` cases | `make test-apis` |
| Deployment volume logic | Update `deployments_test.go`; consider e2e | `make test-unit` |
| ConfigMap reconcile | Unit tests for ensure/validate | `go test ./pkg/controller/external_secrets/...` |
| RBAC if operator reads ConfigMaps cluster-wide | `config/rbac/role.yaml` markers in `controller.go` | `make manifests` |
| OLM CSV descriptors | `make bundle` | `make verify` |

---

## 8. Test & CI Reference

### 8.1 Test Structure

| Tier | Location | Framework |
|------|----------|-----------|
| Unit | `pkg/controller/**/_test.go` | std testing, gomega |
| API | `api/v1alpha1/tests/`, `test/apis/` | envtest + testsuite YAML |
| E2E | `test/e2e/` | Ginkgo v2 |

### 8.2 How to Run Tests Locally

```bash
make test-unit
make test-apis
make test-e2e E2E_GINKGO_LABEL_FILTER="Feature:TrustedCABundle"   # label documented; no spec file on release-1.1 yet
make lint
make verify
```

### 8.3 CI Pipeline

- In-repo: `make verify` (fmt, vet, bindata, generated diff, govulncheck)
- E2E gated by Ginkgo labels; `Feature:Proxy` requires cluster-wide OpenShift proxy

### 8.4 Test Coverage Gaps

| Area | release-1.1 state |
|------|-------------------|
| Proxy + trusted CA volumes | **Unit tests exist** (`deployments_test.go`) |
| Enterprise CA without proxy | **No tests** — greenfield |
| E2E `Feature:TrustedCABundle` | **Documented but no `test/e2e` implementation found** on branch |
| Store `caProvider` e2e for generic enterprise PKI | Only Bitwarden path uses inline `caBundle` |

---

## 9. Developer Workflow

### 9.1 Key Commands Reference

| Command | Purpose |
|---------|---------|
| `make build-operator` | Fast binary compile |
| `make update && make verify` | Full regen + CI parity |
| `make test-unit` | Controller unit tests |
| `make test-apis` | CRD validation |
| `make lint-fix` | Auto-fix lint |

### 9.2 Version Variables

| Variable | Location | Notes |
|----------|----------|-------|
| `EXTERNAL_SECRETS_VERSION` | `Makefile:29` | `v0.20.4` on branch — defines upstream CRD/manifest pin |

### 9.3 Local Development Setup

- Set `RELATED_IMAGE_EXTERNAL_SECRETS` for `make run`
- Working folder at `release-1.1` HEAD per scope constraints

### 9.4 Common Development Scenarios

**How to add operator-level ConfigMap reference (this feature):**

1. Add typed field to `ExternalSecretsConfigSpec` / `CommonConfigs` with kubebuilder validation (name + namespace required).
2. Regenerate: `make update`.
3. Add validation helper in `pkg/controller/external_secrets/` to fetch ConfigMap via uncached/cached client as appropriate.
4. Extend `ensureTrustedCABundleConfigMap` or add `ensureEnterpriseCAConfigMap` — **do not** apply CNO inject label to user-owned enterprise ConfigMaps unless product decision merges platform + enterprise paths (A-011).
5. Extend `updateTrustedCABundleVolumes` or add parallel `updateEnterpriseCABundleVolumes` for controller + webhook deployments.
6. Add unit tests mirroring `deployments_test.go` proxy/CA cases.
7. Add API testsuite YAML + optional e2e with `Feature:TrustedCABundle`.

**How to use store-level CA (no operator code if sufficient):**

- Administrators create ConfigMap with PEM CAs, reference in `ClusterSecretStore` via existing `caProvider` (`type: ConfigMap`, `name`, `key`, `namespace`) — schema already on branch.

---

## 10. Platform & Environment Integration

### 10.1 Security Context & Permissions

- Operand containers: `readOnlyRootFilesystem: true`, non-root — CA must be mounted read-only (`deployments.go` pattern).
- Operator needs RBAC to read referenced ConfigMaps in target namespace(s).

### 10.2 Proxy & Network Configuration

**Existing proxy-driven trusted CA (release-1.1):**

| Constant | Value | File |
|----------|-------|------|
| ConfigMap name | `external-secrets-trusted-ca-bundle` | `constants.go:56` |
| Inject label | `config.openshift.io/inject-trusted-cabundle: "true"` | `constants.go:59` |
| Volume name | `trusted-ca-bundle` | `constants.go:62` |
| Mount path | `/etc/pki/tls/certs` | `constants.go:67` |
| Gate | `getProxyConfiguration() != nil` | `configmap.go:24`, `deployments.go:574` |

CNO populates ConfigMap data; operator only ensures metadata/labels.

**Operator pod (separate):** `config/manager/kustomization.yaml` + `trusted-ca-patch.yaml` mount platform CA for the **operator manager**, not the operand — do not conflate with operand enterprise CA work.

**Enterprise CA requirement (greenfield):** User-supplied ConfigMap reference, independent of proxy, mounted to controller/webhook — **not present** on branch.

**A-011 note:** Code confirms platform operand CA injection is **skipped entirely** when proxy config is nil — no `/etc/pki/tls/certs` mount in that case.

### 10.3 Cloud Provider Integration

- Not central to CA feature; AWS/GCP e2e use public CA endpoints unless store `caProvider` configured.

### 10.4 Build & Compliance Constraints

- FIPS build tags for production (`hack/go-fips.sh`)

### 10.5 Console / UI Integration

- `config/console/` samples for SecretStore — could document `caProvider` usage post-implementation.

### 10.6 Packaging & Lifecycle

- CSV marks proxy-aware; new API fields need CSV schema descriptors via `make bundle`.
- Upstream CRD updates only via operand version tier-3 bump — out of scope for operator-only enterprise CA API.

---

## 11. Risks & Downstream Impacts

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Mount path collision** | Platform CNO bundle and enterprise CA both target `/etc/pki/tls/certs` | Use separate volume/mount or merge strategy; resolve A-011 in plan |
| **Proxy gate removal side effects** | Decoupling CA from proxy may change behavior for proxy-less clusters | Explicit spec tests for proxy absent + enterprise CA enabled |
| **Scope creep into upstream CRDs** | Jira mentions SecretStore CRDs; schema already exists | Limit operator work to API + operand mounts; document store `caProvider` for FR-003 |
| **ConfigMap lifecycle** | Deleted/invalid CM after enable | Validate on reconcile; set ESC Degraded (FR-009) |
| **Watch scope** | User ConfigMap outside operand namespace | May need uncached get + watch mapping |
| **Single mount path for Go TLS** | Go reads system cert dir — multiple CAs must be visible | Supplement vs replace semantics (A-002) |

### 11.1 Assessment Limitations / UNVERIFIED Items

- **Thycotic / IBM Secret Server provider block** in upstream CRD not individually verified — generic `caProvider` applies to HTTPS providers with CA fields.
- **E2E `Feature:TrustedCABundle`**: documented in `docs/testing-guidelines.md:118-141` but **no matching `test/e2e/*.go` on `release-1.1`** — verify before planning e2e tasks.
- **E2E `Feature:Proxy`**: same — no e2e Go match on branch.
- **Exact provider for customer vault** not in repo — store-level `caProvider` may suffice without operator API if only provider TLS verification fails.
- **A-007 / A-011** unresolved — plan must not assume shared ConfigMap or proxy-less platform injection without product decision.

### 11.2 Operator vs SecretStore Ownership (release-1.1)

| Concern | Owner | Branch state | Planned work type |
|---------|-------|--------------|-------------------|
| Mount CA into operand controller/webhook pods | **Downstream operator** (`pkg/controller/external_secrets`) | Proxy-gated CNO injection only | **Delta / greenfield** — new API + reconcile |
| Reference ConfigMap in operator config | **Downstream operator** (`ExternalSecretsConfig`) | Field absent | **Greenfield API** |
| `caProvider` / `caBundle` on SecretStore | **Upstream external-secrets.io** | Present in shipped CRD v0.20.4 | **Document / verify** — no operator schema change |
| Provider TLS verification at sync time | **Upstream operand controller** | Uses store spec | Configure via ClusterSecretStore; not operator reconciler |
| Webhook serving TLS (inbound) | Operator + cert-manager or cert-controller | Existing `certProvider` | Out of scope (A-006) |

**Conclusion:** Primary implementation effort is **operator-owned** (ExternalSecretsConfig enterprise CA + deployment volume wiring). **Store-level CA references are already supported by upstream CRDs** on this branch; customer may need operator pod trust **and/or** store `caProvider` depending on where TLS verification fails. Planning should treat store CA as configuration/documentation path unless gap analysis shows operand version lacks required provider fields.

---

## 12. Quick Reference Card

### Preflight Checklist (run before every PR)

```
1. make update && make verify
2. make test-unit
3. make test-apis    # if API changed
4. make lint
5. make test-e2e E2E_GINKGO_LABEL_FILTER="..."   # if e2e added
```

### Key File Quick-Nav

| I want to... | Look at... |
|---|---|
| Add enterprise CA API field | `api/v1alpha1/external_secrets_config_types.go`, `meta.go` |
| Change operand CA mounts | `pkg/controller/external_secrets/deployments.go` |
| Change platform CA ConfigMap ensure | `pkg/controller/external_secrets/configmap.go` |
| Proxy precedence | `pkg/controller/external_secrets/utils.go` (`getProxyConfiguration`) |
| Install ordering | `pkg/controller/external_secrets/install_external_secrets.go` |
| Store CA schema (read-only) | `config/crd/bases/customresourcedefinition_clustersecretstores.external-secrets.io.yml` |
| Unit test patterns | `pkg/controller/external_secrets/deployments_test.go` |
| API validation tests | `api/v1alpha1/tests/externalsecretsconfig.operator.openshift.io/` |
| Upstream CRD refresh | `Makefile` `EXTERNAL_SECRETS_VERSION`, `hack/update-external-secrets-manifests.sh` |
