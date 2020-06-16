package viewport

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type renderer struct {
	Out            io.Writer
	Y              int
	Height         int
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

const csi = "\x1b["

func changeScrollingRegion(w io.Writer, top, bottom int) {
	fmt.Fprintf(w, csi+"%d;%dr", top, bottom)
}

func moveTo(w io.Writer, row, col int) {
	fmt.Fprintf(w, csi+"%d;%dH", row, col)
}

func cursorDown(w io.Writer, numLines int) {
	fmt.Fprintf(w, csi+"%dB", numLines)
}

func cursorDownString(numLines int) string {
	return fmt.Sprintf(csi+"%dB", numLines)
}

func clearLine(w io.Writer) {
	fmt.Fprint(w, csi+"2K")
}

func insertLine(w io.Writer, numLines int) {
	fmt.Fprintf(w, csi+"%dL", numLines)
}
