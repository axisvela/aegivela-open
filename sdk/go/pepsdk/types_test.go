package pepsdk

import (
	"errors"
	"testing"
)

func TestRouteValidate(t *testing.T) {
	makeValid := func() Route {
		return Route{
			Audience:  "research-api",
			Action:    "datasets.read",
			Scope:     []string{"datasets.read"},
			Kind:      "dataset",
			Reference: "dataset-001",
			Modes:     []Mode{ModeHumanWeb, ModeSystemAPI},
		}
	}

	t.Run("valid", func(t *testing.T) {
		if err := makeValid().Validate(); err != nil {
			t.Fatalf("valid route rejected: %v", err)
		}
	})

	t.Run("nil modes is ok", func(t *testing.T) {
		r := makeValid()
		r.Modes = nil
		if err := r.Validate(); err != nil {
			t.Fatalf("route with nil modes rejected: %v", err)
		}
	})

	t.Run("empty audience", func(t *testing.T) {
		r := makeValid()
		r.Audience = ""
		if err := r.Validate(); !errors.Is(err, ErrInvalidRoute) {
			t.Fatalf("expected ErrInvalidRoute, got %v", err)
		}
	})

	t.Run("empty action", func(t *testing.T) {
		r := makeValid()
		r.Action = ""
		if err := r.Validate(); !errors.Is(err, ErrInvalidRoute) {
			t.Fatalf("expected ErrInvalidRoute, got %v", err)
		}
	})

	t.Run("empty scope", func(t *testing.T) {
		r := makeValid()
		r.Scope = nil
		if err := r.Validate(); !errors.Is(err, ErrInvalidRoute) {
			t.Fatalf("expected ErrInvalidRoute, got %v", err)
		}
	})

	t.Run("empty kind", func(t *testing.T) {
		r := makeValid()
		r.Kind = ""
		if err := r.Validate(); !errors.Is(err, ErrInvalidRoute) {
			t.Fatalf("expected ErrInvalidRoute, got %v", err)
		}
	})

	t.Run("empty reference", func(t *testing.T) {
		r := makeValid()
		r.Reference = ""
		if err := r.Validate(); !errors.Is(err, ErrInvalidRoute) {
			t.Fatalf("expected ErrInvalidRoute, got %v", err)
		}
	})

	t.Run("unsupported mode", func(t *testing.T) {
		r := makeValid()
		r.Modes = []Mode{"unsupported_mode"}
		if err := r.Validate(); !errors.Is(err, ErrInvalidRoute) {
			t.Fatalf("expected ErrInvalidRoute for unsupported mode, got %v", err)
		}
	})
}

func TestAuthorizationInputValidate(t *testing.T) {
	makeHumanWeb := func() AuthorizationInput {
		return AuthorizationInput{
			Mode:        ModeHumanWeb,
			BearerToken: "eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiJ1c2VyLWFkYSJ9.c2lnbmF0dXJlLWV4YW1wbGU",
		}
	}

	makeSystemAPI := func() AuthorizationInput {
		return AuthorizationInput{
			Mode:              ModeSystemAPI,
			WorkloadAssertion: "eyJhbGciOiJFZERTQSJ9.eyJ3b3JrbG9hZCI6Indvcmtsb2FkLXByb2QifQ.c2lnbmF0dXJlLXN5cw",
		}
	}

	makeDelegated := func() AuthorizationInput {
		return AuthorizationInput{
			Mode:                 ModeDelegatedAPI,
			BearerToken:          "eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiJ1c2VyLWFkYSJ9.c2lnbmF0dXJlLWRlbGVnYXRl",
			WorkloadAssertion:    "eyJhbGciOiJFZERTQSJ9.eyJ3b3JrbG9hZCI6Indvcmtsb2FkLXByb2QifQ.c2lnbmF0dXJlLXdvcmtsb2Fk",
			ParentExecutionGrant: "grant-parent-delegated-001",
		}
	}

	t.Run("human_web valid", func(t *testing.T) {
		if err := makeHumanWeb().Validate(); err != nil {
			t.Fatalf("valid human_web rejected: %v", err)
		}
	})

	t.Run("system_api valid", func(t *testing.T) {
		if err := makeSystemAPI().Validate(); err != nil {
			t.Fatalf("valid system_api rejected: %v", err)
		}
	})

	t.Run("delegated_api valid", func(t *testing.T) {
		if err := makeDelegated().Validate(); err != nil {
			t.Fatalf("valid delegated_api rejected: %v", err)
		}
	})

	t.Run("human_web missing bearer", func(t *testing.T) {
		in := makeHumanWeb()
		in.BearerToken = ""
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("system_api missing workload", func(t *testing.T) {
		in := makeSystemAPI()
		in.WorkloadAssertion = ""
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("bearer on system_api", func(t *testing.T) {
		in := makeSystemAPI()
		in.BearerToken = "eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiJ1c2VyLWFkYSJ9.c2lnbmF0dXJlLWV4YW1wbGU"
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for bearer on system, got %v", err)
		}
	})

	t.Run("workload on human_web", func(t *testing.T) {
		in := makeHumanWeb()
		in.WorkloadAssertion = "eyJhbGciOiJFZERTQSJ9.eyJ3b3JrbG9hZCI6Indvcmtsb2FkLXByb2QifQ.c2lnbmF0dXJlLXN5cw"
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for workload on human, got %v", err)
		}
	})

	t.Run("missing grant on delegated_api", func(t *testing.T) {
		in := makeDelegated()
		in.ParentExecutionGrant = ""
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for missing grant on delegated, got %v", err)
		}
	})

	t.Run("missing bearer on delegated_api", func(t *testing.T) {
		in := makeDelegated()
		in.BearerToken = ""
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for missing bearer on delegated, got %v", err)
		}
	})

	t.Run("missing workload on delegated_api", func(t *testing.T) {
		in := makeDelegated()
		in.WorkloadAssertion = ""
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for missing workload on delegated, got %v", err)
		}
	})

	t.Run("extra grant on human_web", func(t *testing.T) {
		in := makeHumanWeb()
		in.ParentExecutionGrant = "some-grant"
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for extra grant on human, got %v", err)
		}
	})

	t.Run("extra workload on system_api", func(t *testing.T) {
		in := makeSystemAPI()
		in.BearerToken = "eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiJ1c2VyLWFkYSJ9.c2lnbmF0dXJlLWV4YW1wbGU"
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for extra bearer on system, got %v", err)
		}
	})

	t.Run("unsupported mode", func(t *testing.T) {
		in := AuthorizationInput{Mode: Mode("unsupported")}
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for unsupported mode, got %v", err)
		}
	})
}

