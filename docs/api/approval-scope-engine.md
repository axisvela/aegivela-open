# Approval Scope Engine Contract

## Overview

`POST /v1/approvals/issue` creates persistent, auditable approval records that downstream grant issuance (`POST /v1/grants/issue`) can bind to via `approval_jti`. Approvals are append-only, immutable, and tenant-scoped. Raw reason text and bearer tokens are never persisted; only SHA-256 digests enter the approval record.

## Authentication

The endpoint is internally authenticated with exactly one valid `X-Aegivela-Internal-Token` header. Requests without a valid internal token receive `401 unauthenticated`.

## Request

### Accepted Fields

```json
{
  "bearer_token": "<identity-bearer-token>",
  "action": "invoke",
  "resource": {"kind": "gateway:backend", "id": "backend-1", "version": "v1"},
  "scope": ["capability:read", "tool:invoke"],
  "expiry": "2026-08-01T00:00:00Z",
  "reason": "approved by security team",
  "policy_version": "policy-v1",
  "session": "session-1",
  "lifecycle": 42,
  "trace_id": "trace-1",
  "evidence_refs": ["ref-a", "ref-b"]
}
```

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `bearer_token` | string | Yes | Identity bearer token resolved via Identity Bridge. |
| `action` | string | Yes | Bounded action (max 256 chars, `[a-zA-Z0-9-._:/]`). |
| `resource` | object | Yes | ResourceRef with `kind`, `id`, and optional `version`. |
| `scope` | []string | Yes | Non-empty scope set in `category:action` format (max 64 entries). Normalized (deduplicated, sorted) before storage. |
| `expiry` | string | Yes | RFC 3339 timestamp. Must not be zero, past, or exceed the principal's remaining lifetime. Maximum approval TTL is 24 hours. |
| `reason` | string | Yes | Human-readable reason text. Only the SHA-256 digest (`reason_digest`) is persisted; raw text is discarded. |
| `policy_version` | string | Yes | Bounded policy version reference. |
| `session` | string | No | Optional session reference derived from or correlated with the principal. |
| `lifecycle` | int64 | No | Optional lifecycle epoch; must be non-negative when present. |
| `trace_id` | string | No | Validated trace identifier (max 128 chars, starts with alphanumeric). Empty is allowed and treated as absent. |
| `evidence_refs` | []string | No | Bounded evidence references (max 32 entries, `[a-zA-Z0-9-._:/]`). Accepted for forward compatibility; not yet column-persisted. |

### Rejected Fields

Any field not listed above is rejected with `400 invalid_approval_request`. In particular, the following identity overrides are always rejected:

- `tenant_id`
- `subject_ref`
- `actor_id`

The caller cannot control tenant, subject, or actor; these are derived exclusively from the resolved Identity Bridge principal.

## Identity Resolution

The handler resolves `bearer_token` via the Identity Bridge using audience `aegivela`. The resulting principal supplies:

- `tenant_id` — the tenant owning the approval.
- `subject_digest` — SHA-256 of the principal's `subject_ref`; the raw subject reference is never stored.
- `actor_id` — the actor that authorized the approval.

If identity resolution fails with `ErrUnauthenticated`, the handler returns `401 unauthenticated`. If the identity bridge is unavailable, the handler returns `503 identity_unavailable`.

## Scope Semantics

Scope entries must follow `category:action` format (e.g., `capability:read`, `tool:invoke`). Entries missing the `:` delimiter are rejected at request validation. The scope set is normalized before storage:

- Entries are trimmed and deduplicated.
- Sorted lexicographically.
- Stored as a `\x00`-separated normalized string in PostgreSQL.

Empty scope and scope entries exceeding 64 items are rejected.

## Reason Redaction

The raw `reason` string is never written to the database. The handler computes `SHA-256(reason)` during request processing and stores only the resulting hex-digest as `reason_digest`. The raw text is discarded immediately after digest computation.

## Response

### Success (201 Created)

```json
{
  "approval_jti": "eac1cc47-c2ff-4dc7-ae6f-c6b05be57b0a"
}
```

The `approval_jti` is a UUID v4 that uniquely identifies the approval record within its tenant. Callers pass this value as `approval_jti` when issuing execution grants.

### Error Mapping

| Condition | Status | Code |
| --- | --- | --- |
| Invalid JSON, unknown fields, missing required fields, invalid scope format, invalid resource, invalid expiry format, past expiry, expiry exceeds maximum TTL, invalid trace_id, invalid evidence_refs | 400 | `invalid_approval_request` |
| Missing or invalid internal token | 401 | `unauthenticated` |
| Failed bearer token resolution (unauthenticated) | 401 | `unauthenticated` |
| Expiry exceeds principal's remaining lifetime | 403 | `authorization_denied` |
| Identity bridge unavailable | 503 | `identity_unavailable` |
| Approval persistence unavailable, JTI generation failure | 503 | `approval_issuance_unavailable` |

## Grant Integration

Downstream grant issuance uses `approval_jti` to bind an execution grant to a pre-existing approval. See the [Execution Grant Issuance Contract](execution-grants.md):

- The grant issuer validates the approval exists, is unexpired, and belongs to the requesting tenant.
- The grant's scope, audience, expiry, action, resource, and task binding must be equal to or narrower than the approval's.
- A revoked or expired approval denies downstream grants with `403 decision_denied`.
- An unavailable approval revocation check returns `503`.

Grants without `approval_jti` preserve existing behavior and require no approval record.

## Persistence Properties

Approval records are append-only with immutable UPDATE/DELETE/TRUNCATE triggers. The `reason_digest` column contains only SHA-256 hex digests; raw reason text, bearer artifacts, and raw subject references are never stored. Evidence references are validated at the HTTP boundary; column persistence for `evidence_refs` is a future migration concern.
