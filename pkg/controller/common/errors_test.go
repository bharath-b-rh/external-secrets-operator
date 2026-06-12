package common

import (
	"errors"
	"fmt"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestIsUserConfigurationNotFound(t *testing.T) {
	t.Parallel()

	cmGR := schema.GroupResource{Resource: "configmaps"}
	notFound := apierrors.NewNotFound(cmGR, "missing")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "user configuration error wrapping NotFound",
			err:  NewUserConfigurationError(notFound, "issuer %q not found", "external-secrets/missing"),
			want: true,
		},
		{
			name: "wrapped user configuration NotFound",
			err:  fmt.Errorf("reconcile deployment: %w", NewUserConfigurationError(notFound, "issuer %q not found", "external-secrets/missing")),
			want: true,
		},
		{
			name: "user configuration error invalid PEM",
			err:  NewUserConfigurationError(errors.New("invalid pem"), "invalid issuer configuration"),
			want: false,
		},
		{
			name: "bare NotFound",
			err:  notFound,
			want: false,
		},
		{name: "nil error", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsUserConfigurationNotFound(tt.err); got != tt.want {
				t.Fatalf("IsUserConfigurationNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUserTrustedCABundleNotFound(t *testing.T) {
	t.Parallel()

	cmGR := schema.GroupResource{Resource: "configmaps"}
	notFound := apierrors.NewNotFound(cmGR, "missing")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "trusted CA bundle error wrapping NotFound",
			err:  NewUserTrustedCABundleError(notFound, "trustedCABundle ConfigMap %q not found", "external-secrets/missing"),
			want: true,
		},
		{
			name: "wrapped trusted CA bundle NotFound",
			err:  fmt.Errorf("reconcile deployment: %w", NewUserTrustedCABundleError(notFound, "trustedCABundle ConfigMap %q not found", "external-secrets/missing")),
			want: true,
		},
		{
			name: "trusted CA bundle error missing key",
			err: NewUserTrustedCABundleError(
				fmt.Errorf("key %q not found", "ca.crt"),
				"trustedCABundle ConfigMap %q does not contain key %q",
				"external-secrets/cm", "ca.crt",
			),
			want: false,
		},
		{
			name: "user configuration NotFound is not trusted CA NotFound",
			err:  NewUserConfigurationError(notFound, "issuer %q not found", "external-secrets/missing"),
			want: false,
		},
		{name: "nil error", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsUserTrustedCABundleNotFound(tt.err); got != tt.want {
				t.Fatalf("IsUserTrustedCABundleNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUserConfigurationError(t *testing.T) {
	t.Parallel()

	notFound := apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "missing")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "user configuration error",
			err:  NewUserConfigurationError(notFound, "issuer %q not found", "external-secrets/missing"),
			want: true,
		},
		{
			name: "trusted CA bundle error is not user configuration",
			err:  NewUserTrustedCABundleError(notFound, "trustedCABundle ConfigMap %q not found", "external-secrets/missing"),
			want: false,
		},
		{name: "nil error", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsUserConfigurationError(tt.err); got != tt.want {
				t.Fatalf("IsUserConfigurationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUserTrustedCABundleError(t *testing.T) {
	t.Parallel()

	notFound := apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "missing")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "trusted CA bundle error",
			err:  NewUserTrustedCABundleError(notFound, "trustedCABundle ConfigMap %q not found", "external-secrets/missing"),
			want: true,
		},
		{
			name: "user configuration error is not trusted CA bundle",
			err:  NewUserConfigurationError(notFound, "issuer %q not found", "external-secrets/missing"),
			want: false,
		},
		{name: "nil error", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsUserTrustedCABundleError(tt.err); got != tt.want {
				t.Fatalf("IsUserTrustedCABundleError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIrrecoverableError(t *testing.T) {
	t.Parallel()

	cmGR := schema.GroupResource{Resource: "configmaps"}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "irrecoverable reconcile error",
			err:  NewIrrecoverableError(errors.New("bad config"), "invalid configuration"),
			want: true,
		},
		{
			name: "user configuration error",
			err:  NewUserConfigurationError(apierrors.NewNotFound(cmGR, "missing"), "issuer %q not found", "external-secrets/missing"),
			want: false,
		},
		{
			name: "trusted CA bundle error",
			err:  NewUserTrustedCABundleError(apierrors.NewNotFound(cmGR, "missing"), "trustedCABundle ConfigMap %q not found", "external-secrets/missing"),
			want: false,
		},
		{name: "nil error", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsIrrecoverableError(tt.err); got != tt.want {
				t.Fatalf("IsIrrecoverableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromClientError(t *testing.T) {
	t.Parallel()

	cmGR := schema.GroupResource{Resource: "configmaps"}

	tests := []struct {
		name       string
		err        error
		wantReason ErrorReason
		wantNil    bool
	}{
		{name: "nil error", err: nil, wantNil: true},
		{
			name:       "NotFound",
			err:        apierrors.NewNotFound(cmGR, "user-ca"),
			wantReason: RetryRequiredError,
		},
		{
			name:       "Forbidden",
			err:        apierrors.NewForbidden(cmGR, "user-ca", errors.New("denied")),
			wantReason: IrrecoverableError,
		},
		{
			name:       "ServiceUnavailable",
			err:        apierrors.NewServiceUnavailable("service unavailable"),
			wantReason: RetryRequiredError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FromClientError(tt.err, "operation failed for %q", "user-ca")
			if tt.wantNil {
				if got != nil {
					t.Fatalf("FromClientError() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("FromClientError() = nil, want reconcile error")
			}
			if got.Reason != tt.wantReason {
				t.Fatalf("Reason = %q, want %q", got.Reason, tt.wantReason)
			}
		})
	}
}

func TestReconcileErrorConstructorsNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  *ReconcileError
	}{
		{name: "NewIrrecoverableError", got: NewIrrecoverableError(nil, "msg")},
		{name: "NewRetryRequiredError", got: NewRetryRequiredError(nil, "msg")},
		{name: "NewUserConfigurationError", got: NewUserConfigurationError(nil, "msg")},
		{name: "NewUserTrustedCABundleError", got: NewUserTrustedCABundleError(nil, "msg")},
		{name: "FromClientError", got: FromClientError(nil, "msg")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != nil {
				t.Fatalf("%s(nil) = %v, want nil", tt.name, tt.got)
			}
		})
	}
}
