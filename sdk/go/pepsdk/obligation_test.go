package pepsdk

import (
	"errors"
	"testing"
)

func TestEnforceObligationsFailsClosed(t *testing.T) {
	t.Parallel()

	decision := GatewayConnectDecision{Obligations: []string{"credential-injection"}, CredentialRef: "cred-1"}
	if err := decision.EnforceObligations(nil); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unenforced named obligation error = %v, want ErrUnavailable", err)
	}
	if err := decision.EnforceObligations(map[string]bool{"credential-injection": true}); err != nil {
		t.Fatalf("enforced named obligation error = %v, want nil", err)
	}

	mitm := GatewayConnectDecision{MitmRequired: true}
	if err := mitm.EnforceObligations(map[string]bool{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unenforced mitm error = %v, want ErrUnavailable", err)
	}
	if err := mitm.EnforceObligations(map[string]bool{"mitm-inspection": true}); err != nil {
		t.Fatalf("enforced mitm error = %v, want nil", err)
	}
}
