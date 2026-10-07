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
- EASEF-DELEGATION binding record, lifecycle epoch, dual-identity artifact claims, and P1-P4 conformance properties.
- Agent Identity Authority records for Twin and Service Agent classes, including immutable authority-root bindings.
- Implemented: Identity Bridge validation of configured enterprise OIDC JWT issuers, audiences, signatures, time claims, and claim-to-principal mappings through cached discovery and JWKS.
- OpenAPI and protobuf definitions for authorization, agent lifecycle, token exchange/pre-authorization, revocation, and adapter activation.
- Policy SDK/PEP error taxonomy: unauthenticated, denied, and unavailable.
- Moduregis port additions for `decision_id`, `evidence_refs`, and non-empty `policy_version` validation.
- Threat model and ADRs for token validation, tenant derivation, audit redaction, and key rotation.

The Policy Decision artifacts establish the F0 request, response, mode, attenuation, outcome, and fail-closed transport semantics only. F0 also delivers internal Identity Bridge JWT validation; product-specific browser/OIDC relying-party flows, optional opaque-token introspection, and external adapters remain future work. Real PDP policy storage, grants, approvals, revocation, attestation, lifecycle resolution, and production evidence persistence remain deferred dependencies.

Generic Web Resource Authorization is implemented: versioned resource/rule/projection contracts, tenant-isolated immutable resource and policy storage, short-lived deny-first UI projections, authenticated internal management APIs, and redacted projection evidence are complete. Implemented: the MODUREGIS adapter's Console OIDC authorization-code + PKCE flow with in-memory refresh (MODUREGIS console consumes the published integration boundary; AEGIVELA remains the authorization authority). The MODUREGIS adapter's runtime switch to `/v1/moduregis/authorize` (F1A port v2) and MODUREGIS-specific resource mapping remain pending.

**Exit gate**

- The Moduregis port maps only a validated allow decision to a non-empty tenant and actor principal.
- Invalid, expired, wrong-audience, unknown-issuer, and tenant-confused tokens fail closed.
- A configured adapter returns `401` for invalid credentials, `403` for policy denial, and `503` only for an unavailable authorization dependency.
- Contract compatibility tests run in both repositories.
- Lifecycle and token-exchange tests prove immutable master binding, scope/audience/expiry attenuation, and post-suspension rejection.

## F1A: MODUREGIS Integration Readiness — `capability:*` Adapter Contract [OSS]

> MODUREGIS owns its Console, OIDC/PKCE integration, browser session management, Capability lifecycle, Catalog, Registry, workflow, and Audit Index. AEGIVELA supplies the versioned adapter contract, action/resource mappings, and cross-repository conformance fixtures.

**Deliverables**

- Implemented: versioned MODUREGIS adapter contract with `capability:read`, `capability:publish`, `adapter:activate`, and `capability:invoke` action/resource mappings, plus positive and negative fixtures.
- Implemented: PDP policy and evidence fields required for MODUREGIS consumption, including `capability:*` and `adapter:activate` actions in the policy evaluator.
- Explicit adapter selection semantics; `DenyAll` remains the no-configuration fallback.
- Implemented: `POST /v1/moduregis/authorize` adapter endpoint evaluating `capability:read`, `capability:publish`, `adapter:activate`, and `capability:invoke`, plus in-repo contract conformance tests. Implemented: cross-repository conformance delivery with the MODUREGIS repository — MODUREGIS consumes the contract schema and fixtures from the versioned Go module (`pepsdk/moduregiscontract`) and validates the identical fixture set in its own CI.

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

