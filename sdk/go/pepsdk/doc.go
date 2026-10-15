// Package pepsdk is the Apache-2.0 AEGIVELA runtime-enforcement client:
//
//   - policy decisions: Evaluate, Authorize, SimulateDecision, VerifyDecision, DecisionJWKS;
//   - gateway/tool enforcement: GatewayConnect, ModuregisAuthorize, ToolInvoke;
//   - revocation: CheckRevocation, RevocationList;
//   - HTTP authorization middleware.
//
// Domain clients live in sibling packages: identitysdk (workload bindings +
// assertions + JWKS), grantsdk (execution grants + approvals + pre-auth +
// token-exchange), attestationsdk (attestations), governancesdk (policies, web
// resources, risk signals). The product-named root package
// github.com/axisrobo/aegivela-open/sdk/go re-exports them all.
//
// Evaluate owns the shared request/response path and returns valid deny/approval
// outcomes as data. Authorize rejects all non-allow outcomes. Neither method
// verifies SignedDecision cryptographically; consumers requiring execution
// authority must verify the artifact and its expected bindings before dispatch.
// V2 inputs use PolicyPrincipalV2 rather than raw JSON or duplicated namespace/
// actor fields. Service/twin agents require workload proof, bound positive dual
// epochs and ParentAuthority; scope, audience, task and expiry cannot expand
// beyond those bounds. A default requested expiry is capped by the parent.
// ParentAuthority metadata is not independently verified proof of delegation.
// Delegated mode prohibits it and presents ParentExecutionGrant instead.
//
// Configure defaults at construction with NewClient(baseURL, httpClient,
// WithHeaders(headers)). The headers are copied, endpoint credentials take
// precedence, and cross-origin redirects are refused when defaults are set.
// Tenant configuration via WithTenant must also finish before concurrent use.
// Evaluate/Authorize HTTP failures expose *APIError with StatusCode; use
// errors.As to distinguish them from transport failures, and errors.Is for
// ErrDenied, ErrUnauthenticated, ErrUnavailable or ErrMalformed. Transport
// failures retain their underlying cancellation/deadline cause. HTTP error
// bodies are not retained in APIError.
//
// Stability: the exported clients and their request/decision types are stable
// within a contract version. The embedded contract data lives in
// pepsdk/moduregiscontract and follows the contracts compatibility policy.
package pepsdk
