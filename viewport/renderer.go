package viewport

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	te "github.com/muesli/termenv"
)

type renderer struct {
	// Out is the io.Writer to which we should render. Generally, this will
	// be stdout.
	Out io.Writer

	// Y is the vertical offset of the rendered area in relation to the
	// terminal window.
	Y int

	// Height is the number of rows to render.
	Height int

	// TerminalHeight is the total height of the terminal.
	TerminalHeight int
}

// clear clears the viewport region.
func (r *renderer) clear() {
	buf := new(bytes.Buffer)
	moveTo(buf, r.Y, 0)
	for i := 0; i < r.Height; i++ {
		clearLine(buf)
		cursorDown(buf, 1)
	}
	r.Out.Write(buf.Bytes())
}

// sync paints the whole area.
func (r *renderer) sync(lines []string) {
	r.clear()
	moveTo(r.Out, r.Y, 0)
	r.writeLines(lines)
}

// write writes to the set writer.
func (r *renderer) writeLines(lines []string) {
	if len(lines) == 0 {
		return
	}
	io.WriteString(r.Out, strings.Join(lines, "\r\n"))
}

// Effectively scroll up. That is, insert a line at the top, pushing
// everything else down. This is roughly how ncurses does it.
func (r *renderer) insertTop(lines []string) {
	changeScrollingRegion(r.Out, r.Y, r.Y+r.Height)
	moveTo(r.Out, r.Y, 0)
	insertLine(r.Out, len(lines))
	r.writeLines(lines)
	changeScrollingRegion(r.Out, 0, r.TerminalHeight)
}

// Effectively scroll down. That is, insert a line at the bottom, pushing
// everything else up. This is roughly how ncurses does it.
func (r *renderer) insertBottom(lines []string) {
	changeScrollingRegion(r.Out, r.Y, r.Y+r.Height)
	moveTo(r.Out, r.Y+r.Height, 0)
	io.WriteString(r.Out, "\r\n"+strings.Join(lines, "\r\n"))
	changeScrollingRegion(r.Out, 0, r.TerminalHeight)
}

// Terminal Control

func changeScrollingRegion(w io.Writer, top, bottom int) {
	fmt.Fprintf(w, te.CSI+"%d;%dr", top, bottom)
}

func moveTo(w io.Writer, row, col int) {
	fmt.Fprintf(w, te.CSI+te.CursorPositionSeq, row, col)
}

func cursorDown(w io.Writer, numLines int) {
	fmt.Fprintf(w, te.CSI+te.CursorDownSeq, numLines)
}

func clearLine(w io.Writer) {
	fmt.Fprintf(w, te.CSI+te.EraseLineSeq, 2)
}

func insertLine(w io.Writer, numLines int) {
	fmt.Fprintf(w, te.CSI+"%dL", numLines)
}

func saveCursorPosition(w io.Writer) {
	fmt.Fprint(w, te.CSI+"s")
}

func restoreCursorPosition(w io.Writer) {
	fmt.Fprint(w, te.CSI+"u")
}
