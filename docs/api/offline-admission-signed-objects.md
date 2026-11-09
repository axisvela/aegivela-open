# Offline-Admission Cryptographic Signed Objects

**Status:** design record, 2026-11-09. Accepted for the design by
`hanhuiwen@axisvela.com`; publication and runtime enablement remain separate.

This document lists every cryptographic signature relevant to F4b, who produces
it, which key domain it uses and what bytes it covers. Review sign-offs are
recorded separately in the internal owner-review register; they are not
cryptographic signatures.

## Object Summary

| Object | Signer | Key domain | Status |
| --- | --- | --- | --- |
| Parent execution grant | AEGIVELA | Existing execution-grant Ed25519 domain | Existing; F4b verifies and references it |
| Offline-admission artifact | AEGIVELA | Purpose-separated `offline-admission` Ed25519 domain | Implemented in the unwired reference |
| Offline trust bundle | Trust-root operator | Purpose-separated offline trust-signing domain | Implemented in the unwired reference |
| Device/local submitter proof | Device or local workload | Device-local credential/key | Consumer-owned; AEGIVELA only defines the binding |
| Consumer receipt / consumption fact | AKROVELA satellite (optional) | AKROVELA-owned key | Consumer-owned; not part of this contract |
| Approval artifact | AEGIVELA | Existing approval Ed25519 domain | Existing; referenced through the parent lineage |

## Offline-Admission Artifact

The only new artifact AEGIVELA issues:

```text
protectedHeader = {"alg":"EdDSA","kid":"<offline key id>","typ":"aegivela-offline-admission+jws"}
payload         = JCS(Artifact claims)
signingInput    = base64url(JCS(protectedHeader)) + "." + base64url(payload)
signature       = Ed25519(privateKey, signingInput)
token           = signingInput + "." + base64url(signature)
```

The signed claim set is fixed by
`contracts/offline-admission/v1.0/offline-admission.schema.json`: parent
reference/digest, namespace/principal, audience, action/resource/structured
digest/task/scope, device binding (`deviceRef`, `deviceKeyId`, `sourceEpoch`,
workspace/site), operation allowlist and target/map/frame constraints, budget,
submitter binding, validity/clock/skew, revocation policy/reference, trust
profile and `offlineDelegation=false`.

`parentGrantDigest` is SHA-256 over the **exact verified parent compact JWS
bytes**, not a reserialized JSON view. The artifact signature covers that digest
and every binding above; no unsigned field may affect admission.

Golden bytes, digests and a verification vector are under
`contracts/offline-admission/v1.0/vectors/`.

## Offline Trust Bundle

The verifier needs keys before disconnecting, so the bundle itself must be a
signed object. Proposed and design-accepted format:

```text
protectedHeader = {"alg":"EdDSA","kid":"<trust root key id>","typ":"aegivela-offline-admission-trust+jws"}
payload         = JCS({
  "apiVersion": "aegivela.io/offline-admission-trust/v1.0",
  "purpose": "offline-admission",
  "issuer": "<trust root issuer>",
  "issuedAt": "<RFC3339>",
  "notAfter": "<RFC3339>",
  "keys": [ {kty,crv,kid,use,alg,purpose,x,notAfter?}, ... ]
})
signingInput    = base64url(header) + "." + base64url(payload)
```

Rules:

- Signed by a trust-root key distinct from the artifact signing key; an artifact
  key cannot authorize a new trust bundle.
- The trust-root public key is distributed out of band or pinned at deployment.
- `purpose`, `alg`, `kty`, `crv`, unique `kid`, bundle/key `notAfter` and
  duplicate-JSON rejection are enforced before use.
- A bundle may rotate/retire verification keys; it cannot retroactively
  invalidate already-attempted offline effects.
- Unknown, expired or unsigned bundles reject.

Implementation status: SignTrustBundle and VerifyTrustBundle implement the JWS
format above, including signature, header purpose, bundle/key validity and
content checks. Trust-root public-key distribution/pinning remains a deployment
and G3 acceptance item; the reference does not fetch it over the network.

## Parent Execution Grant

The parent is the existing AEGIVELA execution grant (`aegivela.io/v2.0`). F4b
does not create or re-sign it. The issuer/verifier treat it as trusted only after
the existing signature, canonical-allow lineage, revocation, lifecycle and
binding checks pass. F4b binds to its exact compact bytes via SHA-256.

Approval artifacts referenced by the parent are likewise existing AEGIVELA
signed artifacts; F4b signs no new approval.

## Consumer-Owned Signatures

These are not AEGIVELA artifacts and must not be represented as such:

- **Submitter proof:** the device/local workload signs or authenticates the
  submitter binding. AEGIVELA signs the constraint, not the local proof.
- **Consumer receipts / consumption facts:** if AKROVELA signs them, it uses its
  own key and format. AEGIVELA defines the reconciliation references, not the
  receipt signature. Whether receipts are signed is an AKROVELA acceptance
  decision.

## Not Signed

The following are never signed as authority artifacts and must not be treated as
such: raw bearer tokens, passwords/MFA assertions, prompts, unrestricted tool
arguments, arbitrary model output, unsigned JSON policy decisions, and evidence
envelopes (evidence is stored redacted, not used as authority).

## Key Separation Summary

| Domain | Purpose |
| --- | --- |
| Execution grant | Parent authority |
| Approval | Human approval artifact |
| Offline admission | New offline artifact |
| Offline trust | Signs trust bundles/key sets |
| Consumer/device | Local submitter and optional receipts |

`kid` from one domain does not authorize another domain. Algorithm is `EdDSA`
for AEGIVELA artifacts; any alternative needs a new reviewed contract version.
