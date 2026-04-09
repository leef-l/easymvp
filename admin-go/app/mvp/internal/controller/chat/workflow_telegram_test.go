package chat

import "testing"

func TestDefaultTelegramCommands(t *testing.T) {
	t.Parallel()

	commands := defaultTelegramCommands()
	if len(commands) != 4 {
		t.Fatalf("len(defaultTelegramCommands()) = %d, want 4", len(commands))
	}

	want := []struct {
		command     string
		description string
	}{
		{command: "start", description: "开始使用 / 帮助"},
		{command: "help", description: "查看所有功能"},
		{command: "list", description: "我的项目列表"},
		{command: "quit", description: "退出对话模式"},
	}

	for i, item := range want {
		if commands[i].Command != item.command || commands[i].Description != item.description {
			t.Fatalf("commands[%d] = %+v, want command=%q description=%q", i, commands[i], item.command, item.description)
		}
	}
}
