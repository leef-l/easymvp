package resourcepath

import (
	"reflect"
	"testing"
)

func TestLooksLikeDirectoryResource(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "frontend/src/components", want: true},
		{value: "frontend/src/components/", want: true},
		{value: "backend/manifest/config/config.yaml", want: false},
		{value: "Dockerfile", want: false},
		{value: "Makefile", want: false},
		{value: ".gitignore", want: false},
		{value: "frontend/src/components/GameBoard.tsx", want: false},
	}

	for _, tc := range tests {
		if got := LooksLikeDirectoryResource(tc.value); got != tc.want {
			t.Fatalf("LooksLikeDirectoryResource(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestFindCodingDirectoryPlaceholders(t *testing.T) {
	got := FindCodingDirectoryPlaceholders([]string{
		"frontend/src/components",
		"frontend/src/components/",
		"frontend/src/hooks/useSnakeGame.ts",
		"Makefile",
		"frontend/src/pages",
	})
	want := []string{"frontend/src/components", "frontend/src/pages"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FindCodingDirectoryPlaceholders() = %#v, want %#v", got, want)
	}
}

func TestFindNewlyIntroducedGovernedRootFiles(t *testing.T) {
	got := FindNewlyIntroducedGovernedRootFiles(
		[]string{"package.json", "scripts/dev.js"},
		[]string{"package.json", ".gitignore", "scripts/build.js", "frontend/.gitignore"},
	)
	want := []string{".gitignore"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FindNewlyIntroducedGovernedRootFiles() = %#v, want %#v", got, want)
	}
}

func TestFindNewlyIntroducedGovernedRootFilesAllowsExistingGovernedRootFile(t *testing.T) {
	got := FindNewlyIntroducedGovernedRootFiles(
		[]string{"./.gitignore", "package.json"},
		[]string{".gitignore", "package.json", "scripts/dev.js"},
	)
	if len(got) != 0 {
		t.Fatalf("FindNewlyIntroducedGovernedRootFiles() = %#v, want empty", got)
	}
}
