package pepsdk

import "errors"

type Mode string

const (
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
	if len(r.Scope) == 0 {
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
}

func (in AuthorizationInput) Validate() error {
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
	return nil
}

type Principal struct {
	TenantID                   string
	SubjectRef                 string
	ActorRef                   string
	ClientID                   string
	WorkloadRef                string
	AgentID                    string
	AgentClass                 string
	MasterID                   string
	OrganizationAuthorityRoot  string
	LifecycleEpoch             int64
	AttestationRef             string
}

type Result struct {
	DecisionID    string
	PolicyVersion string
	EvidenceRefs  []string
	Outcome       string
}
