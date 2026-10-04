# Policy Decision v1alpha3 Contract Mapping

`api_version` is `aegivela.io/v1alpha3`; the protobuf service is `PolicyDecisionService.Evaluate`. v1alpha3 preserves every v1alpha2 field number and type. `PolicyDecisionRequest.parent_execution_grant` is an optional string at protobuf field `13`.

For `delegated_api`, `parent_execution_grant` is required and `parent_authority` is prohibited on HTTP requests. The compact grant is verified for signature, configured issuer, expiry, audience, structured-resource digest, and tenant-scoped revocation before principal resolution or policy evaluation. A verified grant exclusively supplies the parent authority: JTI, exact action/resource, scope, audience, expiry, and task binding. Action/resource equality and monotonic scope/audience/expiry/task binding are enforced before evaluation.

v1alpha3 uses the v1alpha2 structured-resource digest rule. v1alpha1 and v1alpha2 contracts and artifacts are immutable.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `policy_decision_request.api_version` | `PolicyDecisionRequest.api_version` | `PolicyDecisionRequest.api_version` = 1 |
| authorization mode | `policy_decision_request.authorization_mode` | `PolicyDecisionRequest.authorization_mode` | `PolicyDecisionRequest.authorization_mode` = 2 |
| principal | `policy_decision_request.principal` | `PolicyDecisionRequest.principal` | `PolicyDecisionRequest.principal` = 3 |
| action | `policy_decision_request.action` | `PolicyDecisionRequest.action` | `PolicyDecisionRequest.action` = 4 |
| resource | `policy_decision_request.resource` | `PolicyDecisionRequest.resource` | `PolicyDecisionRequest.resource` = 5 |
| requested scope | `policy_decision_request.requested_scope` | `PolicyDecisionRequest.requested_scope` | `PolicyDecisionRequest.requested_scope` = 6 |
| audience | `policy_decision_request.audience` | `PolicyDecisionRequest.audience` | `PolicyDecisionRequest.audience` = 7 |
| requested expiry | `policy_decision_request.requested_expires_at` | `PolicyDecisionRequest.requested_expires_at` | `PolicyDecisionRequest.requested_expires_at` = 8 |
| task binding | `policy_decision_request.task_binding` | `PolicyDecisionRequest.task_binding` | `PolicyDecisionRequest.task_binding` = 9 |
| parent authority | `policy_decision_request.parent_authority` | `PolicyDecisionRequest.parent_authority` | `PolicyDecisionRequest.parent_authority` = 10 |
| risk context | `policy_decision_request.risk_context` | `PolicyDecisionRequest.risk_context` | `PolicyDecisionRequest.risk_context` = 11 |
| trace ID | `policy_decision_request.trace_id` | `PolicyDecisionRequest.trace_id` | `PolicyDecisionRequest.trace_id` = 12 |
| parent execution grant | `policy_decision_request.parent_execution_grant` | `PolicyDecisionRequest.parent_execution_grant` | `PolicyDecisionRequest.parent_execution_grant` = 13 |
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
| namespace | `principal.namespace` | `Principal.namespace` | `Principal.namespace` = 12 |
| authority binding | `principal.authority_binding` | `Principal.authority_binding` | `Principal.authority_binding` = 13 |
| agent epoch | `principal.agent_epoch` | `Principal.agent_epoch` | `Principal.agent_epoch` = 14 |
| identity epoch | `principal.identity_epoch` | `Principal.identity_epoch` | `Principal.identity_epoch` = 15 |
| instance reference | `principal.instance_id` | `Principal.instance_id` | `Principal.instance_id` = 16 |
| resource kind | `resource.kind` | `Resource.kind` | `Resource.kind` = 1 |
| resource reference | `resource.reference` | `Resource.reference` | `Resource.reference` = 2 |
| product ID | `resource.product_id` | `Resource.product_id` | `Resource.product_id` = 3 |
| resource type | `resource.resource_type` | `Resource.resource_type` | `Resource.resource_type` = 4 |
| descriptor ID | `resource.descriptor_id` | `Resource.descriptor_id` | `Resource.descriptor_id` = 5 |
| descriptor version | `resource.descriptor_version` | `Resource.descriptor_version` | `Resource.descriptor_version` = 6 |
| resource attributes | `resource.attributes` | `Resource.attributes` | `Resource.attributes` = 7 |
| parent authority reference | `parent_authority.authority_ref` | `ParentAuthority.authority_ref` | `ParentAuthority.authority_ref` = 1 |
| parent scope | `parent_authority.parent_scope` | `ParentAuthority.parent_scope` | `ParentAuthority.parent_scope` = 2 |
| parent audience | `parent_authority.parent_audience` | `ParentAuthority.parent_audience` | `ParentAuthority.parent_audience` = 3 |
| parent expiry | `parent_authority.parent_expires_at` | `ParentAuthority.parent_expires_at` | `ParentAuthority.parent_expires_at` = 4 |
| parent task binding | `parent_authority.parent_task_binding` | `ParentAuthority.parent_task_binding` | `ParentAuthority.parent_task_binding` = 5 |
| risk level | `risk_context.risk_level` | `RiskContext.risk_level` | `RiskContext.risk_level` = 1 |
| risk reference | `risk_context.risk_ref` | `RiskContext.risk_ref` | `RiskContext.risk_ref` = 2 |
| response API version | `policy_decision_response.api_version` | `PolicyDecisionResponse.api_version` | `PolicyDecisionResponse.api_version` = 1 |
| decision ID | `policy_decision_response.decision_id` | `PolicyDecisionResponse.decision_id` | `PolicyDecisionResponse.decision_id` = 2 |
| outcome | `policy_decision_response.outcome` | `PolicyDecisionResponse.outcome` | `PolicyDecisionResponse.outcome` = 3 |
| policy version | `policy_decision_response.policy_version` | `PolicyDecisionResponse.policy_version` | `PolicyDecisionResponse.policy_version` = 4 |
| expiry | `policy_decision_response.expires_at` | `PolicyDecisionResponse.expires_at` | `PolicyDecisionResponse.expires_at` = 5 |
| obligations | `policy_decision_response.obligations` | `PolicyDecisionResponse.obligations` | `PolicyDecisionResponse.obligations` = 6 |
| evidence references | `policy_decision_response.evidence_refs` | `PolicyDecisionResponse.evidence_refs` | `PolicyDecisionResponse.evidence_refs` = 7 |
| structured resource digest | `policy_decision_response.structured_resource_digest` | `PolicyDecisionResponse.structured_resource_digest` | `PolicyDecisionResponse.structured_resource_digest` = 8 |

## Transport Errors

| HTTP status and error | Meaning |
| --- | --- |
| `401 unauthenticated` | The parent grant is missing, malformed, expired, revoked, incorrectly issued, or bound to another tenant or subject. |
| `403 policy_request_denied` | A verified parent grant would be expanded. |
| `503 parent_execution_grant_verification_unavailable` | Verification or revocation could not be safely completed. |
