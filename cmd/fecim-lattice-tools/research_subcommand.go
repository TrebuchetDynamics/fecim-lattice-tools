package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var researchRunner = runResearchTool

func runResearchSubcommand(args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}
	return researchRunner(args)
}

func runResearchTool(args []string) error {
	python := os.Getenv("FECIM_RESEARCH_PYTHON")
	if python == "" {
		python = "python3"
	}
	root := filepath.Clean(filepath.Join(".", "tools", "research", "research_cli.py"))
	cmdArgs := append([]string{root}, args...)
	cmd := exec.Command(python, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("research tool: %w", err)
	}
	return nil
}
