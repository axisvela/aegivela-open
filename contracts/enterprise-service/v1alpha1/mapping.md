# Enterprise Service Authorization Profile v1alpha1 Mapping

The profile carries a typed enterprise service as a v1alpha2 structured
`resource`. `kind` is fixed to `enterprise_service`. AEGIVELA does not own the
enterprise-architecture catalog; it consumes the trusted, product-owned typed
reference and evaluates it.

## Field Mapping

| Field | Value |
| --- | --- |
| `kind` | `enterprise_service` |
| `reference` | Product-owned catalog identifier |
| `resource_type` | `business_service`, `application_service`, or `technology_service` |
| `product_id` | Owning catalog/product |
| `descriptor_id` | Immutable descriptor identity |
| `descriptor_version` | Descriptor version; a new version is a new descriptor |
| `attributes` | Constraint inputs (below) |

## Attribute Vocabulary

| Attribute | Required | Values |
| --- | --- | --- |
| `data_classification` | yes | `public`, `internal`, `confidential`, `restricted` |
| `accountable_owner` | yes | non-empty reference |
| `segregation_of_duties` | no | `none`, `requester_must_differ_from_owner` |
| `approval_ref` | no | non-empty reference |
| `cost_center` | no | non-empty reference |

## Fail-Closed Rules

- A `kind: enterprise_service` resource without the complete structured field set
  is `invalid_policy_request` (`ErrInvalidRequest`).
- An unknown `resource_type`, an unknown attribute key, a missing required
  attribute, or an unknown enum value is `invalid_policy_request`.
- The canonical `structured_resource_digest` is computed by
  `policycontract.StructuredResourceDigestForResource`; the PEP recomputes it
  from its trusted descriptor instance and rejects a mismatch.
- Caller-supplied attributes or descriptors are comparison-only.
