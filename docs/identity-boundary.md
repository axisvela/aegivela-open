# Identity Boundary

## Family boundary

```text
Agent registry           = NOMIVELA
Identity / authentication = EIDOVELA
Authorization / delegation = AEGIVELA
```

AEGIVELA is the **Agent Authorization & Delegation Fabric**. In the family
integration boundary it consumes a verified principal, rather than owning
Agent registry records, Agent lifecycle or the Agent authentication authority.
It owns authorization, delegation, approval, execution grants, the PEP seam,
authorization revocation and audit evidence. Enterprise IdPs retain human
authentication, passwords, MFA and directory ownership.

This is the target responsibility split, not a claim that all identity code
has been removed. The retained default `builtin` mode and local registration
handlers are described below. EE console sessions are local application
sessions, not Agent identity credentials or a family-wide browser session center.
Cross-product conclusions are maintained in
[Cross-Product Authority Boundaries](https://github.com/axisvela/axisvela/blob/main/AUTHORITY-BOUNDARIES.md).
The [F0 identity-authority design](agent-identity-authority.md) is historical.

## The identity-source seam

`backend/identitysrc` defines the seam:

- `VerifiedPrincipal` is the single normalized representation AEGIVELA may act
  on. It carries references and normalized fields only; raw credentials and
  unrestricted claims never appear in it.
- `Provider` is the only identity dependency of the transport and resolution
  layers. Only `Provider.Resolve` may construct a `VerifiedPrincipal`, so a
  request-supplied principal cannot become an identity by construction.
- Error taxonomy: `ErrUnauthenticated` (401), `ErrDenied` (403),
  `ErrUnavailable` (503). Every failure is fail-closed.

The wire form is published as
`contracts/identity-source/v1alpha1/verified-principal.schema.json`.

## Dual lifecycle

A verified context carries both registry epochs:

- `agent_epoch` — the NOMIVELA Agent business lifecycle.
- `identity_epoch` — the NOMIVELA Agent ID security lifecycle.

For externally sourced Agent identity, AEGIVELA consumes rather than writes
the NOMIVELA lifecycle. It binds both epochs to decisions, delegation contracts
and execution grants. A change in either epoch invalidates or forces a recheck
of the affected artifacts. This does not describe removal of the standalone
local lifecycle implementation.

A context missing dual state, dual epoch, `attestation_ref` or
`credential_generation` is rejected (`ErrInvalidContext`).

## Tenant and namespace

`namespace` is the verified Agent authority namespace; `tenant` is AEGIVELA's
technical isolation key. They are related by configuration, never by string
equality or caller input:

- Each identity source registers the namespaces it may assert, and each
  namespace maps deterministically to exactly one tenant.
- An unknown or unmapped namespace fails closed (`ErrUnavailable`).
- A caller-supplied namespace or tenant is comparison-only.

## Adapters

| Adapter | Source | Status |
|---|---|---|
| `builtin` | Current identity bridge, agent identity and workload identity | Existing behavior; kept as the standalone default |
| OIDC bridge | Generic enterprise OIDC/JWKS IdP | Identity bridge component, not an `AEGIVELA_IDENTITY_SOURCE=oidc` selector value |
| `eidovela` | EIDOVELA verified agent context and introspection | Implemented: `identitysrc/eidovela` |

### EIDOVELA adapter

`identitysrc/eidovela` resolves a credential by:

1. verifying the presenter credential through EIDOVELA introspection
   (`POST /v1/introspect`) with the presenter proof key;
2. reading `GET /v1/verified-agent-context` for the authoritative Agent,
   Agent Identity and Instance state and dual epochs;
3. reading `GET /v1/authentication-state/{agentID}` (or the active credential
   generation when no override exists) for the authentication state.

It derives the authority binding from the registry-enforced agent class
(`twin`/`ephemeral` → human master, `service` → organization root) rather than
from caller input, maps the namespace to the local tenant, and rejects any
context that fails the dual-lifecycle invariants.

The adapter is configured through `identitysrc/eidovela.Config` with a base URL,
HTTP client and a required `NamespaceMap`.

### Selecting the source

The composition root selects the source with `AEGIVELA_IDENTITY_SOURCE`:

- `builtin` (or unset) — the existing identity bridge, local agent registry and
  workload identity. Routes, responses, error codes and database behavior are
  unchanged.
- `eidovela` — the EIDOVELA verified agent context. It requires
  `AEGIVELA_EIDOVELA_URL` and a namespace-to-tenant mapping, supplied either as
  `AEGIVELA_EIDOVELA_NAMESPACES_FILE` (a JSON object of namespace to tenant) or
  as the pair `AEGIVELA_EIDOVELA_NAMESPACE` and `AEGIVELA_EIDOVELA_TENANT`.

An unknown source or incomplete configuration fails startup. With `eidovela`
selected, the agent authorization modes resolve identity, authority binding and
both registry epochs only from the source; the local agent registry is not
consulted for them.

This selection is limited to the Agent authorization identity-resolution path.
The composition root still constructs the local identity service and HTTP
handlers. It does not disable registration/lifecycle routes, migrate their
records or prove that no other path can write local identity state. Do not
configure a second writer for the same NOMIVELA-owned records.

Before claiming deployment convergence, record the selected source, supported
contract versions, remaining local write paths and their enforcement boundary.
Verify single-writer behavior, stale-epoch and revocation rejection, and
unavailable-source fail-closed behavior. Static configuration review or AgentIAM
fixture ownership alone does not establish those properties.

## Identity is not authorization

An identity credential proves who is calling and the current identity state; it
does not authorize an action. Execution authorization keeps the single lineage:

```text
verified principal -> canonical policy decision -> allow
  -> execution grant -> attenuated child grant -> PEP effect
```

A raw identity token is never accepted as an execution grant.

## Conformance boundary

This document describes an implementation mapping to the Agent IAM Series Part 3
`Verified Agent Identity Context` and Part 4 authorization/delegation boundary.
The `agent-iam-spec` normative text governs on any conflict
([ADR-0016](adr/0016-agent-iam-spec-conformance-authority.md)). Canonical allow
lineage, namespace/freshness-aware revocation, lifecycle schema alignment and
sanitized evidence persistence are implemented. Current verification limits and
consumer-owned acceptance remain tracked in [Conformance Status](conformance-status.md)
and the [roadmap](roadmap.md). In-repo conformance does not establish deployment
single-writer convergence or certify the identity components' own behavior.
