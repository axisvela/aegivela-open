// Package aegivela is the Apache-2.0 Go client SDK for AEGIVELA. It is the
// neutral entry point: it re-exports the pepsdk package (the implementation
// the client surface). Consumers may import either this package or pepsdk
// directly; this package exists so the SDK has a product-named root, mirroring
// the other *-open SDKs.
//
// Coverage: policy enforcement (Authorize / GatewayConnect / ModuregisAuthorize
// / CheckRevocation + middleware), workload identity (bindings + assertions +
// JWKS), execution grants (issue / verify / JWKS), and attestations
// (issue / verify / JWKS).
package aegivela

import (
	"github.com/axisrobo/aegivela-open/sdk/go/attestationsdk"
	"github.com/axisrobo/aegivela-open/sdk/go/governancesdk"
	"github.com/axisrobo/aegivela-open/sdk/go/grantsdk"
	"github.com/axisrobo/aegivela-open/sdk/go/identitysdk"
	"github.com/axisrobo/aegivela-open/sdk/go/pepsdk"
)

// Core client.
type (
	Client = pepsdk.Client
	Mode   = pepsdk.Mode
)

// Error sentinels.
var (
	ErrInvalidRoute    = pepsdk.ErrInvalidRoute
	ErrInvalidInput    = pepsdk.ErrInvalidInput
	ErrUnauthenticated = pepsdk.ErrUnauthenticated
	ErrDenied          = pepsdk.ErrDenied
	ErrUnavailable     = pepsdk.ErrUnavailable
)

// API versions and modes.
const (
	APIVersionV1Alpha1 = pepsdk.APIVersionV1Alpha1
	APIVersionV2       = pepsdk.APIVersionV2

	ModeHumanWeb        = pepsdk.ModeHumanWeb
	ModeSystemAPI       = pepsdk.ModeSystemAPI
	ModeDelegatedAPI    = pepsdk.ModeDelegatedAPI
	ModeServiceAgentAPI = pepsdk.ModeServiceAgentAPI
	ModeTwinAgentAPI    = pepsdk.ModeTwinAgentAPI
)

// Constructors.
var (
	NewClient                     = pepsdk.NewClient
	NewAuthorizeMiddleware        = pepsdk.NewAuthorizeMiddleware
	WithEvidenceRecorder          = pepsdk.WithEvidenceRecorder
	WithMode                      = pepsdk.WithMode
	WithRoute                     = pepsdk.WithRoute
	ResultFromContext             = pepsdk.ResultFromContext
	AuthorizationInputFromContext = pepsdk.AuthorizationInputFromContext
	EvidenceFromContext           = pepsdk.EvidenceFromContext
)

// Enforcement types.
type (
	Route                      = pepsdk.Route
	Resource                   = pepsdk.Resource
	AuthorizationInput         = pepsdk.AuthorizationInput
	Principal                  = pepsdk.Principal
	Result                     = pepsdk.Result
	Authorizer                 = pepsdk.Authorizer
	Evidence                   = pepsdk.Evidence
	EvidenceRecorder           = pepsdk.EvidenceRecorder
	Middleware                 = pepsdk.Middleware
	MiddlewareOption           = pepsdk.MiddlewareOption
	GatewayConnectRequest      = pepsdk.GatewayConnectRequest
	GatewayConnectDecision     = pepsdk.GatewayConnectDecision
	ToolAttribution            = pepsdk.ToolAttribution
	ModuregisAuthorizeRequest  = pepsdk.ModuregisAuthorizeRequest
	ModuregisAuthorizeDecision = pepsdk.ModuregisAuthorizeDecision
)

// Revocation.
type (
	SLOClass           = pepsdk.SLOClass
	RevocationSelector = pepsdk.RevocationSelector
)

const (
	ClassPreDispatch  = pepsdk.ClassPreDispatch
	ClassContinuation = pepsdk.ClassContinuation
	ClassConnection   = pepsdk.ClassConnection
)

// Workload identity.
type (
	IdentityClient               = identitysdk.Client
	WorkloadAssertionRequest     = identitysdk.WorkloadAssertionRequest
	CreateWorkloadBindingRequest = identitysdk.CreateWorkloadBindingRequest
)

// Execution grants.
type (
	GrantClient             = grantsdk.Client
	GrantIssueRequest       = grantsdk.GrantIssueRequest
	GrantVerifyRequest      = grantsdk.GrantVerifyRequest
	VerifyGrantResult       = grantsdk.VerifyGrantResult
	ResourceRef             = grantsdk.ResourceRef
	ApprovalRequest         = grantsdk.ApprovalRequest
	PreAuthorizationRequest = grantsdk.PreAuthorizationRequest
	TokenExchangeRequest    = grantsdk.TokenExchangeRequest
)

// Attestations.
type (
	AttestationClient        = attestationsdk.Client
	AttestationRequest       = attestationsdk.AttestationRequest
	AttestationResult        = attestationsdk.AttestationResult
	AttestationVerifyRequest = attestationsdk.AttestationVerifyRequest
)

// Tool invocation + decision verification.
type (
	ToolInvokeRequest     = pepsdk.ToolInvokeRequest
	ToolInvokeResult      = pepsdk.ToolInvokeResult
	DecisionVerifyRequest = pepsdk.DecisionVerifyRequest
	VerifiedDecision      = pepsdk.VerifiedDecision
)

// Governance: policies, web resources, risk signals.
type (
	GovernanceClient             = governancesdk.Client
	Policy                       = governancesdk.Policy
	CreatePolicyRequest          = governancesdk.CreatePolicyRequest
	RiskSignalRequest            = governancesdk.RiskSignalRequest
	CreateWebResourceRequest     = governancesdk.CreateWebResourceRequest
	CreateWebResourceRuleRequest = governancesdk.CreateWebResourceRuleRequest
)

// Domain client constructors.
var (
	NewIdentityClient    = identitysdk.New
	NewGrantClient       = grantsdk.New
	NewAttestationClient = attestationsdk.New
	NewGovernanceClient  = governancesdk.New
)
