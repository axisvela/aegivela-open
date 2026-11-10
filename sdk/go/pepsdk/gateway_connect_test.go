package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func gatewayConnectTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(handler))
	client := NewClient(server.URL, server.Client())
	t.Cleanup(server.Close)
	return server, client
}

func validGatewayConnectInput() GatewayConnectRequest {
	return GatewayConnectRequest{
		Action:            "connection:open",
		AgentID:           "agent-1",
		AgentClass:        "twin",
		WorkloadAssertion: "eyJhbGciOiJFZERTQSJ9.signature",
		TenantID:          "tenant-a",
		TargetHost:        "api.example.com",
		TargetScheme:      "https",
		TargetPath:        "/v1",
		RequestedScope:    []string{"data.read"},
		Audience:          "server-gateway",
	}
}

func TestGatewayConnectAllow(t *testing.T) {
	_, client := gatewayConnectTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Aegivela-Internal-Token") == "" {
			t.Error("missing internal token header")
		}
		response := map[string]any{
			"api_version": "aegivela.io/v1alpha1", "decision_id": "d-1", "outcome": "allow",
			"policy_version": "p-1", "expires_at": "2099-01-01T00:00:00Z", "ttl_seconds": 60,
			"offline_eligible": false, "evidence_refs": []string{"evidence://gateway/d-1"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})
	decision, err := client.GatewayConnect(context.Background(), "internal-token", validGatewayConnectInput())
	if err != nil {
		t.Fatalf("GatewayConnect() = %v", err)
	}
	if decision.DecisionID != "d-1" || decision.Outcome != "allow" || decision.PolicyVersion != "p-1" {
		t.Fatalf("decision = %+v", decision)
	}
	if len(decision.EvidenceRefs) != 1 || decision.EvidenceRefs[0] != "evidence://gateway/d-1" {
		t.Fatalf("evidence refs = %v", decision.EvidenceRefs)
	}
}

func TestGatewayConnectDenyMapsToErrDenied(t *testing.T) {
	_, client := gatewayConnectTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"api_version": "aegivela.io/v1alpha1", "decision_id": "d-1", "outcome": "deny",
			"policy_version": "p-1", "expires_at": "2099-01-01T00:00:00Z", "ttl_seconds": 60,
			"offline_eligible": false, "evidence_refs": []string{"evidence://gateway/d-1"},
		}
		_ = json.NewEncoder(w).Encode(response)
	})
	_, err := client.GatewayConnect(context.Background(), "internal-token", validGatewayConnectInput())
	if err != ErrDenied {
		t.Fatalf("GatewayConnect() deny = %v, want ErrDenied", err)
	}
}

func TestGatewayConnectStatusMapping(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusUnauthorized, ErrUnauthenticated},
		{http.StatusForbidden, ErrDenied},
		{http.StatusServiceUnavailable, ErrUnavailable},
		{http.StatusBadRequest, ErrInvalidInput},
	}
	for _, test := range cases {
		_, client := gatewayConnectTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(test.status)
		})
		_, err := client.GatewayConnect(context.Background(), "internal-token", validGatewayConnectInput())
		if err != test.want {
			t.Errorf("status %d = %v, want %v", test.status, err, test.want)
		}
	}
}

func TestGatewayConnectInvalidInput(t *testing.T) {
	_, client := gatewayConnectTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for invalid input")
	})
	request := validGatewayConnectInput()
	request.WorkloadAssertion = ""
	if _, err := client.GatewayConnect(context.Background(), "internal-token", request); err != ErrInvalidInput {
		t.Fatalf("missing assertion = %v, want ErrInvalidInput", err)
	}
	request = validGatewayConnectInput()
	request.TargetScheme = "ftp"
	if _, err := client.GatewayConnect(context.Background(), "internal-token", request); err != ErrInvalidInput {
		t.Fatalf("bad scheme = %v, want ErrInvalidInput", err)
	}
	if _, err := client.GatewayConnect(context.Background(), "", validGatewayConnectInput()); err != ErrInvalidInput {
		t.Fatalf("empty internal token = %v, want ErrInvalidInput", err)
	}
	request = validGatewayConnectInput()
	request.Action = "tool:invoke"
	if _, err := client.GatewayConnect(context.Background(), "internal-token", request); err != ErrInvalidInput {
		t.Fatalf("tool invoke without attribution = %v, want ErrInvalidInput", err)
	}
}

func TestGatewayConnectRejectsCrossModeSubstitution(t *testing.T) {
	_, client := gatewayConnectTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called for invalid input")
	})
	withAgent := func(mode Mode) GatewayConnectRequest {
		request := validGatewayConnectInput()
		request.AuthorizationMode = mode
		return request
	}
	cases := map[string]GatewayConnectRequest{
		"system_api with agent authority": withAgent(ModeSystemAPI),
		"twin mode with service class": func() GatewayConnectRequest {
			request := withAgent(ModeTwinAgentAPI)
			request.AgentClass = "service"
			return request
		}(),
		"agent id without class": func() GatewayConnectRequest {
			request := validGatewayConnectInput()
			request.AgentClass = ""
			return request
		}(),
		"agent class without id": func() GatewayConnectRequest {
			request := validGatewayConnectInput()
			request.AgentID = ""
			return request
		}(),
		"unknown mode": withAgent(ModeHumanWeb),
	}
	for name, request := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := client.GatewayConnect(context.Background(), "internal-token", request); err != ErrInvalidInput {
				t.Fatalf("err = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestGatewayConnectSystemAPISendsMode(t *testing.T) {
	var received map[string]any
	_, client := gatewayConnectTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		response := map[string]any{
			"api_version": "aegivela.io/v1alpha1", "decision_id": "d-1", "outcome": "allow",
			"policy_version": "p-1", "expires_at": "2099-01-01T00:00:00Z", "ttl_seconds": 60,
			"offline_eligible": false, "evidence_refs": []string{"evidence://gateway/d-1"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})
	request := validGatewayConnectInput()
	request.AuthorizationMode = ModeSystemAPI
	request.AgentID = ""
	request.AgentClass = ""
	if _, err := client.GatewayConnect(context.Background(), "internal-token", request); err != nil {
		t.Fatalf("GatewayConnect() = %v", err)
	}
	if received["authorization_mode"] != "system_api" {
		t.Fatalf("authorization_mode = %v, want system_api", received["authorization_mode"])
	}
}
