package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	tea "github.com/charmbracelet/bubbletea"
)

type TableStyle struct {
	innerStyle table.Styles
	outerStyle lipgloss.Style
}

// What the user opens to, should have general information on the market.
type Dashboard struct {
	name string
	// screen height
	height int
	// screen width
	width int
	// List of tables, tables[0] is commodities, tables[1] is stocks, tables[3] is news
	tables []table.Model
	// index of the focused table
	focused int
	// unfocused style
	focusedStyle TableStyle
	// focused style
	unfocusedStyle TableStyle
}

func (d *Dashboard) Init() tea.Cmd {
	cmdtyTable := table.New()

	stockTable := table.New()

	lipgloss.NewStyle().BorderForeground(lipgloss.Color("#FFFFFF")).Border(lipgloss.NormalBorder())

	foucsedInnerStyle := table.Styles{
		Header:   lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#703FFD")).Foreground(lipgloss.Color("#FFFFFF")),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")),
	}

	unfocusedInnerStyle := table.Styles{
		Header:   lipgloss.NewStyle().BorderForeground(lipgloss.Color("#703FFD")),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFF")),
	}

	d.focusedStyle = TableStyle{
		innerStyle: foucsedInnerStyle,
		outerStyle: lipgloss.NewStyle().BorderForeground(lipgloss.Color("#703FFD")).Border(lipgloss.NormalBorder()),
	}

	d.unfocusedStyle = TableStyle{
		innerStyle: unfocusedInnerStyle,
		outerStyle: lipgloss.NewStyle().BorderForeground(lipgloss.Color("#FFFFFF")).Border(lipgloss.NormalBorder()),
	}

	d.tables = append(d.tables, cmdtyTable, stockTable)

	d.tables[d.focused].SetStyles(d.focusedStyle.innerStyle)
	return nil
}

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

		// Redraw tables

		// NOTE: For some reason using exactly 1/2 the width and 2/3 the screen
		// draws the border past it's boundaries. whatever make it slightly less
		topTablesWidth := int(float64(d.width) * .47)
		topTablesHeight := int(float64(d.height) * .65)

		d.tables[0].SetWidth(topTablesWidth)
		d.tables[1].SetWidth(topTablesWidth)

		d.tables[0].SetHeight(topTablesHeight)
		d.tables[1].SetHeight(topTablesHeight)

		// The width of the Commodity COLUMN.
		cmdtyColumnWidth := int(float64(topTablesHeight) * .66)
		// the width of the 5d, 1d, and current price column
		priceMovementColumnWidth := topTablesWidth - cmdtyColumnWidth
		cmdtyTableColumns := []table.Column{
			{Title: "Commodity", Width: cmdtyColumnWidth},
			{Title: "1D", Width: int(float64(priceMovementColumnWidth) * .33)},
			{Title: "5D", Width: int(float64(priceMovementColumnWidth) * .33)},
			{Title: "Price", Width: int(float64(priceMovementColumnWidth) * .33)},
		}

		d.tables[0].SetColumns(cmdtyTableColumns)

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// unfocus currently focused table
			d.tables[d.focused].Blur()
			if d.focused < len(d.tables)-1 {
				d.tables[d.focused].SetStyles(d.unfocusedStyle.innerStyle)
				d.focused += 1
			} else {
				d.tables[d.focused].SetStyles(d.focusedStyle.innerStyle)
				d.focused = 0
			}
			d.tables[d.focused].Focus()
			d.tables[d.focused].SetStyles(d.focusedStyle.innerStyle)

			log.Infof("Focusing on table %v", d.focused)
		}
	}
	return d, nil
}

func (d *Dashboard) View() string {
	newsStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder())
	return lipgloss.JoinHorizontal(0,
		newsStyle.Render(d.tables[0].View()),
		newsStyle.Render(d.tables[1].View()),
	)

}

type EconomicCalendar struct {
}

func (cal EconomicCalendar) Init() tea.Cmd {
	return nil
}

func (cal EconomicCalendar) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return cal, nil

}

func (cal EconomicCalendar) View() string {
	return "Economic Calendar"
}
