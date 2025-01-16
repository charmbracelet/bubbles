package image

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"strings"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
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

	// The image unique id. A non-zero indicates the image was transmitted successfully.
	id int
	// The image number
	num int

	// Whether the image was transmitted
	didTransmit bool

	// The terminal width and height
	w, h int
}

func newModel(area image.Rectangle) (m Model) {
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
	// TODO: Fix me! This currently doesn't work with PNG
	// ext := filepath.Ext(file)
	// if strings.Contains(ext, "png") {
	// 	// We're done here, there's no need to decode the image.
	// 	m.opts.Format = kitty.PNG
	// 	m.opts.File = file
	// 	return
	// }

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

	// Use a temporary file to store the image data
	m.opts.Transmission = kitty.TempFile

	// TODO: Enable compression
	// m.opts.Compression = kitty.Zlib

	// Set the image size
	bounds := im.Bounds()
	m.opts.ImageWidth = bounds.Dx()
	m.opts.ImageHeight = bounds.Dy()

	// Optimize for JPEG images and alpha transparency
	switch mtyp {
	case "jpeg":
		m.opts.Format = kitty.RGB
	default:
		m.opts.Format = kitty.RGBA
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
	m.didTransmit = false
}

// Area returns the image area in cells.
func (m Model) Area() image.Rectangle {
	return m.area
}

// transmit is a command that transmits the image to the terminal.
func (m *Model) transmit() tea.Msg {
	var seq bytes.Buffer
	if err := ansi.WriteKittyGraphics(&seq, m.m, &m.opts); err != nil {
		// TODO: Error handling
		return nil
	}

	m.didTransmit = true
	return tea.RawMsg{Msg: seq.String()}
}

// Init initializes the image model.
func (m Model) Init() (tea.Model, tea.Cmd) {
	return m, tea.Batch(
	// TODO: Query support
	)
}

// Update updates the image model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.didTransmit = false
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case input.KittyGraphicsEvent:
		if msg.Options.Number == m.num &&
			msg.Options.ID > 0 &&
			bytes.Equal(msg.Payload, []byte("OK")) {
			// Store the actual image id
			m.id = msg.Options.ID
		}
	}

	if !m.didTransmit {
		cmds = append(cmds, m.transmit)
		m.didTransmit = true
	}

	return m, tea.Batch(cmds...)
}

// View returns a string representation to render the image.
func (m Model) View() string {
	if m.id == 0 {
		// TODO: Maybe use a spinner?
		return "Loading image..."
	}

	log.Printf("area: %v", m.area)

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
