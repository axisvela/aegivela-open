# Offline Admission Contract (v1.0)

**Status: PUBLISHED at `aegivela-open` `1b101f2`.**

This directory is the v1.0 machine contract for bounded offline admission. It was
approved for drafting by the user on 2026-11-09 (M1-M8 proposals) and published
as G1 (`684886a`, clarified at `8c0c69f`, issue audience `e5f1645`, task binding
`1b101f2`). It is in the `scripts/sync-oss.ps1` publication list. It is not an
enabled wire surface and does not authorize runtime deployment, device acceptance
or offline execution. The embedded schema/proto/OpenAPI titles and vector status
fields still carry the `DRAFT - NOT PUBLISHED` maturity marker from the published
`1b101f2` revision; correcting them requires a reviewed contract revision and
publication, not an in-place edit.

| File | Purpose |
| --- | --- |
| `offline-admission.schema.json` | Authoritative v1.0 JSON Schema for the artifact and API bodies |
| `offline-admission.openapi.yaml` | HTTP paths (implemented behind a default-off mount) |
| `offline-admission.proto` | v1.0 protobuf mapping; generated binding under `backend/internal/offlineadmissioncontract/pb` |
| `mapping.md` | Field mapping and the approved M1-M8 decisions |
| `vectors/` | Golden canonical bytes, signing input, digests, signature/JWKS vectors and parent token |
| `fixtures/valid` | Structurally valid examples |
| `fixtures/invalid` | Structurally invalid examples |
| `fixtures/semantic-invalid` | Schema-valid examples that violate the approved boundary |

Device trust profile management is a separate draft surface under
`contracts/device-trust-profile/v1.0/`. It mirrors the published
`device-trust-profile.schema.json` content fields; the management contract test
enforces required-field and descriptive-validator parity. It does not mutate the
published v1.0 artifact bytes or import the object schema by `$ref`.

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

The tests validate fixtures, artifact invariants and the contract file set. They
do not prove producer reservation, consumer recovery or device trust.

## Remaining Gates

The contract and its vectors are published. Before deployment or offline
enablement: remaining resource/device/identity/consumer review, a device hardware
proof, producer/consumer PostgreSQL and fault-injection acceptance, and an
approved device trust profile plus evidence recorder at assembly startup. F4b
remains blocked until then.