- Implemented: client enrollment and device/workload identity binding for the Client Gateway (`agentidentity` enrollment and workload binding records).
- Implemented: immutable human-to-agent binding records and master-suspension cascade handling.
- Implemented: Service Agent registration with organization authority-root binding and no synthetic human master.
- Implemented: trusted-principal resolution and five-mode negative tests across `human_web`, `system_api`, `delegated_api`, Twin Agent, and Service Agent authority sources.
- Implemented: workload binding, assertion, JWKS, and binding/JTI revocation checks at the gateway identity boundary.
- Implemented: AEGIVELA adapter for the existing gateway `POST /v1/policy/connect` foundation (`gatewaycontract` connect evaluator and policy adapter).
- Implemented: `POST /v1/gateway/connect` decision endpoint for the Server Gateway, resolving Twin/Service agent and `system_api` workload-assertion principals, granting `tool:invoke` only to verified agent authorities with matching attribution.
- Implemented: short-lived, bounded-TTL connection decisions with explicit degraded-mode eligibility.
- Implemented: Server Gateway PDP, approval, credential-injection, and evidence-export integration (credential/MITM/target obligations, [ADR-0009](adr/0009-gateway-decision-obligation-contract.md)).
- Implemented: revocation and policy-version invalidation delivery to Client Gateway caches (`invalidation_service`).

**Exit gate**

- An enrolled agent can access only an authorized target through the Client and Server Gateway PEPs.
- Expired, revoked, unavailable, or non-offline-eligible decisions block rather than creating a degraded-mode bypass.
- Allow, block, approval, injection, and cache invalidation events are correlated in security evidence.
- All five authority modes reject missing, forged, mismatched, expired, revoked, and cross-tenant principal evidence; resolver, binding, policy, and revocation outages return `503` and fail closed.

## F2A: MODUREGIS Integration Readiness — Approval and Publication Contracts [OSS]

> MODUREGIS owns publication workflow, approval UX, audit indexing, and console behavior. AEGIVELA supplies adapter contracts, policy/approval evidence fields, and cross-repository conformance. The underlying approval record and artifact mechanisms are profile-neutral ([ADR-0006](adr/0006-approval-record-ownership.md)); MODUREGIS publication is their first conformance profile.

**Deliverables**

- Implemented: PDP actions for `capability:publish` and approval-required outcomes expressed in the versioned adapter contract.
- Implemented: approval binding evidence fields: human decision, reason, resource digest/version, scope, expiry, and evidence references.
- Implemented: signed approval artifact contract (`contracts/approval/v1alpha2`): Ed25519 JWS artifacts binding tenant, action, immutable resource version or descriptor-bound structured digest, scope, and expiry, verified offline by product PEPs and never bypassing the authoritative `approval_jti` revocation check.
- Implemented: scope attenuation and delegated-grant contracts for publisher and agent actions.
- Implemented: cross-repository fixture suite with valid/invalid fixtures for approval issue and revocation check (F3A), including positive/negative test vectors across both `contracts/approval/v1alpha2` and `contracts/revocation/v1alpha1` directories; in-repo conformance tests validated in both repositories' CI.

**Exit gate**

- The adapter rejects a publish outside its documented tenant, namespace, scope, or approval window.
- A resubmitted mutation whose parameters do not match or narrow the verified approval artifact bindings is rejected; an unavailable approval verification or revocation dependency fails closed.
- Cross-repository audit correlation contracts are published; fixture parity between `contracts/approval/v1alpha2` and `contracts/revocation/v1alpha1` is verified.

**External integration dependencies (MODUREGIS):** Console refresh-token and enterprise-login redirect handling, product-specific audit index persistence, approval UX, and publication workflow.

## F2B: Agent Gateway Semantic Tool Enforcement [OSS]

**Deliverables**

- Implemented: versioned `tool:invoke` and `backend:request` decision contracts.
- Implemented: signed `tool_id`, `skill_hash`, implementation digest, execution, and four-hop `h -> a -> k -> s` attribution.
- Implemented: pre-authorization windows bound to master, agent, scope, audience, task/resource, expiry, and revocation handle.
- Implemented: argument-level enforcement using policy-relevant projections, classifications, and digests.
- Implemented: audience-bound credential-injection obligations and policy-controlled MITM inspection (structured `credential_ref`/`credential_class`/`mitm_required`/`mitm_scope`/`allowed_target_paths` fields, [ADR-0009](adr/0009-gateway-decision-obligation-contract.md)).
- Implemented: tool-attribution mismatch, argument escalation, and backend confused-deputy conformance tests.
- Implemented: S4 `tool:invoke` adapter contract (`contracts/tool/v1alpha1`) with tool invariant claims (`tool_id`, `skill_hash`, `implementation_digest`), SLO-class-aware revocation recheck (`pre_dispatch`/`continuation`/`connection`), and cross-repo valid/invalid fixtures. A tool grant can only be issued for an exact verified implementation and fails closed on an unavailable revocation check.

