# Approval Issue Contract Mapping v1alpha2

`apiVersion` is the literal `aegivela.io/v2.0`; the protobuf service is `ApprovalService.Issue`. v1alpha2 preserves every v1alpha1 field number and type; request fields 13/14 and response field 3 are the only additions.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `approvalIssueRequest.apiVersion` | `ApprovalIssueRequest.apiVersion` | `ApprovalIssueRequest.apiVersion` = 1 |
| bearer token | `approvalIssueRequest.bearerToken` | `ApprovalIssueRequest.bearerToken` | `ApprovalIssueRequest.bearerToken` = 2 |
| action | `approvalIssueRequest.action` | `ApprovalIssueRequest.action` | `ApprovalIssueRequest.action` = 3 |
| resource | `approvalIssueRequest.resource` | `ApprovalIssueRequest.resource` | `ApprovalIssueRequest.resource` = 4 |
| resource kind | `resource.kind` | `ResourceRef.kind` | `ResourceRef.kind` = 1 |
| resource reference | `resource.reference` | `ResourceRef.reference` | `ResourceRef.reference` = 2 |
| scope | `approvalIssueRequest.scope` | `ApprovalIssueRequest.scope` | `ApprovalIssueRequest.scope` = 5 |
| expiry | `approvalIssueRequest.expiry` | `ApprovalIssueRequest.expiry` | `ApprovalIssueRequest.expiry` = 6 |
| reason | `approvalIssueRequest.reason` | `ApprovalIssueRequest.reason` | `ApprovalIssueRequest.reason` = 7 |
| policy version | `approvalIssueRequest.policyVersion` | `ApprovalIssueRequest.policyVersion` | `ApprovalIssueRequest.policyVersion` = 8 |
| session | `approvalIssueRequest.session` | `ApprovalIssueRequest.session` | `ApprovalIssueRequest.session` = 9 |
| lifecycle | `approvalIssueRequest.lifecycle` | `ApprovalIssueRequest.lifecycle` | `ApprovalIssueRequest.lifecycle` = 10 |
| evidence refs | `approvalIssueRequest.evidenceRefs` | `ApprovalIssueRequest.evidenceRefs` | `ApprovalIssueRequest.evidenceRefs` = 11 |
| trace ID | `approvalIssueRequest.traceId` | `ApprovalIssueRequest.traceId` | `ApprovalIssueRequest.traceId` = 12 |
| structured resource digest | `approvalIssueRequest.structuredResourceDigest` | `ApprovalIssueRequest.structuredResourceDigest` | `ApprovalIssueRequest.structuredResourceDigest` = 13 |
| descriptor version | `approvalIssueRequest.descriptorVersion` | `ApprovalIssueRequest.descriptorVersion` | `ApprovalIssueRequest.descriptorVersion` = 14 |
| response API version | `approvalIssueResponse.apiVersion` | `ApprovalIssueResponse.apiVersion` | `ApprovalIssueResponse.apiVersion` = 1 |
| response approval ID | `approvalIssueResponse.approvalId` | `ApprovalIssueResponse.approvalId` | `ApprovalIssueResponse.approvalId` = 2 |
| signed approval artifact | `approvalIssueResponse.approvalArtifact` | `ApprovalIssueResponse.approvalArtifact` | `ApprovalIssueResponse.approvalArtifact` = 3 |

`structuredResourceDigest` and `descriptorVersion` are required together: a descriptor-bound structured resource approval binds both, and a legacy kind/reference approval binds neither. All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha2 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorizationDenied` | `PERMISSION_DENIED` | The caller is not authorized to issue this approval. |
| `400 invalidApprovalRequest` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocationUnavailable` | `UNAVAILABLE` | Revocation status cannot be safely checked; consumers must fail closed. |
