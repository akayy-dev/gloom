package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type StockBar struct {
	Symbol        string  `json:"symbol"`
	Date          string  `json:"date"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        int     `json:"volume"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	Vwap          float64 `json:"vwap"`
}

type CompanyProfile struct {
	Symbol            string  `json:"symbol"`
	Price             float64 `json:"price"`
	MarketCap         int64   `json:"marketCap"`
	Beta              float64 `json:"beta"`
	LastDividend      float64 `json:"lastDividend"`
	Range             string  `json:"range"`
	Change            float64 `json:"change"`
	ChangePercentage  float64 `json:"changePercentage"`
	Volume            int     `json:"volume"`
	AverageVolume     int     `json:"averageVolume"`
	CompanyName       string  `json:"companyName"`
	Currency          string  `json:"currency"`
	Cik               string  `json:"cik"`
	Isin              string  `json:"isin"`
	Cusip             string  `json:"cusip"`
	ExchangeFullName  string  `json:"exchangeFullName"`
	Exchange          string  `json:"exchange"`
	Industry          string  `json:"industry"`
	Website           string  `json:"website"`
	Description       string  `json:"description"`
	Ceo               string  `json:"ceo"`
	Sector            string  `json:"sector"`
	Country           string  `json:"country"`
	FullTimeEmployees string  `json:"fullTimeEmployees"`
	Phone             string  `json:"phone"`
	Address           string  `json:"address"`
	City              string  `json:"city"`
	State             string  `json:"state"`
	Zip               string  `json:"zip"`
	Image             string  `json:"image"`
	IpoDate           string  `json:"ipoDate"`
	DefaultImage      bool    `json:"defaultImage"`
	IsEtf             bool    `json:"isEtf"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
	IsAdr             bool    `json:"isAdr"`
	IsFund            bool    `json:"isFund"`
}



func GetStockData(symbol string) ([]StockBar, error) {
	var bars []StockBar
	url := fmt.Sprintf("https://financialmodelingprep.com/stable/historical-price-eod/full?symbol=%s&apikey=%s", url.QueryEscape(symbol), os.Getenv("FMP_KEY"))
	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)

	if err != nil {
		return bars, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &bars)
	if err != nil {
		return bars, err
	}
	return bars, nil
}

func GetCompanyProfile(symbol string) (CompanyProfile, error) {
	var profile CompanyProfile
	url := "https://financialmodelingprep.com/stable/profile?symbol=AAPL&apikey=YOUR_API_KEY"
	response, err := http.Get(url)
	if err != nil {
		return profile, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return profile, err
	}

	err = json.Unmarshal(body, profile)
	if err != nil {
		return profile, err
	}
	return profile, nil
}
