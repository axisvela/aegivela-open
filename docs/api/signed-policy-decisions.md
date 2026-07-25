# Signed Policy Decision Contract

## Status and Authority

This F0 contract publishes signed artifacts created only at the canonical policy-evaluation boundary. It supplements the canonical [Policy Decision contract](policy-decision-contract.md); it does not allow a PEP or another HTTP caller to turn arbitrary policy JSON into an authorization artifact.

Each artifact is a compact EdDSA JWS. The protected header contains exactly `alg: "EdDSA"` and the configured `kid`. The payload accepts no unknown or duplicate JSON members. Consumers must reject any other algorithm, an unknown key ID, an invalid signature, malformed claims, an invalid time window, an unexpected issuer or audience, or an artifact longer than 8192 bytes.

## Claims and Outcomes

The signed payload contains the following claims:

| Claim | Meaning |
| --- | --- |
| `iss` | Configured policy-decision issuer. |
| `aud` | Exact target audience. |
| `jti` | Unique signed-decision artifact identifier. It is distinct from `decision_id`. |
| `tenant_id`, `subject_ref`, `actor_id` | Trusted identity bindings. |
| `outcome` | One of `allow`, `deny`, `approval_required`, or `revoked`. |
| `action`, `resource_ref`, `effective_scope` | Exact authorized operation and effective scope. `resource_ref` contains bounded `kind`, `id`, and optional `version`. |
| `policy_version`, `decision_id` | Canonical policy result identifiers. |
| `session_ref`, `lifecycle_epoch`, `task_ref` | Optional lifecycle and task bindings. |
| `nbf`, `iat`, `exp` | Numeric-date validity window. |

All required identifier fields are non-empty bounded identifiers. `effective_scope` is non-empty, unique, and limited to 64 entries. The validity window must be positive and no longer than five minutes. Artifacts are not valid before `nbf`, after `exp`, or when `iat` is in the future.

Every canonical outcome is signed. Only a verified `allow` can authorize execution-grant issuance. Verified `deny`, `approval_required`, and `revoked` outcomes deny issuance and execution; `approval_required` is not a deferred allow.

## Signing and Key Domain

The API process requires an independent decision-signing configuration:

| Environment variable | Purpose |
| --- | --- |
| `AEGIVELA_POLICY_DECISION_SIGNING_KEY_FILE` | Path to exactly one PEM-encoded PKCS#8 Ed25519 private key. |
| `AEGIVELA_POLICY_DECISION_SIGNING_KEY_ID` | Public key identifier placed in `kid`. |
| `AEGIVELA_POLICY_DECISION_ISSUER` | Issuer placed in `iss` and required by verifiers. |

These values are separate from `AEGIVELA_EXECUTION_GRANT_SIGNING_KEY_FILE`, `AEGIVELA_EXECUTION_GRANT_SIGNING_KEY_ID`, and `AEGIVELA_EXECUTION_GRANT_ISSUER`. A decision signing key must not be reused as an execution-grant key.

`GET /v1/policy/decisions/jwks` returns the active decision verification key as a public-only JWK Set. The JWK uses `kty: "OKP"`, `crv: "Ed25519"`, `use: "sig"`, `alg: "EdDSA"`, the configured `kid`, and base64url public coordinate `x`. It never returns private key material.

The canonical policy service constructs signed decisions from a validated request, a principal resolved by `trustedprincipal`, and the canonical evaluator response. `POST /v1/policy/decisions/evaluate` and `POST /v1/policy/decisions/sign` require internal authentication, then resolve the principal from mode-appropriate trusted evidence before evaluation. `human_web` resolves through Identity Bridge; `system_api` and `delegated_api` require a verified workload assertion and active workload binding; Twin and Service Agent modes require their Agent Identity Authority binding and lifecycle checks. A supplied `principal` is comparison-only and a mismatch is rejected. The signing path is not an HTTP signing oracle: callers cannot submit an unsigned decision, arbitrary identity, or policy result for signing. Invalid internal authentication or principal evidence is `401 unauthenticated`; unavailable identity, binding, policy, signing, or revocation dependencies are `503` and fail closed.

## Verification Endpoint

`POST /v1/policy/decisions/verify` is internally authenticated with exactly one valid `X-Aegivela-Internal-Token` header. Its request body is:

```json
{
  "signed_decision": "<compact-jws>",
  "audience": "gateway",
  "tenant_id": "tenant-1",
  "subject_ref": "subject-1",
  "actor_id": "actor-1",
  "action": "invoke",
  "resource_ref": {"kind": "gateway:backend", "id": "backend-1"},
  "scope": ["read"],
  "trace_id": "trace-1",
  "task_ref": "task-1",
  "session_ref": "session-1",
  "policy_version": "policy-v1",
  "lifecycle_epoch": 4
}
```

Unknown fields, trailing content, or an invalid trace ID are `400 invalid_decision_verify_request`. Verification requires exact issuer, audience, tenant, subject, actor, action, resource, policy version, session, task, and lifecycle-epoch bindings. The requested `scope` may be a duplicate-free subset of the signed `effective_scope`; it must not widen that scope. The endpoint returns the verified decision on success. Invalid signatures, malformed artifacts, expired artifacts, or invalid expected bindings are `401 invalid_signed_decision`; a verified artifact that does not match the requested binding or is revoked is `403 decision_denied`; unavailable verification or revocation dependencies are `503 decision_verification_unavailable`. Missing or invalid internal authentication is `401 unauthenticated`.

## Invalidation and Evidence

Verification checks tenant-scoped effective revocation selectors for `decision_jti`, `subject_ref`, `session_ref`, `policy_version`, and a digest of `resource_ref`. Only successful non-revocation results may be cached, and then only until the earlier of artifact expiry and the configured cache lifetime. A missing or failed revocation check is unavailable and fails closed; it never becomes an allow.

Verification emits allowlisted evidence for verified, denied, revoked, and unavailable outcomes. Evidence includes tenant ID, decision digest, decision ID, outcome, action, resource digest, policy version, audience, validated trace ID, and a reason code. It does not include the compact decision token, raw bearer token, credentials, unrestricted arguments, or raw resource payload.

## Consumer Rules

Consumers must fetch or configure only the decision public key, validate all artifact and binding requirements, and fail closed when verification or revocation is unavailable. When a consumer requests an attenuated scope, it must enforce that requested subset rather than the full signed `effective_scope`. An unsigned canonical decision response has no grant or execution authority. A consumer must not override signed principal, audience, action, resource, task, policy, session, or lifecycle fields.
