package pepsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CreateWebResourceRequest is the input to CreateWebResource.
type CreateWebResourceRequest struct {
	ProductID         string
	ResourceID        string
	ParentResourceID  string
	Kind              string
	DisplayName       string
	ProjectionActions json.RawMessage
}

type createWebResourceWire struct {
	APIVersion        string          `json:"api_version"`
	ProductID         string          `json:"product_id"`
	ResourceID        string          `json:"resource_id"`
	ParentResourceID  string          `json:"parent_resource_id"`
	Kind              string          `json:"kind"`
	DisplayName       string          `json:"display_name"`
	ProjectionActions json.RawMessage `json:"projection_actions"`
}

// CreateWebResource registers a web resource; the resource JSON is returned
// verbatim. The caller presents the AEGIVELA internal token.
func (c *Client) CreateWebResource(ctx context.Context, internalToken string, req CreateWebResourceRequest) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.ProductID) == "" ||
		strings.TrimSpace(req.ResourceID) == "" || strings.TrimSpace(req.DisplayName) == "" {
		return nil, ErrInvalidInput
	}
	body, err := json.Marshal(createWebResourceWire{
		APIVersion: APIVersionV1Alpha1, ProductID: req.ProductID, ResourceID: req.ResourceID,
		ParentResourceID: req.ParentResourceID, Kind: req.Kind, DisplayName: req.DisplayName,
		ProjectionActions: req.ProjectionActions,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return c.webJSON(ctx, http.MethodPost, "/v1/web-resources", internalToken, body)
}

// DeprecateWebResource deprecates a web resource
// (POST /v1/web-resources/{productID}/{resourceID}:deprecate); the resource JSON
// is returned verbatim.
func (c *Client) DeprecateWebResource(ctx context.Context, internalToken, productID, resourceID string) (json.RawMessage, error) {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(productID) == "" || strings.TrimSpace(resourceID) == "" {
		return nil, ErrInvalidInput
	}
	path := "/v1/web-resources/" + url.PathEscape(productID) + "/" + url.PathEscape(resourceID) + ":deprecate"
	return c.webJSON(ctx, http.MethodPost, path, internalToken, nil)
}

// CreateWebResourceRuleRequest is the input to CreateWebResourceRule.
type CreateWebResourceRuleRequest struct {
	ProductID       string
	RuleID          string
	ResourceID      string
	Descendants     bool
	SubjectSelector json.RawMessage
	Actions         json.RawMessage
	Effect          string
}

type createWebResourceRuleWire struct {
	APIVersion      string          `json:"api_version"`
	ProductID       string          `json:"product_id"`
	RuleID          string          `json:"rule_id"`
	ResourceID      string          `json:"resource_id"`
	Descendants     bool            `json:"descendants"`
	SubjectSelector json.RawMessage `json:"subject_selector"`
	Actions         json.RawMessage `json:"actions"`
	Effect          string          `json:"effect"`
}

// CreateWebResourceRule creates a web-resource policy rule
// (POST /v1/web-resource-policies).
func (c *Client) CreateWebResourceRule(ctx context.Context, internalToken string, req CreateWebResourceRuleRequest) error {
	if strings.TrimSpace(internalToken) == "" || strings.TrimSpace(req.ProductID) == "" ||
		strings.TrimSpace(req.RuleID) == "" || strings.TrimSpace(req.ResourceID) == "" ||
		len(req.SubjectSelector) == 0 || len(req.Actions) == 0 || strings.TrimSpace(req.Effect) == "" {
		return ErrInvalidInput
	}
	body, err := json.Marshal(createWebResourceRuleWire{
		APIVersion: APIVersionV1Alpha1, ProductID: req.ProductID, RuleID: req.RuleID, ResourceID: req.ResourceID,
		Descendants: req.Descendants, SubjectSelector: req.SubjectSelector, Actions: req.Actions, Effect: req.Effect,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	_, err = c.webJSON(ctx, http.MethodPost, "/v1/web-resource-policies", internalToken, body)
	return err
}
