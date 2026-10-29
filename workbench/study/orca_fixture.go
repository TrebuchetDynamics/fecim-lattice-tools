//go:build orca_fixture

package study

import (
	"context"
	"os"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/project"
)

// RunOrcaFixtureBundleRoot is an opt-in synthetic acceptance seam, absent from
// normal builds. It retains real rooted commits, reconciliation and settlement.
// The caller owns roots until return and must supply an evaluator/report writer.
func RunOrcaFixtureBundleRoot(ctx context.Context, root, citations *os.Root, bundle project.Bundle, opts RunOptions, evaluate experiment.Evaluator, writeReport func(*os.Root, project.Bundle, experiment.Analysis) error) (RunOutcome, error) {
	return runBundleRoot(ctx, root, citations, bundle, opts, evaluate, writeReport)
}
