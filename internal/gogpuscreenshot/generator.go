//go:build !cgo

package gogpuscreenshot

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"fecim-lattice-tools/internal/gogpuapp"
	"fecim-lattice-tools/shared/viewmodel"
)

type appFrameScreenshot struct {
	module   string
	id       viewmodel.ModuleID
	filename string
}

var appFrameScreenshots = []appFrameScreenshot{
	{module: "hysteresis", id: viewmodel.ModuleHysteresis, filename: "hysteresis-p-e-loop.png"},
	{module: "crossbar", id: viewmodel.ModuleCrossbar, filename: "crossbar-heatmap-8x8.png"},
	{module: "mnist", id: viewmodel.ModuleMNIST, filename: "mnist-accuracy-sweep.png"},
	{module: "circuits", id: viewmodel.ModuleCircuits, filename: "circuits-ispp-convergence.png"},
	{module: "comparison", id: viewmodel.ModuleComparison, filename: "comparison-architecture-bars.png"},
	{module: "eda", id: viewmodel.ModuleEDA, filename: "eda-design-overview.png"},
	{module: "docs", id: viewmodel.ModuleDocs, filename: "docs-overview.png"},
}

func Run(args []string) error {
	opts, err := ParseOptions(args)
	if err != nil {
		return err
	}
	return Generate(opts)
}

func Generate(opts Options) error {
	if opts.Width <= 0 || opts.Height <= 0 {
		return fmt.Errorf("screenshot dimensions must be positive, got %dx%d", opts.Width, opts.Height)
	}
	if opts.OutputDir == "" {
		opts.OutputDir = DefaultOptions().OutputDir
	}

	count := 0
	total := matchedScreenshotCount(opts)
	for _, screenshot := range appFrameScreenshots {
		if !opts.Matches(screenshot.module) {
			continue
		}
		count++
		log.Printf("[%d/%d] %s", count, total, screenshot.filename)
		if err := captureAppFrame(opts, screenshot); err != nil {
			return err
		}
	}

	if count == 0 {
		return fmt.Errorf("no screenshots matched -only %q", opts.Only)
	}
	log.Printf("done - %d screens generated in %s/", count, opts.OutputDir)
	return nil
}

func matchedScreenshotCount(opts Options) int {
	count := 0
	for _, screenshot := range appFrameScreenshots {
		if opts.Matches(screenshot.module) {
			count++
		}
	}
	return count
}

func captureAppFrame(opts Options, screenshot appFrameScreenshot) error {
	img, err := gogpuapp.CaptureFrameImage(screenshot.id, opts.Width, opts.Height)
	if err != nil {
		return fmt.Errorf("screenshot: render %s: %w", screenshot.module, err)
	}
	return savePNG(opts.OutputPath(screenshot.filename), img)
}

func savePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("screenshot: mkdir %s: %w", filepath.Dir(path), err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("screenshot: create %s: %w", path, err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("screenshot: encode %s: %w", path, err)
	}
	return nil
}
