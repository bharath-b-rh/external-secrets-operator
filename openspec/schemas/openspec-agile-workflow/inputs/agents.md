This file provides guidance to AI agents working with the **external-secrets-operator** for OpenShift — a Go operator that installs and manages the upstream [external-secrets](https://external-secrets.io/) operand. It uses **controller-runtime only** (no library-go).

## Repository Layout

```
api/v1alpha1/                          # Operator CRD types (v1alpha1 only)
bindata/external-secrets/resources/    # Operand YAML (helm-sourced, go-bindata embedded)
bundle/                                # OLM bundle (CSV, CRDs)
cmd/external-secrets-operator/         # Binary entrypoint (separate go.mod)
config/                                # Kustomize: CRDs, RBAC, manager, samples, console
docs/                                  # API reference, manual test plans
hack/                                  # Manifest update, verify, test-apis, FIPS scripts
pkg/controller/
  client/                              # CtrlClient interface + counterfeiter fakes
  common/                              # Errors, utils, validation, constants
  commontest/                          # Shared unit-test fixtures
  crd_annotator/                       # Optional cert-manager CRD annotation controller
  external_secrets/                    # Primary operand install controller
  external_secrets_manager/            # ESM status aggregation controller
pkg/operator/
  setup_manager.go                     # StartControllers() — wires all reconcilers
  assets/bindata.go                    # Generated go-bindata (DO NOT EDIT)
test/
  apis/                                # envtest + Ginkgo API validation tests
  e2e/                                 # Live-cluster Ginkgo tests (build tag: e2e)
  utils/                               # E2E helpers
tools/                                 # counterfeiter, linters, codegen tools
```

**Workspace:** `go.work` spans root `.`, `./cmd/external-secrets-operator`, `./test`, `./tools`. Vendor via `make update-vendor`.

## Architecture — Single controller-runtime Manager

All controllers register on **one** `ctrl.Manager` in `cmd/external-secrets-operator/main.go`. Never create separate managers.

| Controller | Package | `ControllerName` | Primary CR |
|------------|---------|------------------|------------|
| External Secrets Manager | `pkg/controller/external_secrets_manager/` | `external-secrets-manager` | `ExternalSecretsManager` (singleton `cluster`) |
| External Secrets (operand) | `pkg/controller/external_secrets/` | `external-secrets-controller` | `ExternalSecretsConfig` (singleton `cluster`) |
| CRD Annotator (optional) | `pkg/controller/crd_annotator/` | `crd-annotator` | Operand CRDs (when cert-manager installed) |

**Startup order** (`pkg/operator/setup_manager.go`):
1. `esmcontroller.New().SetupWithManager()`
2. `escontroller.New().SetupWithManager()`
3. If cert-manager CRD present: `crdannotator.New().SetupWithManager()`
4. `esmcontroller.CreateDefaultESMResource()` (uncached client)

**Operand namespace:** `external-secrets` (`OperandDefaultNamespace`).

## Reconciliation Pattern — Create-or-Update (Not SSA-First)

Unlike addon operators that use Server-Side Apply for all operands, this repo primarily uses **create-or-update with deep equality**:

- `createWithFallback()` → `Create()`; on `AlreadyExists`, `UncachedClient.UpdateWithRetry()`
- Existing resources: `common.HasObjectChanged()` → `UpdateWithRetry()`
- Limited SSA: `client.Apply` + `client.FieldOwner(common.ExternalSecretsOperatorCommonName)` only for CR annotation patches

**Install sequence** (`pkg/controller/external_secrets/install_external_secrets.go`):
validateConfig → namespace → networkPolicies → serviceAccounts → certificates → secret → trusted CA ConfigMap → RBAC → services → configMaps → deployments → validatingWebhooks → status/annotation patches

**Managed label:** `app=external-secrets`. Watch label for user refs: `externalsecretsconfig.operator.openshift.io/watching=true`.

**Finalizer:** `externalsecretsconfigs.operator.openshift.io/external-secrets-controller`

## Operator CRDs

| CRD | Scope | Singleton name | Purpose |
|-----|-------|----------------|---------|
| `ExternalSecretsConfig` | Cluster | `cluster` | Operand configuration (webhooks, providers, deployment) |
| `ExternalSecretsManager` | Cluster | `cluster` | Global config, feature flags, status aggregation |

**API package:** `api/v1alpha1/` — group `operator.openshift.io`, version `v1alpha1` only.

**Shared types:** `api/v1alpha1/meta.go` (`ConditionalStatus`, `CommonConfigs`, `ManagementState`), `api/v1alpha1/conditions.go` (`Ready`, `Degraded`).

**Validation:** CEL `+kubebuilder:validation:XValidation`, enums, defaults; YAML testsuites at `api/v1alpha1/tests/<crdName>/*.testsuite.yaml`.

**Never hand-edit:** `zz_generated.deepcopy.go`, `config/crd/bases/operator.openshift.io_*.yaml`.

## Shared Packages — Never Duplicate

### `pkg/controller/client/`

| Symbol | Use for |
|--------|---------|
| `CtrlClient` interface | All client operations (Get/List/Create/Update/UpdateWithRetry/Patch/Delete/StatusUpdate/Exists) |
| `fakes/FakeCtrlClient` | Unit test mocking (counterfeiter) |

Regenerate fake: `go generate ./pkg/controller/client/...`

### `pkg/controller/common/`

| Symbol | Use for |
|--------|---------|
| `NewIrrecoverableError()`, `NewRetryRequiredError()`, `NewUserConfigurationError()`, `FromClientError()` | Error classification |
| `DefaultRequeueTime` (30s) | Requeue after recoverable errors |
| `Decode*ObjBytes` | Deserialize bindata YAML |
| `HasObjectChanged()` | Deep equality before update |
| `ApplyResourceMetadata()`, `AddFinalizer()`, `RemoveFinalizer()` | Metadata and finalizers |
| `IsFeatureEnabled(esm, name)` | ESM feature flag lookup |
| `ExternalSecretsConfigObjectName`, `ExternalSecretsManagerObjectName` | `"cluster"` |
| `CertManagerInjectCAFromAnnotation` | cert-manager CA injection |

## Feature Flags (CR-Based, Not OpenShift FeatureGate)

Features are defined on `ExternalSecretsManager.Spec.Features[]` with `FeatureName` enum (e.g. `UnsafeAllowGenericTargets`).

- Check via `common.IsFeatureEnabled(esm, name)`
- Deployment args via `featureContainerArgs` map in `pkg/controller/external_secrets/constants.go`
- **No** `pkg/features/`, **no** cluster `featuregates.config.openshift.io` discovery

## Operand Manifests & Bindata

Operand YAML comes from upstream helm via `hack/update-external-secrets-manifests.sh $(EXTERNAL_SECRETS_VERSION)` (default `v2.5.0`).

Pipeline: helm template → yq cleanup → split to `bindata/external-secrets/resources/` + `config/crd/bases/customresourcedefinition_*` → `make update-bindata`.

**Never hand-edit:** `pkg/operator/assets/bindata.go`.

## cert-manager Integration (Optional)

- CRD discovery at startup (`IsCertManagerInstalled()`)
- Webhook TLS via cert-manager `Certificate` CR or in-tree cert-controller deployment
- `crd_annotator` adds `cert-manager.io/inject-ca-from` to operand CRDs when cert-manager is present

## Environment Variables (OLM)

| Variable | Purpose |
|----------|---------|
| `RELATED_IMAGE_EXTERNAL_SECRETS` | Operand image |
| `OPERAND_EXTERNAL_SECRETS_IMAGE_VERSION` | Operand version labels |
| `RELATED_IMAGE_BITWARDEN_SDK_SERVER` | Bitwarden SDK sidecar image |
| `BITWARDEN_SDK_SERVER_IMAGE_VERSION` | Bitwarden SDK version |

Missing image env → `NewIrrecoverableError`.

## OpenShift Conventions

- Trusted CA: `config.openshift.io/inject-trusted-cabundle` on proxy CA ConfigMap
- Metrics: HTTPS with OpenShift service CA (`main.go` `loadOpenShiftCACertPool`)
- FIPS: `hack/go-fips.sh` (`GOEXPERIMENT=strictfipsruntime`, `CGO_ENABLED=1`)
- Operator deploy/install: `kubectl apply --server-side` via Makefile `install`/`deploy`

## Code Generation

| Command | Output |
|---------|--------|
| `make generate` | `zz_generated.deepcopy.go` (controller-gen) |
| `make manifests` | CRDs in `config/crd/bases/`, RBAC in `config/rbac/role.yaml` |
| `make update-operand-manifests` | Refresh bindata + operand CRDs from helm |
| `make update-bindata` | `pkg/operator/assets/bindata.go` |
| `make bundle` | OLM bundle |
| `make docs` | `docs/api_reference.md` |

After API type changes: `make generate && make manifests`. After operand version bump: `make update-operand-manifests && make update-bindata`.

## Testing Patterns

### Unit tests (`pkg/controller/**`)

- Colocated `*_test.go`; standard `testing` package (no testify in `pkg/`)
- Table-driven: `tests := []struct { name string; preReq func(...); ... }`
- Mocking: counterfeiter `fakes.FakeCtrlClient` — not `fake.NewClientBuilder`
- Fixtures: `pkg/controller/external_secrets/test_utils.go`, `pkg/controller/commontest/utils.go`
- Run: `make test-unit` (excludes `test/e2e`, `test/apis`, `test/utils`)

### API integration tests (`test/apis/`)

- Ginkgo v2 + Gomega + envtest
- Loads `api/v1alpha1/tests/**/*.testsuite.yaml`
- Run: `make test-apis` (K8s 1.32.0 via envtest)

### E2E tests (`test/e2e/`)

- Build tag: `//go:build e2e`
- Separate module: `test/go.mod`
- Ginkgo + live cluster; default label filter: `Platform: isSubsetOf {AWS}`
- Run: `make test-e2e`
- Labels: `Platform:AWS`, `Provider:Bitwarden`, `API:Bitwarden`, `CrossPlatform:GCP-AWS`

## Makefile — Verification Matrix

| Target | When to run |
|--------|-------------|
| `make test-unit` | Controller / common package changes |
| `make test-apis` | API type, CEL rule, or testsuite YAML changes |
| `make test` | Full pre-PR check (manifests, generate, fmt, vet, test-apis, test-unit) |
| `make verify` | Codegen consistency (bindata, generated, govulncheck, git diff) |
| `make lint` | golangci-lint + kube-api-linter |
| `make build` | Compile check after substantive changes |
| `make bundle` | OLM CSV/CRD changes |

## Common Mistakes

1. Do NOT use SSA-first operand patterns from other operators — this repo uses create-or-update
2. Do NOT create separate `ctrl.Manager` instances — register in `setup_manager.go`
3. Do NOT duplicate client/error helpers — use `pkg/controller/client/` and `pkg/controller/common/`
4. Do NOT hand-edit generated files (`bindata.go`, deepcopy, operator CRD YAML)
5. Do NOT use OpenShift FeatureGate discovery — features are on `ExternalSecretsManager` CR
6. Do NOT assume cert-manager is always present — guard with `IsCertManagerInstalled()`
7. Do NOT use `make test-unit` naming from other operators — this repo uses `make test-unit` (not `make test` alone for unit only; `make test` is full suite)
8. Do NOT import controller internals in e2e when constants should be mirrored — check existing `test/e2e/` patterns first

---

## Per-task testing during `/opsx-apply` (code generation eval gate)

During implementation, each code generation task is verified with **real command execution**. See `stage-gate/CODE_GENERATION_EVAL_PROMPT.md` in the schema package.

| Task type | Verification | Test strategy |
|-----------|-------------|---------------|
| API types | `go build`, `go vet` | `make test-apis` after testsuite YAML |
| Codegen | `make generate && make manifests && make verify` | Consistency check |
| Controller logic | `go test ./pkg/controller/...` | Co-generated `_test.go` with `FakeCtrlClient` |
| Operand manifests | `make update-operand-manifests && make update-bindata && make verify` | `make verify-bindata` |
| OLM bundle | `make bundle` | Bundle validation |
| ESM / feature flags | `go test ./pkg/controller/external_secrets_manager/...` | Unit tests for feature wiring |

---

## Execution agent routing

Use these **Assigned Agent** IDs in `tasks.md` §3 when **`AgentRoutingMode: PROVIDED`**. Each task gets exactly one primary agent.

| Agent ID | Scope | Route when task touches | OAPE / execution |
|----------|-------|-------------------------|------------------|
| **API_Agent** | CRD/API types, markers, `.testsuite.yaml` | `api/v1alpha1/`, `api/v1alpha1/tests/` | `api-generate` or `api-generate-tests` (verification-only) |
| **OperatorController_Agent** | Reconciliation, operand install, wiring | `pkg/controller/external_secrets/`, `pkg/controller/external_secrets_manager/`, `pkg/controller/crd_annotator/`, `pkg/operator/setup_manager.go`, `cmd/external-secrets-operator/` | `api-implement` |
| **ManifestsBindata_Agent** | Operand YAML, version pins | `bindata/`, `hack/update-external-secrets-manifests.sh`, `Makefile` `EXTERNAL_SECRETS_VERSION`, `config/crd/bases/customresourcedefinition_*` | Manual — `make update-operand-manifests`, `make update-bindata` |
| **WebhookTLS_Agent** | Webhook TLS, cert-manager certificates | Webhook deployments, `certificates.go`, trusted CA ConfigMap | Manual |
| **RBACSecurity_Agent** | RBAC, network policies | `config/rbac/`, `rbacs.go`, `networkpolicy.go` | Manual |
| **OLMRelease_Agent** | OLM bundle, CSV, relatedImages | `config/manifests/`, `bundle/`, Makefile bundle targets | Manual — `make bundle` |
| **Testing_Agent** | E2E and API test authoring | `test/e2e/`, `test/apis/`, `test/utils/` | `e2e-generate` when task is e2e |
| **Docs_Agent** | User-facing docs | `README.md`, `docs/` | Manual |

### Controller routing rules

- **Operand controller** (`pkg/controller/external_secrets/`): create-or-update reconcilers; follow `install_external_secrets.go` ordering
- **ESM controller** (`pkg/controller/external_secrets_manager/`): status aggregation, default ESM creation, feature flags
- **CRD annotator** (`pkg/controller/crd_annotator/`): only when cert-manager is installed
- **API before controller**: CRD field tasks must complete (and pass `make test-apis`) before controller tasks that reconcile those fields

### Verification pairing

- API changes → pair with `api/v1alpha1/tests/*.testsuite.yaml` tasks (`API_Agent`, verification-only)
- Controller changes → pair with unit tests (`make test-unit`) and e2e when user-visible (`Testing_Agent`)
- Operand version bumps → pair with `make verify` and relevant e2e smoke paths

---

## Stage-Specific Agent Guidance

### Repo-Assessment Stage Hints

When assessing `external-secrets-operator`, document:

- **Single manager, three reconcilers** — not dual library-go + controller-runtime
- **Create-or-update** operand pattern — not SSA-first (limited SSA for annotations only)
- **Two operator CRDs** — `ExternalSecretsConfig` + `ExternalSecretsManager` (both singleton `cluster`)
- **Operand CRDs** from upstream helm in `config/crd/bases/customresourcedefinition_*`
- **Feature flags on ESM CR** — not OpenShift FeatureSet gates
- **cert-manager optional** — CRD annotator only when cert-manager CRD exists
- **Test commands:** `make test-unit`, `make test-apis`, `make test-e2e`
- **FIPS build** via `hack/go-fips.sh`

Anti-patterns (forbidden without branch evidence):
- Claiming library-go or SSA-first addon patterns apply to this repo
- Using OpenShift FeatureGate / TechPreview cluster discovery patterns from other operators
- Documenting operand controllers or CRDs that do not exist on the target branch

### Planning Stage Hints

Prefer operator-native thinking for external-secrets:
- `ExternalSecretsConfig` / `ExternalSecretsManager` API evolution
- Operand install ordering and webhook TLS paths
- Bindata/helm manifest pipeline (`EXTERNAL_SECRETS_VERSION`)
- Provider integrations (Bitwarden SDK sidecar, AWS/GCP test paths)
- RBAC blast radius (secrets, cluster-scoped operand CRDs)
- OLM bundle and `RELATED_IMAGE_*` env vars

Default repo pin when none provided:
```
primary_repo: "https://github.com/openshift/external-secrets-operator"
branch: "main"
```

### Validation Stage Hints

When the spec touches operators, secrets, webhooks, providers, or OpenShift platform integration, assess:

- API & CRD lifecycle (singleton rules, immutability, CEL validation)
- Install / uninstall / reconcile semantics (finalizers, operand namespace)
- RBAC & blast radius (secret access, cluster-scoped writes)
- Webhook TLS (cert-manager vs in-tree cert-controller)
- Provider matrix (Bitwarden, AWS, GCP — see e2e labels)
- Observability (Ready/Degraded conditions on ESC)
- Upgrade / operand version skew (`EXTERNAL_SECRETS_VERSION`)
