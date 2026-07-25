# Revocation Check Contract Mapping v1alpha1

`api_version` is the literal `aegivela.io/v1alpha1`; the protobuf service is `RevocationService.Check`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `revocation_check_request.api_version` | `RevocationCheckRequest.api_version` | `RevocationCheckRequest.api_version` = 1 |
| SLO class | `revocation_check_request.class` | `RevocationCheckRequest.class` | `RevocationCheckRequest.class` = 2 |
| selectors | `revocation_check_request.selectors` | `RevocationCheckRequest.selectors` | `RevocationCheckRequest.selectors` = 3 |
| selector kind | `selector.kind` | `Selector.kind` | `Selector.kind` = 1 |
| selector value | `selector.value` | `Selector.value` | `Selector.value` = 2 |
| lifecycle epoch | `revocation_check_request.lifecycle_epoch` | `RevocationCheckRequest.lifecycle_epoch` | `RevocationCheckRequest.lifecycle_epoch` = 4 |
| trace ID | `revocation_check_request.trace_id` | `RevocationCheckRequest.trace_id` | `RevocationCheckRequest.trace_id` = 5 |
| response API version | `revocation_check_response.api_version` | `RevocationCheckResponse.api_version` | `RevocationCheckResponse.api_version` = 1 |
| SLO class | `revocation_check_response.class` | `RevocationCheckResponse.class` | `RevocationCheckResponse.class` = 2 |
| outcome | `revocation_check_response.outcome` | `RevocationCheckResponse.outcome` | `RevocationCheckResponse.outcome` = 3 |
| checked at | `revocation_check_response.checked_at` | `RevocationCheckResponse.checked_at` | `RevocationCheckResponse.checked_at` = 4 |
| trace ID | `revocation_check_response.trace_id` | `RevocationCheckResponse.trace_id` | `RevocationCheckResponse.trace_id` = 5 |

SLO classes are `pre_dispatch`, `continuation`, and `connection`. `pre_dispatch` and `continuation` bypass the negative cache; `connection` may serve a bounded-TTL negative-cache hit. Tenant is derived from the internal caller authority and is never caller-supplied.

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalid_revocation_check_request` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocation_unavailable` | `UNAVAILABLE` | Revocation status cannot be safely checked; consumers must fail closed. |
