# OSS / Enterprise Edition Boundary

AEGIVELA ships in two editions: the **Open Core (OSS)** edition, which is the general-purpose security fabric for agentic systems, and the **Enterprise Edition (EE)**, which adds enterprise identity federation, continuous evaluation, operational controls, and SIEM integration.

## Boundary Principle

Every component available in the OSS edition must be deployable and useful without any EE dependency. EE features extend OSS, never replace or gate it. The OSS edition is production-grade for its scope; the EE edition integrates with enterprise infrastructure that is inherently organization-specific.

## OSS Edition — AEGIVELA Open Core

The OSS edition includes every profile-neutral security primitive and every AEGIVELA-owned integration contract. It is the foundation applicable to any agentic system regardless of enterprise IdP or SIEM vendor.

| Milestone | Rationale |
| --- | --- |
| F0 — Contract and Trust Foundation | Identity binding, policy evaluation, approval records, execution grants, revocation selectors, attestation policy, security evidence references — every deployment needs these |
| F1A — MODUREGIS Adapter Contract | Versioned adapter contract defines action/resource mappings, not a MODUREGIS implementation; MODUREGIS owns its own adapter runtime |
| F1B — Agent Gateway Identity | Agent identity binding, master-suspension cascade, five-mode principal resolution, workload assertions — universal agent security model |
| F1C — Web/API PEP | Web and API authorization middleware (`human_web`, `system_api`, `delegated_api`) — general-purpose PEP contract and Go reference SDK |
| F2A — Approval Contracts | Signed approval artifacts, scope attenuation, delegated-grant contracts, approval evidence fields — universal approval/privacy primitives |
| F2B — Tool Enforcement | `tool:invoke` attribution chain (`h -> a -> k -> s`), argument classification, skill hash, target host enforcement, credential injection contracts — universal agent tool security |
| F3A — Invocation Contracts | Execution-bound grant audience registry, revocation recheck contract (`pre_dispatch`/`continuation`/`connection` SLO classes), capability invoke mapping — universal execution security contracts |

## Enterprise Edition — AEGIVELA EE

The EE edition layers enterprise infrastructure integration on top of the OSS core. These features depend on organization-specific infrastructure (enterprise IdP, directory service, SIEM product, deployment fabric).

| Feature | Rationale |
| --- | --- |
| SAML Federation | Enterprise IdPs that speak SAML rather than OIDC. The OSS Identity Bridge validates OIDC; SAML assertions require IdP-specific protocol handling |
| Directory/SCIM Enrichment | Enterprise directory systems (LDAP, Azure AD, Okta) for attribute enrichment (groups, roles, department). Enrichment is additive to verified token claims, never a replacement |
| SPIFFE/SVID Workload Identity | X.509 SVID verification with trust bundles for enterprise workload identity fabrics (SPIRE, Istio). OSS supports JWKS workload assertions; EE adds X.509 |
| CAEP Risk Signals | Continuous Access Evaluation Protocol signal intake, risk-to-revocation escalation, and policy reevaluation on risk change. OSS has `RiskThreshold` in PDP rules; EE adds full CAEP event ingestion and workflow |
| SIEM Export | Standardized JSON evidence envelope in vendor-specific formats (CEF/LEEF), streaming export to SIEM products. OSS has structured redacted evidence; EE adds SIEM integration adapters |
| Policy Simulation | Dry-run policy evaluation with hypothetical requests for policy authoring, testing, and audit preview. OSS evaluates real requests; EE adds simulation mode |
| Key Rotation | Automated signing key rotation with overlapping `kid` windows, eliminating the OSS restart-based rotation in ADR-0001/0003 |
| Regional Deployment | Multi-region PostgreSQL replication, tenant affinity routing, and region-local JWKS/revocation caching for low-latency PDP checks across geographic deployments |
| Tenant Administration | Admin UI and API for tenant lifecycle (creation, suspension, deletion), per-tenant identity issuer configuration, and tenant-scoped audit views |

## Repository Structure

```
aegivela-open/         Public-facing: README, binary releases, Go SDK, examples, docs
aegivela/              Core implementation (private; open-source decision pending)
aegivela-ee/           Enterprise extensions (private; licensed)
```

- **`aegivela-open`** hosts the project homepage, installation guide, quickstart examples, and Go SDK reference. It links to the OSS binary releases and documentation. It does not contain source code.
- **`aegivela`** is the core Go implementation. All OSS edition features are implemented here.
- **`aegivela-ee`** contains EE-specific packages: SAML bridge, SCIM enrichment, SPIFFE SVID verifier, CAEP signal ingestion adapters, SIEM export formats, policy simulation, key rotation automation, and tenant administration. EE code imports OSS packages from `aegivela` and adds enterprise extension points.

## Horizon

The current `aegivela` repository is private. An open-source release of the OSS core is under evaluation. The OSS/EE boundary documented here will govern what ships in each distribution when the decision is made.

## Licensing

| Repository | License |
| --- | --- |
| `aegivela` (core implementation) | [GNU Affero General Public License v3.0](../LICENSE) |
| `aegivela-open` (public-facing) | [Apache License 2.0](https://github.com/axisrobo/aegivela-open/blob/main/LICENSE) |
| `aegivela-ee` (enterprise extensions) | Commercial / Enterprise License |
