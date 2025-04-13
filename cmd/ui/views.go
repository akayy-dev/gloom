package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	tea "github.com/charmbracelet/bubbletea"
)

type Commodity struct {
	Name           string
	Price          float64
	OneDayMovement float64
}

type CommodityUpdateMsg []Commodity

// What the user opens to, should have general information on the market.
type Dashboard struct {
	name string
	// screen height
	height int
	// screen width
	width int
	// List of tables, tables[0] is commodities, tables[1] is stocks, tables[3] is news
	tables  []table.Model
	focused int
}

func (d *Dashboard) Init() tea.Cmd {
	// Cmdty and stock table width

	cmdtyTable := table.New()

	stockTable := table.New()

	stockTable.Columns()
	d.tables = append(d.tables, cmdtyTable, stockTable)
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
				d.focused += 1
			} else {
				d.focused = 0
			}
			d.tables[d.focused].Focus()

			log.Infof("Focusing on table %v", d.focused)
		}
	}
	return d, nil
}

func (d *Dashboard) View() string {
	newsStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder())
	return lipgloss.JoinHorizontal(0, newsStyle.Render(d.tables[0].View()), newsStyle.Render(d.tables[1].View()))

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
