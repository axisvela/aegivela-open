package pepsdk

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// Request performs an HTTP request against the client's base URL. It is the
// reuse seam for the sibling domain packages (identitysdk, grantsdk,
// attestationsdk): they supply explicit headers (e.g. the internal token) and
// an optional JSON body; the response is returned unread.
func (c *Client) Request(ctx context.Context, method, path string, headers map[string]string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.httpClient.Do(req)
}

// BaseURL returns the client's base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// DecodeJSON decodes a size-bounded JSON response body. When strict, unknown
// fields are rejected.
func DecodeJSON(r io.Reader, target any, strict bool) error {
	return decodeBoundedResponse(r, target, strict)
}

// GenerateTraceID returns a fresh hexadecimal trace id.
func GenerateTraceID() string { return generateTraceID() }
