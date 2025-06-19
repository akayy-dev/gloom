package utils

import (
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
)

func GetCurrentOHLCV(symbols []string) ([]finance.Equity, error) {
	tickers := []finance.Equity{}

	for _, ticker := range symbols {
		q, err := equity.Get(ticker)
		if err != nil {
			return tickers, err
		}
		tickers = append(tickers, *q)
	}

	return tickers, nil
}