**Exit gate**

- A protected backend request with absent, forged, or digest-mismatched tool attribution is rejected at the Gateway PEP.
- An allowed tool cannot exceed its delegated scope, argument constraints, backend target, or credential class.
- An exchanged or pre-authorized artifact cannot widen parent scope, audience, expiry, or task binding.
- A revoked tool subject, grant, or implementation produces a definitive rejection at both pre-dispatch and continuation boundaries.

## F3A: MODUREGIS Integration Readiness — Activation, Invocation, and Revocation [OSS]

> MODUREGIS owns Governor execution, PRAXOVELA/RHEOVELA dispatch, and audit index. AEGIVELA supplies adapter activation decisions, execution-bound grant contracts, and revocation semantics. The audience registry and revocation SLO classes are profile-neutral ([ADR-0007](adr/0007-grant-audience-registry.md), [ADR-0008](adr/0008-revocation-propagation-slo.md)); MODUREGIS runtimes are their first conformance consumers.

**Deliverables**

- Implemented: adapter activation decision contract for Moduregis Governor: activation only with verified evidence and an allowed decision for its exact adapter version.
- Implemented: execution-bound grant contract for `capability:invoke`, audience-bound to product-specific execution runtimes.
- Implemented: registered grant audience registry: exact-match, startup-validated audience list with optional per-audience TTL shortening; issuance rejects unregistered audiences and verifiers fail startup on unregistered expected audiences.
- Implemented: revocation recheck contract (`pre_dispatch`, `continuation`, `connection` SLO classes) with cache-free authoritative checks for pre-dispatch and continuation, plus a PEP SDK continuation-recheck helper covering grant, subject, resource/policy-version selectors, and lifecycle epoch.
- Implemented: revocation selector contract with in-repo conformance; cross-repository pre-dispatch and continuation-bound rejection fixtures await the MODUREGIS repository.
- Implemented: agent lifecycle epoch and master-to-agent suspension cascade semantics for the adapter.

**Exit gate**

- Cross-repository fixtures prove an adapter activates only with verified evidence and allowed decision.
- Revoked subject, grant, implementation, or Capability version produces a definitive rejection at the adapter boundary, both pre-dispatch and at declared continuation boundaries; an unavailable revocation check fails closed rather than falling back to cached allow.
- A grant for an unregistered or wrong-runtime audience cannot be issued or verified.

**External integration dependencies (MODUREGIS):** Governor execution, PRAXOVELA/RHEOVELA dispatch, Harmovela correlation, product audit index persistence, and deployment release.

## F4: Enterprise Federation and Continuous Evaluation [EE]

**Status**: ✅ **Completed** — all S1–S8 slices delivered and reviewed (spec + quality both APPROVED).

**Deliverables** (all implemented across slices S1–S8)

