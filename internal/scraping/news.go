package scraping

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"gloomberg/internal/shared"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	feed "github.com/mmcdole/gofeed"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Struct that the channel uses to send a status code along with additional information
type StatusUpdate struct {
	// Status Code, negative numbers generally mean an error
	StatusCode int
	// if an error is raised, the error as a string will be put here 
	StatusMessage string
}

type NewsUpdate []NewsArticle

type NewsArticle struct {
	Title           string
	PublicationDate time.Time
	Bullets         []string
	URL             string
	Source          string
	Readable        bool
	Content         string
}

// Response from Gemini when scraping news articles
type GeminiResponse struct {
	Success bool     `json:"success"`
	Bullets []string `json:"bullets"`
	Content string   `json:"content"`
}

// Sanitize json to be properly marsalled
func sanitizeJSON(input []byte) []byte {
	// Replace unescaped " with escaped ones within JSON strings
	re := regexp.MustCompile(`(?m)([^\\])\\([^"\\/bfnrt]|$)`)
	fixed := re.ReplaceAll(input, []byte(`${1}\\\\${2}`))
	return fixed
}

// Use AI to scrape the content off an articles page.
// NOTE: Currently returns a too many requests error on a lot of yahoo finance articles.
// my buest guess as to why this happens is because http.Get is just a curl wrapper, and without
// a proper user agent yahoo blocks requests. the solution to this is to migrate to colly.
func PromptNewsURL(article *NewsArticle, progressChan *chan StatusUpdate, ctx context.Context) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_KEY")))

	if err != nil {
		// NOTE: This does not seem to work, the screen shows all the other
		// status messages then throws up an error with the error code,
		// technically not broken since it shows the error code,
		// I'll worry about it later
		if strings.Contains(err.Error(), "API key not valid") {
			log.Errorf("Invalid API key provided for Gemini Client: %s", err)
			(*progressChan) <- StatusUpdate{
				StatusCode: -1,
				StatusMessage: "Gemini key is not valid, did you set $GEMINI_KEY",
			}
		} else {
			log.Errorf("Error while creating Gemini Client: %s", err)
			(*progressChan) <- StatusUpdate{
				StatusCode: -1,
				StatusMessage: err.Error(),
			}
		}
		return
	}

	defer client.Close()
	model := client.GenerativeModel("gemini-2.0-flash")
	model.ResponseMIMEType = "application/json"

	(*progressChan) <- StatusUpdate{
		StatusCode: 0,
	}

	log.Infof("Requesting content from %s", article.URL)

	htmlReq, err := http.NewRequest("GET", article.URL, nil)
	if err != nil {
		log.Errorf("Error while creating http request: %s", err)
		(*progressChan) <- StatusUpdate{
			StatusCode: -1,
			StatusMessage: err.Error(),
		}
		return
	}

	htmlReq.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
	httpClient := http.Client{Timeout: 10 * time.Second}
	htmlSrc, err := httpClient.Do(htmlReq)
	log.Info("Sent request")
	if err != nil {
		// check for timeout error
		if os.IsTimeout(err) {
			log.Error("HTTP request timed out")
			(*progressChan) <- StatusUpdate{
				StatusCode: -1,
				StatusMessage: "HTTP request timed out",
			}
			return
		} else if errors.Is(err, context.DeadlineExceeded) {
			log.Error("HTTP request context deadline exceeded")
			(*progressChan) <- StatusUpdate{
				StatusCode: -1,
				StatusMessage: "HTTP request context deadline exceeded",
			}
			return
		} else {
			log.Errorf("Error while getting article: %s", err)
			(*progressChan) <- StatusUpdate{
				StatusCode: -1,
				StatusMessage: err.Error(),
			}
			return
		}

	}

	if htmlSrc.ContentLength > 5*1024*1024 { // 5 MB
		(*progressChan) <- StatusUpdate{
			StatusCode: -1,
			StatusMessage: "HTML page too large, cancelling request",
		}
		return
	}

	(*progressChan) <- StatusUpdate{
		StatusCode: 1,
	}

	log.Info("Request was successful")
	defer htmlSrc.Body.Close()

	log.Info("Reading bytes from article")
	htmlBytes, err := io.ReadAll(htmlSrc.Body)
	if err != nil {
		log.Errorf("Error encountered while reading HTML content: %s", err)
		(*progressChan) <- StatusUpdate{
			StatusCode: 1,
			StatusMessage: err.Error(),
		}
		return
	}

	(*progressChan) <- StatusUpdate{
		StatusCode: 2,
	}

	// start the gemini request
	req := []genai.Part{
		genai.Blob{MIMEType: "text/html", Data: htmlBytes},
		genai.Text(`
		You are a helpful AI assistant for webscraping.
		I will send you the HTML content of an news website, your job is to convert the article from HTML to markdown.
		Make sure you ONLY format the article, do not format the advertisements on the page or any of the article suggestions.
		 Also please do not include the metadata in your article like the title, time of publication, or author.
		Formatting should not just copy the text, but make use of the multitude of features that markdown offers,
		including matching <h1>-<h6> tags with their appropriate heading in markdown,
		along with rendering lists and tables, as well as anything else that can be properly represented in markdown.
		It should be noted that this text will be displayed in a terminal
		window, so you should not include any HTML entities in the outputted
		JSON, just format those entities into the characers/strings they represent.
		Format your responses in JSON like this:
		{
			"success": true // whether or not you were able to successfully access and scrape the articles full contents
			"bullets": []string // up to 5 bullet points summarizing the article
			"content": <CONTENT> // the content of the article in a markdown formatted string
		}
		`),
	}

	log.Info("Sending bytedata to gemini")
	resp, err := model.GenerateContent(ctx, req...)
	if err != nil {
		log.Errorf("Error while generating content: %s", err)
		(*progressChan) <- StatusUpdate{
			StatusCode:    -1,
			StatusMessage: "Error while generating content",
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error("Gemini API call timeout exceeded")
		} else {
			log.Errorf("Error while generating content: %s", err)
		}
		(*progressChan) <- StatusUpdate{
			StatusCode:    -1,
			StatusMessage: err.Error(),
		}
		return
	}

	(*progressChan) <- StatusUpdate{
		StatusCode: 3,
	}

	// TODO: Marshal the JSON response to a struct and return an article update message.
	var response GeminiResponse
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			sanitized := sanitizeJSON([]byte(txt))
			if err := json.Unmarshal(sanitized, &response); err != nil {
				log.Error(err)
				log.Info(txt)
			}
			if response.Success {
				article.Content = response.Content
				// BUG: For some reason this does not work, article is still *rendered as* unreadable.
				article.Readable = true
				article.Bullets = response.Bullets
			} else {
				(*progressChan) <- StatusUpdate{
					StatusCode: -1,
					StatusMessage: "Gemini was unable to parse the article",
				}
				return
			}
		}
	}

	(*progressChan) <- StatusUpdate{
		StatusCode: 4,
		StatusMessage: "Completed",
	}

	close(*progressChan)
	log.Info("Finished talking to Gemini, closing channels.")
}

