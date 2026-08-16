# AEGIVELA Open

**AEGIVELA** is the AxisRobo Agent Identity, Authorization, and Security Fabric — a general-purpose security platform for agentic systems.

This is the public-facing repository for the AEGIVELA Open Core (OSS) edition. It contains the project overview, installation guide, Go SDK reference, quickstart examples, and links to binary releases. The core implementation lives in a private repository pending an open-source decision.

## What AEGIVELA Does

AEGIVELA provides profile-neutral security primitives for any agentic system:

- **Identity Binding** — Human, Agent, and Workload identity resolution from enterprise IdP tokens, workload assertions, and agent binding records
- **Policy Decision Point** — Versioned policy evaluation with fail-closed semantics, structured resource descriptors, and risk-level thresholds
- **Approvals** — Signed approval artifacts with scope/expiry/resource binding, non-amplifying delegation chains
- **Execution Grants** — Short-lived, audience-bound Ed25519 JWS grants with revocation selectors
- **Revocation** — Generic tenant-scoped revocation across grants, decisions, approvals, subjects, sessions, and policies
- **Attestation** — Artifact digest verification, signer identity, runtime/workload evidence
- **PEP SDK** — Go middleware for web/API authorization (`human_web`, `system_api`, `delegated_api`, twin agent, service agent modes)
- **Security Evidence** — Redacted, SIEM-compatible JSON evidence envelopes with cross-repository correlation

## OSS / Enterprise Boundary

The Open Core edition includes every profile-neutral security primitive and versioned integration contract. The Enterprise Edition (`aegivela-ee`) adds enterprise infrastructure integration: SAML federation, SCIM enrichment, SPIFFE/SVID, CAEP risk signals, SIEM export, policy simulation, key rotation, and tenant administration.

