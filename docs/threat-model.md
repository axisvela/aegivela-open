# AEGIVELA Threat Model

## Scope and Attacker Model

This threat model covers the AEGIVELA control plane: Identity Bridge, Policy Decision Point and signed policy decisions, execution grants, Approval Scope Engine, token exchange and pre-authorization, revocation, Attestation Service, and security evidence. Enterprise IdP internals, product domain state (MODUREGIS Registry, gateway runtimes), and product audit indexes are out of scope; trust boundaries are defined in [architecture.md](architecture.md).

Attacker classes:

- **External forger:** crafts or replays tokens and signed artifacts without valid keys.
- **Malicious tenant principal:** holds valid credentials and attempts cross-tenant access or authority widening.
- **Compromised PEP or adapter:** submits crafted requests to AEGIVELA endpoints.
- **Key/compromise attacker:** obtains one signing key or floods verification with unknown key IDs.
- **Insider with read access** to PostgreSQL or log/evidence pipelines.

Protected invariants: tenant isolation, non-amplification of delegated authority, revocation timeliness, signing-key confidentiality, and audit/evidence confidentiality.

## Global Mitigations

- All AEGIVELA-signed artifacts are compact EdDSA (Ed25519) JWS with a pinned algorithm and configured `kid` ([ADR-0001](adr/0001-use-ed25519-jws-for-signed-artifacts.md)), separate keys per artifact domain ([ADR-0003](adr/0003-use-separate-signing-keys.md)).
- Persistence and evidence are digest-only ([ADR-0002](adr/0002-use-digest-only-persistence.md)).
- Revocation uses generic tenant-scoped selectors ([ADR-0004](adr/0004-use-revocation-selectors.md)).
- Every security dependency failure denies ([ADR-0005](adr/0005-fail-closed-on-unavailable.md)).
- Every endpoint is internally authenticated (`X-Aegivela-Internal-Token`), strictly decodes bodies (no unknown fields, no trailing content, size caps), and derives identity only from verified principals.

## 1. Identity Spoofing

| Threat | Mitigation |
| --- | --- |
| Forged enterprise token | Identity Bridge accepts only configured issuers; `iss` never triggers dynamic discovery. Only allowlisted asymmetric algorithms (no `none`, no HMAC). Signature verified against discovered JWKS (5 s deadline, 1 MiB cap, no redirects). Exact issuer, audience, and `exp`/`nbf`/`iat` checks. `tenant_id`, `subject_ref`, `actor_id`, `client_id` are mapped only from verified claims; every endpoint rejects caller-supplied identity fields as unknown fields. |
| Forged signed decision | Decisions are compact EdDSA JWS: protected header is exactly `alg: "EdDSA"` plus the configured decision `kid`; unknown `kid`, other algorithms, duplicate/unknown claims, or artifacts over 8192 bytes are rejected. Validity window is positive and at most five minutes. Verification requires exact issuer, audience, tenant, subject, actor, action, resource, policy version, session, task, and lifecycle-epoch bindings. `POST /v1/policy/decisions/evaluate` and `POST /v1/policy/decisions/sign` require internal caller authentication and resolve the principal from trusted mode-specific evidence. They evaluate and sign only that resolved principal; caller-supplied principal, policy result, and unsigned decision data have no signing authority. Invalid authentication or principal evidence is `401`; unavailable identity, policy, or revocation dependencies are `503` and deny. |
| Forged execution grant | Grants use an independent signing key and issuer from decisions. Issuance requires a verified `allow` decision plus a bridge-resolved bearer token; unsigned decision JSON and principal overrides have no authority. Grant verification pins `EdDSA`, the configured grant `kid`, exact issuer/audience, time claims, and all binding claims; failure is `401`/`403`, never partial acceptance. |
| Forged workload assertion | Workload assertions use an independent Ed25519 key, issuer, and `GET /v1/workload-assertions/jwks` endpoint. Verification requires the configured issuer and audience, an active exact tenant-scoped workload binding, and available binding and JTI revocation checks. A forged, expired, mismatched, or revoked assertion is unauthenticated or denied; unavailable key, binding, or revocation dependencies return `503` and fail closed. |

