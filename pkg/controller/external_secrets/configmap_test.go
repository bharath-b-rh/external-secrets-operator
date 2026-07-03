package external_secrets

import (
	"context"
	"strings"
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

const testEnterpriseCAPEM = `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHBfpEHIha0MA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNVBAMMCXRl
c3Qtcm9vdDAeFw0yNTAxMDEwMDAwMDBaFw0zNTAxMDEwMDAwMDBaMBQxEjAQBgNV
BAMMCXRlc3Qtcm9vdDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABGv4Motp3jQb
Fd4Fz8x0Y3V0Y2gtY2VydC1kYXRhLWZvci11bml0LXRlc3RzLW9ubHk=
-----END CERTIFICATE-----`

func TestValidateEnterpriseCAConfigMap(t *testing.T) {
	ref := &operatorv1alpha1.AdditionalTrustedCAConfigMapRef{
		Name:      "enterprise-ca-bundle",
		Namespace: "corporate-certs",
	}

	tests := []struct {
		name    string
		preReq  func(*fakes.FakeCtrlClient)
		wantErr string
	}{
		{
			name: "missing ConfigMap returns configuration error",
			preReq: func(m *fakes.FakeCtrlClient) {
				m.GetCalls(func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					return apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "configmaps"}, key.Name)
				})
			},
			wantErr: "enterprise CA ConfigMap corporate-certs/enterprise-ca-bundle not found",
		},
		{
			name: "missing key returns irrecoverable error",
			preReq: func(m *fakes.FakeCtrlClient) {
				m.GetCalls(func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					cm := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace},
						Data:       map[string]string{"other-key": testEnterpriseCAPEM},
					}
					cm.DeepCopyInto(obj.(*corev1.ConfigMap))
					return nil
				})
			},
			wantErr: `enterprise CA ConfigMap corporate-certs/enterprise-ca-bundle is missing required key "ca-bundle.crt"`,
		},
		{
			name: "invalid PEM returns irrecoverable error",
			preReq: func(m *fakes.FakeCtrlClient) {
				m.GetCalls(func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					cm := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace},
						Data:       map[string]string{defaultEnterpriseCAConfigMapKey: "not-a-pem-bundle"},
					}
					cm.DeepCopyInto(obj.(*corev1.ConfigMap))
					return nil
				})
			},
			wantErr: `enterprise CA ConfigMap corporate-certs/enterprise-ca-bundle key "ca-bundle.crt" must contain PEM-encoded CA certificates`,
		},
		{
			name: "valid ConfigMap passes validation",
			preReq: func(m *fakes.FakeCtrlClient) {
				m.GetCalls(func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					cm := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace},
						Data:       map[string]string{defaultEnterpriseCAConfigMapKey: testEnterpriseCAPEM},
					}
					cm.DeepCopyInto(obj.(*corev1.ConfigMap))
					return nil
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := testReconciler(t)
			mock := &fakes.FakeCtrlClient{}
			if tt.preReq != nil {
				tt.preReq(mock)
			}
			r.UncachedClient = mock

			err := r.validateEnterpriseCAConfigMap(ref)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateEnterpriseCAConfigMap() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateEnterpriseCAConfigMap() expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateEnterpriseCAConfigMap() error = %v, want message containing %q", err, tt.wantErr)
			}
			if strings.Contains(tt.wantErr, "not found") && !common.IsConfigurationError(err) {
				t.Fatalf("expected configuration error for missing ConfigMap, got %T", err)
			}
			if (strings.Contains(tt.wantErr, "missing required key") || strings.Contains(tt.wantErr, "must contain PEM")) &&
				!common.IsIrrecoverableError(err) {
				t.Fatalf("expected irrecoverable error for invalid enterprise CA ConfigMap, got %T", err)
			}
		})
	}
}

func TestEnsureEnterpriseTrustedCAConfigMap(t *testing.T) {
	esc := commontest.TestExternalSecretsConfig()
	esc.Spec.ApplicationConfig.AdditionalTrustedCAConfigMapRef = &operatorv1alpha1.AdditionalTrustedCAConfigMapRef{
		Name:      "enterprise-ca-bundle",
		Namespace: "corporate-certs",
	}
	resourceMetadata := testResourceMetadata(esc)

	t.Run("creates synced ConfigMap in operand namespace", func(t *testing.T) {
		r := testReconciler(t)
		mock := &fakes.FakeCtrlClient{}
		mock.GetCalls(func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace},
				Data:       map[string]string{defaultEnterpriseCAConfigMapKey: testEnterpriseCAPEM},
			}
			cm.DeepCopyInto(obj.(*corev1.ConfigMap))
			return nil
		})
		mock.ExistsCalls(func(ctx context.Context, key types.NamespacedName, obj client.Object) (bool, error) {
			return false, nil
		})
		var created *corev1.ConfigMap
		mock.CreateCalls(func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
			created = obj.(*corev1.ConfigMap)
			return nil
		})
		r.CtrlClient = mock
		r.UncachedClient = mock

		if err := r.ensureEnterpriseTrustedCAConfigMap(esc, resourceMetadata); err != nil {
			t.Fatalf("ensureEnterpriseTrustedCAConfigMap() error = %v", err)
		}
		if created == nil {
			t.Fatal("expected synced ConfigMap to be created")
		}
		if created.Name != enterpriseTrustedCABundleConfigMapName {
			t.Fatalf("expected synced ConfigMap name %q, got %q", enterpriseTrustedCABundleConfigMapName, created.Name)
		}
		if created.Namespace != commontest.TestExternalSecretsNamespace {
			t.Fatalf("expected synced ConfigMap in operand namespace %q, got %q", commontest.TestExternalSecretsNamespace, created.Namespace)
		}
		if created.Data[defaultEnterpriseCAConfigMapKey] != testEnterpriseCAPEM {
			t.Fatalf("expected synced ConfigMap data to match enterprise source")
		}
	})
}
