package table

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	CellStyle         lipgloss.Style
	TitleCellStyle    lipgloss.Style
	SelectedCellStyle lipgloss.Style
	HeaderStyle       lipgloss.Style
	RowStyle          lipgloss.Style
}

var indigo = lipgloss.Color("#5A56E0")

func DefaultStyles() Styles {
	cell := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		MaxHeight(1)

	return Styles{
		CellStyle: cell,

		TitleCellStyle: cell.Copy().
			Bold(true),

		SelectedCellStyle: cell.Copy().
			Background(indigo),

		HeaderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(indigo).
			BorderBottom(true),

		RowStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#D9DCCF")).
			BorderBottom(true),
	}
}
