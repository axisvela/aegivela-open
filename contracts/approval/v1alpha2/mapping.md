# Approval Issue Contract Mapping v1alpha2

`api_version` is the literal `aegivela.io/v1alpha2`; the protobuf service is `ApprovalService.Issue`. v1alpha2 preserves every v1alpha1 field number and type; request fields 13/14 and response field 3 are the only additions.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `approval_issue_request.api_version` | `ApprovalIssueRequest.api_version` | `ApprovalIssueRequest.api_version` = 1 |
| bearer token | `approval_issue_request.bearer_token` | `ApprovalIssueRequest.bearer_token` | `ApprovalIssueRequest.bearer_token` = 2 |
| action | `approval_issue_request.action` | `ApprovalIssueRequest.action` | `ApprovalIssueRequest.action` = 3 |
| resource | `approval_issue_request.resource` | `ApprovalIssueRequest.resource` | `ApprovalIssueRequest.resource` = 4 |
| resource kind | `resource.kind` | `ResourceRef.kind` | `ResourceRef.kind` = 1 |
| resource reference | `resource.reference` | `ResourceRef.reference` | `ResourceRef.reference` = 2 |
| scope | `approval_issue_request.scope` | `ApprovalIssueRequest.scope` | `ApprovalIssueRequest.scope` = 5 |
| expiry | `approval_issue_request.expiry` | `ApprovalIssueRequest.expiry` | `ApprovalIssueRequest.expiry` = 6 |
| reason | `approval_issue_request.reason` | `ApprovalIssueRequest.reason` | `ApprovalIssueRequest.reason` = 7 |
| policy version | `approval_issue_request.policy_version` | `ApprovalIssueRequest.policy_version` | `ApprovalIssueRequest.policy_version` = 8 |
| session | `approval_issue_request.session` | `ApprovalIssueRequest.session` | `ApprovalIssueRequest.session` = 9 |
| lifecycle | `approval_issue_request.lifecycle` | `ApprovalIssueRequest.lifecycle` | `ApprovalIssueRequest.lifecycle` = 10 |
| evidence refs | `approval_issue_request.evidence_refs` | `ApprovalIssueRequest.evidence_refs` | `ApprovalIssueRequest.evidence_refs` = 11 |
| trace ID | `approval_issue_request.trace_id` | `ApprovalIssueRequest.trace_id` | `ApprovalIssueRequest.trace_id` = 12 |
| structured resource digest | `approval_issue_request.structured_resource_digest` | `ApprovalIssueRequest.structured_resource_digest` | `ApprovalIssueRequest.structured_resource_digest` = 13 |
| descriptor version | `approval_issue_request.descriptor_version` | `ApprovalIssueRequest.descriptor_version` | `ApprovalIssueRequest.descriptor_version` = 14 |
| response API version | `approval_issue_response.api_version` | `ApprovalIssueResponse.api_version` | `ApprovalIssueResponse.api_version` = 1 |
| response approval ID | `approval_issue_response.approval_id` | `ApprovalIssueResponse.approval_id` | `ApprovalIssueResponse.approval_id` = 2 |
| signed approval artifact | `approval_issue_response.approval_artifact` | `ApprovalIssueResponse.approval_artifact` | `ApprovalIssueResponse.approval_artifact` = 3 |

`structured_resource_digest` and `descriptor_version` are required together: a descriptor-bound structured resource approval binds both, and a legacy kind/reference approval binds neither. All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha2 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorization_denied` | `PERMISSION_DENIED` | The caller is not authorized to issue this approval. |
| `400 invalid_approval_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocation_unavailable` | `UNAVAILABLE` | Revocation status cannot be safely checked; consumers must fail closed. |
