# Tool Invocation Contract — v1alpha1

The `tool:invoke` adapter action binds an execution grant to an exact tool
implementation. A grant issued for `tool:invoke` carries a `tool_invariant`
claim (tool_id, skill_hash, implementation_digest) that a product PEP MUST
verify offline before dispatch.

## Wire Mapping

| Wire field | Go claim | Notes |
|---|---|---|
| `tool_id` | `ToolInvariant.ToolID` | Exact tool registry identifier |
| `skill_hash` | `ToolInvariant.SkillHash` | SHA-256 of the enrolled skill digest |
| `implementation_digest` | `ToolInvariant.ImplementationDigest` | SHA-256 of the tool binary/hash |
| `execution_id` | `TaskRef` / `ToolInvariant.TaskID` | Bound execution handle |
| `slo_class` | `ToolInvariant.SLOClass` | `pre_dispatch` \| `continuation` \| `connection` |

## Enforcement

- Offline: `executiongrant.enforceToolInvariants` rejects any invocation that
  widens verified tool bounds or runs an unverified implementation.
- Revocation: `pre_dispatch` and `continuation` SLO classes perform a
  cache-free authoritative recheck; an unavailable check fails closed.
- `connection` is eligible for the cached offline decision window and rechecks
  when a checker is available.
