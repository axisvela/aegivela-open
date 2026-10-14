package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateAndGetPolicy(t *testing.T) {
	var createBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/policies":
			_ = json.NewDecoder(r.Body).Decode(&createBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"policy_id":"p1","version":"v1","tenant_id":"t1","rules":[],"activated_at":"2026-01-01T00:00:00Z","created_at":"2026-01-01T00:00:00Z"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/policies/p1":
			_, _ = w.Write([]byte(`{"policy_id":"p1","version":"v1","tenant_id":"t1","rules":[{"id":"r"}],"activated_at":"2026-01-01T00:00:00Z","created_at":"2026-01-01T00:00:00Z"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())

	created, err := c.CreatePolicy(context.Background(), "itok", CreatePolicyRequest{
		PolicyID: "p1", Version: "v1", Rules: json.RawMessage(`[{"id":"r"}]`), ActivatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}
	if created.PolicyID != "p1" {
		t.Errorf("created = %+v", created)
	}
	if createBody["policy_id"] != "p1" {
		t.Errorf("body = %v", createBody)
	}
	got, err := c.GetPolicy(context.Background(), "itok", "p1")
	if err != nil {
		t.Fatalf("GetPolicy: %v", err)
	}
	if string(got.Rules) != `[{"id":"r"}]` {
		t.Errorf("rules = %s", got.Rules)
	}
}
