# Risk Signal Contract Mapping v1alpha1

`apiVersion` is the literal `aegivela.io/v2.0`; the protobuf service is `RiskSignalService.Process`.

Risk signals are CAEP-style continuous evaluation events. The risk level determines the action:
- `low`: logged for audit only; no active revocation
- `medium`: logged + PDP will deny next evaluation for rules with `RiskThreshold` below `medium`
- `high`: subject and session revocation records are written; in-flight grants invalidated on next verification
- `critical`: subject, session, and all subject grants are immediately revoked

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `riskSignalRequest.apiVersion` | `RiskSignalRequest.apiVersion` | `RiskSignalRequest.apiVersion` = 1 |
| subject ref | `riskSignalRequest.subjectRef` | `RiskSignalRequest.subjectRef` | `RiskSignalRequest.subjectRef` = 2 |
| session ref | `riskSignalRequest.sessionRef` | `RiskSignalRequest.sessionRef` | `RiskSignalRequest.sessionRef` = 3 |
| risk level | `riskSignalRequest.riskLevel` | `RiskSignalRequest.riskLevel` | `RiskSignalRequest.riskLevel` = 4 |
| risk ref | `riskSignalRequest.riskRef` | `RiskSignalRequest.riskRef` | `RiskSignalRequest.riskRef` = 5 |
| reason | `riskSignalRequest.reason` | `RiskSignalRequest.reason` | `RiskSignalRequest.reason` = 6 |
| trace ID | `riskSignalRequest.traceId` | `RiskSignalRequest.traceId` | `RiskSignalRequest.traceId` = 7 |
| response API version | `riskSignalResponse.apiVersion` | `RiskSignalResponse.apiVersion` | `RiskSignalResponse.apiVersion` = 1 |
| signal ID | `riskSignalResponse.signalId` | `RiskSignalResponse.signalId` | `RiskSignalResponse.signalId` = 2 |
| risk level | `riskSignalResponse.riskLevel` | `RiskSignalResponse.riskLevel` | `RiskSignalResponse.riskLevel` = 3 |
| action | `riskSignalResponse.action` | `RiskSignalResponse.action` | `RiskSignalResponse.action` = 4 |
| trace ID | `riskSignalResponse.traceId` | `RiskSignalResponse.traceId` | `RiskSignalResponse.traceId` = 5 |

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalidRiskSignal` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocationUnavailable` | `UNAVAILABLE` | Revocation writes cannot be performed safely; consumers must fail closed. |
