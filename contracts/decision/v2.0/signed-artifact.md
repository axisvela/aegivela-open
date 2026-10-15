# V2 Signed Decision Semantics

The HTTP contract is camelCase. The compact EdDSA JWS payload retains canonical
snake_case claim names; do not rename its signed bytes to match HTTP DTOs.

## Issuance

Only the policy boundary signs its own PDP result. A configured v2 evaluation
performs one policy evaluation and returns its result with `signedDecision`.
Missing verified credential validity, invalid signer configuration, inconsistent
PDP version/digest/expiry or oversized encoding fails closed, without unsigned
success. Raw reference evaluation is not an execution-authority channel.

Credential expiry is internal resolution metadata, not request JSON. Requested
expiry is bounded by actual credential validity and a five-minute decision TTL.
Delegated proof lifetime is the intersection of human and workload credentials.
Agent proof lifetime is the workload credential lifetime, narrowed by any
explicit identity-context validity. Parent bounds are checked before narrowing.
No human subject or identity expiry is synthesized from request preferences.

## Identity Claims

| Identity | Required signed bindings |
| --- | --- |
| Human | Verified tenant, namespace, subject and actor |
| System | Verified tenant, namespace, client, workload and organization authority root; no fabricated subject |
| Service Agent | Agent actor/id/class, workload, namespace, `organization_root` binding, `organization_authority_root`, positive `agent_epoch` and `identity_epoch` |
| Twin Agent | Agent actor/id/class, workload, namespace, `human_master` binding, `master_id`, positive dual epochs; no organization-root substitution |

Instance and attestation references are signed and compared exactly when
present. Service/Twin artifacts carry an empty `subject_ref`, not an invented
human. Client/workload references, actual session reference and task binding
are retained. Legacy artifact versions cannot carry the new v2 identity fields
or be relabelled as v2 through the verify HTTP endpoint.

## Verification

Verification requires configured issuer/key, EdDSA signature, audience,
not-before/issued-at/expiry, canonical resource digest, action/resource, scope
and exact expected principal bindings, including authority root, dual epochs,
instance and attestation. HTTP verification uses the requested resource context,
not merely the digest copied from the artifact. Revocation failures deny.
The verify endpoint uses `X-Aegivela-Internal-Token`; JWKS contains public keys.

Producer and verifier share limits: token 8192 bytes, decoded header 512 bytes,
decoded payload 4096 bytes, bounded key identifiers. Oversized artifacts are
rejected before being returned. Never log raw tokens or private keys.

Only a verified allow is eligible for a grant. Deny, approval-required and revoked
results may be signed for evidence but cannot authorize issuance. Caller-supplied
agent parent metadata is not itself verified delegation authority.

## Acceptance Limits

The SDK/HTTP tests use real decision signatures and verification, with identity,
workload and revocation providers stubbed. They are not deployed-provider or
PostgreSQL acceptance. The existing revocation checker/cache semantics remain;
matching expected epochs does not fetch fresh registry state automatically.
Consumers must obtain authoritative current context and enforce freshness.
Session/task claims are signed, but the current verify DTO does not expose
independent expected session/task fields; consumers needing those expectations
must enforce them separately. Basic resource references are covered by the SDK
tests; SDK-backed structured-resource verification is not certified here.
Public grant issuance and PRAXOVELA adapter migration remain separate work.
