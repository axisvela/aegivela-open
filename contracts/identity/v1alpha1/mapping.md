# Identity Resolve Contract Mapping

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `IdentityBridgeService.Resolve`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `identity_resolve_request.api_version` | `IdentityResolveRequest.api_version` | `IdentityResolveRequest.api_version` = 1 |
| bearer token | `identity_resolve_request.bearer_token` | `IdentityResolveRequest.bearer_token` | `IdentityResolveRequest.bearer_token` = 2 |
| audience | `identity_resolve_request.audience` | `IdentityResolveRequest.audience` | `IdentityResolveRequest.audience` = 3 |
| trace ID | `identity_resolve_request.trace_id` | `IdentityResolveRequest.trace_id` | `IdentityResolveRequest.trace_id` = 4 |
| response API version | `identity_resolve_response.api_version` | `IdentityResolveResponse.api_version` | `IdentityResolveResponse.api_version` = 1 |
| principal | `identity_resolve_response.principal` | `IdentityResolveResponse.principal` | `IdentityResolveResponse.principal` = 2 |
| tenant ID | `principal.tenant_id` | `Principal.tenant_id` | `Principal.tenant_id` = 1 |
| subject reference | `principal.subject_ref` | `Principal.subject_ref` | `Principal.subject_ref` = 2 |
| actor reference | `principal.actor_ref` | `Principal.actor_ref` | `Principal.actor_ref` = 3 |
| client ID | `principal.client_id` | `Principal.client_id` | `Principal.client_id` = 4 |
| workload reference | `principal.workload_ref` | `Principal.workload_ref` | `Principal.workload_ref` = 5 |
| agent ID | `principal.agent_id` | `Principal.agent_id` | `Principal.agent_id` = 6 |
| agent class | `principal.agent_class` | `Principal.agent_class` | `Principal.agent_class` = 7 |
| master ID | `principal.master_id` | `Principal.master_id` | `Principal.master_id` = 8 |
| organization authority root | `principal.organization_authority_root` | `Principal.organization_authority_root` | `Principal.organization_authority_root` = 9 |
| lifecycle epoch | `principal.lifecycle_epoch` | `Principal.lifecycle_epoch` | `Principal.lifecycle_epoch` = 10 |
| attestation reference | `principal.attestation_ref` | `Principal.attestation_ref` | `Principal.attestation_ref` = 11 |

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalid_identity_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 identity_unavailable` | `UNAVAILABLE` | An identity cannot be safely resolved; consumers must fail closed. |