func GetAllNews() tea.Msg {
	var news []NewsArticle

	// TODO: Refactor this code to get news from every RSS feed in the config file
	// NOTE: Program crashed the first time I tried, but it's 2 in the morning so what do I know

	rssFeeds := shared.Koanf.Strings("news.rss_feeds")

	for _, url := range rssFeeds {
		log.Infof("Getting news from %s", url)
		news = append(news, GetRSSFeed(url)...)
	}

	news = append(news, GetTENews()...)

	sort.Slice(news, func(i, j int) bool {
		return news[i].PublicationDate.After(news[j].PublicationDate)
	})

	return NewsUpdate(news)
}

// Get `NewsArticle`s from an RSS url
func GetRSSFeed(RSS_URL string) []NewsArticle {
	fp := feed.Parser{}

	feed, err := fp.ParseURL(RSS_URL)

	if err != nil {
		log.Error("Failed to get Yahoo News data", "error: ", err)
	}

	source := feed.Title // title of the RSS feed
	var articles []NewsArticle
	for _, item := range feed.Items {
		article := NewsArticle{
			Title:           item.Title,
			Source:          source,
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

	client := &http.Client{Timeout: 10 * time.Second}

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
			Source:          "TradingEconomics",
			Readable:        true,
			Content:         n.Description,
			PublicationDate: parsedTime,
		}

		articleData = append(articleData, article)
	}

	return articleData

}
