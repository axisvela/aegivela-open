# Offline-Admission Key Purpose and Lifecycle

**Status:** local draft contract decision, 2026-11-09. Not published; no key
service, JWKS endpoint or verifier is enabled.

## Purpose Separation

Offline admission is a separate signing domain from ordinary execution grants.
An artifact is accepted only when all of the following agree:

- protected header `alg` is exactly `EdDSA`;
- protected header `typ` is exactly `aegivela-offline-admission+jws`;
- payload `purpose` is exactly `offline-admission`;
- the header `kid` resolves to an offline-admission key, not merely to any key
  present in the grant JWKS.

The proposed JWKS entry shape is:

```json
{
  "kty": "OKP",
  "crv": "Ed25519",
  "kid": "offline-key-1",
  "use": "sig",
  "alg": "EdDSA",
  "purpose": "offline-admission",
  "x": "<base64url public key>"
}
```

`purpose` is a contract-required JWK member for this set. Private members (`d`)
must never appear. Duplicate `kid` values, a conflicting `use`/`alg`/`crv`, or a
key without the required purpose reject at configuration load. The proposed
endpoint is `/v1/offline-admissions/jwks`; it is separate from
`/v1/grants/jwks` and is not implemented.

## Distribution and Bootstrap

- A consumer must install an authenticated offline-admission trust bundle
  **before** disconnecting. The bundle carries the purpose, keys, and a finite
  validity bound (`notAfter`).
- Key resolution never fetches a caller-provided URL and never accepts a new
  public key merely because it accompanies the artifact.
- Unknown `kid`, an expired bundle, a key outside its validity window, or a
  purpose/algorithm mismatch rejects. Missing trust metadata is not an implicit
  long-lived allow.
- The artifact cannot outlive the verified trust/key validity. Trust bounds
  shorter than the artifact expiry shorten admission.
- The local reference `offlineadmission.VerifyTrustBundle` authenticates the
  bundle with a pinned trust-root key and enforces the purpose, algorithm,
  unique-`kid`, bundle/key `notAfter` and duplicate-member rules.
  `SignTrustBundle` produces the corresponding JWS. It does not provide hardware
  device proof or fetch keys over the network.

## Issuance

- Only an active offline-admission key may sign. A retired or compromised key
  cannot issue new artifacts.
- `kid` is unique per signing domain. Reusing an execution-grant key requires a
  reviewed re-keying decision; it is not implied by sharing a JWKS.
- Key selection is tenant/namespace scoped through configuration, not selected
  by the caller.

## Rotation and Overlap

1. Publish the new public key in a new trust bundle.
2. Switch issuance to the new `kid`.
3. Keep the previous key verify-only during overlap.
4. Retire it after the overlap deadline.

The overlap deadline is the earlier of the bundle `notAfter` and:

```text
last artifact issued with the retiring key
  + maximum artifact lifetime (900 s)
  + maximum clock skew (5000 ms)
  + permitted publication delay
```

A retired key verifies only inside that window. No new artifact may use it.
During disconnection an already-distributed trust bundle cannot learn of a
rotation; the finite staleness/stop policy in the artifact governs admission.
Instantaneous offline retirement is not promised.

## Compromise

Connected consumers reject a revoked key or trust bundle immediately. A
disconnected consumer can only stop at its declared freshness/stop deadline;
the contract must not claim immediate remote compromise awareness. Emergency
response includes revoking the key, publishing a new trust bundle, halting
issuance, and reconciling already-attempted work without restoring its authority.

## Evidence and Storage

- Public keys and `kid` may appear in evidence. Private keys, seeds or derived
  signing material must not.
- Private keys must be protected at rest under the deployment key-management
  policy; this draft does not implement key storage.
- The `kid` in evidence never substitutes for the signature, binding checks or
  revocation policy.

## Vector

`vectors/signature.json` and `vectors/jwks.json` freeze a non-secret verification
pair:

- The private seed is **publicly derivable** from
  `SHA-256('aegivela-offline-admission-test-vector-v1')`; it is not a credential
  and must never be used for real authority.
- Only the public JWK, the signature over the fixed signing input and its
  SHA-256 are committed.
- The Go contract test verifies the signature, rejects a tampered input and
  checks the fixed header/typ/JWKS purpose values.

This proves the byte/signature contract, not runtime key rotation, compromise
propagation or offline trust-bundle enforcement. Those remain G2/G3 acceptance
items.
