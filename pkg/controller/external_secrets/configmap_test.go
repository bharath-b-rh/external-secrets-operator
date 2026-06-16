package external_secrets

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/openshift/external-secrets-operator/api/v1alpha1"
	"github.com/openshift/external-secrets-operator/pkg/controller/client/fakes"
	"github.com/openshift/external-secrets-operator/pkg/controller/common"
	"github.com/openshift/external-secrets-operator/pkg/controller/commontest"
)

// testESCWithProxy returns an ExternalSecretsConfig with a proxy configured so that
// ensureTrustedCABundleConfigMap proceeds past the proxy-nil guard.
func testESCWithProxy() *operatorv1alpha1.ExternalSecretsConfig {
	esc := commontest.TestExternalSecretsConfig()
	esc.Spec.ApplicationConfig.Proxy = &operatorv1alpha1.ProxyConfig{
		HTTPProxy: "http://proxy.example.com:3128",
	}
	return esc
}

func TestEnsureTrustedCABundleConfigMap(t *testing.T) {
	cnoData := map[string]string{"ca-bundle.crt": "cert-data"}

	tests := []struct {
		name             string
		resourceMetadata common.ResourceMetadata
		preReq           func(r *Reconciler, cached *fakes.FakeCtrlClient, uncached *fakes.FakeCtrlClient)
		wantErr          string
		wantCreate       bool
		wantUpdate       bool
		noProxy          bool
	}{
		{
			name:             "ConfigMap created when it does not exist",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, _ client.Object) (bool, error) {
					return false, nil
				})
				cached.CreateCalls(func(_ context.Context, obj client.Object, _ ...client.CreateOption) error {
					cm, ok := obj.(*corev1.ConfigMap)
					if !ok {
						t.Errorf("expected ConfigMap, got %T", obj)
					}
					if cm.Name != trustedCABundleConfigMapName {
						t.Errorf("expected name %s, got %s", trustedCABundleConfigMapName, cm.Name)
					}
					if cm.Labels[trustedCABundleInjectLabel] != "true" {
						t.Errorf("expected inject label to be 'true', got %q", cm.Labels[trustedCABundleInjectLabel])
					}
					return nil
				})
			},
			wantCreate: true,
		},
		{
			name:             "ConfigMap exists in correct state, no update needed",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, obj client.Object) (bool, error) {
					existing := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      trustedCABundleConfigMapName,
							Namespace: commontest.TestExternalSecretsNamespace,
							Labels:    getTrustedCABundleLabels(controllerDefaultResourceLabels),
						},
						Data: cnoData,
					}
					existing.DeepCopyInto(obj.(*corev1.ConfigMap))
					return true, nil
				})
			},
			wantUpdate: false,
		},
		{
			name:             "ConfigMap exists with wrong labels, update triggered",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, obj client.Object) (bool, error) {
					existing := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      trustedCABundleConfigMapName,
							Namespace: commontest.TestExternalSecretsNamespace,
							Labels:    map[string]string{"app": "something-else"},
						},
						Data: cnoData,
					}
					existing.DeepCopyInto(obj.(*corev1.ConfigMap))
					return true, nil
				})
				cached.UpdateWithRetryCalls(func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
					cm, ok := obj.(*corev1.ConfigMap)
					if !ok {
						t.Errorf("expected ConfigMap, got %T", obj)
					}
					if cm.Labels[trustedCABundleInjectLabel] != "true" {
						t.Errorf("expected inject label restored, got %q", cm.Labels[trustedCABundleInjectLabel])
					}
					if cm.Data["ca-bundle.crt"] != "cert-data" {
						t.Errorf("CNO-managed data should be preserved, got %v", cm.Data)
					}
					return nil
				})
			},
			wantUpdate: true,
		},
		{
			// Regression test for: labels updating with app: external-secrets of cm
			// external-secrets-trusted-ca-bundle will block the reconcile in the http_proxy env.
			//
			// When the managed label (app=external-secrets) is removed from the ConfigMap,
			// the object falls out of the label-filtered cache. The cached Exists() returns
			// false, Create() fails with AlreadyExists, and the controller must use the
			// uncached client to fetch and restore the correct labels instead of returning
			// an error that blocks reconciliation.
			name:             "Create returns AlreadyExists (label-filtered cache miss): uncached client restores labels",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, uncached *fakes.FakeCtrlClient) {
				// Cached client doesn't see the ConfigMap because its label was changed.
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, _ client.Object) (bool, error) {
					return false, nil
				})
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return apierrors.NewAlreadyExists(schema.GroupResource{Resource: "configmaps"}, trustedCABundleConfigMapName)
				})
				// Uncached client fetches the real object from the API server.
				uncached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, obj client.Object) (bool, error) {
					existing := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:            trustedCABundleConfigMapName,
							Namespace:       commontest.TestExternalSecretsNamespace,
							ResourceVersion: "1000",
							Labels:          map[string]string{"app": "update-secrets"},
						},
						Data: cnoData,
					}
					existing.DeepCopyInto(obj.(*corev1.ConfigMap))
					return true, nil
				})
				uncached.UpdateWithRetryCalls(func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
					cm, ok := obj.(*corev1.ConfigMap)
					if !ok {
						t.Errorf("expected ConfigMap, got %T", obj)
					}
					if cm.Labels["app"] != "external-secrets" {
						t.Errorf("expected app=external-secrets label restored, got %q", cm.Labels["app"])
					}
					if cm.Labels[trustedCABundleInjectLabel] != "true" {
						t.Errorf("expected inject label restored, got %q", cm.Labels[trustedCABundleInjectLabel])
					}
					if cm.Data["ca-bundle.crt"] != "cert-data" {
						t.Errorf("CNO-managed data should be preserved, got %v", cm.Data)
					}
					return nil
				})
			},
			wantUpdate: true,
		},
		{
			name:             "proxy not configured: ConfigMap reconcile skipped",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			noProxy:          true,
		},
		{
			name:             "cached Exists check fails",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, _ client.Object) (bool, error) {
					return false, commontest.ErrTestClient
				})
			},
			wantErr: "failed to check external-secrets/external-secrets-trusted-ca-bundle trusted CA bundle ConfigMap resource already exists: test client error",
		},
		{
			name:             "Create fails with non-AlreadyExists error",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, _ client.Object) (bool, error) {
					return false, nil
				})
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return commontest.ErrTestClient
				})
			},
			wantErr: "failed to create external-secrets/external-secrets-trusted-ca-bundle trusted CA bundle ConfigMap resource: test client error",
		},
		{
			name:             "uncached Exists fails after AlreadyExists from Create",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, uncached *fakes.FakeCtrlClient) {
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, _ client.Object) (bool, error) {
					return false, nil
				})
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return apierrors.NewAlreadyExists(schema.GroupResource{Resource: "configmaps"}, trustedCABundleConfigMapName)
				})
				uncached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, _ client.Object) (bool, error) {
					return false, commontest.ErrTestClient
				})
			},
			wantErr: "failed to fetch existing external-secrets/external-secrets-trusted-ca-bundle trusted CA bundle ConfigMap for restoration: test client error",
		},
		{
			name:             "uncached UpdateWithRetry fails after AlreadyExists from Create",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, uncached *fakes.FakeCtrlClient) {
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, _ client.Object) (bool, error) {
					return false, nil
				})
				cached.CreateCalls(func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
					return apierrors.NewAlreadyExists(schema.GroupResource{Resource: "configmaps"}, trustedCABundleConfigMapName)
				})
				uncached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, obj client.Object) (bool, error) {
					existing := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:            trustedCABundleConfigMapName,
							Namespace:       commontest.TestExternalSecretsNamespace,
							ResourceVersion: "1000",
							Labels:          map[string]string{"app": "update-secrets"},
						},
					}
					existing.DeepCopyInto(obj.(*corev1.ConfigMap))
					return true, nil
				})
				uncached.UpdateWithRetryCalls(func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
					return commontest.ErrTestClient
				})
			},
			wantErr: "failed to restore external-secrets/external-secrets-trusted-ca-bundle trusted CA bundle ConfigMap to desired state: test client error",
		},
		{
			name:             "cached UpdateWithRetry fails when ConfigMap has wrong labels",
			resourceMetadata: testResourceMetadata(commontest.TestExternalSecretsConfig()),
			preReq: func(_ *Reconciler, cached *fakes.FakeCtrlClient, _ *fakes.FakeCtrlClient) {
				cached.ExistsCalls(func(_ context.Context, _ types.NamespacedName, obj client.Object) (bool, error) {
					existing := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      trustedCABundleConfigMapName,
							Namespace: commontest.TestExternalSecretsNamespace,
							Labels:    nil,
						},
					}
					existing.DeepCopyInto(obj.(*corev1.ConfigMap))
					return true, nil
				})
				cached.UpdateWithRetryCalls(func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
					return commontest.ErrTestClient
				})
			},
			wantErr: "failed to update external-secrets/external-secrets-trusted-ca-bundle trusted CA bundle ConfigMap resource: test client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := testReconciler(t)
			cached := &fakes.FakeCtrlClient{}
			uncached := &fakes.FakeCtrlClient{}
			if tt.preReq != nil {
				tt.preReq(r, cached, uncached)
			}
			r.CtrlClient = cached
			r.UncachedClient = uncached

			var esc *operatorv1alpha1.ExternalSecretsConfig
			if tt.noProxy {
				esc = commontest.TestExternalSecretsConfig()
			} else {
				esc = testESCWithProxy()
			}

			err := r.ensureTrustedCABundleConfigMap(esc, tt.resourceMetadata)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("ensureTrustedCABundleConfigMap() err = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ensureTrustedCABundleConfigMap() unexpected error: %v", err)
			}

			if tt.wantCreate && cached.CreateCallCount() == 0 {
				t.Error("expected Create to be called, but it wasn't")
			}
			if tt.wantUpdate && cached.UpdateWithRetryCallCount() == 0 && uncached.UpdateWithRetryCallCount() == 0 {
				t.Error("expected UpdateWithRetry to be called (on cached or uncached client), but it wasn't")
			}
			if !tt.wantUpdate && (cached.UpdateWithRetryCallCount() > 0 || uncached.UpdateWithRetryCallCount() > 0) {
				t.Error("expected no UpdateWithRetry call, but one was made")
			}
		})
	}
}
