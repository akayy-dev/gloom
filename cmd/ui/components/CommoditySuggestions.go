package components

import (
	"encoding/json"
	"fmt"
	"gloomberg/internal/shared"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Suggestion struct {
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Currency          string `json:"currency"`
	StockExchange     string `json:"stockExchange"`
	ExchangeShortName string `json:"exchangeShortName"`
}

func (e Suggestion) Title() string {
	return e.Name
}

func (e Suggestion) Description() string {
	return fmt.Sprintf("%s - %s", e.Symbol, e.StockExchange)
}

func (e Suggestion) FilterValue() string {
	return fmt.Sprintf("%s %s", e.Name, e.StockExchange)
}

// Return all stocks accessible by FMP.
func GetStockSuggestions(symbol string) []Suggestion {
	// NOTE: QueryEscape formats characters like spaces so the request doesn't break.
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/search?query=%s&apikey=%s", url.QueryEscape(symbol), os.Getenv("FMP_KEY"))

	client := http.Client{Timeout: 15 * time.Second} // The api returns a pretty big set of data, so it's best if we have it time out.
	resp, err := client.Get(url)

	if err != nil {
		shared.UserLog.Fatal("Fatal error ocurred while requesting listed stocks from GetStockSuggestions()", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		shared.UserLog.Fatal("Fatal error occurred while reading body in GetStockSuggestions()", err)
	}

	var list []Suggestion
	err = json.Unmarshal(body, &list)

	if err != nil {
		shared.UserLog.Fatalf("Fatal error occurred while Unmarshaling StockList in GetStockSuggestions() err: %s, JSON %b", err, body)
	}

	return list
}

type CommoditySuggestions struct {
	Symbols     []Suggestion
	SearchQuery string
	List        list.Model
	Width       int
	Height      int
}

func (s *CommoditySuggestions) Init() tea.Cmd {
	s.Symbols = GetStockSuggestions(s.SearchQuery)

	// Convert the symbols list to a list of item interfaces (typejack)
	items := make([]list.Item, len(s.Symbols))
	for i, symbol := range s.Symbols {
		items[i] = symbol
	}

	/*
		TODO: Create a custom delegate,
			it looks like it can simplify calling a function
			on selection.
	*/
	delegate := list.NewDefaultDelegate()

	// Change the styling of the currently selecte
	accentColor := lipgloss.Color(shared.Koanf.String("theme.accentColor"))
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(accentColor).BorderForeground(accentColor)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(accentColor).BorderForeground(accentColor)

	s.List = list.New(items, delegate, s.Width, s.Height)
	s.List.Title = "Select a stock"
	s.List.SetShowHelp(false)
	s.List.SetFilteringEnabled(true)

	return nil
}

func (s *CommoditySuggestions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// BUG: Pressing q quits, should look at main.go
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.Width = msg.Width / 2
		s.Height = int(float64(msg.Height) * .8)
	case tea.KeyMsg:
		switch key := msg.String(); key {
		case "esc":
			// only close the overlay if the user isn't currently searching
			if !s.List.SettingFilter() {
				log.Info("Closing suggestions")
				return s, func() tea.Msg { return shared.ModalCloseMsg(true) }
			}
		}
	}

	s.List.SetWidth(s.Width)
	s.List.SetHeight(s.Height)
	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}

func (s *CommoditySuggestions) View() string {
	titleStyle := shared.Renderer.NewStyle().Bold(true).Foreground(lipgloss.Color(shared.Koanf.String("theme.accentColor")))
	listStyle := shared.Renderer.NewStyle().Border(lipgloss.RoundedBorder()).Width(s.Width).Height(s.Height)
	s.List.Styles.Title = titleStyle
	s.List.Styles.ActivePaginationDot = shared.Renderer.NewStyle().Foreground(lipgloss.Color(shared.Koanf.String("theme.accentColor")))
	return listStyle.Render(s.List.View())
}

func (s CommoditySuggestions) GetKeys() []key.Binding {
	keys := s.List.KeyMap
	if s.List.SettingFilter() {
		escape := key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("<esc>", "Cancel filter"),
		)
		return []key.Binding{keys.CursorUp, keys.CursorDown, escape}
	} else {
		return []key.Binding{keys.CursorUp, keys.CursorDown, keys.Filter, keys.NextPage, keys.PrevPage}
	}
}
