package identitysdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIssueWorkloadAssertion(t *testing.T) {
	var token string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = r.Header.Get("X-Aegivela-Internal-Token")
		if r.URL.Path != "/v1/workload-assertions/issue" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"workload_assertion":"jwt-1"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())
	tok, err := c.IssueWorkloadAssertion(context.Background(), "itok", WorkloadAssertionRequest{
		BindingID: "b", ClientID: "c", WorkloadRef: "w", Audience: "rheovela",
	})
	if err != nil || tok != "jwt-1" {
		t.Fatalf("tok=%q err=%v", tok, err)
	}
	if token != "itok" {
		t.Errorf("token header = %q", token)
	}
	if _, err := c.IssueWorkloadAssertion(context.Background(), "", WorkloadAssertionRequest{}); err == nil {
		t.Error("expected invalid input error")
	}
}

func TestCreateWorkloadBindingAndJWKS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"keys":[]}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"BindingID":"b1"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())
	raw, err := c.CreateWorkloadBinding(context.Background(), "itok", CreateWorkloadBindingRequest{
		SubjectRef: "s", OrganizationRoot: "o", ClientID: "c", WorkloadRef: "w",
	})
	if err != nil || !strings.Contains(string(raw), "b1") {
		t.Fatalf("raw=%s err=%v", raw, err)
	}
	jwks, err := c.WorkloadAssertionJWKS(context.Background())
	if err != nil || !strings.Contains(string(jwks), "keys") {
		t.Fatalf("jwks=%s err=%v", jwks, err)
	}
	_ = json.Marshal
}
