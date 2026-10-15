# Policy Decision Verification Contract Mapping

`apiVersion` is the literal `aegivela.io/v2.0`; the protobuf service is `PolicyDecisionVerificationService.Verify`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| verify request API version | `policyDecisionVerifyRequest.apiVersion` | `PolicyDecisionVerifyRequest.apiVersion` | `PolicyDecisionVerifyRequest.apiVersion` = 1 |
| signed decision | `policyDecisionVerifyRequest.signedDecision` | `PolicyDecisionVerifyRequest.signedDecision` | `PolicyDecisionVerifyRequest.signedDecision` = 2 |
| expected bindings | `policyDecisionVerifyRequest.expectedBindings` | `PolicyDecisionVerifyRequest.expectedBindings` | `PolicyDecisionVerifyRequest.expectedBindings` = 3 |
| expected principal | `expectedBindings.principal` | `ExpectedBindings.principal` | `ExpectedDecisionBindings.principal` = 1 |
| expected action | `expectedBindings.action` | `ExpectedBindings.action` | `ExpectedDecisionBindings.action` = 2 |
| expected resource | `expectedBindings.resource` | `ExpectedBindings.resource` | `ExpectedDecisionBindings.resource` = 3 |
| expected scope | `expectedBindings.scope` | `ExpectedBindings.scope` | `ExpectedDecisionBindings.scope` = 4 |
| expected audience | `expectedBindings.audience` | `ExpectedBindings.audience` | `ExpectedDecisionBindings.audience` = 5 |
| response decision ID | `policyDecisionVerifyResponse.decisionId` | `PolicyDecisionVerifyResponse.decisionId` | `PolicyDecisionVerifyResponse.decisionId` = 2 |
| valid | `policyDecisionVerifyResponse.valid` | `PolicyDecisionVerifyResponse.valid` | `PolicyDecisionVerifyResponse.valid` = 3 |
| outcome | `policyDecisionVerifyResponse.outcome` | `PolicyDecisionVerifyResponse.outcome` | `PolicyDecisionVerifyResponse.outcome` = 4 |
| policy version | `policyDecisionVerifyResponse.policyVersion` | `PolicyDecisionVerifyResponse.policyVersion` | `PolicyDecisionVerifyResponse.policyVersion` = 5 |
| expires at | `policyDecisionVerifyResponse.expiresAt` | `PolicyDecisionVerifyResponse.expiresAt` | `PolicyDecisionVerifyResponse.expiresAt` = 6 |
| obligations | `policyDecisionVerifyResponse.obligations` | `PolicyDecisionVerifyResponse.obligations` | `PolicyDecisionVerifyResponse.obligations` = 7 |
| evidence refs | `policyDecisionVerifyResponse.evidenceRefs` | `PolicyDecisionVerifyResponse.evidenceRefs` | `PolicyDecisionVerifyResponse.evidenceRefs` = 8 |
| revoked | `policyDecisionVerifyResponse.revoked` | `PolicyDecisionVerifyResponse.revoked` | `PolicyDecisionVerifyResponse.revoked` = 9 |

Decision outcomes are `allow`, `deny`, `approvalRequired`, and `revoked`. The zero enum value is an invalid sentinel; valid outcome numbers are 1 through 4. Existing protobuf field numbers and types remain unchanged. GET /v1/policy/decisions/jwks returns public keys; verify uses `X-Aegivela-Internal-Token`. See [signed artifact semantics](signed-artifact.md) for identity, validity, encoding limits and acceptance boundaries.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorizationDenied` | `PERMISSION_DENIED` | The caller is not authorized for decision verification. |
| `400 invalidDecisionRequest` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 decisionUnavailable` | `UNAVAILABLE` | A signed decision cannot be safely verified; consumers must fail closed. |
