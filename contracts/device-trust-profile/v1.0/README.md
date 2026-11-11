# Device Trust Profile Management Contract (v1.0)

**Status: PUBLISHED at `aegivela-open` `v1.2.13` (2026-11-11).**

This is the management surface for the device trust profile object defined by
`contracts/offline-admission/v1.0/device-trust-profile.schema.json`. It does not
replace or mutate that object schema.

| File | Purpose |
| --- | --- |
| `management.schema.json` | Create-request, record and error schemas |
| `management.openapi.yaml` | List/create/get/approve/retire paths |
| `fixtures/` | Valid and invalid examples |

The runtime management API is default-off and requires the internal token,
tenant/actor headers, a PostgreSQL store and an evidence recorder. Publication
does not authorize deployment or enablement.

Publication approval: owner direction, 2026-11-11. This is a document and
publication record, not a cryptographic signature or a hardware trust proof.
