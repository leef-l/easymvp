package service

import (
	"context"
	"errors"
	"testing"
)

func TestBuildStartupSnapshotSafeMode(t *testing.T) {
	t.Parallel()

	snapshot := buildStartupSnapshot(context.Background(), StartupConfig{
		SafeMode:          true,
		BrainServeBaseURL: "http://127.0.0.1:7701",
		Options: StartupConfigOptions{
			DataRoot:          StartupOption{Value: "./var", UsingDefault: true},
			DBPath:            StartupOption{Value: "./var/data/easymvp.db", UsingDefault: true},
			MigrationPath:     StartupOption{Value: "./manifest/migrations", UsingDefault: true},
			BrainServeBaseURL: StartupOption{Value: "http://127.0.0.1:7701", UsingDefault: true},
			ServerAddress:     StartupOption{Value: ":8000", UsingDefault: true},
			SafeMode:          StartupOption{Value: "true", UsingDefault: false},
		},
	}, WorkerManagerStatus{}, nil)

	if snapshot.Status != "recovery" {
		t.Fatalf("snapshot status = %q, want recovery", snapshot.Status)
	}
	if snapshot.Ready {
		t.Fatal("safe mode snapshot should not be ready")
	}
	if snapshot.WorkerStatus != "disabled-safe-mode" {
		t.Fatalf("worker status = %q, want disabled-safe-mode", snapshot.WorkerStatus)
	}
	if len(snapshot.Diagnostics) == 0 {
		t.Fatal("expected safe mode diagnostics")
	}
}

func TestBuildStartupSnapshotRuntimeFailure(t *testing.T) {
	t.Parallel()

	snapshot := buildStartupSnapshot(context.Background(), StartupConfig{
		BrainServeBaseURL: "http://127.0.0.1:7701",
		Options: StartupConfigOptions{
			BrainServeBaseURL: StartupOption{Value: "http://127.0.0.1:7701", UsingDefault: false},
			DataRoot:          StartupOption{Value: "/tmp/easymvp", UsingDefault: false},
			DBPath:            StartupOption{Value: "/tmp/easymvp/data/easymvp.db", UsingDefault: false},
			MigrationPath:     StartupOption{Value: "/tmp/easymvp/migrations", UsingDefault: false},
			ServerAddress:     StartupOption{Value: ":8000", UsingDefault: false},
			SafeMode:          StartupOption{Value: "false", UsingDefault: false},
		},
	}, WorkerManagerStatus{Started: true}, errors.New("dial tcp 127.0.0.1:7701: connect refused"))

	if snapshot.Status != "attention" {
		t.Fatalf("snapshot status = %q, want attention", snapshot.Status)
	}
	if snapshot.RuntimeStatus != "degraded" {
		t.Fatalf("runtime status = %q, want degraded", snapshot.RuntimeStatus)
	}
	if snapshot.WorkerStatus != "running" {
		t.Fatalf("worker status = %q, want running", snapshot.WorkerStatus)
	}
}
