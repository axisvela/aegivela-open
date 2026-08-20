# MODUREGIS Adapter Contract Mapping v1alpha1

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `ModuregisAdapter.Authorize`.

This contract defines the four MODUREGIS Capability Governance profile actions and their required fields per action class. The adapter unconditionally maps an unconfigured or DenyAll state to `503 authorization_unconfigured`. Only a validated `allow` outcome with non-empty `tenant_id`, `actor_id`, and `policy_version` satisfies the MODUREGIS port.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `adapter_authorization_request.api_version` | `AdapterAuthorizationRequest.api_version` | `AdapterAuthorizationRequest.api_version` = 1 |
| bearer token | `adapter_authorization_request.bearer_token` | `AdapterAuthorizationRequest.bearer_token` | `AdapterAuthorizationRequest.bearer_token` = 2 |
| action | `adapter_authorization_request.action` | `AdapterAuthorizationRequest.action` | `AdapterAuthorizationRequest.action` = 3 |
| resource | `adapter_authorization_request.resource` | `AdapterAuthorizationRequest.resource` | `AdapterAuthorizationRequest.resource` = 4 |
| resource kind | `resource.kind` | `ResourceRef.kind` | `ResourceRef.kind` = 1 |
| resource reference | `resource.reference` | `ResourceRef.reference` | `ResourceRef.reference` = 2 |
| scope | `adapter_authorization_request.scope` | `AdapterAuthorizationRequest.scope` | `AdapterAuthorizationRequest.scope` = 5 |
| approval artifact | `adapter_authorization_request.approval_artifact` | `AdapterAuthorizationRequest.approval_artifact` | `AdapterAuthorizationRequest.approval_artifact` = 6 |
| adapter ID | `adapter_authorization_request.adapter_id` | `AdapterAuthorizationRequest.adapter_id` | `AdapterAuthorizationRequest.adapter_id` = 7 |
| adapter version | `adapter_authorization_request.adapter_version` | `AdapterAuthorizationRequest.adapter_version` | `AdapterAuthorizationRequest.adapter_version` = 8 |
| execution ID | `adapter_authorization_request.execution_id` | `AdapterAuthorizationRequest.execution_id` | `AdapterAuthorizationRequest.execution_id` = 9 |
| trace ID | `adapter_authorization_request.trace_id` | `AdapterAuthorizationRequest.trace_id` | `AdapterAuthorizationRequest.trace_id` = 10 |
| response API version | `adapter_authorization_response.api_version` | `AdapterAuthorizationResponse.api_version` | `AdapterAuthorizationResponse.api_version` = 1 |
| decision ID | `adapter_authorization_response.decision_id` | `AdapterAuthorizationResponse.decision_id` | `AdapterAuthorizationResponse.decision_id` = 2 |
| outcome | `adapter_authorization_response.outcome` | `AdapterAuthorizationResponse.outcome` | `AdapterAuthorizationResponse.outcome` = 3 |
| policy version | `adapter_authorization_response.policy_version` | `AdapterAuthorizationResponse.policy_version` | `AdapterAuthorizationResponse.policy_version` = 4 |
| tenant ID | `adapter_authorization_response.tenant_id` | `AdapterAuthorizationResponse.tenant_id` | `AdapterAuthorizationResponse.tenant_id` = 5 |
| actor ID | `adapter_authorization_response.actor_id` | `AdapterAuthorizationResponse.actor_id` | `AdapterAuthorizationResponse.actor_id` = 6 |
| agent ID | `adapter_authorization_response.agent_id` | `AdapterAuthorizationResponse.agent_id` | `AdapterAuthorizationResponse.agent_id` = 7 |
| master ID | `adapter_authorization_response.master_id` | `AdapterAuthorizationResponse.master_id` | `AdapterAuthorizationResponse.master_id` = 8 |
| workload ID | `adapter_authorization_response.workload_id` | `AdapterAuthorizationResponse.workload_id` | `AdapterAuthorizationResponse.workload_id` = 9 |
| subject ref | `adapter_authorization_response.subject_ref` | `AdapterAuthorizationResponse.subject_ref` | `AdapterAuthorizationResponse.subject_ref` = 10 |
| expires at | `adapter_authorization_response.expires_at` | `AdapterAuthorizationResponse.expires_at` | `AdapterAuthorizationResponse.expires_at` = 11 |
| evidence refs | `adapter_authorization_response.evidence_refs` | `AdapterAuthorizationResponse.evidence_refs` | `AdapterAuthorizationResponse.evidence_refs` = 12 |
| grant token | `adapter_authorization_response.grant_token` | `AdapterAuthorizationResponse.grant_token` | `AdapterAuthorizationResponse.grant_token` = 13 |

## Per-Action Required Fields

| Action | Additionally required fields | Returns |
| --- | --- | --- |
| `capability:read` | none beyond base | Principal fields |
| `capability:publish` | `scope`, `approval_artifact` | Principal fields; `approval_required` is governed state, resubmitted with approval reference |
| `adapter:activate` | `adapter_id`, `adapter_version` | Principal fields; activation only with verified attestation evidence and an allowed decision for the exact adapter version |
| `capability:invoke` | `scope`, `execution_id` | Principal fields plus `grant_token` (short-lived, audience-bound execution grant) |

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | Invalid, expired, wrong-issuer, or wrong-audience bearer artifact |
| `403 authorization_denied` | `PERMISSION_DENIED` | Policy denial, scope attenuation failure, revoked, expired approval, or binding mismatch |
| `503 authorization_unconfigured` | `UNAVAILABLE` | Adapter is not configured (DenyAll) or AEGIVELA dependency is unavailable |

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Module Publication

The schema and fixtures in this directory ship byte-identically inside the
Go module as `backend/pepsdk/moduregiscontract` (embedded under `files/`),
kept in parity by `sync_test.go`. Downstream repositories (MODUREGIS)
consume the contract by requiring `github.com/axisrobo/aegivela/backend` at
a `backend/v*` module tag and must not copy these files. Changes here
require re-running the parity sync (copy into
`backend/pepsdk/moduregiscontract/files/`) and a new module tag before
downstream consumers pick them up.
