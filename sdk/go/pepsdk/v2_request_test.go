package pepsdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestAuthorizeV2UsesCamelCaseWireFields(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","result":"allow","evidenceRefs":["e"]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	if _, err := c.Authorize(context.Background(),
		Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "dataset", Reference: "d1"},
		AuthorizationInput{
			Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2,
			Principal:   &PolicyPrincipalV2{Namespace: "ns", ActorRef: "ar", SubjectRef: "u1"},
			TaskBinding: "task-1", TraceID: "trace-1",
		}); err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	for _, key := range []string{"apiVersion", "authorizationMode", "requestedScope", "requestedExpiresAt", "taskBinding", "traceId"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("camelCase field %q missing from %v", key, body)
		}
	}
	for _, key := range []string{"api_version", "authorization_mode", "requested_scope", "requested_expires_at", "task_binding", "trace_id"} {
		if _, ok := body[key]; ok {
			t.Fatalf("legacy snake_case field %q present in %v", key, body)
		}
	}
	principal, ok := body["principal"].(map[string]any)
	if !ok || principal["subjectRef"] != "u1" {
		t.Fatalf("principal = %v", body["principal"])
	}
	if _, ok := principal["subject_ref"]; ok {
		t.Fatal("legacy principal.subject_ref present")
	}
}

func TestEvaluateRejectsOversizeResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","result":"allow","evidenceRefs":["` + strings.Repeat("x", (1<<20)+16) + `"]}`))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(),
		Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "dataset", Reference: "d1"},
		AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2,
			Principal: &PolicyPrincipalV2{Namespace: "ns", ActorRef: "ar"}})
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("oversize response error = %v, want ErrMalformed", err)
	}
}

func TestEvaluateRejectsTrailingJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","result":"allow","evidenceRefs":["e"]}{"extra":true}`))
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(),
		Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "dataset", Reference: "d1"},
		AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2,
			Principal: &PolicyPrincipalV2{Namespace: "ns", ActorRef: "ar"}})
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("trailing JSON error = %v, want ErrMalformed", err)
	}
}
