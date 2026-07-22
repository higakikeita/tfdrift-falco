package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/aws"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/spf13/cobra"
)

// maxDriftExitCode caps the drift-count exit code well below the shell-reserved
// range (126+) so "N drifts" stays a plain, unambiguous status.
const maxDriftExitCode = 250

// newScanCmd builds the `tfdrift scan` subcommand: a one-shot, read-only
// reconcile between Terraform state and live cloud state. No Falco, no
// CloudTrail — deterministic and CI-friendly (#334, ADR-0014).
func newScanCmd() *cobra.Command {
	var (
		scanConfig  string
		scanRegions []string
		scanOutput  string
		failOnDrift bool
	)
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "One-shot reconcile: compare live cloud state against Terraform state and report drift",
		Long: `scan performs a single read-only reconcile between your Terraform state and the
actual cloud resources (AWS), then reports unmanaged / missing / modified
resources and exits with a code reflecting the result.

Unlike the daemon it needs no Falco and no CloudTrail — only read access to the
Terraform state and the cloud provider. It is deterministic and suited to CI
(nightly drift gate) and to answering "right now, does reality match my code?".

Exit code: 0 = no drift; otherwise the number of drifted resources (capped at
250). Use --fail-on-drift=false to always exit 0 and only report.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			code, err := runScan(scanConfig, scanRegions, scanOutput, failOnDrift)
			if err != nil {
				return err
			}
			if code != 0 {
				os.Exit(code)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&scanConfig, "config", "", "config file (default is config.yaml)")
	cmd.Flags().StringSliceVar(&scanRegions, "region", nil, "AWS region(s) to scan; overrides config (e.g. --region us-east-1,ap-northeast-1)")
	cmd.Flags().StringVar(&scanOutput, "output", "human", "output mode: human or json")
	cmd.Flags().BoolVar(&failOnDrift, "fail-on-drift", true, "exit non-zero (drift count) when drift is found")
	return cmd
}

// runScan executes the reconcile and returns the process exit code.
func runScan(cfgPath string, regionsOverride []string, output string, failOnDrift bool) (int, error) {
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return 0, fmt.Errorf("load config %q: %w", cfgPath, err)
	}
	if !cfg.Providers.AWS.Enabled {
		return 0, fmt.Errorf("scan currently supports AWS only; enable providers.aws in %s", cfgPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	sm, err := terraform.NewStateManager(cfg.Providers.AWS.State)
	if err != nil {
		return 0, fmt.Errorf("create state manager: %w", err)
	}
	if err := sm.Load(ctx); err != nil {
		return 0, fmt.Errorf("load terraform state: %w", err)
	}
	tfResources := sm.GetAllResources()

	regions := regionsOverride
	if len(regions) == 0 {
		regions = cfg.Providers.AWS.Regions
	}
	if len(regions) == 0 {
		regions = []string{"us-east-1"}
	}

	// Discover across all regions first, then compare ONCE: comparing per
	// region would flag a resource as "missing" whenever it lives in another
	// region. Dedup by ID so a global resource (IAM/S3) seen in several regions
	// isn't counted as multiple unmanaged resources.
	var allAWS []*types.DiscoveredResource
	seen := make(map[string]bool)
	for _, region := range regions {
		dc, err := aws.NewDiscoveryClient(ctx, region)
		if err != nil {
			return 0, fmt.Errorf("create discovery client (%s): %w", region, err)
		}
		res, err := dc.DiscoverAll(ctx)
		if err != nil {
			return 0, fmt.Errorf("discover resources (%s): %w", region, err)
		}
		for _, r := range res {
			if r != nil && !seen[r.ID] {
				seen[r.ID] = true
				allAWS = append(allAWS, r)
			}
		}
	}

	drift := aws.CompareStateWithActual(tfResources, allAWS)

	report := renderDriftReport(drift, output, len(tfResources), len(allAWS), regions)
	fmt.Println(report)

	return exitCodeForDrift(driftTotal(drift), failOnDrift), nil
}

// driftTotal is the number of drifted resources across all categories.
func driftTotal(d *types.DriftResult) int {
	if d == nil {
		return 0
	}
	return len(d.UnmanagedResources) + len(d.MissingResources) + len(d.ModifiedResources)
}

// exitCodeForDrift maps a drift count to a process exit code (0 = clean).
func exitCodeForDrift(total int, failOnDrift bool) int {
	if total == 0 || !failOnDrift {
		return 0
	}
	if total > maxDriftExitCode {
		return maxDriftExitCode
	}
	return total
}

// renderDriftReport formats the reconcile result. Pure (no IO) so it is unit
// tested without cloud access.
func renderDriftReport(d *types.DriftResult, output string, tfCount, awsCount int, regions []string) string {
	if d == nil {
		d = &types.DriftResult{}
	}
	if output == "json" {
		payload := map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"regions":   regions,
			"summary": map[string]int{
				"terraform_resources": tfCount,
				"cloud_resources":     awsCount,
				"unmanaged":           len(d.UnmanagedResources),
				"missing":             len(d.MissingResources),
				"modified":            len(d.ModifiedResources),
				"total_drift":         driftTotal(d),
			},
			"drift": d,
		}
		b, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Sprintf(`{"error":%q}`, err.Error())
		}
		return string(b)
	}

	var b strings.Builder
	total := driftTotal(d)
	fmt.Fprintf(&b, "TFDrift scan — regions: %s\n", strings.Join(regions, ", "))
	fmt.Fprintf(&b, "  terraform resources: %d | cloud resources: %d\n", tfCount, awsCount)
	if total == 0 {
		b.WriteString("\n✅ No drift: live cloud state matches Terraform state.\n")
		return b.String()
	}
	fmt.Fprintf(&b, "\n⚠️  Drift detected: %d resource(s) — unmanaged=%d missing=%d modified=%d\n",
		total, len(d.UnmanagedResources), len(d.MissingResources), len(d.ModifiedResources))

	if len(d.UnmanagedResources) > 0 {
		b.WriteString("\nUnmanaged (in cloud, not in Terraform):\n")
		for _, r := range d.UnmanagedResources {
			fmt.Fprintf(&b, "  + %s %s (%s)\n", r.Type, r.ID, r.Region)
		}
	}
	if len(d.MissingResources) > 0 {
		b.WriteString("\nMissing (in Terraform, not in cloud):\n")
		for _, r := range d.MissingResources {
			fmt.Fprintf(&b, "  - %s.%s (%s)\n", r.Type, r.Name, r.ID)
		}
	}
	if len(d.ModifiedResources) > 0 {
		b.WriteString("\nModified (attribute differences):\n")
		for _, r := range d.ModifiedResources {
			fmt.Fprintf(&b, "  ~ %s %s\n", r.ResourceType, r.ResourceID)
			for _, f := range r.Differences {
				fmt.Fprintf(&b, "      %s: terraform=%v actual=%v\n", f.Field, f.TerraformValue, f.ActualValue)
			}
		}
	}
	return b.String()
}
