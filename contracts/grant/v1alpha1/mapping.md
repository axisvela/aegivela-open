# Execution Grant Contract Mapping

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `ExecutionGrantService`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| issue request API version | `grant_issue_request.api_version` | `GrantIssueRequest.api_version` | `GrantIssueRequest.api_version` = 1 |
| bearer token | `grant_issue_request.bearer_token` | `GrantIssueRequest.bearer_token` | `GrantIssueRequest.bearer_token` = 2 |
| signed allow decision | `grant_issue_request.signed_allow_decision` | `GrantIssueRequest.signed_allow_decision` | `GrantIssueRequest.signed_allow_decision` = 3 |
| requested scope | `grant_issue_request.requested_scope` | `GrantIssueRequest.requested_scope` | `GrantIssueRequest.requested_scope` = 4 |
| pre-authorization JTI | `grant_issue_request.pre_authorization_jti` | `GrantIssueRequest.pre_authorization_jti` | `GrantIssueRequest.pre_authorization_jti` = 5 |
| approval JTI | `grant_issue_request.approval_jti` | `GrantIssueRequest.approval_jti` | `GrantIssueRequest.approval_jti` = 6 |
| trace ID | `grant_issue_request.trace_id` | `GrantIssueRequest.trace_id` | `GrantIssueRequest.trace_id` = 7 |
| grant ID | `execution_grant.grant_id` | `GrantIssueResponse.grant_id` | `GrantIssueResponse.grant_id` = 2 |
| principal | `execution_grant.principal` | `GrantIssueResponse.principal` | `GrantIssueResponse.principal` = 3 |
| scope | `execution_grant.scope` | `GrantIssueResponse.scope` | `GrantIssueResponse.scope` = 4 |
| audience | `execution_grant.audience` | `GrantIssueResponse.audience` | `GrantIssueResponse.audience` = 5 |
| expires at | `execution_grant.expires_at` | `GrantIssueResponse.expires_at` | `GrantIssueResponse.expires_at` = 6 |
| issued at | `execution_grant.issued_at` | `GrantIssueResponse.issued_at` | `GrantIssueResponse.issued_at` = 7 |
| task binding | `execution_grant.task_binding` | `GrantIssueResponse.task_binding` | `GrantIssueResponse.task_binding` = 8 |
| pre-authorization JTI | `execution_grant.pre_authorization_jti` | `GrantIssueResponse.pre_authorization_jti` | `GrantIssueResponse.pre_authorization_jti` = 9 |
| approval JTI | `execution_grant.approval_jti` | `GrantIssueResponse.approval_jti` | `GrantIssueResponse.approval_jti` = 10 |
| evidence refs | `execution_grant.evidence_refs` | `GrantIssueResponse.evidence_refs` | `GrantIssueResponse.evidence_refs` = 11 |
| verify signed grant | `grant_verify_request.signed_grant` | `GrantVerifyRequest.signed_grant` | `GrantVerifyRequest.signed_grant` = 2 |
| expected bindings | `grant_verify_request.expected_bindings` | `GrantVerifyRequest.expected_bindings` | `GrantVerifyRequest.expected_bindings` = 3 |
| expected audience | `expected_bindings.audience` | `ExpectedBindings.audience` | `ExpectedBindings.audience` = 1 |
| expected scope | `expected_bindings.scope` | `ExpectedBindings.scope` | `ExpectedBindings.scope` = 2 |
| expected task binding | `expected_bindings.task_binding` | `ExpectedBindings.task_binding` | `ExpectedBindings.task_binding` = 3 |
| valid | `grant_verify_response.valid` | `GrantVerifyResponse.valid` | `GrantVerifyResponse.valid` = 3 |
| expired | `grant_verify_response.expired` | `GrantVerifyResponse.expired` | `GrantVerifyResponse.expired` = 8 |
| revoked | `grant_verify_response.revoked` | `GrantVerifyResponse.revoked` | `GrantVerifyResponse.revoked` = 9 |
| revocation check epoch | `grant_verify_response.revocation_check_epoch` | `GrantVerifyResponse.revocation_check_epoch` | `GrantVerifyResponse.revocation_check_epoch` = 10 |

The GET /v1/grants/jwks endpoint returns a public JWKS document. All mutation endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorization_denied` | `PERMISSION_DENIED` | The caller is not authorized for the requested grant operation. |
| `400 invalid_grant_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocation_unavailable` | `UNAVAILABLE` | Revocation status cannot be safely checked; consumers must fail closed. |
