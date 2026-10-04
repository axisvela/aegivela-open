# Identity Boundary

## Family boundary

```text
Agent registry           = NOMIVELA
Identity / authentication = EIDOVELA
Authorization / delegation = AEGIVELA
```

AEGIVELA is the **Agent Authorization & Delegation Fabric**. It does not own an
Agent registry, an Agent lifecycle, an authentication authority, a password
store, MFA, a directory, or browser sessions. It consumes a verified principal
and owns authorization, delegation, approval, execution grants, the PEP seam,
authorization revocation and audit evidence.

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

AEGIVELA writes no lifecycle. It binds both epochs to decisions, delegation
contracts and execution grants. A change in either epoch invalidates or forces a
recheck of the affected artifacts.

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
| `oidc` | Generic enterprise OIDC/JWKS IdP | Extracted from the identity bridge |
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
([ADR-0016](adr/0016-agent-iam-spec-conformance-authority.md)). Remaining AEGIVELA
Part 4 blockers stay tracked in `docs/roadmap.md`: canonical allow lineage in
token exchange, namespace- and freshness-consistent revocation checks, unified
lifecycle schema, and evidence rejection-event persistence.
