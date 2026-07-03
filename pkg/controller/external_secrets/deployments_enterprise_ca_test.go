package external_secrets

import (
	"context"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/external-secrets-operator/api/v1alpha1"
	"github.com/openshift/external-secrets-operator/pkg/controller/client/fakes"
)

func setupEnterpriseCAGetStub(m *fakes.FakeCtrlClient) {
	m.GetCalls(func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
		cm := &corev1.ConfigMap{
			Data: map[string]string{defaultEnterpriseCAConfigMapKey: testEnterpriseCAPEM},
		}
		cm.DeepCopyInto(obj.(*corev1.ConfigMap))
		return nil
	})
}

func TestUpdateOperandTrustedCAVolumesEnterprise(t *testing.T) {
	trustedMount := corev1.VolumeMount{
		Name:      trustedCABundleVolumeName,
		MountPath: trustedCABundleMountPath,
		ReadOnly:  true,
	}

	baseDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: controllerContainerName}},
					},
				},
			},
		}
	}

	expectedEnterpriseVolume := corev1.Volume{
		Name: trustedCABundleVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: enterpriseTrustedCABundleConfigMapName},
				Items: []corev1.KeyToPath{
					{Key: defaultEnterpriseCAConfigMapKey, Path: defaultEnterpriseCAConfigMapKey},
				},
			},
		},
	}

	expectedProjectedVolume := corev1.Volume{
		Name: trustedCABundleVolumeName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{
						ConfigMap: &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{Name: trustedCABundleConfigMapName},
							Optional:             ptr.To(true),
						},
					},
					{
						ConfigMap: &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{Name: enterpriseTrustedCABundleConfigMapName},
							Items: []corev1.KeyToPath{
								{Key: defaultEnterpriseCAConfigMapKey, Path: enterpriseTrustedCAProjectedFileName},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		assetName      string
		esc            *v1alpha1.ExternalSecretsConfig
		setupUncached  func(*fakes.FakeCtrlClient)
		wantErr        bool
		expectedVolume *corev1.Volume
	}{
		{
			name:      "enterprise only mounts synced ConfigMap on controller",
			assetName: controllerDeploymentAssetName,
			esc: &v1alpha1.ExternalSecretsConfig{
				Spec: v1alpha1.ExternalSecretsConfigSpec{
					ApplicationConfig: v1alpha1.ApplicationConfig{
						CommonConfigs: v1alpha1.CommonConfigs{
							AdditionalTrustedCAConfigMapRef: &v1alpha1.AdditionalTrustedCAConfigMapRef{
								Name:      "enterprise-ca-bundle",
								Namespace: "corporate-certs",
							},
						},
					},
				},
			},
			setupUncached:  setupEnterpriseCAGetStub,
			expectedVolume: &expectedEnterpriseVolume,
		},
		{
			name:      "proxy and enterprise use projected volume on webhook",
			assetName: webhookDeploymentAssetName,
			esc: &v1alpha1.ExternalSecretsConfig{
				Spec: v1alpha1.ExternalSecretsConfigSpec{
					ApplicationConfig: v1alpha1.ApplicationConfig{
						CommonConfigs: v1alpha1.CommonConfigs{
							Proxy: &v1alpha1.ProxyConfig{HTTPProxy: "http://proxy:8080"},
							AdditionalTrustedCAConfigMapRef: &v1alpha1.AdditionalTrustedCAConfigMapRef{
								Name:      "enterprise-ca-bundle",
								Namespace: "corporate-certs",
							},
						},
					},
				},
			},
			setupUncached:  setupEnterpriseCAGetStub,
			expectedVolume: &expectedProjectedVolume,
		},
		{
			name:      "enterprise ref does not mount on cert-controller",
			assetName: certControllerDeploymentAssetName,
			esc: &v1alpha1.ExternalSecretsConfig{
				Spec: v1alpha1.ExternalSecretsConfigSpec{
					ApplicationConfig: v1alpha1.ApplicationConfig{
						CommonConfigs: v1alpha1.CommonConfigs{
							AdditionalTrustedCAConfigMapRef: &v1alpha1.AdditionalTrustedCAConfigMapRef{
								Name:      "enterprise-ca-bundle",
								Namespace: "corporate-certs",
							},
						},
					},
				},
			},
			setupUncached: setupEnterpriseCAGetStub,
		},
		{
			name:      "missing enterprise ConfigMap returns error",
			assetName: controllerDeploymentAssetName,
			esc: &v1alpha1.ExternalSecretsConfig{
				Spec: v1alpha1.ExternalSecretsConfigSpec{
					ApplicationConfig: v1alpha1.ApplicationConfig{
						CommonConfigs: v1alpha1.CommonConfigs{
							AdditionalTrustedCAConfigMapRef: &v1alpha1.AdditionalTrustedCAConfigMapRef{
								Name:      "missing",
								Namespace: "corporate-certs",
							},
						},
					},
				},
			},
			setupUncached: func(m *fakes.FakeCtrlClient) {
				m.GetCalls(func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					return apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "configmaps"}, key.Name)
				})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := testReconciler(t)
			mock := &fakes.FakeCtrlClient{}
			if tt.setupUncached != nil {
				tt.setupUncached(mock)
			}
			r.UncachedClient = mock

			deployment := baseDeployment()
			err := r.updateProxyConfiguration(deployment, tt.esc, tt.assetName)
			if tt.wantErr {
				if err == nil {
					t.Fatal("updateProxyConfiguration() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("updateProxyConfiguration() unexpected error: %v", err)
			}

			if tt.expectedVolume == nil {
				for _, volume := range deployment.Spec.Template.Spec.Volumes {
					if volume.Name == trustedCABundleVolumeName {
						t.Fatalf("expected no trusted CA volume, found %+v", volume)
					}
				}
				return
			}

			if len(deployment.Spec.Template.Spec.Volumes) != 1 {
				t.Fatalf("expected one volume, got %+v", deployment.Spec.Template.Spec.Volumes)
			}
			if !reflect.DeepEqual(deployment.Spec.Template.Spec.Volumes[0], *tt.expectedVolume) {
				t.Fatalf("volume mismatch.\nExpected: %+v\nActual: %+v", *tt.expectedVolume, deployment.Spec.Template.Spec.Volumes[0])
			}
			if !reflect.DeepEqual(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, []corev1.VolumeMount{trustedMount}) {
				t.Fatalf("volume mount mismatch: %+v", deployment.Spec.Template.Spec.Containers[0].VolumeMounts)
			}
		})
	}
}

func TestUpdateOperandTrustedCAVolumesRemoval(t *testing.T) {
	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: controllerContainerName,
							VolumeMounts: []corev1.VolumeMount{
								{Name: trustedCABundleVolumeName, MountPath: trustedCABundleMountPath, ReadOnly: true},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: trustedCABundleVolumeName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: enterpriseTrustedCABundleConfigMapName},
								},
							},
						},
					},
				},
			},
		},
	}

	r := testReconciler(t)
	esc := &v1alpha1.ExternalSecretsConfig{}
	if err := r.updateProxyConfiguration(deployment, esc, controllerDeploymentAssetName); err != nil {
		t.Fatalf("updateProxyConfiguration() error = %v", err)
	}
	if len(deployment.Spec.Template.Spec.Volumes) != 0 {
		t.Fatalf("expected trusted CA volume removed, got %+v", deployment.Spec.Template.Spec.Volumes)
	}
	if len(deployment.Spec.Template.Spec.Containers[0].VolumeMounts) != 0 {
		t.Fatalf("expected trusted CA volume mount removed, got %+v", deployment.Spec.Template.Spec.Containers[0].VolumeMounts)
	}
}
