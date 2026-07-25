# Agent Identity Authority

## Decision

AEGIVELA owns an **Agent Identity Authority** and its security-scoped Agent Identity Registry. This is not a second enterprise human IdP and not a product Agent/Capability Registry.

The Enterprise IdP remains authoritative for human authentication, passwords, MFA, directory records, and SCIM. AEGIVELA consumes its verified human assertions.

## F0 Current Behavior And Limits

F0 implements tenant-isolated Twin and Service Agent registration, immutable authority binding, lifecycle state transitions, PostgreSQL persistence, migration-level audit fields, and an internal-token-protected HTTP API. Activation in F0 is a lifecycle API operation behind the internal PEP; it does not verify workload enrollment or attestation.

F0 does not implement PDP authorization, verified workload enrollment or attestation, short-lived artifact issuance, external evidence or audit export, human authentication, or grant validation. Those capabilities remain future work.

It begins as an AEGIVELA module backed by PostgreSQL, not a separate deployable system. It may later become independently deployable only when its security SLA, tenancy, or scaling requirements require that boundary.

## F0 Persisted Schema

```text
tenant_id, agent_id, agent_class, authority_root,
master_id?, organization_authority_root?, maximum_scopes,
workload_selector, lifecycle_state, lifecycle_epoch,
created_at, updated_at
```

F0 does not persist `owner_ref`, `creation_source`, `autonomy_policy`, credential references, attestation requirements, `revoked_at`, or `policy_version`. Those are future target metadata. The registry contains no Agent prompt, memory, tool configuration, Capability catalog entry, business workflow, or execution history.

## Agent Classes

| Class | Authority root | Cardinality | Required binding |
| --- | --- | --- | --- |
| Twin Agent | Named human | One human may own `0..N` Twins; each Twin has exactly one master | Immutable `master_id` |
| Service Agent | Organization or system | One organization authority root may own `0..N` Service Agents | Immutable `organization_id` or `authority_root` |

Changing a Twin's master or a Service Agent's organization root is not an in-place edit. AEGIVELA revokes the old identity and creates a new agent ID, preserving auditability and preventing an old grant from being reinterpreted under a new authority.

## Future Registration Flows

### Twin Agent Registration

1. A verified human, or an authorized product acting through that human's delegation, requests a Twin registration.
2. AEGIVELA derives `tenant_id` and `master_id` from the verified human identity; neither is accepted as caller-controlled input.
3. The request declares the requested maximum scope, workload selector, autonomy profile, and attestation requirement.
4. A future PDP verifies the requested envelope is a subset of the human's effective authority and any tenant policy.
5. A future authority flow records evidence and returns enrollment requirements.
6. A future verified workload/device completes enrollment and attestation before activation.

### Service Agent Registration

1. A tenant administrator or an authorized deployment controller requests a Service Agent registration.
2. AEGIVELA derives the tenant and organization/system authority root from the authenticated administrator or deployment identity.
3. A future PDP verifies the requested service envelope, workload selector, owner reference, and attestation requirement.
4. A future authority flow activates only after workload enrollment and attestation.

Service Agents cannot declare a human `master_id`. A request that needs to act for a user must use a Twin or a delegated API artifact with explicit user authority.

## Future Authentication and Issuance

The future Agent Identity Authority is an agent authorization issuer, not a human login provider. It will accept proof from a registered workload, such as mTLS, SPIFFE/SVID, device-bound key, or an approved client credential, then issue a short-lived AEGIVELA artifact only when the agent is active and attested.

The F0 HTTP API internal-token boundary is separate from the Agent Identity Authority and is not a human login mechanism or an issuance credential. It is a temporary internal PEP guard for an adapter or proxy that supplies verified request context; F1 OIDC Identity Bridge replaces that adapter.

For Twin Agents, every issued artifact contains the server-injected immutable `master_id`. For Service Agents, it contains the organization/system authority root and no user identity. Both artifact types bind agent ID, workload proof, lifecycle epoch, audience, scope, expiry, revocation handle, and parent authority reference where delegated.

## Lifecycle and Product Coordination

F0 exposes registration and lifecycle requests through its internal PEP. Products cannot activate an agent by updating their own product record alone. F0 activation does not verify runtime attestation. Future contracts may accept attestation reports; a product configuration may remain present after suspension, but future grant and PEP enforcement remain out of F0 scope.

Future lifecycle policy may cascade a human's security suspension to every Twin bound to that human. It does not cascade a human suspension to Service Agents because they are not human-delegated; organization policy controls their lifecycle.

## Future Security Gates

- Future registration fails closed for unverified requester identity, tenant mismatch, scope escalation, missing workload selector, or invalid attestation policy.
- The F0 HTTP API fails closed unless its internal adapter token and each trusted context header are present exactly once and valid; it does not authenticate a human user.
- Future issuance fails closed for `created`, `suspended`, or `revoked` lifecycle states; stale lifecycle epoch; revoked credential; or unavailable revocation check.
- Future credentials are proof-bound and short-lived. Raw private keys, bearer tokens, and credentials are not stored in audit records.
- Future registration, enrollment, activation, issuance, denial, suspension, and revocation operations emit redacted evidence with actor, authority root, agent ID, policy version, and trace ID; F0 retains only migration-level audit fields.
