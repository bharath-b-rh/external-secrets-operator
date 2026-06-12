package external_secrets

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/external-secrets-operator/pkg/controller/client/fakes"
	"github.com/openshift/external-secrets-operator/pkg/controller/common"
	"github.com/openshift/external-secrets-operator/pkg/controller/commontest"
)

func TestCreateWithFallback(t *testing.T) {
	resourceMetadata := common.ResourceMetadata{
		Labels: controllerDefaultResourceLabels,
	}

	tests := []struct {
		name          string
		preReq        func(cached *fakes.FakeCtrlClient, uncached *fakes.FakeCtrlClient)
		customDesired func() client.Object
		wantErr       string
		wantUpdate    bool
	}{
		{
			name: "resource created successfully",
			preReq: func(cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return nil
				})
			},
		},
		{
			name: "create fails with non-AlreadyExists error",
			preReq: func(cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return commontest.ErrTestClient
				})
			},
			wantErr: "failed to create ServiceAccount test-ns/test-sa: test client error",
		},
		{
			name: "GVK resolved from object TypeMeta when pre-set, skipping scheme lookup",
			preReq: func(cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return commontest.ErrTestClient
				})
			},
			customDesired: func() client.Object {
				return &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{Kind: "ServiceAccount", APIVersion: "v1"},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-sa",
						Namespace: "test-ns",
						Labels:    controllerDefaultResourceLabels,
					},
				}
			},
			wantErr: "failed to create ServiceAccount test-ns/test-sa: test client error",
		},
		{
			name: "AlreadyExists triggers uncached update to restore labels",
			preReq: func(cached *fakes.FakeCtrlClient, uncached *fakes.FakeCtrlClient) {
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return apierrors.NewAlreadyExists(schema.GroupResource{Resource: "serviceaccounts"}, "test-sa")
				})
				uncached.UpdateWithRetryCalls(func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
					if obj.GetLabels()["app"] != "external-secrets" {
						t.Errorf("expected managed label app=external-secrets to be present, got %v", obj.GetLabels())
					}
					return nil
				})
			},
			wantUpdate: true,
		},
		{
			name: "AlreadyExists but uncached update fails",
			preReq: func(cached *fakes.FakeCtrlClient, uncached *fakes.FakeCtrlClient) {
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return apierrors.NewAlreadyExists(schema.GroupResource{Resource: "serviceaccounts"}, "test-sa")
				})
				uncached.UpdateWithRetryCalls(func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
					return commontest.ErrTestClient
				})
			},
			wantErr:    "failed to restore ServiceAccount test-ns/test-sa to desired state: test client error",
			wantUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := testReconciler(t)
			cached := &fakes.FakeCtrlClient{}
			uncached := &fakes.FakeCtrlClient{}
			if tt.preReq != nil {
				tt.preReq(cached, uncached)
			}
			r.CtrlClient = cached
			r.UncachedClient = uncached

			var desired client.Object
			if tt.customDesired != nil {
				desired = tt.customDesired()
			} else {
				desired = &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-sa",
						Namespace: "test-ns",
						Labels:    controllerDefaultResourceLabels,
					},
				}
			}

			err := r.createWithFallback(desired, resourceMetadata, "test-ns/test-sa", commontest.TestExternalSecretsConfig())

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("createWithFallback() err = %v, wantErr = %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("createWithFallback() unexpected error: %v", err)
			}

			if tt.wantUpdate && uncached.UpdateWithRetryCallCount() == 0 {
				t.Error("expected UncachedClient.UpdateWithRetry to be called, but it wasn't")
			}
			if !tt.wantUpdate && uncached.UpdateWithRetryCallCount() > 0 {
				t.Error("expected no UncachedClient.UpdateWithRetry call, but one was made")
			}
		})
	}
}

// TestCreateWithFallback_GVKResolutionFallback verifies that when the scheme has no
// registered types, createWithFallback falls back to the Go type name in log and error messages.
func TestCreateWithFallback_GVKResolutionFallback(t *testing.T) {
	resourceMetadata := common.ResourceMetadata{
		Labels: controllerDefaultResourceLabels,
	}

	r := testReconciler(t)
	// Use an empty scheme so apiutil.GVKForObject cannot resolve the type.
	r.Scheme = runtime.NewScheme()

	cached := &fakes.FakeCtrlClient{}
	cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
		return commontest.ErrTestClient
	})
	r.CtrlClient = cached
	r.UncachedClient = &fakes.FakeCtrlClient{}

	desired := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test-ns",
			Labels:    controllerDefaultResourceLabels,
		},
	}

	err := r.createWithFallback(desired, resourceMetadata, "test-ns/test-sa", commontest.TestExternalSecretsConfig())
	wantErr := "failed to create *v1.ServiceAccount test-ns/test-sa: test client error"
	if err == nil || err.Error() != wantErr {
		t.Errorf("createWithFallback() err = %v, wantErr = %v", err, wantErr)
	}
}

