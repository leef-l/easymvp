package service

import (
	"sort"
	"testing"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// ---------------------------------------------------------------------------
// computeTopologicalLayers tests
// ---------------------------------------------------------------------------

func TestComputeTopologicalLayers_EmptyInput(t *testing.T) {
	result := computeTopologicalLayers(nil, nil)
	if result != nil {
		t.Fatalf("expected nil for empty input, got %v", result)
	}
}

func TestComputeTopologicalLayers_SingleTask(t *testing.T) {
	tasks := []entity.WorkflowCompiledTasks{
		{Id: "t1"},
	}
	result := computeTopologicalLayers(tasks, nil)
	if len(result) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(result))
	}
	if len(result[0]) != 1 || result[0][0] != "t1" {
		t.Fatalf("expected layer [t1], got %v", result[0])
	}
}

func TestComputeTopologicalLayers_LinearChain(t *testing.T) {
	// t1 -> t2 -> t3 (t2 depends on t1, t3 depends on t2)
	tasks := []entity.WorkflowCompiledTasks{
		{Id: "t1"},
		{Id: "t2"},
		{Id: "t3"},
	}
	deps := []entity.TaskDependencies{
		{TaskId: "t2", DependsOnTaskId: "t1"},
		{TaskId: "t3", DependsOnTaskId: "t2"},
	}

	result := computeTopologicalLayers(tasks, deps)
	if len(result) != 3 {
		t.Fatalf("expected 3 layers for linear chain, got %d", len(result))
	}
	if result[0][0] != "t1" {
		t.Fatalf("expected t1 in layer 0, got %v", result[0])
	}
	if result[1][0] != "t2" {
		t.Fatalf("expected t2 in layer 1, got %v", result[1])
	}
	if result[2][0] != "t3" {
		t.Fatalf("expected t3 in layer 2, got %v", result[2])
	}
}

func TestComputeTopologicalLayers_ParallelTasks(t *testing.T) {
	// t1, t2, t3 all independent -> should be in same layer
	tasks := []entity.WorkflowCompiledTasks{
		{Id: "t1"},
		{Id: "t2"},
		{Id: "t3"},
	}

	result := computeTopologicalLayers(tasks, nil)
	if len(result) != 1 {
		t.Fatalf("expected 1 layer for parallel tasks, got %d", len(result))
	}
	if len(result[0]) != 3 {
		t.Fatalf("expected 3 tasks in layer 0, got %d", len(result[0]))
	}
}

func TestComputeTopologicalLayers_DiamondShape(t *testing.T) {
	// t1 -> t2, t1 -> t3, t2 -> t4, t3 -> t4 (diamond)
	tasks := []entity.WorkflowCompiledTasks{
		{Id: "t1"},
		{Id: "t2"},
		{Id: "t3"},
		{Id: "t4"},
	}
	deps := []entity.TaskDependencies{
		{TaskId: "t2", DependsOnTaskId: "t1"},
		{TaskId: "t3", DependsOnTaskId: "t1"},
		{TaskId: "t4", DependsOnTaskId: "t2"},
		{TaskId: "t4", DependsOnTaskId: "t3"},
	}

	result := computeTopologicalLayers(tasks, deps)
	if len(result) != 3 {
		t.Fatalf("expected 3 layers for diamond, got %d", len(result))
	}

	// Layer 0: t1
	if len(result[0]) != 1 || result[0][0] != "t1" {
		t.Fatalf("expected [t1] in layer 0, got %v", result[0])
	}

	// Layer 1: t2, t3 (order may vary)
	layer1 := make([]string, len(result[1]))
	copy(layer1, result[1])
	sort.Strings(layer1)
	if len(layer1) != 2 || layer1[0] != "t2" || layer1[1] != "t3" {
		t.Fatalf("expected [t2, t3] in layer 1, got %v", layer1)
	}

	// Layer 2: t4
	if len(result[2]) != 1 || result[2][0] != "t4" {
		t.Fatalf("expected [t4] in layer 2, got %v", result[2])
	}
}

