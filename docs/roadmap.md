# AEGIVELA Delivery Roadmap

**Reviewed:** 2026-11-09. This is the current delivery view, not a release
announcement or approval of a new authorization contract.

## Delivery Rules

Each release must prove a security boundary: reject invalid authority, preserve
tenant isolation under failure, and emit correlatable, sanitized evidence.
An endpoint, passing unit tests, or a code review alone does not close a
consumer deployment or operational SLO gate.

- **Implemented:** code and contracts exist; verification scope is stated separately.
- **Covered:** executable tests exist; skipped integration tests are not passing evidence.
- **Accepted:** the owning consumer supplies a dated result pinned to revisions and configuration.
- **Released:** applicable gates and synchronized Core/Open publication are recorded.
- **Proposed:** requirements or sequencing only; no authority or API is granted.

NOMIVELA owns Agent registration/lifecycle; EIDOVELA owns authentication.
AEGIVELA consumes verified identity and owns authorization/delegation. The
retained standalone identity behavior is documented in [Identity Boundary](identity-boundary.md).
MODUREGIS and Gateway profiles progress independently. Shared primitives remain
profile-neutral; product proxy, workflow, dispatch and audit-index code stays
with its product owner.

**[OSS]** denotes general security primitives and integration contracts.
**[EE]** denotes enterprise infrastructure and operational extensions. See
[OSS/EE Boundary](oss-ee-boundary.md). A private intake location does not assign
an edition.

## Current Baseline

The wire contract is `aegivela.io/v2.0` alongside frozen `v1alphaN` transports.
Current Go modules use `/v3`; product versions, module versions and wire versions
must not be conflated. Source pins, local test evidence, tag discrepancies and
external acceptance live in [Conformance Status](conformance-status.md).

Historical GA and v1.x release records have moved to
[Release History](release-history.md). They do not certify the current checkout.
No synchronized Core/Open release or new tag is asserted by this review.

**Active execution policy:** local CI first, as directed by the user on
2026-11-09. Core/EE/SDK tests and real PostgreSQL integration are the current
development gate. Remote Actions scheduling/billing and EE private-checkout
credentials (WI-10) are deferred, not prerequisites for local progress. This
does not waive F6 owner approval, consumer acceptance or future release gates.

## Stage Status

| Stage | AEGIVELA-owned implementation | Remaining gate |
| --- | --- | --- |
| F0 [OSS] Contract and trust foundation | Implemented: versioned contracts, OIDC validation, principal resolution, PDP storage, grants, approvals, attestation and evidence | Maintain compatibility and failure/isolation tests; consumer evidence belongs to the profile gates below |
| F1A [OSS] MODUREGIS adapter | Implemented: four action mappings, authorize endpoint, embedded contract fixtures | Product runtime selection, route/resource mapping and deployment acceptance |
| F1B [OSS] Gateway identity and connection | Implemented: trusted modes, workload binding/assertions, bounded decisions, invalidation and obligations | Partial real-service acceptance recorded; full degraded/approval/injection/invalidation and failure matrix remains open |
| F1C [OSS] Web/API PEP | Implemented: SDK, mode separation, verified delegated-parent enforcement and v2.0 transport | Product-owned route/configuration acceptance |
| F2A [OSS] Approval/publication contracts | Implemented: signed artifacts, scope/expiry/resource bindings and revocation contracts | MODUREGIS mutation/approval UX, audit correlation and runtime rejection evidence |
| F2B [OSS] Semantic tool enforcement | Implemented: exact implementation attribution, argument projections, obligations and SLO-aware rechecks | Tool-runtime fixture execution and real dispatch/continuation enforcement |
| F3A [OSS] Activation/invocation/revocation | Implemented: audience registry, execution bindings and authoritative rechecks | MODUREGIS Governor/dispatch and consumer pre-dispatch/continuation rejection |
| F4 [EE] Federation/continuous evaluation | S1-S8 implementation and prior reviews recorded | Current deployment-specific SLO, incident-response and audit-completeness evidence must be supplied separately |
| F5 [OSS/EE] AgentIAM conformance | All 13 owned behaviors passed locally with PostgreSQL 18.1; remote PostgreSQL suite evidence verified; v2 contract/evidence hardening implemented | Composite/consumer acceptance and exploratory descriptor pilot |
| F6 [OSS/EE scope review] Bounded offline admission | G0 design/owner review accepted by `hanhuiwen@axisvela.com`; M1-M8 and G1 coordinates approved; published v1.0 contract, vectors, key lifecycle, signed-objects record, reference issuer/verifier behind a default-off gate, signed trust-bundle verification, pinned trust roots, atomic ledger with issuance evidence and G3/profile schemas exist; G1 published at `aegivela-open` `684886a`, clarified at `8c0c69f`, issue audience at `e5f1645`, task binding at `1b101f2`; HTTP transport + config assembly can mount behind a default-off flag; no deployment or enablement | G3: clean AKROVELA `16e3fa2` passes the package, recovery 10x, all 20 component cases and the full backend suite; AEGIVELA resolved reservation-proof (signed artifact) and disconnected-revocation (bounded anchor) semantics, pending AKROVELA reruns; the named device-profile proof and the reviewed service+mount configuration are implemented in the reference (ADR-0023); owner-approved profile provisioning, deployment and enablement remain. AEGIVELA added: the named `DeviceTrustProfile` structural model and artifact-profile binding (`internal/offlineadmission`, schema-aligned) wired into the issuer/verifier as an optional `Profiles` set (default off; an unapproved or mismatched proof profile fails closed); and the reviewed mount gate `httpapi.MountOfflineAdmission` with operations config (`docs/operations/offline-admission.md`), default-off; enabling also requires an approved device trust profile and an evidence recorder, enforced at assembly startup |

