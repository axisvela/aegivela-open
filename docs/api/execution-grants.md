# Execution Grant Issuance Contract

## Signed Decision Prerequisite

`POST /v1/grants/issue` issues an execution grant only from a verified signed policy decision with `outcome: "allow"` that was produced from a trusted principal. A bearer token is resolved only to confirm that principal's identity and binding; it is not grant authority by itself. It does not accept a raw policy decision response, a caller-supplied principal, or caller overrides for audience, action, resource, task, policy version, session, or lifecycle epoch.

The endpoint is internally authenticated with exactly one valid `X-Aegivela-Internal-Token` header. Its body accepts only:

```json
{
  "bearer_token": "<identity-bearer-token>",
  "signed_allow_decision": "<compact-EdDSA-JWS>",
  "scope": ["read"],
  "trace_id": "trace-1"
}
```

`bearer_token`, `signed_allow_decision`, and a valid `trace_id` are required. `scope` is optional: omitted scope uses the signed `effective_scope`; supplied scope must be a duplicate-free subset of that scope. The issued grant carries the requested attenuated scope, which consumers must enforce instead of the full signed `effective_scope`. The endpoint rejects unknown fields and trailing content with `400 invalid_grant_issue_request`. Therefore unsigned decision JSON and every principal or signed-binding override are rejected before grant issuance.

## Issuance Checks

The service first authenticates the signed artifact, then resolves the bearer token for the signed audience. It verifies exact tenant, subject, actor, audience, action, resource, task, policy version, session, and lifecycle-epoch bindings between the trusted principal and signed decision. It also checks decision invalidation. Only after the decision verifies and is `allow` does it issue and persist the grant. This establishes the authority lineage `trusted principal -> canonical allow decision -> execution grant`; token exchange can only further attenuate the verified decision or grant.

The grant expiry is the earliest of 15 minutes from issuance, the resolved principal expiry, and the signed-decision expiry. A grant is therefore never extended beyond the identity or signed policy authorization that produced it.

## Failure Semantics

| Condition | Response |
| --- | --- |
| Missing or invalid internal token, bearer identity, or signed artifact | `401 unauthenticated` or `401 invalid_signed_decision` as applicable. |
| Malformed request or attempted override | `400 invalid_grant_issue_request`. |
| Non-allow decision, binding mismatch, scope widening, or invalidated decision | `403 decision_denied`. |
| Identity, decision verification, revocation, or issuance dependency unavailable | `503 identity_unavailable`, `503 decision_verification_unavailable`, or `503 grant_issuance_unavailable`. |

All unavailable dependencies fail closed. See the [Signed Policy Decision contract](signed-policy-decisions.md) for JWS claims, decision key discovery, binding verification, invalidation, and redacted evidence requirements.
