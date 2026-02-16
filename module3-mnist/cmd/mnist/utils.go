package mnistcli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func parseLevelList(levelsStr string) ([]int, error) {
	trimmed := strings.TrimSpace(levelsStr)
	if trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	levelSet := make(map[int]struct{})
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		level, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid level %q", part)
		}
		levelSet[level] = struct{}{}
	}
	if len(levelSet) == 0 {
		return nil, fmt.Errorf("no valid levels found")
	}
	result := make([]int, 0, len(levelSet))
	for level := range levelSet {
		result = append(result, level)
	}
	sort.Ints(result)
	return result, nil
}

func parseDirList(dirsStr string) ([]string, error) {
	trimmed := strings.TrimSpace(dirsStr)
	if trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	dirSet := make(map[string]struct{})
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		dirSet[filepath.Clean(part)] = struct{}{}
	}
	if len(dirSet) == 0 {
		return nil, fmt.Errorf("no valid directories found")
	}
	result := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		result = append(result, dir)
	}
	sort.Strings(result)
	return result, nil
}

func resolveWeightsPath(loadFile string) (string, error) {
	if loadFile != "" {
		return loadFile, nil
	}

	candidates := []string{
		filepath.Join("data", "pretrained-weigths", "pretrained_weights.json"),
		filepath.Join("data", "pretrained-weights", "pretrained_weights.json"),
		filepath.Join("module3-mnist", "data", "pretrained_weights.json"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("default weights not found (checked %s)", strings.Join(candidates, ", "))
}
