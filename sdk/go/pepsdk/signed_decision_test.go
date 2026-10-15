package pepsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorizeExposesSignedDecision(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/policy/decisions/evaluate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"api_version":"aegivela.io/v1alpha1","decision_id":"d1","outcome":"allow","policy_version":"v1","evidence_refs":["e"],"signed_decision":"jws-artifact"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	res, err := c.Authorize(context.Background(),
		Route{Audience: "au", Action: "tool:invoke", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
		AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt"})
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if res.SignedDecision != "jws-artifact" {
		t.Fatalf("SignedDecision = %q, want jws-artifact", res.SignedDecision)
	}
}