func TestComputeTopologicalLayers_CycleDetection(t *testing.T) {
	// t1 -> t2 -> t3 -> t1 (cycle)
	tasks := []entity.WorkflowCompiledTasks{
		{Id: "t1"},
		{Id: "t2"},
		{Id: "t3"},
	}
	deps := []entity.TaskDependencies{
		{TaskId: "t2", DependsOnTaskId: "t1"},
		{TaskId: "t3", DependsOnTaskId: "t2"},
		{TaskId: "t1", DependsOnTaskId: "t3"},
	}

	result := computeTopologicalLayers(tasks, deps)
	if result != nil {
		t.Fatalf("expected nil for cyclic graph, got %v", result)
	}
}

func TestComputeTopologicalLayers_DanglingDependency(t *testing.T) {
	// t2 depends on t_removed (not in task set) - should be ignored
	tasks := []entity.WorkflowCompiledTasks{
		{Id: "t1"},
		{Id: "t2"},
	}
	deps := []entity.TaskDependencies{
		{TaskId: "t2", DependsOnTaskId: "t_removed"},
	}

	result := computeTopologicalLayers(tasks, deps)
	if len(result) != 1 {
		t.Fatalf("expected 1 layer (dangling dep ignored), got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("expected 2 tasks in layer 0, got %d", len(result[0]))
	}
}

// ---------------------------------------------------------------------------
// buildDependencyGraph tests
// ---------------------------------------------------------------------------

func TestBuildDependencyGraph_Empty(t *testing.T) {
	graph := buildDependencyGraph(nil)
	if len(graph) != 0 {
		t.Fatalf("expected empty graph, got %v", graph)
	}
}

func TestBuildDependencyGraph_MultipleDeps(t *testing.T) {
	deps := []entity.TaskDependencies{
		{TaskId: "t2", DependsOnTaskId: "t1"},
		{TaskId: "t3", DependsOnTaskId: "t1"},
		{TaskId: "t3", DependsOnTaskId: "t2"},
	}

	graph := buildDependencyGraph(deps)

	if len(graph["t2"]) != 1 || graph["t2"][0] != "t1" {
		t.Fatalf("expected t2 depends on [t1], got %v", graph["t2"])
	}

	t3Deps := graph["t3"]
	sort.Strings(t3Deps)
	if len(t3Deps) != 2 || t3Deps[0] != "t1" || t3Deps[1] != "t2" {
		t.Fatalf("expected t3 depends on [t1, t2], got %v", t3Deps)
	}
}

// ---------------------------------------------------------------------------
// validateDAG tests
// ---------------------------------------------------------------------------

func TestValidateDAG_EmptyGraph(t *testing.T) {
	err := validateDAG(map[string][]string{})
	if err != nil {
		t.Fatalf("expected nil error for empty graph, got %v", err)
	}
}

func TestValidateDAG_ValidDAG(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {},
	}
	err := validateDAG(graph)
	if err != nil {
		t.Fatalf("expected valid DAG, got error: %v", err)
	}
}

func TestValidateDAG_CycleDetected(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	err := validateDAG(graph)
	if err == nil {
		t.Fatal("expected cycle detection error, got nil")
	}
}

func TestValidateDAG_SelfCycle(t *testing.T) {
	graph := map[string][]string{
		"a": {"a"},
	}
	err := validateDAG(graph)
	if err == nil {
		t.Fatal("expected self-cycle detection error, got nil")
	}
}

func TestValidateDAG_ComplexValidGraph(t *testing.T) {
	// Diamond + extra node
	graph := map[string][]string{
		"a": {"c", "d"},
		"b": {"c"},
		"c": {},
		"d": {},
		"e": {"a", "b"},
	}
	err := validateDAG(graph)
	if err != nil {
		t.Fatalf("expected valid DAG, got error: %v", err)
	}
}
