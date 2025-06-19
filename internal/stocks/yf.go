package stocks

import (
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
	// "github.com/piquette/finance-go/datetime"
	// "github.com/piquette/finance-go/chart"
)

type OHLCVTickerUpdateMsg []finance.Equity

func GetCurrentOHLCV(symbol string) (finance.Equity, error) {
	q, err := equity.Get(symbol)
	if err != nil {
		return *q, err
	}
	return *q, nil
}