- **S1 SPIFFE+CAEP**: Core public `spiffe` package + EE `coreclient` with trust-domain-bound SVID verification + `POST /v1/risk/signals` risk signal forwarding + EE risk→revocation flow. (Commits: S1 push).
- **S2 SAML XML-DSig**: Real XML-DSig verification via `goxmldsig`; critical signature-wrapping bypass fix + SHA-256 enforcement. (Commits: S2 push).
- **S3 Policy Simulation**: Core internal-token-gated `POST /v1/policy/decisions/simulate` + ADR-0011 + EE `coreclient.Evaluate` with correct error mapping (400/401/403/503). (Commits: S3 push).
- **S4 SIEM CEF/LEEF**: `FormatCEF`/`FormatLEEF` with newline/CR escaping, deterministic field order, LEEF conformant header with EventID+0x09 delimiter, CEF Device Version. `format` query param on evidence endpoint (ndjson/cef/leef). (Commits: S4 push).
- **S5 Key Rotation PG+JWKS**: Tenant-aware `signing_keys` PostgreSQL migration + `public_jwk`/`private_jwk` columns + `BuildJWKSet` strips private fields + `GET /.well-known/jwks.json` endpoint with tenant filtering + `405 Method Not Allowed` for non-GET. (Commits: S5 push).
- **S6 Region Tenant Affinity**: `region` column on `tenants` table + `AEGIVELA_REGION` affinity filter in `List`/`Get` (SQL + app-layer defense) + index on `region`. (Commits: S6 push).
- **S7 Tenant IdP Config CRUD**: `idp_config JSONB` column on `tenants` + `Suspend` 404 when not found + Update/Create/Delete IdP config per tenant. (Commits: S7 push).
- **S8 CI+Version**: EE CI workflow (`ee.yml`) + version alignment across repos + integration test infrastructure design. (Commit: S8).

**Exit gate** ✅ **Satisfied**

- A policy or risk signal can revoke applicable grants within the agreed propagation SLO without cross-tenant impact.
- Federation and workload identity pass conformance, incident-response, and audit-completeness exercises.
All F4 exit gate criteria are validated through sub-agent two-stage review (spec + quality) for each slice.

## GA Boundary Reached — `1.0.0` Release

The AEGIVELA Open Core reaches its first General Availability release at `v1.0.0`. This is marked by:

- **F1–F4** fully implemented and sub-agent-reviewed (spec + quality both APPROVED).
- **GA Boundary ADR** — `docs/adr/0013-ga-boundary.md` declares the GA boundary conditions met.
- **Version synchronization** — `aegivela` core `v1.0.0` + `aegivela-open` `v1.0.0` (shared version).
- **EE independent** — `aegivela-ee` at `v1.1.0` (first enterprise release at 1.x).
- **API contracts** — published and versioned; no breaking changes to public packages.
- **Tenant isolation** — preserved across all tenant-scoped operations.
- **Fail-closed** — all security gates verified and tested.

