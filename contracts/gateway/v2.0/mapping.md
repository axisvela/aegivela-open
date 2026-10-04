# Gateway Connect Contract Mapping v1alpha1

`apiVersion` is the literal `aegivela.io/v2.0`; the protobuf service is `GatewayService.Connect`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `connectRequest.apiVersion` | `GatewayConnectRequest.apiVersion` | `GatewayConnectRequest.apiVersion` = 1 |
| action | `connectRequest.action` | `GatewayConnectRequest.action` | `GatewayConnectRequest.action` = 2 |
| authorization mode | `connectRequest.authorizationMode` | `GatewayConnectRequest.authorizationMode` | `GatewayConnectRequest.authorizationMode` = 16 |
| tenant ID | `connectRequest.tenantId` | `GatewayConnectRequest.tenantId` | `GatewayConnectRequest.tenantId` = 3 |
| agent ID | `connectRequest.agentId` | `GatewayConnectRequest.agentId` | `GatewayConnectRequest.agentId` = 4 |
| agent class | `connectRequest.agentClass` | `GatewayConnectRequest.agentClass` | `GatewayConnectRequest.agentClass` = 5 |
| workload assertion | `connectRequest.workloadAssertion` | `GatewayConnectRequest.workloadAssertion` | `GatewayConnectRequest.workloadAssertion` = 6 |
| tool attribution | `connectRequest.attribution` | `GatewayConnectRequest.attribution` | `GatewayConnectRequest.attribution` = 7 |
| argument digest | `connectRequest.argumentDigest` | `GatewayConnectRequest.argumentDigest` | `GatewayConnectRequest.argumentDigest` = 8 |
| argument classification | `connectRequest.argumentClassification` | `GatewayConnectRequest.argumentClassification` | `GatewayConnectRequest.argumentClassification` = 9 |
| target host | `connectRequest.targetHost` | `GatewayConnectRequest.targetHost` | `GatewayConnectRequest.targetHost` = 10 |
| target scheme | `connectRequest.targetScheme` | `GatewayConnectRequest.targetScheme` | `GatewayConnectRequest.targetScheme` = 11 |
| target path | `connectRequest.targetPath` | `GatewayConnectRequest.targetPath` | `GatewayConnectRequest.targetPath` = 12 |
| requested scope | `connectRequest.requestedScope` | `GatewayConnectRequest.requestedScope` | `GatewayConnectRequest.requestedScope` = 13 |
| audience | `connectRequest.audience` | `GatewayConnectRequest.audience` | `GatewayConnectRequest.audience` = 14 |
| trace ID | `connectRequest.traceId` | `GatewayConnectRequest.traceId` | `GatewayConnectRequest.traceId` = 15 |
| response API version | `connectResponse.apiVersion` | `GatewayConnectResponse.apiVersion` | `GatewayConnectResponse.apiVersion` = 1 |
| decision ID | `connectResponse.decisionId` | `GatewayConnectResponse.decisionId` | `GatewayConnectResponse.decisionId` = 2 |
| outcome | `connectResponse.outcome` | `GatewayConnectResponse.outcome` | `GatewayConnectResponse.outcome` = 3 |
| policy version | `connectResponse.policyVersion` | `GatewayConnectResponse.policyVersion` | `GatewayConnectResponse.policyVersion` = 4 |
| expires at | `connectResponse.expiresAt` | `GatewayConnectResponse.expiresAt` | `GatewayConnectResponse.expiresAt` = 5 |
| TTL seconds | `connectResponse.ttlSeconds` | `GatewayConnectResponse.ttlSeconds` | `GatewayConnectResponse.ttlSeconds` = 6 |
| offline eligible | `connectResponse.offlineEligible` | `GatewayConnectResponse.offlineEligible` | `GatewayConnectResponse.offlineEligible` = 7 |
| scope | `connectResponse.scope` | `GatewayConnectResponse.scope` | `GatewayConnectResponse.scope` = 8 |
| obligations | `connectResponse.obligations` | `GatewayConnectResponse.obligations` | `GatewayConnectResponse.obligations` = 9 |
| evidence refs | `connectResponse.evidenceRefs` | `GatewayConnectResponse.evidenceRefs` | `GatewayConnectResponse.evidenceRefs` = 10 |
| credential ref | `connectResponse.credentialRef` | `GatewayConnectResponse.credentialRef` | `GatewayConnectResponse.credentialRef` = 11 |
| credential class | `connectResponse.credentialClass` | `GatewayConnectResponse.credentialClass` | `GatewayConnectResponse.credentialClass` = 12 |
| MITM required | `connectResponse.mitmRequired` | `GatewayConnectResponse.mitmRequired` | `GatewayConnectResponse.mitmRequired` = 13 |
| MITM scope | `connectResponse.mitmScope` | `GatewayConnectResponse.mitmScope` | `GatewayConnectResponse.mitmScope` = 14 |
| allowed target paths | `connectResponse.allowedTargetPaths` | `GatewayConnectResponse.allowedTargetPaths` | `GatewayConnectResponse.allowedTargetPaths` = 15 |

Actions are `connection:open`, `backend:request`, `credential:inject`, and `tool:invoke`. Agent classes are `twin` and `service`. `authorizationMode` is `systemApi`, `serviceAgentApi`, or `twinAgentApi`; it binds the workload-assertion artifact to the caller's authority surface and must agree with the agent fields (`systemApi` carries no agent authority; the agent modes require the matching `agentClass` and `agentId`). A contradiction is a cross-mode substitution and is rejected with `400 invalidRequest` before principal resolution and policy evaluation. The endpoint resolves Twin and Service Agent authorities for agent-classed callers and non-agent (`systemApi`) callers by workload assertion; caller-supplied tenant and agent identifiers are comparison-only. `tool:invoke` requires a verified agent authority whose attribution agent reference matches the verified agent identity.

The endpoint requires the `X-Aegivela-Internal-Token` header.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalidRequest` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `403 authorizationDenied` | `PERMISSION_DENIED` | The decision is denied or revoked. |
| `503 authorizationUnconfigured` | `UNAVAILABLE` | An authorization dependency is unavailable; consumers must fail closed. |
