package selection

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/muesli/termenv"
)

const (
	// DefaultTemplate defines the appearance of the selection and
	// can be copied as a starting point for a custom template.
	DefaultTemplate = `
{{- if .Label -}}
  {{ Bold .Label }}
{{ end -}}
{{ if .Filter }}
  {{- print "Filter: " .FilterInput }}
{{ end }}

{{- range  $i, $choice := .Choices }}
  {{- if IsScrollUpHintPosition $i }}
    {{- "⇡ " -}}
  {{- else if IsScrollDownHintPosition $i -}}
    {{- "⇣ " -}} 
  {{- else -}}
    {{- "  " -}}
  {{- end -}} 

  {{- if eq $.SelectedIndex $i }}
   {{- Foreground "32" (Bold (print "▸ " $choice.String "\n")) }}
  {{- else }}
    {{- print "  " $choice.String "\n"}}
  {{- end }}
{{- end}}`

	// DefaultFilterPlaceholder is printed instead of the
	// filter text when no filter text was entered yet.
	DefaultFilterPlaceholder = "Type to filter choices"
)

// Model is a configurable selection prompt with optional filtering
// and pagination.
type Model struct {
	// Choices represent all selectable choices of the selection.
	// Slices of arbitrary types can be converted to a slice of
	// choices using the helpers StringChoices, StringerChoices
	// and SliceChoices.
	Choices []*Choice

	// Label holds the the prompt text or question that is printed
	// above the choices in the default template (if not empty).
	Label string

	// Filter is a function that decides whether a given choice
	// should be displayed based on the text entered by the user
	// into the filter input field. If Filter is nil, filtering
	// will be disabled.
	Filter func(filterText string, choice *Choice) bool

	// FilterPlaceholder holds the text that is displayed in the
	// filter input field when no text was entered by the user yet.
	// If empty, the DefaultFilterPlaceholder is used. If Filter
	// is nil, filtering is disabled and FilterPlaceholder does
	// nothing.
	FilterPlaceholder string

	// Template holds the display template. A custom template can
	// be used to completely customize the appearance of the
	// selection prompt. If empty, DefaultTemplate is used.
	Template string

	// PageSize is the number of choices that are displayed at
	// once. If PageSize is smaller than the number of choices,
	// pagination is enabled. If PageSize is 0, pagenation is
	// always disabled.
	PageSize int

	// KeyMap determines with which keys the selection prompt is
	// controlled. By default, DefaultKeyMap is used.
	KeyMap KeyMap

	// Err holds errors that may occur during the execution of
	// the selection prompt.
	Err error

	filterInput      textinput.Model
	currentChoices   []*Choice
	availableChoices int
	currentIdx       int
	scrollOffset     int
	width            int
	tmpl             *template.Template
}

// ensure that the Model interface is implemented.
var _ tea.Model = &Model{}

func NewModel() Model {
	return Model{
		Template:          DefaultTemplate,
		FilterPlaceholder: DefaultFilterPlaceholder,
		KeyMap:            DefaultKeyMap,
	}
}

// Run executes the selection prompt in standalone mode.
func (m *Model) Run() (*Choice, error) {
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		return nil, err
	}

	choice, err := m.Choice()
	if err != nil {
		return nil, err
	}

	return choice, err
}

// Init initializes the selection prompt model.
func (m *Model) Init() tea.Cmd {
	if len(m.Choices) == 0 {
		m.Err = fmt.Errorf("no choices provided")

		return tea.Quit
	}

	m.reindexChoices()

	m.tmpl = template.New("")
	m.tmpl.Funcs(termenv.TemplateFuncs(termenv.ColorProfile()))
	m.tmpl.Funcs(template.FuncMap{
		"IsScrollDownHintPosition": func(idx int) bool {
			return m.canScrollDown() && (idx == len(m.currentChoices)-1)
		},
		"IsScrollUpHintPosition": func(idx int) bool {
			return m.canScrollUp() && idx == 0 && m.scrollOffset > 0
		},
	})

	m.tmpl, m.Err = m.tmpl.Parse(m.Template)
	if m.Err != nil {
		return tea.Quit
	}

	m.filterInput = textinput.NewModel()
	m.filterInput.Placeholder = m.FilterPlaceholder
	m.filterInput.Prompt = ""
	m.filterInput.Focus()
	m.width = 70
	m.currentChoices, m.availableChoices = m.filteredAndPagedChoices()

	return textinput.Blink
}

