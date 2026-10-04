# Verified Principal Contract Mapping

`contracts/identity-source/v1alpha1/verified-principal.schema.json` is the wire
form of the normalized verified principal that AEGIVELA may act on. It carries
references and normalized fields only; raw credentials and unrestricted claims
never appear in it. Only an identity-source `Provider` may produce this value.

## Field Mapping

| Schema property | `identitysrc.VerifiedPrincipal` | Source | Notes |
| --- | --- | --- | --- |
| `tenant_id` | `TenantID` | AEGIVELA mapping | Technical isolation key; derived from `namespace` by configuration. |
| `namespace` | `Namespace` | NOMIVELA | Semantic Agent authority namespace. |
| `subject_ref` | `SubjectRef` | Enterprise IdP | Human subject reference when the agent acts for a human. |
| `actor_id` | `ActorID` | EIDOVELA | Authenticated actor reference. |
| `client_id` | `ClientID` | EIDOVELA | Calling client reference. |
| `agent_id` | `AgentID` | NOMIVELA | Registered Agent identifier. |
| `agent_ref` | `AgentRef` | NOMIVELA | Stable Agent reference. |
| `agent_class` | `AgentClass` | NOMIVELA | `twin`, `service`, `ephemeral`, or `simulation`. |
| `authority_binding` | `AuthorityBinding` | NOMIVELA | `human_master` (Twin) or `organization_root` (Service). |
| `authority_root` | `AuthorityRoot` | NOMIVELA | Immutable master or organization root reference. |
| `master_id` | `MasterID` | NOMIVELA | Present for Twin Agents; absent for Service Agents. |
| `agent_state` | `AgentState` | NOMIVELA | Only `active` is accepted. |
| `agent_epoch` | `AgentEpoch` | NOMIVELA | Agent business lifecycle epoch; changes invalidate decisions and grants. |
| `identity_state` | `IdentityState` | NOMIVELA | `active` or `bound`. |
| `identity_epoch` | `IdentityEpoch` | NOMIVELA | Agent ID security lifecycle epoch; changes invalidate that Agent ID's artifacts. |
| `instance_id` | `InstanceID` | EIDOVELA | Workload instance reference. |
| `attestation_ref` | `AttestationRef` | EIDOVELA | Reference to verified attestation evidence. |
| `credential_generation` | `CredentialGeneration` | EIDOVELA | Active credential generation. |
| `identity_assurance` | `IdentityAssurance` | EIDOVELA | Optional assurance level. |
| `session_ref` | `SessionRef` | EIDOVELA | Optional session reference. |
| `issuer` | `Issuer` | EIDOVELA | Verified context issuer. |
| `issued_at` | `IssuedAt` | EIDOVELA | Context issuance time. |
| `expires_at` | `ExpiresAt` | EIDOVELA | Context expiry time. |

## Error Taxonomy

| Error | HTTP | Meaning |
| --- | --- | --- |
| `ErrUnauthenticated` | `401` | The presenter credential is missing, malformed, or unverifiable. |
| `ErrDenied` | `403` | The verified context is not authorized for the requested audience. |
| `ErrUnavailable` | `503` | The identity source, namespace mapping, or a dependency is unavailable; fail closed. |
| `ErrInvalidContext` | `503` | The context is missing a required invariant (dual state, dual epoch, `attestation_ref`, or `credential_generation`). |

Every failure is fail-closed. `namespace` maps deterministically to exactly one
`tenant`; an unknown or unmapped namespace fails closed.
