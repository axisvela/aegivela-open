package pepsdk

import (
	"errors"
	"time"
)

type Mode string

const (
	APIVersionV1Alpha1 = "aegivela.io/v1alpha1"
	APIVersionV2       = "aegivela.io/v2.0"

	ModeHumanWeb        Mode = "human_web"
	ModeSystemAPI       Mode = "system_api"
	ModeDelegatedAPI    Mode = "delegated_api"
	ModeServiceAgentAPI Mode = "service_agent_api"
	ModeTwinAgentAPI    Mode = "twin_agent_api"
)

var (
	ErrInvalidRoute    = errors.New("invalid route")
	ErrInvalidInput    = errors.New("invalid authorization input")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrDenied          = errors.New("authorization denied")
	ErrUnavailable     = errors.New("authorization unavailable")
)

type Resource struct {
	Kind      string `json:"kind"`
	Reference string `json:"reference"`
}

type Route struct {
	Audience  string
	Action    string
	Scope     []string
	Kind      string
	Reference string
	Modes     []Mode
}

func (r Route) Validate() error {
	if r.Audience == "" {
		return ErrInvalidRoute
	}
	if r.Action == "" {
		return ErrInvalidRoute
	}
	if !validStringSet(r.Scope) {
		return ErrInvalidRoute
	}
	if r.Kind == "" {
		return ErrInvalidRoute
	}
	if r.Reference == "" {
		return ErrInvalidRoute
	}
	for _, m := range r.Modes {
		switch m {
		case ModeHumanWeb, ModeSystemAPI, ModeDelegatedAPI, ModeServiceAgentAPI, ModeTwinAgentAPI:
		default:
			return ErrInvalidRoute
		}
	}
	return nil
}

type AuthorizationInput struct {
	Mode                 Mode
	BearerToken          string
	WorkloadAssertion    string
	ParentExecutionGrant string
	APIVersion           string
	TaskBinding          string
	// Principal contains comparison-only identity hints for v2. The authority
	// must resolve and verify them from the presented credentials.
	Principal *PolicyPrincipalV2
	// ParentAuthority is the v2 agent-mode attenuation bound. Delegated mode
	// instead requires ParentExecutionGrant and prohibits caller-supplied bounds.
	ParentAuthority    *ParentAuthority
	RequestedExpiresAt time.Time
	TraceID            string
	// RiskLevel / RiskRef optionally carry a risk context onto a v2 evaluate
	// request.
	RiskLevel string
	RiskRef   string
}

