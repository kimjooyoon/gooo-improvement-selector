# Improvement selector contract v1

The contract has one immutable denominator and one explicit selector order.
Every denominator cell names a Gooo activity, a generated artifact, a metric
path, and its evaluator. The source and IR are digest-bound before candidate
generation.

## Candidate authority

The selector consumes observations. It does not write to a subject checkout,
invoke a refactor, or create a commit. A candidate artifact is a declaration
with `DECLARE_ONLY`, `NO_INPUT_REPOSITORY_WRITE`, and `NO_APPLY` operations.

## Exact paired resources

For each candidate, the measurement receipt carries integer `before` and
`after` values for:

- `memory_kib`
- `build_wall_ms`
- `test_wall_ms`
- `conformance_wall_ms`

The receipt key is the tuple `(scenario_id, toolchain, input_digest,
contract_digest)`. All four values must match the current run. The selector
computes `after - before` without normalization or weighting.

If a pair is absent or its key does not match, the candidate is UNKNOWN. If
two exact delta vectors move in opposite directions on different axes, the
comparison is `INCOMPARABLE_UNKNOWN`. If all gates and exact evidence are
closed, the ordered delta tuple is compared lexicographically; the first
different axis determines the result and there is no aggregate value.

## Scenario matrix

| class | count | expected state | expected decision |
|---|---:|---|---|
| normal | 2 | CLOSED | PROMOTE or DEFER |
| unknown | 3 | UNKNOWN | INCOMPARABLE_UNKNOWN |
| refuted | 4 | REFUTED | REFUTE |

`REFUTED > UNKNOWN > CLOSED` is applied whenever the scenario contains
multiple candidate outcomes. UNKNOWN coordinates are preserved in both the
candidate decision and the scenario selection.
