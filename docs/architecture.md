# AEGIVELA Architecture Baseline

AEGIVELA is a reusable security control plane and SDK family. It is not a MODUREGIS subsystem and does not require a Capability Registry to operate. AEGIVELA is integrated by a product-specific adapter and enforced where that product's trust boundary requires a PEP.

## Authority Model

AEGIVELA is the security authority for identity binding, authorization, approval, delegated scope, revocation, attestation policy, and associated security evidence. A policy decision authorizes a specific action; it never transfers Registry, Catalog, or workflow ownership to AEGIVELA.

| Object or decision | Authoritative owner | AEGIVELA responsibility |
| --- | --- | --- |
| User authentication, MFA, directory, SCIM | Enterprise IdP | Validate and map identity assertions only |
| Capability and version lifecycle | Moduregis Registry | Evaluate requested lifecycle action in the Moduregis profile |
| Gateway transport, MITM, injection, and backend routing | Gateway runtime | Supply policy/identity/grant contracts and PEP SDK |
| Agent, tool, backend, and data resource model | Integrating product | Evaluate an integrating product's declared action and resource |
| Execution grant and delegated authority | AEGIVELA | Issue or validate short-lived, audience- and instance-bound grants |
| Immutable control-plane audit index | Moduregis | Supply decision, policy version, and evidence references |
| Raw security logs and SIEM export | Evidence producer / SIEM | Produce correlation-safe evidence references |

## Core Components

### Identity Bridge

The Identity Bridge trusts configured enterprise issuers and validates OIDC JWTs through issuer discovery and JWKS. It does not perform opaque-token introspection or parse SAML assertions. Directory and SCIM data may enrich subject attributes but cannot override a verified token's issuer, subject, audience, expiry, or tenant mapping. See the [Identity Bridge integration guide](api/identity-bridge.md).

The F0 bridge returns only this verified principal:

```text
bridge_principal = {
  tenant_id, subject_ref, actor_id, issuer, client_id,
  identity_assurance?, session_ref?, issued_at, expires_at
}
```

`tenant_id`, `subject_ref`, `actor_id`, `client_id`, optional assurance/session references, and token times are derived only from verified claims and configured mapping rules. A caller must never supply them as authorization overrides.

Later authorization processing may enrich the bridge principal with `authorization_mode`, `authority_root`, `workload_id`, `agent_id`, immutable `master_id`, `organization_id`, `lifecycle_epoch`, and `delegation_chain`. Those fields are not returned by the F0 bridge and must come from the applicable binding records and authorization context. For Twin Agents, immutable `master_id` and optional workload proof are resolved from the EASEF-IAM binding record. For Service Agents, `agent_id` and organization authority root are resolved without manufacturing a human master. See [authorization modes](authorization-modes.md) and the [Agent Identity Authority](agent-identity-authority.md).

### Policy Decision Point

The PDP evaluates a versioned action and resource contract. `POST /v1/policy/decisions/evaluate` and `POST /v1/policy/decisions/sign` authenticate their internal caller, then use `trustedprincipal` to resolve the normalized principal before evaluation. The resolver accepts only the mode-appropriate trusted source: Identity Bridge for `human_web`, workload assertion plus active binding for `system_api` and `delegated_api`, and Agent Identity Authority bindings and lifecycle state for Twin and Service Agents. Request principals are comparison-only and a mismatch is rejected. The Moduregis profile uses `capability:read`, `capability:publish`, `adapter:activate`, and `capability:invoke`; the Gateway profile uses `connection:open`, `tool:invoke`, `backend:request`, and `credential:inject`. Its input combines the resolved principal with resource identifiers, requested scope, risk context, approval state, and revocation state. Its output is a versioned decision, not a mutable authorization session.

Policy evaluation is fail-closed. An unavailable issuer, unknown policy version, expired credential, missing tenant binding, or revocation-check failure must not be interpreted as allow.

Workload bindings are tenant-scoped, append-only records that bind a client and workload to an organization authority root. A workload assertion is a separately signed, audience-bound, five-minute proof of an active binding; it is verified with its own JWKS and subject to binding and JTI revocation. A forged or invalid trusted-principal source is unauthenticated (`401`); an unavailable identity, binding, policy, or revocation dependency is unavailable (`503`) and must deny.

Policy Decision v1alpha1 is immutable and accepts only legacy resource kind/reference pairs. v1alpha2 is a separate, explicitly routed contract that can carry legacy resources during migration or an all-or-none descriptor-bound structured resource. For structured resources, the PDP validates the registered descriptor and returns a canonical `structured_resource_digest` bound to the resource identity and validated attributes. A partial resource, unknown descriptor, invalid attributes, or unavailable descriptor validation fails closed.

Signed v1alpha2 decisions and execution grants retain a canonical digest for every resource: structured resources use their descriptor-bound digest and legacy resources use the domain-separated `legacy-v1alpha2` kind/reference digest. Before execution, a v1alpha2 PEP recomputes the applicable digest from its trusted current resource instance; an absent, unexpected, or unequal digest rejects execution. A v1alpha1 artifact cannot authorize a v1alpha2 structured-resource PEP. Evidence includes only the digest and descriptor identifiers, not raw structured attributes.

