package pepsdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEvaluateRawOutcome(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/policy/decisions/evaluate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"api_version":"aegivela.io/v1alpha1","decision_id":"d1","outcome":"approval_required","policy_version":"v1","evidence_refs":["e"],"signed_decision":"jws"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	res, err := c.Evaluate(context.Background(),
		Route{Audience: "au", Action: "tool:invoke", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
		AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if res.Outcome != "approval_required" {
		t.Fatalf("Outcome = %q, want approval_required", res.Outcome)
	}
	if res.SignedDecision != "jws" {
		t.Fatalf("SignedDecision = %q", res.SignedDecision)
	}
}

func TestEvaluateRejectsUnsupportedVersionAndInvalidPrincipal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("invalid input reached server")
		_, _ = w.Write([]byte(`{"decision_id":"d1","outcome":"allow","policy_version":"v1"}`))
	}))
	defer srv.Close()
	for _, input := range []AuthorizationInput{
		{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: "unknown"},
		{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2, Principal: json.RawMessage(`[]`)},
		{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2, Principal: json.RawMessage(`null`)},
		{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2, Principal: json.RawMessage(`{`)},
	} {
		_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(),
			Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"}, input)
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("input %+v: error = %v, want ErrInvalidInput", input, err)
		}
	}
}

func TestEvaluateAndAuthorizeV2OutcomeSemantics(t *testing.T) {
	for _, outcome := range []string{"allow", "deny", "approvalRequired", "revoked"} {
		t.Run(outcome, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"apiVersion": APIVersionV2, "decisionId": "d1", "policyVersion": "v1",
					"outcome": outcome, "evidenceRefs": []string{"e"}, "signedDecision": "unverified-artifact",
				})
			}))
			defer srv.Close()
			c := NewClient(srv.URL, srv.Client())
			route := Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"}
			input := AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2}
			res, err := c.Evaluate(context.Background(), route, input)
			if err != nil || res.Outcome != outcome || res.SignedDecision != "unverified-artifact" {
				t.Fatalf("Evaluate = %+v, %v", res, err)
			}
			authorized, err := c.Authorize(context.Background(), route, input)
			if outcome == "allow" {
				if err != nil || authorized.DecisionID != res.DecisionID {
					t.Fatalf("Authorize = %+v, %v", authorized, err)
				}
			} else if authorized != nil || !errors.Is(err, ErrDenied) {
				t.Fatalf("non-allow Authorize = %+v, %v", authorized, err)
			}
		})
	}
}

func TestEvaluateRejectsDisallowedMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("disallowed mode reached server")
		_, _ = w.Write([]byte(`{"decision_id":"d1","outcome":"allow","policy_version":"v1"}`))
	}))
	defer srv.Close()
	_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(),
		Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1", Modes: []Mode{ModeSystemAPI}},
		AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestEvaluateRejectsMalformedV2Envelope(t *testing.T) {
	for _, body := range []string{
		`{"apiVersion":"wrong","decisionId":"d1","result":"allow","evidenceRefs":["e"]}`,
		`{"apiVersion":"aegivela.io/v2.0","result":"allow","evidenceRefs":["e"]}`,
		`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","result":"allow","outcome":"deny","evidenceRefs":["e"]}`,
		`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","outcome":"allow","evidenceRefs":["e"]}`,
		`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","result":"allow","evidenceRefs":["e"],"unexpected":true}`,
	} {
		t.Run(body, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(body))
			}))
			defer srv.Close()
			_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(),
				Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
				AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2})
			if err == nil {
				t.Fatal("malformed response accepted")
			}
		})
	}
}

func TestEvaluateForbiddenIsDenied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errorCode":"authorizationDenied"}`))
	}))
	defer srv.Close()
	for _, version := range []string{APIVersionV1Alpha1, APIVersionV2} {
		_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(),
			Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
			AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: version})
		if !errors.Is(err, ErrDenied) {
			t.Errorf("version %s: error = %v, want ErrDenied", version, err)
		}
	}
}

func TestEvaluateFailureClassification(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 429, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"errorCode":"unknown"}`))
			}))
			defer srv.Close()
			_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(),
				Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
				AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: APIVersionV2})
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != status {
				t.Fatalf("error = %v, want APIError status %d", err, status)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewClient("http://127.0.0.1:1", nil).Evaluate(ctx,
		Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
		AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt"})
	var apiErr *APIError
	if errors.As(err, &apiErr) || !errors.Is(err, context.Canceled) || !errors.Is(err, ErrUnavailable) {
		t.Fatalf("transport error lost classification/cause: %v", err)
	}
}

func TestEvaluateMalformedIsClassified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	for _, version := range []string{APIVersionV1Alpha1, APIVersionV2} {
		_, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(),
			Route{Audience: "au", Action: "read", Scope: []string{"read"}, Kind: "tool", Reference: "t1"},
			AuthorizationInput{Mode: ModeHumanWeb, BearerToken: "bt", APIVersion: version})
		if !errors.Is(err, ErrMalformed) {
			t.Errorf("version %s: error = %v, want ErrMalformed", version, err)
		}
	}
}
