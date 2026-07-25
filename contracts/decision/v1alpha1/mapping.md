# Policy Decision Verification Contract Mapping

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `PolicyDecisionVerificationService.Verify`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| verify request API version | `policy_decision_verify_request.api_version` | `PolicyDecisionVerifyRequest.api_version` | `PolicyDecisionVerifyRequest.api_version` = 1 |
| signed decision | `policy_decision_verify_request.signed_decision` | `PolicyDecisionVerifyRequest.signed_decision` | `PolicyDecisionVerifyRequest.signed_decision` = 2 |
| expected bindings | `policy_decision_verify_request.expected_bindings` | `PolicyDecisionVerifyRequest.expected_bindings` | `PolicyDecisionVerifyRequest.expected_bindings` = 3 |
| expected principal | `expected_bindings.principal` | `ExpectedBindings.principal` | `ExpectedDecisionBindings.principal` = 1 |
| expected action | `expected_bindings.action` | `ExpectedBindings.action` | `ExpectedDecisionBindings.action` = 2 |
| expected resource | `expected_bindings.resource` | `ExpectedBindings.resource` | `ExpectedDecisionBindings.resource` = 3 |
| expected scope | `expected_bindings.scope` | `ExpectedBindings.scope` | `ExpectedDecisionBindings.scope` = 4 |
| expected audience | `expected_bindings.audience` | `ExpectedBindings.audience` | `ExpectedDecisionBindings.audience` = 5 |
| response decision ID | `policy_decision_verify_response.decision_id` | `PolicyDecisionVerifyResponse.decision_id` | `PolicyDecisionVerifyResponse.decision_id` = 2 |
| valid | `policy_decision_verify_response.valid` | `PolicyDecisionVerifyResponse.valid` | `PolicyDecisionVerifyResponse.valid` = 3 |
| outcome | `policy_decision_verify_response.outcome` | `PolicyDecisionVerifyResponse.outcome` | `PolicyDecisionVerifyResponse.outcome` = 4 |
| policy version | `policy_decision_verify_response.policy_version` | `PolicyDecisionVerifyResponse.policy_version` | `PolicyDecisionVerifyResponse.policy_version` = 5 |
| expires at | `policy_decision_verify_response.expires_at` | `PolicyDecisionVerifyResponse.expires_at` | `PolicyDecisionVerifyResponse.expires_at` = 6 |
| obligations | `policy_decision_verify_response.obligations` | `PolicyDecisionVerifyResponse.obligations` | `PolicyDecisionVerifyResponse.obligations` = 7 |
| evidence refs | `policy_decision_verify_response.evidence_refs` | `PolicyDecisionVerifyResponse.evidence_refs` | `PolicyDecisionVerifyResponse.evidence_refs` = 8 |
| revoked | `policy_decision_verify_response.revoked` | `PolicyDecisionVerifyResponse.revoked` | `PolicyDecisionVerifyResponse.revoked` = 9 |

Decision outcomes are `allow`, `deny`, `approval_required`, and `revoked`. This is the first public v1alpha1 protobuf baseline: `DECISION_OUTCOME_UNSPECIFIED=0` is an invalid sentinel; valid outcomes use `ALLOW=1`, `DENY=2`, `APPROVAL_REQUIRED=3`, and `REVOKED=4`. The GET /v1/policy/decisions/jwks endpoint returns a public JWKS document. The verify endpoint requires the `X-AEGIVELA-PEP` internal header. Future releases must preserve this baseline's enum values, field numbers, and field types.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `403 authorization_denied` | `PERMISSION_DENIED` | The caller is not authorized for decision verification. |
| `400 invalid_decision_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 decision_unavailable` | `UNAVAILABLE` | A signed decision cannot be safely verified; consumers must fail closed. |
