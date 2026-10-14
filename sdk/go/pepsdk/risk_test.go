package pepsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubmitRiskSignal(t *testing.T) {
	var path, token string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, token = r.URL.Path, r.Header.Get("X-Aegivela-Internal-Token")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()
	c := NewClient(srv.URL, srv.Client())
	if err := c.SubmitRiskSignal(context.Background(), "itok", RiskSignalRequest{SubjectRef: "s", RiskRef: "r"}); err != nil {
		t.Fatalf("SubmitRiskSignal: %v", err)
	}
	if path != "/v1/risk/signals" || token != "itok" {
		t.Errorf("path=%q token=%q", path, token)
	}
}

func TestSubmitRiskSignal_InvalidInput(t *testing.T) {
	c := NewClient("http://127.0.0.1:1", http.DefaultClient)
	if err := c.SubmitRiskSignal(context.Background(), "", RiskSignalRequest{}); err != ErrInvalidInput {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}
