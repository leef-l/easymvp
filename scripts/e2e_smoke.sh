#!/usr/bin/env bash
# End-to-end smoke test: verifies Core ↔ Brain connectivity and basic API flows.
# Expects both services to be running:
#   Core on CORE_URL (default http://127.0.0.1:8000)
#   Brain on BRAIN_URL (default http://127.0.0.1:7701)

set -euo pipefail

CORE_URL="${CORE_URL:-http://127.0.0.1:8000}"
BRAIN_URL="${BRAIN_URL:-http://127.0.0.1:7701}"
TEST_PROJECT_ID=""
PASS=0
FAIL=0

pass() { echo "  PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "  FAIL: $1"; FAIL=$((FAIL + 1)); }

echo "== Phase 1: Service Health =="

if curl --fail --silent --max-time 5 "$CORE_URL/api/v3/system/healthz" >/dev/null 2>&1; then
  pass "Core healthz reachable"
else
  fail "Core healthz unreachable at $CORE_URL"
fi

if curl --fail --silent --max-time 5 "$BRAIN_URL/v1/health" >/dev/null 2>&1; then
  pass "Brain health reachable"
else
  fail "Brain health unreachable at $BRAIN_URL"
fi

echo ""
echo "== Phase 2: Create Test Project =="

CREATE_RESPONSE=$(curl --silent --max-time 10 \
  -X POST "$CORE_URL/api/v3/projects" \
  -H "Content-Type: application/json" \
  -d '{"name":"e2e-smoke-test","project_category":"web_app","goal_summary":"E2E smoke test project","workspace_root":"/tmp/e2e-smoke","repo_root":"/tmp/e2e-smoke"}' \
  2>/dev/null || echo "")

if [ -n "$CREATE_RESPONSE" ]; then
  TEST_PROJECT_ID=$(echo "$CREATE_RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',d).get('resource_id',''))" 2>/dev/null || echo "")
  if [ -n "$TEST_PROJECT_ID" ]; then
    pass "Project created: $TEST_PROJECT_ID"
  else
    fail "Project creation returned no resource_id"
  fi
else
  fail "Project creation request failed"
fi

echo ""
echo "== Phase 3: Workspace View =="

if [ -n "$TEST_PROJECT_ID" ]; then
  WS_RESPONSE=$(curl --silent --max-time 10 \
    "$CORE_URL/api/v3/projects/$TEST_PROJECT_ID/workspace-view" 2>/dev/null || echo "")
  if [ -n "$WS_RESPONSE" ]; then
    pass "Workspace view returned data"
  else
    fail "Workspace view returned empty"
  fi
else
  fail "Skipped workspace view (no project ID)"
fi

echo ""
echo "== Phase 4: Brain JSON-RPC Ping =="

RPC_RESPONSE=$(curl --silent --max-time 15 \
  -X POST "$BRAIN_URL/rpc" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"brain/execute","params":{"instruction":"ping","budget":{"max_turns":1}}}' \
  2>/dev/null || echo "")

if [ -n "$RPC_RESPONSE" ]; then
  HAS_JSONRPC=$(echo "$RPC_RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print('yes' if d.get('jsonrpc')=='2.0' else 'no')" 2>/dev/null || echo "no")
  if [ "$HAS_JSONRPC" = "yes" ]; then
    pass "Brain JSON-RPC responds with valid envelope"
  else
    fail "Brain JSON-RPC response missing jsonrpc field"
  fi
else
  fail "Brain JSON-RPC request failed"
fi

echo ""
echo "== Phase 5: Plan View =="

if [ -n "$TEST_PROJECT_ID" ]; then
  PLAN_RESPONSE=$(curl --silent --max-time 10 \
    "$CORE_URL/api/v3/projects/$TEST_PROJECT_ID/plan-view" 2>/dev/null || echo "")
  if [ -n "$PLAN_RESPONSE" ]; then
    pass "Plan view returned data"
  else
    fail "Plan view returned empty"
  fi
else
  fail "Skipped plan view (no project ID)"
fi

echo ""
echo "== Results =="
echo "  Passed: $PASS"
echo "  Failed: $FAIL"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
echo "  All smoke tests passed."
