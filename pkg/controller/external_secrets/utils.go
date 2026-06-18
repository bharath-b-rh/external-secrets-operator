package external_secrets

import (
	"context"
	"fmt"
	"net/url"
	"os"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
