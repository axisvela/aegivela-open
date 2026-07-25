# Security Evidence Contract Mapping v1alpha1

`api_version` is the literal `aegivela.io/v1alpha1`.

AEGIVELA evidence is a redacted, immutable record of a security-relevant event. Every domain (decision, grant, approval, attestation, revocation, risk, gateway, exchange, identity) emits exactly one envelope per terminal outcome. The envelope never carries raw bearer tokens, workload assertions, credentials, unrestricted resource attributes, unrestricted argument payloads, or human-readable reason text.

## Field Semantics

| Field | Required | Purpose |
| --- | --- | --- |
| `api_version` | yes | Contract version |
| `evidence_id` | yes | Unique event identifier |
| `tenant_id` | yes | Tenant isolation key |
| `domain` | yes | Which AEGIVELA subsystem emitted the event |
| `outcome` | yes | `allow`, `deny`, `approval_required`, `revoked`, `unavailable`, `revoked_subject`, `logged` |
| `reason_code` | yes | Machine-readable reason (`scope_denied`, `revoked`, `expired`, `risk_escalated`, etc.) |
| `trace_id` | yes | Request correlation identifier |
| `decision_id` | no | Linked policy decision ID |
| `grant_jti` | no | Linked execution grant JTI |
| `approval_jti` | no | Linked approval record JTI |
| `policy_version` | no | Policy version that produced the decision |
| `subject_digest` | no | SHA-256 of the subject reference |
| `actor_id` | no | Direct actor identifier |
| `action` | no | The action being evaluated |
| `resource_digest` | no | SHA-256 of the resource identity |
| `resource_kind` | no | Resource type descriptor |
| `risk_level` | no | `low`, `medium`, `high`, `critical` |
| `risk_ref` | no | Risk signal reference |
| `audience` | no | Intended audience for the action |
| `evidence_refs` | no | Linked evidence references for correlation chain |
| `timestamp` | yes | RFC 3339 UTC when the event occurred |

## Export / SIEM Integration

The envelope is a self-contained JSON object suitable for:
- Streaming to a SIEM via structured log agents (Fluentd, Vector, Logstash)
- Persistence in a product audit index (MODUREGIS Audit Index) alongside its native events
- Correlation across repositories using `trace_id`, `decision_id`, and `grant_jti`

AEGIVELA emits evidence through a `Recorder` interface instantiated per domain. The recorder's output format matches this schema.
