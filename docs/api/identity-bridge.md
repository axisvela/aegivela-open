# Identity Bridge Integration Guide

The Identity Bridge validates enterprise-issued OIDC JWTs for configured issuers and returns a normalized principal. It is an internal adapter/PEP API, not a public login API.

## Issuer Configuration

Set `AEGIVELA_IDENTITY_ISSUERS_FILE` to the path of a JSON file containing an issuer array. AEGIVELA reads at most 1 MiB from this file. Within that read limit, it uses strict JSON decoding: unknown fields are rejected, and a second JSON value or trailing non-whitespace content is rejected. Content after the 1 MiB read limit is not evaluated. The array must contain at least one issuer.

```json
[
  {
    "issuer_url": "https://issuer.example",
    "discovery_url": "https://issuer.example/.well-known/openid-configuration",
    "allowed_audiences": ["aegivela"],
    "allowed_algorithms": ["RS256"],
    "tenant_claim": "tenant_id",
    "subject_claim": "sub",
    "actor_claim": "actor_id",
    "client_id_claim": "client_id",
    "session_claim": "sid",
    "assurance_claim": "acr",
    "cache_ttl": "5m",
    "allow_insecure_localhost": false
  }
]
```

Each object may contain only the fields shown above. `issuer_url` and `discovery_url` must be absolute HTTPS URLs. HTTP is accepted only when `allow_insecure_localhost` is `true` and the hostname is exactly `localhost`; this is for local testing only. `allowed_audiences`, `allowed_algorithms`, `tenant_claim`, `subject_claim`, `client_id_claim`, and a positive `cache_ttl` are required. `actor_claim`, `session_claim`, and `assurance_claim` may be omitted or empty. `allowed_algorithms` accepts only `RS256`, `RS384`, `RS512`, `ES256`, `ES384`, and `ES512`. Issuer URLs must be unique.

The token `iss` selects only an already configured issuer. It never enables dynamic issuer discovery. Discovery must report the configured issuer and a valid JWKS URL.

## Resolve Endpoint

`POST /v1/identity/resolve` accepts a bearer JWT and its intended audience:

```json
{
  "bearer_token": "<enterprise-oidc-jwt>",
  "audience": "aegivela",
  "trace_id": "trace-123"
}
```

`bearer_token` and `audience` are required and non-blank. `trace_id` is optional. When present, it must be empty or a 1-128 byte ASCII identifier whose first byte is alphanumeric and whose remaining bytes are alphanumeric, `.`, `_`, or `-`; invalid values return `400 invalid_identity_request`. The request body accepts no other fields and is limited to 1 MiB.

On success, the response is the normalized principal:

```json
{
  "tenant_id": "tenant-a",
  "subject_ref": "subject-a",
  "actor_id": "actor-a",
  "issuer": "https://issuer.example",
  "client_id": "client-a",
  "identity_assurance": "high",
  "session_ref": "session-a",
  "issued_at": "2026-07-16T12:00:00Z",
  "expires_at": "2026-07-16T13:00:00Z"
}
```

`identity_assurance` and `session_ref` are omitted when their configured claims are absent. The response never contains the bearer JWT, raw claims, credentials, or evidence.

The endpoint requires exactly one non-blank `X-Aegivela-Internal-Token` header whose value exactly matches `AEGIVELA_INTERNAL_AUTH_TOKEN`. It does not require `X-Aegivela-Tenant-ID`, `X-Aegivela-Actor-ID`, or `X-Aegivela-Authority-Root`; those headers do not influence the resolved principal. The authenticated adapter or PEP must keep this token private and must not log it.

Error responses are JSON objects with only a `code` field:

| Status | Body | Condition |
| --- | --- | --- |
| 400 | `{"code":"invalid_identity_request"}` | Invalid JSON, unknown request field, trailing content, blank/missing audience, or invalid trace ID. |
| 401 | `{"code":"unauthenticated"}` | Missing, blank, duplicate, or invalid internal token; missing/blank bearer token; invalid JWT, issuer, algorithm, audience, signature, time claims, or mapped claim. |
| 413 | `{"code":"request_too_large"}` | Request body exceeds 1 MiB. |
| 503 | `{"code":"identity_unavailable"}` | Discovery or JWKS cannot be obtained or refreshed. |

## Verification And Caching

The bridge accepts only configured issuers and configured asymmetric signing algorithms. It verifies the JWT signature; exact issuer and requested audience; and `exp`, `nbf`, and `iat`. It maps configured tenant, subject, actor, client, session, and assurance claims into the response. The tenant and subject claims must be non-empty strings. The client claim must also be non-empty. When the configured actor claim is absent, `actor_id` defaults to `subject_ref`; an empty present actor claim is rejected.

Discovery and JWKS are cached per issuer for that issuer's `cache_ttl`. Each discovery or JWKS HTTP request has a 5-second context deadline, redirects are not followed, and each response is capped at 1 MiB; a non-`200` status, timeout, malformed document, or oversized response makes identity resolution unavailable. A still-valid cached key set is used for a matching `kid`. On an unknown `kid`, the bridge attempts one synchronous JWKS refresh and retries the lookup once. Unknown-`kid` refreshes are coalesced for an issuer and then suppressed for 30 seconds; a token with an unknown key during that interval is unauthenticated. An expired cache is not used if refresh fails: the request returns `503 identity_unavailable`.

## Evidence Redaction

The configured API process emits redacted `ResolveEvidence` as JSON log records through the standard logger for successful, unauthenticated, and unavailable identity resolutions. The evidence contains the issuer, tenant ID, SHA-256 digest of the subject reference, client ID, validated trace ID, outcome, and reason code. It excludes raw bearer JWTs, credentials, full claim payloads, and JWKS documents. Do not add bearer tokens or raw claims to logs, audit records, or evidence envelopes.

## Local Testing

For a local OIDC test issuer, set both issuer URLs to `http://localhost/...` and set `allow_insecure_localhost` to `true`. Do not use that setting outside local testing. Configure `AEGIVELA_IDENTITY_ISSUERS_FILE` along with `DATABASE_URL` and `AEGIVELA_INTERNAL_AUTH_TOKEN`, apply migrations, and run:

```sh
go run ./backend/cmd/aegivela-api
```

Run the backend test suite with:

```sh
go test -count=1 ./backend/...
go vet ./backend/...
```

## Non-Goals

The Identity Bridge does not implement browser login, OIDC authorization-code redirects, PKCE, refresh tokens, logout, password or MFA handling, SAML assertion parsing, opaque-token introspection, SCIM, or directory synchronization.
