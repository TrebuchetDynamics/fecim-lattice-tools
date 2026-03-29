// Package io provides shared file I/O utilities for FeCIM tools.
package io

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePath checks that a file path is non-empty and does not contain
// path-traversal sequences. It returns the cleaned path or an error.
// Use this at system boundaries where paths originate from user input,
// configuration files, or external data.
func ValidatePath(path string) (string, error) {
	cleaned, err := validatePathNotEmpty(path)
	if err != nil {
		return "", err
	}

	// Reject paths that resolve to pure traversal (e.g. "../../../etc/passwd").
	// We check the cleaned path for leading ".." components. Absolute paths
	// are allowed because callers like SaveToFile legitimately use them.
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path traversal detected: %q", path)
	}

	return cleaned, nil
}

// validatePathNotEmpty checks that a path is non-empty and returns the
// cleaned version. It does not reject traversal so that internal callers
// using legitimate relative paths (e.g. "../experimental-data/...") are
// not blocked.
func validatePathNotEmpty(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("file path must not be empty")
	}
	return filepath.Clean(path), nil
}

// SaveJSON writes data to a JSON file with pretty formatting.
// Creates parent directories if they don't exist.
// Returns an error if the path is empty or if marshaling/writing fails.
func SaveJSON(path string, data interface{}) error {
	cleanPath, err := validatePathNotEmpty(path)
	if err != nil {
		return fmt.Errorf("invalid save path: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal with indentation for readability
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cleanPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", cleanPath, err)
	}

	return nil
}

// LoadJSON reads a JSON file and unmarshals it into the target.
// Returns an error if the path is empty or if reading/unmarshaling fails.
func LoadJSON(path string, target interface{}) error {
	cleanPath, err := validatePathNotEmpty(path)
	if err != nil {
		return fmt.Errorf("invalid load path: %w", err)
	}

	jsonBytes, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", cleanPath, err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", cleanPath, err)
	}

	return nil
}

// SaveJSONCompact writes data to a JSON file without formatting.
// Useful for large data files where size matters more than readability.
func SaveJSONCompact(path string, data interface{}) error {
	cleanPath, err := validatePathNotEmpty(path)
	if err != nil {
		return fmt.Errorf("invalid save path: %w", err)
	}

	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(cleanPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", cleanPath, err)
	}

	return nil
}

// FileExists checks if a file exists and is not a directory.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

// DirExists checks if a directory exists.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}

// EnsureDir creates a directory and all parent directories if they don't exist.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}
