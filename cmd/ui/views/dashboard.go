package views

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"gloomberg/cmd/ui/components"
	"gloomberg/internal/scraping"
	"gloomberg/internal/utils"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

func commodityUpdateTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return scraping.GetCommodities()
	})
}

// Update the stock prices every 5 seconds.
func stockUpdateTick(symbols []string) tea.Cmd {
	utils.UserLog.Info("stockUpdateTick")
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		var rows []RowData
		for _, symbol := range symbols {
			q, err := utils.GetCurrentOHLCV(symbol)
			if err != nil {
				utils.UserLog.Errorf("Error fetching data for %s: %v", symbol, err)
				continue
			}
			row := RowData{
				CompanyName:   q.ShortName,
				Symbol:        q.Symbol,
				Price:         q.RegularMarketPrice,
				PercentChange: q.RegularMarketChange,
				SMA:           q.Quote.FiftyDayAverage,
			}
			rows = append(rows, row)
		}
		return WatchlistUpdateMsg{Rows: rows, Refresh: true}
	})
}

func (d *Dashboard) GetWatchList(refresh bool) tea.Msg {
	var rows []RowData
	for _, symbol := range d.WatchList {
		utils.UserLog.Infof("Getting data for %s", symbol)
		q, err := utils.GetCurrentOHLCV(symbol)
		if err != nil {
			utils.UserLog.Errorf("Error fetching data for %s: %v", symbol, err)
			continue
		}
		row := RowData{
			CompanyName:   q.ShortName,
			Symbol:        q.Symbol,
			Price:         q.RegularMarketPrice,
			PercentChange: q.RegularMarketChange,
			SMA:           q.Quote.FiftyDayAverage,
		}
		rows = append(rows, row)
	}
	return WatchlistUpdateMsg{Rows: rows, Refresh: refresh}
}

type WatchlistUpdateMsg struct {
	Rows []RowData
	// Should recieving this WatchlistUpdateMsg
	// send another WatchlistUpdateMsg in a couple of seconds?
	Refresh bool
}

type RowData struct {
	CompanyName   string
	Symbol        string
	Price         float64
	PercentChange float64
	SMA           float64
}

// Return a table.Row for the stock table to use
func (d RowData) Render() table.Row {
	var color string
	if d.PercentChange >= 0 {
		color = "\033[38;5;46m" // green
	} else {
		color = "\033[38;5;196m" // red
	}
	return table.Row{
		fmt.Sprintf("%s%s (%s)", color, d.CompanyName, d.Symbol),
		fmt.Sprintf("$%.2f", d.SMA),
		fmt.Sprintf("$%.2f", d.Price),
		// NOTE: Adding return-to-normal escape code (\033[0m) breaks table width, doesn't matter though,
		// seems lipgloss can handle it
		fmt.Sprintf("%.2f", d.PercentChange),
	}
}

type DisplayOverlayMsg tea.Model

type TableStyle struct {
	innerStyle table.Styles
	outerStyle lipgloss.Style
}

type Dashboard struct {
	Name string
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
	// map the row in the table to an actual news article
	articleMap map[int]scraping.NewsArticle

	// Stock watchlist
	WatchList []string
}

