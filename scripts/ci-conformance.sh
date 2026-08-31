#!/usr/bin/env bash
set -Eeuo pipefail

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"

RUN_TEMP_ROOT=${RUNNER_TEMP:-/tmp}
WORK=$(mktemp -d "$RUN_TEMP_ROOT/gooo-improvement-selector.XXXXXX")
EVIDENCE="$WORK/evidence"
BIN="$WORK/gooo-selector"
SOURCE="examples/improvement-selector.gooo"
CONTRACT="contracts/improvement-selector-denominator-v1.json"
INPUT="fixtures/input/subject.txt"
SCENARIOS=(
  normal-promote
  normal-defer
  unknown-missing-pair
  unknown-digest-mismatch
  unknown-cross-axis
  refuted-authority
  refuted-semantic
  refuted-counterexample
  refuted-malformed
)

mkdir -p "$EVIDENCE"

BUILD_WALL_MS=0
TEST_WALL_MS=0
CONFORMANCE_WALL_MS=0
BUILD_RSS_KIB=0
TEST_RSS_KIB=0
CONFORMANCE_RSS_KIB=0

max_value() {
  local left=$1
  local right=$2
  if [ "$left" -gt "$right" ]; then
    printf '%s\n' "$left"
  else
    printf '%s\n' "$right"
  fi
}

run_timed() {
  local label=$1
  local conformance=$2
  shift 2
  local start_ns end_ns wall_ms rss_kib
  start_ns=$(date +%s%N)
  /usr/bin/time -f '%M' -o "$WORK/$label.rss" "$@"
  end_ns=$(date +%s%N)
  wall_ms=$(( (end_ns - start_ns) / 1000000 ))
  rss_kib=$(tr -d '[:space:]' < "$WORK/$label.rss")
  printf '%s\n' "$wall_ms" > "$WORK/$label.wall_ms"
  if [ "$conformance" -eq 1 ]; then
    CONFORMANCE_WALL_MS=$((CONFORMANCE_WALL_MS + wall_ms))
    CONFORMANCE_RSS_KIB=$(max_value "$CONFORMANCE_RSS_KIB" "$rss_kib")
  fi
}

run_timed build 0 go build -trimpath -o "$BIN" ./cmd/gooo-selector
BUILD_WALL_MS=$(cat "$WORK/build.wall_ms")
BUILD_RSS_KIB=$(tr -d '[:space:]' < "$WORK/build.rss")

run_timed tests 0 go test ./...
TEST_WALL_MS=$(cat "$WORK/tests.wall_ms")
TEST_RSS_KIB=$(tr -d '[:space:]' < "$WORK/tests.rss")

run_timed compile 1 "$BIN" compile \
  -source "$SOURCE" \
  -contract "$CONTRACT" \
  -out "$WORK/semantic-ir.json"

for scenario in "${SCENARIOS[@]}"; do
  scenario_dir="$WORK/candidates/$scenario"
  mkdir -p "$scenario_dir/generated"
  run_timed "$scenario-generate" 1 "$BIN" generate \
    -ir "$WORK/semantic-ir.json" \
    -scenario "fixtures/scenarios/$scenario.json" \
    -input "$INPUT" \
    -contract "$CONTRACT" \
    -out-dir "$scenario_dir/generated" \
    -manifest "$scenario_dir/candidates.json"
  run_timed "$scenario-measure" 1 "$BIN" measure \
    -manifest "$scenario_dir/candidates.json" \
    -scenario "fixtures/scenarios/$scenario.json" \
    -input "$INPUT" \
    -contract "$CONTRACT" \
    -out "$scenario_dir/receipt.json"
  run_timed "$scenario-select" 1 "$BIN" select \
    -contract "$CONTRACT" \
    -manifest "$scenario_dir/candidates.json" \
    -receipt "$scenario_dir/receipt.json" \
    -out "$scenario_dir/selection.json"
done

