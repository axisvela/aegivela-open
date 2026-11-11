# AEGIVELA Authorization Modes

## Principal Model

AEGIVELA keeps the authority being exercised separate from the actor that makes a request and the workload that runs it:

```text
subject:        whose authority permits the effect
actor:          the human, agent, or workload directly making the request
client:         the OAuth client or registered application
workload:       the attested runtime/device proving the client or agent instance
authority root: a human principal or an organization/system principal
```

The Identity Bridge derives these from trusted issuer claims, registered client metadata, Agent Identity Authority records, and workload attestation. A caller cannot select them with request headers or claims it controls.

## Supported Modes

| Mode | Typical path | Authority | Required proof | Effective scope |
| --- | --- | --- | --- | --- |
| `human_web` | Browser -> Web app -> API | Human subject and actor | OIDC authorization code with PKCE | Human permission intersect requested scope and policy |
| `system_api` | Service/workload -> API Gateway -> API | Organization/system root | Client credential plus workload identity or mTLS | System grant intersect workload grant, requested scope, and policy |
| `delegated_api` | Service -> API Gateway -> API for a user | Human subject; service/workload actor | User delegation artifact plus client/workload proof | Human permission intersect client grant, delegated scope, requested scope, and policy |
| `service_agent_api` | Service Agent -> Gateway -> API | Organization/system root; Service Agent actor | Agent registration plus workload attestation | Organization authority intersect agent envelope, workload grant, requested scope, and policy |
| `twin_agent_api` | User Twin -> Gateway -> API | Human master; Twin Agent actor | Immutable master binding plus workload attestation | Master permission intersect Twin envelope, workload grant, tool permission, requested scope, and policy |

The v1alpha1 Policy Decision contract validates these mode boundaries at the transport layer. `human_web` requires a subject and forbids agent, master, and organization-root identities. `system_api` requires an organization root, client, and workload and forbids subject and master identities. `delegated_api` requires subject, client, workload, and parent authority while forbidding agent and master identities. `service_agent_api` requires a Service Agent identity, organization root, workload, and lifecycle epoch without a master. `twin_agent_api` requires a Twin identity, master, workload, and lifecycle epoch without an organization root.

Every protected endpoint declares which modes and artifact types it accepts. A workload token cannot impersonate a user; a Service Agent cannot claim a human master; and a Twin grant cannot satisfy a system-only endpoint.

## V2 Policy SDK Boundary

Updated 2026-10-15: the camelCase `aegivela.io/v2.0` HTTP principal carries
`namespace`, `authorityBinding`, `agentEpoch` and `identityEpoch`. Service agents
require `organizationRoot` binding; Twins require `humanMaster`. The validator
uses both positive registry epochs for v2 agents, with no fallback to the old
single lifecycle epoch. Validated request snapshots deep-copy both epoch
pointers so later caller mutation cannot change established bindings.

The Go PEP SDK represents this input as `PolicyPrincipalV2`, with
`ParentAuthority` for service/twin-agent scope, audience, expiry and task bounds.
Requested expiry is caller-controlled within those bounds; the default is
capped by the parent. Human/system modes prohibit raw parent authority, and
delegated mode continues to require the cryptographically verified
`parentExecutionGrant` path. Principal values remain comparison hints: trusted
workload verification and the identity source supply the authoritative values.

`TestSDKV2AgentEvaluateAgainstHandler` exercises the real SDK, HTTP decoder,
trusted-principal resolver, validator and static evaluator. Workload verification
and identity lookup are test stubs, not a deployed EIDOVELA/AEGIVELA acceptance
claim. Negative cases cover stale dual epochs, namespace/tenant mismatches,
invalid workload proof and identity-source outages.

Raw reference evaluation may omit the signed artifact. The configured production
v2 path now uses `SigningService.Evaluate`: one PDP evaluation, capped verified
credential lifetime and a signed result, with no unsigned fallback on failure.
The SDK still exposes unverified data until artifact/binding verification.
Caller-supplied parent bounds are not independently signed delegation proof.

