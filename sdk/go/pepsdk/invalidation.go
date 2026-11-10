package pepsdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/axisvela/aegivela-open/sdk/go/v3/httpx"
)

// InvalidationEvent is one tenant-scoped authority invalidation event.
type InvalidationEvent struct {
	EventID    string
	Type       string
	TenantID   string
	ResourceID string
	PolicyID   string
	Version    string
	CreatedAt  time.Time
}

// InvalidationResponse is the result of a tenant-scoped invalidation poll.
type InvalidationResponse struct {
	Events    []InvalidationEvent
	CheckedAt time.Time
}

// PollInvalidations polls authority invalidations for one tenant since a time.
func (c *Client) PollInvalidations(ctx context.Context, internalToken, tenantID string, since time.Time) (*InvalidationResponse, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(tenantID) == "" {
		return nil, ErrInvalidInput
	}
	query := url.Values{"tenant": {tenantID}}
	if !since.IsZero() {
		query.Set("since", since.UTC().Format(time.RFC3339))
	}
	headers := map[string]string{"X-Aegivela-Internal-Token": internalToken}
	resp, err := c.hc.Request(ctx, http.MethodGet, "/v1/gateway/invalidations?"+query.Encode(), headers, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, ErrUnauthenticated
	case http.StatusBadRequest:
		return nil, ErrInvalidInput
	default:
		return nil, ErrUnavailable
	}
	var wire struct {
		APIVersion string `json:"apiVersion"`
		Events     []struct {
			EventID    string    `json:"eventId"`
			Type       string    `json:"type"`
			TenantID   string    `json:"tenantId"`
			ResourceID string    `json:"resourceId"`
			PolicyID   string    `json:"policyId"`
			Version    string    `json:"version"`
			CreatedAt  time.Time `json:"createdAt"`
		} `json:"events"`
		CheckedAt time.Time `json:"checkedAt"`
	}
	if err := httpx.DecodeBounded(resp.Body, &wire, true); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	response := &InvalidationResponse{CheckedAt: wire.CheckedAt}
	for _, event := range wire.Events {
		response.Events = append(response.Events, InvalidationEvent{
			EventID: event.EventID, Type: event.Type, TenantID: event.TenantID,
			ResourceID: event.ResourceID, PolicyID: event.PolicyID, Version: event.Version,
			CreatedAt: event.CreatedAt,
		})
	}
	return response, nil
}
