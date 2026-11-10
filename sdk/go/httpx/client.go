// Package httpx is the module-internal HTTP transport shared by the AEGIVELA
// SDK packages (pepsdk, identitysdk, grantsdk, attestationsdk). It lives under
// internal/ so it is importable by the module's packages but not by external
// consumers — the transport is not part of the public SDK surface.
package httpx

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MaxResponseBytes bounds how much of a response body is read.
const MaxResponseBytes = 1 << 20

// Retry defaults: idempotent requests are retried on transport failures and 5xx
// responses, with a linear backoff starting at defaultBackoff.
const (
	defaultMaxAttempts = 3
	defaultBackoff     = 100 * time.Millisecond
)

// Client is the shared HTTP transport.
type Client struct {
	BaseURL  string
	HTTP     *http.Client
	TenantID string
	headers  http.Header
	// maxAttempts bounds retries for idempotent requests.
	maxAttempts int
	// backoff is the first retry delay; it grows linearly per attempt.
	backoff time.Duration
}

// New creates a transport client. A nil httpClient gets a 10s-timeout default.
func New(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{BaseURL: baseURL, HTTP: httpClient, maxAttempts: defaultMaxAttempts, backoff: defaultBackoff}
}

// SetRetry bounds the attempts for idempotent requests (transport failures and
// 5xx). attempts < 1 disables retries. Non-idempotent methods are never retried.
func (c *Client) SetRetry(attempts int) {
	c.maxAttempts = attempts
}

// WithTenant sets the tenant scope (used by the revocation-check contract).
func (c *Client) WithTenant(tenantID string) *Client {
	c.TenantID = tenantID
	return c
}

// WithHeaders copies construction-time defaults. Explicit endpoint headers win.
// Configured headers must never be forwarded to a different origin by redirects.
func (c *Client) WithHeaders(headers map[string]string) {
	c.headers = make(http.Header, len(headers))
	for key, value := range headers {
		c.headers.Set(key, value)
	}
	if len(c.headers) == 0 {
		return
	}
	hc := *c.HTTP
	checkRedirect := hc.CheckRedirect
	hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > 0 && (req.URL.Scheme != via[0].URL.Scheme || req.URL.Host != via[0].URL.Host) {
			return fmt.Errorf("refusing redirect of configured headers to another origin")
		}
		if checkRedirect != nil {
			return checkRedirect(req, via)
		}
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		return nil
	}
	c.HTTP = &hc
}

// Do applies defaults without modifying the caller's request and retries
// idempotent requests on transport failures and 5xx responses. A fresh attempt
// replays the request body via GetBody, so the caller's request is untouched.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	attempts := c.maxAttempts
	if attempts < 1 || !retryable(req.Method) {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 && !sleep(req.Context(), time.Duration(i)*c.backoff) {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, req.Context().Err()
		}
		resp, err := c.attempt(req, i > 0)
		if err != nil {
			lastErr = err
			if i == attempts-1 {
				return nil, err
			}
			continue
		}
		if resp.StatusCode >= 500 && i < attempts-1 {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("httpx: status %d from %s", resp.StatusCode, req.URL)
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

// attempt performs a single request, applying default headers and, on a retry,
// a fresh body from GetBody so the original request is never reused.
func (c *Client) attempt(req *http.Request, replay bool) (*http.Response, error) {
	if len(c.headers) == 0 && !replay {
		return c.HTTP.Do(req)
	}
	clone := req.Clone(req.Context())
	for key, values := range c.headers {
		if _, exists := clone.Header[key]; !exists {
			clone.Header[key] = append([]string(nil), values...)
		}
	}
	if replay && req.Body != nil {
		if req.GetBody == nil {
			return nil, fmt.Errorf("httpx: cannot replay request body for retry")
		}
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		clone.Body = body
	}
	return c.HTTP.Do(clone)
}

// retryable reports whether a method is safe to re-send after a transient
// failure.
func retryable(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	}
	return false
}

// sleep waits for d, returning false when the context is cancelled first. A
// non-positive d returns immediately.
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// Request performs an HTTP request against BaseURL. When body is non-nil it is
// sent as JSON (Content-Type set). headers are applied to the request.
func (c *Client) Request(ctx context.Context, method, path string, headers map[string]string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.Do(req)
}

// DecodeBounded reads a size-bounded JSON body into target. When strict,
// unknown fields are rejected.
func DecodeBounded(body io.Reader, target any, strict bool) error {
	encoded, err := io.ReadAll(io.LimitReader(body, MaxResponseBytes+1))
	if err != nil {
		return err
	}
	if len(encoded) > MaxResponseBytes {
		return fmt.Errorf("response exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	if strict {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("trailing response data")
	}
	return nil
}

// DecodeStrict is DecodeBounded with strict field checking.
func DecodeStrict(body io.Reader, target any) error {
	return DecodeBounded(body, target, true)
}

// GenerateTraceID returns a fresh hexadecimal trace id.
func GenerateTraceID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "fallback-trace-id"
	}
	return hex.EncodeToString(buf[:])
}
