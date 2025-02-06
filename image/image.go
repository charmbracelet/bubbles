package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/iterm2"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/charmbracelet/x/input"
)

// number is a global number used to generate unique image numbers.
var number int32

// nextImageNumber returns the next unique image number.
func nextNumber() int32 {
	return atomic.AddInt32(&number, 1)
}

// Protocol is the terminal graphics protocol used to render the image.
type Protocol byte

// Graphic protocol constants.
const (
	HalfBlocks Protocol = iota + 1
	Sixel
	ITerm2
	Kitty
)

// Model represents a terminal graphics image.
type Model struct {
	// The protocol used
	Protocol Protocol
	// The area covering the image in cells
	area image.Rectangle
	// The image data (exclusive with file)
	m image.Image
	// The image file path (exclusive with m)
	file string

	// The image options
	opts kitty.Options

	// seq contains the encoded image sequence buffer to render the image.
	seq string

	// The image unique id. A non-zero indicates the image was transmitted successfully.
	id int
	// The image number
	num int

	// The terminal width and height
	w, h int

	// laterDraw indicates if the image is being drawn for the first time.
	laterDraw bool
}

func newModel(area image.Rectangle) (m Model) {
	m.Protocol = ITerm2

	// We always use virtual placement for images
	m.opts.VirtualPlacement = true
	// Always chunk the image
	m.opts.Chunk = true
	// Transmit and put/display the image
	m.opts.Action = kitty.TransmitAndPut

	num := int(nextNumber())
	m.num = num
	m.opts.Number = num

	m.SetArea(area)

	return
}

// NewLocal creates a new image model from a local file.
func NewLocal(file string, area image.Rectangle) (m Model, err error) {
	m = newModel(area)
	m.file = file
	m.area = area

	f, err := os.Open(file)
	if err != nil {
		return m, fmt.Errorf("could not open image file: %w", err)
	}

	defer f.Close() //nolint:errcheck

	im, mtyp, err := image.Decode(f)
	if err != nil {
		return m, fmt.Errorf("could not decode image: %w", err)
	}

	m.m = im

	// Set the image size
	bounds := im.Bounds()
	m.opts.ImageWidth = bounds.Dx()
	m.opts.ImageHeight = bounds.Dy()

	// Optimize for JPEG images and alpha transparency
	switch mtyp {
	case "png":
		m.opts.Format = kitty.PNG
	case "jpeg":
		m.opts.Format = kitty.RGB
	default:
		m.opts.Format = kitty.RGBA
	}

	switch m.opts.Format {
	case kitty.PNG:
		m.opts.File = file
		m.opts.Transmission = kitty.File
	default:
		// Use a temporary file to store the image data
		m.opts.Transmission = kitty.TempFile
		m.opts.Compression = kitty.Zlib
	}

	return
}

// New creates a new image model given an image and an area in cells
func New(im image.Image, area image.Rectangle) Model {
	m := newModel(area)
	m.opts.Transmission = kitty.Direct
	// m.opts.Compression = kitty.Zlib
	m.opts.Format = kitty.RGBA
	m.m = im
	// Set the image size
	bounds := im.Bounds()
	m.opts.ImageWidth = bounds.Dx()
	m.opts.ImageHeight = bounds.Dy()
	m.SetArea(area)
	return m
}

// ID returns the image id unique with respect to the terminal.
func (m Model) ID() int {
	return m.id
}

// Number returns the image number unique with respect to the library.
func (m Model) Number() int {
	return m.num
}

// SetArea sets the image area in cells.
func (m *Model) SetArea(area image.Rectangle) {
	m.area = area
	m.opts.Columns = m.area.Dx()
	m.opts.Rows = m.area.Dy()
}

// Area returns the image area in cells.
func (m Model) Area() image.Rectangle {
	return m.area
}

// imageMsg is a message that transmits the image to the terminal.
type imageMsg string

// renderKitty returns the Kitty graphics sequence to render the image.
func (m *Model) renderKittyCmd() tea.Msg {
	var seq bytes.Buffer
	if err := ansi.WriteKittyGraphics(&seq, m.m, &m.opts); err != nil {
		// TODO: Error handling
		return imageMsg("")
	}

	return imageMsg(seq.String())
}

// renderIterm2 returns the iTerm2 graphics sequence to render the image.
func (m *Model) renderIterm2Cmd() tea.Msg {
	var buf bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := png.Encode(enc, m.m); err != nil {
		return imageMsg("")
	}

	if err := enc.Close(); err != nil {
		return imageMsg("")
	}

	return imageMsg(ansi.ITerm2(iterm2.File{
		Width:  iterm2.Cells(m.area.Dx()),
		Height: iterm2.Cells(m.area.Dy()),
		Inline: true,
		// DoNotMoveCursor:   true,
		IgnoreAspectRatio: true,
		Content:           buf.Bytes(),
	}))
}

// Init initializes the image model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		// m.renderKittyCmd,
		m.renderIterm2Cmd,
	// TODO: Query support
	)
}

// Update updates the image model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		cmds = append(cmds, tea.ClearScreen)
	case input.KittyGraphicsEvent:
		if msg.Options.Number == m.num &&
			msg.Options.ID > 0 &&
			bytes.Equal(msg.Payload, []byte("OK")) {
			// Store the actual image id
			m.id = msg.Options.ID
		}
	case imageMsg:
		m.seq = string(msg)
	}

	m.laterDraw = true

	return m, tea.Batch(cmds...)
}

// View returns a string representation to render the image.
func (m Model) View() string {
	// Build Kitty graphics unicode place holders
	var fgSeq string
	var extra int
	var r, g, b int
	extra, r, g, b = m.id>>24&0xff, m.id>>16&0xff, m.id>>8&0xff, m.id&0xff

	if r == 0 && g == 0 {
		fgSeq = ansi.Style{}.ForegroundColor(ansi.ExtendedColor(b)).String() //nolint:gosec
	} else {
		fgSeq = ansi.Style{}.ForegroundColor(color.RGBA{
			R: uint8(r), //nolint:gosec
			G: uint8(g), //nolint:gosec
			B: uint8(b), //nolint:gosec
			A: 0xff,
		}).String()
	}

	var s strings.Builder
	width := min(m.area.Dx(), m.w)
	height := min(m.area.Dy(), m.h)
	s.WriteString(ansi.ResetStyle)

	for y := 0; y < height; y++ {
		// As an optimization, we only write the fg color sequence id, and
		// column-row data once on the first cell. The terminal will handle
		// the rest.
		s.WriteString(fgSeq)
		s.WriteRune(kitty.Placeholder)
		s.WriteRune(kitty.Diacritic(y))
		s.WriteRune(kitty.Diacritic(0))
		if extra > 0 {
			s.WriteRune(kitty.Diacritic(extra))
		}

		for x := 1; x < width; x++ {
			s.WriteRune(kitty.Placeholder)
		}

		s.WriteString(ansi.ResetStyle)
		if y != m.area.Dy()-1 {
			s.WriteByte('\n')
		}
	}

	if m.laterDraw && m.Protocol == ITerm2 {
		// Move the cursor to the top left corner of the image
		s.WriteString(ansi.CursorBackward(m.area.Dx()))
		s.WriteString(ansi.CursorUp(m.area.Dy()))
	}

	// Write the image sequence
	s.WriteString(m.seq)

	return s.String()
}

// Rect returns a rectangle from the given x, y, width, and height.
func Rect(x, y, w, h int) image.Rectangle {
	return image.Rect(x, y, x+w, y+h)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
