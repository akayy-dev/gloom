package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"gloomberg/internal/scraping"
	"gloomberg/internal/stocks"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

type DisplayOverlayMsg tea.Model

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
	cmdtyTable := table.New(table.WithFocused(false))

	stockTable := table.New(table.WithFocused(false))

	newsTable := table.New(table.WithFocused(false))

	foucsedInnerStyle := table.Styles{
		Header: Renderer.NewStyle().
			Align(lipgloss.Center).
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")),
		Cell:     Renderer.NewStyle(),
		Selected: Renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")),
	}

	unfocusedInnerStyle := table.Styles{
		Header: Renderer.NewStyle().
			BorderForeground(lipgloss.Color("#703FFD")).
			Bold(false),
		Cell:     Renderer.NewStyle(),
		Selected: Renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFF")),
	}

	d.focusedStyle = TableStyle{
		innerStyle: foucsedInnerStyle,
		outerStyle: Renderer.NewStyle().BorderForeground(lipgloss.Color("#703FFD")).Border(lipgloss.NormalBorder()),
	}

	d.unfocusedStyle = TableStyle{
		innerStyle: unfocusedInnerStyle,
		outerStyle: Renderer.NewStyle().BorderForeground(lipgloss.Color("#FFFFFF")).Border(lipgloss.NormalBorder()),
	}

	d.tables = append(d.tables, cmdtyTable, stockTable, newsTable)

	// keeps other tables from having boldface columns by default
	for i, _ := range d.tables {
		d.tables[i].SetStyles(d.unfocusedStyle.innerStyle)
	}

	d.tables[0].Focus()

	var watchlist []string
	watchlist = append(watchlist, "SPY", "FEZ", "AAPL")
	return tea.Batch(scraping.GetCommodities,
		scraping.GetAllNews,
		func() tea.Msg { return scraping.CommodityUpdateTick() },
		// TODO: Find a way to dynamically get the tickers to search, perhaps a config file or
		// database entry for user preferences?
		func() tea.Msg { return stocks.GetCurrentOHLCV(watchlist) },
	)
}

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

		// Redraw tables

		// NOTE: For some reason using exactly 1/2 the width and 2/3 the screen
		// draws the border past it's boundaries. whatever make it slightly less
		topTablesWidth := int(float64(d.width) * .49)
		topTablesHeight := int(float64(d.height) * .65)

		d.tables[0].SetWidth(topTablesWidth)
		d.tables[1].SetWidth(topTablesWidth)

		d.tables[0].SetHeight(topTablesHeight)
		d.tables[1].SetHeight(topTablesHeight)

		// The width of the Commodity COLUMN.
		cmdtyColumnWidth := int(float64(topTablesWidth) * 1 / 2)
		// the width of the 5d, 1d, and current price column
		priceMovementColumnWidth := topTablesWidth - cmdtyColumnWidth
		cmdtyTableColumns := []table.Column{
			{Title: "Commodity", Width: cmdtyColumnWidth},
			{Title: "1D", Width: int(float64(priceMovementColumnWidth) * 1 / 3)},
			{Title: "7D", Width: int(float64(priceMovementColumnWidth) * 1 / 3)},
			{Title: "Price", Width: int(float64(priceMovementColumnWidth) * 1 / 3)},
		}

		d.tables[0].SetColumns(cmdtyTableColumns)

		stockColumns := []table.Column{
			{Title: "Symbol", Width: int(float64(topTablesWidth) * 1 / 2)},
			{Title: "7D", Width: int(float64(priceMovementColumnWidth) * 1 / 6)},
			{Title: "1D", Width: int(float64(topTablesWidth) * 1 / 6)},
			{Title: "Price", Width: int(float64(topTablesWidth) * 1 / 6)},
		}

		d.tables[1].SetColumns(stockColumns)

		newsTableWidth := topTablesWidth * 2
		newsColumns := []table.Column{
			{Title: "Headline", Width: int(math.Ceil(float64(newsTableWidth) * .75))},
			{Title: "Source", Width: int(math.Ceil(float64(newsTableWidth) * .125))},
			{Title: "Date", Width: int(math.Ceil(float64(newsTableWidth) * .125))},

			// Data columns, don't actually show on the table, that's why their width is zero
			{Title: "Readable", Width: 0}, // empty if the article cannot be read.
			{Title: "Title", Width: 0},
			{Title: "Content", Width: 0},
			{Title: "URL", Width: 0},
			{Title: "Source", Width: 0},
		}

		d.tables[2].SetColumns(newsColumns)
		d.tables[2].SetWidth(newsTableWidth)
		d.tables[2].SetHeight(d.height - topTablesHeight - 5)

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// unfocus currently focused table
			// BUG: Can't fit everything into table
			d.tables[d.focused].Blur()
			if d.focused < len(d.tables)-1 {
				d.tables[d.focused].SetStyles(d.unfocusedStyle.innerStyle)
				d.focused += 1
			} else {
				d.tables[d.focused].SetStyles(d.unfocusedStyle.innerStyle)
				d.focused = 0
			}
			d.tables[d.focused].Focus()
			d.tables[d.focused].SetStyles(d.focusedStyle.innerStyle)

			UserLog.Infof("Focusing on table %v", d.focused)
		case "shift+tab":
			d.tables[d.focused].Blur()
			if d.focused > 0 {
				d.tables[d.focused].SetStyles(d.unfocusedStyle.innerStyle)
				d.focused -= 1
			} else {
				d.tables[d.focused].SetStyles(d.unfocusedStyle.innerStyle)
				d.focused = len(d.tables) - 1
			}
			d.tables[d.focused].Focus()
			d.tables[d.focused].SetStyles(d.focusedStyle.innerStyle)

			UserLog.Infof("Focusing on table %v", d.focused)

		case "enter":
			UserLog.Info("enter pressed")

			switch d.focused {
			// different actions depending on which table is focused
			case 2:
				selectedStory := d.tables[2].SelectedRow()

				// get rid of nerd font book character
				if selectedStory[3] == "1" {
					content := selectedStory[5]

					// get the story title without the unicode book character in front of it
					titleString := strings.TrimLeft(selectedStory[0], " ")
					UserLog.Info("reading news story", "CONTENT", content)
					newsOverlay := NewsModal{
						title:   string(titleString),
						content: content,
						w:       d.width / 2,
						h:       int(float64(d.height) * .8),
					}
					return d, func() tea.Msg { return (&newsOverlay) }
				}

			}
		}
		d.tables[d.focused], cmd = d.tables[d.focused].Update(msg)

	case scraping.CommodityUpdateMsg:
		UserLog.Info("Commodity Data Recieved")
		rows := []table.Row{}
		for _, cmdty := range msg {
			var color string
			if cmdty.OneDayMovement >= 0 {
				color = "#30FF1E"
			} else {
				color = "#FF211D"
			}

			style := Renderer.NewStyle().Foreground(lipgloss.Color(color))
			rows = append(rows, table.Row{
				// NOTE: Attempting to add color to other columns results in visual bug.
				style.Render(cmdty.Name), fmt.Sprintf("%.2f%%", cmdty.OneDayMovement), fmt.Sprintf("%.2f%%", cmdty.WeeklyMovement), fmt.Sprintf("$%.2f", cmdty.Price),
			})
		}
		d.tables[0].SetRows(rows)
		UserLog.Info("Got commodity data")
		return d, scraping.CommodityUpdateTick()

	case scraping.NewsUpdate:
		UserLog.Info("Got news update")

		rows := []table.Row{}

		for _, article := range msg {
			// Format the publication date
			var formattedTime string

			year, month, day := article.PublicationDate.Date()
			nowYear, nowMonth, nowDay := time.Now().Date()
			if year == nowYear && month == nowMonth && day == nowDay {
				// If the article was published today, format it as HH:MM AM/PM
				formattedTime = article.PublicationDate.Format("03:04 PM")
			} else {
				// Otherwise, format it as MM/DD
				formattedTime = article.PublicationDate.Format("01/02")
			}

			var flaggedTitle string // the title with a flag to show whether or not it's readable
			if article.Readable {
				flaggedTitle = fmt.Sprintf("%s %s", "", article.Title)
			} else {
				flaggedTitle = article.Title
			}

			var readable string // 0 if not, 1 if so

			if article.Readable {
				readable = "1"
			} else {
				readable = "0"
			}

			rows = append(rows, table.Row{
				flaggedTitle,
				article.Source,
				formattedTime,

				// data columns
				readable,
				article.Title,
				article.Content,
				article.URL,
				article.Source,
			})
		}
		d.tables[2].SetRows(rows)

	case stocks.OHLCVTickerUpdateMsg:
		UserLog.Info("Got stock data")
		var rows []table.Row
		for _, row := range msg {
			rows = append(rows, table.Row{
				row.Ticker, "", "", fmt.Sprintf("$%.2f", row.TngoLast),
			})
			UserLog.Infof("Adding row for %s", row.Ticker)
		}
		d.tables[1].SetRows(rows)
	}

	return d, cmd
}

func (d *Dashboard) View() string {
	foucsedBorder := Renderer.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#703FFD"))
	unfocusedBorder := Renderer.NewStyle().Border(lipgloss.NormalBorder())

	var styledTables []string
	for _, t := range d.tables {
		if t.Focused() {
			styledTables = append(styledTables, foucsedBorder.Render(t.View()))
		} else {
			styledTables = append(styledTables, unfocusedBorder.Render(t.View()))
		}
	}

	upperDiv := lipgloss.JoinHorizontal(0,
		styledTables[0], styledTables[1],
	)

	content := lipgloss.JoinVertical(0, upperDiv, styledTables[2])
	return content

}

// Economic Calendar Tab
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
