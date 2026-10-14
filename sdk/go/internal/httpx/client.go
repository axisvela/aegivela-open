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

// Client is the shared HTTP transport.
type Client struct {
	BaseURL  string
	HTTP     *http.Client
	TenantID string
}

// New creates a transport client. A nil httpClient gets a 10s-timeout default.
func New(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{BaseURL: baseURL, HTTP: httpClient}
}

// WithTenant sets the tenant scope (used by the revocation-check contract).
func (c *Client) WithTenant(tenantID string) *Client {
	c.TenantID = tenantID
	return c
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
	return c.HTTP.Do(req)
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
