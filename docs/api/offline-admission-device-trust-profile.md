# Offline-Admission Device Trust Profile

**Status:** required deployment profile, 2026-11-09. AEGIVELA does not attest
hardware and cannot create this proof. Without an accepted profile, offline
admission rejects.

## Why

A valid signature and a correct `sourceEpoch` do not prove that a journal cannot
be rolled back or a device cannot be cloned. The artifact therefore binds a
`trustProfileRef`/`trustProfileVersion`; the consumer may admit offline only when
the deployed device matches an accepted profile.

## Required Profile Content

| Field | Required content |
| --- | --- |
| `profileRef`, `profileVersion` | Immutable identity of the reviewed profile |
| `deviceClass` | Exact supported hardware/OS class; no wildcard |
| `keyCustody` | Where the device key lives and what can export/copy it |
| `sourceEpochSource` | How the durable boot epoch is read and why process restart preserves it |
| `singleConsumer` | What prevents two consumers using one allocation concurrently |
| `rollbackResistance` | Concrete mechanism and evidence for journal/state rollback |
| `cloneResistance` | Concrete mechanism and evidence for device/key cloning |
| `enforceableBounds` | Which time/key/budget bounds the device can prove |
| `rejectWhenUnproven` | Explicit reject behavior when a mechanism is unavailable |
| `evidenceRefs` | Test reports, hardware attestations or deployment evidence |
| `approver`, `approvedAt` | Named identity/device owner and date |

A self-declared boolean such as `"cloneResistant": true` is not evidence.
Unsupported profiles, partial mechanisms and unknown versions must reject.

## Template (Non-Wire)

```json
{
  "profileRef": "akrovela-device-profile",
  "profileVersion": "1",
  "deviceClass": "site-controller-a",
  "keyCustody": "TPM-backed non-exportable key",
  "sourceEpochSource": "secure element monotonic boot counter",
  "singleConsumer": "exclusive local service lock owned by the executor",
  "rollbackResistance": "monotonic NV counter checked before admission",
  "cloneResistance": "device certificate bound to secure element",
  "enforceableBounds": ["clock", "key", "budget", "epoch"],
  "rejectWhenUnproven": true,
  "evidenceRefs": ["evidence:device-profile/1"],
  "approver": "identity-device-owner@example",
  "approvedAt": "2026-11-09"
}
```

The template is illustrative; it is not an accepted hardware claim.

## Acceptance Requirements

1. The identity/device owner verifies the mechanism, not just the document.
2. The consumer proves the mechanism fails closed when unavailable (counter read
   error, missing key, unknown device, rollback detected).
3. The profile states its limits: a restored journal or cloned device that
   cannot be detected must cause rejection, not silent admission.
4. AEGIVELA records the profile reference in the artifact but does not
   independently verify hardware; a mismatch or unknown profile rejects.
5. Reconnection reconciliation detects facts after the effect; it does not
   retroactively prevent an offline effect already attempted.

Structural validation is available at
`contracts/offline-admission/v1.0/device-trust-profile.schema.json`. Passing it
means the profile is complete, not that the mechanism works.

## Current State

No hardware profile has been independently accepted. The reference verifier
checks the artifact's `trustProfileRef` presence, not hardware truth. F4b remains
blocked until a profile and its failure evidence are recorded.