func TestGetWithCacheFallback(t *testing.T) {
	t.Parallel()

	key := types.NamespacedName{Name: "user-ca", Namespace: OperandDefaultNamespace}
	cm := testUserCAConfigMap("user-ca", mustPEMCert(t, true), nil)
	cacheUnavailableErr := errors.New("cache unavailable")

	tests := []struct {
		name         string
		setupClients func(t *testing.T) (*fakes.FakeCtrlClient, *fakes.FakeCtrlClient)
		wantErr      error
		assertGot    func(t *testing.T, got *corev1.ConfigMap)
	}{
		{
			name: "cache hit",
			setupClients: func(t *testing.T) (*fakes.FakeCtrlClient, *fakes.FakeCtrlClient) {
				t.Helper()
				cached := &fakes.FakeCtrlClient{}
				uncached := &fakes.FakeCtrlClient{}
				cached.GetCalls(func(_ context.Context, _ types.NamespacedName, obj client.Object) error {
					cm.DeepCopyInto(obj.(*corev1.ConfigMap))
					return nil
				})
				uncached.GetCalls(func(context.Context, types.NamespacedName, client.Object) error {
					t.Fatal("uncached Get should not be called when cache hits")
					return nil
				})
				return cached, uncached
			},
			assertGot: func(t *testing.T, got *corev1.ConfigMap) {
				t.Helper()
				if got.Name != cm.Name {
					t.Fatalf("got ConfigMap name %q, want %q", got.Name, cm.Name)
				}
			},
		},
		{
			name: "cache miss falls back to uncached",
			setupClients: func(t *testing.T) (*fakes.FakeCtrlClient, *fakes.FakeCtrlClient) {
				t.Helper()
				cached := &fakes.FakeCtrlClient{}
				uncached := &fakes.FakeCtrlClient{}
				cached.GetReturns(apierrors.NewNotFound(corev1.Resource("configmaps"), key.Name))
				uncached.GetCalls(func(_ context.Context, _ types.NamespacedName, obj client.Object) error {
					cm.DeepCopyInto(obj.(*corev1.ConfigMap))
					return nil
				})
				return cached, uncached
			},
			assertGot: func(t *testing.T, got *corev1.ConfigMap) {
				t.Helper()
				if got.Data[UserCABundleKeyPath] == "" {
					t.Fatal("expected ConfigMap data from uncached client")
				}
			},
		},
		{
			name: "non-not-found cache error",
			setupClients: func(t *testing.T) (*fakes.FakeCtrlClient, *fakes.FakeCtrlClient) {
				t.Helper()
				cached := &fakes.FakeCtrlClient{}
				uncached := &fakes.FakeCtrlClient{}
				cached.GetReturns(cacheUnavailableErr)
				uncached.GetCalls(func(context.Context, types.NamespacedName, client.Object) error {
					t.Fatal("uncached Get should not be called on non-NotFound cache errors")
					return nil
				})
				return cached, uncached
			},
			wantErr: cacheUnavailableErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cached, uncached := tt.setupClients(t)
			r := &Reconciler{ctx: context.Background(), CtrlClient: cached, UncachedClient: uncached}
			got := &corev1.ConfigMap{}
			err := r.getWithCacheFallback(key, got)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("getWithCacheFallback() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("getWithCacheFallback() error = %v", err)
			}
			if tt.assertGot != nil {
				tt.assertGot(t, got)
			}
		})
	}
}

func TestUpdateWatchLabel(t *testing.T) {
	t.Parallel()

	key := types.NamespacedName{Name: "user-ca", Namespace: OperandDefaultNamespace}
	baseCM := testUserCAConfigMap("user-ca", mustPEMCert(t, true), nil)
	labeledCM := testUserCAConfigMap("user-ca", mustPEMCert(t, true), map[string]string{
		WatchedResourceLabelKey: WatchedResourceLabelValue,
	})

	tests := []struct {
		name        string
		cm          *corev1.ConfigMap
		wantPatch   bool
		assertPatch func(t *testing.T, obj client.Object)
	}{
		{
			name:      "patches watch label",
			cm:        baseCM,
			wantPatch: true,
			assertPatch: func(t *testing.T, obj client.Object) {
				t.Helper()
				labels := obj.GetLabels()
				if labels[WatchedResourceLabelKey] != WatchedResourceLabelValue {
					t.Fatalf("patch labels = %v, want watch label set", labels)
				}
			},
		},
		{
			name:      "skips patch when label already set",
			cm:        labeledCM,
			wantPatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cached := &fakes.FakeCtrlClient{}
			uncached := &fakes.FakeCtrlClient{}
			stubGet := func(_ context.Context, _ types.NamespacedName, obj client.Object) error {
				tt.cm.DeepCopyInto(obj.(*corev1.ConfigMap))
				return nil
			}
			cached.GetCalls(stubGet)
			uncached.GetCalls(stubGet)

			var patched bool
			uncached.PatchCalls(func(_ context.Context, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
				if !tt.wantPatch {
					t.Fatal("Patch should not be called when watch label is already set")
				}
				patched = true
				if tt.assertPatch != nil {
					tt.assertPatch(t, obj)
				}
				return nil
			})

			r := &Reconciler{ctx: context.Background(), CtrlClient: cached, UncachedClient: uncached}
			if err := r.updateWatchLabel(key, &corev1.ConfigMap{}); err != nil {
				t.Fatalf("updateWatchLabel() error = %v", err)
			}
			if tt.wantPatch && !patched {
				t.Fatal("expected Patch to be called")
			}
		})
	}
}
