package service

import "testing"

func TestCollectRunArtifactCandidateRootsSkipsEmptyExecutionID(t *testing.T) {
	t.Parallel()

	roots := runArtifactRoots{
		RunsRoot:   "/tmp/easymvp/runs",
		ReplayRoot: "/tmp/easymvp/replay",
	}

	got := collectRunArtifactCandidateRoots(roots, "run_123", "")
	if len(got) != 2 {
		t.Fatalf("unexpected candidate count: got %d want %d", len(got), 2)
	}
	if got[0] != "/tmp/easymvp/runs/run_123" {
		t.Fatalf("unexpected runs root candidate: got %s", got[0])
	}
	if got[1] != "/tmp/easymvp/replay/run_123" {
		t.Fatalf("unexpected replay root candidate: got %s", got[1])
	}
}

func TestClassifyReplayKindUsesFilenameHints(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		item runArtifactFile
		want string
	}{
		{name: "tool call", item: runArtifactFile{RelPath: "replay/20260419_tool-call_0002.json", FileName: "20260419_tool-call_0002.json"}, want: "tool_call"},
		{name: "tool result", item: runArtifactFile{RelPath: "replay/tool_result_0003.json", FileName: "tool_result_0003.json"}, want: "tool_result"},
		{name: "browser trace", item: runArtifactFile{RelPath: "artifacts/browser_0004.json", FileName: "browser_0004.json"}, want: "browser_trace"},
		{name: "default", item: runArtifactFile{RelPath: "replay/step_0001.json", FileName: "step_0001.json"}, want: "step_snapshot"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyReplayKind(tc.item); got != tc.want {
				t.Fatalf("unexpected replay kind: got %s want %s", got, tc.want)
			}
		})
	}
}

func TestClassifyLogStreamKindUsesFilenameHints(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		item runArtifactFile
		want string
	}{
		{name: "stderr", item: runArtifactFile{RelPath: "logs/20260419_stderr_0001.log", FileName: "20260419_stderr_0001.log"}, want: "stderr"},
		{name: "system", item: runArtifactFile{RelPath: "logs/system_0002.log", FileName: "system_0002.log"}, want: "system"},
		{name: "tool", item: runArtifactFile{RelPath: "logs/tool_0003.log", FileName: "tool_0003.log"}, want: "tool"},
		{name: "stdout default", item: runArtifactFile{RelPath: "logs/20260419_stdout_0004.log", FileName: "20260419_stdout_0004.log"}, want: "stdout"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyLogStreamKind(tc.item); got != tc.want {
				t.Fatalf("unexpected stream kind: got %s want %s", got, tc.want)
			}
		})
	}
}

func TestClassifyRunArtifactPathIgnoresCheckpointAndMetaFiles(t *testing.T) {
	t.Parallel()

	root := "/tmp/easymvp/runs/run_123"
	cases := []struct {
		name string
		path string
		want string
	}{
		{name: "checkpoint json", path: "/tmp/easymvp/runs/run_123/checkpoints/20260419_checkpoint_run_123.json", want: ""},
		{name: "meta json", path: "/tmp/easymvp/runs/run_123/meta.json", want: ""},
		{name: "replay json", path: "/tmp/easymvp/runs/run_123/replay/20260419_replay_run_123_step_0001.json", want: "replay"},
		{name: "stdout log", path: "/tmp/easymvp/runs/run_123/logs/20260419_stdout_0001.log", want: "log"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyRunArtifactPath(root, tc.path); got != tc.want {
				t.Fatalf("unexpected artifact kind: got %q want %q", got, tc.want)
			}
		})
	}
}
