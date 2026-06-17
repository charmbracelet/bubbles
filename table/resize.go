package table

func (m Model) cellFrameWidth() int {
	return max(
		m.styles.Header.GetHorizontalFrameSize(),
		m.styles.Cell.GetHorizontalFrameSize(),
	)
}

func (m Model) effectiveColumnWidths() []int {
	widths := make([]int, len(m.cols))
	for i, col := range m.cols {
		widths[i] = col.Width
	}
	if m.tableWidth <= 0 {
		return widths
	}

	return resizeColumnWidths(widths, m.tableWidth, m.cellFrameWidth())
}

func resizeColumnWidths(widths []int, target, frame int) []int {
	out := append([]int(nil), widths...)

	var indices []int
	for i, w := range out {
		if w > 0 {
			indices = append(indices, i)
		}
	}

	n := len(indices)
	if n == 0 {
		return out
	}

	sum := 0
	for _, i := range indices {
		sum += out[i] + frame
	}
	if sum == target {
		return out
	}

	contentTarget := target - n*frame
	if contentTarget < n {
		contentTarget = n
	}

	contentSum := 0
	for _, i := range indices {
		contentSum += out[i]
	}

	switch {
	case contentSum < contentTarget:
		extra := contentTarget - contentSum
		for _, i := range indices {
			add := extra * out[i] / contentSum
			out[i] += add
			extra -= add
		}
		for j := 0; extra > 0; j++ {
			out[indices[j%len(indices)]]++
			extra--
		}
	case contentSum > contentTarget:
		remaining := contentTarget
		for k, i := range indices {
			if k == len(indices)-1 {
				out[i] = max(1, remaining)
				break
			}
			w := max(1, out[i]*contentTarget/contentSum)
			out[i] = w
			remaining -= w
		}
	}

	for {
		sum = 0
		for _, i := range indices {
			sum += out[i] + frame
		}
		if sum == target {
			break
		}
		if sum < target {
			out[indices[0]]++
			continue
		}
		shrunk := false
		for _, i := range indices {
			if out[i] > 1 {
				out[i]--
				shrunk = true
				break
			}
		}
		if !shrunk {
			break
		}
	}

	return out
}