// Choice returns the choice that is currently selected or the final
// choice after the prompt has concluded.
func (m *Model) Choice() (*Choice, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	if len(m.currentChoices) == 0 {
		return nil, fmt.Errorf("no choices")
	}

	if m.currentIdx < 0 || m.currentIdx >= len(m.currentChoices) {
		return nil, fmt.Errorf("choice index out of bounds")
	}

	return m.currentChoices[m.currentIdx], nil
}

// Update updates the model based on the received message.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Err != nil {
		return m, tea.Quit
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch {
		case keyMatches(key, m.KeyMap.Abort):
			m.Err = fmt.Errorf("selection was aborted")

			return m, tea.Quit
		case keyMatches(key, m.KeyMap.ClearFilter):
			m.filterInput.SetValue("")

			return m, nil
		case keyMatches(key, m.KeyMap.Select):
			if len(m.currentChoices) == 0 {
				return m, nil
			}

			return m, tea.Quit
		case keyMatches(key, m.KeyMap.Down):
			m.cursorDown()

			return m, nil
		case keyMatches(key, m.KeyMap.Up):
			m.cursorUp()

			return m, nil
		case keyMatches(key, m.KeyMap.ScrollDown):
			m.scrollDown()

			return m, nil
		case keyMatches(key, m.KeyMap.ScrollUp):
			m.scrollUp()

			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case error:
		m.Err = msg

		return m, tea.Quit
	}

	if m.Filter == nil {
		return m, cmd
	}

	previousFilter := m.filterInput.Value()

	m.filterInput, cmd = m.filterInput.Update(msg)

	if m.filterInput.Value() != previousFilter {
		m.currentIdx = 0
		m.scrollOffset = 0
		m.currentChoices, m.availableChoices = m.filteredAndPagedChoices()
	}

	return m, cmd
}

// View renders the selection prompt.
func (m *Model) View() string {
	viewBuffer := &bytes.Buffer{}

	err := m.tmpl.Execute(viewBuffer, map[string]interface{}{
		"Label":         m.Label,
		"Filter":        m.Filter != nil,
		"FilterInput":   m.filterInput.View(),
		"Choices":       m.currentChoices,
		"NChoices":      len(m.currentChoices),
		"SelectedIndex": m.currentIdx,
		"PageSize":      m.PageSize,
		"IsPaged":       m.PageSize > 0 && len(m.currentChoices) > m.PageSize,
		"AllChoices":    m.Choices,
		"NAllChoices":   len(m.Choices),
	})
	if err != nil {
		m.Err = err

		return "Template Error: " + err.Error()
	}

	return wrap.String(wordwrap.String(viewBuffer.String(), m.width), m.width)
}

func (m Model) filteredAndPagedChoices() ([]*Choice, int) {
	choices := []*Choice{}

	var available, ignored int

	for _, choice := range m.Choices {
		if m.Filter != nil && !m.Filter(m.filterInput.Value(), choice) {
			continue
		}

		available++

		if m.PageSize > 0 && len(choices) >= m.PageSize {
			break
		}

		if (m.PageSize > 0) && (ignored < m.scrollOffset) {
			ignored++

			continue
		}

		choices = append(choices, choice)
	}

	return choices, available
}

func (m *Model) canScrollDown() bool {
	if m.PageSize <= 0 || m.availableChoices <= m.PageSize {
		return false
	}

	if m.scrollOffset+m.PageSize >= len(m.Choices) {
		return false
	}

	return true
}

func (m *Model) canScrollUp() bool {
	return m.scrollOffset > 0
}

func (m *Model) cursorDown() {
	if m.currentIdx == len(m.currentChoices)-1 && m.canScrollDown() {
		m.scrollDown()
	}

	m.currentIdx = min(len(m.currentChoices)-1, m.currentIdx+1)
}

func (m *Model) cursorUp() {
	if m.currentIdx == 0 && m.canScrollUp() {
		m.scrollUp()
	}

	m.currentIdx = max(0, m.currentIdx-1)
}

func (m *Model) scrollDown() {
	if m.PageSize <= 0 || m.scrollOffset+m.PageSize >= m.availableChoices {
		return
	}

	m.currentIdx = max(0, m.currentIdx-1)
	m.scrollOffset++
	m.currentChoices, m.availableChoices = m.filteredAndPagedChoices()
}

func (m *Model) scrollUp() {
	if m.PageSize <= 0 || m.scrollOffset <= 0 {
		return
	}

	m.currentIdx = min(len(m.currentChoices)-1, m.currentIdx+1)
	m.scrollOffset--
	m.currentChoices, m.availableChoices = m.filteredAndPagedChoices()
}

func (m *Model) reindexChoices() {
	for i, choice := range m.Choices {
		choice.Index = i
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
