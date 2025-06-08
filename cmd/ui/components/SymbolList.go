package components

import (
	"encoding/json"
	"fmt"
	"gloomberg/internal/shared"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	listStyle = shared.Renderer.NewStyle().Border(lipgloss.RoundedBorder())
)

type StockListEntry struct {
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Currency          string `json:"currency"`
	StockExchange     string `json:"stockExchange"`
	ExchangeShortName string `json:"exchangeShortName"`
}

func (e StockListEntry) Title() string {
	return e.Name
}

func (e StockListEntry) Description() string {
	return fmt.Sprintf("%s - %s", e.Symbol, e.StockExchange)
}

func (e StockListEntry) FilterValue() string {
	return e.Title()
}

// Return all stocks accessible by FMP.
func GetStockSuggestions(symbol string) []StockListEntry {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/search?query=%s&apikey=%s", symbol, os.Getenv("FMP_KEY"))

	client := http.Client{Timeout: 15 * time.Second} // The api returns a pretty big set of data, so it's best if we have it time out.
	resp, err := client.Get(url)

	if err != nil {
		shared.UserLog.Fatal("Fatal error ocurred while requesting listed stocks from GetAllListedStocks()", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		shared.UserLog.Fatal("Fatal error occurred while reading body in GetAllListedStocks()", err)
	}

	var list []StockListEntry
	err = json.Unmarshal(body, &list)

	if err != nil {
		shared.UserLog.Fatal("Fatal error occurred while Unmarshaling StockList in GetAllListedStocks()", err)
	}

	shared.UserLog.Info("Got the following stock suggestions")
	for _, stock := range list {
		shared.UserLog.Info(stock.Name)
	}

	return list
}

type StockSuggestions struct {
	Symbols     []StockListEntry
	SearchQuery string
	List        list.Model
	Width       int
	Height      int
}

func (s *StockSuggestions) Init() tea.Cmd {
	s.Symbols = GetStockSuggestions(s.SearchQuery)

	// Convert the symbols list to a list of item interfaces (typejack)
	items := make([]list.Item, len(s.Symbols))
	for i, symbol := range s.Symbols {
		items[i] = symbol
	}

	s.List = list.New(items, list.NewDefaultDelegate(), s.Width, s.Height)
	// BUG: The width does not set until a filtering, even with this extra call.
	log.Info("Width", s.Width)
	s.List.SetWidth(s.Width)
	s.List.Title = "Select a stock"
	s.List.SetFilteringEnabled(true)

	return nil
}

func (s *StockSuggestions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// BUG: Pressing q quits, should look at main.go
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch key := msg.String(); key {

		}
	}
	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}

func (s *StockSuggestions) View() string {
	return listStyle.Render(s.List.View())
}
