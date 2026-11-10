package pepsdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testAuthorizer struct {
	allow bool
	err   error
}

func (a *testAuthorizer) Authorize(_ context.Context, _ Route, _ AuthorizationInput) (*Result, error) {
	if a.err != nil {
		return nil, a.err
	}
	if a.allow {
		return &Result{DecisionID: "d-1", PolicyVersion: "v1", Outcome: "allow"}, nil
	}
	return nil, ErrDenied
}

type testEvidenceRecorder struct {
	evidences []Evidence
}

func (r *testEvidenceRecorder) Record(_ context.Context, e Evidence) {
	r.evidences = append(r.evidences, e)
}

func TestMiddlewareHandlerCalling(t *testing.T) {
	tests := []struct {
		name       string
		allow      bool
		err        error
		wantStatus int
		wantCalled bool
	}{
		{"allow", true, nil, http.StatusOK, true},
		{"deny", false, nil, http.StatusForbidden, false},
		{"unauthenticated", false, ErrUnauthenticated, http.StatusUnauthorized, false},
		{"unavailable", false, ErrUnavailable, http.StatusServiceUnavailable, false},
		{"unknown error", false, errors.New("unknown"), http.StatusServiceUnavailable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewAuthorizeMiddleware(&testAuthorizer{allow: tt.allow, err: tt.err})
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Authorization", "Bearer eyJ.a.b")
			rec := httptest.NewRecorder()

			mw.Middleware(next).ServeHTTP(rec, req)

			if called != tt.wantCalled {
				t.Fatalf("handler called=%v, want=%v", called, tt.wantCalled)
			}
			if tt.wantStatus > 0 && rec.Code != tt.wantStatus {
				t.Errorf("status=%d, want=%d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestMiddlewareContextValues(t *testing.T) {
	t.Run("ResultFromContext after allow", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		var got *Result
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = ResultFromContext(r.Context())
		})
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if got == nil || got.DecisionID != "d-1" {
			t.Fatalf("unexpected result: %+v", got)
		}
	})

	t.Run("ResultFromContext nil from empty", func(t *testing.T) {
		if ResultFromContext(context.Background()) != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("AuthorizationInputFromContext after allow", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		var got *AuthorizationInput
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = AuthorizationInputFromContext(r.Context())
		})
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if got == nil || got.BearerToken != "eyJ.a.b" || got.Mode != ModeHumanWeb {
			t.Fatalf("unexpected input: %+v", got)
		}
	})

	t.Run("AuthorizationInputFromContext nil", func(t *testing.T) {
		if AuthorizationInputFromContext(context.Background()) != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("EvidenceFromContext after allow", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		var got *Evidence
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = EvidenceFromContext(r.Context())
		})
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if got == nil || got.Outcome != "allow" {
			t.Fatalf("unexpected evidence: %+v", got)
		}
	})

	t.Run("EvidenceFromContext nil", func(t *testing.T) {
		if EvidenceFromContext(context.Background()) != nil {
			t.Fatal("expected nil")
		}
	})
}

func TestMiddlewareEvidenceRecorder(t *testing.T) {
	recorder := &testEvidenceRecorder{}

	t.Run("called on allow", func(t *testing.T) {
		recorder.evidences = nil
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true}, WithEvidenceRecorder(recorder))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")
		req.Header.Set("X-Aegivela-Trace-ID", "trace-001")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if len(recorder.evidences) != 1 {
			t.Fatalf("expected 1 evidence, got %d", len(recorder.evidences))
		}
		e := recorder.evidences[0]
		if e.Outcome != "allow" || e.DecisionID != "d-1" || e.TraceID != "trace-001" {
			t.Errorf("evidence fields: outcome=%q decision=%q trace=%q", e.Outcome, e.DecisionID, e.TraceID)
		}
	})

	t.Run("called on deny", func(t *testing.T) {
		recorder.evidences = nil
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: false}, WithEvidenceRecorder(recorder))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if len(recorder.evidences) != 1 || recorder.evidences[0].Outcome != "deny" {
			t.Fatalf("evidence: count=%d outcome=%q", len(recorder.evidences), recorder.evidences[0].Outcome)
		}
	})

	t.Run("called on header error", func(t *testing.T) {
		recorder.evidences = nil
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true}, WithEvidenceRecorder(recorder))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rec := httptest.NewRecorder()

		mw.Middleware(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status=%d", rec.Code)
		}
		if len(recorder.evidences) != 1 || recorder.evidences[0].ReasonCode != "invalid_pep_request" {
			t.Fatalf("evidence: count=%d reason=%q", len(recorder.evidences), recorder.evidences[0].ReasonCode)
		}
	})
}

func TestMiddlewareHeaderValidation(t *testing.T) {
	t.Run("duplicate Authorization rejected", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Add("Authorization", "Bearer eyJ.a.b")
		req.Header.Add("Authorization", "Bearer eyJ.c.d")

		rec := httptest.NewRecorder()
		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})

	t.Run("duplicate workload header rejected", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Add("X-Aegivela-Workload-Assertion", "eyJ.a.b")
		req.Header.Add("X-Aegivela-Workload-Assertion", "eyJ.c.d")

		rec := httptest.NewRecorder()
		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})
}

