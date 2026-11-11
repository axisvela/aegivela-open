# Offline-Admission Contract Draft

**Status:** technical draft, 2026-11-09; not a published API or implementation.
The user explicitly approved the conservative boundary for technical review
and new-version drafting. The approval record is maintained in the internal
owner-review register.
That approval does not establish device acceptance, an independently reviewed
trust profile, a stable wire version or permission to enable offline execution.

## Approved First-Profile Boundary

- Derive only from a verified parent execution grant with canonical signed-allow
  lineage. A raw bearer or direct Decision is not an accepted parent.
- The signed audience equals the parent's audience and is already bound to the
  intended device. No implicit audience replacement or narrowing relation.
- Artifact lifetime is at most 900 seconds and never outlives the parent.
- Reserve authority online; a satellite consumes but cannot issue, sign child
  authority, partition a new budget offline or re-delegate.
- Unknown clock/key/trust bounds, or inability to prove required rollback/clone
  resistance, reject offline admission. A local flag never enables the profile.

Existing online `pre_dispatch`/`continuation` checks are unchanged. The profile
grants no Mission, assignment, safety or physical-effect authority. NOMIVELA/
EIDOVELA remain lifecycle/authentication owners; AEGIVELA owns authorization;
AKROVELA owns local consumption/fencing/recovery; resource owners define the
operation model; KINETOVELA retains Mission/assignment/effect ownership.

## Candidate Field Groups

These groups describe required semantics, not a frozen schema. Candidate names
use camelCase and reference fields use `...Ref`. Contract versions follow
`v<major>.<minor>` without prerelease suffixes when assigned. This document's
filename/status, not a draft wire literal, identifies its review state.

| Group | Candidate content | Required binding |
| --- | --- | --- |
| Envelope | `apiVersion`, `grantId`, `issuer`, `issuedAt`, `notBefore`, `expiresAt` | Distinct artifact purpose/version; complete signed payload |
| Parent | `parentGrantRef`, `parentGrantDigest` | Exact verified parent artifact and parent grant identity; no Decision-parent discriminator in the first profile |
| Principal | `namespace`, verified mode-shaped `principal` | Exact parent identity, authority root/master and dual Agent/identity epochs where applicable; no synthetic human subject |
| Target | `audience`, `action`, `resourceRef`, `structuredResourceDigest`, `taskBinding`, bounded `scope` | Preserve or narrow the verified parent, with exact audience/action/task and immutable resource bindings |
| Device | Device/workload/instance references, device key reference, trusted `sourceEpoch`, workspace/site references | Device key possession and trusted boot context; namespace-to-tenant mapping is configured, not caller-selected |
| Constraints | Immutable operation allowlist, target/map/frame references and bounded payload projections/digests | Exact resource-owner semantics and units; unknown constraints reject |
| Reservation | `budgetSourceRef`, `reservationRef`, partition identity, operation/resource bounds | Online policy-authorized allocation and non-overlapping committed reservation |
| Submitter | Permitted submitter reference, proof profile/version and binding | Locally authenticated caller, distinct from device/Agent class and model output |
| Time/revocation | Clock policy, trusted freshness anchor, revocation reference/epoch, signed offline eligibility and stop limits | Conservative authority-time interval and explicit finite disconnected policy |
| Trust/keys | Trust-profile reference/version and verification-key identity | Real verified evidence, finite trusted key/bundle validity and algorithm/purpose binding |
| Evidence | Controlled decision/grant/approval/trace references and sanitized reconciliation facts | Correlatable outcome without raw credentials, sensitive payloads or unrestricted arguments |

`tenantId` is not a normative interoperable selector. Derive the internal tenant
from verified `namespace`. Workspace/site references cannot override it.
`sourceEpoch` is a distinct boot/source binding, never a substitute for
`agentEpoch` or `identityEpoch`. Its concrete wire type and proof profile must
be agreed with the identity/device owner before the machine contract is frozen.

## Online Issuance and Partitioning

1. Authenticate the issuer caller and verify the parent execution grant,
   including signature, canonical allow lineage, current revocation/lifecycle
   state and exact namespace/principal/action/resource/task/audience bindings.
2. Verify device/workload/key/boot evidence and the allowed local submitter
   profile through reviewed owners. Self-reported fields are comparison-only.
3. Evaluate policy permitting this offline profile and a specific numeric budget
   source. A scope-bearing parent alone is not a finite consumable allowance.
4. Validate constraints and bounds against both policy and parent. Compute
   expiry as the earliest parent/profile/identity/trust bound, with an absolute
   900-second maximum. No missing bound may be interpreted as an infinite limit.
5. Atomically reserve online capacity and record the immutable device-bound
   partition, request identity/digest and issuance evidence. Concurrent requests
   cannot both allocate the same final credit. An uncertain commit blocks
   issuance/retry until the authoritative result is resolved.
