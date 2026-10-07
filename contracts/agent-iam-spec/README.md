# AgentIAM Cross-Repository Conformance Fixtures (vendored)

`cross-repo/` is a vendored copy of `agent-iam-spec`'s
`conformance/cross-repo/` directory.

- Series version: `0.4.1-draft` (manifest `generated` 2026-10-04)
- Source: <https://github.com/AICTRI/agent-iam-spec> (`conformance/cross-repo/`)

The fixtures are **owned by the AgentIAM series**; AEGIVELA only executes the
subset whose `contract.surface` it owns. Refresh by re-copying the directory from
the source at a pinned series version and updating this header plus the pinned
version noted in `backend/internal/conformance`.

## Ownership

| `contract.surface` | Owner | Parts |
| --- | --- | --- |
| `aegivela.authorization-plane` | AEGIVELA | 4, 6 |
| `agent-iam-series` | AEGIVELA (Part 7 claim) | 7 |
| `eidovela.registry-consumer` | EIDOVELA | 3, 5 |

## Execution

`go test ./backend/internal/conformance` validates every fixture against
[`fixture.schema.json`](cross-repo/fixture.schema.json), enforces that every
AEGIVELA fixture has a harness mapping, and executes the fixtures the harness
supports. The logical `operation.http` paths in the fixtures are
transport-agnostic; the harness maps each to the equivalent AEGIVELA endpoint
(see `docs/conformance-status.md`).
