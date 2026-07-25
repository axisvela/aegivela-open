# Token Exchange and Pre-Authorization Contract

## Overview

`POST /v1/token-exchange` issues a short-lived execution grant only by attenuating a verified parent execution grant or a canonical `allow` decision produced from a trusted principal. An enterprise bearer token is evidence for trusted-principal resolution, not authority to mint a grant by itself. `POST /v1/pre-authorizations/issue` creates a persistent, tenant-scoped window that bounds how many downstream execution grants may be issued under a single authorization decision.

Both endpoints are internally authenticated with exactly one valid `X-Aegivela-Internal-Token` header. They depend on the Identity Bridge to resolve bearer tokens into normalized principals.

---

## Token Exchange

### Endpoint

`POST /v1/token-exchange`

### Subject Token Types

| Type | Description |
| --- | --- |
| `enterprise_bearer` | An IdP-issued bearer token resolved to a trusted principal and paired with that principal's canonical verified `allow` decision. The principal and decision expiries bound the exchange grant's maximum lifetime. An unverified bearer has no exchange authority. |
| `execution_grant` | An existing execution grant (delegation chain). The exchange grant's scope, audience, action, resource, and expiry are attenuated from the subject grant. |

### Request

```json
{
  "subject_token": "<bearer-or-execution-grant-token>",
  "subject_token_type": "enterprise_bearer",
  "signed_allow_decision": "<compact-EdDSA-JWS-required-for-enterprise-bearer>",
  "scope": ["capability:read"],
  "audience": "gateway",
  "action": "invoke",
  "resource_ref": {"kind": "gateway:backend", "id": "backend-1", "version": "v1"},
  "trace_id": "trace-1"
}
```

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `subject_token` | string | Yes | The bearer or execution grant token to exchange. |
| `subject_token_type` | string | Yes | Must be `enterprise_bearer` or `execution_grant`. |
| `signed_allow_decision` | string | For `enterprise_bearer` | Canonical verified `allow` decision produced from the resolved trusted principal. Not accepted as a replacement for an execution-grant parent. |
| `scope` | []string | Yes | Non-empty scope set. Must be a subset (or equal to) the subject token's scope for execution grant exchanges. |
| `audience` | string | Yes | Target audience for the issued grant. For execution grant exchanges, must match the subject grant's audience. |
| `action` | string | Yes | Action authorized by the issued grant. For execution grant exchanges, must match the subject grant's action. |
| `resource_ref` | object | Yes | ResourceRef with `kind`, `id`, and optional `version`. For execution grant exchanges, must match the subject grant's resource. |
| `trace_id` | string | No | Validated trace identifier (max 128 chars, starts with alphanumeric). |

### Attenuation Semantics

The parent authority is either a verified execution grant or a canonical verified `allow` decision created from a trusted principal. Token exchange preserves the parent tenant, action, resource, task, lifecycle, audience, scope, and expiry bounds; it never turns an unverified bearer artifact or raw policy JSON into authority.

When exchanging an execution grant (`subject_token_type: "execution_grant"`):

- **Scope:** The exchange request scope must be a subset of the subject grant's scope. Widening scope is rejected with `403 exchange_denied`.
- **Audience:** Must match the subject grant's audience exactly. Cross-audience exchange is rejected with `403 exchange_denied`.
- **Action:** Must match the subject grant's action exactly. Cross-action exchange is rejected with `403 exchange_denied`.
- **Resource:** Must match the subject grant's resource exactly. Cross-resource exchange is rejected with `403 exchange_denied`.
- **Expiry:** The issued exchange grant expires at the earliest of: 5 minutes from now, the subject grant's expiry, and the subject principal's expiry. Explicit requested expiry wider than this bound is rejected.

### Exchange Grant Properties

- Maximum TTL: 5 minutes
- Policy version: `exchange-v1` (fixed, not caller-controlled)
- Expiry is bounded by both the subject token's lifetime and the 5-minute maximum

### Response

**Success (201 Created):**

```json
{"grant": "<execution-grant-token>"}
```

### Failure Semantics

| Condition | Status | Code |
| --- | --- | --- |
| Missing or empty subject token, audience, action | 401 | `unauthenticated` |
| Invalid subject token type | 401 | `unauthenticated` |
| Invalid JSON, unknown fields | 400 | `invalid_exchange_request` |
| Enterprise bearer token, canonical allow decision, or execution grant not authenticated | 401 | `unauthenticated` |
| Authenticated artifact lacks an allow parent or violates parent binding | 403 | `exchange_denied` |
| Scope widening, audience/action/resource mismatch, expiry extension | 403 | `exchange_denied` |
| Identity, policy-decision, grant-verification, or revocation dependency unavailable | 503 | `token_exchange_unavailable` |
| Missing or invalid internal token | 401 | `unauthenticated` |
| Request body too large | 413 | `request_too_large` |

---

## Pre-Authorization Window

### Endpoint

`POST /v1/pre-authorizations/issue`

### Purpose

A pre-authorization window provisions a quota-bounded authorization envelope that enables downstream grant issuance without requiring a fresh signed policy decision for each grant. The window specifies:

- **Max grants:** An upper bound on how many execution grants may be issued under this window (1-100).
- **Effective scope:** The scope set to which all issued grants are attenuated.
- **Expiry:** The time at which the window closes (and remaining quota is forfeited).

