# Workload Assertions API

`v1alpha1` workload bindings and assertions are internal AEGIVELA APIs. They bind a client and workload to the tenant and organization authority root resolved from trusted internal context.

`POST /v1/workload-bindings` and `POST /v1/workload-assertions/issue` require exactly one valid `X-Aegivela-Internal-Token` plus exactly one tenant, actor, and authority-root trusted-context header. Bodies must supply `actor_ref` and `organization_authority_root` that exactly match the trusted context; they cannot supply `tenant_id` or a principal. Unknown fields and trailing JSON are rejected with `400`.

Binding registration accepts matching `actor_ref` and `organization_authority_root`, `client_id`, `workload_id`, opaque `evidence_ref`, and `trace_id`. The server derives the tenant, creates the binding identifier, and stores an active immutable binding linked to the trusted actor and root. Assertion issuance repeats the matching actor/root values with a client/workload pair, audience, expiry, and trace ID. It finds an active binding linked to that trusted authority and emits a compact Ed25519 assertion only in a successful `200` response. Assertion expiry must be in the future and no more than five minutes from issue time.

Invalid authentication or inconsistent trusted authority is `401 unauthenticated`; malformed requests and invalid expiry are `400`; unavailable binding or key dependencies are `503`. Error responses and logs do not include raw assertions. `GET /v1/workload-assertions/jwks` returns only public `OKP` / `Ed25519` signing material.

The published schema, OpenAPI description, fixtures, and compatibility baseline are under `contracts/workload/v1alpha1/`.
