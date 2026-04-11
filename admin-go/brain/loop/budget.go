// Package loop implements the Agent Loop Runner contract defined in
// 22-Agent-Loop规格.md.
//
// The Agent Loop Runner is the main-process component that drives a BrainKernel
// Run through a sequence of Turns: each Turn assembles a three-layer-cached
// ChatRequest, invokes the llm.Provider (non-streaming or streaming), dispatches
// any resulting tool.Tool calls, sanitizes tool results, updates the Run-level
// Budget, checks for stuck-loop patterns, and decides the next State
// transition. See 22 §2 (Architecture), §3 (Prompt Cache), §5 (Budget), §6
// (Turn anatomy), §7 (Streaming), §8 (Tool Result Sanitizer), and §9
// (LoopDetector) for normative behavior.
//
// Everything exported here is a frozen v1 contract — the concrete Runner,
// cache builder, stream consumer, sanitizer, and loop detector implementations
// will be filled in by the main-line team; the skeleton only declares the
// shape so the rest of BrainKernel can compile against it.
package loop

import (
	"time"

	"easymvp/brain/llm"
)

// Budget is the Run-level resource envelope enforced by the Agent Loop Runner.
// Every Turn, before calling the llm.Provider, the Runner MUST update the
// Used* / Elapsed* counters and call CheckTurn / CheckCost to decide whether
// to proceed. See 22-Agent-Loop规格.md §5.
type Budget struct {
	// MaxTurns is the hard upper bound on Turn count for the whole Run.
	// A Runner MUST fail the Run with a budget.exhausted BrainError once
	// UsedTurns reaches MaxTurns. See 22 §5.1.
	MaxTurns int

	// MaxCostUSD is the hard upper bound on accumulated LLM + tool cost
	// for the whole Run, in US dollars. See 22 §5.2.
	MaxCostUSD float64

	// MaxToolCalls is the hard upper bound on the number of tool
	// invocations allowed across all Turns of the Run. See 22 §5.3.
	MaxToolCalls int

	// MaxLLMCalls is the hard upper bound on the number of llm.Provider
	// Complete / Stream calls allowed across all Turns of the Run.
	// See 22 §5.3.
	MaxLLMCalls int

	// MaxDuration is the wall-clock upper bound on the whole Run. The
	// Runner MUST abort with a budget.timeout BrainError once ElapsedTime
	// exceeds MaxDuration. See 22 §5.4.
	MaxDuration time.Duration

	// UsedTurns is the monotonically increasing count of Turns executed so
	// far. It MUST be updated exactly once per Turn, after the Turn has
	// been persisted. See 22 §5.1 and §6.4.
	UsedTurns int

	// UsedCostUSD is the monotonically increasing accumulated cost of the
	// Run, in US dollars, summed over llm.Usage.CostUSD on every Turn.
	// See 22 §5.2.
	UsedCostUSD float64

	// UsedToolCalls is the monotonically increasing count of tool
	// invocations across all Turns of the Run. See 22 §5.3.
	UsedToolCalls int

	// UsedLLMCalls is the monotonically increasing count of llm.Provider
	// Complete / Stream calls across all Turns of the Run. See 22 §5.3.
	UsedLLMCalls int

	// ElapsedTime is the wall-clock time consumed so far by the Run,
	// measured as time.Now().Sub(Run.StartedAt) at check time. See 22 §5.4.
	ElapsedTime time.Duration
}

// CheckTurn validates that a new Turn is allowed to start under the current
// Budget counters. It MUST return a *errors.BrainError with Class=Permanent
// and ErrorCode="budget.turns_exhausted" when UsedTurns >= MaxTurns, and nil
// otherwise. See 22-Agent-Loop规格.md §5.1.
func (b *Budget) CheckTurn() error {
	panic("unimplemented: 22-Agent-Loop规格.md §5.1 Budget.CheckTurn")
}

// CheckCost validates that the current Budget still has headroom for another
// LLM / tool call. It MUST return a *errors.BrainError with
// ErrorCode="budget.cost_exhausted" when UsedCostUSD >= MaxCostUSD, and nil
// otherwise. See 22-Agent-Loop规格.md §5.2.
func (b *Budget) CheckCost() error {
	panic("unimplemented: 22-Agent-Loop规格.md §5.2 Budget.CheckCost")
}

// Remaining returns a snapshot of the still-available Budget in the form
// consumed by llm.ChatRequest.RemainingBudget. The snapshot is a point-in-time
// copy and MUST NOT share mutable state with the Budget receiver. See
// 22-Agent-Loop规格.md §5 and §6.3 (ChatRequest.RemainingBudget).
func (b *Budget) Remaining() llm.BudgetSnapshot {
	panic("unimplemented: 22-Agent-Loop规格.md §5 Budget.Remaining")
}
