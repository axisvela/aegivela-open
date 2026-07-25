# Policy Decision Contract

## Status and Authority

The F0 Policy Decision contract is `aegivela.io/v1alpha1`. Its canonical normative definition is the JSON Schema at [`contracts/policy/v1alpha1/policy-decision.schema.json`](../../contracts/policy/v1alpha1/policy-decision.schema.json). It defines a request or response document, rejects unknown properties, and fixes the API version literal. v1alpha1 is immutable: it supports only `resource.kind` and `resource.reference`, and rejects every structured-resource field.

The OpenAPI 3.1 document at [`contracts/policy/v1alpha1/policy-decision.openapi.yaml`](../../contracts/policy/v1alpha1/policy-decision.openapi.yaml) maps that contract to `POST /v1/policy/decisions:evaluate`. The protobuf service at [`contracts/policy/v1alpha1/policy-decision.proto`](../../contracts/policy/v1alpha1/policy-decision.proto) maps it to `policy.v1alpha1.PolicyDecisionService.Evaluate`. The maintained field-by-field mapping is [`contracts/policy/v1alpha1/mapping.md`](../../contracts/policy/v1alpha1/mapping.md). This is the first public v1alpha1 protobuf baseline. `AUTHORIZATION_MODE_UNSPECIFIED=0` and `DECISION_OUTCOME_UNSPECIFIED=0` are invalid sentinels; omitted enum fields decode to those values and adapters reject them. Valid modes and outcomes have nonzero assignments.

v1alpha2 is a distinct contract at [`contracts/policy/v1alpha2`](../../contracts/policy/v1alpha2). Its schema, OpenAPI document, protobuf service, mapping, baseline, and fixtures are version-specific. Evaluation routes exclusively by the literal `api_version`; consumers must never infer a version from resource fields.

## Request and Response

A request always supplies `api_version`, `authorization_mode`, `principal`, `action`, `resource`, non-empty unique `requested_scope`, `audience`, `requested_expires_at`, and `trace_id`. `principal` always supplies tenant and actor references. `parent_authority` records the authority reference, allowed scope and audience, expiry, and task binding when a mode requires delegation. The response always supplies `api_version`, `decision_id`, `outcome`, non-empty `policy_version`, `expires_at`, `obligations`, and at least one `evidence_refs` value.

`tenant_id` is security scoped input, not a caller-selectable routing hint. PEPs must derive and validate principals from trusted identity, client, workload, and Agent Identity Authority sources. Evidence references must not contain raw bearer tokens, credentials, unrestricted arguments, or raw sensitive payloads.

## Structured Resources and Artifacts

v1alpha2 retains legacy `resource.kind` and `resource.reference` to support migration. It adds an all-or-none structured set: `product_id`, `resource_type`, `descriptor_id`, `descriptor_version`, and `attributes`. When that set is present, the PDP validates it against the registered descriptor and returns `structured_resource_digest`, a canonical digest binding the product, resource type, descriptor identity/version, resource reference, and descriptor-validated attributes. A legacy resource uses a domain-separated `legacy-v1alpha2` canonical digest over `kind` and `reference`. A partial set, an unknown descriptor, or invalid attributes is `invalid_policy_request`; descriptor validation unavailability fails closed.

A v1alpha2 decision response and every execution grant derived from it must carry a nonempty digest. PEPs derive the applicable structured or `legacy-v1alpha2` digest from their trusted current resource instance before execution and reject a missing, unexpected, or unequal digest. They record only the digest and descriptor identifiers in evidence, never raw structured attributes.

Artifacts are version-bound. A v1alpha1 decision or grant cannot authorize a v1alpha2 structured-resource PEP. A v1alpha2 structured artifact must not be downgraded to a legacy artifact or reused for a different descriptor, resource instance, or attribute set.

## Authorization Modes

| Mode | Required principal boundary | Parent authority |
| --- | --- | --- |
| `human_web` | Human `subject_ref`; no agent, master, or organization root | Not required |
| `system_api` | Organization root, `client_id`, and `workload_ref`; no human subject or master | Not required |
| `delegated_api` | Human subject plus client and workload; no agent or master | Required |
| `service_agent_api` | Service `agent_id`, organization root, workload, and lifecycle epoch; no master | Required |
| `twin_agent_api` | Twin `agent_id`, immutable `master_id`, workload, and lifecycle epoch; no organization root | Required |

Every protected endpoint declares accepted modes. A credential or decision produced for one mode must not satisfy an endpoint accepting another mode.

## Parent Attenuation

`delegated_api`, `service_agent_api`, and `twin_agent_api` require `parent_authority`. Child authority may only narrow from that parent:

- Every requested scope must be in `parent_scope`.
- `audience` must be in `parent_audience`.
- `requested_expires_at` must not be later than `parent_expires_at`.
- `task_binding` must exactly equal `parent_task_binding`.

Missing or malformed parent authority is `invalid_policy_request`; a scope, audience, expiry, or task-binding widening is a denial. When a matching policy is available, evaluator transports represent a structurally parseable widening as a conforming `deny` response with empty obligations. PEPs must not repair, substitute, or widen parent values.

