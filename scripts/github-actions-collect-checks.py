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
            "GitHub Actions validate regression manifest and delivery policy",
            "runtime",
            os.environ.get("VALIDATE_RESULT"),
            "GitHub Actions backend-guard job validate-regression executed bash ./test-workspaces/validate.sh",
            "GitHub Actions workflow backend-guard / job validate-regression",
        ),
        (
            "test-backend",
            "GitHub Actions backend test",
            "test",
            os.environ.get("TEST_BACKEND_RESULT"),
            "GitHub Actions backend-guard job test-backend executed cd admin-go && go test ./...",
            "GitHub Actions workflow backend-guard / job test-backend",
        ),
        (
            "test-codegen",
            "GitHub Actions codegen test",
            "test",
            os.environ.get("TEST_CODEGEN_RESULT"),
            "GitHub Actions backend-guard job test-codegen executed cd admin-go/codegen && go test ./...",
            "GitHub Actions workflow backend-guard / job test-codegen",
        ),
        (
            "build-services",
            "GitHub Actions build system/ai/mvp",
            "build",
            os.environ.get("BUILD_SERVICES_RESULT"),
            "GitHub Actions backend-guard job build-services executed CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build system ai mvp",
            "GitHub Actions workflow backend-guard / job build-services",
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
