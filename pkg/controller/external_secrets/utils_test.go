package external_secrets

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
