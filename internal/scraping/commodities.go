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
	// How many times we have retried to get the data
	retries := 0
	// Maximum amount of retries allowed
	maxRetries := 3

	var cmdtyData []Commodity
	// practice reading a news article
	URL := "https://tradingeconomics.com/commodities"
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.6312.86 Safari/537.36"),
		colly.AllowURLRevisit(),
		colly.AllowedDomains("tradingeconomics.com"),
	)

	c.OnRequest(func(r *colly.Request) {
		log.Infof("Getting commodities from %s", r.URL)
	})

	// Rate Limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	})

	c.OnError(func(response *colly.Response, err error) {

		if retries < maxRetries {
			retries++
			log.Warnf("Retry %d/%d: %v", retries, maxRetries, err)
			response.Request.Retry()
		} else {
			log.Error("Something went wrong: ", err)
		}
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
