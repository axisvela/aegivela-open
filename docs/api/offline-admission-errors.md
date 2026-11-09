# Offline-Admission Stable Errors

**Status:** local draft contract decision, 2026-11-09. Not published.

The error envelope has `apiVersion`, `errorCode`, `reasonCode` and `message`
(see `contracts/offline-admission/v1.0/offline-admission.schema.json#/$defs/error`).
`errorCode` is the transport-level class; `reasonCode` is the stable semantic
cause. `message` is human-readable and must never contain tokens, credentials,
raw payloads or sensitive attributes.

## Transport Classes

| `errorCode` | HTTP | Meaning |
| --- | --- | --- |
| `invalid_offline_admission_request` | 400 | Malformed, unknown or wide request |
| `unauthenticated` | 401 | Missing/invalid caller, submitter or signature identity |
| `authorization_denied` | 403 | Verified request denied by authority, binding, budget or delegation |
| `revocation_unavailable` | 503 | Authoritative revocation dependency unavailable; fail closed |

## Reason Codes

| `reasonCode` | Class | Producer/consumer | Meaning |
| --- | --- | --- | --- |
| `offline_grant_invalid` | 403 | Both | Artifact/claim structure valid JSON but semantically invalid |
| `offline_grant_expired` | 403 | Both | Current verified time is at/after expiry |
| `offline_grant_not_yet_valid` | 403 | Both | Current verified time is before not-before |
| `offline_grant_revoked` | 403 | Both | Parent/grant/implementation is revoked |
| `offline_grant_audience_mismatch` | 403 | Both | Audience differs from parent/expected device |
| `offline_grant_device_epoch_stale` | 403 | Consumer | `sourceEpoch` does not match trusted boot state |
| `offline_grant_scope_widening` | 403 | Producer | Requested scope/action/resource/task extends the parent |
| `offline_grant_budget_exhausted` | 403 | Both | No committed reservation or remaining allowed units |
| `offline_grant_not_delegable` | 403 | Producer | Offline issuance, re-delegation or new partition attempted |
| `offline_grant_submitter_unauthenticated` | 401 | Consumer | Local submitter proof missing, invalid or not bound |
| `offline_grant_signature_invalid` | 401 | Both | Wrong algorithm/purpose/key, bad signature or unknown key |
| `offline_grant_recovery_required` | 503 | Consumer | Local ledger state uncertain; affected admissions blocked |
| `revocation_unavailable` | 503 | Both | Authoritative revocation check unavailable |
| `invalid_offline_admission_request` | 400 | Producer | Unknown field, bad type/unit, duplicate member or wide payload |
| `unauthenticated` | 401 | Both | Caller identity or proof absent/invalid |
| `authorization_denied` | 403 | Both | Generic denial without a narrower reason |

## Rules

- Unknown `errorCode`/`reasonCode` rejects. Do not infer a reason from `message`.
- A signature failure is authentication (`401`), not policy denial (`403`).
- Unavailable revocation is `503`, never a cached or inferred allow.
- `offline_grant_recovery_required` blocks affected admissions until the local
  ledger is authoritatively resolved; it does not authorize replay or refund.
- The same semantic failure keeps the same `reasonCode` across producer,
  consumer and local verification. HTTP status applies only where a transport
  exists; a disconnected consumer may surface the reason locally.
- Adding a new reason code requires a reviewed contract revision; removing or
  renaming one is breaking.

Fixtures: `fixtures/valid/error.json` and
`fixtures/invalid/error-unknown-reason.json`. The contract test enforces the
closed enum.
