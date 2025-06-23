package utils

import (
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/equity"
)

func GetCurrentOHLCV(symbol string) (finance.Equity, error) {
	q, err := equity.Get(symbol)
	if err != nil {
		return *q, err
	}
	return *q, nil
}
