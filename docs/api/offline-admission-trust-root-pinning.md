# Offline-Admission Trust-Root Distribution and Pinning

**Status:** deployment guidance plus a reference parser, 2026-11-09. No key
distribution service is implemented.

## Principle

The trust-root key that authenticates offline trust bundles must reach the
consumer through deployment, not through the artifact or the bundle it signs.
If an attacker can replace both the bundle and the root key, the signature adds
no protection.

## Provisioned Files

1. `trust-roots.json` — the pinned trust-root public keys.
2. A separate deployment value — the SHA-256 fingerprint of the exact file
   bytes, delivered through a different channel (environment, secret, deployment
   annotation or release metadata).

Suggested file shape:

```json
{
  "roots": [
    {
      "kty": "OKP",
      "crv": "Ed25519",
      "kid": "trust-root-1",
      "use": "sig",
      "alg": "EdDSA",
      "purpose": "offline-admission-trust",
      "x": "<base64url public key>"
    }
  ]
}
```

Rules:

- The fingerprint is lowercase SHA-256 hex over the exact UTF-8 file bytes.
- The fingerprint must not be stored inside the same file or distributed in the
  same channel.
- Duplicate `kid`, wrong purpose (`offline-admission-trust`, not
  `offline-admission`), wrong algorithm, invalid key bytes and unknown fields
  reject.
- A missing or mismatched fingerprint fails closed; there is no "TOFU" mode.

Reference implementation: `offlineadmission.ParsePinnedTrustRoots(encoded,
fingerprint)` returns the root key map used by `VerifyTrustBundle`, and
`TrustRootsFingerprint(encoded)` produces the out-of-band value.

## Rotation

1. Generate a new trust-root key in the trust domain.
2. Publish a new trust bundle set signed by the new root.
3. Distribute the new `trust-roots.json` and its fingerprint through deployment.
4. Keep the old root pinned only for a finite overlap.
5. Retire the old root after no accepted bundle depends on it.

A trust-root key cannot be rotated by an artifact key. Compromise requires a
deployment update, bundle reissue and consumer reconciliation; disconnected
devices cannot learn the change immediately.

## Deployment Responsibilities

| Concern | Owner |
| --- | --- |
| Generate/protect trust-root key | Trust-root operator |
| Produce/sign trust bundles | Trust-root operator with AEGIVELA key-purpose policy |
| Distribute file + separate fingerprint | Deployment/device fleet owner |
| Verify fingerprint at startup | Consumer (AKROVELA satellite service) |
| Record trust-root revision in evidence | Consumer and fleet owner |

Hardware key custody and anti-clone proof remain in the
[device trust profile](offline-admission-device-trust-profile.md).
