# Revocation Check Contract Mapping v1alpha1

`apiVersion` is the literal `aegivela.io/v2.0`; the protobuf service is `RevocationService.Check`.

| Semantic field | JSON Schema property | OpenAPI property | Protobuf field number and name |
| --- | --- | --- | --- |
| request API version | `revocationCheckRequest.apiVersion` | `RevocationCheckRequest.apiVersion` | `RevocationCheckRequest.apiVersion` = 1 |
| SLO class | `revocationCheckRequest.class` | `RevocationCheckRequest.class` | `RevocationCheckRequest.class` = 2 |
| selectors | `revocationCheckRequest.selectors` | `RevocationCheckRequest.selectors` | `RevocationCheckRequest.selectors` = 3 |
| selector kind | `selector.kind` | `Selector.kind` | `Selector.kind` = 1 |
| selector value | `selector.value` | `Selector.value` | `Selector.value` = 2 |
| lifecycle epoch | `revocationCheckRequest.lifecycleEpoch` | `RevocationCheckRequest.lifecycleEpoch` | `RevocationCheckRequest.lifecycleEpoch` = 4 |
| trace ID | `revocationCheckRequest.traceId` | `RevocationCheckRequest.traceId` | `RevocationCheckRequest.traceId` = 5 |
| response API version | `revocationCheckResponse.apiVersion` | `RevocationCheckResponse.apiVersion` | `RevocationCheckResponse.apiVersion` = 1 |
| SLO class | `revocationCheckResponse.class` | `RevocationCheckResponse.class` | `RevocationCheckResponse.class` = 2 |
| outcome | `revocationCheckResponse.outcome` | `RevocationCheckResponse.outcome` | `RevocationCheckResponse.outcome` = 3 |
| checked at | `revocationCheckResponse.checkedAt` | `RevocationCheckResponse.checkedAt` | `RevocationCheckResponse.checkedAt` = 4 |
| trace ID | `revocationCheckResponse.traceId` | `RevocationCheckResponse.traceId` | `RevocationCheckResponse.traceId` = 5 |

SLO classes are `preDispatch`, `continuation`, and `connection`. `preDispatch` and `continuation` bypass the negative cache; `connection` may serve a bounded-TTL negative-cache hit. Tenant is derived from the internal caller authority and is never caller-supplied.

All endpoints require the `X-AEGIVELA-PEP` internal header. This is the first public v1alpha1 protobuf baseline.

## Transport Errors

| HTTP status and error | gRPC status | Meaning |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | The caller is not authenticated. |
| `400 invalidRevocationCheckRequest` | `INVALID_ARGUMENT` | The request cannot be validated against this contract. |
| `503 revocationUnavailable` | `UNAVAILABLE` | Revocation status cannot be safely checked; consumers must fail closed. |