**Release Tags**

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.0.0` |
| `aegivela-open` | `v1.0.0` |
| `aegivela-ee` (EE) | `v1.1.0` |

**Exit gate criteria** fully met. Roadmap F4 section updated to reflect completed status.

## F5: Agent IAM Series Authorization Conformance [OSS/EE]

**Purpose**

Close the remaining gaps between AEGIVELA contracts/runtime behavior and the
Agent IAM Series, principally Part 3 (Identity and Authentication), Part 4
(Authorization and Delegation), Part 5 (Federation), Part 6 (Audit), and Part 7
(Conformance). This stage does not turn AEGIVELA into an identity registry or a
general business-service catalog. A separate Agent Registry remains authoritative
for Agent/Agent ID registration and discovery; EIDOVELA remains the authentication
authority; a product or enterprise-architecture repository remains the owner of
business, application, and technology-service models.

**Deliverables**

- Implemented: Canonical allow lineage — Enterprise bearer exchange consumes a
  verified PDP `allow` Decision with identical principal, action, resource,
  scope, and audience bindings; the exchange service never synthesizes an
  `allow` outcome.
- Implemented: Part 4 contract alignment — published the Part 4 request/response
  mapping in [Part 4 Contract Mapping](part4-contract-mapping.md), covering
  Decision, Execution Grant, approval, attenuation, and revocation with
  `namespace`, lifecycle epoch, parent lineage, policy version, task,
  obligations, and the exact action/resource/audience bindings. The
  `policy/v1alpha3` principal contract publishes the registry-consumer fields
  (`namespace`, `authority_binding`, `agent_epoch`, `identity_epoch`,
  `instance_id`) alongside the legacy single `lifecycle_epoch`; `obligations`
  are published on the decision response. The frozen `v1alpha1`/`v1alpha2`
  transports are unchanged.
- Implemented: Revocation boundary alignment — the revocation check endpoint,
  published contracts, PEP SDK, and runtime agree on the selector type and value
  and the `pre_dispatch`/`continuation`/`connection` freshness class; an
  unavailable dependency fails closed.
- Implemented: Lifecycle and activation alignment — the Agent lifecycle
  OpenAPI/schema, DTOs and routes are reconciled with the runtime. Activation
  consumes a verified enrollment result: `POST /v1/agents/{id}/enroll` verifies
  the evidence prescribed by the active profile (device_backed,
  human_authorized, api_only) and is the only path from `created` to `active`;
  `/activate` only restores an already-enrolled suspended agent and returns
  `409 enrollment_required` otherwise. Activation is never inferred from
  caller-supplied identity fields.
- Implemented: Trusted identity context — `identitysrc` consumes the EIDOVELA
  verified agent context, and only a `Provider` may construct a principal. The
  Identity Resolve and Evidence schemas now carry `namespace`, authority
  binding, the NOMIVELA dual lifecycle epochs and the instance reference.
  AEGIVELA does not re-derive namespace, Agent class, workload or lifecycle
  epoch from request input.
- Implemented: Part 6 evidence parity — sanitized evidence envelopes are
  persisted to the append-only `evidence_events` table through one shared
  `securityevidence/postgres` recorder ([ADR-0014](adr/0014-persist-sanitized-security-evidence.md)).
  Outcome and rejection events from identity resolution, execution-grant
  issuance and verification, policy decisions, and web-resource projections
  correlate by tenant, namespace, trace ID, and the controlled
  Decision/Grant/Approval references. Raw tokens, credentials, unrestricted
  prompts, and raw tool arguments remain prohibited.
- Implemented: [EE] Signing and key rotation — `aegivela-ee` persists rotation
  records to `ee_signing_keys` through a `keyrotation.Persister` and loads them
  into the operational signer at startup (environment as fallback). A retired
  key remains valid for verification until the domain overlap window elapses:
  `max artifact lifetime + permitted clock skew + publication delay`.
- Implemented: Attestation policy enforcement — issuance requires every
  assertion group in the configured profile (`AEGIVELA_ATTESTATION_REQUIRED_GROUPS`;
  default artifact, signer, workload, environment). The workload group is
  satisfied only by a verified workload assertion; the self-reported
  `workload_id`/`workload_evidence` issuance path is removed. This aligns the
  Attestation Service with the profile contract for agent/MCP issuer
  trust (Parts 2 and 4).
- Implemented: Gateway mode enforcement — the gateway connect request carries an
  explicit `authorization_mode` that must agree with the agent class and
  identifier (`system_api` carries no agent authority); a cross-mode token
  substitution is rejected before principal resolution and policy evaluation.
  The non-agent `system_api` surface validates end-to-end, and the contract
  publishes `authorization_mode` with valid and substituted fixtures.
- Implemented: Enterprise service authorization profile — defined
  in [Enterprise Service Authorization Profile](enterprise-service-authorization-profile.md)
  and published as `contracts/enterprise-service/v1alpha1`. A product-owned typed
  service reference (business, application, or technology service) is carried as
  a v1alpha2 structured resource descriptor and validated against the profile
  vocabulary on the policy-request path; the PDP evaluates it with
  data-classification, segregation-of-duties, accountable owner, approval, and
  cost-center constraints. AEGIVELA does not own the enterprise architecture
  catalog; the profile defines only the trusted reference and its canonical
  `structured_resource_digest`.

**Exit gate**

- Token exchange tests prove that only a verified `allow` Decision derives a
  grant and that all child grants attenuate scope, audience, expiry, action,
  resource, and task.
- Contract, SDK, and runtime tests produce identical namespace-scoped revocation
  behavior for every freshness class; unavailable dependencies return a
  fail-closed result.
- Lifecycle activation tests reject missing, forged, stale, or self-reported
  enrollment/attestation evidence.
- Cross-repository fixtures from `agent-iam-spec` Parts 3–7 run against the
  AEGIVELA contract surface, including positive and negative audit events.
  *Pending external delivery:* the fixture set is owned by the `agent-iam-spec`
  series repository; AEGIVELA publishes the contract surface and in-repo
  contract tests and executes the cross-repository fixtures once they are
  delivered.
- A pilot resource descriptor derived from an enterprise architecture model is
  authorized as an immutable typed resource, not as a free-form service name or
  natural-language prompt.

**Exit-gate status:** the token-exchange, revocation, and lifecycle criteria are
satisfied by in-repo tests; the `agent-iam-spec` cross-repository fixture
criterion is pending external delivery, and the pilot resource-descriptor
criterion remains exploratory.

## F5 Release — `1.1.0`

The Agent IAM Series authorization conformance stage (F5) ships as Open Core
`v1.1.0`.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.1.0` |
| `aegivela` backend module | `backend/v1.1.0` |
| `aegivela-open` | `v1.1.0` |
| `aegivela-ee` | `v1.2.0` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.1.0`. The
Enterprise Edition counts independently (`v1.2.0`).

## Patch Release — `1.1.1`

A patch on the F5 line closes the tracked `policy/v1alpha3` transport gap and
completes the identity-source positioning:

- Generated the `policy/v1alpha3` protobuf descriptor, mirrored the OpenAPI to
  the JSON Schema, and extended the drift test so the JSON Schema, OpenAPI,
  protobuf source, generated descriptor, and baseline move together.
- Renamed the `EASEF-IAM` profile to `EASEF-DELEGATION`, repositioned the
  product boundary (NOMIVELA registry, EIDOVELA authentication), added
  [ADR-0015](adr/0015-identity-source-boundary.md), and completed
  `contracts/identity-source/v1alpha1` with a mapping, fixtures, and a schema
  test.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.1.1` |
