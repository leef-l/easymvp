#!/usr/bin/env python3
import glob
import json
import os
import sys


def web_antd_artifacts() -> list[dict]:
    checks: list[dict] = []
    for path in sorted(glob.glob(".tmp/web-antd-guard/*.json")):
        with open(path, "r", encoding="utf-8") as fh:
            payload = json.load(fh)
        if isinstance(payload, dict):
            checks.append(payload)
    return checks


def backend_env() -> list[dict]:
    def map_result(value: str) -> str:
        text = (value or "").strip().lower()
        if text == "success":
            return "passed"
        if text in {"failure", "cancelled"}:
            return "failed"
        return "pending"

    checks: list[dict] = []
    for job, name, kind, result, summary, command in [
        (
            "validate-regression",
            "validate regression manifest and delivery policy",
            "runtime",
            os.environ.get("VALIDATE_RESULT"),
            "bash ./test-workspaces/validate.sh",
            "bash ./test-workspaces/validate.sh",
        ),
        (
            "test-backend",
            "backend go test ./...",
            "test",
            os.environ.get("TEST_BACKEND_RESULT"),
            "cd admin-go && go test ./...",
            "go test ./...",
        ),
        (
            "test-codegen",
            "codegen go test ./...",
            "test",
            os.environ.get("TEST_CODEGEN_RESULT"),
            "cd admin-go/codegen && go test ./...",
            "go test ./...",
        ),
        (
            "build-services",
            "build system/ai/mvp",
            "build",
            os.environ.get("BUILD_SERVICES_RESULT"),
            "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build system ai mvp",
            "go build system ai mvp",
        ),
    ]:
        normalized = (result or "").strip().lower()
        if not normalized or normalized == "skipped":
            continue
        checks.append(
            {
                "name": name,
                "kind": kind,
                "status": map_result(normalized),
                "summary": f"{summary}; job_result={normalized}",
                "command": command,
                "workflow": "backend-guard",
                "job": job,
            }
        )
    return checks


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: github-actions-collect-checks.py <web-antd-artifacts|backend-env>", file=sys.stderr)
        return 2

    mode = sys.argv[1]
    if mode == "web-antd-artifacts":
        checks = web_antd_artifacts()
    elif mode == "backend-env":
        checks = backend_env()
    else:
        print(f"unsupported mode: {mode}", file=sys.stderr)
        return 2

    json.dump(checks, sys.stdout, ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
