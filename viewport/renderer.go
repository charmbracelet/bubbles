package viewport

import (
	"bytes"
	"fmt"
	"io"
)

type renderer struct {
	Out            io.Writer
	Y              int
	Height         int
	TerminalWidth  int
	TerminalHeight int
}

// sync paints the whole area.
func (r *renderer) sync(content string) {
	r.clear()
	moveTo(r.Out, r.Y, 0)
	r.write(content)
}

// clear clears the viewport region.
func (r *renderer) clear() {
	b := new(bytes.Buffer)
	moveTo(b, r.Y, 0)
	for i := 0; i < r.Height; i++ {
		clearLine(b)
		cursorDown(b, 1)
	}
	r.Out.Write(b.Bytes())
}

// write writes to the set writer.
func (r *renderer) write(s string) {
	if len(s) == 0 {
		return
	}

	buf := new(bytes.Buffer)
	for _, r := range []rune(s) {
		if r == '\n' {
			buf.WriteString("\r\n")
			continue
		}
		buf.WriteRune(r)
	}

	r.Out.Write(buf.Bytes())
}

// Effectively scroll up. That is, insert a line at the top, scrolling
// everything else down. This is roughly how ncurses does it.
func (r *renderer) insertTop(line string) {
	changeScrollingRegion(r.Out, r.Y, r.Y+r.Height)
	moveTo(r.Out, r.Y, 0)
	insertLine(r.Out, 1)
	io.WriteString(r.Out, line)
	changeScrollingRegion(r.Out, r.TerminalWidth, r.TerminalHeight)
}

// Effectively scroll down. That is, insert a line at the bottom, scrolling
// everything else up. This is roughly how ncurses does it.
func (r *renderer) insertBottom(line string) {
	changeScrollingRegion(r.Out, r.Y, r.Y+r.Height)
	moveTo(r.Out, r.Y+r.Height, 0)
	io.WriteString(r.Out, "\n"+line)
	changeScrollingRegion(r.Out, r.TerminalWidth, r.TerminalHeight)
}

// Screen command

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
