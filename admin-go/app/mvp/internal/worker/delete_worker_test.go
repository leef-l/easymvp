package worker

import (
	"errors"
	"os"
	"testing"
)

func TestShouldFallbackToDirectRedisErr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "auth error", err: errors.New("NOAUTH Authentication required"), want: true},
		{name: "nil redis object", err: errors.New("the Redis object is nil"), want: true},
		{name: "client unavailable", err: errors.New("redis client unavailable"), want: true},
		{name: "generic timeout", err: errors.New("i/o timeout"), want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldFallbackToDirectRedisErr(tc.err); got != tc.want {
				t.Fatalf("shouldFallbackToDirectRedisErr(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestResolveRedisPass(t *testing.T) {
	t.Parallel()

	oldPass := os.Getenv("REDIS_PASS")
	oldPassword := os.Getenv("REDIS_PASSWORD")
	defer func() {
		_ = os.Setenv("REDIS_PASS", oldPass)
		_ = os.Setenv("REDIS_PASSWORD", oldPassword)
	}()

	_ = os.Unsetenv("REDIS_PASS")
	_ = os.Unsetenv("REDIS_PASSWORD")
	if got := resolveRedisPass(); got != "" {
		t.Fatalf("resolveRedisPass() without env = %q, want empty", got)
	}

	_ = os.Setenv("REDIS_PASSWORD", "password-only")
	if got := resolveRedisPass(); got != "password-only" {
		t.Fatalf("resolveRedisPass() with REDIS_PASSWORD = %q, want %q", got, "password-only")
	}

	_ = os.Setenv("REDIS_PASS", "pass-first")
	if got := resolveRedisPass(); got != "pass-first" {
		t.Fatalf("resolveRedisPass() with REDIS_PASS = %q, want %q", got, "pass-first")
	}
}