Once exhausted (`used_grants >= max_grants`), further grant issuance attempts under this window's JTI fail with `403 grant_denied`. Expired windows similarly deny issuance.

### Request

```json
{
  "bearer_token": "<identity-bearer-token>",
  "action": "deploy",
  "resource_ref": {"kind": "gateway:backend", "id": "backend-1", "version": "v1"},
  "scope": ["capability:read", "tool:invoke"],
  "expiry": "2099-01-01T00:00:00Z",
  "policy_version": "policy-v1",
  "audience": "gateway",
  "trace_id": "trace-123",
  "task_ref": "task-1",
  "max_grants": 10
}
```

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `bearer_token` | string | Yes | Identity bearer token resolved via Identity Bridge. |
| `action` | string | Yes | Bounded action (max 256 chars, `[a-zA-Z0-9-._:/]`). |
| `resource_ref` | object | Yes | ResourceRef with `kind`, `id`, and optional `version`. |
| `scope` | []string | Yes | Non-empty scope set in `category:action` format (max 64 entries, `:` delimiter required). |
| `expiry` | string | Yes | RFC 3339 timestamp. Must not be past, must not exceed 24 hours from now, and must not exceed principal expiry. |
| `policy_version` | string | Yes | Bounded policy version reference. |
| `audience` | string | Yes | Target audience for grants issued under this window. |
| `trace_id` | string | No | Validated trace identifier. |
| `task_ref` | string | No | Optional task reference. |
| `max_grants` | int | Yes | Maximum number of grants issuable under this window (1-100). |

### Evidence Redaction

The following fields are derived from identity resolution and are never stored in raw form:

| Raw Value | Stored As | Redaction |
| --- | --- | --- |
| `principal.SubjectRef` | `subject_digest` | SHA-256 hex digest |
| `ResourceRef` (kind, id, version) | `resource_digest` | SHA-256 hex digest of `kind\x00id\x00version` |
| Raw subject reference, bearer token | Not stored | Discarded immediately |

The `bearer_token` is consumed only for identity resolution; it is never written to PostgreSQL or logged.

### Scope Semantics

- Scope entries must follow `category:action` format (e.g., `capability:read`).
- Entries missing the `:` delimiter are rejected.
- Each scope entry is validated against `validPreAuthField` (`[a-zA-Z0-9-._:/]`, max 256 chars).
- Duplicate scope entries are rejected at the PostgreSQL layer (normalization deduplicates).
- Maximum 64 entries.

### Window Exhaustion

The window tracks `used_grants` against `max_grants`. Each successful grant issuance that references this window's JTI atomically increments `used_grants`. When `used_grants` reaches `max_grants`:

- Further `ConsumeGrant` calls return `ErrGrantExhausted`.
- The window row remains immutable; `used_grants` is the only mutable column.
- The row is never deleted; UPDATE, DELETE, and TRUNCATE are blocked by PostgreSQL triggers.
- Expired windows (where `expires_at <= NOW()`) also deny `ConsumeGrant`.

### Response

**Success (201 Created):**

```json
{"pre_authorization_jti": "eac1cc47-c2ff-4dc7-ae6f-c6b05be57b0a"}
```

The `pre_authorization_jti` is a UUID v4 that uniquely identifies the window within its tenant.

### Failure Semantics

| Condition | Status | Code |
| --- | --- | --- |
| Invalid JSON, unknown fields, missing required fields | 400 | `invalid_preauthorization_request` |
| Scope missing `:` delimiter, empty scope, scope > 64 entries | 400 | `invalid_preauthorization_request` |
| Invalid resource_ref | 400 | `invalid_preauthorization_request` |
| Past expiry, expiry exceeds 24h TTL, malformed expiry | 400 | `invalid_preauthorization_request` |
| `max_grants` < 1 or > 100 | 400 | `invalid_preauthorization_request` |
| Invalid trace_id | 400 | `invalid_preauthorization_request` |
| Expiry exceeds principal's remaining lifetime | 403 | `authorization_denied` |
| Expired principal | 403 | `authorization_denied` |
| Missing or invalid internal token | 401 | `unauthenticated` |
| Identity resolution failure (unauthenticated) | 401 | `unauthenticated` |
| Identity bridge unavailable | 503 | `identity_unavailable` |
| Repository unavailable, JTI generation failure | 503 | `preauthorization_issuance_unavailable` |
| Request body too large | 413 | `request_too_large` |

### Grant Integration

Downstream grant issuance uses `pre_authorization_jti` to bind grants to a pre-existing window. See the [Execution Grant Issuance Contract](execution-grants.md):

- The grant coordinator validates the window exists, is unexpired, and belongs to the requesting tenant.
- The grant's scope, audience, action, resource, and expiry must be equal to or narrower than the window's.
- Each successful grant atomically increments `used_grants`.
- An exhausted or expired window denies downstream grants with `403 grant_denied`.

### Immutable Row Properties

The `pre_authorizations` table uses PostgreSQL triggers to enforce immutability:
- `pre_authorizations_update_guard`: Blocks UPDATE operations.
- `pre_authorizations_delete_guard`: Blocks DELETE operations.
- `pre_authorizations_truncate_immutable`: Blocks TRUNCATE operations.
- `ConsumeGrant` is the only path that modifies a row (incrementing `used_grants` atomically).
