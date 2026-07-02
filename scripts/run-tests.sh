#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPORTS_DIR="${REPO_ROOT}/test-reports"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"

# Defaults
TARGET="${1:-./internal/...}"
TIMEOUT="${TEST_TIMEOUT:-120m}"
PARALLEL="${TEST_PARALLEL:-4}"
REPORT_NAME=""

usage() {
    cat <<EOF
Usage: $(basename "$0") [TARGET] [OPTIONS]

Run acceptance tests with automatic report generation.

TARGET (positional, default: ./internal/...):
  ./internal/...                          All tests
  ./internal/resources/vpc/               Single resource
  ./internal/resources/instance/          Single resource
  ./internal/datasources/flavors/         Single data source

OPTIONS (via environment variables):
  TEST_TIMEOUT   Test timeout (default: 120m)
  TEST_PARALLEL  Parallelism (default: 4)
  TEST_RUN       Regex to filter specific test functions (e.g., TestAccVPC_basic)
  REPORT_NAME    Custom report name suffix (default: auto-generated from target)
  TF_ACC         Must be set to 1 (sourced from .env.test)

Examples:
  ./scripts/run-tests.sh                                          # All tests
  ./scripts/run-tests.sh ./internal/resources/vpc/                # VPC only
  TEST_RUN=TestAccVPC_basic ./scripts/run-tests.sh                # Single test
  TEST_PARALLEL=1 ./scripts/run-tests.sh ./internal/resources/instance/

EOF
    exit 0
}

[[ "${1:-}" == "-h" || "${1:-}" == "--help" ]] && usage

# Derive a short name for the report from the target path
derive_report_name() {
    local target="$1"
    if [[ "$target" == "./internal/..." || "$target" == "./..." ]]; then
        echo "all"
    else
        echo "$target" | sed 's|^\./internal/||' | sed 's|resources/||' | sed 's|datasources/|ds-|' | sed 's|/$||' | tr '/' '-'
    fi
}

REPORT_NAME="${REPORT_NAME:-$(derive_report_name "$TARGET")}"
REPORT_FILE="${REPORTS_DIR}/${TIMESTAMP}-${REPORT_NAME}"

mkdir -p "$REPORTS_DIR"

# Verify TF_ACC is set
if [[ "${TF_ACC:-}" != "1" ]]; then
    echo "ERROR: TF_ACC is not set to 1. Source .env.test first:"
    echo "  source .env.test"
    exit 1
fi

# Build go test args
GO_TEST_ARGS=(-v -timeout "$TIMEOUT" -parallel "$PARALLEL" -count=1)

if [[ -n "${TEST_RUN:-}" ]]; then
    GO_TEST_ARGS+=(-run "$TEST_RUN")
fi

echo "═══════════════════════════════════════════════════════════════"
echo "  Acceptance Test Run"
echo "═══════════════════════════════════════════════════════════════"
echo "  Target:      $TARGET"
echo "  Timeout:     $TIMEOUT"
echo "  Parallel:    $PARALLEL"
echo "  Run filter:  ${TEST_RUN:-<none>}"
echo "  Report:      ${REPORT_FILE}.txt"
echo "  Started:     $(date)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

START_TIME=$(date +%s)

# Run tests with JSON output for structured parsing, tee to both console and file
set +e
go test "$TARGET" "${GO_TEST_ARGS[@]}" -json 2>&1 | tee "${REPORT_FILE}.jsonl"
EXIT_CODE=${PIPESTATUS[0]}
set -e

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Generate human-readable report from JSON lines
echo ""
echo "Generating report..."

# Parse results
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0
FAILED_TESTS=""

while IFS= read -r line; do
    action=$(echo "$line" | jq -r '.Action // empty' 2>/dev/null)
    test_name=$(echo "$line" | jq -r '.Test // empty' 2>/dev/null)
    pkg=$(echo "$line" | jq -r '.Package // empty' 2>/dev/null)

    [[ -z "$action" || -z "$test_name" ]] && continue

    case "$action" in
        pass)
            TOTAL=$((TOTAL + 1))
            PASSED=$((PASSED + 1))
            ;;
        fail)
            TOTAL=$((TOTAL + 1))
            FAILED=$((FAILED + 1))
            FAILED_TESTS="${FAILED_TESTS}\n  - ${pkg}::${test_name}"
            ;;
        skip)
            TOTAL=$((TOTAL + 1))
            SKIPPED=$((SKIPPED + 1))
            ;;
    esac
done < "${REPORT_FILE}.jsonl"

# Write summary report
cat > "${REPORT_FILE}.md" <<EOF
# Test Report: ${REPORT_NAME}

**Date:** $(date -u +"%Y-%m-%d %H:%M:%S UTC")
**Target:** \`${TARGET}\`
**Duration:** ${DURATION}s
**Filter:** ${TEST_RUN:-none}
**Parallel:** ${PARALLEL}

## Summary

| Metric | Count |
|--------|-------|
| Total  | ${TOTAL} |
| Passed | ${PASSED} |
| Failed | ${FAILED} |
| Skipped| ${SKIPPED} |

**Result: $([ $EXIT_CODE -eq 0 ] && echo "PASS ✅" || echo "FAIL ❌")**

EOF

if [[ -n "$FAILED_TESTS" ]]; then
    cat >> "${REPORT_FILE}.md" <<EOF
## Failed Tests
$(echo -e "$FAILED_TESTS")

## Failure Details

\`\`\`
$(grep -A 5 '"Action":"fail"' "${REPORT_FILE}.jsonl" | jq -r '.Output // empty' 2>/dev/null | head -100)
\`\`\`

EOF
fi

cat >> "${REPORT_FILE}.md" <<EOF
## Environment

| Variable | Set |
|----------|-----|
| ACECLOUD_API_URL | $([ -n "${ACECLOUD_API_URL:-}" ] && echo "✅" || echo "❌") |
| ACECLOUD_API_KEY_ID | $([ -n "${ACECLOUD_API_KEY_ID:-}" ] && echo "✅" || echo "❌") |
| ACECLOUD_REGION | $([ -n "${ACECLOUD_REGION:-}" ] && echo "✅" || echo "❌") |
| ACECLOUD_PROJECT_ID | $([ -n "${ACECLOUD_PROJECT_ID:-}" ] && echo "✅" || echo "❌") |
| ACECLOUD_FLAVOR_ID | $([ -n "${ACECLOUD_FLAVOR_ID:-}" ] && echo "✅" || echo "❌") |
| ACECLOUD_IMAGE_ID | $([ -n "${ACECLOUD_IMAGE_ID:-}" ] && echo "✅" || echo "❌") |
| ACECLOUD_EXTERNAL_NETWORK_ID | $([ -n "${ACECLOUD_EXTERNAL_NETWORK_ID:-}" ] && echo "✅" || echo "❌") |

---
*Generated by \`scripts/run-tests.sh\`*
EOF

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  Test Run Complete"
echo "═══════════════════════════════════════════════════════════════"
echo "  Result:    $([ $EXIT_CODE -eq 0 ] && echo "PASS ✅" || echo "FAIL ❌")"
echo "  Total:     $TOTAL tests"
echo "  Passed:    $PASSED"
echo "  Failed:    $FAILED"
echo "  Skipped:   $SKIPPED"
echo "  Duration:  ${DURATION}s"
echo "  Report:    ${REPORT_FILE}.md"
echo "  Raw logs:  ${REPORT_FILE}.jsonl"
echo "═══════════════════════════════════════════════════════════════"

exit $EXIT_CODE
