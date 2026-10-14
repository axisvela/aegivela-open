package governancesdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPoliciesWebAndRisk(t *testing.T) {
	var policyBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/policies":
			_ = json.NewDecoder(r.Body).Decode(&policyBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"policy_id":"p1","version":"v1","rules":[]}`))
		case r.URL.Path == "/v1/web-permissions:resolve":
			_, _ = w.Write([]byte(`{"api_version":"aegivela.io/v1alpha1","permissions":[]}`))
		case r.URL.Path == "/v1/web-resources/moduregis":
			_, _ = w.Write([]byte(`[{"resource_id":"console"}]`))
		case r.URL.Path == "/v1/risk/signals":
			w.WriteHeader(http.StatusAccepted)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())

	if _, err := c.CreatePolicy(context.Background(), "itok", CreatePolicyRequest{
		PolicyID: "p1", Version: "v1", Rules: json.RawMessage(`[]`), ActivatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}
	if policyBody["policy_id"] != "p1" {
		t.Errorf("policy body = %v", policyBody)
	}
	if raw, err := c.ResolveWebPermissions(context.Background(), "itok", "moduregis", "console.catalog"); err != nil || !strings.Contains(string(raw), "permissions") {
		t.Fatalf("permissions=%s err=%v", raw, err)
	}
	if raw, err := c.ListWebResources(context.Background(), "itok", "moduregis", ""); err != nil || !strings.Contains(string(raw), "console") {
		t.Fatalf("resources=%s err=%v", raw, err)
	}
	if err := c.SubmitRiskSignal(context.Background(), "itok", RiskSignalRequest{SubjectRef: "s", RiskRef: "r"}); err != nil {
		t.Fatalf("SubmitRiskSignal: %v", err)
	}
}