See [OSS/EE Boundary](https://github.com/axisrobo/aegivela-open/blob/main/docs/oss-ee-boundary.md) for the full split.

## Quick Start

### Binary Download

Download the latest `aegivela-api` binary for your platform from [GitHub Releases](https://github.com/axisrobo/aegivela-open/releases):

```bash
# Linux amd64 / arm64
curl -L -o aegivela-api https://github.com/axisrobo/aegivela-open/releases/latest/download/aegivela-api-linux-amd64
chmod +x aegivela-api
curl -L -o aegivela-api https://github.com/axisrobo/aegivela-open/releases/latest/download/aegivela-api-linux-arm64
chmod +x aegivela-api

# macOS amd64 / arm64
curl -L -o aegivela-api https://github.com/axisrobo/aegivela-open/releases/latest/download/aegivela-api-darwin-amd64
chmod +x aegivela-api
curl -L -o aegivela-api https://github.com/axisrobo/aegivela-open/releases/latest/download/aegivela-api-darwin-arm64
chmod +x aegivela-api

# Windows amd64 (PowerShell)
Invoke-WebRequest -Uri "https://github.com/axisrobo/aegivela-open/releases/latest/download/aegivela-api-windows-amd64.exe" -OutFile aegivela-api.exe
```

Verify integrity against the [`SHA256SUMS`](https://github.com/axisrobo/aegivela-open/releases/latest/download/SHA256SUMS) manifest.

### Docker Image

```bash
# Coming soon
docker pull ghcr.io/axisrobo/aegivela:latest
```

### Runtime Configuration

The API requires PostgreSQL and the following environment variables:

| Variable | Required | Purpose |
| --- | --- | --- |
| `DATABASE_URL` | yes | PostgreSQL connection string |
| `AEGIVELA_INTERNAL_AUTH_TOKEN` | yes | Internal service authentication |
| `AEGIVELA_IDENTITY_ISSUERS_FILE` | no | Path to OIDC issuer configuration JSON |
| `AEGIVELA_POLICY_DECISION_SIGNING_KEY_FILE` | see below | Ed25519 private key for policy decision signing |
| `AEGIVELA_POLICY_DECISION_SIGNING_KEY_ID` | see below | Key ID (kid) |
| `AEGIVELA_POLICY_DECISION_ISSUER` | see below | Issuer URL |
| `AEGIVELA_EXECUTION_GRANT_SIGNING_KEY_FILE` | see below | Ed25519 private key for execution grant signing |
| `AEGIVELA_EXECUTION_GRANT_SIGNING_KEY_ID` | see below | Key ID (kid) |
| `AEGIVELA_EXECUTION_GRANT_ISSUER` | see below | Issuer URL |
| `AEGIVELA_GRANT_AUDIENCE_REGISTRY_FILE` | see below | Grant audience allowlist JSON |
| `AEGIVELA_LISTEN_ADDR` | no | Listen address override (default `:1886`) |

Signing keys are required when decision/grant signing is enabled. Generate keys with:

```bash
openssl genpkey -algorithm ed25519 -out decision.key
openssl genpkey -algorithm ed25519 -out grant.key
```

For local development, see the [development guide](docs/api/identity-bridge.md).

## Integration Profiles

AEGIVELA is consumed through three integration profiles that share the same identity, decision, grant, revocation, and evidence contracts:

| Profile | Consumer | Actions |
| --- | --- | --- |
| **MODUREGIS Capability Governance** | MODUREGIS Console/Broker | `capability:read`, `capability:publish`, `adapter:activate`, `capability:invoke` |
| **Web/API Gateway** | Product APIs, API Gateways | `human_web`, `system_api`, `delegated_api` |
| **Agent Gateway Runtime** | AXIS-gateway Client/Server | `connection:open`, `tool:invoke`, `backend:request`, `credential:inject` |

## Go SDK

```go
import "github.com/axisrobo/aegivela-open/sdk/go/pepsdk"

client := pepsdk.NewClient("https://aegivela.internal:1886", http.DefaultClient)
result, err := client.Authorize(ctx, route, input)
```

Full SDK reference: [docs/sdk/](https://github.com/axisrobo/aegivela-open/blob/main/docs/sdk/)

## Documentation

- [Architecture](https://github.com/axisrobo/aegivela-open/blob/main/docs/architecture.md)
- [Roadmap](https://github.com/axisrobo/aegivela-open/blob/main/docs/roadmap.md)
- [Authorization Modes](https://github.com/axisrobo/aegivela-open/blob/main/docs/authorization-modes.md)
- [OSS/EE Boundary](https://github.com/axisrobo/aegivela-open/blob/main/docs/oss-ee-boundary.md)
- [ADR Index](https://github.com/axisrobo/aegivela-open/blob/main/docs/adr/)

## Repository Layout

```
aegivela-open/        ← You are here. Project homepage, docs, SDK reference, examples.
├── README.md
├── LICENSE            (Apache 2.0)
├── docs/              Architecture, roadmap, ADRs, API contracts
├── contracts/         Versioned JSON Schema / OpenAPI / protobuf contracts
├── sdk/go/
│   ├── go.mod         Go module (github.com/axisrobo/aegivela-open/sdk/go)
│   └── pepsdk/        PEP middleware and authorization client
└── examples/
    └── web-api-pep/   HTTP API protected by the PEP SDK
aegivela/              Core Go implementation (private, AGPL-3.0)
aegivela-ee/           Enterprise extensions (private, Enterprise License)
```

## License

The AEGIVELA Open Core (`aegivela-open`) is licensed under the [Apache License 2.0](LICENSE).

| Repository | License |
| --- | --- |
| `aegivela-open` (this repo) | Apache 2.0 |
| `aegivela` (core implementation) | GNU Affero GPL v3.0 |
| `aegivela-ee` (enterprise extensions) | Enterprise License |

## Development

The `aegivela-open` repository is synced from the private `aegivela` core repository via `aegivela/scripts/sync-oss.ps1`. To update the open repository after a core change:

```powershell
.\scripts\sync-oss.ps1 -SourceRoot ..\Aegivela -TargetRoot .
```

Test files are excluded from sync (they reference internal packages). The open SDK is tested via the examples.

Open documentation edits may be made directly in this repository. Contract and SDK files are authoritative in the core repository and should be synced rather than manually edited here.
