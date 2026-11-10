package pepsdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func moduregisTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(handler))
	client := NewClient(server.URL, server.Client())
	t.Cleanup(server.Close)
	return server, client
}

func validModuregisInput() ModuregisAuthorizeRequest {
	return ModuregisAuthorizeRequest{
		Action:            "capability:read",
		ResourceKind:      "moduregis:capability",
		ResourceReference: "cap-1",
		BearerToken:       "bearer-token",
		Scope:             []string{"capability:read"},
	}
}

func TestModuregisAuthorizeAllow(t *testing.T) {
	_, client := moduregisTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Aegivela-Internal-Token") == "" {
			t.Error("missing internal token header")
		}
		response := map[string]any{
			"api_version": "aegivela.io/v1alpha1", "decision_id": "d-1", "outcome": "allow",
			"policy_version": "p-1", "tenant_id": "tenant-a", "actor_id": "actor-1",
			"evidence_refs": []string{"evidence://e/1"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})
	decision, err := client.ModuregisAuthorize(context.Background(), "internal-token", validModuregisInput())
	if err != nil {
		t.Fatalf("ModuregisAuthorize() = %v", err)
	}
	if decision.DecisionID != "d-1" || decision.Outcome != "allow" || decision.PolicyVersion != "p-1" {
		t.Fatalf("decision = %+v", decision)
	}
	if decision.TenantID != "tenant-a" || decision.ActorID != "actor-1" {
		t.Fatalf("tenant/actor = %s/%s", decision.TenantID, decision.ActorID)
	}
}

func TestModuregisAuthorizeDenyMapsToErrDenied(t *testing.T) {
	_, client := moduregisTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"api_version": "aegivela.io/v1alpha1", "decision_id": "d-1", "outcome": "deny",
			"policy_version": "p-1", "tenant_id": "tenant-a", "actor_id": "actor-1",
			"evidence_refs": []string{"evidence://e/1"},
		}
		_ = json.NewEncoder(w).Encode(response)
	})
	_, err := client.ModuregisAuthorize(context.Background(), "internal-token", validModuregisInput())
	if err != ErrDenied {
		t.Fatalf("deny = %v, want ErrDenied", err)
	}
}

func TestModuregisAuthorizeStatusMapping(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusUnauthorized, ErrUnauthenticated},
		{http.StatusForbidden, ErrDenied},
		{http.StatusBadRequest, ErrInvalidInput},
		{http.StatusServiceUnavailable, ErrUnavailable},
	}
	for _, test := range cases {
		_, client := moduregisTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(test.status)
		})
		_, err := client.ModuregisAuthorize(context.Background(), "internal-token", validModuregisInput())
		if err != test.want {
			t.Errorf("status %d = %v, want %v", test.status, err, test.want)
		}
	}
}

func TestModuregisAuthorizeInvalidInput(t *testing.T) {
	_, client := moduregisTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for invalid input")
	})
	request := validModuregisInput()
	request.BearerToken = ""
	if _, err := client.ModuregisAuthorize(context.Background(), "internal-token", request); err != ErrInvalidInput {
		t.Fatalf("missing bearer = %v, want ErrInvalidInput", err)
	}
	request = validModuregisInput()
	request.Action = "capability:publish"
	if _, err := client.ModuregisAuthorize(context.Background(), "internal-token", request); err != ErrInvalidInput {
		t.Fatalf("publish without artifact = %v, want ErrInvalidInput", err)
	}
	request = validModuregisInput()
	request.Action = "adapter:activate"
	if _, err := client.ModuregisAuthorize(context.Background(), "internal-token", request); err != ErrInvalidInput {
		t.Fatalf("activate without adapter id = %v, want ErrInvalidInput", err)
	}
	if _, err := client.ModuregisAuthorize(context.Background(), "", validModuregisInput()); err != ErrInvalidInput {
		t.Fatalf("empty internal token = %v, want ErrInvalidInput", err)
	}
}

func TestModuregisAuthorizeDoesNotSendNamespace(t *testing.T) {
	var got map[string]any
	_, client := moduregisTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body = %v", err)
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal body = %v", err)
		}
		response := map[string]any{
			"api_version": "aegivela.io/v1alpha1", "decision_id": "d-1", "outcome": "allow",
			"policy_version": "p-1", "tenant_id": "tenant-a", "actor_id": "actor-1",
		}
		_ = json.NewEncoder(w).Encode(response)
	})
	if _, err := client.ModuregisAuthorize(context.Background(), "internal-token", validModuregisInput()); err != nil {
		t.Fatalf("ModuregisAuthorize() = %v", err)
	}
	if _, present := got["namespace"]; present {
		t.Fatalf("request carried a top-level namespace field not defined by the v1alpha1 contract: %#v", got)
	}
}

func TestModuregisAuthorizeParsesAgentPrincipalEnvelope(t *testing.T) {
	_, client := moduregisTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"api_version":"aegivela.io/v1alpha1","decision_id":"dec-1","outcome":"allow","policy_version":"p-1","tenant_id":"tenant-a","actor_id":"agent-9","agent_id":"agent-9","master_id":"user-1","workload_id":"spiffe://aegivela.example/gateway","subject_ref":"https://issuer.example/agents/9","evidence_refs":["evidence://e/1"],"expires_at":"2026-08-19T00:00:00Z","grant_token":""}`)
	})
	decision, err := client.ModuregisAuthorize(context.Background(), "internal-token", ModuregisAuthorizeRequest{
		Action: "capability:read", ResourceKind: "capability", ResourceReference: "cap-1", BearerToken: "token",
	})
	if err != nil {
		t.Fatalf("ModuregisAuthorize() error = %v", err)
	}
	if decision.AgentID != "agent-9" || decision.MasterID != "user-1" ||
		decision.WorkloadID != "spiffe://aegivela.example/gateway" || decision.SubjectRef != "https://issuer.example/agents/9" {
		t.Fatalf("decision agent envelope = %#v", decision)
	}
}