func TestAuthorizationInputAgentModes(t *testing.T) {
	makeServiceAgent := func() AuthorizationInput {
		return AuthorizationInput{
			Mode:              ModeServiceAgentAPI,
			WorkloadAssertion: "eyJhbGciOiJFZERTQSJ9.eyJ3b3JrbG9hZCI6Indvcmtsb2FkLXByb2QifQ.c2lnbmF0dXJlLXNlcnZpY2U",
		}
	}

	makeTwinAgent := func() AuthorizationInput {
		return AuthorizationInput{
			Mode:              ModeTwinAgentAPI,
			WorkloadAssertion: "eyJhbGciOiJFZERTQSJ9.eyJ3b3JrbG9hZCI6Indvcmtsb2FkLXByb2QifQ.c2lnbmF0dXJlLXR3aW4",
		}
	}

	t.Run("service_agent_api valid", func(t *testing.T) {
		if err := makeServiceAgent().Validate(); err != nil {
			t.Fatalf("valid service_agent_api rejected: %v", err)
		}
	})

	t.Run("twin_agent_api valid", func(t *testing.T) {
		if err := makeTwinAgent().Validate(); err != nil {
			t.Fatalf("valid twin_agent_api rejected: %v", err)
		}
	})

	t.Run("service_agent_api missing workload", func(t *testing.T) {
		in := makeServiceAgent()
		in.WorkloadAssertion = ""
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("twin_agent_api missing workload", func(t *testing.T) {
		in := makeTwinAgent()
		in.WorkloadAssertion = ""
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("bearer on service_agent_api", func(t *testing.T) {
		in := makeServiceAgent()
		in.BearerToken = "eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiJ1c2VyLWFkYSJ9.c2lnbmF0dXJlLWV4YW1wbGU"
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for bearer on service_agent_api, got %v", err)
		}
	})

	t.Run("bearer on twin_agent_api", func(t *testing.T) {
		in := makeTwinAgent()
		in.BearerToken = "eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiJ1c2VyLWFkYSJ9.c2lnbmF0dXJlLWV4YW1wbGU"
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for bearer on twin_agent_api, got %v", err)
		}
	})

	t.Run("grant on service_agent_api", func(t *testing.T) {
		in := makeServiceAgent()
		in.ParentExecutionGrant = "grant-123"
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for grant on service_agent_api, got %v", err)
		}
	})

	t.Run("grant on twin_agent_api", func(t *testing.T) {
		in := makeTwinAgent()
		in.ParentExecutionGrant = "grant-123"
		if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for grant on twin_agent_api, got %v", err)
		}
	})
}

func TestRouteAgentModes(t *testing.T) {
	t.Run("service_agent_api accepted", func(t *testing.T) {
		r := Route{
			Audience:  "api",
			Action:    "read",
			Scope:     []string{"read"},
			Kind:      "dataset",
			Reference: "d1",
			Modes:     []Mode{ModeServiceAgentAPI},
		}
		if err := r.Validate(); err != nil {
			t.Fatalf("Route with service_agent_api rejected: %v", err)
		}
	})

	t.Run("twin_agent_api accepted", func(t *testing.T) {
		r := Route{
			Audience:  "api",
			Action:    "read",
			Scope:     []string{"read"},
			Kind:      "dataset",
			Reference: "d1",
			Modes:     []Mode{ModeTwinAgentAPI},
		}
		if err := r.Validate(); err != nil {
			t.Fatalf("Route with twin_agent_api rejected: %v", err)
		}
	})

	t.Run("all five modes accepted", func(t *testing.T) {
		r := Route{
			Audience:  "api",
			Action:    "read",
			Scope:     []string{"read"},
			Kind:      "dataset",
			Reference: "d1",
			Modes:     []Mode{ModeHumanWeb, ModeSystemAPI, ModeDelegatedAPI, ModeServiceAgentAPI, ModeTwinAgentAPI},
		}
		if err := r.Validate(); err != nil {
			t.Fatalf("Route with all five modes rejected: %v", err)
		}
	})
}