## 2. Tenant Confusion

| Threat | Mitigation |
| --- | --- |
| Cross-tenant data access | `tenant_id` comes only from the verified token's configured tenant claim; `X-Aegivela-Tenant-ID`-style headers and body fields are rejected or ignored. Every PostgreSQL aggregate, unique constraint (`(tenant_id, jti)`), index, query, and revocation cache key begins with `tenant_id`. Approval and pre-authorization checks require the record to belong to the requesting tenant. Migrations carry tenant-isolation tests. |
| Tenant-scoped selector leakage | Revocation selectors and stored metadata hold SHA-256 digests (subject, resource, grant), not raw identifiers, so one tenant's identifiers never appear in shared tables, caches, or evidence. Cache keys are tenant-keyed, preventing a cached no-revocation result from applying across tenants. |

## 3. Scope Expansion

| Threat | Mitigation |
| --- | --- |
| Wider-than-allowed scope in grants | A requested scope must be a duplicate-free subset of the signed decision's `effective_scope`; the grant carries the attenuated subset and PEPs must enforce it. Grant expiry is the earliest of 15 minutes, principal expiry, and decision expiry. |
| Approval window widening | A grant bound to an `approval_jti` must be equal to or narrower than the approval on scope, audience, expiry, action, resource, and task; with both `parent_jti` and `approval_jti` present it must be narrower than both. Approvals are append-only with a 24-hour maximum TTL. |
| Pre-authorization abuse | Windows are quota-bounded (`max_grants` 1-100), expire within 24 hours, and fix the scope/audience envelope; grants under a window must be equal or narrower. EASEF-IAM P2/P3 hold structurally: `master_id` is injected from the authoritative record (never caller-supplied), and token exchange only attenuates (subset scope, exact audience/action/resource match, non-increasing expiry). Widening attempts are `403`, not errors of degree. |

## 4. Revocation Bypass

| Threat | Mitigation |
| --- | --- |
| Stale cache | Only negative (no-revocation) results are cached, with TTL bounded by both the configured maximum and the artifact's remaining lifetime; a cached "not revoked" can never outlive the artifact. Cache keys include tenant, JTI, subject/session, policy version, resource digest, and lifecycle epoch. Expired cache entries are never served. |
| Unknown KID flooding | The bridge attempts at most one synchronous JWKS refresh per unknown `kid`; refreshes are coalesced per issuer and suppressed for 30 seconds, so attacker-chosen `kid`s cost bounded work and the token is unauthenticated during suppression. Artifact verifiers accept only the configured `kid`; there is no remote key fetch for AEGIVELA artifacts. |
| Concurrent refresh amplification | Per-issuer refresh coalescing collapses concurrent misses into one fetch; each fetch has a 5-second deadline and 1 MiB response cap. On refresh failure an expired cache is not used: resolution returns `503 identity_unavailable`. |
| Revocation check evasion | Revocation is checked after cryptographic verification inside the same verify call; an unavailable check returns `ErrRevocationUnavailable` and maps to `503`/deny, never an allow ([ADR-0005](adr/0005-fail-closed-on-unavailable.md)). |

## 5. Key Compromise

| Threat | Mitigation |
| --- | --- |
| Single-key rotation without overlap | F0 rotates by controlled restart with a new `kid`; the old `kid` becomes unknown immediately, so artifacts from a compromised removed key fail closed. The availability cost of non-overlapping rotation is accepted and documented; overlapping keys and KMS/HSM are deferred work ([ADR-0003](adr/0003-use-separate-signing-keys.md)). |
| Key reuse across artifact types | Decision, grant, attestation, and workload assertion domains each require an independent Ed25519 key, `kid`, and issuer via separate configuration; contracts state a key must not be reused across domains. Cross-domain confusion fails cryptographically (wrong key) and by issuer/audience binding. Startup fails when signing is enabled with a missing or invalid key file. JWKS endpoints publish public keys only. |

