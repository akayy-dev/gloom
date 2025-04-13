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
	// create and style newstable
	newsTable := CreateNewsTable()
	stockTable := table.New(
		table.WithHeight(d.height),
		table.WithWidth(d.width/3),
	)
	d.tables = append(d.tables, newsTable, stockTable)
	return nil
}

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

		d.tables[0].SetWidth(int(float64(d.width) * 0.65))
		d.tables[1].SetWidth(int(float64(d.width) * 0.32))

		d.tables[0].SetHeight(int(float64(d.height) * 0.65))
		d.tables[1].SetHeight(int(float64(d.height) * 0.65))

	case tea.KeyMsg:
		log.Info(msg.String())
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
			log.Infof("Focusing on table %d", d.focused)
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
