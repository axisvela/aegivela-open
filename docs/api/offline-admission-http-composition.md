# Offline-Admission HTTP Composition

**Status:** transport and config assembly implemented; mounted only when the
explicit flag is true and a complete config file is supplied. Default off; no
deployment or offline enablement is authorized.

## Routes

| Method | Path | Authentication |
| --- | --- | --- |
| `POST` | `/v1/offline-admissions/issue` | `X-Aegivela-Internal-Token` |
| `POST` | `/v1/offline-admissions/verify` | `X-AEGIVELA-PEP` |
| `GET` | `/v1/offline-admissions/jwks` | Public |

The issue and verify bodies are bounded to 64 KiB. The service layer returns the
canonical response bytes; the transport does not reinterpret artifact claims.

## Enable Flag

`AEGIVELA_OFFLINE_ADMISSION_ENABLED` is disabled by default. Only `1`, `true` or
`yes` (case/space tolerant) enable it; empty, `0`, `false`, `off` and `no` keep
the routes absent. When true, `AEGIVELA_OFFLINE_ASSEMBLY_FILE` must point to a
complete assembly config; missing or invalid config aborts startup. The flag does
not enable offline admission by itself.

## Composition Seam

`httpapi.RegisterOfflineAdmissionRoutes` requires:

- an `InternalTokenResolver` for both credentials;
- an issue service (`Issue(ctx, []byte) ([]byte, error)`);
- a verify service (`Verify(ctx, []byte) ([]byte, error)`);
- a public JWKS provider.

Incomplete configuration returns an error. Service failures map to the stable
error envelope; unknown failures return `503`. Authentication failures return
`401`; oversized or empty bodies return `400`.

## Assembly Config

When enabled, `AEGIVELA_OFFLINE_ASSEMBLY_FILE` supplies:

```json
{
  "signing_key_file": "/path/offline-key.pem",
  "signing_key_id": "offline-key-1",
  "issuer": "https://aegivela.example",
  "trust_roots_file": "/path/trust-roots.json",
  "trust_roots_fingerprint": "<sha256>",
  "trust_bundle_file": "/path/trust-bundle.jws"
}
```

Startup loads the Ed25519 signing key, pinned trust roots and signed trust
bundle, then mounts the three routes with the existing internal-token resolver
(issue) and PEP credential (verify). The parent grant is verified with the
declared `audience`, `requestedScope` and optional `taskBinding`; the budget
source must already exist for the verified tenant/parent. Issuance reserves in
PostgreSQL and signs only the committed allocation payload.

Default-off behavior and complete-config failure are tested; the assembly itself
is not deployed or enabled in this repository.

## Tests

The transport tests cover: routes absent until registered, issue/verify
authentication separation, public JWKS, error-envelope mapping, body limits and
incomplete configuration. They do not prove issuance or admission behavior.
