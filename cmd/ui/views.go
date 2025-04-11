package main

import (
	"fmt"
	"gloomberg/internal"

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
	name       string
	cmdtyTable table.Model
	newsTable  table.Model
	height     int
	width      int
	tables     []table.Model
	focused    int
}

func (d Dashboard) Init() tea.Cmd {
	cmdtyData := []Commodity{
		{
			Name:           "GOLD/USD",
			Price:          250,
			OneDayMovement: 5,
		},
		{
			Name:           "SILVER/USD",
			Price:          30,
			OneDayMovement: 1.2,
		},
		{
			Name:           "OIL/USD",
			Price:          85,
			OneDayMovement: -0.5,
		},
	}

	return tea.Batch(func() tea.Msg {
		log.Info("Init command executing")
		return CommodityUpdateMsg(cmdtyData)
	}, internal.GetNewsData())
}

func (d Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Info("Received message", "type", fmt.Sprintf("%T", msg))
	var returnCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if d.newsTable.Focused() {
				log.Info("Blurring")
				d.newsTable.Blur()
			} else {
				log.Info("Focusing")
				d.newsTable.Focus()
			}
		}
	case CommodityUpdateMsg:
		log.Info("Commodity update")
		cmdtyRows := []table.Row{}
		for _, cmdty := range msg {
			// add rowdata to cmdtyRows here.
			row := table.Row{
				cmdty.Name,
				fmt.Sprintf("%.2f", cmdty.Price),
				fmt.Sprintf("%.2f%%", cmdty.OneDayMovement),
			}
			cmdtyRows = append(cmdtyRows, row)
		}

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderForeground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#703FFD")).
			BorderBottom(true).
			Bold(true)

		d.cmdtyTable = table.New(
			table.WithRows(cmdtyRows),
			table.WithColumns([]table.Column{
				{Title: "Commodity", Width: int(float64(d.width) * 0.75)},
				{Title: "Price", Width: int(float64(d.width) * .1)},
				{Title: "1D", Width: int(float64(d.width) * .1)},
			}),
			table.WithStyles(s),
			table.WithFocused(true),
		)
	case tea.WindowSizeMsg:
		d.height = msg.Height
		d.width = msg.Width

		log.Infof("Width set to %d, Height set to %d", d.height, d.width)

	case internal.NewsMsg:
		var rows []table.Row
		for _, article := range msg.Feed {
			rows = append(rows, table.Row{
				article.Title,
				article.TimePublished,
			})
		}

		columns := []table.Column{
			{Title: "Headline", Width: int(float64(d.width) * 0.8)},
			{Title: "Time", Width: int(float64(d.width) * 0.15)},
		}

		s := table.DefaultStyles()

		s.Header = s.Header.
			Background(lipgloss.Color("#703FFD")).
			BorderBottom(true).
			Bold(true)

		d.newsTable = table.New(
			table.WithRows(rows),
			table.WithColumns(columns),
			table.WithStyles(s),
			table.WithFocused(false),
		)
	}

	d.cmdtyTable, _ = d.cmdtyTable.Update(msg)
	d.newsTable, _ = d.newsTable.Update(msg)
	return d, returnCmd
}

func (d Dashboard) View() string {
	cmtdyStyle := lipgloss.NewStyle().
		Align(lipgloss.Top, lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		Render(d.cmdtyTable.View())
	newsStyle := lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Bottom).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#FFFFFF")).
		Render(d.newsTable.View())
	return lipgloss.JoinVertical(0, d.name, cmtdyStyle, newsStyle)
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
