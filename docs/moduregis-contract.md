# AEGIVELA-Moduregis Contract v0.1

## Purpose

This contract replaces the temporary `DenyAll` adapter only after both sides implement and verify it. It deliberately matches Moduregis's existing `Authorizer` and `ActivationAuthorizer` ports without moving product governance into AEGIVELA.

## Caller Requirements

Moduregis sends the bearer artifact unchanged to AEGIVELA with a required authorization requirement:

```json
{
  "action": "capability:read",
  "namespace": "engineering"
}
```

The bearer artifact is an enterprise IdP access token or an AEGIVELA-issued delegated grant. AEGIVELA validates its signature or introspects it against an allowlisted issuer, verifies issuer, audience, expiry, and tenant mapping, then derives the principal. Moduregis must not accept caller-provided `tenant_id`, `actor_id`, or `policy_version` as trusted input.

## Decision Response

The HTTP and gRPC APIs use one semantic response model. Transport-specific status codes cannot change authorization meaning.

```json
{
  "decision_id": "dec_...",
  "outcome": "allow",
  "principal": {
    "tenant_id": "tenant_...",
    "actor_id": "agent_...",
    "agent_id": "agent_...",
    "master_id": "user_...",
    "workload_id": "spiffe://example.org/agent-runtime/worker-1",
    "subject_ref": "https://issuer.example/agents/456"
  },
  "policy_version": "2026-07-15.1",
  "evidence_refs": ["aegivela://evidence/dec_..."],
  "expires_at": "2026-07-15T12:05:00Z"
}
```

Valid outcomes are `allow`, `deny`, `approval_required`, and `revoked`. Only `allow` can satisfy the Moduregis port. `approval_required` is a governed state, not a successful authorization; Moduregis presents its approval workflow and resubmits with the resulting approval reference. When the approval reference is a signed approval artifact (`approval/v1alpha2`), it carries cryptographically bound tenant, action, immutable resource version or descriptor-bound structured digest, scope, and expiry; Moduregis must treat verified artifact bindings as the only admissible approval parameters and must fail closed when verification or the authoritative `approval_jti` revocation check is unavailable. See [ADR-0006](adr/0006-approval-record-ownership.md). A malformed, unavailable, or expired response is unavailable and therefore fail-closed.

For the current Catalog port, `allow` requires non-empty `tenant_id`, `actor_id`, and `policy_version`. The adapter maps `deny`, `approval_required`, and `revoked` to `ErrDenied`; it maps transport and service failures to `ErrUnavailable`.

Before enabling the adapter, Moduregis must extend its port with `ErrUnauthenticated`. Invalid, expired, wrong-issuer, and wrong-audience bearer artifacts map to that error and return `401`; they are not AEGIVELA outages. `DenyAll` remains mapped to `ErrUnavailable` and `503 authorization_unconfigured`.

## Required Moduregis Port Changes

The current `Principal` carries only `tenant_id`, `actor_id`, and `policy_version`. Before the production adapter is enabled, it must also carry `decision_id` and `evidence_refs`; the API must reject an allow result with an empty `policy_version`. For an agent subject, it also carries `agent_id`, agent class, workload reference or proof, lifecycle epoch, and grant/decision expiry. A Twin Agent additionally carries immutable `master_id`; a Service Agent carries its organization authority root and no synthetic master. `ActivationDecision` likewise requires `decision_id`. These additive fields let Moduregis persist the exact AEGIVELA decision in its Audit Index without treating an uncorrelated policy result as authoritative.

`actor_id` remains the direct caller executing the request. For a Twin Agent, `master_id` is the accountable human authority and must be derived from the AEGIVELA binding record. For a Service Agent, the organization authority root is the accountable authority. Moduregis can apply a stricter product policy, but it must reject an agent decision whose required authority binding, lifecycle epoch, tenant, or grant expiry is invalid.

## Adapter Activation Decision

`adapter:activate` is bound to the exact immutable adapter ID and version:

```json
{
  "tenant_id": "tenant_...",
  "namespace": "engineering",
  "adapter_id": "praxovela.execution",
  "adapter_version": "1.2.0",
  "actor_id": "user_...",
  "trace_id": "trace_...",
  "authorization_artifact": "bearer-or-delegated-grant",
  "evidence_refs": ["moduregis://attestations/att_..."]
}
```

AEGIVELA returns `allowed`, `policy_version`, and evidence references. The decision must verify that the actor derived from the artifact matches the requested actor, that the resource version is immutable, that attestation evidence is acceptable, and that no applicable revocation exists. An AEGIVELA outage must leave the adapter inactive.

## Execution Grants

For invocation, AEGIVELA issues a short-lived signed grant with at least these bindings:

```text
tenant_id, actor_id, agent_id?, agent_class?, master_id?, organization_authority_root?,
workload_id or cnf?, lifecycle_epoch?,
capability_id, capability_version, implementation_id, execution_id, audience,
scope, policy_version, issued_at, expires_at, parent_authority_ref?,
delegation_chain, revocation_handle, evidence_refs
```

The Broker and executor PEP reject a grant when any binding differs from the invocation, its audience is wrong, it is expired, or its revocation handle is invalid. The `audience` field must be a registered runtime audience (for example `moduregis:praxovela.execution`) per [ADR-0007](adr/0007-grant-audience-registry.md). The Governor must perform a `pre_dispatch` revocation check before dispatch and `continuation` rechecks at declared boundaries via `POST /v1/revocations/check` ([ADR-0008](adr/0008-revocation-propagation-slo.md)), halting on any `503`. The grant contains no long-lived secret and never authorizes a different Capability version.

## Audit Handoff

For login-derived access, authorization, denial, approval, activation, and revocation, AEGIVELA provides `decision_id`, `policy_version`, `trace_id`, `actor_id`, resource reference, outcome, and evidence references. Moduregis writes these into its append-only Audit Index. Raw IdP assertions, credentials, payloads, and secrets are not copied into the index.

## Compatibility and Rollout

1. Publish versioned OpenAPI and protobuf definitions from the AEGIVELA `contracts` package.
2. Add a Moduregis HTTP or gRPC adapter behind explicit configuration; retain `DenyAll` as the default.
3. Run contract tests against allow, deny, invalid token, unavailable PDP, tenant mismatch, expired grant, approval-required, and revoked-token cases.
4. Enable only a non-production tenant first; require an audit record for every result.
5. Promote after denial, revocation, and activation gates meet the defined evidence criteria.
