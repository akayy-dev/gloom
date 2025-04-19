package scraping

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	feed "github.com/mmcdole/gofeed"
)

type NewsUpdate []NewsArticle

type NewsArticle struct {
	Title           string
	PublicationDate time.Time
	URL             string
	Source          string
	Readable        bool
	Content         string
}

func GetAllNews() tea.Msg {
	news := append(GetYahooNews(), GetTENews()...)

	sort.Slice(news, func(i, j int) bool {
		return news[i].PublicationDate.After(news[j].PublicationDate)
	})

	return NewsUpdate(news)
}

func GetYahooNews() []NewsArticle {
	// NOTE: This technically could get news from any RSS url.
	RSS_URL := "https://news.yahoo.com/rss/finance"

	fp := feed.Parser{}

	feed, err := fp.ParseURL(RSS_URL)

	if err != nil {
		log.Error("Failed to get Yahoo News data", "error: ", err)
	}

	var articles []NewsArticle
	for _, item := range feed.Items {
		article := NewsArticle{
			Title:           item.Title,
			Source:          "Yahoo Finance",
			Readable:        false,
			PublicationDate: *item.PublishedParsed,
			URL:             item.Link,
		}

		articles = append(articles, article)
	}

	return articles
}

type TENewsJSON []struct {
	ID          int         `json:"ID"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	URL         string      `json:"url"`
	Author      string      `json:"author"`
	Country     string      `json:"country"`
	Category    string      `json:"category"`
	Image       interface{} `json:"image"`
	Importance  int         `json:"importance"`
	Date        string      `json:"date"`
	Expiration  string      `json:"expiration"`
	HTML        interface{} `json:"html"`
	Type        interface{} `json:"type"`
}

func GetTENews() []NewsArticle {
	URL := "https://tradingeconomics.com/ws/stream.ashx?start=0&size=20"

	client := &http.Client{}

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "en-US,en;q=0.6")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Brave/135.0.0.0 Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://tradingeconomics.com/stream")
	// req.Header.Set("Cookie", `ASP.NET_SessionId=2mqytqet5kkj3v0vtyuolosk; __zlcmid=1R7nX6Tt1VX2J4Z; .ASPXAUTH=4699F73B28B7F163F6E449E54626AD57C2177E3BD023765231FF5520D50F11BE9E9ACFE5968AA936E8AE84880EEE824B358DE03709CFBFCCF6A13A66A7FA3850F5C1701EFAA818563F0B74DEF7920109A02D9B756EA30E6D6307CC51BE83A74346; TEUsername=fKkcBEv4cCREdCkbhMfMnJoCJ6JB3rFiKy5MsLxl/==; TENickName=AhaduKebede; TEUserInfo=21ecaf01-8c32-46a6-aa3a-997c6b8d8dc6; TEName=Ahadu Kebede; TEUserEmail=ahadukebede@gmail.com; cal-timezone-offset=-240; TEServer=TEIIS2`)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(gz)

	if err != nil {
		panic(err)
	}

	var news TENewsJSON
	json.Unmarshal(body, &news)

	var articleData []NewsArticle

	for _, n := range news {

		// TODO: Timestamps seem to be 4 hours ahead of EST, write code to account for this discrepancy
		strippedTime := strings.Split(n.Date, ".")[0] // get rid of millisecond data, useless and causes errors
		parsedTime, err := time.Parse("2006-01-02T15:04:05", strippedTime)

		if err != nil {
			// panic(err)
			log.Error("Cannot parse datetime for TradingEconomics article", "error: ", err)
			return articleData
		}

		article := NewsArticle{
			Title:           n.Title,
			Source:          "TE",
			Readable:        true,
			Content:         n.Description,
			PublicationDate: parsedTime,
		}

		articleData = append(articleData, article)
	}

	return articleData

}
