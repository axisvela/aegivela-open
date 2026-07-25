# AEGIVELA Delivery Roadmap

## Release Principle

Each release is a vertically tested security boundary. A service is not complete because an endpoint exists: it must reject invalid authority, emit correlatable evidence, and preserve tenant isolation under failure.

F0 establishes reusable AEGIVELA contracts. The MODUREGIS Capability Governance profile and the Agent Gateway Runtime Access Control profile then progress independently; neither is a prerequisite for the other. Shared identity, grant, revocation, attestation, and evidence contracts must remain profile-neutral.

The OSS/EE boundary is documented in [OSS/EE Boundary](oss-ee-boundary.md). Stages marked **[OSS]** ship in the open-core edition; stages marked **[EE]** require enterprise infrastructure and ship in the Enterprise Edition.

## Edition Legend

- **[OSS]** — Profile-neutral security primitives and versioned integration contracts
- **[EE]** — Enterprise identity federation, continuous evaluation, operational controls

## F0: Contract and Trust Foundation [OSS]

**Deliverables**

- Implemented: canonical Policy Decision JSON Schema, OpenAPI and protobuf mappings, protobuf compatibility baseline, positive/negative/attenuation fixtures, and static F0 reference evaluator.
- Versioned principal, policy decision, approval, revocation, execution-grant, and security-evidence schemas.
- EASEF-IAM binding record, lifecycle epoch, dual-identity artifact claims, and P1-P4 conformance properties.
- Agent Identity Authority records for Twin and Service Agent classes, including immutable authority-root bindings.
- Implemented: Identity Bridge validation of configured enterprise OIDC JWT issuers, audiences, signatures, time claims, and claim-to-principal mappings through cached discovery and JWKS.
- OpenAPI and protobuf definitions for authorization, agent lifecycle, token exchange/pre-authorization, revocation, and adapter activation.
- Policy SDK/PEP error taxonomy: unauthenticated, denied, and unavailable.
- Moduregis port additions for `decision_id`, `evidence_refs`, and non-empty `policy_version` validation.
- Threat model and ADRs for token validation, tenant derivation, audit redaction, and key rotation.

The Policy Decision artifacts establish the F0 request, response, mode, attenuation, outcome, and fail-closed transport semantics only. F0 also delivers internal Identity Bridge JWT validation; product-specific browser/OIDC relying-party flows, optional opaque-token introspection, and external adapters remain future work. Real PDP policy storage, grants, approvals, revocation, attestation, lifecycle resolution, and production evidence persistence remain deferred dependencies.

Generic Web Resource Authorization is implemented: versioned resource/rule/projection contracts, tenant-isolated immutable resource and policy storage, short-lived deny-first UI projections, authenticated internal management APIs, and redacted projection evidence are complete. The MODUREGIS adapter, its Console/OIDC/PKCE integration, and MODUREGIS-specific resource mapping remain pending.

**Exit gate**

- The Moduregis port maps only a validated allow decision to a non-empty tenant and actor principal.
- Invalid, expired, wrong-audience, unknown-issuer, and tenant-confused tokens fail closed.
- A configured adapter returns `401` for invalid credentials, `403` for policy denial, and `503` only for an unavailable authorization dependency.
- Contract compatibility tests run in both repositories.
- Lifecycle and token-exchange tests prove immutable master binding, scope/audience/expiry attenuation, and post-suspension rejection.

## F1A: MODUREGIS Integration Readiness — `capability:*` Adapter Contract [OSS]

> MODUREGIS owns its Console, OIDC/PKCE integration, browser session management, Capability lifecycle, Catalog, Registry, workflow, and Audit Index. AEGIVELA supplies the versioned adapter contract, action/resource mappings, and cross-repository conformance fixtures.

**Deliverables**

- Versioned MODUREGIS adapter contract with `capability:read`, `capability:publish`, `adapter:activate`, and `capability:invoke` action/resource mappings.
- PDP policy and evidence fields required for MODUREGIS consumption.
- Explicit adapter selection semantics; `DenyAll` remains the no-configuration fallback.
- Cross-repository fixture suite proving the adapter maps only a validated allow decision to a non-empty tenant/actor principal.

**Exit gate**

