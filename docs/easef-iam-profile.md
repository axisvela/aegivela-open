# EASEF-IAM Dual-Identity Profile

## Scope

This AEGIVELA profile applies when an agent acts under a named human's delegated authority. It does not replace enterprise OIDC, SAML, directory, or workload identity; it adds agent-to-human binding and lifecycle rules those systems do not provide alone.

```text
h: verified human principal          Enterprise IdP authentication and approval authority
a: registered agent principal        Delegated authority and lifecycle subject
w: verified workload or device       Runtime attestation and token-proof subject
```

For a tool-mediated backend call, the Gateway profile adds the tool/skill and backend dimensions: `h -> a -> k -> s`, while `w` proves the workload running `a` or `k`.

## Authoritative Agent Record

For every human-delegated Twin Agent, the Agent Identity Authority stores or resolves:

```text
agent_id, tenant_id, master_id, maximum_scopes, autonomy_policy,
lifecycle_state, permitted_workload_selectors, attestation_requirements,
created_at, policy_version
```

AEGIVELA injects `master_id` from this record; it never accepts it from an agent request or caller-provided JWT claim. The binding is immutable for an agent identifier: reassignment revokes the old agent and creates a new identity record. One human may own zero or more Twin Agents; each Twin has exactly one master. An organization-owned Service Agent uses the separate organization authority-root model described in [authorization modes](authorization-modes.md); it must not silently claim a human `master_id`.

## Formal Security Properties

| Property | Invariant | Enforcement point |
| --- | --- | --- |
| P1 Principal accountability | Every valid agent artifact has exactly one `master_id`, equal to the agent's immutable registered master. | Identity Bridge issuance and PEP validation |
| P2 No privilege escalation | `artifact.scopes subset-of agent.maximum_scopes subset-of master.effective_scopes`. | Registration, issuance, PDP, and PEP |
| P3 Delegation soundness | A child artifact has scopes, audience, expiry, and task binding no broader than its parent authority. | Token exchange and grant issuance |
| P4 Revocation safety | A suspended or revoked agent has no artifact accepted by a PEP. | Lifecycle transaction, Revocation Service, and PEP |

These properties require mandatory master binding, scope-subset validation, monotonic exchange, and lifecycle-aware revocation. Product policy may narrow authority further but cannot widen these structural limits.

## Artifact Contract

All access tokens, pre-authorization tokens, and execution grants for this profile carry these protected claims or equivalent signed fields:

```text
jti, typ, tenant_id, sub=agent_id, master_id, workload_id or cnf,
scope, audience, issued_at, expires_at, policy_version,
parent_authority_ref, lifecycle_epoch, revocation_handle, trace_id
```

`workload_id` is present when workload binding is required. `cnf` may bind proof-of-possession material instead of exposing a stable workload identifier. A PEP validates signature, issuer, audience, expiry, tenant, lifecycle epoch, revocation handle, and resource-specific bindings before allowing an effect.

## Lifecycle

| State | Issue new artifacts | Existing artifact acceptance | Transition rule |
| --- | --- | --- | --- |
| `created` | No | No | Attestation and scope validation are incomplete |
| `active` | Yes, within delegated scope | Per policy | Normal operation |
| `suspended` | No | No | Reactivation requires explicit transition and fresh issuance |
| `revoked` | No | No | Terminal; identifier is never reused |

Suspension changes the state, increments the lifecycle epoch, and creates revocation records for affected artifacts before reporting success. Suspension or deactivation of a master triggers suspension of its delegated agents. PEPs receive invalidation events and reject artifacts with stale lifecycle epochs or revocation state.

Distributed propagation is not assumed to be instantaneous. A strict PEP performs online revocation/lifecycle validation before each high-risk effect and has no post-suspension acceptance window. A bounded-cache PEP may use only an explicitly offline-eligible, unexpired decision; it cannot permit credential injection, destructive actions, new targets, or scope expansion while disconnected.

## Token Exchange and Pre-Authorization

Token exchange only attenuates authority:

```text
child.scope       subset-of parent.scope
child.audience    subset-of parent.audience
child.expires_at  <= parent.expires_at
child.task_binding is equal to or narrower than parent.task_binding
```

Pre-authorization is a time-bounded approval window for autonomous use of already delegated authority, never a privilege elevation mechanism. A verified `master_id` session requests it; the Approval Gateway binds the result to agent, scope, audience, task/resource reference, reason/evidence, `valid_until`, and policy version. The resulting token is revocable and obeys the same subset rules.

## Required Evidence

Every issuance, exchange, approval, denial, suspension, revocation, and PEP decision emits redacted evidence containing its token or decision ID, agent, master, workload reference where applicable, action/resource, policy version, trace ID, outcome, and evidence references. Raw credentials, secrets, and unrestricted arguments are excluded.
