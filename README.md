# gooo-improvement-selector

`gooo-improvement-selector` is a read-only, evidence-first loop for comparing
multiple Gooo self-improvement candidates. It declares candidates in `.gooo`,
compiles that source into semantic IR, generates declaration-only candidate
artifacts in caller-owned temporary output, records measurement receipts, and
selects `PROMOTE`, `DEFER`, `REFUTE`, or `INCOMPARABLE_UNKNOWN`.

The selector never applies a candidate, writes to an input checkout, creates a
commit, or pushes a change. Candidate artifacts carry the explicit operations
`DECLARE_ONLY`, `NO_INPUT_REPOSITORY_WRITE`, and `NO_APPLY`.

The evidence chain is:

`.gooo source` → `semantic IR` → `candidate artifact` → `measurement receipt`
→ `lexicographic selector` → `human report`

## Contract

The fixed denominator is exactly 12 cells. The selector order is explicit and
unweighted:

`AUTHORITY_GUARDRAIL` → `SEMANTIC_CONFORMANCE` →
`COUNTEREXAMPLE_PRESERVATION` → `EXACT_PAIRED_RESOURCE_DELTAS`

The four exact integer resource axes are `memory_kib`, `build_wall_ms`,
`test_wall_ms`, and `conformance_wall_ms`. The selector computes `after -
before` directly, with no score, percentage, weighted sum, or natural-language
inference. A missing before/after pair or a mismatched
`scenario_id/toolchain/input_digest/contract_digest` key is `UNKNOWN`.
Opposing resource directions are `INCOMPARABLE_UNKNOWN`. Lexicographic order
uses the four axes in the declared order. State precedence is
`REFUTED > UNKNOWN > CLOSED`.

Every `UNKNOWN` record contains the six operational fields `stage`, `step`,
`reason`, `unknown_class`, `next_operation`, and `blocked_by`.

Exact integer runtime observations include memory, build/test/conformance time,
reused, executed, skipped, generated artifacts, and the authority boundary:

```text
repository_writes=0
local_test_executions=0
cross_project_required_gates=0
```

Other repositories are optional inputs by immutable release identity and
digest only; this repository does not check them out.

## Scenario matrix

The conformance matrix defines exactly the requested minimum:

| class | scenarios | expected result |
| --- | --- | --- |
| normal (2) | `normal-promote`, `normal-defer` | `CLOSED`, `PROMOTE` or `DEFER` |
| unknown (3) | `unknown-missing-pair`, `unknown-digest-mismatch`, `unknown-cross-axis` | `UNKNOWN`, `INCOMPARABLE_UNKNOWN` |
| refuted (4) | `refuted-authority`, `refuted-semantic`, `refuted-counterexample`, `refuted-malformed` | `REFUTED`, `REFUTE` |

The fixture runner writes evidence under `$RUNNER_TEMP` and checks that the
input repository remains clean. The root README is the repository inventory;
generated artifacts and receipts are not committed.

## CI-only execution

The required verification runs in GitHub Actions with Go `1.27.x`. Per the
development boundary, local Go build/test/vet execution is not evidence for
this repository.

```sh
./scripts/ci-conformance.sh
```

The workflow builds and tests the selector in CI, runs all nine scenarios,
checks the 12/12 denominator, verifies the six-field UNKNOWN coordinates, and
emits a machine-readable report plus human-readable report and SHA-256-bound
artifacts in temporary output.

## Inventory

```text
cmd/gooo-selector/                    CLI entrypoint
selector/                             source compiler, pipeline, selector, report
contracts/improvement-selector-*.json fixed 12-cell denominator
examples/improvement-selector.gooo    declarative source
fixtures/scenarios/                   2 normal + 3 unknown + 4 refuted cases
scripts/ci-conformance.sh             CI fixture and report runner
docs/                                 contract and v0.1.0 release evidence
.gooo/                                ownership and development provenance
```

## Release

`v0.1.0` is immutable. Consumers verify the tag target and the SHA-256 digest
of the release evidence bundle before using it as an optional input.

## License

MIT.
