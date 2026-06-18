package external_secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"go.uber.org/zap/zapcore"

	operatorv1alpha1 "github.com/openshift/external-secrets-operator/api/v1alpha1"
	"github.com/openshift/external-secrets-operator/pkg/controller/common"
)

func getNamespace(_ *operatorv1alpha1.ExternalSecretsConfig) string {
	return externalsecretsDefaultNamespace
}

func updateNamespace(obj client.Object, esc *operatorv1alpha1.ExternalSecretsConfig) {
	obj.SetNamespace(getNamespace(esc))
}

func containsProcessedAnnotation(esc *operatorv1alpha1.ExternalSecretsConfig) bool {
	_, exist := esc.GetAnnotations()[controllerProcessedAnnotation]
	return exist
}

func addProcessedAnnotation(esc *operatorv1alpha1.ExternalSecretsConfig) bool {
	annotations := esc.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string, 1)
	}
	if _, exist := annotations[controllerProcessedAnnotation]; !exist {
		annotations[controllerProcessedAnnotation] = "true"
		esc.SetAnnotations(annotations)
		return true
	}
	return false
}

func (r *Reconciler) updateCondition(esc *operatorv1alpha1.ExternalSecretsConfig, prependErr error) error {
	if err := r.updateStatus(r.ctx, esc); err != nil {
		errUpdate := fmt.Errorf("failed to update %s/%s status: %w", esc.GetNamespace(), esc.GetName(), err)
		if prependErr != nil {
			return utilerrors.NewAggregate([]error{prependErr, errUpdate})
		}
		return errUpdate
	}
	return nil
}

