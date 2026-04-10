package orchestrator

import (
	"testing"

	"easymvp/app/mvp/internal/workflow/autonomy"
)

func TestShouldDeferToDecisionCenterRework(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		resp *autonomy.DecisionResponse
		want bool
	}{
		{
			name: "nil response falls back",
			resp: nil,
			want: false,
		},
		{
			name: "audit only falls back",
			resp: &autonomy.DecisionResponse{
				Handled: false,
			},
			want: false,
		},
		{
			name: "waiting human keeps fallback",
			resp: &autonomy.DecisionResponse{
				Handled:       true,
				HumanRequired: true,
			},
			want: false,
		},
		{
			name: "auto executed defers",
			resp: &autonomy.DecisionResponse{
				Handled:       true,
				HumanRequired: false,
			},
			want: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldDeferToDecisionCenterRework(tc.resp); got != tc.want {
				t.Fatalf("shouldDeferToDecisionCenterRework(%+v) = %v, want %v", tc.resp, got, tc.want)
			}
		})
	}
}
