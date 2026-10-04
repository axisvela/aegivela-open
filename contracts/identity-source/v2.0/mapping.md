# Verified Principal Contract Mapping

`contracts/identity-source/v2.0/verified-principal.schema.json` is the wire
form of the normalized verified principal that AEGIVELA may act on. It carries
references and normalized fields only; raw credentials and unrestricted claims
never appear in it. Only an identity-source `Provider` may produce this value.

## Field Mapping

| Schema property | `identitysrc.VerifiedPrincipal` | Source | Notes |
| --- | --- | --- | --- |
| `tenantId` | `TenantID` | AEGIVELA mapping | Technical isolation key; derived from `namespace` by configuration. |
| `namespace` | `Namespace` | NOMIVELA | Semantic Agent authority namespace. |
| `subjectRef` | `SubjectRef` | Enterprise IdP | Human subject reference when the agent acts for a human. |
| `actorId` | `ActorID` | EIDOVELA | Authenticated actor reference. |
| `clientId` | `ClientID` | EIDOVELA | Calling client reference. |
| `agentId` | `AgentID` | NOMIVELA | Registered Agent identifier. |
| `agentRef` | `AgentRef` | NOMIVELA | Stable Agent reference. |
| `agentClass` | `AgentClass` | NOMIVELA | `twin`, `service`, `ephemeral`, or `simulation`. |
| `authorityBinding` | `AuthorityBinding` | NOMIVELA | `humanMaster` (Twin) or `organizationRoot` (Service). |
| `authorityRoot` | `AuthorityRoot` | NOMIVELA | Immutable master or organization root reference. |
| `masterId` | `MasterID` | NOMIVELA | Present for Twin Agents; absent for Service Agents. |
| `agentState` | `AgentState` | NOMIVELA | Only `active` is accepted. |
| `agentEpoch` | `AgentEpoch` | NOMIVELA | Agent business lifecycle epoch; changes invalidate decisions and grants. |
| `identityState` | `IdentityState` | NOMIVELA | `active` or `bound`. |
| `identityEpoch` | `IdentityEpoch` | NOMIVELA | Agent ID security lifecycle epoch; changes invalidate that Agent ID's artifacts. |
| `instanceId` | `InstanceID` | EIDOVELA | Workload instance reference. |
| `attestationRef` | `AttestationRef` | EIDOVELA | Reference to verified attestation evidence. |
| `credentialGeneration` | `CredentialGeneration` | EIDOVELA | Active credential generation. |
| `identityAssurance` | `IdentityAssurance` | EIDOVELA | Optional assurance level. |
| `sessionRef` | `SessionRef` | EIDOVELA | Optional session reference. |
| `issuer` | `Issuer` | EIDOVELA | Verified context issuer. |
| `issuedAt` | `IssuedAt` | EIDOVELA | Context issuance time. |
| `expiresAt` | `ExpiresAt` | EIDOVELA | Context expiry time. |

## Error Taxonomy

| Error | HTTP | Meaning |
| --- | --- | --- |
| `ErrUnauthenticated` | `401` | The presenter credential is missing, malformed, or unverifiable. |
| `ErrDenied` | `403` | The verified context is not authorized for the requested audience. |
| `ErrUnavailable` | `503` | The identity source, namespace mapping, or a dependency is unavailable; fail closed. |
| `ErrInvalidContext` | `503` | The context is missing a required invariant (dual state, dual epoch, `attestationRef`, or `credentialGeneration`). |

Every failure is fail-closed. `namespace` maps deterministically to exactly one
`tenant`; an unknown or unmapped namespace fails closed.
