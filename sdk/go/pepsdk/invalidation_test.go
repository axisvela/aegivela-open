package pepsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPollInvalidations(t *testing.T) {
	var path, token, tenant, since string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, token = r.URL.Path, r.Header.Get("X-Aegivela-Internal-Token")
		tenant, since = r.URL.Query().Get("tenant"), r.URL.Query().Get("since")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v1alpha1","events":[{"eventId":"e1","type":"policy_revoked","tenantId":"tenant-a","policyId":"p1","version":"v2","createdAt":"2026-11-10T00:00:00Z"}],"checkedAt":"2026-11-10T00:00:01Z"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	response, err := client.PollInvalidations(context.Background(), "internal-token", "tenant-a", time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("PollInvalidations: %v", err)
	}
	if path != "/v1/gateway/invalidations" || token != "internal-token" || tenant != "tenant-a" || since == "" {
		t.Fatalf("request path=%q token=%q tenant=%q since=%q", path, token, tenant, since)
	}
	if len(response.Events) != 1 || response.Events[0].Type != "policy_revoked" || response.Events[0].EventID != "e1" {
		t.Fatalf("response = %#v", response)
	}
}

func TestPollInvalidationsRequiresInput(t *testing.T) {
	client := NewClient("https://example.invalid", http.DefaultClient)
	if _, err := client.PollInvalidations(context.Background(), "", "tenant-a", time.Time{}); err == nil {
		t.Fatal("missing internal token must be rejected")
	}
	if _, err := client.PollInvalidations(context.Background(), "token", "", time.Time{}); err == nil {
		t.Fatal("missing tenant must be rejected")
	}
}
