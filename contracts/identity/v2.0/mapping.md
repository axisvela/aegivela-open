# Identity Resolve Contract Mapping

`apiVersion` is the literal `aegivela.io/v2.0`; the protobuf service is `IdentityBridgeService.Resolve`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `identityResolveRequest.apiVersion` | `IdentityResolveRequest.apiVersion` | `IdentityResolveRequest.apiVersion` = 1 |
| bearer token | `identityResolveRequest.bearerToken` | `IdentityResolveRequest.bearerToken` | `IdentityResolveRequest.bearerToken` = 2 |
| audience | `identityResolveRequest.audience` | `IdentityResolveRequest.audience` | `IdentityResolveRequest.audience` = 3 |
| trace ID | `identityResolveRequest.traceId` | `IdentityResolveRequest.traceId` | `IdentityResolveRequest.traceId` = 4 |
| response API version | `identityResolveResponse.apiVersion` | `IdentityResolveResponse.apiVersion` | `IdentityResolveResponse.apiVersion` = 1 |
| principal | `identityResolveResponse.principal` | `IdentityResolveResponse.principal` | `IdentityResolveResponse.principal` = 2 |
| tenant ID | `principal.tenantId` | `Principal.tenantId` | `Principal.tenantId` = 1 |
| subject reference | `principal.subjectRef` | `Principal.subjectRef` | `Principal.subjectRef` = 2 |
| actor reference | `principal.actorRef` | `Principal.actorRef` | `Principal.actorRef` = 3 |
| client ID | `principal.clientId` | `Principal.clientId` | `Principal.clientId` = 4 |
| workload reference | `principal.workloadRef` | `Principal.workloadRef` | `Principal.workloadRef` = 5 |
| agent ID | `principal.agentId` | `Principal.agentId` | `Principal.agentId` = 6 |
| agent class | `principal.agentClass` | `Principal.agentClass` | `Principal.agentClass` = 7 |
| master ID | `principal.masterId` | `Principal.masterId` | `Principal.masterId` = 8 |
| organization authority root | `principal.organizationAuthorityRoot` | `Principal.organizationAuthorityRoot` | `Principal.organizationAuthorityRoot` = 9 |
| lifecycle epoch | `principal.lifecycleEpoch` | `Principal.lifecycleEpoch` | `Principal.lifecycleEpoch` = 10 |
| attestation reference | `principal.attestationRef` | `Principal.attestationRef` | `Principal.attestationRef` = 11 |
| namespace | `principal.namespace` | `Principal.namespace` | `Principal.namespace` = 12 |
| authority binding | `principal.authorityBinding` | `Principal.authorityBinding` | `Principal.authorityBinding` = 13 |
| agent epoch | `principal.agentEpoch` | `Principal.agentEpoch` | `Principal.agentEpoch` = 14 |
| identity epoch | `principal.identityEpoch` | `Principal.identityEpoch` | `Principal.identityEpoch` = 15 |
| instance ID | `principal.instanceId` | `Principal.instanceId` | `Principal.instanceId` = 16 |

`namespace`, `authorityBinding`, `agentEpoch`, `identityEpoch`, and `instanceId` are the NOMIVELA/EIDOVELA registry-consumer fields. They are optional on the wire and populated only when the credential was resolved through an identity source that carries the dual lifecycle (for example EIDOVELA). A context missing dual state, dual epoch, `attestationRef`, or `credentialGeneration` is rejected upstream and never reaches this response.

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalidIdentityRequest` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 identityUnavailable` | `UNAVAILABLE` | An identity cannot be safely resolved; consumers must fail closed. |
