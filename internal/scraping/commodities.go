package scraping

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/gocolly/colly"
)

type Commodity struct {
	Name           string
	Price          float64
	OneDayMovement float64
	WeeklyMovement float64
}

type CommodityUpdateMsg []Commodity

func CommodityUpdateTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return GetCommodities()
	})
}

func GetCommodities() tea.Msg {
	var cmdtyData []Commodity
	// practice reading a news article
	URL := "https://tradingeconomics.com/commodities"
	c := colly.NewCollector(
		colly.UserAgent("Android"),
		colly.AllowURLRevisit(),
		colly.AllowedDomains("tradingeconomics.com"),
	)

	c.OnRequest(func(r *colly.Request) {
		log.Info("Visiting", r.URL)
	})

	c.OnError(func(response *colly.Response, err error) {
		log.Fatal("Something went wrong: ", err)
	})

	c.OnHTML("tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {

			var cols []string
			row.ForEach("td", func(i int, col *colly.HTMLElement) {
				cols = append(cols, strings.TrimSpace(col.Text))
			})

			cmtdtyName := strings.Split(cols[0], "  ")[0] // get only the commodity name, not the USD/symbol.

			cols[1] = strings.ReplaceAll(cols[1], ",", "") // Remove comma for parsing number to float
			price, err := strconv.ParseFloat(cols[1], 64)

			if err != nil {
				panic(err)
			}

			cols[3] = strings.ReplaceAll(cols[3], "%", "")
			oneDayMovement, err := strconv.ParseFloat(cols[3], 64)

			if err != nil {
				panic(err)
			}

			cols[4] = strings.ReplaceAll(cols[4], "%", "")
			weeklyMovement, err := strconv.ParseFloat(cols[4], 64)

			if err != nil {
				panic(err)
			}

			cmdtyData = append(cmdtyData, Commodity{
				Name:           cmtdtyName,
				Price:          price,
				OneDayMovement: oneDayMovement,
				WeeklyMovement: weeklyMovement,
			})

		})
	})

	c.Visit(URL)

	return CommodityUpdateMsg(cmdtyData)
}
