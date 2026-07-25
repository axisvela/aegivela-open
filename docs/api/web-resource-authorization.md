# Web Resource Authorization API

Web Resource Authorization is the generic AEGIVELA capability for centrally managed, product-scoped UI resource trees and short-lived permission projections. It is not a business-action authorization API and does not grant access to a product backend.

## Resource and Rule Management

An integrating product registers its immutable UI resources and tenant-scoped projection rules through the internal AEGIVELA API. Resource identifiers are scoped by `product_id`; rules are additionally scoped by the authenticated tenant. The trusted internal context derives tenant and actor identity. Request bodies must not carry a tenant, principal, role, token, or arbitrary authorization context.

Resources expose only the projection actions `visible`, `enabled`, `read`, `edit`, and `submit`. A product retains ownership of its UI semantics, resource tree, domain state, and business actions.

## Permission Projection

`POST /v1/web-permissions:resolve` accepts a `product_id` and optional `root_resource_id` in JSON, plus exactly one `Authorization: Bearer <enterprise-oidc-jwt>` header and exactly one non-blank `X-Aegivela-Identity-Audience` header. `X-Aegivela-Internal-Token` remains the PEP authentication boundary and is not an identity artifact. AEGIVELA verifies the bearer through the configured Identity Bridge, rejects expired identities, and checks tenant-scoped subject and session revocations before resolving the projection.

The JSON body never accepts a bearer token, audience, tenant, principal, actor, role, or arbitrary authorization context. Missing or invalid identity, a tenant mismatch, or a revoked subject/session returns `401 unauthenticated`; unavailable Identity Bridge or revocation dependencies return `503 web_resource_unavailable`. If no Identity Bridge is configured, projection resolution returns `503`.

The projection is UI-only. A browser may use it to hide or disable controls, but it is not a credential and must never be treated as a backend authorization decision. Every product API or mutation PEP must independently request or enforce its applicable server-side AEGIVELA PDP decision and fail closed when that check is unavailable.

## Cache and Evidence Requirements

Products may cache a projection only until its `expires_at`. They must discard it on expiry and invalidate it when a referenced policy version or resource-tree version changes; a stale or unavailable refresh must not preserve access.

Each successful resolution records redacted security evidence: tenant, product, digested root resource, permission-set ID, policy/resource versions, trace ID when supplied, permission count, and expiry. Evidence excludes the permission map, bearer tokens, raw form fields, unrestricted request bodies, and principal secrets.

## Product Boundary

The generic Web Resource Authorization capability is available for product integrations. The MODUREGIS-specific adapter, including its Console/OIDC/PKCE integration and MODUREGIS resource mapping, remains pending. MODUREGIS continues to own its Capability lifecycle, Registry, Catalog, workflow, and Audit Index.
