package external_secrets

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		name       string
		preReq     func(cached *fakes.FakeCtrlClient, uncached *fakes.FakeCtrlClient)
		wantErr    string
		wantUpdate bool
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
			wantErr: "failed to create test-ns/test-sa: test client error",
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
			wantErr:    "failed to restore test-ns/test-sa to desired state: test client error",
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

			desired := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sa",
					Namespace: "test-ns",
					Labels:    controllerDefaultResourceLabels,
				},
			}

			err := r.createWithFallback(desired, resourceMetadata, "test-ns/test-sa")

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