### Web Resource Authorization

Web Resource Authorization is a complete generic capability for tenant-scoped UI resource trees and short-lived permission projections. It resolves a trusted internal principal into a redacted-evidence-backed browser projection with deny precedence and visibility closure. The projection is UI-only: every product backend remains responsible for server-side PDP/PEP enforcement and must fail closed rather than accepting a browser permission map. Integrating products own their resource semantics and cache invalidation on projection expiry, policy-version changes, or resource-tree changes. The MODUREGIS-specific adapter and its Console/OIDC/PKCE integration remain pending.

### Approval and Scope Engine

Approval binds a human decision to the tenant, actor, action, immutable resource version, scope, reason, evidence references, and expiry. Approvals may be issued with an optional signed artifact (`approval/v1alpha2`, fifth signing domain per [ADR-0006](adr/0006-approval-record-ownership.md)) so product PEPs verify bindings offline while grant verification still enforces the `approval_jti` revocation selector. Scope attenuation produces only subsets of the parent authority. Delegation must preserve non-amplification:

```text
scope(child) subset-of scope(parent)
Twin:    authority(tool) <= authority(agent) <= authority(human)
Service: authority(tool) <= authority(agent) <= authority(organization_root)
```

Child artifacts must also have an audience, expiry, and task/resource binding no broader than their parent artifact.

Authority has one lineage: a trusted principal produces a canonical policy decision; only a verified `allow` decision may produce an execution grant; token exchange may only attenuate that verified parent grant or canonical allow decision. A raw bearer artifact establishes no delegated authority by itself, and every child preserves the parent's tenant, audience, scope, expiry, action, resource, task, and lifecycle constraints.

### Revocation and Evidence

Revocation supports subject/session, execution grant, Capability version, implementation, and policy-version invalidation. Rechecks occur at three SLO classes ([ADR-0008](adr/0008-revocation-propagation-slo.md)): `pre_dispatch` and `continuation` are authoritative cache-free checks, while `connection` may serve a bounded-TTL negative cache. The API and every PEP check revocation before a privileged action; long-running operations recheck at continuation boundaries. Evidence records retain references, digests, and correlation identifiers, not raw secrets or unredacted payloads.

### Attestation Service

The Attestation Service verifies artifact digests, signer identity, runtime/workload evidence, and environment assertions. It does not trust an agent, tool, gateway, or executor to self-assign an assurance level. Its results are evidence references that policy may require for grant issuance, credential injection, deployment activation, or high-risk tool invocation.

### Security Evidence SDK

The SDK emits a stable, redacted evidence envelope for Argus, SIEM, Harmovela, and a consuming product's Audit Index. It carries identity and decision references, action/resource references, trace/correlation identifiers, policy version, outcome, and hashes or classifications of protected inputs. It never requires raw secrets, bearer tokens, or unrestricted arguments in audit output.

## Integration Profiles

### MODUREGIS Capability Governance

Moduregis owns Capability lifecycle, Registry, Catalog, product workflow, and its append-only Audit Index. AEGIVELA returns policy decisions and evidence references through the documented adapter. The Console redirects to the enterprise IdP using authorization-code flow with PKCE; it does not manage passwords. See [the Moduregis contract](moduregis-contract.md).

### Web and API Gateway Access

Web applications and API Gateways consume AEGIVELA as OIDC relying parties and PEPs. The Gateway distinguishes human web access, system API access, and system-on-behalf-of-user access instead of using an undifferentiated service account. AEGIVELA publishes a versioned Web/API PEP contract and Go reference SDK; consuming products own OIDC authorization-code-with-PKCE, session management, and refresh-token handling. See [authorization modes](authorization-modes.md) and the [Web/API PEP integration guide](api/web-api-pep.md).

### Agent Gateway Runtime Access Control

The Gateway profile places PEPs around agent-originated traffic. The existing `AXIS-gateway` client and server gateway processes remain independent runtime components. AEGIVELA's adapter evolves their policy API from connection routing toward signed, tool-level and argument-aware enforcement. See [the Gateway profile](gateway-profile.md).

## Trust Boundaries

```text
Browser -> Enterprise IdP -> Product UI -> AEGIVELA PDP -> Product PEP
Agent/Skill/Tool -> Client Gateway PEP -> Server Gateway PEP -> API/LLM/BizService
Moduregis Broker -> AEGIVELA execution grant -> PRAXOVELA or RHEOVELA PEP
```

Products consume AEGIVELA's verified principal or signed grant rather than inventing another security authority. PEPs fail closed if a required decision, grant validation, or revocation check is unavailable. A bounded local decision cache is permitted only when the policy explicitly authorizes degraded mode and the decision has not expired.

## Non-Goals

- A username/password database, MFA challenge service, or general browser session center.
- A duplicate tenant, Capability lifecycle, or generic product database.
- A Catalog visibility rule used as a substitute for PDP enforcement.
- Long-lived executor credentials or unbound bearer grants.
- A central event broker, planner, workflow engine, Registry, proxy, or gateway runtime.
