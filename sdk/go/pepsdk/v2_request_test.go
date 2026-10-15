package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorizeV2CarriesPrincipalAndRisk(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/policy/decisions/evaluate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","result":"allow","evidenceRefs":["e"]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	_, err := c.Authorize(context.Background(),
		Route{Audience: "au", Action: "tool:invoke", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
		AuthorizationInput{
			Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2,
			Principal: &PolicyPrincipalV2{Namespace: "ns", ActorRef: "ar", SubjectRef: "user-1"},
			RiskLevel: "high", RiskRef: "rr",
		})
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	principal, ok := body["principal"].(map[string]any)
	if !ok || principal["namespace"] != "ns" || principal["actorRef"] != "ar" || principal["subjectRef"] != "user-1" {
		t.Fatalf("principal = %v", body["principal"])
	}
	risk, ok := body["riskContext"].(map[string]any)
	if !ok || risk["riskLevel"] != "high" || risk["riskRef"] != "rr" {
		t.Fatalf("riskContext = %v", body["riskContext"])
	}
}
