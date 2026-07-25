# Agent Lifecycle Mapping

| JSON Schema | OpenAPI | Protobuf |
|---|---|---|
| `agent_register_request` | `AgentRegisterRequest` | `AgentRegisterRequest` |
| `agent_register_response` | `AgentRecord` | `AgentRecord` |
| `agent_get_request` | path parameter `agent_id` | `AgentGetRequest` |
| `agent_get_response` | `AgentRecord` | `AgentRecord` |
| `agent_lifecycle_request` | `AgentLifecycleRequest` | `AgentLifecycleRequest` |
| `agent_lifecycle_response` | `AgentRecord` | `AgentRecord` |
| `agent_class` enum | `agent_class` inline enum | `AgentClass` enum |
| `lifecycle_state` enum | `lifecycle_state` inline enum | `LifecycleState` enum |
| `lifecycle_action` enum | `action` inline enum | `LifecycleAction` enum |
