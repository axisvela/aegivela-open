# Offline Admission Contract Draft (v1.0)

**Status: DRAFT — NOT PUBLISHED.**

This directory is a local machine-contract draft for bounded offline admission.
It is approved only for drafting by the user on 2026-11-09 (M1-M8 proposals).
It is not in the `scripts/sync-oss.ps1` publication list, is not an enabled wire
contract, and does not authorize runtime implementation, device acceptance or
offline execution.

| File | Purpose |
| --- | --- |
| `offline-admission.schema.json` | Authoritative draft JSON Schema for the artifact and API bodies |
| `offline-admission.openapi.yaml` | Proposed HTTP paths (not implemented) |
| `offline-admission.proto` | Draft protobuf mapping; generated binding under `backend/internal/offlineadmissioncontract/pb` |
| `mapping.md` | Field mapping and the approved M1-M8 decisions |
| `vectors/` | Golden canonical bytes, signing input, digests, signature/JWKS vectors and parent token |
| `fixtures/valid` | Structurally valid examples |
| `fixtures/invalid` | Structurally invalid examples |
| `fixtures/semantic-invalid` | Schema-valid examples that violate the approved boundary |

Fixed strings and canonicalization are documented in
`docs/api/offline-admission-canonicalization.md`; stable errors are in
`docs/api/offline-admission-errors.md`.

## Approved Boundary

- Parent is a verified execution grant with canonical signed-allow lineage.
- Audience equals the parent's and is already device-bound.
- Lifetime is at most 900 seconds and never exceeds the parent.
- Authority is reserved online; no offline issuance or re-delegation.
- Unknown clock/key/trust or unproven rollback/clone resistance rejects.

## Local Validation

```powershell
go test -count=1 ./backend/internal/offlineadmissioncontract
& .\tools\buf.ps1 lint
```

The tests validate fixtures, artifact invariants and the draft file set. They do
not prove producer reservation, consumer recovery, device trust or publication.

## Publication Gates

Before this becomes a published contract: remaining resource/device/identity/
consumer review, exact wire version/path assignment, canonical-byte golden
vectors, producer/consumer PostgreSQL and fault-injection acceptance, and an
addition to the OSS sync list with reviewed public verification. F4b remains
blocked until then.
