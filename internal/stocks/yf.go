package stocks

import (
	"gloomberg/internal/shared"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
)

type OHLCVTickerUpdateMsg []finance.Equity

func GetCurrentOHLCV(symbols []string) tea.Msg {
	tickers := []finance.Equity{}

	for _, ticker := range symbols {
		q, err := equity.Get(ticker)
		if err != nil {
			shared.UserLog.Fatalf("Error ocurre while getting equity: %v", err)
		}
		tickers = append(tickers, *q)
	}

	return OHLCVTickerUpdateMsg(tickers)
}
