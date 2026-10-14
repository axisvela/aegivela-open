package pepsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type webPermissionResolutionWire struct {
	APIVersion     string `json:"api_version"`
	ProductID      string `json:"product_id"`
	RootResourceID string `json:"root_resource_id,omitempty"`
}

// ResolveWebPermissions resolves the permission set for the caller's web
// principal over a product's root resource (POST /v1/web-permissions:resolve).
// The permission set is returned verbatim. The caller presents the AEGIVELA
// internal token.
func (c *Client) ResolveWebPermissions(ctx context.Context, internalToken, productID, rootResourceID string) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(productID) == "" {
		return nil, ErrInvalidInput
	}
	body, err := json.Marshal(webPermissionResolutionWire{
		APIVersion: APIVersionV1Alpha1, ProductID: productID, RootResourceID: rootResourceID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return c.webJSON(ctx, http.MethodPost, "/v1/web-permissions:resolve", internalToken, body)
}

// ListWebResources lists active web resources for a product, optionally under a
// parent (GET /v1/web-resources/{productID}). Resources are returned verbatim.
func (c *Client) ListWebResources(ctx context.Context, internalToken, productID, parentResourceID string) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(productID) == "" {
		return nil, ErrInvalidInput
	}
	path := "/v1/web-resources/" + url.PathEscape(productID)
	if parentResourceID != "" {
		path += "?parent_resource_id=" + url.QueryEscape(parentResourceID)
	}
	return c.webJSON(ctx, http.MethodGet, path, internalToken, nil)
}

// GetWebResource fetches one active web resource
// (GET /v1/web-resources/{productID}/{resourceID}). Returned verbatim.
func (c *Client) GetWebResource(ctx context.Context, internalToken, productID, resourceID string) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(productID) == "" || strings.TrimSpace(resourceID) == "" {
		return nil, ErrInvalidInput
	}
	path := "/v1/web-resources/" + url.PathEscape(productID) + "/" + url.PathEscape(resourceID)
	return c.webJSON(ctx, http.MethodGet, path, internalToken, nil)
}

func (c *Client) webJSON(ctx context.Context, method, path, internalToken string, body []byte) (json.RawMessage, error) {
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	var httpReq *http.Request
	var err error
	if reader != nil {
		httpReq, err = http.NewRequestWithContext(ctx, method, c.hc.BaseURL+path, reader)
	} else {
		httpReq, err = http.NewRequestWithContext(ctx, method, c.hc.BaseURL+path, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("X-Aegivela-Internal-Token", internalToken)
	resp, err := c.hc.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthenticated
	case http.StatusForbidden, http.StatusNotFound:
		return nil, ErrDenied
	case http.StatusBadRequest:
		return nil, ErrInvalidInput
	case http.StatusServiceUnavailable:
		return nil, ErrUnavailable
	}
	if resp.StatusCode >= 300 {
		return nil, ErrUnavailable
	}
	var raw json.RawMessage
	if err := decodeBoundedResponse(resp.Body, &raw, false); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return raw, nil
}