- Missing adapter configuration returns MODUREGIS `503 authorization_unconfigured`; a policy denial returns `403 authorization_denied` via the adapter.
- An unconfigured, invalid, or cross-tenant decision fails closed at the adapter boundary.

**External integration dependencies (MODUREGIS):** Console OIDC authorization-code with PKCE, browser session handling, token refresh, product audit index, and deployment release acceptance.

## F1C: Web and API Gateway Authorization — Published PEP Contract [OSS]

**Deliverables**

- Versioned PEP contract (`contracts/webapi/v1alpha1`) with OpenAPI, JSON Schema, fixtures, and compatibility baseline.
- Go reference middleware and PEP SDK (`backend/pepsdk`) covering `human_web`, `system_api`, and `delegated_api` modes.
- Policy Decision v1alpha3 verified parent execution grant enforcement for `delegated_api`.
- Integration test suite: human success, system success, delegated success, mode-token substitution rejection, parent grant verification, and unavailable-dependency fail-closed tests.
- Integration guide (`docs/api/web-api-pep.md`) describing route configuration, credential headers, error mapping, and product ownership of OIDC/PKCE sessions.

**Exit gate**

- A system token cannot access a user-delegated endpoint, and a user token cannot satisfy a system-only endpoint.
- Effective delegated scope, audience, expiry, action, resource, and task binding are the intersection of the verified parent grant and policy.
- Invalid workload evidence for each mode returns `401`; a verified parent grant whose bindings do not attenuate returns `403`; unavailable identity, binding, grant verification, revocation, or PDP dependencies return `503` and fail closed.

## F1B: Agent Gateway Identity and Connection Enforcement [OSS]

**Deliverables**

- Client enrollment and device/workload identity binding for the Client Gateway.
- Immutable human-to-agent binding records and master-suspension cascade handling.
- Service Agent registration with organization authority-root binding and no synthetic human master.
- Trusted-principal resolution and five-mode negative tests across `human_web`, `system_api`, `delegated_api`, Twin Agent, and Service Agent authority sources.
- Workload binding, assertion, JWKS, and binding/JTI revocation checks at the gateway identity boundary.
- AEGIVELA adapter for the existing gateway `POST /v1/policy/connect` foundation.
- Short-lived, bounded-TTL connection decisions with explicit degraded-mode eligibility.
- Server Gateway PDP, approval, credential-injection, and evidence-export integration.
- Revocation and policy-version invalidation delivery to Client Gateway caches.

**Exit gate**

- An enrolled agent can access only an authorized target through the Client and Server Gateway PEPs.
- Expired, revoked, unavailable, or non-offline-eligible decisions block rather than creating a degraded-mode bypass.
- Allow, block, approval, injection, and cache invalidation events are correlated in security evidence.
- All five authority modes reject missing, forged, mismatched, expired, revoked, and cross-tenant principal evidence; resolver, binding, policy, and revocation outages return `503` and fail closed.

## F2A: MODUREGIS Integration Readiness — Approval and Publication Contracts [OSS]

> MODUREGIS owns publication workflow, approval UX, audit indexing, and console behavior. AEGIVELA supplies adapter contracts, policy/approval evidence fields, and cross-repository conformance. The underlying approval record and artifact mechanisms are profile-neutral ([ADR-0006](adr/0006-approval-record-ownership.md)); MODUREGIS publication is their first conformance profile.

**Deliverables**

- PDP actions for `capability:publish` and approval-required outcomes expressed in the versioned adapter contract.
- Approval binding evidence fields: human decision, reason, resource digest/version, scope, expiry, and evidence references.
- Signed approval artifact contract (`contracts/approval/v1alpha2`): Ed25519 JWS artifacts binding tenant, action, immutable resource version or descriptor-bound structured digest, scope, and expiry, verified offline by product PEPs and never bypassing the authoritative `approval_jti` revocation check.
- Scope attenuation and delegated-grant contracts for publisher and agent actions.
- Cross-repository fixture suite proving publisher cannot write outside its tenant, namespace, scope, or approval window, including approval-laundering negatives: version substitution, namespace substitution, scope widening, and expired approval window replay.

**Exit gate**

- The adapter rejects a publish outside its documented tenant, namespace, scope, or approval window.
- A resubmitted mutation whose parameters do not match or narrow the verified approval artifact bindings is rejected; an unavailable approval verification or revocation dependency fails closed.
- Cross-repository audit correlation contracts are published.

