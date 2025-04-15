package scraping

import (
	"sort"
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
	Content         string
}

func GetYahooNews() tea.Msg {
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
			PublicationDate: *item.PublishedParsed,
			URL:             item.Link,
		}

		articles = append(articles, article)
	}

	sort.Slice(articles, func(i, j int) bool {
		return articles[i].PublicationDate.After(articles[j].PublicationDate)
	})

	return NewsUpdate(articles)
}
