package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"easymvp/app/mvp/internal/regression"
	workspacepkg "easymvp/app/mvp/internal/workspace"
)

var validateManifest = regression.ValidateManifest

var inspectRiskDeliveryPolicies = func() workspacepkg.RiskDeliveryPolicyReport {
	return workspacepkg.InspectRiskDeliveryPolicies(context.Background())
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fail(err)
	}
}

func run(args []string, stdout io.Writer) error {
	manifestPath, err := resolveManifestPath(args)
	if err != nil {
		return err
	}

	report, err := validateManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("manifest 校验失败: %w", err)
	}
	riskReport := inspectRiskDeliveryPolicies()
	if !riskReport.OK() {
		return fmt.Errorf("风险交付矩阵校验失败: %s；告警: %s", riskReport.Summary(), strings.Join(riskReport.Warnings, "；"))
	}

	_, _ = fmt.Fprintf(
		stdout,
		"regression manifest OK: path=%s scenarios=%d ready=%d planned=%d\n",
		report.ManifestPath,
		report.ScenarioCount,
		report.ReadyCount,
		report.PlannedCount,
	)
	_, _ = fmt.Fprintf(stdout, "workspace delivery policies OK: %s\n", riskReport.Summary())
	return nil
}

func resolveManifestPath(args []string) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return filepath.Clean(args[0]), nil
	}
	return regression.ResolveManifestPath()
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "regression guard failed: %v\n", err)
	os.Exit(1)
}
