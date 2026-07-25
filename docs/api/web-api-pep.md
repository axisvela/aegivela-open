# Web/API PEP Integration Guide

AEGIVELA publishes a product-neutral, versioned Web/API PEP contract and Go reference SDK. A consuming web application or API gateway places the PEP middleware in front of protected server-side routes and derives authority from enterprise-issued OIDC tokens and workload assertions. AEGIVELA does not operate browser sessions, OIDC redirects, PKCE callbacks, refresh-token handling, or logout.

## Route Configuration

Every protected route declares its trusted server-side attributes. The caller cannot supply or override them:

- `audience` — exact target audience for the policy evaluation.
- `modes` — the allowed authorization modes for this route (`human_web`, `system_api`, `delegated_api`).
- `action` — the action the route performs.
- `resource` — the concrete resource kind and reference this route accesses.
- `scope` — the requested scope set.

## Credential Headers

The consumer supplies mode-appropriate credentials as headers on the incoming HTTP request, never as route configuration:

| Mode | Required headers |
| --- | --- |
| `human_web` | `Authorization: Bearer <enterprise-access-token>` |
| `system_api` | `X-Aegivela-Workload-Assertion: <compact-assertion>` |
| `delegated_api` | `Authorization: Bearer <enterprise-access-token>`, `X-Aegivela-Workload-Assertion: <compact-assertion>`, and a v1alpha3 signed execution grant |

Duplicate or mode-incompatible headers are rejected before the authorization call.

## Authority Resolution

The PEP middleware constructs the following internal evaluate request from route configuration and headers, never from application input:

1. The route supplies audience, action, resource, and requested scope.
2. The mode and credentials come from the incoming request headers.
3. AEGIVELA resolves the trusted principal from the credentials: Identity Bridge for `human_web`, workload assertion and binding for `system_api`, and both plus verified parent grant for `delegated_api`.
4. AEGIVELA evaluates policy and returns an enforceable decision.

A non-`allow` outcome denies access. An unavailable identity, binding, policy, or revocation dependency returns `503 authorization_unavailable` and fails closed.

## Middleware Error Mapping

| Condition | HTTP status | Body |
| --- | --- | --- |
| Invalid, missing, forged, expired, revoked, or mode-incompatible evidence | `401` | `{"code":"unauthenticated"}` |
| Trusted identity/workload with non-allow policy result | `403` | `{"code":"authorization_denied"}` |
| Identity Bridge, workload, grant verification, PDP, or revocation dependency unavailable | `503` | `{"code":"authorization_unavailable"}` |
| Invalid route configuration or malformed credential headers | `400` | `{"code":"invalid_pep_request"}` |

## Result Context

After a successful allow decision, the middleware stores the resolved `Result` in the request context. Downstream handlers retrieve it with:

```go
result, ok := pepsdk.ResultFromContext(r.Context())
```

The context never contains raw bearer tokens, workload assertions, or parent execution grants.

## Evidence Redaction

An `EvidenceRecorder` receives a terminal record for every outcome. Evidence contains mode, route action/resource, decision ID, policy version, evidence references, trace ID, outcome, and reason code. It excludes all credential-bearing headers and raw payloads.

## Product Responsibility

The consuming product owns:

- OIDC authorization-code-with-PKCE, redirect handling, session management, and refresh-token lifecycle.
- Out-of-band workload binding and assertion issuance through AEGIVELA's workload API before the protected route is reached.
- Parent execution grant issuance (or token exchange) for delegated API routes.
- Any product-specific resource semantics, cache invalidation, and downstream enforcement.

## Local Testing

Use a local OIDC test issuer and workload assertion configuration as described in the [Identity Bridge integration guide](identity-bridge.md) and the [operations development guide](../operations/development.md). The Go PEP SDK includes a reference implementation and integration test coverage that exercises all three modes and their failure paths against the in-process AEGIVELA API.
