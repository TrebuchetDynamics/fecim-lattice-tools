package display

func ShouldShowSelectedOnly(overlayEnabled bool, isSelected bool, cellSize int) bool {
	return overlayEnabled && isSelected && cellSize >= 36
}
