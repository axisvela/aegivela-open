package pepsdk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestModuregisAuthorizeSuccess(t *testing.T) {
	var gotReq ModuregisAuthorizeRequest
	var gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-AEGIVELA-PEP")
		if r.URL.Path != "/v1/moduregis/authorize" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(b, &gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ModuregisAuthorizeDecision{
			DecisionID:    "dec-1",
			Outcome:       "allow",
			PolicyVersion: "pv-1",
			TenantID:      "tenant-1",
			GrantToken:    "grant-1",
			ExpiresAt:     time.Now().Add(time.Hour),
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	req := ModuregisAuthorizeRequest{
		Action:            "capability.invoke",
		ResourceKind:      "capability",
		ResourceReference: "cap/order/create@v1",
		BearerToken:       "caller-token",
		Scope:             []string{"order:create"},
	}
	dec, err := c.ModuregisAuthorize(context.Background(), "pep-internal", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dec.DecisionID != "dec-1" || dec.Outcome != "allow" || dec.GrantToken != "grant-1" {
		t.Errorf("unexpected decision: %+v", dec)
	}
	if gotToken != "pep-internal" {
		t.Errorf("expected X-AEGIVELA-PEP=pep-internal, got %q", gotToken)
	}
	if gotReq.Action != "capability.invoke" || gotReq.TraceID == "" {
		t.Errorf("request not marshaled correctly: %+v", gotReq)
	}
}

func TestModuregisAuthorizeDenied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ModuregisAuthorizeDecision{Outcome: "deny"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	_, err := c.ModuregisAuthorize(context.Background(), "pep", ModuregisAuthorizeRequest{
		Action:            "x",
		ResourceKind:      "k",
		ResourceReference: "r",
	})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("expected ErrDenied, got %v", err)
	}
}

func TestModuregisAuthorizeStatusMapping(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusUnauthorized, ErrUnauthenticated},
		{http.StatusForbidden, ErrDenied},
		{http.StatusInternalServerError, ErrUnavailable},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
		}))
		c := NewClient(srv.URL, srv.Client())
		_, err := c.ModuregisAuthorize(context.Background(), "pep", ModuregisAuthorizeRequest{
			Action: "x", ResourceKind: "k", ResourceReference: "r",
		})
		srv.Close()
		if !errors.Is(err, tc.want) {
			t.Fatalf("status %d: expected %v, got %v", tc.status, tc.want, err)
		}
	}
}

func TestModuregisAuthorizeInvalidInput(t *testing.T) {
	c := NewClient("http://example.invalid", nil)
	if _, err := c.ModuregisAuthorize(context.Background(), "", ModuregisAuthorizeRequest{
		Action: "x", ResourceKind: "k", ResourceReference: "r",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty token, got %v", err)
	}
	if _, err := c.ModuregisAuthorize(context.Background(), "pep", ModuregisAuthorizeRequest{
		ResourceKind: "k", ResourceReference: "r",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty action, got %v", err)
	}
}