## Existing Security Gates

### F0: Contract and Trust Foundation [OSS]

The early static evaluator is not the current implementation limit: real PDP
policy storage, grants, approvals, revocation, attestation and evidence
persistence are implemented. Generic Web Resource Authorization includes
immutable tenant-scoped storage, deny-first projections and redacted evidence.
Local Agent identity/enrollment remains a standalone compatibility behavior,
not a transfer of NOMIVELA's family authority.

Exit evidence must reject invalid issuer/audience/signature/expiry and tenant
confusion, preserve immutable master binding and dual epochs, and prove
scope/audience/expiry attenuation. A configured PEP maps invalid identity to
401, denial to 403 and unavailable authorization dependencies to 503.

### F1A / F2A / F3A: MODUREGIS [OSS]

AEGIVELA supplies `capability:read`, `capability:publish`, `adapter:activate`
and `capability:invoke`, approvals, grants and revocation. MODUREGIS owns its
Console sessions, Capability state, Governor, publication workflow and audit
index. The initial OIDC/PKCE implementation was previously recorded; that is
not evidence of complete refresh, enterprise redirect or release acceptance.

Exit evidence must show an unconfigured adapter fails closed, an allow has
non-empty trusted principal/policy/evidence bindings, publication matches its
approval, and activation/invocation uses the exact verified resource/version.
Wrong audience, revoked subject/grant/implementation/Capability version and
unavailable pre-dispatch or continuation checks block execution. Consumers
return fixture and deployment results pinned to the published contracts.

### F1B / F1C / F2B: PEP and Gateway [OSS]