**External integration dependencies (MODUREGIS):** Console refresh-token and enterprise-login redirect handling, product-specific audit index persistence, approval UX, and publication workflow.

## F2B: Agent Gateway Semantic Tool Enforcement [OSS]

**Deliverables**

- Versioned `tool:invoke` and `backend:request` decision contracts.
- Signed `tool_id`, `skill_hash`, implementation digest, execution, and four-hop `h -> a -> k -> s` attribution.
- Pre-authorization windows bound to master, agent, scope, audience, task/resource, expiry, and revocation handle.
- Argument-level enforcement using policy-relevant projections, classifications, and digests.
- Audience-bound credential-injection obligations and policy-controlled MITM inspection.
- Tool-attribution mismatch, argument escalation, and backend confused-deputy conformance tests.

**Exit gate**

- A protected backend request with absent, forged, or digest-mismatched tool attribution is rejected at the Gateway PEP.
- An allowed tool cannot exceed its delegated scope, argument constraints, backend target, or credential class.
- An exchanged or pre-authorized artifact cannot widen parent scope, audience, expiry, or task binding.

## F3A: MODUREGIS Integration Readiness — Activation, Invocation, and Revocation [OSS]

> MODUREGIS owns Governor execution, PRAXOVELA/RHEOVELA dispatch, and audit index. AEGIVELA supplies adapter activation decisions, execution-bound grant contracts, and revocation semantics. The audience registry and revocation SLO classes are profile-neutral ([ADR-0007](adr/0007-grant-audience-registry.md), [ADR-0008](adr/0008-revocation-propagation-slo.md)); MODUREGIS runtimes are their first conformance consumers.

**Deliverables**

- Adapter activation decision contract for Moduregis Governor: activation only with verified evidence and an allowed decision for its exact adapter version.
- Execution-bound grant contract for `capability:invoke`, audience-bound to product-specific execution runtimes.
- Registered grant audience registry: exact-match, startup-validated audience list with optional per-audience TTL shortening; issuance rejects unregistered audiences and verifiers fail startup on unregistered expected audiences.
- Revocation recheck contract (`pre_dispatch`, `continuation`, `connection` SLO classes) with cache-free authoritative checks for pre-dispatch and continuation, plus a PEP SDK continuation-recheck helper covering grant, subject, resource/policy-version selectors, and lifecycle epoch.
- Revocation selector contract and cross-repository conformance proving pre-dispatch and continuation-bound rejection.
- Agent lifecycle epoch and master-to-agent suspension cascade semantics for the adapter.

**Exit gate**

- Cross-repository fixtures prove an adapter activates only with verified evidence and allowed decision.
- Revoked subject, grant, implementation, or Capability version produces a definitive rejection at the adapter boundary, both pre-dispatch and at declared continuation boundaries; an unavailable revocation check fails closed rather than falling back to cached allow.
- A grant for an unregistered or wrong-runtime audience cannot be issued or verified.

**External integration dependencies (MODUREGIS):** Governor execution, PRAXOVELA/RHEOVELA dispatch, Harmovela correlation, product audit index persistence, and deployment release.

## F4: Enterprise Federation and Continuous Evaluation [EE]

**Deliverables**

- SAML federation, directory/SCIM enrichment, SPIFFE/SVID workload identity, and user-agent-workload binding.
- CAEP-style risk and session signals with policy reevaluation and revocation propagation targets.
- SIEM export, policy simulation, key rotation, regional deployment, and tenant administration controls.

**Exit gate**

- A policy or risk signal can revoke applicable grants within the agreed propagation SLO without cross-tenant impact.
- Federation and workload identity pass conformance, incident-response, and audit-completeness exercises.

## Required Test Matrix

| Area | Minimum proof |
| --- | --- |
| Identity | issuer, audience, expiry, signature, subject, and tenant mapping validation |
| Isolation | no caller tenant override; tenant-scoped policy, cache, and audit queries |
| Policy | allow, deny, approval-required, policy-version and scope checks |
| Delegation | non-amplification across human, agent, tool, and workload identities |
| Revocation | pre-dispatch rejection and long-running continuation rejection |
| Resilience | PDP/IdP timeout or malformed response denies without data exposure |
| Evidence | decision-to-Audit-Index correlation without raw credential or payload retention |