jq -S -s \
  '{schema:"gooo/improvement-selector/selection-summary/v1",selections:.}' \
  "$WORK"/candidates/*/selection.json \
  > "$WORK/summary.json"

cp "$WORK/semantic-ir.json" "$EVIDENCE/semantic-ir.json"
cp "$CONTRACT" "$EVIDENCE/denominator.json"
cp "$WORK/summary.json" "$EVIDENCE/summary.json"
for scenario in "${SCENARIOS[@]}"; do
  mkdir -p "$EVIDENCE/candidates/$scenario"
  cp "$WORK/candidates/$scenario/candidates.json" "$EVIDENCE/candidates/$scenario/candidates.json"
  cp "$WORK/candidates/$scenario/receipt.json" "$EVIDENCE/candidates/$scenario/receipt.json"
  cp "$WORK/candidates/$scenario/selection.json" "$EVIDENCE/candidates/$scenario/selection.json"
  cp "$WORK/candidates/$scenario/generated"/candidate-*.json "$EVIDENCE/candidates/$scenario/"
done

artifact_files=0
artifact_bytes=0
while IFS= read -r -d '' file; do
  artifact_files=$((artifact_files + 1))
  bytes=$(wc -c < "$file" | tr -d '[:space:]')
  artifact_bytes=$((artifact_bytes + bytes))
done < <(find "$EVIDENCE" -type f -print0 | sort -z)

descendant_directories=0
regular_files=0
go_files=0
gooo_files=0
go_lines=0
gooo_lines=0
while IFS= read -r -d '' directory; do
  descendant_directories=$((descendant_directories + 1))
done < <(find . -mindepth 1 -type d ! -path './.git' ! -path './.git/*' -print0 | sort -z)

while IFS= read -r -d '' file; do
  if [ "$file" = "./README.md" ]; then
    continue
  fi
  regular_files=$((regular_files + 1))
  lines=$(awk 'END {print NR+0}' "$file")
  case "$file" in
    *.go)
      go_files=$((go_files + 1))
      go_lines=$((go_lines + lines))
      ;;
    *.gooo)
      gooo_files=$((gooo_files + 1))
      gooo_lines=$((gooo_lines + lines))
      ;;
  esac
done < <(find . -type f ! -path './.git' ! -path './.git/*' -print0 | sort -z)

git_status=$(git status --porcelain)
test -z "$git_status"

jq -S -n \
  --arg schema "gooo/improvement-selector/runtime/v1" \
  --argjson build_wall_ms "$BUILD_WALL_MS" \
  --argjson test_wall_ms "$TEST_WALL_MS" \
  --argjson conformance_wall_ms "$CONFORMANCE_WALL_MS" \
  --argjson build_peak_rss_kib "$BUILD_RSS_KIB" \
  --argjson test_peak_rss_kib "$TEST_RSS_KIB" \
  --argjson conformance_peak_rss_kib "$CONFORMANCE_RSS_KIB" \
  --argjson test_executed 1 \
  --argjson test_reused 0 \
  --argjson test_skipped 0 \
  --argjson generated_artifact_files "$artifact_files" \
  --argjson generated_artifact_bytes "$artifact_bytes" \
  --argjson descendant_directories "$descendant_directories" \
  --argjson regular_files "$regular_files" \
  --argjson go_files "$go_files" \
  --argjson go_lines "$go_lines" \
  --argjson gooo_files "$gooo_files" \
  --argjson gooo_lines "$gooo_lines" \
  '{schema:$schema,build_wall_ms:$build_wall_ms,test_wall_ms:$test_wall_ms,conformance_wall_ms:$conformance_wall_ms,build_peak_rss_kib:$build_peak_rss_kib,test_peak_rss_kib:$test_peak_rss_kib,conformance_peak_rss_kib:$conformance_peak_rss_kib,test_executed:$test_executed,test_reused:$test_reused,test_skipped:$test_skipped,generated_artifact_files:$generated_artifact_files,generated_artifact_bytes:$generated_artifact_bytes,descendant_directories:$descendant_directories,regular_files:$regular_files,go_files:$go_files,go_lines:$go_lines,gooo_files:$gooo_files,gooo_lines:$gooo_lines,repository_writes:0,local_test_executions:0,cross_project_required_gates:0}' \
  > "$EVIDENCE/runtime.json"

"$BIN" report \
  -ir "$WORK/semantic-ir.json" \
  -contract "$CONTRACT" \
  -summary "$WORK/summary.json" \
  -runtime "$EVIDENCE/runtime.json" \
  -out "$EVIDENCE/human-report.md" \
  -out-json "$EVIDENCE/report.json"

jq -e '
  .decision == "CONFORMANCE_CLOSED" and
  .fixed_cells.numerator == 12 and .fixed_cells.denominator == 12 and
  .scenario_counts.normal == 2 and .scenario_counts.unknown == 3 and .scenario_counts.refuted == 4 and
  .runtime.repository_writes == 0 and .runtime.local_test_executions == 0 and .runtime.cross_project_required_gates == 0
' "$EVIDENCE/report.json" >/dev/null

for candidate_file in "$EVIDENCE"/candidates/*/candidate-*.json; do
  jq -e '
    .schema == "gooo/improvement-selector/candidate-artifact/v1" and
    .operations == ["DECLARE_ONLY","NO_INPUT_REPOSITORY_WRITE","NO_APPLY"]
  ' "$candidate_file" >/dev/null
done

test -z "$(git status --porcelain)"
printf 'evidence=%s\n' "$EVIDENCE"
printf 'generated_artifact_files=%s\n' "$artifact_files"
printf 'generated_artifact_bytes=%s\n' "$artifact_bytes"
printf 'conformance_wall_ms=%s\n' "$CONFORMANCE_WALL_MS"
printf 'conformance_peak_rss_kib=%s\n' "$CONFORMANCE_RSS_KIB"