| `aegivela` backend module | `backend/v1.1.1` |
| `aegivela-open` | `v1.1.1` |
| `aegivela-ee` | `v1.2.0` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.1.1`. The
Enterprise Edition counts independently (`v1.2.0`).

## Release — `1.2.0`

A minor release completing the v2.0 camelCase contract surface across the
policy service and aligning the four schema-only integration contracts. Major
and minor remain unchanged from the F5 line; the v2.0 contracts are served
alongside the frozen v1alphaN transports, which are not modified.

- The policy service serves the v2.0 camelCase contract for every routed
  endpoint: evaluate/sign, decision verify/JWKS, grant issue, approval,
  pre-authorization, token exchange, attestation, agent lifecycle, identity
  resolve, workload binding/assertion, gateway connect, MODUREGIS authorize,
  `tool:invoke`, risk signal, revocation check, and web resources.
- Added the execution-grant verification surface: `POST /v1/grants/verify`
  enforces the PEP-declared audience, scope coverage, and task binding, and
  `GET /v1/grants/jwks` publishes the grant signing public keys. A grant bound
  to a task can never be verified without that task binding; verification
  authenticates the PEP credential (`X-AEGIVELA-PEP`).
- Aligned the four schema-only contracts to v2.0: `identity-source`
  (normalized verified principal), `enterprise-service` (typed descriptor
  vocabulary with a pinned canonical `structuredResourceDigest`),
  `webapi` PEP (camelCase modes, strictly bounded and fail-closed SDK response
  handling), and `evidence` (canonical camelCase envelope).
- Security evidence now serializes through one canonical v2 envelope path:
  producer-local fields are normalized and non-contract fields are dropped, so
  raw resource references and producer-specific properties never reach the
  append-only store or SIEM export. Evidence ID generation fails closed instead
  of panicking when entropy is unavailable.
- Identity-source v2 decoding fails closed on unknown authentication state,
  inconsistent Twin master bindings, and mixed-version documents.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.0` |
