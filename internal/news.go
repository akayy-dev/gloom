package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/log"

	tea "github.com/charmbracelet/bubbletea"
)

type NewsData struct {
	Items                    string `json:"items"`
	SentimentScoreDefinition string `json:"sentiment_score_definition"`
	RelevanceScoreDefinition string `json:"relevance_score_definition"`
	Feed                     []struct {
		Title                string   `json:"title"`
		URL                  string   `json:"url"`
		TimePublished        string   `json:"time_published"`
		Authors              []string `json:"authors"`
		Summary              string   `json:"summary"`
		BannerImage          string   `json:"banner_image"`
		Source               string   `json:"source"`
		CategoryWithinSource string   `json:"category_within_source"`
		SourceDomain         string   `json:"source_domain"`
		Topics               []struct {
			Topic          string `json:"topic"`
			RelevanceScore string `json:"relevance_score"`
		} `json:"topics"`
		OverallSentimentScore float64 `json:"overall_sentiment_score"`
		OverallSentimentLabel string  `json:"overall_sentiment_label"`
		TickerSentiment       []struct {
			Ticker               string `json:"ticker"`
			RelevanceScore       string `json:"relevance_score"`
			TickerSentimentScore string `json:"ticker_sentiment_score"`
			TickerSentimentLabel string `json:"ticker_sentiment_label"`
		} `json:"ticker_sentiment"`
	} `json:"feed"`
}

type NewsMsg NewsData

func GetNewsData() tea.Cmd {
	URL := fmt.Sprintf("https://www.alphavantage.co/query?function=NEWS_SENTIMENT&topics=economy_macro&finance&apikey=%v", os.Getenv("ALPHA_KEY"))

	resp, err := http.Get(URL)

	if err != nil {
		log.Fatal("Failed to get news")
	}
	defer resp.Body.Close()

	var dataStruct NewsData
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response body")
	}
	json.Unmarshal(body, &dataStruct)

	log.Info("Successfully grabbed JSON data")
	log.Info(dataStruct)

	return func() tea.Msg {
		return NewsMsg(dataStruct)
	}
}
