package stocks

import (
	"gloomberg/internal/shared"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
	// "github.com/piquette/finance-go/datetime"
	// "github.com/piquette/finance-go/chart"
)

type OHLCVTickerUpdateMsg []finance.Equity

func GetCurrentOHLCV(symbols []string) tea.Msg {
	tickers := []finance.Equity{}

	for _, ticker := range symbols {
		q, err := equity.Get(ticker)
		if err != nil {
			shared.UserLog.Fatalf("Error ocurre while getting equity (%s): %v", ticker, err)
		}
		tickers = append(tickers, *q)
	}

	return OHLCVTickerUpdateMsg(tickers)
}

// func GetSymbolChart(symbol string, interval datetime.Interval) {
// 	params := &chart.Params{
// 		Symbol: symbol,
// 		Interval: interval,
// 	}
//
// 	iter := chart.Get(params)
// }
