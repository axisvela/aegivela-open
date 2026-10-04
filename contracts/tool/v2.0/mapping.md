# Tool Invocation Contract — v2.0

The `tool:invoke` adapter action binds an execution grant to an exact tool
implementation. A grant issued for `tool:invoke` carries a `toolInvariant`
claim (toolId, skillHash, implementationDigest) that a product PEP MUST
verify offline before dispatch.

## Wire Mapping

| Wire field | Go claim | Notes |
|---|---|---|
| `toolId` | `ToolInvariant.ToolID` | Exact tool registry identifier |
| `skillHash` | `ToolInvariant.SkillHash` | SHA-256 of the enrolled skill digest |
| `implementationDigest` | `ToolInvariant.ImplementationDigest` | SHA-256 of the tool binary/hash |
| `executionId` | `TaskRef` / `ToolInvariant.TaskID` | Bound execution handle |
| `sloClass` | `ToolInvariant.SLOClass` | `preDispatch` \| `continuation` \| `connection` |

## Enforcement

- Offline: `executiongrant.enforceToolInvariants` rejects any invocation that
  widens verified tool bounds or runs an unverified implementation.
- Revocation: `preDispatch` and `continuation` SLO classes perform a
  cache-free authoritative recheck; an unavailable check fails closed.
- `connection` is eligible for the cached offline decision window and rechecks
  when a checker is available.
