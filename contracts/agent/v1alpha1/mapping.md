# Agent Lifecycle Mapping

`api_version` is the literal `aegivela.io/v1alpha1`. Every request is authenticated with the `X-AEGIVELA-PEP` internal header, and the trusted authority context (`tenant`, `actor`, `authority root`) is supplied by headers, never by the request body. A request that carries caller-supplied authority fields is rejected.

| Semantic field | JSON Schema | OpenAPI | Protobuf |
|---|---|---|---|
| register request | `agent_register_request` | `AgentRegisterRequest` | `AgentRegisterRequest` |
| agent record response | `agent_record` | `AgentRecord` | `AgentRecord` |
| get request | `agent_get_request` | path parameter `agent_id` | `AgentGetRequest` |
| enroll request | `agent_enroll_request` | `AgentEnrollRequest` | `AgentEnrollRequest` |
| enroll response | `agent_enroll_response` | `AgentEnrollResponse` | `AgentEnrollResponse` |
| agent class | `agent_class` enum | `AgentClass` | `AgentClass` enum |
| lifecycle state | `lifecycle_state` enum (`created`, `active`, `suspended`, `revoked`) | `LifecycleState` | `LifecycleState` enum |
| enrollment strategy | `enrollment_strategy` enum (`device_backed`, `human_authorized`, `api_only`) | `EnrollmentStrategy` | `EnrollmentStrategy` enum |

## Routes

| Route | Body | Effect |
|---|---|---|
| `POST /v1/agents` | `AgentRegisterRequest` | Registers a `created` agent. |
| `GET /v1/agents/{agent_id}` | none | Returns the agent record. |
| `POST /v1/agents/{agent_id}/enroll` | `AgentEnrollRequest` | Verifies the enrollment/attestation evidence prescribed by the active profile and activates the agent. The only path from `created` to `active`. |
| `POST /v1/agents/{agent_id}/activate` | none | Restores a suspended agent. A `created` (not-yet-enrolled) agent returns `409 enrollment_required`; activation is never inferred from caller-supplied identity fields. |
| `POST /v1/agents/{agent_id}/suspend` | none | Suspends an active agent. |
| `POST /v1/agents/{agent_id}/revoke` | none | Revokes an agent. |

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `400 invalid_agent_request` / `invalid_enrollment_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorization_denied` / `enrollment_denied` | `PERMISSION_DENIED` | The authority or enrollment evidence did not match. |
| `404 agent_not_found` | `NOT_FOUND` | The agent does not exist for the caller's tenant and authority. |
| `409 agent_exists` / `invalid_lifecycle_transition` / `enrollment_required` | `FAILED_PRECONDITION` | The lifecycle transition is not permitted, or enrollment is required first. |
| `503 identity_store_unavailable` / `identity_unavailable` | `UNAVAILABLE` | A dependency is unavailable; consumers must fail closed. |