## 6. Audit Leakage

| Threat | Mitigation |
| --- | --- |
| Raw tokens in logs | Identity Bridge evidence excludes bearer JWTs, raw claims, and JWKS documents; grant and decision flows never log compact JWS. Error responses contain only a `code` field and echo no claims. |
| Unrestricted arguments in evidence | Evidence envelopes carry only allowlisted fields: tenant, digests (decision/grant/subject/resource/scope), action, audience, policy version, outcome, reason code, and validated trace ID. Free-text reasons are stored only as `reason_digest`. |
| Bearer artifacts in database | PostgreSQL stores grant/attestation metadata as digests; the compact JWS is never persisted. Pre-authorization and approval rows are append-only with UPDATE/DELETE/TRUNCATE blocked by triggers. See [ADR-0002](adr/0002-use-digest-only-persistence.md). |

## 7. Replay

| Threat | Mitigation |
| --- | --- |
| Grant re-submission | A grant is bound to exact tenant, subject, actor, audience, action, resource, policy version, session, task, and lifecycle epoch; replay against any other context fails binding verification. Lifetime is at most 15 minutes; `grant_jti` selectors revoke individually. |
| Decision re-use | Decisions live at most five minutes, are verified against the exact request binding, and their expiry bounds any issued grant. `decision_jti` revocation shortens the window further. |
| Workload assertion replay | Assertions live at most five minutes and bind tenant, client, workload, binding ID/version, organization root, audience, and JTI. Verification requires the active exact binding and tenant-scoped JTI/binding revocation, preventing replay after binding revocation or into another workload context. |
| Pre-auth window replay | Window consumption is an atomic `used_grants` increment; exhaustion or expiry denies with `403 grant_denied`. Windows are tenant-bound with a fixed scope envelope and a `pre_authorization_jti` revocation selector. |

## 8. Timing

| Threat | Mitigation |
| --- | --- |
| TOCTOU between verification and revocation | Revocation is evaluated after signature and binding checks within the same verification operation; long-running operations recheck at continuation boundaries. EASEF suspension writes revocation records and increments the lifecycle epoch in the lifecycle transaction before reporting success, so a PEP with a stale epoch rejects the artifact. |
| Expired-but-uncached decisions | Time validity (`nbf`/`iat`/`exp`) is intrinsic to the artifact and re-checked on every verification; it is never cached. Only revocation state is cached, negatively, with TTL capped by artifact expiry, so caching cannot extend an artifact's life. |

## 9. Attestation Spoofing

| Threat | Mitigation |
| --- | --- |
| Forged artifact digests | `artifact_digest` is a signed claim bound to `artifact_ref`, signer identity, tenant, and subject; consumers compare it against the actual artifact digest at the consumption point. The Attestation Service never trusts an agent to self-assign assurance; policy may require attestation evidence before grant issuance or credential injection. |
| Missing assertion types | All four assertion groups (artifact, signer, workload, environment) are mandatory at issuance and verification; partial attestations are rejected. |
| Environment claim injection | Claims are bounded (1-16 pairs, restricted character set, 256-character limits; workload evidence 1-8 pairs). Request bodies cannot carry `tenant_id`, `subject_ref`, or `issuer`; those derive from the resolved identity and server configuration. Attestation metadata persists digests only, and artifacts expire after one hour with `attestation_jti`, `workload_id`, and `artifact_digest` selectors available for revocation. |

## Residual Risks

- Non-overlapping key rotation causes a brief issuance outage per domain; artifacts verify fail-closed throughout.
- Pull-only revocation means propagation delay is bounded by the negative-cache TTL, never zero.
- A compromised enterprise IdP is out of scope; damage is contained by audience/issuer binding, short derivative TTLs, and subject/session revocation selectors.
