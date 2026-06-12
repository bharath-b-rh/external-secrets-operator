package external_secrets

import (
	"errors"
	"strings"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	operatorv1alpha1 "github.com/openshift/external-secrets-operator/api/v1alpha1"
	"github.com/openshift/external-secrets-operator/pkg/controller/common"
)

func TestTrustedCABundleDeploymentFailureConditions(t *testing.T) {
	t.Parallel()

	const observedGeneration int64 = 5

	tests := []struct {
		name                 string
		reconcileErr         error
		isFatal              bool
		isTrustedCABundle    bool
		wantDegradedStatus   metav1.ConditionStatus
		wantReadyStatus      metav1.ConditionStatus
		wantReadyReason      string
		wantDegradedContains string
	}{
		{
			name: "trusted CA bundle error",
			reconcileErr: common.NewUserTrustedCABundleError(
				errors.New("invalid pem"),
				"trustedCABundle ConfigMap %q key %q has invalid PEM",
				"external-secrets/user-ca", "ca-bundle.crt",
			),
			isTrustedCABundle:    true,
			wantDegradedStatus:   metav1.ConditionTrue,
			wantReadyStatus:      metav1.ConditionFalse,
			wantReadyReason:      operatorv1alpha1.ReasonFailed,
			wantDegradedContains: "trustedCABundle configuration is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			degraded, ready := deploymentFailureConditions(observedGeneration, tt.reconcileErr, tt.isFatal, false, tt.isTrustedCABundle)

			if degraded.ObservedGeneration != observedGeneration {
				t.Fatalf("degraded ObservedGeneration = %d, want %d", degraded.ObservedGeneration, observedGeneration)
			}
			if ready.ObservedGeneration != observedGeneration {
				t.Fatalf("ready ObservedGeneration = %d, want %d", ready.ObservedGeneration, observedGeneration)
			}
			if degraded.Status != tt.wantDegradedStatus {
				t.Fatalf("degraded status = %q, want %q", degraded.Status, tt.wantDegradedStatus)
			}
			if ready.Status != tt.wantReadyStatus {
				t.Fatalf("ready status = %q, want %q", ready.Status, tt.wantReadyStatus)
			}
			if tt.wantReadyReason != "" && ready.Reason != tt.wantReadyReason {
				t.Fatalf("ready reason = %q, want %q", ready.Reason, tt.wantReadyReason)
			}
			if tt.wantDegradedContains != "" && !strings.Contains(degraded.Message, tt.wantDegradedContains) {
				t.Fatalf("degraded message = %q, want substring %q", degraded.Message, tt.wantDegradedContains)
			}
		})
	}
}

func TestTrustedCABundleFailureResult(t *testing.T) {
	t.Parallel()

	cmGR := schema.GroupResource{Resource: "configmaps"}
	notFoundErr := common.NewUserTrustedCABundleError(
		apierrors.NewNotFound(cmGR, "user-ca"),
		"trustedCABundle ConfigMap %q not found",
		"external-secrets/user-ca",
	)
	invalidErr := common.NewUserTrustedCABundleError(
		errors.New("invalid pem"),
		"trustedCABundle ConfigMap %q key %q has invalid PEM",
		"external-secrets/user-ca", "ca-bundle.crt",
	)

	tests := []struct {
		name            string
		reconcileErr    error
		errUpdate       error
		wantRequeue     time.Duration
		wantReturnError error
	}{
		{
			name:         "NotFound requeues",
			reconcileErr: notFoundErr,
			wantRequeue:  common.DefaultRequeueTime,
		},
		{
			name:         "invalid PEM waits for ConfigMap watch",
			reconcileErr: invalidErr,
		},
		{
			name:            "status update error is returned",
			reconcileErr:    invalidErr,
			errUpdate:       errors.New("status update failed"),
			wantReturnError: errors.New("status update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := trustedCABundleFailureResult(tt.reconcileErr, tt.errUpdate)
			if tt.wantReturnError != nil {
				if err == nil || err.Error() != tt.wantReturnError.Error() {
					t.Fatalf("trustedCABundleFailureResult() err = %v, want %v", err, tt.wantReturnError)
				}
				return
			}
			if err != nil {
				t.Fatalf("trustedCABundleFailureResult() unexpected err = %v", err)
			}
			if result.RequeueAfter != tt.wantRequeue {
				t.Fatalf("RequeueAfter = %v, want %v", result.RequeueAfter, tt.wantRequeue)
			}
		})
	}
}
