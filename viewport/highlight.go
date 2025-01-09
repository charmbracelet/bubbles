package viewport

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func parseMatches(
	matches [][]int,
	lineWidths []int,
) (highlights []highlightInfo) {
	line := 0
	processed := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// safety check
		// XXX: return an error instead
		if start > end {
			panic(fmt.Sprintf("invalid match: %d, %d", start, end))
		}

		hi := highlightInfo{}
		hiline := [][2]int{}
		for line < len(lineWidths) {
			width := lineWidths[line]

			// out of bounds
			if start > processed+width {
				line++
				processed += width
				continue
			}

			colstart := max(0, start-processed)
			colend := clamp(end-processed, colstart, width)

			if start >= processed && start <= processed+width {
				hi.lineStart = line
			}
			if end <= processed+width {
				hi.lineEnd = line
			}

			// fmt.Printf(
			// 	"line=%d linestart=%d lineend=%d colstart=%d colend=%d start=%d end=%d processed=%d width=%d hi=%+v\n",
			// 	line, hi.lineStart, hi.lineEnd, colstart, colend, start, end, processed, width, hi,
			// )

			hiline = append(hiline, [2]int{colstart, colend})
			if end > processed+width {
				if colend > 0 {
					hi.lines = append(hi.lines, hiline)
				}
				hiline = [][2]int{}
				line++
				processed += width
				continue
			}
			if end <= processed+width {
				if colend > 0 {
					hi.lines = append(hi.lines, hiline)
				}
				break
			}
		}
		highlights = append(highlights, hi)
	}
	return
}

type highlightInfo struct {
	lineStart, lineEnd int
	lines              [][][2]int
}

func (hi highlightInfo) inLineRange(line int) bool {
	return line >= hi.lineStart && line <= hi.lineEnd
}

func (hi highlightInfo) forLine(line int) [][2]int {
	if !hi.inLineRange(line) {
		return nil
	}
	return hi.lines[line-hi.lineStart]
}

func (hi highlightInfo) coords() (line int, col int) {
	if len(hi.lines) == 0 {
		return hi.lineStart, 0
	}
	return hi.lineStart, hi.lines[0][0][0]
}

func makeHilightRanges(
	highlights []highlightInfo,
	line int,
	style lipgloss.Style,
) []lipgloss.Range {
	result := []lipgloss.Range{}
	for _, hi := range highlights {
		if !hi.inLineRange(line) {
			// out of range
			continue
		}

		for _, lihi := range hi.forLine(line) {
			result = append(result, lipgloss.NewRange(
				lihi[0], lihi[1],
				style,
			))
		}
	}
	return result
}
