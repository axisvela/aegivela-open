package pepsdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Authorizer interface {
	Authorize(ctx context.Context, route Route, input AuthorizationInput) (*Result, error)
}

type EvidenceRecorder interface {
	Record(ctx context.Context, evidence Evidence)
}

type Evidence struct {
	Mode        Mode
	Action      string
	Resource    Resource
	DecisionID  string
	PolicyVersion string
	EvidenceRefs []string
	TraceID     string
	Outcome     string
	ReasonCode  string
}

type contextKey struct{ name string }

var (
	resultCtxKey  = contextKey{name: "pepsdk-result"}
	inputCtxKey   = contextKey{name: "pepsdk-input"}
	evidenceCtxKey = contextKey{name: "pepsdk-evidence"}
)

func ResultFromContext(ctx context.Context) *Result {
	v, _ := ctx.Value(resultCtxKey).(*Result)
	return v
}

func AuthorizationInputFromContext(ctx context.Context) *AuthorizationInput {
	v, _ := ctx.Value(inputCtxKey).(*AuthorizationInput)
	return v
}

func EvidenceFromContext(ctx context.Context) *Evidence {
	v, _ := ctx.Value(evidenceCtxKey).(*Evidence)
	return v
}

type MiddlewareOption func(*Middleware)

func WithEvidenceRecorder(recorder EvidenceRecorder) MiddlewareOption {
	return func(m *Middleware) {
		m.evidenceRecorder = recorder
	}
}

func WithRoute(route Route) MiddlewareOption {
	return func(m *Middleware) {
		m.route = route
		m.routeSet = true
	}
}

func WithMode(mode Mode) MiddlewareOption {
	return func(m *Middleware) {
		m.mode = mode
	}
}

type Middleware struct {
	authorizer       Authorizer
	evidenceRecorder EvidenceRecorder
	route            Route
	routeSet         bool
	mode             Mode
}

func NewAuthorizeMiddleware(authorizer Authorizer, opts ...MiddlewareOption) *Middleware {
	if authorizer == nil {
		panic("pepsdk: nil authorizer")
	}
	m := &Middleware{
		authorizer: authorizer,
		mode:       ModeHumanWeb,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := m.route
		traceID := r.Header.Get("X-Aegivela-Trace-ID")

		input, evidence, err := m.extractCredentials(r)
		if err != nil {
			evidence.Mode = m.mode
			evidence.TraceID = traceID
			evidence.ReasonCode = "invalid_pep_request"
			m.recordEvidence(r.Context(), evidence)
			writeError(w, http.StatusBadRequest, "invalid_pep_request")
			return
		}
		evidence.Mode = m.mode
		evidence.TraceID = traceID
		evidence.Action = route.Action
		evidence.Resource = Resource{Kind: route.Kind, Reference: route.Reference}

		if m.routeSet {
			if err := m.route.Validate(); err != nil {
				evidence.ReasonCode = "invalid_pep_request"
				m.recordEvidence(r.Context(), evidence)
				writeError(w, http.StatusBadRequest, "invalid_pep_request")
				return
			}
		}

		result, err := m.authorizer.Authorize(r.Context(), route, input)
		if err != nil {
			evidence.Outcome = "deny"
			evidence.ReasonCode = "authorization_denied"
			if errors.Is(err, ErrUnauthenticated) {
				evidence.ReasonCode = "unauthenticated"
				m.recordEvidence(r.Context(), evidence)
				writeError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if errors.Is(err, ErrDenied) {
				m.recordEvidence(r.Context(), evidence)
				writeError(w, http.StatusForbidden, "authorization_denied")
				return
			}
			evidence.ReasonCode = "authorization_unavailable"
			m.recordEvidence(r.Context(), evidence)
			writeError(w, http.StatusServiceUnavailable, "authorization_unavailable")
			return
		}

		evidence.DecisionID = result.DecisionID
		evidence.PolicyVersion = result.PolicyVersion
		evidence.EvidenceRefs = result.EvidenceRefs
		evidence.Outcome = result.Outcome

		ctx := context.WithValue(r.Context(), resultCtxKey, result)
		ctx = context.WithValue(ctx, inputCtxKey, &input)
		ctx = context.WithValue(ctx, evidenceCtxKey, &evidence)
		r = r.WithContext(ctx)

		m.recordEvidence(r.Context(), evidence)
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) recordEvidence(ctx context.Context, e Evidence) {
	if m.evidenceRecorder != nil {
		m.evidenceRecorder.Record(ctx, e)
	}
}

func (m *Middleware) extractCredentials(r *http.Request) (AuthorizationInput, Evidence, error) {
	bearerTokens := r.Header["Authorization"]
	workloads := r.Header["X-Aegivela-Workload-Assertion"]

	switch m.mode {
	case ModeHumanWeb:
		if len(workloads) > 0 {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		if len(bearerTokens) != 1 {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		token := extractBearerToken(bearerTokens[0])
		if token == "" {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		return AuthorizationInput{
			Mode:        ModeHumanWeb,
			BearerToken: token,
		}, Evidence{}, nil

	case ModeSystemAPI:
		if len(bearerTokens) > 0 {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		if len(workloads) != 1 {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		return AuthorizationInput{
			Mode:              ModeSystemAPI,
			WorkloadAssertion: workloads[0],
		}, Evidence{}, nil

	case ModeDelegatedAPI:
		if len(bearerTokens) != 1 {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		if len(workloads) != 1 {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		bearer := extractBearerToken(bearerTokens[0])
		if bearer == "" {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		return AuthorizationInput{
			Mode:              ModeDelegatedAPI,
			BearerToken:       bearer,
			WorkloadAssertion: workloads[0],
		}, Evidence{}, nil

	case ModeServiceAgentAPI, ModeTwinAgentAPI:
		if len(bearerTokens) > 0 {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		if len(workloads) != 1 {
			return AuthorizationInput{}, Evidence{}, ErrInvalidInput
		}
		return AuthorizationInput{
			Mode:              m.mode,
			WorkloadAssertion: workloads[0],
		}, Evidence{}, nil

	default:
		return AuthorizationInput{}, Evidence{}, ErrInvalidInput
	}
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return header[len(prefix):]
}

func writeError(w http.ResponseWriter, statusCode int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"api_version": "aegivela.io/v1alpha1",
		"code":        code,
	})
}