The SDK and gateway enforce mode separation, trusted workload/Agent identity,
parent attenuation, tool/skill/implementation attribution and target/credential
obligations. The [2026-10-13 record](praxovela-integration.md#consumer-integration-status-2026-10-13-real-service-verified)
reports LIMENORA CONNECT allow/deny for system and federated Service Agent paths.
That historical result does not close all five-mode or current-version gates.

Exit evidence must cover forged/missing/mismatched/stale/cross-tenant identity,
unavailable dependencies, denied targets, argument escalation, tool substitution,
approval, injection/MITM obligations and correlated cache invalidation.
Unenforced obligations block effects. Connection caching is bounded and policy
controlled; `pre_dispatch` and `continuation` remain cache-free authoritative
checks. Existing degraded connection eligibility is not offline operation admission.

### F4: Enterprise Federation and Continuous Evaluation [EE]

The delivered slices cover SPIFFE/CAEP risk-to-revocation, SAML XML-DSig,
simulation, CEF/LEEF export, persistent key rotation/JWKS, region affinity,
tenant IdP configuration and CI. Historical review completion is preserved in
[Release History](release-history.md), not substituted for operational proof.

Deployment acceptance requires bounded risk/policy revocation propagation with
tenant isolation, federation/workload conformance, incident-response and audit
completeness. Record environment, revisions, measured results and failures;
regional tenant filtering alone does not prove multi-region replication/routing.

### F5: AgentIAM Authorization Conformance [OSS/EE]

Implemented deliverables include canonical signed allow lineage, dual-epoch
identity consumption, revocation freshness alignment, verified enrollment,
profile-enforced attestation, sanitized evidence and the typed enterprise-service
descriptor profile. Wire contracts use camelCase v2.0; historical snake_case
`v1alphaN` documents are frozen compatibility lines.

The delivered AgentIAM set has 27 fixtures. AEGIVELA owns 10 authorization/audit
fixtures plus 3 Part 7 claim fixtures; 14 identity/federation behaviors belong to
EIDOVELA. All 13 AEGIVELA behaviors have tests. Schema validation of all 27 does
not prove their behaviors. The transaction-consistency behavior requires
PostgreSQL. The initial local run skipped it; after the user provided a database
instance, all 13 owned behaviors passed locally with PostgreSQL 18.1 and verbose
transaction-consistency output. Separately, the
[exact-source remote Core run](https://github.com/axisvela/aegivela/actions/runs/42064201299)
passed on 2026-11-05 with PostgreSQL and the database variable set, including
the uncached conformance package. Historical remote and later local results are
separate evidence; neither proves future F6 budget/recovery semantics.

Exit evidence must include the database-backed atomic grant/evidence failure
test, namespace/epoch/freshness/attenuation rejection and compatible transports.
The enterprise-architecture resource descriptor pilot remains exploratory until
a product catalog supplies an immutable real descriptor and acceptance result.
The component claim does not certify a composite NOMIVELA/EIDOVELA deployment.

## F6: Bounded Offline Admission [Proposed]

**State:** G0 design/owner review accepted by `hanhuiwen@axisvela.com` on
2026-11-09 (R1-R11, M1-M8, G1 coordinates); the requester source revision is not
yet frozen and runtime/device evidence remains open. **Receiving owner:**
AEGIVELA. **Consumer:** AKROVELA F4b, still blocked.

The intake is `aegivela-ee/docs/akrovela-offline-admission-requirements.md` in
the sibling private repository. It includes uncommitted requester corrections,
is not a revision-pinned public contract, and is not owner acceptance. The
requester's reviewed public baseline is `aegivela-open` `0a6a801` (since
force-rewritten; current public head is `8a26bdc`).

F6 concerns admitting a new bounded operation while disconnected, not executing
an already-admitted F4a operation. No new `robotics` principal class is implied.
Existing authorization modes are selected by actual authority. Agent lifecycle,
boot/source epoch, workload identity and local submitter identity remain distinct.

### Responsibility Proposal

General authorization, budget reservation/partition semantics, signed artifacts,
verification and compatibility belong in Core/public contracts under the
existing edition principle. Enterprise device-trust integrations and managed
operations may extend them in EE. Final scope requires owner review.

AKROVELA owns domain constraint enforcement, durable local admission/consumption,
execution fencing, recovery and receipt reporting. AEGIVELA defines authority
reconciliation semantics; KINETOVELA retains Mission/assignment/effect authority.
No safety authority, credential injection, destructive actions, new targets or
scope expansion is granted by this request.

### Ordered Gates

| Gate | Required output | Opens |
| --- | --- | --- |
| F6-G0 Owner/security review | Named reviewer; pinned requester revision; dispositions for R1-R11; namespace/device/parent/audience, budget, clock, revocation and rollback/clone decisions | Contract authoring, not consumer implementation |
| F6-G1 Public contract | New versioned schema, OpenAPI/protobuf mapping, exact canonical signed bytes, errors, key rules, fixtures and compatibility; approved exact public revision | AKROVELA contract review and implementation planning |
| F6-G2 Core implementation | Reviewed issuance/reservation/partition/reconciliation behavior, migration/isolation/failure tests, verifier and SDK behavior against G1 | Candidate runtime integration, not deployment approval |
| F6-G3 Consumer acceptance | AKROVELA local authentication, atomic consumption, concurrent/retry/crash/restart, rollback/clone and disconnected-policy evidence | Release consideration within the accepted trust model |

Do not mutate `aegivela.io/v2.0` in place or expose fixture logical paths as
implemented endpoints. Parent audience equality, existing 15-minute grant TTL,
identity/decision expiry and online revocation requirements are current
constraints, not automatically relaxed for F6. Classify additive resource
limits separately from per-operation ceilings such as speed and payload size.

Reject absent offline eligibility, unknown clock/key/trust bounds and exhausted
authority. Signatures and software counters alone cannot prevent journal rollback
or device cloning. Remote revocation is not instantaneous while disconnected;
the reviewed policy must bound staleness and stop time. Uncertain execution
retains its reservation; reconnect facts never restore revoked authority.

No new artifact version, endpoint, TTL, staleness allowance or hardware trust
profile is assigned by this roadmap. The [owner review register](offline-admission-review.md)
tracks R1-R11 and the [technical response proposal](offline-admission-proposal.md)
supplies recommendations beyond the scoped user approval. The
[prose contract draft](api/offline-admission-draft.md) is explicitly permitted;
it is not a published schema, API or device approval. The
[requester model cross-reference](offline-admission-requester-mapping.md) lists
M1-M8; the [G0 decision proposals](offline-admission-decision-proposals.md) were
approved by the user for a local machine draft. The contract under
`contracts/offline-admission/v1.0` is published at `aegivela-open` `1b101f2` and
included in the OSS sync list; a reference runtime exists behind a default-off
mount. Remaining resource/device/identity/consumer review, device hardware proof
and runtime enablement still gate G2/G3. The internal readiness plan is
`superpowers/plans/2026-11-09-offline-admission-readiness.md` (not a public
distribution artifact); it starts with G0. Deployment requires the remaining approvals.

## Priorities and Dependencies

| Priority | Work | Owner | Completion evidence |
| --- | --- | --- | --- |
| P1 | Release/module/dependency convergence (WI-18) | AEGIVELA + EE/Open maintainers | Exact source/module/public pins, reachable public CI checkouts and clean-checkout CI; no historical tag fabrication |
| P1 | F6-G0 review (WI-19) | AEGIVELA + AKROVELA + identity/resource owners | Design/owner review accepted (R1-R11, M1-M8); freeze the requester revision and record resource/device/identity/consumer inputs |
| P1 | Existing consumer acceptance (WI-01..04, WI-13..15) | MODUREGIS, LIMENORA, product PEPs, PRAXOVELA | Revision-pinned path-specific rejection and effect-gating evidence |
| Baseline complete | Database verification (WI-20) | AEGIVELA | Local PostgreSQL 18.1 Core/EE suites and all 13 owned conformance behaviors passed; preserve coverage limits and keep the repaired key fixture enabled in CI |
| P2 | F6-G1, then G2/G3 | Respective gate owners | Predecessor approval; no parallel implementation against an unsettled contract |
| P2 | Refresh/workflow/audit, continuation and distribution (WI-05..07, WI-11, WI-16..17) | Product owners + AEGIVELA | Product runtime and public self-verification evidence |
| P3 | Correlation, descriptor pilot and operations (WI-08..09, WI-12) | Product owners + AEGIVELA | Real pilot inputs and deployment/runbook/observability proof |

G0 review may proceed alongside existing integration work; G1-G3 are sequential.
No dates or release numbers are promised before owner scope and dependencies
are accepted. Track detailed work and results in [Conformance Status](conformance-status.md).

## Required Test Matrix

| Area | Minimum proof |
| --- | --- |
| Identity | issuer/audience/expiry/signature, tenant mapping, dual epochs and live-instance lease rejection |
| Isolation | no caller tenant override; tenant-scoped policies, indexes, reservations, caches and evidence |
| Policy | allow/deny/approval, policy version, immutable resource and scope checks |
| Delegation | signed allow lineage and non-amplification across identity, audience, expiry and task |
| Revocation | authoritative pre-dispatch/continuation; bounded connection cache; explicit F6 disconnected policy |
| Resilience | unavailable or malformed dependencies fail closed without credential/payload exposure |
| Evidence | correlated redacted envelopes; grant/issuance evidence commit or roll back together |
| Offline budgets | reservation conservation, non-overlap, concurrent exhaustion, retries and uncertain execution |
| Offline recovery | durable start intent, no duplicate execution/refund; clock/key/epoch/rollback/clone rejection or reviewed limitations |

## Update Discipline

Update status, owner and evidence together. Keep product runtime acceptance
distinct from in-repo coverage and historical review. Security authority, grant,
lifecycle, PEP and external contract changes require an ADR or architecture
update. Publish contracts before enabling dependent adapters. Only close a gate
with the owning repository's applicable results; never infer it from a skipped
test, a proposal, a local `replace`, or an unrelated fixture.
