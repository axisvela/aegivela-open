# Offline Admission v1.0 Draft Mapping

This mapping records the user-approved M1-M8 technical proposals. It is a local
draft, not a published compatibility statement or an owner acceptance by the
resource/identity/device/consumer teams.

## Artifact Bindings

| Field | Source / invariant |
| --- | --- |
| `apiVersion` | Const `aegivela.io/offline-admission/v1.0` (draft; not assigned publicly) |
| `offlineGrantId` | AEGIVELA-generated allocation identity |
| `purpose` | Const `offline-admission`; rejects cross-purpose token substitution |
| `issuer` | AEGIVELA artifact issuer |
| `issuedAt` | Issuance time |
| `parentGrantRef` | Verified parent execution-grant identity |
| `parentGrantDigest` | Lowercase SHA-256 hex over exact verified parent compact bytes |
| `namespace` | Verified interoperable anchor; maps to internal tenant by configuration |
| `principal` | Exact parent mode-shaped identity, authority binding and dual epochs |
| `audience` | Must equal the verified parent audience and identify the device |
| `action`, `resourceRef`, `structuredResourceDigest`, `taskBinding`, `scope` | Preserve or narrow the parent; no widening |
| `device.deviceRef`, `device.deviceKeyId`, `device.sourceEpoch` | Device identity, key and durable boot epoch; source epoch is never an Agent epoch |
| `device.workspaceRef`, `device.siteRef` | Signed references; they cannot override namespace/tenant |
| `operations.allowlist` | Exact resource-owner operation types; AEGIVELA does not interpret domain semantics |
| `operations.targetRefs/mapRefs/frameRefs` | Exact immutable target constraints |
| `budget` | Online-reserved integer budget (see below) |
| `submitter` | Authenticated local submitter and proof-profile reference |
| `validity` | `authority-utc`, not-before/expiry and maximum skew 5000 ms |
| `revocation` | `connectedOnly` by default, or a reviewed `boundedDisconnected` policy with positive staleness |
| `trust` | Trust-profile reference/version; the artifact is not proof of device trust |
| `offlineDelegation` | Const `false`; no offline re-delegation |
| `evidenceRefs` | Correlated sanitized evidence references |

## Budget Units (M2/M3)

| Field | Meaning | Accounting |
| --- | --- | --- |
| `maxOperations` | Admitted operations that may invoke the executor | Per admission; a retry inside one admission does not consume another unit |
| `totalDurationMs` | Accumulated executor duration | Cumulative, nonnegative integer |
| `totalGoalDistanceMm` | Accumulated goal distance | Cumulative |
| `totalPathLengthMm` | Accumulated accepted path length | Cumulative |
| `maxSpeedMmPerS` | Motion speed ceiling | Per-operation ceiling, not additive |
| `maxPayloadBytes` | Encoded payload ceiling | Per-request ceiling, not additive |

Overflow, unsafe rounding, negatives and unapproved unit conversion reject.
Resource-owner confirmation of the first-profile vocabulary remains required
before publication.

The signed artifact is the committed-reservation attestation: `offlineGrantId`
plus the signed `budget` prove the allocation and its limits. AEGIVELA signs only
after the reservation commits. No separate reservation-proof field exists, and a
consumer must reject locally asserted reservations without a valid artifact
signature.

## Identity, Scope and Canonicalization (M1/M4/M5/M7)

- AKROVELA `tenant_id` must equal the internal tenant mapped from `namespace`;
  caller-supplied scope is comparison-only.
- `deviceRef` is the AKROVELA `node_id`; `sourceEpoch` is `node_boot_epoch`,
  durable across process restart. It is distinct from `agentEpoch`/`identityEpoch`.
- Dedupe identity: `(namespace, deviceRef, sourceEpoch, operationKey, attemptId,
  bindingDigest, payloadDigest)`. Identical retries return the persisted result;
  conflicting reuse rejects before dispatch.
- Canonical bytes: compact JWS with Ed25519 and RFC 8785 JCS payload. The exact
  protected-header/purpose strings and golden vectors remain a publication gate.
- A coordinator may relay the artifact as transport only; it must not alter any
  signed field or become an issuer.

## Clock and Disconnected Policy (M6)

- Clock domain `authority-utc`; `maxClockSkewMs` is at most 5000.
- Lifetime is at most 900 seconds and never exceeds the parent.
- Default revocation policy is `connectedOnly`; a consumer must reject it for
  disconnected admission. There is no offline staleness allowance.
- `boundedDisconnected` requires a positive `maxOfflineStalenessMs` and a trusted
  local freshness anchor recording the last authoritative revocation observation.
  Admission is allowed only while `now - anchor.checkedAt <= maxOfflineStalenessMs`,
  the anchor is not revoked, the anchor is not from the future beyond the clock
  skew, and the artifact remains within validity/stop bounds. Missing anchor
  rejects.
- An undelivered remote revocation event is not observable offline. The bounded
  policy limits exposure; reconnection records late facts and never restores
  revoked authority.
- Uncertain execution retains its reservation; expiry or reconnect alone does
  not release it.

## F4a Relationship (M8)

The AEGIVELA allocation is the sole authority upper bound. AKROVELA F4a capacity
must be derived from or no greater than it, and both must reference the same
authority identity. Two issued artifacts referring to one parent must not create
two independent spending pools.

## Issue Request Audience

`issue_request.audience` is required. It is the client-declared expected device
and parent-grant audience. AEGIVELA verifies the parent grant against that exact
audience and rejects a mismatch before reserving or signing. A service must not
derive the expected audience from an unverified parent payload.

## Fixed in the Local Draft

- Protected header string, artifact purpose, canonicalization subset and
  parent/payload/binding digest definitions:
  `docs/api/offline-admission-canonicalization.md`.
- Golden canonical bytes and digests under `vectors/`, recomputed by the Go
  contract test.
- Closed error/reason code taxonomy and HTTP classes:
  `docs/api/offline-admission-errors.md` and the schema `error` definition.

## Still Open Before Publication

- Resource-owner approval of the first operation vocabulary and units.
- Identity/device owner approval of the proof profile and rollback/clone test.
- Consumer approval of ledger/recovery state transitions.
- Runtime key rotation, trust-bundle enforcement and compromise propagation.
- Public version/path decision.
- Producer PostgreSQL and consumer fault-injection acceptance.