6. Sign only the committed allocation and return its artifact. A crash after
   reservation but before delivery may strand capacity, never create a second
   allocation on retry. Signing failure cannot expose an uncommitted grant.

Identical issuance identity plus canonical request digest resolves to the same
reservation/artifact result; conflicting reuse rejects. Whether the exact signed
artifact bytes are retained encrypted or reconstructed deterministically is a
storage/key-lifecycle decision for the implementation review, not permission
to mint another independent allowance. Do not store raw bearer artifacts in
audit/evidence records.

Only AEGIVELA creates child partitions, while connected. The sum of outstanding
partitions and settled charges cannot exceed the authorized additive allocation.
Two artifacts referring to the same allocation do not create two spending pools.
Any refinement of a partition remains subject to the same conservation ledger;
this draft gives a satellite no authority to do that offline.

## Budget Semantics

Operation count and cumulative resource quantities require exact, nonnegative,
bounded accounting. Speed, per-operation goal distance and maximum payload size
are ceilings, not automatically additive resources. The resource owner must
freeze integer units, maxima, aggregate/per-operation meaning, rounding and
overflow rejection before schemas and executable constraint validation.

Each requested operation reserves its safe worst-case bounds before admission.
Do not admit an unbounded operation expecting a lower measured settlement.
Runtime enforcement of duration/path/speed/payload limits remains the consumer's
responsibility; an admission record is not proof of physical compliance.

Expiry, revocation, device silence and reconnection alone do not release
outstanding reservations. Reclamation requires reviewed terminal fencing and
authoritative evidence of non-use or settlement. Uncertain execution continues
to consume its reservation. A late authenticated receipt records facts without
restoring authority or allowing revoked/expired work to be newly admitted.

## Local Admission and Recovery

Before each new admission the consumer must:

1. Verify artifact version/purpose, signature, issuer/key trust and parent/device/
   audience/principal/epoch bindings using an approved offline proof profile.
2. Authenticate the local submitter and match its signed constraint; arbitrary
   local processes or model-produced identity fields are not authority.
3. Establish bounded authority time and verify the current offline eligibility,
   expiry/stop/freshness limits and all operation/resource/payload constraints.
4. Serialize access to the reservation ledger. Persist admission identity,
   canonical request digest and budget consumption atomically before execution.
5. Persist execution-start intent durably before invoking the executor. Do not
   re-invoke an uncertain effect on retry or restart.

An identical local admission identity/digest returns the previously recorded
state; a conflicting digest rejects. Persistence failure or uncertain commit
quarantines the affected budget and blocks further admissions until recovery
has authoritative evidence. It cannot be resolved by assuming a failed HTTP
response means the physical effect never began.

Recovery states must distinguish committed admission, durable start intent,
execution outcome known/unknown and authenticated settlement. A database
transaction cannot cover an arbitrary physical executor. No unconditional
exactly-once physical-effect or automatic refund guarantee is made.

Rollback/clone protection is an acceptance prerequisite, not a boolean claim in
the artifact. A copied key, journal and matching epoch can otherwise replay
authority. A verified deployment profile must identify the trusted monotonic/
non-clonable mechanism and test its failure behavior. If it cannot enforce the
required bounds, this profile must reject offline admission, even with a valid
signature and a journal that appears consistent.

## Time, Revocation and Key Policy

Use a verified authority-UTC uncertainty interval. Its lower bound must satisfy
not-before; its upper bound must remain strictly before expiry and every
applicable offline deadline. Unknown/excessive uncertainty or time rollback
rejects. Recheck immediately before admission, not just when receiving a grant.

The artifact cannot exceed 900 seconds or parent validity. Review the concrete
identity/instance lease and trusted-key validity limits as additional shortening
bounds; current lease checking does not imply every such limit is already
present in existing execution-grant claims. No numeric clock skew or nonzero
revocation staleness default is assigned by this draft.

Disconnected use requires signed eligibility, an authenticated non-resettable
freshness anchor and an explicit finite stop limit. When connected, perform
authoritative revocation checks. While disconnected, do not promise knowledge
of undelivered parent suspension or key compromise. Stop at the earliest expiry,
freshness or policy deadline; missing policy means no offline admission.

Offline key resolution uses an authenticated pre-distributed trust root/bundle
with finite validity, not a caller-provided URL or an unknown key supplied with
the request. Rotation/retirement and compromise behavior must fit the same
reviewed disconnected bound. Unknown, unusable or expired trust keys reject.

## Encoding and Errors for Review

