# Approval Issue Contract Mapping

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `ApprovalService.Issue`.

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
| response approval ID | `approval.approval_id` | `ApprovalIssueResponse.approval_id` | `ApprovalIssueResponse.approval_id` = 2 |
| response principal | `approval.principal` | `ApprovalIssueResponse.principal` | `ApprovalIssueResponse.principal` = 3 |
| response action | `approval.action` | `ApprovalIssueResponse.action` | `ApprovalIssueResponse.action` = 4 |
| response resource | `approval.resource` | `ApprovalIssueResponse.resource` | `ApprovalIssueResponse.resource` = 5 |
| response scope | `approval.scope` | `ApprovalIssueResponse.scope` | `ApprovalIssueResponse.scope` = 6 |
| expires at | `approval.expires_at` | `ApprovalIssueResponse.expires_at` | `ApprovalIssueResponse.expires_at` = 7 |
| issued at | `approval.issued_at` | `ApprovalIssueResponse.issued_at` | `ApprovalIssueResponse.issued_at` = 8 |
| reason | `approval.reason` | `ApprovalIssueResponse.reason` | `ApprovalIssueResponse.reason` = 9 |
| policy version | `approval.policy_version` | `ApprovalIssueResponse.policy_version` | `ApprovalIssueResponse.policy_version` = 10 |
| session | `approval.session` | `ApprovalIssueResponse.session` | `ApprovalIssueResponse.session` = 11 |
| lifecycle | `approval.lifecycle` | `ApprovalIssueResponse.lifecycle` | `ApprovalIssueResponse.lifecycle` = 12 |
| evidence refs | `approval.evidence_refs` | `ApprovalIssueResponse.evidence_refs` | `ApprovalIssueResponse.evidence_refs` = 13 |

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorization_denied` | `PERMISSION_DENIED` | The caller is not authorized to issue this approval. |
| `400 invalid_approval_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocation_unavailable` | `UNAVAILABLE` | Revocation status cannot be safely checked; consumers must fail closed. |
