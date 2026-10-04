# Enterprise Service Authorization Profile v2.0 Mapping

The profile carries a typed enterprise service as a v1alpha2 structured
`resource`. `kind` is fixed to `enterpriseService`. AEGIVELA does not own the
enterprise-architecture catalog; it consumes the trusted, product-owned typed
reference and evaluates it.

## Field Mapping

| Field | Value |
| --- | --- |
| `kind` | `enterpriseService` |
| `reference` | Product-owned catalog identifier |
| `resourceType` | `businessService`, `applicationService`, or `technologyService` |
| `productId` | Owning catalog/product |
| `descriptorId` | Immutable descriptor identity |
| `descriptorVersion` | Descriptor version; a new version is a new descriptor |
| `attributes` | Constraint inputs (below) |

## Attribute Vocabulary

| Attribute | Required | Values |
| --- | --- | --- |
| `dataClassification` | yes | `public`, `internal`, `confidential`, `restricted` |
| `accountableOwner` | yes | non-empty reference |
| `segregationOfDuties` | no | `none`, `requesterMustDifferFromOwner` |
| `approvalRef` | no | non-empty reference |
| `costCenter` | no | non-empty reference |

## Fail-Closed Rules

- A `kind: enterpriseService` resource without the complete structured field set
  is `invalidPolicyRequest` (`ErrInvalidRequest`).
- An unknown `resourceType`, an unknown attribute key, a missing required
  attribute, or an unknown enum value is `invalidPolicyRequest`.
- The canonical `structuredResourceDigest` is computed by
  `policycontract.StructuredResourceDigestForResource`; the PEP recomputes it
  from its trusted descriptor instance and rejects a mismatch.
- Caller-supplied attributes or descriptors are comparison-only.