func (d *Dashboard) Init() tea.Cmd {
	// make article map
	d.articleMap = make(map[int]scraping.NewsArticle)

	cmdtyTable := table.New(
		table.WithFocused(false),
	)
	stockTable := table.New(table.WithFocused(false))

	newsTable := table.New(table.WithFocused(false))

	accentColor := utils.Koanf.String("theme.accentColor")

	foucsedInnerStyle := table.Styles{
		Header: utils.Renderer.NewStyle().
			Align(lipgloss.Center).
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")),
		Cell:     utils.Renderer.NewStyle(),
		Selected: utils.Renderer.NewStyle().Bold(true).Foreground(lipgloss.Color(accentColor)),
	}

	unfocusedInnerStyle := table.Styles{
		Header: utils.Renderer.NewStyle().
			BorderForeground(lipgloss.Color(accentColor)).
			Bold(false),
		Cell:     utils.Renderer.NewStyle(),
		Selected: utils.Renderer.NewStyle().Bold(true).Foreground(lipgloss.Color(accentColor)),
	}

	d.focusedStyle = TableStyle{
		innerStyle: foucsedInnerStyle,
		outerStyle: utils.Renderer.NewStyle().
			BorderForeground(lipgloss.Color(accentColor)).
			Border(lipgloss.NormalBorder()),
	}

	d.unfocusedStyle = TableStyle{
		innerStyle: unfocusedInnerStyle,
		outerStyle: utils.Renderer.NewStyle().BorderForeground(lipgloss.Color("#FFFFFF")).Border(lipgloss.NormalBorder()),
	}

	d.tables = append(d.tables, cmdtyTable, stockTable, newsTable)

	// keeps other tables from having boldface columns by default
	for i, _ := range d.tables {
		d.tables[i].SetStyles(d.unfocusedStyle.innerStyle)
	}

	d.tables[0].Focus()

	d.WatchList = utils.Koanf.Strings("dashboard.tickers")
	return tea.Batch(scraping.GetCommodities,
		scraping.GetAllNews,
		func() tea.Msg { return commodityUpdateTick() },
		func() tea.Msg { return d.GetWatchList(true) },
	)
}

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height - 1

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
			{Title: "SMA (50d)", Width: int(float64(topTablesWidth) * 2 / 10)},
			{Title: "Price", Width: int(float64(topTablesWidth) * 2 / 10)},
			{Title: "%", Width: int(float64(topTablesWidth) * 1 / 10)},
		}

		d.tables[1].SetColumns(stockColumns)

		newsTableWidth := topTablesWidth * 2
		newsColumns := []table.Column{
			{Title: "Headline", Width: int(math.Ceil(float64(newsTableWidth) * .75))},
			{Title: "Source", Width: int(math.Ceil(float64(newsTableWidth) * .125))},
			{Title: "Date", Width: int(math.Ceil(float64(newsTableWidth) * .125))},

			{Title: "index", Width: 0},
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

			utils.UserLog.Infof("Focusing on table %v", d.focused)
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

			utils.UserLog.Infof("Focusing on table %v", d.focused)

		case "enter":
			utils.UserLog.Info("enter pressed")

			switch d.focused {
			// different actions depending on which table is focused
			case 2: // news table
				rowID, err := strconv.Atoi(d.tables[2].SelectedRow()[3]) // index of the article in the articleMap
				if err != nil {
					utils.UserLog.Fatal(err)
				}
				selectedStory := d.articleMap[rowID]
				newsOverlay := components.NewsModal{
					Article: &selectedStory,
					W:       d.width / 2,
					H:       int(float64(d.height) * .8),
				}
				return d, func() tea.Msg { return (&newsOverlay) }

			}
		case "a":
			// add symbol on stock table
			if d.focused == 1 {
				return d, func() tea.Msg {
					return utils.PromptOpenMsg{
						Prompt: "Search for a stock: ",
						CallbackFunc: func(s string) tea.Msg {
							// TODO: Create an overlay for the current search query.
							stocklist := components.CommoditySuggestions{
								SearchQuery: s,
								Width:       d.width / 2,
								Height:      int(float64(d.height) * .8),
								CallbackFunc: func(s components.Suggestion) tea.Msg {
									// TODO: Create a function that can add to the watchlist
									// and refresh the display, instead of waiting for the next
									// update message.
									d.WatchList = append(d.WatchList, s.Symbol)
									utils.Program.Send(d.GetWatchList(false))
									return utils.SendNotificationMsg{
										Message:     fmt.Sprintf("Adding $%s to watchlist", s.Symbol),
										DisplayTime: 3000,
									}
								},
							}
							return DisplayOverlayMsg(&stocklist)
						},
					}
				}
			}
		}
		d.tables[d.focused], cmd = d.tables[d.focused].Update(msg)

	case scraping.CommodityUpdateMsg:
		utils.UserLog.Info("Commodity Data Recieved")
		rows := []table.Row{}
		for _, cmdty := range msg {
			var color string
			if cmdty.OneDayMovement >= 0 {
				color = "\033[38;5;46m" // green
			} else {
				color = "\033[38;5;196m" // red
			}

			rows = append(rows, table.Row{
				// NOTE: Attempting to add color to other columns results in visual bug.
				fmt.Sprintf("%s%s", color, cmdty.Name), fmt.Sprintf("%.2f%%", cmdty.OneDayMovement), fmt.Sprintf("%.2f%%", cmdty.WeeklyMovement), fmt.Sprintf("$%.2f", cmdty.Price),
			})
		}
		d.tables[0].SetRows(rows)
		utils.UserLog.Info("Got commodity data")
		return d, commodityUpdateTick()

	case scraping.NewsUpdate:
		utils.UserLog.Info("Got news update")

		rows := []table.Row{}

		for i, article := range msg {
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

			newsRow := table.Row{
				flaggedTitle,
				article.Source,
				formattedTime,
				strconv.Itoa(i), // index in articleMap (as a string)
			}

			d.articleMap[i] = article

			rows = append(rows, newsRow)
		}
		d.tables[2].SetRows(rows)

	case WatchlistUpdateMsg:
		utils.UserLog.Info("Got stock data (WatchlistUpdateMsg)")
		var tableRows []table.Row
		for _, row := range msg.Rows {
			tableRows = append(tableRows, row.Render())
			utils.UserLog.Infof("Adding row for %s", row.Symbol)
		}
		d.tables[1].SetRows(tableRows)
		return d, stockUpdateTick(d.WatchList)
	}

	return d, cmd
}

func (d *Dashboard) GetKeys() []key.Binding { // TODO: Change to have actual type safety
	keyList := []key.Binding{
		key.NewBinding(
			key.WithHelp("<tab>", "Switch Focus"),
		),
		key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "Move up"),
		),
		key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "Move down"),
		),
	}

	// FIXME: This does not work, I'm assuming I have to send an Update 🙄.
	// There should be vue-type reactive data
	if d.focused == 1 {
		keyList = append(keyList, key.NewBinding(
			key.WithHelp("a", "Add Stock"),
			key.WithKeys("a", "add"),
		))
	}
	if d.focused == 3 {
		keyList = append(keyList, key.NewBinding(
			key.WithHelp("<enter>", "Read article"),
			key.WithKeys("enter", "select"),
		))

	}

	return keyList
}

func (d *Dashboard) View() string {
	accentColor := utils.Koanf.String("theme.accentColor")

	foucsedBorder := utils.Renderer.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(accentColor))
	unfocusedBorder := utils.Renderer.NewStyle().Border(lipgloss.NormalBorder())

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
