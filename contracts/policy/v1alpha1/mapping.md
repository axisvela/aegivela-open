# Policy Decision Contract Mapping

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `PolicyDecisionService.Evaluate`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `policy_decision_request.api_version` | `PolicyDecisionRequest.api_version` | `PolicyDecisionRequest.api_version` = 1 |
| authorization mode | `policy_decision_request.authorization_mode` | `PolicyDecisionRequest.authorization_mode` | `PolicyDecisionRequest.authorization_mode` = 2 |
| principal | `policy_decision_request.principal` | `PolicyDecisionRequest.principal` | `PolicyDecisionRequest.principal` = 3 |
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
| action | `policy_decision_request.action` | `PolicyDecisionRequest.action` | `PolicyDecisionRequest.action` = 4 |
| resource | `policy_decision_request.resource` | `PolicyDecisionRequest.resource` | `PolicyDecisionRequest.resource` = 5 |
| resource kind | `resource.kind` | `Resource.kind` | `Resource.kind` = 1 |
| resource reference | `resource.reference` | `Resource.reference` | `Resource.reference` = 2 |
| requested scope | `policy_decision_request.requested_scope` | `PolicyDecisionRequest.requested_scope` | `PolicyDecisionRequest.requested_scope` = 6 |
| audience | `policy_decision_request.audience` | `PolicyDecisionRequest.audience` | `PolicyDecisionRequest.audience` = 7 |
| requested expiry | `policy_decision_request.requested_expires_at` | `PolicyDecisionRequest.requested_expires_at` | `PolicyDecisionRequest.requested_expires_at` = 8 |
| task binding | `policy_decision_request.task_binding` | `PolicyDecisionRequest.task_binding` | `PolicyDecisionRequest.task_binding` = 9 |
| parent authority | `policy_decision_request.parent_authority` | `PolicyDecisionRequest.parent_authority` | `PolicyDecisionRequest.parent_authority` = 10 |
| parent authority reference | `parent_authority.authority_ref` | `ParentAuthority.authority_ref` | `ParentAuthority.authority_ref` = 1 |
| parent scope | `parent_authority.parent_scope` | `ParentAuthority.parent_scope` | `ParentAuthority.parent_scope` = 2 |
| parent audience | `parent_authority.parent_audience` | `ParentAuthority.parent_audience` | `ParentAuthority.parent_audience` = 3 |
| parent expiry | `parent_authority.parent_expires_at` | `ParentAuthority.parent_expires_at` | `ParentAuthority.parent_expires_at` = 4 |
| parent task binding | `parent_authority.parent_task_binding` | `ParentAuthority.parent_task_binding` | `ParentAuthority.parent_task_binding` = 5 |
| risk context | `policy_decision_request.risk_context` | `PolicyDecisionRequest.risk_context` | `PolicyDecisionRequest.risk_context` = 11 |
| risk level | `risk_context.risk_level` | `RiskContext.risk_level` | `RiskContext.risk_level` = 1 |
| risk reference | `risk_context.risk_ref` | `RiskContext.risk_ref` | `RiskContext.risk_ref` = 2 |
| trace ID | `policy_decision_request.trace_id` | `PolicyDecisionRequest.trace_id` | `PolicyDecisionRequest.trace_id` = 12 |
| response API version | `policy_decision_response.api_version` | `PolicyDecisionResponse.api_version` | `PolicyDecisionResponse.api_version` = 1 |
| decision ID | `policy_decision_response.decision_id` | `PolicyDecisionResponse.decision_id` | `PolicyDecisionResponse.decision_id` = 2 |
| outcome | `policy_decision_response.outcome` | `PolicyDecisionResponse.outcome` | `PolicyDecisionResponse.outcome` = 3 |
| policy version | `policy_decision_response.policy_version` | `PolicyDecisionResponse.policy_version` | `PolicyDecisionResponse.policy_version` = 4 |
| expiry | `policy_decision_response.expires_at` | `PolicyDecisionResponse.expires_at` | `PolicyDecisionResponse.expires_at` = 5 |
| obligations | `policy_decision_response.obligations` | `PolicyDecisionResponse.obligations` | `PolicyDecisionResponse.obligations` = 6 |
| evidence references | `policy_decision_response.evidence_refs` | `PolicyDecisionResponse.evidence_refs` | `PolicyDecisionResponse.evidence_refs` = 7 |

Authorization modes are `human_web`, `system_api`, `delegated_api`, `service_agent_api`, and `twin_agent_api`. Decision outcomes are `allow`, `deny`, `approval_required`, and `revoked`. This is the first public v1alpha1 protobuf baseline: `AUTHORIZATION_MODE_UNSPECIFIED=0` and `DECISION_OUTCOME_UNSPECIFIED=0` are invalid sentinels; valid modes use values `1` through `5`, and valid outcomes use `ALLOW=1`, `DENY=2`, `APPROVAL_REQUIRED=3`, and `REVOKED=4`. Omitted proto3 scalar enum fields decode to their unspecified zero values and adapters must reject them. Future releases must preserve this baseline's enum values, field numbers, and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalid_policy_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract or policy input requirements. |
| `503 policy_unavailable` | `UNAVAILABLE` | A policy decision cannot be safely obtained; consumers must fail closed. |

Structural attenuation widening is represented as a conforming `deny` decision when an evaluator has a matching available policy. Malformed requests and unavailable policy remain transport errors.
