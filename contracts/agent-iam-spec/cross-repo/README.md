[English](README.md) · [简体中文](README.zh-CN.md)

# Cross-Repository Conformance Fixtures

Language-neutral, executable conformance fixtures for **AgentIAM** (the Agent IAM
Series) Parts 3–7. They exist so that an independent implementation can prove
interoperability against its published contract surface without reading another
implementation's source.

This program implements the requirement recorded in ADR-0016 of the AEGIVELA
reference implementation: the cross-repository fixture set is **owned by
`agent-iam-spec`** and **executed by each implementation**. A normative
specification clause, not a vendor contract, is the source of every expected
outcome.

## What a fixture is

A fixture is a JSON document conforming to [`fixture.schema.json`](fixture.schema.json):

| Field | Meaning |
|---|---|
| `id` | Stable fixture identifier. |
| `part` / `clause` | The AgentIAM part and clause the fixture exercises. |
| `threatRef` | Optional threat classification (`T1` PoP binding, `T2` workload attestation, `T4` lifecycle/revocation, `T7` exchange, `T8` audience binding, `I1` instance state, `F1` federation, `BR` broker issuance). |
| `kind` | `positive` (must accept) or `negative` (must reject). |
| `contract` | The contract surface and version the fixture targets, e.g. `{"surface": "enrollment", "version": "v3.0"}`. |
| `given` | Preconditions the harness must establish. |
| `operation.http` | The method, path, headers, and body to send. |
| `expected` | `outcome`, the HTTP `status`, an optional stable `reasonCode`, and human-readable `assertions`. |

`expected.reasonCode` is advisory: an implementation MAY use its own error
taxonomy as long as it rejects for the stated reason and the `assertions` hold.

## Running the fixtures

A conforming harness:

1. starts the implementation under test and provisions the `given` state;
2. performs `operation.http` against the implementation's contract surface;
3. compares the response to `expected` (status, outcome, and assertions);
4. reports pass/fail per `id`.

The fixtures are deliberately HTTP-shaped but transport-agnostic: a harness may
map `operation.http` onto any equivalent call that preserves the request
semantics and the contract version.

## Coverage

| Part | Fixtures |
|---|---|
| Part 3 — Identity and Authentication | enrollment, challenge single-use, token audience, epoch invalidation, PoP replay, credential generation, Registry Context fail-closed, workload proof profile version, JWKS rotation overlap |
| Part 4 — Authorization and Delegation | decision→grant, deny blocks grant, delegation non-amplification, approval binding, pre-dispatch revocation, credential-injection obligation, obligation enforcement |
| Part 5 — Cross-Domain Federation | active trust, principal isolation, brokered-field forgery, trust disable, brokered exchange success |
| Part 6 — Audit and Security Events | event persistence, redaction, transactional consistency |
| Part 7 — Conformance | claim completeness, missing part, profile composition |

See [`manifest.json`](manifest.json) for the machine-readable index.

## What a passing run proves, and does not prove

A passing run proves that the implementation accepts the positive fixtures and
rejects the negative fixtures at the stated clause, on the declared contract
version. It is **not** a certification, does not replace a deployment-specific
security review, and does not prove enterprise capabilities (HSM/KMS custody,
multi-region operation, administrative consoles) that are outside AgentIAM.

## Adding a fixture

Add a JSON document under the part directory, register it in `manifest.json`, and
run `node conformance/validate.mjs`. Every negative fixture MUST cite the clause
it protects; every security-critical clause SHOULD have at least one negative
fixture.
