# Gateway Connect Contract Mapping v1alpha1

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `GatewayService.Connect`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `connect_request.api_version` | `GatewayConnectRequest.api_version` | `GatewayConnectRequest.api_version` = 1 |
| action | `connect_request.action` | `GatewayConnectRequest.action` | `GatewayConnectRequest.action` = 2 |
| tenant ID | `connect_request.tenant_id` | `GatewayConnectRequest.tenant_id` | `GatewayConnectRequest.tenant_id` = 3 |
| agent ID | `connect_request.agent_id` | `GatewayConnectRequest.agent_id` | `GatewayConnectRequest.agent_id` = 4 |
| agent class | `connect_request.agent_class` | `GatewayConnectRequest.agent_class` | `GatewayConnectRequest.agent_class` = 5 |
| workload assertion | `connect_request.workload_assertion` | `GatewayConnectRequest.workload_assertion` | `GatewayConnectRequest.workload_assertion` = 6 |
| tool attribution | `connect_request.attribution` | `GatewayConnectRequest.attribution` | `GatewayConnectRequest.attribution` = 7 |
| argument digest | `connect_request.argument_digest` | `GatewayConnectRequest.argument_digest` | `GatewayConnectRequest.argument_digest` = 8 |
| argument classification | `connect_request.argument_classification` | `GatewayConnectRequest.argument_classification` | `GatewayConnectRequest.argument_classification` = 9 |
| target host | `connect_request.target_host` | `GatewayConnectRequest.target_host` | `GatewayConnectRequest.target_host` = 10 |
| target scheme | `connect_request.target_scheme` | `GatewayConnectRequest.target_scheme` | `GatewayConnectRequest.target_scheme` = 11 |
| target path | `connect_request.target_path` | `GatewayConnectRequest.target_path` | `GatewayConnectRequest.target_path` = 12 |
| requested scope | `connect_request.requested_scope` | `GatewayConnectRequest.requested_scope` | `GatewayConnectRequest.requested_scope` = 13 |
| audience | `connect_request.audience` | `GatewayConnectRequest.audience` | `GatewayConnectRequest.audience` = 14 |
| trace ID | `connect_request.trace_id` | `GatewayConnectRequest.trace_id` | `GatewayConnectRequest.trace_id` = 15 |
| response API version | `connect_response.api_version` | `GatewayConnectResponse.api_version` | `GatewayConnectResponse.api_version` = 1 |
| decision ID | `connect_response.decision_id` | `GatewayConnectResponse.decision_id` | `GatewayConnectResponse.decision_id` = 2 |
| outcome | `connect_response.outcome` | `GatewayConnectResponse.outcome` | `GatewayConnectResponse.outcome` = 3 |
| policy version | `connect_response.policy_version` | `GatewayConnectResponse.policy_version` | `GatewayConnectResponse.policy_version` = 4 |
| expires at | `connect_response.expires_at` | `GatewayConnectResponse.expires_at` | `GatewayConnectResponse.expires_at` = 5 |
| TTL seconds | `connect_response.ttl_seconds` | `GatewayConnectResponse.ttl_seconds` | `GatewayConnectResponse.ttl_seconds` = 6 |
| offline eligible | `connect_response.offline_eligible` | `GatewayConnectResponse.offline_eligible` | `GatewayConnectResponse.offline_eligible` = 7 |
| scope | `connect_response.scope` | `GatewayConnectResponse.scope` | `GatewayConnectResponse.scope` = 8 |
| obligations | `connect_response.obligations` | `GatewayConnectResponse.obligations` | `GatewayConnectResponse.obligations` = 9 |
| evidence refs | `connect_response.evidence_refs` | `GatewayConnectResponse.evidence_refs` | `GatewayConnectResponse.evidence_refs` = 10 |
| credential ref | `connect_response.credential_ref` | `GatewayConnectResponse.credential_ref` | `GatewayConnectResponse.credential_ref` = 11 |
| credential class | `connect_response.credential_class` | `GatewayConnectResponse.credential_class` | `GatewayConnectResponse.credential_class` = 12 |
| MITM required | `connect_response.mitm_required` | `GatewayConnectResponse.mitm_required` | `GatewayConnectResponse.mitm_required` = 13 |
| MITM scope | `connect_response.mitm_scope` | `GatewayConnectResponse.mitm_scope` | `GatewayConnectResponse.mitm_scope` = 14 |
| allowed target paths | `connect_response.allowed_target_paths` | `GatewayConnectResponse.allowed_target_paths` | `GatewayConnectResponse.allowed_target_paths` = 15 |

Actions are `connection:open`, `backend:request`, `credential:inject`, and `tool:invoke`. Agent classes are `twin` and `service`. The endpoint resolves Twin and Service Agent authorities for agent-classed callers and non-agent (`system_api`) callers by workload assertion; caller-supplied tenant and agent identifiers are comparison-only. `tool:invoke` requires a verified agent authority whose attribution agent reference matches the verified agent identity.

The endpoint requires the `X-Aegivela-Internal-Token` header.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalid_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `403 authorization_denied` | `PERMISSION_DENIED` | The decision is denied or revoked. |
| `503 authorization_unconfigured` | `UNAVAILABLE` | An authorization dependency is unavailable; consumers must fail closed. |
