package fsm

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLStateMachine 通用 YAML 状态机定义（用于 workflow/stage/task 三套）。
type YAMLStateMachine struct {
	Initial string                    `yaml:"initial"`
	States  map[string]YAMLStateEntry `yaml:"states"`
}

// YAMLStateEntry 单个状态的 YAML 定义。
type YAMLStateEntry struct {
	Terminal    bool     `yaml:"terminal"`
	Transitions []string `yaml:"transitions"`
}

// yamlFSMFiles YAML 文件搜索路径。
var yamlFSMFiles = map[string][]string{
	"workflow": {
		"manifest/fsm/workflow.yaml",
		"../manifest/fsm/workflow.yaml",
		"../../manifest/fsm/workflow.yaml",
	},
	"stage": {
		"manifest/fsm/stage.yaml",
		"../manifest/fsm/stage.yaml",
		"../../manifest/fsm/stage.yaml",
	},
	"task": {
		"manifest/fsm/task.yaml",
		"../manifest/fsm/task.yaml",
		"../../manifest/fsm/task.yaml",
	},
}

// LoadYAMLStateMachine 加载指定名称的 YAML 状态机定义。
func LoadYAMLStateMachine(name string) (*YAMLStateMachine, error) {
	paths, ok := yamlFSMFiles[name]
	if !ok {
		return nil, fmt.Errorf("未知 FSM 名称: %s", name)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var sm YAMLStateMachine
		if err := yaml.Unmarshal(data, &sm); err != nil {
			return nil, fmt.Errorf("解析 %s 失败: %w", path, err)
		}
		return &sm, nil
	}
	return nil, fmt.Errorf("未找到 %s YAML 文件", name)
}

// ValidateWorkflowFSM 验证 workflow YAML 与 Go 代码的转换矩阵一致性。
func ValidateWorkflowFSM() error {
	sm, err := LoadYAMLStateMachine("workflow")
	if err != nil {
		return err
	}

	var errs []string

	// 验证初始状态
	if sm.Initial != WorkflowInitial().String() {
		errs = append(errs, fmt.Sprintf("初始状态不一致: YAML=%s Go=%s", sm.Initial, WorkflowInitial()))
	}

	// 验证每个状态的转换列表
	for state, entry := range sm.States {
		ws := ParseWorkflowStatus(state)
		if ws == "" {
			errs = append(errs, fmt.Sprintf("YAML 包含未知状态: %s", state))
			continue
		}

		goTargets := WorkflowTargets(ws)
		goSet := make(map[string]bool)
		for _, t := range goTargets {
			goSet[t.String()] = true
		}

		yamlSet := make(map[string]bool)
		for _, t := range entry.Transitions {
			yamlSet[t] = true
		}

		for t := range yamlSet {
			if !goSet[t] {
				errs = append(errs, fmt.Sprintf("YAML 中 %s→%s 在 Go 代码中不存在", state, t))
			}
		}
		for t := range goSet {
			if !yamlSet[t] {
				errs = append(errs, fmt.Sprintf("Go 代码中 %s→%s 在 YAML 中不存在", state, t))
			}
		}

		// 验证终态标记
		if entry.Terminal != IsTerminal(ws) {
			errs = append(errs, fmt.Sprintf("终态标记不一致: %s YAML=%v Go=%v", state, entry.Terminal, IsTerminal(ws)))
		}
	}

	// 检查 Go 中存在但 YAML 中缺失的状态
	for state := range workflowTransitions {
		if _, ok := sm.States[state.String()]; !ok {
			errs = append(errs, fmt.Sprintf("Go 代码中存在状态 %s 但 YAML 中缺失", state))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("workflow FSM 一致性校验失败:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// ValidateStageFSM 验证 stage YAML 与 Go 代码的转换矩阵一致性。
func ValidateStageFSM() error {
	sm, err := LoadYAMLStateMachine("stage")
	if err != nil {
		return err
	}

	var errs []string
	if sm.Initial != StageInitial().String() {
		errs = append(errs, fmt.Sprintf("初始状态不一致: YAML=%s Go=%s", sm.Initial, StageInitial()))
	}

	for state, entry := range sm.States {
		ss := ParseStageStatus(state)
		if ss == "" {
			errs = append(errs, fmt.Sprintf("YAML 包含未知状态: %s", state))
			continue
		}

		goTargets := StageTargets(ss)
		goSet := make(map[string]bool)
		for _, t := range goTargets {
			goSet[t.String()] = true
		}
		for _, t := range entry.Transitions {
			if !goSet[t] {
				errs = append(errs, fmt.Sprintf("YAML 中 %s→%s 在 Go 代码中不存在", state, t))
			}
		}
		for t := range goSet {
			found := false
			for _, yt := range entry.Transitions {
				if yt == t {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, fmt.Sprintf("Go 代码中 %s→%s 在 YAML 中不存在", state, t))
			}
		}
	}

	for state := range stageTransitions {
		if _, ok := sm.States[state.String()]; !ok {
			errs = append(errs, fmt.Sprintf("Go 代码中存在状态 %s 但 YAML 中缺失", state))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("stage FSM 一致性校验失败:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// ValidateTaskFSM 验证 task YAML 与 Go 代码的转换矩阵一致性。
func ValidateTaskFSM() error {
	sm, err := LoadYAMLStateMachine("task")
	if err != nil {
		return err
	}

	var errs []string
	if sm.Initial != TaskInitial().String() {
		errs = append(errs, fmt.Sprintf("初始状态不一致: YAML=%s Go=%s", sm.Initial, TaskInitial()))
	}

	for state, entry := range sm.States {
		ts := ParseTaskStatus(state)
		if ts == "" {
			errs = append(errs, fmt.Sprintf("YAML 包含未知状态: %s", state))
			continue
		}

		goTargets := TaskTargets(ts)
		goSet := make(map[string]bool)
		for _, t := range goTargets {
			goSet[t.String()] = true
		}
		for _, t := range entry.Transitions {
			if !goSet[t] {
				errs = append(errs, fmt.Sprintf("YAML 中 %s→%s 在 Go 代码中不存在", state, t))
			}
		}
		for t := range goSet {
			found := false
			for _, yt := range entry.Transitions {
				if yt == t {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, fmt.Sprintf("Go 代码中 %s→%s 在 YAML 中不存在", state, t))
			}
		}
	}

	for state := range taskTransitions {
		if _, ok := sm.States[state.String()]; !ok {
			errs = append(errs, fmt.Sprintf("Go 代码中存在状态 %s 但 YAML 中缺失", state))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("task FSM 一致性校验失败:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// ValidateAllFSM 验证所有 FSM 的 YAML 与 Go 代码一致性。
func ValidateAllFSM() error {
	var errs []string
	if err := ValidateWorkflowFSM(); err != nil {
		errs = append(errs, err.Error())
	}
	if err := ValidateStageFSM(); err != nil {
		errs = append(errs, err.Error())
	}
	if err := ValidateTaskFSM(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fmt.Errorf("FSM 一致性校验失败:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}
