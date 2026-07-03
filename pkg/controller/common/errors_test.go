package common

import (
	"fmt"
	"testing"
)

func TestConfigurationError(t *testing.T) {
	err := NewConfigurationError(fmt.Errorf("configmaps \"enterprise-ca\" not found"), "enterprise CA ConfigMap corporate-certs/enterprise-ca not found")
	if err == nil {
		t.Fatal("expected configuration error")
	}
	if !IsConfigurationError(err) {
		t.Fatalf("expected IsConfigurationError true, got false")
	}
	if IsIrrecoverableError(err) {
		t.Fatalf("configuration error should not be irrecoverable")
	}
}
