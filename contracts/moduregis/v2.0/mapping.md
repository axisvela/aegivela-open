# MODUREGIS Adapter Contract Mapping v1alpha1

`apiVersion` is the literal `aegivela.io/v2.0`; the protobuf service is `ModuregisAdapter.Authorize`.

This contract defines the four MODUREGIS Capability Governance profile actions and their required fields per action class. The adapter unconditionally maps an unconfigured or DenyAll state to `503 authorizationUnconfigured`. Only a validated `allow` outcome with non-empty `tenantId`, `actorId`, and `policyVersion` satisfies the MODUREGIS port.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `adapterAuthorizationRequest.apiVersion` | `AdapterAuthorizationRequest.apiVersion` | `AdapterAuthorizationRequest.apiVersion` = 1 |
| bearer token | `adapterAuthorizationRequest.bearerToken` | `AdapterAuthorizationRequest.bearerToken` | `AdapterAuthorizationRequest.bearerToken` = 2 |
| action | `adapterAuthorizationRequest.action` | `AdapterAuthorizationRequest.action` | `AdapterAuthorizationRequest.action` = 3 |
| resource | `adapterAuthorizationRequest.resource` | `AdapterAuthorizationRequest.resource` | `AdapterAuthorizationRequest.resource` = 4 |
| resource kind | `resource.kind` | `ResourceRef.kind` | `ResourceRef.kind` = 1 |
| resource reference | `resource.reference` | `ResourceRef.reference` | `ResourceRef.reference` = 2 |
| scope | `adapterAuthorizationRequest.scope` | `AdapterAuthorizationRequest.scope` | `AdapterAuthorizationRequest.scope` = 5 |
| approval artifact | `adapterAuthorizationRequest.approvalArtifact` | `AdapterAuthorizationRequest.approvalArtifact` | `AdapterAuthorizationRequest.approvalArtifact` = 6 |
| adapter ID | `adapterAuthorizationRequest.adapterId` | `AdapterAuthorizationRequest.adapterId` | `AdapterAuthorizationRequest.adapterId` = 7 |
| adapter version | `adapterAuthorizationRequest.adapterVersion` | `AdapterAuthorizationRequest.adapterVersion` | `AdapterAuthorizationRequest.adapterVersion` = 8 |
| execution ID | `adapterAuthorizationRequest.executionId` | `AdapterAuthorizationRequest.executionId` | `AdapterAuthorizationRequest.executionId` = 9 |
| trace ID | `adapterAuthorizationRequest.traceId` | `AdapterAuthorizationRequest.traceId` | `AdapterAuthorizationRequest.traceId` = 10 |
| response API version | `adapterAuthorizationResponse.apiVersion` | `AdapterAuthorizationResponse.apiVersion` | `AdapterAuthorizationResponse.apiVersion` = 1 |
| decision ID | `adapterAuthorizationResponse.decisionId` | `AdapterAuthorizationResponse.decisionId` | `AdapterAuthorizationResponse.decisionId` = 2 |
| outcome | `adapterAuthorizationResponse.outcome` | `AdapterAuthorizationResponse.outcome` | `AdapterAuthorizationResponse.outcome` = 3 |
| policy version | `adapterAuthorizationResponse.policyVersion` | `AdapterAuthorizationResponse.policyVersion` | `AdapterAuthorizationResponse.policyVersion` = 4 |
| tenant ID | `adapterAuthorizationResponse.tenantId` | `AdapterAuthorizationResponse.tenantId` | `AdapterAuthorizationResponse.tenantId` = 5 |
| actor ID | `adapterAuthorizationResponse.actorId` | `AdapterAuthorizationResponse.actorId` | `AdapterAuthorizationResponse.actorId` = 6 |
| agent ID | `adapterAuthorizationResponse.agentId` | `AdapterAuthorizationResponse.agentId` | `AdapterAuthorizationResponse.agentId` = 7 |
| master ID | `adapterAuthorizationResponse.masterId` | `AdapterAuthorizationResponse.masterId` | `AdapterAuthorizationResponse.masterId` = 8 |
| workload ID | `adapterAuthorizationResponse.workloadId` | `AdapterAuthorizationResponse.workloadId` | `AdapterAuthorizationResponse.workloadId` = 9 |
| subject ref | `adapterAuthorizationResponse.subjectRef` | `AdapterAuthorizationResponse.subjectRef` | `AdapterAuthorizationResponse.subjectRef` = 10 |
| expires at | `adapterAuthorizationResponse.expiresAt` | `AdapterAuthorizationResponse.expiresAt` | `AdapterAuthorizationResponse.expiresAt` = 11 |
| evidence refs | `adapterAuthorizationResponse.evidenceRefs` | `AdapterAuthorizationResponse.evidenceRefs` | `AdapterAuthorizationResponse.evidenceRefs` = 12 |
| grant token | `adapterAuthorizationResponse.grantToken` | `AdapterAuthorizationResponse.grantToken` | `AdapterAuthorizationResponse.grantToken` = 13 |

## Per-Action Required Fields

| Action | Additionally required fields | Returns |
| --- | --- | --- |
| `capability:read` | none beyond base | Principal fields |
| `capability:publish` | `scope`, `approvalArtifact` | Principal fields; `approvalRequired` is governed state, resubmitted with approval reference |
| `adapter:activate` | `adapterId`, `adapterVersion` | Principal fields; activation only with verified attestation evidence and an allowed decision for the exact adapter version |
| `capability:invoke` | `scope`, `executionId` | Principal fields plus `grantToken` (short-lived, audience-bound execution grant) |

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | Invalid, expired, wrong-issuer, or wrong-audience bearer artifact |
| `403 authorizationDenied` | `PERMISSION_DENIED` | Policy denial, scope attenuation failure, revoked, expired approval, or binding mismatch |
| `503 authorizationUnconfigured` | `UNAVAILABLE` | Adapter is not configured (DenyAll) or AEGIVELA dependency is unavailable |

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Module Publication

The schema and fixtures in this directory ship byte-identically inside the
Go module as `backend/pepsdk/moduregiscontract` (embedded under `files/`),
kept in parity by `syncTest.go`. Downstream repositories (MODUREGIS)
consume the contract by requiring `github.com/axisrobo/aegivela/backend` at
a `backend/v*` module tag and must not copy these files. Changes here
require re-running the parity sync (copy into
`backend/pepsdk/moduregiscontract/files/`) and a new module tag before
downstream consumers pick them up.
