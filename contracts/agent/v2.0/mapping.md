# Agent Lifecycle Mapping

`apiVersion` is the literal `aegivela.io/v2.0`. Every request is authenticated with the `X-AEGIVELA-PEP` internal header, and the trusted authority context (`tenant`, `actor`, `authority root`) is supplied by headers, never by the request body. A request that carries caller-supplied authority fields is rejected.

| Semantic field | JSON Schema | OpenAPI | Protobuf |
|---|---|---|---|
| register request | `agentRegisterRequest` | `AgentRegisterRequest` | `AgentRegisterRequest` |
| agent record response | `agentRecord` | `AgentRecord` | `AgentRecord` |
| get request | `agentGetRequest` | path parameter `agentId` | `AgentGetRequest` |
| enroll request | `agentEnrollRequest` | `AgentEnrollRequest` | `AgentEnrollRequest` |
| enroll response | `agentEnrollResponse` | `AgentEnrollResponse` | `AgentEnrollResponse` |
| agent class | `agentClass` enum | `AgentClass` | `AgentClass` enum |
| lifecycle state | `lifecycleState` enum (`created`, `active`, `suspended`, `revoked`) | `LifecycleState` | `LifecycleState` enum |
| enrollment strategy | `enrollmentStrategy` enum (`deviceBacked`, `humanAuthorized`, `apiOnly`) | `EnrollmentStrategy` | `EnrollmentStrategy` enum |

## Routes

| Route | Body | Effect |
|---|---|---|
| `POST /v1/agents` | `AgentRegisterRequest` | Registers a `created` agent. |
| `GET /v1/agents/{agentId}` | none | Returns the agent record. |
| `POST /v1/agents/{agentId}/enroll` | `AgentEnrollRequest` | Verifies the enrollment/attestation evidence prescribed by the active profile and activates the agent. The only path from `created` to `active`. |
| `POST /v1/agents/{agentId}/activate` | none | Restores a suspended agent. A `created` (not-yet-enrolled) agent returns `409 enrollmentRequired`; activation is never inferred from caller-supplied identity fields. |
| `POST /v1/agents/{agentId}/suspend` | none | Suspends an active agent. |
| `POST /v1/agents/{agentId}/revoke` | none | Revokes an agent. |

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `400 invalidAgentRequest` / `invalidEnrollmentRequest` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorizationDenied` / `enrollmentDenied` | `PERMISSION_DENIED` | The authority or enrollment evidence did not match. |
| `404 agentNotFound` | `NOT_FOUND` | The agent does not exist for the caller's tenant and authority. |
| `409 agentExists` / `invalidLifecycleTransition` / `enrollmentRequired` | `FAILED_PRECONDITION` | The lifecycle transition is not permitted, or enrollment is required first. |
| `503 identityStoreUnavailable` / `identityUnavailable` | `UNAVAILABLE` | A dependency is unavailable; consumers must fail closed. |
