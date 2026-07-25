# Policy Decision v1alpha3 Contract Mapping

`api_version` is `aegivela.io/v1alpha3`; the protobuf service is `PolicyDecisionService.Evaluate`. v1alpha3 preserves every v1alpha2 field number and type. `PolicyDecisionRequest.parent_execution_grant` is an optional string at protobuf field `13`.

For `delegated_api`, `parent_execution_grant` is required and `parent_authority` is prohibited on HTTP requests. The compact grant is verified for signature, configured issuer, expiry, audience, structured-resource digest, and tenant-scoped revocation before principal resolution or policy evaluation. A verified grant exclusively supplies the parent authority: JTI, exact action/resource, scope, audience, expiry, and task binding. Action/resource equality and monotonic scope/audience/expiry/task binding are enforced before evaluation.

v1alpha3 uses the v1alpha2 structured-resource digest rule. v1alpha1 and v1alpha2 contracts and artifacts are immutable.

## Transport Errors

| HTTP status and error | Meaning |
| --- | --- |
| `401 unauthenticated` | The parent grant is missing, malformed, expired, revoked, incorrectly issued, or bound to another tenant or subject. |
| `403 policy_request_denied` | A verified parent grant would be expanded. |
| `503 parent_execution_grant_verification_unavailable` | Verification or revocation could not be safely completed. |
