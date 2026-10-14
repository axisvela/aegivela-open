package grantsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIssueApprovalAndPreAuthorization(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/approvals/issue":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"artifact":"ap-1"}`))
		case "/v1/pre-authorizations/issue":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"artifact":"pa-1"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())

	ap, err := c.IssueApproval(context.Background(), "itok", ApprovalRequest{
		BearerToken: "b", Action: "tool:invoke", Resource: ResourceRef{Kind: "tool", ID: "t1"}, Scope: []string{"read"},
	})
	if err != nil || !strings.Contains(string(ap), "ap-1") {
		t.Fatalf("approval=%s err=%v", ap, err)
	}
	pa, err := c.IssuePreAuthorization(context.Background(), "itok", PreAuthorizationRequest{
		BearerToken: "b", Action: "tool:invoke", Resource: ResourceRef{Kind: "tool", ID: "t1"}, Scope: []string{"read"},
	})
	if err != nil || !strings.Contains(string(pa), "pa-1") {
		t.Fatalf("preauth=%s err=%v", pa, err)
	}
	_ = json.Marshal
}
