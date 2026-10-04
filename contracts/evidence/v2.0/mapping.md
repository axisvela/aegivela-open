# Security Evidence Contract Mapping v2.0

`apiVersion` is the literal `aegivela.io/v2.0`.

AEGIVELA evidence is a redacted, immutable record of a security-relevant event. Every domain (decision, grant, approval, attestation, revocation, risk, gateway, exchange, identity) emits exactly one envelope per terminal outcome. The envelope never carries raw bearer tokens, workload assertions, credentials, unrestricted resource attributes, unrestricted argument payloads, or human-readable reason text.

## Field Semantics

| Field | Required | Purpose |
| --- | --- | --- |
| `apiVersion` | yes | Contract version |
| `evidenceId` | yes | Unique event identifier |
| `tenantId` | yes | Tenant isolation key |
| `namespace` | no | Verified Agent authority namespace, for namespace-scoped correlation |
| `agentEpoch` | no | NOMIVELA Agent business lifecycle epoch bound to the event |
| `identityEpoch` | no | NOMIVELA Agent ID security lifecycle epoch bound to the event |
| `domain` | yes | Which AEGIVELA subsystem emitted the event |
| `outcome` | yes | `allow`, `deny`, `approvalRequired`, `revoked`, `unavailable`, `revokedSubject`, `logged` |
| `reasonCode` | yes | Machine-readable reason (`scopeDenied`, `revoked`, `expired`, `riskEscalated`, etc.) |
| `traceId` | yes | Request correlation identifier |
| `decisionId` | no | Linked policy decision ID |
| `grantJti` | no | Linked execution grant JTI |
| `approvalJti` | no | Linked approval record JTI |
| `policyVersion` | no | Policy version that produced the decision |
| `subjectDigest` | no | SHA-256 of the subject reference |
| `actorId` | no | Direct actor identifier |
| `action` | no | The action being evaluated |
| `resourceDigest` | no | SHA-256 of the resource identity |
| `resourceKind` | no | Resource type descriptor |
| `riskLevel` | no | `low`, `medium`, `high`, `critical` |
| `riskRef` | no | Risk signal reference |
| `audience` | no | Intended audience for the action |
| `evidenceRefs` | no | Linked evidence references for correlation chain |
| `timestamp` | yes | RFC 3339 UTC when the event occurred |

## Export / SIEM Integration

The envelope is a self-contained JSON object suitable for:
- Streaming to a SIEM via structured log agents (Fluentd, Vector, Logstash)
- Persistence in a product audit index (MODUREGIS Audit Index) alongside its native events
- Correlation across repositories using `traceId`, `namespace`, `decisionId`, and `grantJti`

AEGIVELA emits evidence through a `Recorder` interface instantiated per domain. The recorder's output format matches this schema. One `securityevidence/postgres` recorder persists the same envelope to the append-only `evidenceEvents` table, indexed by `tenantId`, `namespace`, `traceId`, `decisionId`, `grantJti`, `approvalJti`, and `policyVersion` so Decision, Grant, Approval, and PEP events correlate ([ADR-0014](../../../docs/adr/0014-persist-sanitized-security-evidence.md)); log and SIEM export continue alongside it.
