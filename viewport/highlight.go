package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

// parseMatches converts the given matches into highlight ranges.
//
// Assumptions:
// - matches are measured in bytes, e.g. what [regex.FindAllStringIndex] would return;
// - matches were made against the given content;
// - matches are in order
// - matches do not overlap
//
// We'll then convert the ranges into [highlightInfo]s, which hold the starting
// line and the grapheme positions.
func parseMatches(
	matches [][]int,
	content string,
) []highlightInfo {
	if len(matches) == 0 {
		return nil
	}

	line := 0
	graphemePos := 0
	previousLinesOffset := 0
	bytePos := 0

	highlights := make([]highlightInfo, 0, len(matches))
	gr := uniseg.NewGraphemes(ansi.Strip(content))

matchLoop:
	for _, match := range matches {
		hi := highlightInfo{
			lines: map[int][][2]int{},
		}
		byteStart, byteEnd := match[0], match[1]

		for byteStart > bytePos {
			if !gr.Next() {
				break
			}
			graphemePos += len(gr.Str())
			if content[bytePos] == '\n' {
				previousLinesOffset = graphemePos
				line++
			}
			bytePos++
		}

		hi.lineStart = line
		hi.lineEnd = line

		graphemeStart := graphemePos
		graphemeEnd := graphemePos

		for byteEnd >= bytePos {
			if bytePos == byteEnd {
				graphemeEnd = graphemePos
				colstart := max(0, graphemeStart-previousLinesOffset)
				colend := max(graphemeEnd-previousLinesOffset, colstart)

				// fmt.Printf(
				// 	"no line=%d linestart=%d lineend=%d colstart=%d colend=%d start=%d end=%d processed=%d width=%d\n",
				// 	line, hi.lineStart, hi.lineEnd, colstart, colend, graphemeStart, graphemeEnd, previousLinesOffset, graphemePos-previousLinesOffset,
				// )
				//
				if colend > colstart {
					hi.lines[line] = append(hi.lines[line], [2]int{colstart, colend})
					hi.lineEnd = line
				}
				highlights = append(highlights, hi)
				continue matchLoop
			}

			if content[bytePos] == '\n' {
				graphemeEnd = graphemePos
				colstart := max(0, graphemeStart-previousLinesOffset)
				colend := max(graphemeEnd-previousLinesOffset+1, colstart)

				// fmt.Printf(
				// 	"nl line=%d linestart=%d lineend=%d colstart=%d colend=%d start=%d end=%d processed=%d width=%d\n",
				// 	line, hi.lineStart, hi.lineEnd, colstart, colend, graphemeStart, graphemeEnd, previousLinesOffset, graphemePos-previousLinesOffset,
				// )

				if colend > colstart {
					hi.lines[line] = append(hi.lines[line], [2]int{colstart, colend})
					hi.lineEnd = line
				}

				previousLinesOffset = graphemePos + len(gr.Str())
				line++
			}

			if !gr.Next() {
				break
			}

			bytePos++
			graphemePos += len(gr.Str())
		}

		highlights = append(highlights, hi)
	}

	return highlights
}

type highlightInfo struct {
	// in which line this highlight starts and ends
	lineStart, lineEnd int

	// the grapheme highlight ranges for each of these lines
	lines map[int][][2]int
}

// func (hi *highlightInfo) addToLine(line int, rng [2]int) {
// 	hi.lines[line] = append(hi.lines[line], rng)
// }
//
// func (hi highlightInfo) forLine(line int) ([][2]int, bool) {
// 	got, ok := hi.lines[line]
// 	return got, ok
// }

func (hi highlightInfo) coords() (line int, col int) {
	// if len(hi.lines) == 0 {
	return hi.lineStart, 0
	// }
	// return hi.lineStart, hi.lines[line][0][0]
}

func makeHilightRanges(
	highlights []highlightInfo,
	line int,
	style lipgloss.Style,
) []lipgloss.Range {
	result := []lipgloss.Range{}
	for _, hi := range highlights {
		lihis, ok := hi.lines[line]
		if !ok {
			continue
		}
		for _, lihi := range lihis {
			if lihi == [2]int{} {
				continue
			}
			result = append(result, lipgloss.NewRange(lihi[0], lihi[1], style))
		}
	}
	return result
}