// updateStatus is for updating the status subresource of externalsecretsconfigs.operator.openshift.io.
func (r *Reconciler) updateStatus(ctx context.Context, changed *operatorv1alpha1.ExternalSecretsConfig) error {
	namespacedName := client.ObjectKeyFromObject(changed)
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		r.log.V(4).Info("updating externalsecretsconfigs.operator.openshift.io status", "request", namespacedName)
		current := &operatorv1alpha1.ExternalSecretsConfig{}
		if err := r.Get(ctx, namespacedName, current); err != nil {
			return fmt.Errorf("failed to fetch externalsecretsconfigs.operator.openshift.io %q for status update: %w", namespacedName, err)
		}
		changed.Status.DeepCopyInto(&current.Status)

		if err := r.StatusUpdate(ctx, current); err != nil {
			return fmt.Errorf("failed to update externalsecretsconfigs.operator.openshift.io %q status: %w", namespacedName, err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// validateExternalSecretsConfig is for validating the ExternalSecretsConfig CR fields, apart from the
// CEL validations present in CRD.
func (r *Reconciler) validateExternalSecretsConfig(esc *operatorv1alpha1.ExternalSecretsConfig) error {
	if isCertManagerConfigEnabled(esc) {
		if _, ok := r.optionalResourcesList[certificateCRDGKV]; !ok {
			return fmt.Errorf("spec.controllerConfig.certProvider.certManager.mode is set, but cert-manager is not installed")
		}
	}
	return nil
}

// isCertManagerConfigEnabled returns whether CertManagerConfig is enabled in ExternalSecretsConfig CR Spec.
func isCertManagerConfigEnabled(esc *operatorv1alpha1.ExternalSecretsConfig) bool {
	return esc.Spec.ControllerConfig.CertProvider != nil &&
		esc.Spec.ControllerConfig.CertProvider.CertManager != nil &&
		common.EvalMode(esc.Spec.ControllerConfig.CertProvider.CertManager.Mode)
}

// isBitwardenConfigEnabled returns whether BitwardenSecretManagerProvider is enabled in ExternalSecretsConfig CR Spec.
func isBitwardenConfigEnabled(esc *operatorv1alpha1.ExternalSecretsConfig) bool {
	return esc.Spec.Plugins.BitwardenSecretManagerProvider != nil &&
		common.EvalMode(esc.Spec.Plugins.BitwardenSecretManagerProvider.Mode)
}

func getLogLevel(esc *operatorv1alpha1.ExternalSecretsConfig, esm *operatorv1alpha1.ExternalSecretsManager) string {
	var logLevel int32 = 1
	if esc.Spec.ApplicationConfig.LogLevel != 0 {
		logLevel = esc.Spec.ApplicationConfig.LogLevel
	} else if esm.Spec.GlobalConfig != nil && esm.Spec.GlobalConfig.LogLevel != 0 {
		logLevel = esm.Spec.GlobalConfig.LogLevel
	}
	switch logLevel {
	case 0, 1, 2:
		return zapcore.Level(logLevel).String()
	case 4, 5:
		return zapcore.DebugLevel.String()
	}
	return zapcore.InfoLevel.String()
}

func getOperatingNamespace(esc *operatorv1alpha1.ExternalSecretsConfig) string {
	return esc.Spec.ApplicationConfig.OperatingNamespace
}

func (r *Reconciler) IsCertManagerInstalled() bool {
	_, ok := r.optionalResourcesList[certificateCRDGKV]
	return ok
}

// hasProxyURLs reports whether a ProxyConfig carries at least one proxy URL
// (HTTPProxy, HTTPSProxy, or NoProxy). Control-plane fields like
// NetworkPolicyProvisioning are intentionally excluded — they configure
// operator behavior, not proxy endpoints.
func hasProxyURLs(p *operatorv1alpha1.ProxyConfig) bool {
	return p != nil && (p.HTTPProxy != "" || p.HTTPSProxy != "" || p.NoProxy != "")
}

// getProxyConfiguration returns the proxy configuration based on precedence.
// The precedence order is: ExternalSecretsConfig > ExternalSecretsManager > OLM environment variables.
// Only proxy URL fields (HTTPProxy, HTTPSProxy, NoProxy) participate in the
// precedence decision; a ProxyConfig that carries only control-plane fields
// (e.g. NetworkPolicyProvisioning) is treated as empty and falls through to the
// next source. After resolving the URLs, any NetworkPolicyProvisioning value set
// on the ExternalSecretsConfig CR is merged into the result so that administrators
// can control network-policy management independently of where the proxy URLs originate.
func (r *Reconciler) getProxyConfiguration(esc *operatorv1alpha1.ExternalSecretsConfig) (*operatorv1alpha1.ProxyConfig, error) {
	var proxyConfig *operatorv1alpha1.ProxyConfig

	switch {
	case hasProxyURLs(esc.Spec.ApplicationConfig.Proxy):
		proxyConfig = esc.Spec.ApplicationConfig.Proxy
	case r.esm.Spec.GlobalConfig != nil && hasProxyURLs(r.esm.Spec.GlobalConfig.Proxy):
		proxyConfig = r.esm.Spec.GlobalConfig.Proxy
	default:
		// Fall back to OLM environment variables
		olmHTTPProxy := os.Getenv(httpProxyEnvVar)
		olmHTTPSProxy := os.Getenv(httpsProxyEnvVar)
		olmNoProxy := os.Getenv(noProxyEnvVar)

		// Only create proxy config if at least one OLM env var is set
		if olmHTTPProxy != "" || olmHTTPSProxy != "" || olmNoProxy != "" {
			proxyConfig = &operatorv1alpha1.ProxyConfig{
				HTTPProxy:  olmHTTPProxy,
				HTTPSProxy: olmHTTPSProxy,
				NoProxy:    olmNoProxy,
			}
		}
	}

	// Merge NetworkPolicyProvisioning from the CR even when proxy URLs came
	// from a lower-priority source, so administrators can control network-policy
	// management independently.
	if crProxy := esc.Spec.ApplicationConfig.Proxy; crProxy != nil && proxyConfig != nil {
		if crProxy.NetworkPolicyProvisioning != "" {
			proxyConfig.NetworkPolicyProvisioning = crProxy.NetworkPolicyProvisioning
		}
	}

	if proxyConfig != nil {
		if err := validateProxy(proxyConfig.HTTPProxy); err != nil {
			return nil, fmt.Errorf("failed to validate HTTP proxy: %w", err)
		}
		if err := validateProxy(proxyConfig.HTTPSProxy); err != nil {
			return nil, fmt.Errorf("failed to validate HTTPS proxy: %w", err)
		}
	}

	return proxyConfig, nil
}

// validateProxy checks if the proxy address configured is a valid URL.
func validateProxy(rawURL string) error {
	if rawURL == "" {
		return nil
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL configured: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("proxy URL must include valid scheme and host")
	}

	return nil
}

// createWithFallback attempts to create a resource and handles the AlreadyExists error that
// occurs when the resource exists on the API server but is invisible to the label-filtered
// informer cache (e.g. because the managed label app=external-secrets was externally removed).
//
// When Create returns AlreadyExists, this method bypasses the stale cache by using the
// uncached client to update the resource directly on the API server, restoring the managed
// labels and annotations to the desired state.
//
// It records a "created" event on success, or a "restored" event when the AlreadyExists
// fallback path is taken.
func (r *Reconciler) createWithFallback(desired client.Object, resourceMetadata common.ResourceMetadata, resourceName string, esc *operatorv1alpha1.ExternalSecretsConfig) error {
	kind := desired.GetObjectKind().GroupVersionKind().Kind
	if kind == "" {
		gvk, gvkErr := apiutil.GVKForObject(desired, r.Scheme)
		if gvkErr != nil {
			r.log.V(5).Info("could not determine GVK, falling back to Go type name",
				"type", fmt.Sprintf("%T", desired), "err", gvkErr)
			kind = fmt.Sprintf("%T", desired)
		} else {
			kind = gvk.Kind
		}
	}

	if err := r.Create(r.ctx, desired); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return common.FromClientError(err, "failed to create %s %s", kind, resourceName)
		}

		r.log.V(1).Info("resource exists on API server but absent from label-filtered cache, restoring desired state",
			"kind", kind, "name", resourceName)
		common.RemoveObsoleteAnnotations(desired, resourceMetadata)
		if err := r.UncachedClient.UpdateWithRetry(r.ctx, desired); err != nil {
			return common.FromClientError(err, "failed to restore %s %s to desired state", kind, resourceName)
		}
		r.eventRecorder.Eventf(esc, corev1.EventTypeNormal, "Reconciled", "%s resource %s restored to desired state", kind, resourceName)
		return nil
	}
	r.eventRecorder.Eventf(esc, corev1.EventTypeNormal, "Reconciled", "%s resource %s created", kind, resourceName)
	return nil
}

// jsonPatchOp is a single JSON Patch operation.
type jsonPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// patchResourceMetadata fully replaces the labels and annotations on an object using a
// JSON Patch, leaving all other fields untouched. This is safe for co-managed
// resources where this operator owns all metadata but data fields are owned by an external
// controller (e.g. CNO-managed ConfigMap data, cert-controller-managed TLS Secret data).
//
// JSON Patch "add" on an existing path replaces the entire value, so the resulting
// labels/annotations on the server exactly match desired.
//
// RemoveObsoleteAnnotations is called defensively to strip any deleted annotation keys
// from desired before building the patch, regardless of whether the caller already did so.
func (r *Reconciler) patchResourceMetadata(desired client.Object, resourceMetadata common.ResourceMetadata) error {
	common.RemoveObsoleteAnnotations(desired, resourceMetadata)

	annotations := desired.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	labels := desired.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	ops := []jsonPatchOp{
		{Op: "add", Path: "/metadata/labels", Value: labels},
		{Op: "add", Path: "/metadata/annotations", Value: annotations},
	}
	patchBytes, err := json.Marshal(ops)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata patch: %w", err)
	}
	return r.CtrlClient.Patch(r.ctx, desired, client.RawPatch(types.JSONPatchType, patchBytes))
}
