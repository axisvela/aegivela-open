# Policy Decision v2.0 Contract Mapping

`apiVersion` is `aegivela.io/v2.0`; the protobuf service is `PolicyDecisionService.Evaluate`. v2.0 uses camelCase JSON property names. `PolicyDecisionRequest.parentExecutionGrant` is an optional string at protobuf field `13`.

For `delegatedApi`, `parentExecutionGrant` is required and `parentAuthority` is prohibited on HTTP requests. The compact grant is verified for signature, configured issuer, expiry, audience, structured-resource digest, and tenant-scoped revocation before principal resolution or policy evaluation. A verified grant exclusively supplies the parent authority: JTI, exact action/resource, scope, audience, expiry, and task binding. Action/resource equality and monotonic scope/audience/expiry/task binding are enforced before evaluation.

v1alpha3 uses the v1alpha2 structured-resource digest rule. v1alpha1 and v1alpha2 contracts and artifacts are immutable.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `policyDecisionRequest.apiVersion` | `PolicyDecisionRequest.apiVersion` | `PolicyDecisionRequest.apiVersion` = 1 |
| authorization mode | `policyDecisionRequest.authorizationMode` | `PolicyDecisionRequest.authorizationMode` | `PolicyDecisionRequest.authorizationMode` = 2 |
| principal | `policyDecisionRequest.principal` | `PolicyDecisionRequest.principal` | `PolicyDecisionRequest.principal` = 3 |
| action | `policyDecisionRequest.action` | `PolicyDecisionRequest.action` | `PolicyDecisionRequest.action` = 4 |
| resource | `policyDecisionRequest.resource` | `PolicyDecisionRequest.resource` | `PolicyDecisionRequest.resource` = 5 |
| requested scope | `policyDecisionRequest.requestedScope` | `PolicyDecisionRequest.requestedScope` | `PolicyDecisionRequest.requestedScope` = 6 |
| audience | `policyDecisionRequest.audience` | `PolicyDecisionRequest.audience` | `PolicyDecisionRequest.audience` = 7 |
| requested expiry | `policyDecisionRequest.requestedExpiresAt` | `PolicyDecisionRequest.requestedExpiresAt` | `PolicyDecisionRequest.requestedExpiresAt` = 8 |
| task binding | `policyDecisionRequest.taskBinding` | `PolicyDecisionRequest.taskBinding` | `PolicyDecisionRequest.taskBinding` = 9 |
| parent authority | `policyDecisionRequest.parentAuthority` | `PolicyDecisionRequest.parentAuthority` | `PolicyDecisionRequest.parentAuthority` = 10 |
| risk context | `policyDecisionRequest.riskContext` | `PolicyDecisionRequest.riskContext` | `PolicyDecisionRequest.riskContext` = 11 |
| trace ID | `policyDecisionRequest.traceId` | `PolicyDecisionRequest.traceId` | `PolicyDecisionRequest.traceId` = 12 |
| parent execution grant | `policyDecisionRequest.parentExecutionGrant` | `PolicyDecisionRequest.parentExecutionGrant` | `PolicyDecisionRequest.parentExecutionGrant` = 13 |
| subject reference | `principal.subjectRef` | `Principal.subjectRef` | `Principal.subjectRef` = 2 |
| actor reference | `principal.actorRef` | `Principal.actorRef` | `Principal.actorRef` = 3 |
| client ID | `principal.clientId` | `Principal.clientId` | `Principal.clientId` = 4 |
| workload reference | `principal.workloadRef` | `Principal.workloadRef` | `Principal.workloadRef` = 5 |
| agent ID | `principal.agentId` | `Principal.agentId` | `Principal.agentId` = 6 |
| agent class | `principal.agentClass` | `Principal.agentClass` | `Principal.agentClass` = 7 |
| master ID | `principal.masterId` | `Principal.masterId` | `Principal.masterId` = 8 |
| organization authority root | `principal.organizationAuthorityRoot` | `Principal.organizationAuthorityRoot` | `Principal.organizationAuthorityRoot` = 9 |
| attestation reference | `principal.attestationRef` | `Principal.attestationRef` | `Principal.attestationRef` = 11 |
| namespace | `principal.namespace` | `Principal.namespace` | `Principal.namespace` = 12 |
| authority binding | `principal.authorityBinding` | `Principal.authorityBinding` | `Principal.authorityBinding` = 13 |
| agent epoch | `principal.agentEpoch` | `Principal.agentEpoch` | `Principal.agentEpoch` = 14 |
| identity epoch | `principal.identityEpoch` | `Principal.identityEpoch` | `Principal.identityEpoch` = 15 |
| instance reference | `principal.instanceId` | `Principal.instanceId` | `Principal.instanceId` = 16 |
| resource kind | `resource.kind` | `Resource.kind` | `Resource.kind` = 1 |
| resource reference | `resource.reference` | `Resource.reference` | `Resource.reference` = 2 |
| product ID | `resource.productId` | `Resource.productId` | `Resource.productId` = 3 |
| resource type | `resource.resourceType` | `Resource.resourceType` | `Resource.resourceType` = 4 |
| descriptor ID | `resource.descriptorId` | `Resource.descriptorId` | `Resource.descriptorId` = 5 |
| descriptor version | `resource.descriptorVersion` | `Resource.descriptorVersion` | `Resource.descriptorVersion` = 6 |
| resource attributes | `resource.attributes` | `Resource.attributes` | `Resource.attributes` = 7 |
| parent authority reference | `parentAuthority.authorityRef` | `ParentAuthority.authorityRef` | `ParentAuthority.authorityRef` = 1 |
| parent scope | `parentAuthority.parentScope` | `ParentAuthority.parentScope` | `ParentAuthority.parentScope` = 2 |
| parent audience | `parentAuthority.parentAudience` | `ParentAuthority.parentAudience` | `ParentAuthority.parentAudience` = 3 |
| parent expiry | `parentAuthority.parentExpiresAt` | `ParentAuthority.parentExpiresAt` | `ParentAuthority.parentExpiresAt` = 4 |
| parent task binding | `parentAuthority.parentTaskBinding` | `ParentAuthority.parentTaskBinding` | `ParentAuthority.parentTaskBinding` = 5 |
| risk level | `riskContext.riskLevel` | `RiskContext.riskLevel` | `RiskContext.riskLevel` = 1 |
| risk reference | `riskContext.riskRef` | `RiskContext.riskRef` | `RiskContext.riskRef` = 2 |
| response API version | `policyDecisionResponse.apiVersion` | `PolicyDecisionResponse.apiVersion` | `PolicyDecisionResponse.apiVersion` = 1 |
| decision ID | `policyDecisionResponse.decisionId` | `PolicyDecisionResponse.decisionId` | `PolicyDecisionResponse.decisionId` = 2 |
| outcome | `policyDecisionResponse.outcome` | `PolicyDecisionResponse.outcome` | `PolicyDecisionResponse.outcome` = 3 |
| policy version | `policyDecisionResponse.policyVersion` | `PolicyDecisionResponse.policyVersion` | `PolicyDecisionResponse.policyVersion` = 4 |
| expiry | `policyDecisionResponse.expiresAt` | `PolicyDecisionResponse.expiresAt` | `PolicyDecisionResponse.expiresAt` = 5 |
| obligations | `policyDecisionResponse.obligations` | `PolicyDecisionResponse.obligations` | `PolicyDecisionResponse.obligations` = 6 |
| evidence references | `policyDecisionResponse.evidenceRefs` | `PolicyDecisionResponse.evidenceRefs` | `PolicyDecisionResponse.evidenceRefs` = 7 |
| structured resource digest | `policyDecisionResponse.structuredResourceDigest` | `PolicyDecisionResponse.structuredResourceDigest` | `PolicyDecisionResponse.structuredResourceDigest` = 8 |

## Transport Errors

| HTTP status and error | Meaning |
| --- | --- |
| `401 unauthenticated` | The parent grant is missing, malformed, expired, revoked, incorrectly issued, or bound to another tenant or subject. |
| `403 policyRequestDenied` | A verified parent grant would be expanded. |
| `503 parentExecutionGrantVerificationUnavailable` | Verification or revocation could not be safely completed. |
