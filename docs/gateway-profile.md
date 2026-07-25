# AEGIVELA Agent Gateway Profile

## Purpose

This profile applies AEGIVELA to a general agent access-control path, independently of MODUREGIS:

```text
Skill / Tool on desktop or Agent Host
  -> Client Gateway PEP :10255
  -> Server Gateway PEP :10256
  -> Backend API / LLM / Business Service
```

The existing `AXIS-gateway` repository owns the Go proxy, `CONNECT` tunneling, certificate MITM, request forwarding, credential injection, and gateway deployment. AEGIVELA owns the portable security contracts and services consumed at those PEPs. This avoids turning AEGIVELA into another proxy implementation while allowing the same PDP, grant, approval, revocation, and evidence model to secure other agent runtimes.

## Current Reference Surface

| Runtime | Placement | Existing foundation | AEGIVELA integration direction |
| --- | --- | --- | --- |
| Client Gateway | Desktop, developer machine, or Agent Host; default `:10255` | HTTP forwarding, CONNECT, local MITM foundation, short-TTL policy cache, policy-driven intercept, approval wait | Device/workload identity, short-lived decisions, cache invalidation, local tool PEP |
| Server Gateway | Datacenter, container, or cloud; default `:10256` | PostgreSQL resolver, centralized policy route, header/query injection, AWS SigV4 finalizer, audit/telemetry and approval foundations | PDP adapter, credential-injection authorization, approval coordination, central evidence export |

The present gateway route `POST /v1/policy/connect` is a connection-level foundation: it receives `agent_token`, target host, scheme, and path. That is insufficient for semantic control of an agent action.

## Required Semantic Authorization Context

Before a protected tool call, the gateway PEP must assemble a signed attribution context:

```text
human principal (h)
  -> delegated agent principal (a)
  -> tool or skill principal (k: tool_id, skill_hash, implementation digest)
  -> backend service principal (s)
```

The resulting decision request binds:

```text
tenant_id, device/workload identity, h/a/k/s references, execution_id,
tool_id, skill_hash, capability or task reference, target service, method,
path, declared argument schema/classification, requested credential scope,
risk signals, policy version, and trace_id
```

When the caller is a Twin Agent, this context is resolved through the EASEF-IAM profile: `a` has one immutable `master_id`, and the local device or runtime proves `w` through enrollment and workload attestation. For a Service Agent, `a` instead has an organization authority root and no `master_id`. The PEP rejects a missing required authority binding, a workload outside the agent's permitted selectors, a stale lifecycle epoch, or a child grant broader than its parent authority.

Raw arguments, credentials, and tokens are not required in the decision or audit record. The PEP sends only the minimum value needed for enforcement, such as an argument digest, type, sensitivity label, or policy-relevant field projection.

The PDP enforces:

```text
authority(k) <= authority(a) <= authority(h)
requested_scope subset-of delegated_scope
backend_request_authority = authority(k) intersection authority(s)
every protected request has non-empty tool/skill attribution
```

For Service Agents, the first relation is replaced by `authority(k) <= authority(a) <= authority(organization_root)`. The authority mode determines which relation is valid; the PEP never substitutes one for the other.

This closes the current connection-level gap: host/path allowlisting alone cannot show which tool initiated an operation or whether its arguments are permitted.

## Enforcement Semantics

| PDP outcome | Client Gateway behavior | Server Gateway behavior |
| --- | --- | --- |
| allow | Forward only within the returned obligations | Recheck audience, scope, target, expiry, and revocation before egress/injection |
| deny or revoked | Block locally and emit redacted evidence | Block egress and emit redacted evidence |
| approval-required | Hold within bounded timeout and create approval request | Coordinate approval and issue a new bound decision after approval |
| rate-limit | Apply returned limit without widening scope | Apply shared limit where central coordination is required |
| unavailable | Deny, except an explicitly offline-eligible cached decision | Deny; do not inject credentials |

Local degraded mode is not a bypass. Only a non-expired, locally cached decision marked offline-eligible may be used, and it cannot permit credential injection, high-risk operations, new targets, or a broader scope than the original decision. Revocation and policy invalidation clear matching cache entries through a client-initiated HTTPS channel, optionally upgraded to WebSocket or SSE.

For high-autonomy actions, a pre-authorization token may replace a per-call approval only when it is bound to the same agent, master, target scope, audience, task/resource reference, and validity window. It grants no scope not already delegated to the agent.

## Credential and Inspection Controls

Credential injection is an effect, not proof of authorization. The PEP may inject a header, query value, request-body value, or AWS SigV4 signature only when the AEGIVELA decision explicitly names the target, credential class, scope, and expiry. Secrets remain in the approved vault or server-side secret provider and are never included in grants, policy telemetry, or audit evidence.

MITM is enabled only by policy for target scopes that require inspection or injection. The PEP records the intercept reason, but audit records retain only redacted request metadata and policy-relevant digests.

## Delivery Gates

1. Device and workload enrollment binds a client installation to an approved workload identity and agent lifecycle record.
2. Extend the gateway policy contract from `connect` to tool/skill attribution and policy-relevant argument context.
3. Enforce `tool_id` and `skill_hash` at the client PEP; reject absent or digest-mismatched attribution.
4. Bind every injected credential and backend egress action to a short-lived AEGIVELA decision or grant.
5. Propagate revocation and policy invalidation to local caches within the agreed SLO.
6. Export correlated allow, deny, approval, injection, and revocation evidence to Argus/SIEM/Harmovela.

No Gateway profile is production-ready until all six gates pass in an end-to-end agent-to-backend test.