func TestMiddlewareSystemAPIHeaders(t *testing.T) {
	t.Run("valid workload only", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		mw.mode = ModeSystemAPI
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("X-Aegivela-Workload-Assertion", "eyJ.a.b")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if !called {
			t.Fatal("handler not called")
		}
	})

	t.Run("missing workload rejected", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		mw.mode = ModeSystemAPI
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rec := httptest.NewRecorder()

		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})

	t.Run("Authorization header rejected for system API", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		mw.mode = ModeSystemAPI
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("X-Aegivela-Workload-Assertion", "eyJ.a.b")
		req.Header.Set("Authorization", "Bearer eyJ.a.b")
		rec := httptest.NewRecorder()

		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})
}

func TestMiddlewareDelegatedAPIHeaders(t *testing.T) {
	mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
	mw.mode = ModeDelegatedAPI

	t.Run("both headers ok", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")
		req.Header.Set("X-Aegivela-Workload-Assertion", "eyJ.c.d")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if !called {
			t.Fatal("delegated with both headers should pass")
		}
	})

	t.Run("missing bearer rejected", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("X-Aegivela-Workload-Assertion", "eyJ.c.d")
		rec := httptest.NewRecorder()

		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})

	t.Run("missing workload rejected", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")
		rec := httptest.NewRecorder()

		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})
}

func TestMiddlewareErrorBodyJSON(t *testing.T) {
	mw := NewAuthorizeMiddleware(&testAuthorizer{err: ErrDenied})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer eyJ.a.b")
	rec := httptest.NewRecorder()

	mw.Middleware(next).ServeHTTP(rec, req)

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatal("Content-Type is not application/json")
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if body["api_version"] != "aegivela.io/v1alpha1" {
		t.Errorf("api_version=%v", body["api_version"])
	}
}

func TestMiddlewareRouteConfig(t *testing.T) {
	validRoute := Route{Audience: "test-api", Action: "read", Scope: []string{"read"}, Kind: "test", Reference: "test-001"}

	t.Run("valid route passes", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true}, WithRoute(validRoute))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")

		rec := httptest.NewRecorder()
		mw.Middleware(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status=%d", rec.Code)
		}
	})

	t.Run("invalid route rejected", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true}, WithRoute(Route{Audience: "test-api"}))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")

		rec := httptest.NewRecorder()
		mw.Middleware(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("route passed to authorizer", func(t *testing.T) {
		auth := &captureAuthorizer{}
		mw := NewAuthorizeMiddleware(auth, WithRoute(validRoute))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Authorization", "Bearer eyJ.a.b")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if auth.lastRoute.Audience != "test-api" {
			t.Errorf("route.Audience=%q", auth.lastRoute.Audience)
		}
	})
}

type captureAuthorizer struct {
	lastRoute Route
	lastInput AuthorizationInput
}

func (a *captureAuthorizer) Authorize(_ context.Context, r Route, input AuthorizationInput) (*Result, error) {
	a.lastRoute = r
	a.lastInput = input
	return &Result{DecisionID: "d-1", PolicyVersion: "v1", Outcome: "allow"}, nil
}

func TestMiddlewareNilAuthorizerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewAuthorizeMiddleware(nil)
}

func TestMiddlewareWithMode(t *testing.T) {
	mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true}, WithMode(ModeSystemAPI))
	if mw.mode != ModeSystemAPI {
		t.Errorf("mode=%q", mw.mode)
	}
}

func TestMiddlewareAgentModeHeaders(t *testing.T) {
	t.Run("service_agent_api valid workload only", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		mw.mode = ModeServiceAgentAPI
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("X-Aegivela-Workload-Assertion", "eyJ.a.b")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if !called {
			t.Fatal("handler not called for valid service_agent_api")
		}
	})

	t.Run("twin_agent_api valid workload only", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		mw.mode = ModeTwinAgentAPI
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("X-Aegivela-Workload-Assertion", "eyJ.a.b")

		mw.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

		if !called {
			t.Fatal("handler not called for valid twin_agent_api")
		}
	})

	t.Run("service_agent_api missing workload", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		mw.mode = ModeServiceAgentAPI
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rec := httptest.NewRecorder()

		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})

	t.Run("service_agent_api rejects bearer token", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		mw.mode = ModeServiceAgentAPI
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("X-Aegivela-Workload-Assertion", "eyJ.a.b")
		req.Header.Set("Authorization", "Bearer eyJ.a.b")
		rec := httptest.NewRecorder()

		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})

	t.Run("twin_agent_api rejects bearer token", func(t *testing.T) {
		mw := NewAuthorizeMiddleware(&testAuthorizer{allow: true})
		mw.mode = ModeTwinAgentAPI
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("X-Aegivela-Workload-Assertion", "eyJ.a.b")
		req.Header.Set("Authorization", "Bearer eyJ.a.b")
		rec := httptest.NewRecorder()

		mw.Middleware(next).ServeHTTP(rec, req)

		if called || rec.Code != http.StatusBadRequest {
			t.Fatalf("called=%v status=%d", called, rec.Code)
		}
	})
}
