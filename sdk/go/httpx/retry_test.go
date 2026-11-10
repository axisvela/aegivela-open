package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newRetryClient(t *testing.T, baseURL string, attempts int) *Client {
	t.Helper()
	c := New(baseURL, nil)
	c.SetRetry(attempts)
	c.backoff = time.Millisecond
	return c
}

func TestDoRetriesIdempotent5xxThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newRetryClient(t, srv.URL, 3)
	resp, err := c.Request(context.Background(), http.MethodGet, "/jwks", nil, nil)
	if err != nil {
		t.Fatalf("Request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestDoDoesNotRetryDefinitive4xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := newRetryClient(t, srv.URL, 3)
	resp, err := c.Request(context.Background(), http.MethodGet, "/jwks", nil, nil)
	if err != nil {
		t.Fatalf("Request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls = %d, want 1 (a definitive 4xx is not retried)", got)
	}
}

func TestDoDoesNotRetryNonIdempotent(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := newRetryClient(t, srv.URL, 3)
	resp, err := c.Request(context.Background(), http.MethodPost, "/v1/grants/issue", nil, []byte(`{}`))
	if err != nil {
		t.Fatalf("Request error: %v", err)
	}
	defer resp.Body.Close()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls = %d, want 1 (POST is not retried)", got)
	}
}

func TestDoRetriesTransportError(t *testing.T) {
	var calls int32
	c := New("http://unused.invalid", &http.Client{
		Transport: transportFunc(func(*http.Request) (*http.Response, error) {
			atomic.AddInt32(&calls, 1)
			return nil, errors.New("connection refused")
		}),
	})
	c.SetRetry(3)
	c.backoff = time.Millisecond

	if _, err := c.Request(context.Background(), http.MethodGet, "/jwks", nil, nil); err == nil {
		t.Fatal("expected a transport error")
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestWithRetryDisabledDoesNotRetry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := newRetryClient(t, srv.URL, 1)
	resp, err := c.Request(context.Background(), http.MethodGet, "/jwks", nil, nil)
	if err != nil {
		t.Fatalf("Request error: %v", err)
	}
	defer resp.Body.Close()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls = %d, want 1 (retry disabled)", got)
	}
}

func TestDoReplaysRequestBodyOnIdempotentRetry(t *testing.T) {
	var calls int32
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if atomic.AddInt32(&calls, 1) < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newRetryClient(t, srv.URL, 3)
	resp, err := c.Request(context.Background(), http.MethodPut, "/v1/x", nil, []byte(`{"body":"payload"}`))
	if err != nil {
		t.Fatalf("Request error: %v", err)
	}
	defer resp.Body.Close()
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
	for i, b := range bodies {
		if !strings.Contains(b, "payload") {
			t.Fatalf("attempt %d body = %q, want the replayed payload", i, b)
		}
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