## Outcomes and Enforcement

The only execution-permitting outcome is `allow`. `deny`, `approval_required`, and `revoked` block execution. `approval_required` is not a deferred allow: a separate approved decision or grant is required before execution. `revoked` blocks execution even when another property appears valid. PEPs must apply all response obligations before executing an `allow`, retain the decision ID, policy version, trace ID, and evidence references for their security evidence, and enforce response expiry.

| Outcome | PEP action |
| --- | --- |
| `allow` | Execute only after obligations are satisfied and before `expires_at`. |
| `deny` | Do not execute. |
| `approval_required` | Do not execute; obtain a separately valid approval path. |
| `revoked` | Do not execute; invalidate any associated local authorization state. |

## Transport Errors

Policy denial is a `200` response with `outcome: deny`, `approval_required`, or `revoked`; it is not an HTTP or gRPC transport error.

| HTTP | gRPC | Meaning and PEP action |
| --- | --- | --- |
| `401 unauthenticated` | `UNAUTHENTICATED` | Caller identity is not authenticated. Do not execute. |
| `400 invalid_policy_request` | `INVALID_ARGUMENT` | Contract or policy input validation failed. Do not execute; correct the trusted integration. |
| `503 policy_unavailable` | `UNAVAILABLE` | A safe decision cannot be obtained. Fail closed; do not use a stale or default allow. |

Any malformed response, unknown enum, unsupported API version, failed schema validation, unavailable dependency, or expired decision is also a fail-closed condition.

## Static F0 Reference Evaluator

`backend/internal/policycontract` contains an in-memory static reference evaluator for conformance and local integration. It validates the version, structural fields, mode boundaries, parent presence, attenuation, request expiry, configured static policy, known outcomes, obligations, and evidence references. Structurally parseable denials become `deny` decisions only when a matching static policy is available; malformed requests and unavailable policy remain errors. It emits a decision ID and the selected static policy version.

It is not a production PDP. It does not validate OIDC issuers, signatures, audiences, or claims; resolve grants or approvals; perform revocation checks; validate real attestation or lifecycle state; persist evidence; or query a dynamic policy store. Those identity, OIDC, grant, approval, revocation, attestation, lifecycle, and production PDP dependencies remain deferred beyond F0. A production consumer must fail closed until those checks are available.

## Fixtures and Protobuf Baseline

The fixture corpus under [`contracts/policy/v1alpha1/fixtures`](../../contracts/policy/v1alpha1/fixtures) is part of the contract:

- `valid/` covers all five modes plus `deny`, `approval_required`, and `revoked` responses.
- `invalid/` covers unknown modes, incompatible principal combinations, and missing parent authority.
- `attenuation-invalid/` covers scope escalation, audience widening, expiry extension, and task-binding widening.

[`contracts/policy/v1alpha1/protobuf-baseline.json`](../../contracts/policy/v1alpha1/protobuf-baseline.json) records every protobuf field number, field type, and enum number. It is the first public v1alpha1 compatibility baseline; future releases must preserve its enum values, field numbers, and field types. Generated Go protobuf output is derived, not an editing target. Run the pinned resolver and linter from the repository root:

```powershell
powershell -ExecutionPolicy Bypass -File .\tools\buf.ps1 lint
```

## Consumer Requirements

MODUREGIS must send its declared mode and trusted tenant/actor context, map only an unexpired `allow` to the guarded operation, apply obligations, and record decision/evidence references. It must preserve its intentional `503 authorization_unconfigured` behavior when no AEGIVELA adapter is configured.

Web consumers must use `human_web` only after trusted OIDC relying-party validation and must not treat browser session state as a decision. API Gateway consumers must distinguish `system_api` from `delegated_api`, validate both the human delegation and calling client/workload for delegated calls, and reject cross-mode token substitution.

Agent Gateway consumers must use `service_agent_api` or `twin_agent_api` as applicable, validate immutable Twin master binding or Service Agent organization-root binding, lifecycle epoch, workload proof, parent attenuation, and decision expiry before each protected action. They must not execute when a required revocation, attestation, grant, or policy dependency is unavailable.

## Versioning and Migration

v1alpha1 consumers remain on their existing immutable contract; they must not send or accept structured-resource fields or artifacts. A product migrating to v1alpha2 first publishes and registers its descriptors, then sends the complete structured resource set, and enables its PEP digest recomputation and artifact-version checks before accepting structured decisions or grants. v1alpha2 can temporarily carry legacy-only resources; those requests receive the required `legacy-v1alpha2` canonical digest over `kind` and `reference`, but do not gain structured-resource authorization.

Adding an optional field is minor-version compatible only within a mutable contract version. A required-field change, removed or renumbered protobuf field, removed enum value, or any change to outcome or transport-error semantics is breaking and requires a new API version. v1alpha1 and v1alpha2 protobuf baselines freeze their published enum values, field numbers, and field types. New versions must publish updated JSON Schema, OpenAPI/protobuf mappings, protobuf baseline, positive and negative fixtures, and compatibility rules before consumers enable them.
