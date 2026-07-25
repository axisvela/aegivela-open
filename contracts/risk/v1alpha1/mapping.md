# Risk Signal Contract Mapping v1alpha1

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `RiskSignalService.Process`.

Risk signals are CAEP-style continuous evaluation events. The risk level determines the action:
- `low`: logged for audit only; no active revocation
- `medium`: logged + PDP will deny next evaluation for rules with `RiskThreshold` below `medium`
- `high`: subject and session revocation records are written; in-flight grants invalidated on next verification
- `critical`: subject, session, and all subject grants are immediately revoked

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `risk_signal_request.api_version` | `RiskSignalRequest.api_version` | `RiskSignalRequest.api_version` = 1 |
| subject ref | `risk_signal_request.subject_ref` | `RiskSignalRequest.subject_ref` | `RiskSignalRequest.subject_ref` = 2 |
| session ref | `risk_signal_request.session_ref` | `RiskSignalRequest.session_ref` | `RiskSignalRequest.session_ref` = 3 |
| risk level | `risk_signal_request.risk_level` | `RiskSignalRequest.risk_level` | `RiskSignalRequest.risk_level` = 4 |
| risk ref | `risk_signal_request.risk_ref` | `RiskSignalRequest.risk_ref` | `RiskSignalRequest.risk_ref` = 5 |
| reason | `risk_signal_request.reason` | `RiskSignalRequest.reason` | `RiskSignalRequest.reason` = 6 |
| trace ID | `risk_signal_request.trace_id` | `RiskSignalRequest.trace_id` | `RiskSignalRequest.trace_id` = 7 |
| response API version | `risk_signal_response.api_version` | `RiskSignalResponse.api_version` | `RiskSignalResponse.api_version` = 1 |
| signal ID | `risk_signal_response.signal_id` | `RiskSignalResponse.signal_id` | `RiskSignalResponse.signal_id` = 2 |
| risk level | `risk_signal_response.risk_level` | `RiskSignalResponse.risk_level` | `RiskSignalResponse.risk_level` = 3 |
| action | `risk_signal_response.action` | `RiskSignalResponse.action` | `RiskSignalResponse.action` = 4 |
| trace ID | `risk_signal_response.trace_id` | `RiskSignalResponse.trace_id` | `RiskSignalResponse.trace_id` = 5 |

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalid_risk_signal` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocation_unavailable` | `UNAVAILABLE` | Revocation writes cannot be performed safely; consumers must fail closed. |
