# Execution Grant Contract Mapping

`apiVersion` is the literal `aegivela.io/v2.0`; the protobuf service is `ExecutionGrantService`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| issue request API version | `grantIssueRequest.apiVersion` | `GrantIssueRequest.apiVersion` | `GrantIssueRequest.apiVersion` = 1 |
| bearer token | `grantIssueRequest.bearerToken` | `GrantIssueRequest.bearerToken` | `GrantIssueRequest.bearerToken` = 2 |
| signed allow decision | `grantIssueRequest.signedAllowDecision` | `GrantIssueRequest.signedAllowDecision` | `GrantIssueRequest.signedAllowDecision` = 3 |
| requested scope | `grantIssueRequest.requestedScope` | `GrantIssueRequest.requestedScope` | `GrantIssueRequest.requestedScope` = 4 |
| pre-authorization JTI | `grantIssueRequest.preAuthorizationJti` | `GrantIssueRequest.preAuthorizationJti` | `GrantIssueRequest.preAuthorizationJti` = 5 |
| approval JTI | `grantIssueRequest.approvalJti` | `GrantIssueRequest.approvalJti` | `GrantIssueRequest.approvalJti` = 6 |
| trace ID | `grantIssueRequest.traceId` | `GrantIssueRequest.traceId` | `GrantIssueRequest.traceId` = 7 |
| grant ID | `executionGrant.grantId` | `GrantIssueResponse.grantId` | `GrantIssueResponse.grantId` = 2 |
| principal | `executionGrant.principal` | `GrantIssueResponse.principal` | `GrantIssueResponse.principal` = 3 |
| scope | `executionGrant.scope` | `GrantIssueResponse.scope` | `GrantIssueResponse.scope` = 4 |
| audience | `executionGrant.audience` | `GrantIssueResponse.audience` | `GrantIssueResponse.audience` = 5 |
| expires at | `executionGrant.expiresAt` | `GrantIssueResponse.expiresAt` | `GrantIssueResponse.expiresAt` = 6 |
| issued at | `executionGrant.issuedAt` | `GrantIssueResponse.issuedAt` | `GrantIssueResponse.issuedAt` = 7 |
| task binding | `executionGrant.taskBinding` | `GrantIssueResponse.taskBinding` | `GrantIssueResponse.taskBinding` = 8 |
| pre-authorization JTI | `executionGrant.preAuthorizationJti` | `GrantIssueResponse.preAuthorizationJti` | `GrantIssueResponse.preAuthorizationJti` = 9 |
| approval JTI | `executionGrant.approvalJti` | `GrantIssueResponse.approvalJti` | `GrantIssueResponse.approvalJti` = 10 |
| evidence refs | `executionGrant.evidenceRefs` | `GrantIssueResponse.evidenceRefs` | `GrantIssueResponse.evidenceRefs` = 11 |
| verify signed grant | `grantVerifyRequest.signedGrant` | `GrantVerifyRequest.signedGrant` | `GrantVerifyRequest.signedGrant` = 2 |
| expected bindings | `grantVerifyRequest.expectedBindings` | `GrantVerifyRequest.expectedBindings` | `GrantVerifyRequest.expectedBindings` = 3 |
| expected audience | `expectedBindings.audience` | `ExpectedBindings.audience` | `ExpectedBindings.audience` = 1 |
| expected scope | `expectedBindings.scope` | `ExpectedBindings.scope` | `ExpectedBindings.scope` = 2 |
| expected task binding | `expectedBindings.taskBinding` | `ExpectedBindings.taskBinding` | `ExpectedBindings.taskBinding` = 3 |
| valid | `grantVerifyResponse.valid` | `GrantVerifyResponse.valid` | `GrantVerifyResponse.valid` = 3 |
| expired | `grantVerifyResponse.expired` | `GrantVerifyResponse.expired` | `GrantVerifyResponse.expired` = 8 |
| revoked | `grantVerifyResponse.revoked` | `GrantVerifyResponse.revoked` | `GrantVerifyResponse.revoked` = 9 |
| revocation check epoch | `grantVerifyResponse.revocationCheckEpoch` | `GrantVerifyResponse.revocationCheckEpoch` | `GrantVerifyResponse.revocationCheckEpoch` = 10 |

The GET /v1/grants/jwks endpoint returns a public JWKS document. All mutation endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline. Future releases must preserve this baseline's field numbers and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorizationDenied` | `PERMISSION_DENIED` | The caller is not authorized for the requested grant operation. |
| `400 invalidGrantRequest` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocationUnavailable` | `UNAVAILABLE` | Revocation status cannot be safely checked; consumers must fail closed. |
