package engine

import "testing"

func TestArchitectReplyRequestsContinuation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "explicit token", content: "这是第一段\n[AUTO_CONTINUE_NEXT]", want: true},
		{name: "plain continue text is not enough", content: "这是第 1/4 段，如需继续请回复继续。", want: false},
		{name: "final patch payload", content: "```json\n{\"task_patches\":[{\"task_name\":\"登录\",\"description\":\"补齐 OAuth 回调说明\"}]}\n```", want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := architectReplyRequestsContinuation(tc.content); got != tc.want {
				t.Fatalf("architectReplyRequestsContinuation() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExtractArchitectTaskPatchesMergesMultipleCodeBlocks(t *testing.T) {
	t.Parallel()

	content := "第一段\n```json\n{\"task_patches\":[{\"task_name\":\"登录\",\"description\":\"更新登录描述\",\"affected_resources\":[\"frontend/login.tsx\"]}]}\n```\n\n第二段\n```json\n{\"task_patches\":[{\"task_name\":\"资料页\",\"description\":\"补齐保存逻辑\",\"depends_on\":[\"登录\"]}]}\n```"
	patches, err := extractArchitectTaskPatches(content)
	if err != nil {
		t.Fatalf("extractArchitectTaskPatches() error = %v", err)
	}
	if len(patches) != 2 {
		t.Fatalf("expected 2 patches, got %d", len(patches))
	}
	if patches[0].TaskName != "登录" || patches[1].TaskName != "资料页" {
		t.Fatalf("unexpected patches: %+v", patches)
	}
	if len(patches[1].DependsOn) != 1 || patches[1].DependsOn[0] != "登录" {
		t.Fatalf("unexpected depends_on: %+v", patches[1])
	}
}

func TestExtractArchitectTaskPatchesSupportsSinglePatchObject(t *testing.T) {
	t.Parallel()

	content := "```json\n{\"task_name\":\"订单中心\",\"description\":\"修正批次顺序\",\"batch_no\":2,\"reason\":\"先完成库存接口\"}\n```"
	patches, err := extractArchitectTaskPatches(content)
	if err != nil {
		t.Fatalf("extractArchitectTaskPatches() error = %v", err)
	}
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].TaskName != "订单中心" || patches[0].BatchNo == nil || *patches[0].BatchNo != 2 {
		t.Fatalf("unexpected patch: %+v", patches[0])
	}
}