| `aegivela` backend module | `backend/v1.2.0` |
| `aegivela-open` | `v1.2.0` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.0`. The
Enterprise Edition counts independently: it releases `v1.2.1`, which pins the
Open Core backend module at `github.com/axisrobo/aegivela/backend v1.2.0` and
builds standalone (no shared Go workspace required).

## Patch Release — `1.2.1`

A patch on the v1.2.0 line fixes a protobuf contract defect and makes proto
lint enforceable in CI:

- `contracts/tool/v2.0/tool-invoke.proto` incorrectly reused the v1alpha1 proto
  package (`aegivela.tool.v1alpha1`) and `go_package`, producing duplicate
  symbol collisions against `tool/v1alpha1`. It now declares
  `aegivela.tool.v2_0` and its own `go_package`.
- `buf lint` now passes. The STANDARD rules that the v2.0 camelCase wire
  contract intentionally violates (field casing, enum value casing, and the
  package version suffix) are explicitly excepted in `buf.yaml`, matching the
  existing exception list.
- Backend CI installs the pinned `buf` (v1.50.0) and runs `buf lint`, so future
  proto drift fails the build.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.1` |
| `aegivela` backend module | `backend/v1.2.1` |
| `aegivela-open` | `v1.2.1` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.1`. The
Enterprise Edition counts independently (`v1.2.1`, pinning `backend v1.2.0`).

## Patch Release — `1.2.2`

A patch completing the Open Core distribution surface:

- `scripts/sync-oss.ps1` now publishes `contracts/enterprise-service` and
  `contracts/tool` to `aegivela-open`. Both are AEGIVELA-owned `[OSS]`
  integration contracts (F5 enterprise-service authorization profile and F2B
  `tool:invoke` enforcement) and were previously published in the core
  repository only.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.2` |
| `aegivela` backend module | `backend/v1.2.2` |
| `aegivela-open` | `v1.2.2` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.2`. The
Enterprise Edition counts independently (`v1.2.1`, pinning `backend v1.2.0`).

## Patch Release — `1.2.3`

A patch that makes the v2.0 protobuf wire contracts machine-enforced:

- Versioned the v2.0 `go_package` paths for the domains that previously shared
  the v1alphaN output path, so `buf generate` no longer collides.
- Generated protobuf Go bindings for every contract domain under
  `backend/internal/<domain>contract/pb`.
- Added the `backend/internal/contractdrift` test, which pins every v2.0
  domain's message field numbers, field types, and enum values against the
  populated `protobuf-baseline.json`. A rename, renumber, or type change now
  fails the build; run with `-contract-update` to regenerate after an
  intentional change.
- Backend CI runs `buf lint` and verifies the generated protobuf is current.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.3` |
| `aegivela` backend module | `backend/v1.2.3` |
| `aegivela-open` | `v1.2.3` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.3`. The
Enterprise Edition counts independently (`v1.2.1`, pinning `backend v1.2.0`).

## Patch Release — `1.2.4`

A patch extending v2.0 contract enforcement and repairing two invalid artifacts:

- Extended the `contractdrift` test with schema↔OpenAPI property parity for
  every v2.0 domain: any definition with inline properties and a same-named
  OpenAPI component must expose the identical property-name set.
- Repaired `contracts/token-exchange/v2.0` OpenAPI (a literal `` `n`` had
  replaced a newline) and `contracts/approval/v2.0` OpenAPI (an unquoted regex
  made the document invalid YAML).
- Renamed the `contracts/token-exchange/v2.0` proto field
  `requested_resourceRef` to `requestedResourceRef`, matching the schema and the
  v2.0 camelCase convention, and regenerated its bindings and baseline.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.4` |
| `aegivela` backend module | `backend/v1.2.4` |
| `aegivela-open` | `v1.2.4` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.4`. The
Enterprise Edition counts independently (`v1.2.2`, pinning `backend v1.2.4`).

## Patch Release — `1.2.5`

A patch closing the schema↔proto side of v2.0 contract enforcement:

- Extended the `contractdrift` test with schema↔proto field-name parity, so
  every schema definition with inline properties and a same-named protobuf
  message must expose the identical field set.