The policy v2 JSON Schema/OpenAPI and fixtures now use v2 identifiers and
camelCase modes, with required agent authority bindings and dual epochs. The
request contract includes bearer/workload proof fields; optional protobuf fields
14/15 preserve the existing field numbers, and Go bindings were regenerated
with protoc-gen-go v1.36.11. The public open contract mirrors these artifacts.
SDK serialization for all five modes is checked against the public JSON Schema.
Every invalid fixture must become schema-valid when only its named defect is
repaired; schema rejection alone is not evidence of semantic attenuation.

Regeneration used `go run github.com/bufbuild/buf/cmd/buf@v1.55.1 generate --path
contracts/policy/v2.0/policy-decision.proto` and the checked-in pinned plugin
configuration. Go selected toolchain 1.26.9 for that generator command; product
Go module versions were not changed. Generated descriptors were not hand-edited.

Evaluate and sign reject legacy snake_case mode/binding input and incomplete
mode-shaped principal hints before resolution. Required agent epochs/authority
fields cannot be omitted and silently replaced by identity lookup results.
Human/system raw parent authority is
rejected. V2 delegated requests now enforce exact parent action/resource as well
as scope/audience/expiry/task attenuation before policy evaluation or signing.
Regression tests cover rejected action/resource expansion and unchanged positive
requests using a stub verified-grant provider; they do not certify deployed
cryptographic grant verification.

Human resolution preserves namespace supplied by the verified identity source,
and rejects mismatching request hints. It does not infer namespace from tenant
or accept the request namespace as authority. The human-v2 integration test now
consumes the official typed SDK rather than extending the old backend SDK copy.

Trusted resolution now retains credential expiry and session references as
internal, non-JSON principal metadata. Human/workload proof intersection and
optional identity-context expiry may only narrow lifetime. The signer adapter
no longer manufactures expiry from requested expiry. Unknown validity denies
issuance; v2 response `signedDecision` occupies optional protobuf field 9.

V2 JWS verification binds actual service organization roots or Twin masters,
without inventing human subjects, plus namespace, actor/class, client/workload,
dual epochs, instance and attestation. It checks signature/key/issuer/audience,
resource digest from expected context, scope, time and revocation. V2 verify
does not accept relabelled older artifacts. Encoding/key-ID bounds are shared
by issuer and verifier; unverifiable oversized results are not returned.

The SDK/HTTP tests exercise real decision cryptography, with identity/workload
and revocation providers stubbed. Public grant issuance, fresh registry rechecks,
deployed provider acceptance and PRAXOVELA adapter migration remain separate.
See [signed artifact semantics](../contracts/decision/v2.0/signed-artifact.md)
for the exact acceptance limits, including session/task expectation and SDK
structured-resource coverage limits.

## Web and API Gateway

For `human_web`, the web application is an OIDC relying party. It redirects to the Enterprise IdP using authorization code with PKCE, consumes AEGIVELA's verified identity and policy result, and never owns passwords or a second session center.

The API Gateway acts as PEP for `system_api` and `delegated_api`. A delegated call requires both the human subject artifact and a verified calling service/workload. Its user authority is not inferred from a service account:

```text
delegated_api.effective_scope =
  human_scope intersection client_grant intersection delegated_scope intersection policy
```

## Agent-to-API Access

For a Twin, the Gateway validates `agent_id`, `master_id`, workload proof, lifecycle epoch, tool/skill attribution where present, and the bound scope/audience/expiry. One user may have many Twins, but each Twin has exactly one immutable master.

For a Service Agent, the Gateway validates `agent_id`, organization/system authority root, workload proof, lifecycle epoch, and its service envelope. It does not require or fabricate `master_id`.

```text
twin_agent_api.effective_scope =
  master_scope intersection twin_envelope intersection workload_grant intersection tool_permission intersection policy

service_agent_api.effective_scope =
  organization_authority intersection service_agent_envelope intersection workload_grant intersection policy
```

In every mode, scope, audience, expiry, target resource, and task binding may only narrow from established authority. A missing identity, policy, attestation, lifecycle, or revocation check denies the request.

## Audit Semantics

Evidence records include `authorization_mode`, subject/authority root, actor, client, workload, action, resource, policy version, decision ID, and trace ID. Twin and Service Agent records include agent ID; Twin records include `master_id`. Records exclude raw bearer tokens, credentials, and unrestricted arguments.

This makes audit reconstruction unambiguous: human web action, organization system action, system action on behalf of a human, Twin action on behalf of its master, and Service Agent action under organization authority are distinct event types.