Encoding is purpose-separated Ed25519 compact JWS. The exact protected-header
members, purpose string, JCS canonical bytes, parent/payload/binding digests and
golden vectors are fixed in
[Canonicalization and Golden Vectors](offline-admission-canonicalization.md).
Reject duplicate members, unknown versions or algorithms, `none`, unsigned
artifacts and cross-purpose token substitution. Key reuse across purposes
requires explicit authorization; it is not inherited from an ordinary grant's
`kid` or JWKS membership.

Parent digest is SHA-256 of the exact verified compact artifact bytes, not an
arbitrary JSON reserialization. Number formatting now covers finite IEEE-754
doubles; signature vectors and key lifecycle remain implementation gates.

Errors distinguish malformed/unauthenticated, authorization denied, unavailable
verification and local recovery-required states. The closed `errorCode` and
`reasonCode` set, HTTP classes and fail-closed rules are fixed in
[Stable Errors](offline-admission-errors.md). Signatures fail as authentication,
not policy denial; unavailable revocation is `503`; local ledger uncertainty
blocks affected admissions.

## Required Acceptance Matrix

| Case | Required result |
| --- | --- |
| Valid parent/device/profile/reservation and bounded operation | Admit once inside all reviewed limits |
| Direct Decision/raw bearer parent | Reject in the first profile |
| Parent audience A, device audience B | Reject; no implicit audience narrowing |
| Lifetime 901 seconds, or child expiry after parent | Reject |
| Valid signature but no committed authorized reservation | Reject |
| Duplicate issuance/admission identity with same digest | Existing result; no additional allowance or effect |
| Same identity with conflicting digest | Reject |
| Overlapping partitions or two final-credit requests | No double allocation/admission |
| Satellite issues/re-delegates/partitions authority offline | Reject |
| Scope/action/resource/task/namespace/dual-epoch/boot mismatch | Reject |
| Unsupported unit, unsafe rounding, overflow or unbounded payload | Reject |
| Clock interval crosses validity/freshness deadline | Reject |
| Missing signed offline eligibility, unknown trust or key | Reject |
| Journal rollback/cloned device cannot be bounded by proven profile | Reject the offline profile, not merely log after effects |
| Uncertain commit/crash around durable start/executor call | Quarantine reservation; no replay or automatic refund |
| Late receipt after parent revocation/expiry | Deduplicated authenticated fact only; no renewed authority |
| Invalid/none/unknown algorithm, duplicate JSON or wrong purpose | Reject |

Static signature/schema fixtures are necessary but cannot establish ledger
recovery or physical/device trust. Producer PostgreSQL concurrency/rollback/
isolation and consumer fault-injection tests must be separate acceptance suites.

## R1-R11 Coverage and Remaining Decisions

R1-R2 map to parent/identity/device bindings and issuance; R3 to resource-owned
constraints; R4 to allocation, consumption and recovery; R5-R6 to time and
disconnected revocation; R7 to submitter authentication; R8 to encoding/keys;
R9 to errors; R10 to static and stateful acceptance; R11 to a new versioned
contract and reviewed real API paths.

The [requester model cross-reference](../offline-admission-requester-mapping.md)
maps AKROVELA's current field vocabulary and lists unresolved decisions M1-M8
(tenant/workspace/site mapping, count vs attempts, resource units, canonical
bytes, idempotency identity, clock bounds, relay/audience, F4a double-booking).

The protected header/purpose strings, constrained JCS canonicalization,
digests, golden vectors and stable error taxonomy are now fixed locally (see
links above). Before G1: resource owners must approve operation vocabulary/units
and fencing; identity/device owners must specify proof and rollback/clone
interfaces; consumer owners must review ledger/recovery; signature vectors, key lifecycle,
version/paths and a frozen request snapshot remain open. There is currently no schema/protobuf/OpenAPI extension,
endpoint, signer, migration or public revision implementing this draft.

G0 is partially advanced, not closed. G1 remains unclosed; F4b deployment is
blocked. The user explicitly permitted this draft while those independent
decisions remain open. Publishing/enabling the feature needs separate evidence
and approval, not a status-field edit.

## Published Contract

The approved M1-M8 proposals produced the contract at
`contracts/offline-admission/v1.0`, published as G1 at `aegivela-open`
`1b101f2`:

- `offline-admission.schema.json` — authoritative structural schema;
- `offline-admission.openapi.yaml` — paths implemented behind a default-off mount;
- `offline-admission.proto` and generated bindings;
- valid, invalid and semantic-invalid fixtures;
- `backend/internal/offlineadmissioncontract` tests for fixtures, boundary
  invariants, OpenAPI presence and schema/protobuf field parity.

The directory is in `scripts/sync-oss.ps1`. A reference issuer/verifier and HTTP
composition exist, mounted only behind a reviewed default-off gate; deployment
and offline enablement remain unapproved.
