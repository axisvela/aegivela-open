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
