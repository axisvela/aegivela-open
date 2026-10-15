package pepsdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEvaluateV2AgentCarriesBoundAuthority(t *testing.T) {
	for _, mode := range []Mode{ModeServiceAgentAPI, ModeTwinAgentAPI} {
		t.Run(string(mode), func(t *testing.T) {
			input := v2AgentInput(mode)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var wire struct {
					Mode      string            `json:"authorizationMode"`
					Principal PolicyPrincipalV2 `json:"principal"`
					Parent    ParentAuthority   `json:"parentAuthority"`
					ExpiresAt time.Time         `json:"requestedExpiresAt"`
					TraceID   string            `json:"traceId"`
				}
				if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
					t.Error(err)
				}
				if wire.Mode != modeV2(mode) || wire.Principal.InstanceID != "instance-1" ||
					wire.Principal.AgentEpoch == nil || *wire.Principal.AgentEpoch != 2 ||
					wire.Principal.IdentityEpoch == nil || *wire.Principal.IdentityEpoch != 3 ||
					wire.Parent.AuthorityRef != "parent-1" || !wire.ExpiresAt.Equal(input.RequestedExpiresAt) || wire.TraceID != "trace-1" {
					t.Errorf("wire = %+v", wire)
				}
				_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","outcome":"allow","policyVersion":"v1","evidenceRefs":["e"]}`))
			}))
			defer srv.Close()
			if _, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(), v2AgentRoute(), input); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEvaluateV2AgentRejectsMissingOrExpandedAuthority(t *testing.T) {
	for name, mutate := range map[string]func(*AuthorizationInput){
		"principal":          func(in *AuthorizationInput) { in.Principal = nil },
		"agent":              func(in *AuthorizationInput) { in.Principal.AgentID = "" },
		"namespace":          func(in *AuthorizationInput) { in.Principal.Namespace = "" },
		"actor":              func(in *AuthorizationInput) { in.Principal.ActorRef = "" },
		"class":              func(in *AuthorizationInput) { in.Principal.AgentClass = "twin" },
		"root":               func(in *AuthorizationInput) { in.Principal.OrganizationAuthorityRoot = "" },
		"master":             func(in *AuthorizationInput) { in.Principal.MasterID = "master" },
		"workload":           func(in *AuthorizationInput) { in.Principal.WorkloadRef = "" },
		"binding":            func(in *AuthorizationInput) { in.Principal.AuthorityBinding = "humanMaster" },
		"agent epoch":        func(in *AuthorizationInput) { in.Principal.AgentEpoch = nil },
		"identity epoch":     func(in *AuthorizationInput) { zero := int64(0); in.Principal.IdentityEpoch = &zero },
		"assertion":          func(in *AuthorizationInput) { in.WorkloadAssertion = "" },
		"bearer":             func(in *AuthorizationInput) { in.BearerToken = "human" },
		"parent":             func(in *AuthorizationInput) { in.ParentAuthority = nil },
		"parent ref":         func(in *AuthorizationInput) { in.ParentAuthority.AuthorityRef = "" },
		"parent scope":       func(in *AuthorizationInput) { in.ParentAuthority.Scope = []string{"write"} },
		"duplicate scope":    func(in *AuthorizationInput) { in.ParentAuthority.Scope = []string{"read", "read"} },
		"empty scope":        func(in *AuthorizationInput) { in.ParentAuthority.Scope = nil },
		"parent audience":    func(in *AuthorizationInput) { in.ParentAuthority.Audience = []string{"other"} },
		"duplicate audience": func(in *AuthorizationInput) { in.ParentAuthority.Audience = []string{"axon", "axon"} },
		"parent expiry":      func(in *AuthorizationInput) { in.ParentAuthority.ExpiresAt = time.Now().Add(-time.Minute) },
		"missing expiry":     func(in *AuthorizationInput) { in.ParentAuthority.ExpiresAt = time.Time{} },
		"expanded expiry":    func(in *AuthorizationInput) { in.RequestedExpiresAt = in.ParentAuthority.ExpiresAt.Add(time.Second) },
		"task":               func(in *AuthorizationInput) { in.TaskBinding = "other" },
		"missing task":       func(in *AuthorizationInput) { in.ParentAuthority.TaskBinding = "" },
		"parent grant":       func(in *AuthorizationInput) { in.ParentExecutionGrant = "grant" },
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("invalid authority reached server")
			}))
			defer srv.Close()
			input := v2AgentInput(ModeServiceAgentAPI)
			mutate(&input)
			_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(), v2AgentRoute(), input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestEvaluateV2ProhibitsRawParentOutsideAgentModes(t *testing.T) {
	for _, mode := range []Mode{ModeHumanWeb, ModeSystemAPI, ModeDelegatedAPI} {
		t.Run(string(mode), func(t *testing.T) {
			input := v2AgentInput(ModeServiceAgentAPI)
			input.Mode = mode
			if mode == ModeHumanWeb {
				input.BearerToken, input.WorkloadAssertion = "user", ""
			}
			if mode == ModeDelegatedAPI {
				input.BearerToken, input.ParentExecutionGrant = "user", "signed-parent-grant"
			}
			if !errors.Is(input.Validate(), ErrInvalidInput) {
				t.Fatal("caller-supplied parent authority accepted outside agent modes")
			}
		})
	}
}

func TestEvaluateV2AgentDefaultExpiryIsBoundedByParent(t *testing.T) {
	input := v2AgentInput(ModeServiceAgentAPI)
	input.RequestedExpiresAt = time.Time{}
	input.ParentAuthority.ExpiresAt = time.Now().Add(time.Minute).UTC()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wire struct {
			ExpiresAt time.Time `json:"requestedExpiresAt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
			t.Error(err)
		}
		if !wire.ExpiresAt.Equal(input.ParentAuthority.ExpiresAt) {
			t.Errorf("expiry %s, parent %s", wire.ExpiresAt, input.ParentAuthority.ExpiresAt)
		}
		_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","outcome":"allow","policyVersion":"v1","evidenceRefs":["e"]}`))
	}))
	defer srv.Close()
	if _, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(), v2AgentRoute(), input); err != nil {
		t.Fatal(err)
	}
}

func v2AgentRoute() Route {
	return Route{Audience: "axon", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"}
}

func v2AgentInput(mode Mode) AuthorizationInput {
	agentEpoch, identityEpoch := int64(2), int64(3)
	principal := &PolicyPrincipalV2{
		Namespace: "ns", ActorRef: "agent-1", AgentID: "agent-1", AgentClass: "service",
		WorkloadRef: "workload-1", OrganizationAuthorityRoot: "org-1", AuthorityBinding: "organizationRoot",
		AgentEpoch: &agentEpoch, IdentityEpoch: &identityEpoch, InstanceID: "instance-1",
	}
	if mode == ModeTwinAgentAPI {
		principal.AgentClass, principal.MasterID, principal.AuthorityBinding = "twin", "master-1", "humanMaster"
		principal.OrganizationAuthorityRoot = ""
	}
	return AuthorizationInput{
		APIVersion: APIVersionV2, Mode: mode, Principal: principal, WorkloadAssertion: "assertion",
		TaskBinding: "task-1", TraceID: "trace-1", RequestedExpiresAt: time.Now().Add(30 * time.Second).UTC(),
		ParentAuthority: &ParentAuthority{AuthorityRef: "parent-1", Scope: []string{"read"}, Audience: []string{"axon"}, ExpiresAt: time.Now().Add(time.Minute).UTC(), TaskBinding: "task-1"},
	}
}
