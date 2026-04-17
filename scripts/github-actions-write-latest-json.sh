#!/usr/bin/env bash
set -euo pipefail

OUTPUT_PATH="${EASYMVP_CI_OUTPUT:-${1:-.easymvp/ci/latest.json}}"

mkdir -p "$(dirname "$OUTPUT_PATH")"

python3 - "$OUTPUT_PATH" <<'PY'
import json
import os
import sys


def normalize_status(value: str) -> str:
    text = (value or "").strip().lower()
    if text in {"success", "succeeded", "passed", "pass", "completed"}:
        return "passed"
    if text in {"failure", "failed", "error", "cancelled", "canceled", "timed_out"}:
        return "failed"
    if text in {"", "queued", "queue", "running", "in_progress", "pending", "requested", "waiting"}:
        return "pending"
    return text


def normalize_kind(value: str) -> str:
    text = (value or "").strip().lower()
    if text in {"lint", "test", "build", "runtime"}:
        return text
    if text in {"browser", "e2e"}:
        return "browser"
    return ""


def compact(item: dict) -> dict:
    return {key: value for key, value in item.items() if value not in ("", None, [], {})}


output_path = sys.argv[1]
workflow = os.environ.get("EASYMVP_CI_WORKFLOW", "").strip()
pipeline = os.environ.get("EASYMVP_CI_PIPELINE", "").strip() or workflow
summary = os.environ.get("EASYMVP_CI_SUMMARY", "").strip()
run_id = os.environ.get("EASYMVP_CI_RUN_ID", "").strip()
run_url = os.environ.get("EASYMVP_CI_RUN_URL", "").strip()
status_raw = os.environ.get("EASYMVP_CI_STATUS", "").strip()
status = normalize_status(status_raw) if status_raw else ""

checks_raw = os.environ.get("EASYMVP_CI_CHECKS_JSON", "").strip() or "[]"
try:
    parsed_checks = json.loads(checks_raw)
except json.JSONDecodeError as exc:
    raise SystemExit(f"EASYMVP_CI_CHECKS_JSON is not valid JSON: {exc}") from exc

if not isinstance(parsed_checks, list):
    raise SystemExit("EASYMVP_CI_CHECKS_JSON must be a JSON array")

checks = []
for raw in parsed_checks:
    if not isinstance(raw, dict):
        continue
    name = str(raw.get("name", "")).strip()
    kind = normalize_kind(str(raw.get("kind", "")))
    check_status = normalize_status(str(raw.get("status", "")))
    check_summary = str(raw.get("summary", "")).strip()
    command = str(raw.get("command", "")).strip()
    check_workflow = str(raw.get("workflow", "")).strip() or workflow
    job = str(raw.get("job", "")).strip()
    runner = "github_actions"

    if not name:
        if kind:
            name = f"github actions {kind}"
        elif job:
            name = job
        else:
            name = "github actions check"

    checks.append(
        compact(
            {
                "name": name,
                "kind": kind,
                "status": check_status,
                "summary": check_summary,
                "command": command,
                "runner": runner,
                "workflow": check_workflow,
                "job": job,
            }
        )
    )

check_kinds = sorted({item["kind"] for item in checks if item.get("kind")})

if not status:
    if not checks:
        status = "pending"
    else:
        statuses = [item.get("status", "pending") for item in checks]
        if any(item == "failed" for item in statuses):
            status = "failed"
        elif any(item == "pending" for item in statuses):
            status = "pending"
        else:
            status = "passed"

if not summary:
    passed = sum(1 for item in checks if item.get("status") == "passed")
    failed = sum(1 for item in checks if item.get("status") == "failed")
    pending = sum(1 for item in checks if item.get("status") == "pending")
    summary = f"checks={len(checks)} passed={passed} failed={failed} pending={pending}"

payload = compact(
    {
        "status": status,
        "tool": "github_actions",
        "pipeline": pipeline,
        "summary": summary,
        "workflow": workflow,
        "runId": run_id,
        "runUrl": run_url,
        "checkKinds": check_kinds,
        "checks": checks,
    }
)

with open(output_path, "w", encoding="utf-8") as fh:
    json.dump(payload, fh, ensure_ascii=False, indent=2)
    fh.write("\n")
PY

echo "Wrote GitHub Actions CI latest snapshot to $OUTPUT_PATH"
