package scrollbar

// Msg signals that scrollbar parameters must be updated.
type Msg struct {
	Total   int
	Visible int
	Offset  int
}

// HeightMsg signals that scrollbar height must be updated.
type HeightMsg int

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