- Added the missing `signedDecision` field to the `token-exchange/v2.0`
  `TokenExchangeRequest`; the schema and OpenAPI already required it, so the
  proto was incomplete. Regenerated its bindings and baseline.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.5` |
| `aegivela` backend module | `backend/v1.2.5` |
| `aegivela-open` | `v1.2.5` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.5`. The
Enterprise Edition counts independently (`v1.2.3`, pinning `backend v1.2.5`).

## Patch Release — `1.2.6`

A patch that starts executing the AgentIAM cross-repository conformance fixtures
and aligns decision obligations with the published contract:

- Vendored the `agent-iam-spec` cross-repository fixtures under
  `contracts/agent-iam-spec/cross-repo` and added `backend/internal/conformance`,
  which validates every fixture against its schema, enforces that each
  AEGIVELA-owned fixture has a mapping, and executes six Part 4 fixtures
  (revocation pre-dispatch, decision deny/allow grant derivation, approval
  binding, delegation non-amplification, credential-injection obligation).
- Aligned the `decision verify` response and the signed-decision claims with the
  published `decision/v2.0` contract: verified decisions now carry
  `obligations`, propagated from the policy decision.

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.6` |
| `aegivela` backend module | `backend/v1.2.6` |
| `aegivela-open` | `v1.2.6` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.6`. The
Enterprise Edition counts independently (`v1.2.3`, pinning `backend v1.2.5`).

## Patch Release — `1.2.7`

A patch that adds the evidence ingestion boundary and extends the conformance
harness to AEGIVELA's Part 6 fixtures:

- Added `securityevidence.Ingest` ([ADR-0020](adr/0020-evidence-ingestion-boundary.md)):
  a validated ingestion gate that rejects prohibited fields (raw tokens,
  credentials, proofs, prompts, arguments, attributes), rejects unknown fields,
  requires the correlation fields, and normalizes the event to the canonical v2
  envelope.
- Extended the cross-repository harness to execute the Part 6
  `security-event-persisted` and `event-redaction` fixtures (eight AEGIVELA
  fixtures executed in total).

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.7` |
| `aegivela` backend module | `backend/v1.2.7` |
| `aegivela-open` | `v1.2.7` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.7`. The
Enterprise Edition counts independently (`v1.2.3`, pinning `backend v1.2.5`).

## Patch Release — `1.2.8`

A patch that satisfies AEGIVELA's Part 7 conformance-claim obligations:

- Added the AgentIAM Part 7 claim validator `conformance.ValidateClaim`, which
  enforces composite-profile composition (Identity = Parts 2, 3; Authorization =
  2, 3, 4; Federated = 2, 3, 4, 5).
- Published AEGIVELA's component conformance claim at
  [conformance-claim.md](conformance-claim.md).
- Extended the cross-repository harness to execute the three Part 7 fixtures
  (eleven AEGIVELA fixtures executed in total).

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.8` |
| `aegivela` backend module | `backend/v1.2.8` |
| `aegivela-open` | `v1.2.8` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.8`. The
Enterprise Edition counts independently (`v1.2.3`, pinning `backend v1.2.5`).

## Patch Release — `1.2.9`

A patch that adds PEP obligation enforcement and executes the matching fixture:

- Added `pepsdk.GatewayConnectDecision.EnforceObligations`
  ([ADR-0009](adr/0009-gateway-decision-obligation-contract.md)): the PEP fails
  closed when a named obligation, required credential injection, or required
  MITM inspection has not been enforced before the effect is dispatched.
- Extended the cross-repository harness to execute the Part 4
  `obligation-enforcement` fixture (twelve of the thirteen AEGIVELA fixtures).

| Repository | Tag |
|---|---|
| `aegivela` (core) | `v1.2.9` |
| `aegivela` backend module | `backend/v1.2.9` |
| `aegivela-open` | `v1.2.9` |

Open Core is synchronized — `aegivela` and `aegivela-open` share `v1.2.9`. The
Enterprise Edition counts independently (`v1.2.3`, pinning `backend v1.2.5`).

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
