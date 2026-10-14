package pepsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebResources(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/web-permissions:resolve":
			_, _ = w.Write([]byte(`{"api_version":"aegivela.io/v1alpha1","permissions":[]}`))
		case "/v1/web-resources/moduregis":
			_, _ = w.Write([]byte(`[{"resource_id":"console"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	perms, err := c.ResolveWebPermissions(context.Background(), "itok", "moduregis", "console.catalog")
	if err != nil || !strings.Contains(string(perms), "permissions") {
		t.Fatalf("ResolveWebPermissions raw=%s err=%v", perms, err)
	}
	res, err := c.ListWebResources(context.Background(), "itok", "moduregis", "")
	if err != nil || !strings.Contains(string(res), "console") {
		t.Fatalf("ListWebResources raw=%s err=%v", res, err)
	}
}
