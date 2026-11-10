package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRevocationCheckClear(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-AEGIVELA-PEP") != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"api_version": "aegivela.io/v1alpha1",
			"class":       "pre_dispatch",
			"outcome":     "clear",
			"checked_at":  "2026-07-25T00:00:00Z",
			"trace_id":    "trace-1",
		})
	}))
	defer server.Close()
	client := NewClient(server.URL, server.Client()).WithTenant("tenant-a")
	err := client.CheckRevocation(context.Background(), "test-token", ClassPreDispatch, []RevocationSelector{{Kind: "grant_jti", Value: "123e4567-e89b-42d3-a456-426614174000"}}, nil)
	if err != nil {
		t.Fatalf("CheckRevocation() = %v", err)
	}
}

func TestRevocationCheckRevokedReturnsDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"api_version": "aegivela.io/v1alpha1",
			"class":       "continuation",
			"outcome":     "revoked",
			"checked_at":  "2026-07-25T00:00:05Z",
			"trace_id":    "trace-2",
		})
	}))
	defer server.Close()
	client := NewClient(server.URL, server.Client()).WithTenant("tenant-a")
	err := client.CheckRevocation(context.Background(), "tok", ClassContinuation, []RevocationSelector{{Kind: "subject_ref", Value: "user-1"}}, nil)
	if err != ErrDenied {
		t.Fatalf("CheckRevocation() = %v, want ErrDenied", err)
	}
}

func TestRevocationCheckUnavailableFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client := NewClient(server.URL, server.Client()).WithTenant("tenant-a")
	err := client.CheckRevocation(context.Background(), "tok", ClassPreDispatch, []RevocationSelector{{Kind: "grant_jti", Value: "jti"}}, nil)
	if err != ErrUnavailable {
		t.Fatalf("CheckRevocation() = %v, want ErrUnavailable", err)
	}
}

func TestRevocationCheckInvalidClass(t *testing.T) {
	client := NewClient("http://localhost", nil)
	err := client.CheckRevocation(context.Background(), "tok", "background", []RevocationSelector{{Kind: "grant_jti", Value: "jti"}}, nil)
	if err != ErrInvalidInput {
		t.Fatalf("CheckRevocation() = %v, want ErrInvalidInput", err)
	}
}
