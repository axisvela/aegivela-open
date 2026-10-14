package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestIssueWorkloadAssertion(t *testing.T) {
	var mu sync.Mutex
	var gotPath, gotToken string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		gotPath = r.URL.Path
		gotToken = r.Header.Get("X-Aegivela-Internal-Token")
		gotBody = body
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"workload_assertion": "assertion-jwt"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	tok, err := c.IssueWorkloadAssertion(context.Background(), "internal-tok", WorkloadAssertionRequest{
		SubjectRef: "axond", OrganizationRoot: "org-1", BindingID: "b-1",
		ClientID: "axond", WorkloadRef: "axond-worker", Audience: "rheovela",
	})
	if err != nil {
		t.Fatalf("IssueWorkloadAssertion: %v", err)
	}
	if tok != "assertion-jwt" {
		t.Errorf("token = %q", tok)
	}
	mu.Lock()
	defer mu.Unlock()
	if gotPath != "/v1/workload-assertions/issue" {
		t.Errorf("path = %q", gotPath)
	}
	if gotToken != "internal-tok" {
		t.Errorf("internal token header = %q", gotToken)
	}
	if gotBody["api_version"] != APIVersionV1Alpha1 || gotBody["audience"] != "rheovela" {
		t.Errorf("body = %v", gotBody)
	}
}

func TestIssueWorkloadAssertion_Errors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	_, err := c.IssueWorkloadAssertion(context.Background(), "tok", WorkloadAssertionRequest{
		BindingID: "b", ClientID: "c", WorkloadRef: "w", Audience: "rheovela",
	})
	if err != ErrUnauthenticated {
		t.Fatalf("err = %v, want ErrUnauthenticated", err)
	}
}

func TestIssueWorkloadAssertion_InvalidInput(t *testing.T) {
	c := NewClient("http://127.0.0.1:1", http.DefaultClient)
	if _, err := c.IssueWorkloadAssertion(context.Background(), "", WorkloadAssertionRequest{Audience: "x"}); err != ErrInvalidInput {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}
