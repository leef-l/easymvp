package service

import (
	"fmt"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// ---------------------------------------------------------------------------
// Topological Layer Computation (DAG-based task scheduling)
// ---------------------------------------------------------------------------

// computeTopologicalLayers partitions tasks into dependency layers using Kahn's
// algorithm. Layer 0 contains tasks with zero in-degree (no dependencies);
// each subsequent layer contains tasks whose dependencies are all satisfied by
// earlier layers. Returns nil if a cycle is detected.
func computeTopologicalLayers(tasks []entity.WorkflowCompiledTasks, deps []entity.TaskDependencies) [][]string {
	if len(tasks) == 0 {
		return nil
	}

	graph := buildDependencyGraph(deps)

	// Collect the full set of task IDs present in the task list.
	taskSet := make(map[string]bool, len(tasks))
	for _, t := range tasks {
		taskSet[t.Id] = true
	}

	// Compute in-degree for each task. Only count edges whose source is also
	// in the task set (ignore dangling references to removed tasks).
	inDegree := make(map[string]int, len(tasks))
	for _, t := range tasks {
		inDegree[t.Id] = 0
	}
	for taskID, depIDs := range graph {
		if !taskSet[taskID] {
			continue
		}
		count := 0
		for _, dep := range depIDs {
			if taskSet[dep] {
				count++
			}
		}
		inDegree[taskID] = count
	}

	visited := make(map[string]bool, len(tasks))
	var layers [][]string

	for len(visited) < len(taskSet) {
		var layer []string
		for id := range taskSet {
			if visited[id] {
				continue
			}
			if inDegree[id] == 0 {
				layer = append(layer, id)
			}
		}
		if len(layer) == 0 {
			// Remaining tasks form a cycle — cannot proceed.
			return nil
		}
		for _, id := range layer {
			visited[id] = true
		}
		// Decrease in-degree for dependents of this layer.
		for _, id := range layer {
			for taskID, depIDs := range graph {
				if visited[taskID] {
					continue
				}
				for _, dep := range depIDs {
					if dep == id {
						inDegree[taskID]--
					}
				}
			}
		}
		layers = append(layers, layer)
	}

	return layers
}

// buildDependencyGraph constructs a map from task_id to the list of task IDs
// it depends on, based on the task_dependencies table rows.
func buildDependencyGraph(deps []entity.TaskDependencies) map[string][]string {
	graph := make(map[string][]string, len(deps))
	for _, d := range deps {
		graph[d.TaskId] = append(graph[d.TaskId], d.DependsOnTaskId)
	}
	return graph
}

// validateDAG checks the dependency graph for cycles. Returns nil if the graph
// is a valid DAG, or an error describing the cycle if one is detected.
// Uses iterative Kahn's algorithm: if not all nodes are consumed, a cycle exists.
func validateDAG(graph map[string][]string) error {
	// Collect all nodes referenced in the graph (both as keys and values).
	nodes := make(map[string]bool)
	for k, deps := range graph {
		nodes[k] = true
		for _, d := range deps {
			nodes[d] = true
		}
	}
	if len(nodes) == 0 {
		return nil
	}

	// Build in-degree map.
	inDegree := make(map[string]int, len(nodes))
	for n := range nodes {
		inDegree[n] = 0
	}
	for _, deps := range graph {
		for _, d := range deps {
			inDegree[d]++
		}
	}

	// Kahn's: repeatedly remove zero in-degree nodes.
	queue := make([]string, 0, len(nodes))
	for n, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, n)
		}
	}

	consumed := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		consumed++
		for _, dep := range graph[node] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if consumed < len(nodes) {
		return fmt.Errorf("cycle detected in dependency graph: %d of %d nodes could not be resolved", len(nodes)-consumed, len(nodes))
	}
	return nil
}
