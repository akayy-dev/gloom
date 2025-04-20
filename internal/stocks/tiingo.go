package stocks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type TiingoIEXTicker []struct {
	Ticker            string    `json:"ticker"`
	Timestamp         time.Time `json:"timestamp"`
	LastSaleTimestamp any       `json:"lastSaleTimestamp"`
	QuoteTimestamp    any       `json:"quoteTimestamp"`
	Open              float64   `json:"open"`
	High              float64   `json:"high"`
	Low               float64   `json:"low"`
	Mid               any       `json:"mid"`
	TngoLast          float64   `json:"tngoLast"`
	Last              any       `json:"last"`
	LastSize          any       `json:"lastSize"`
	BidSize           any       `json:"bidSize"`
	BidPrice          any       `json:"bidPrice"`
	AskPrice          any       `json:"askPrice"`
	AskSize           any       `json:"askSize"`
	Volume            int       `json:"volume"`
	PrevClose         float64   `json:"prevClose"`
}

type OHLCVTickerUpdateMsg TiingoIEXTicker

func GetCurrentOHLCV(symbols []string) tea.Msg {
	builder := strings.Builder{}
	builder.WriteString("https://api.tiingo.com/iex/")

	for i, symbol := range symbols {
		builder.WriteString(symbol)
		if i < len(symbols) {
			builder.WriteString(",")
		}
	}

	builder.WriteString(fmt.Sprintf("?token=%s", os.Getenv("TIINGO_KEY")))
	endpoint := builder.String()
	log.Infof("Getting stock data from endpoint %s", endpoint)
	client := &http.Client{}
	request, err := http.NewRequest("GET", endpoint, nil)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data *TiingoIEXTicker
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
	}

	return OHLCVTickerUpdateMsg(*data)
}