func (in AuthorizationInput) Validate() error {
	if in.APIVersion != "" && in.APIVersion != APIVersionV1Alpha1 && in.APIVersion != APIVersionV2 {
		return ErrInvalidInput
	}
	if in.APIVersion == APIVersionV2 {
		if in.Principal == nil || in.Principal.Namespace == "" || in.Principal.ActorRef == "" {
			return ErrInvalidInput
		}
	} else if in.Principal != nil || in.ParentAuthority != nil {
		return ErrInvalidInput
	}
	switch in.Mode {
	case ModeHumanWeb:
		if in.BearerToken == "" {
			return ErrInvalidInput
		}
		if in.WorkloadAssertion != "" {
			return ErrInvalidInput
		}
		if in.ParentExecutionGrant != "" {
			return ErrInvalidInput
		}
	case ModeSystemAPI:
		if in.WorkloadAssertion == "" {
			return ErrInvalidInput
		}
		if in.BearerToken != "" {
			return ErrInvalidInput
		}
		if in.ParentExecutionGrant != "" {
			return ErrInvalidInput
		}
	case ModeDelegatedAPI:
		if in.BearerToken == "" {
			return ErrInvalidInput
		}
		if in.WorkloadAssertion == "" {
			return ErrInvalidInput
		}
		if in.ParentExecutionGrant == "" {
			return ErrInvalidInput
		}
	case ModeServiceAgentAPI, ModeTwinAgentAPI:
		if in.WorkloadAssertion == "" {
			return ErrInvalidInput
		}
		if in.BearerToken != "" {
			return ErrInvalidInput
		}
		if in.ParentExecutionGrant != "" {
			return ErrInvalidInput
		}
	default:
		return ErrInvalidInput
	}
	if in.APIVersion == APIVersionV2 {
		if in.Mode != ModeServiceAgentAPI && in.Mode != ModeTwinAgentAPI {
			if in.ParentAuthority != nil {
				return ErrInvalidInput
			}
			return nil
		}
		p := in.Principal
		if p.AgentID == "" || p.WorkloadRef == "" || p.AgentEpoch == nil || *p.AgentEpoch < 1 || p.IdentityEpoch == nil || *p.IdentityEpoch < 1 {
			return ErrInvalidInput
		}
		if in.Mode == ModeServiceAgentAPI {
			if p.AgentClass != "service" || p.OrganizationAuthorityRoot == "" || p.MasterID != "" || p.AuthorityBinding != "organizationRoot" {
				return ErrInvalidInput
			}
		} else if p.AgentClass != "twin" || p.MasterID == "" || p.OrganizationAuthorityRoot != "" || p.AuthorityBinding != "humanMaster" {
			return ErrInvalidInput
		}
		parent := in.ParentAuthority
		if parent == nil || parent.AuthorityRef == "" || !validStringSet(parent.Scope) || !validStringSet(parent.Audience) ||
			parent.ExpiresAt.IsZero() || parent.TaskBinding == "" || in.TaskBinding != parent.TaskBinding {
			return ErrInvalidInput
		}
	}
	return nil
}

// PolicyPrincipalV2 follows the supplier's camelCase principal contract.
// Epoch pointers distinguish absent values from explicitly invalid zero epochs.
type PolicyPrincipalV2 struct {
	Namespace                 string `json:"namespace"`
	SubjectRef                string `json:"subjectRef,omitempty"`
	ActorRef                  string `json:"actorRef"`
	ClientID                  string `json:"clientId,omitempty"`
	WorkloadRef               string `json:"workloadRef,omitempty"`
	AgentID                   string `json:"agentId,omitempty"`
	AgentClass                string `json:"agentClass,omitempty"`
	MasterID                  string `json:"masterId,omitempty"`
	OrganizationAuthorityRoot string `json:"organizationAuthorityRoot,omitempty"`
	AuthorityBinding          string `json:"authorityBinding,omitempty"`
	AgentEpoch                *int64 `json:"agentEpoch,omitempty"`
	IdentityEpoch             *int64 `json:"identityEpoch,omitempty"`
	InstanceID                string `json:"instanceId,omitempty"`
	AttestationRef            string `json:"attestationRef,omitempty"`
}

type ParentAuthority struct {
	AuthorityRef string    `json:"authorityRef"`
	Scope        []string  `json:"parentScope"`
	Audience     []string  `json:"parentAudience"`
	ExpiresAt    time.Time `json:"parentExpiresAt"`
	TaskBinding  string    `json:"parentTaskBinding"`
}

func validStringSet(values []string) bool {
	if len(values) == 0 {
		return false
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

type Principal struct {
	TenantID                  string
	SubjectRef                string
	ActorRef                  string
	ClientID                  string
	WorkloadRef               string
	AgentID                   string
	AgentClass                string
	MasterID                  string
	OrganizationAuthorityRoot string
	LifecycleEpoch            int64
	AttestationRef            string
}

type Result struct {
	DecisionID    string
	PolicyVersion string
	EvidenceRefs  []string
	Outcome       string
	// SignedDecision is the signed-decision artifact returned by evaluate,
	// when present, so the caller can verify it against the decision JWKS
	// (pepsdk.DecisionJWKS) before acting. Empty when the decision is unsigned.
	SignedDecision string
}
