package main

import (
	"context"
	"fmt"
	"gloomberg/internal/scraping"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// Message for when the article is finished scraping
type UpdateContentMsg scraping.NewsArticle

type UpdateStatusMsg int

type NewsModal struct {
	article *scraping.NewsArticle
	// width
	w int
	h int
	// viewport model
	vp viewport.Model

	// glamour renderer
	styler *glamour.TermRenderer

	// context for prompt function.
	newsCtx context.Context
	// cancel function for newsCtx
	newsCtxCancel context.CancelFunc
	// whether or not the article has loaded
	loading bool
	// status message
	statusMessage string

	// channel for progress updates for newsscraping
	progressChan chan int
}

// begin newsscraping
func scrapeNews(article *scraping.NewsArticle, status *chan int, ctx *context.Context) tea.Cmd {
	log.Info("scrapeNews CMD")
	return func() tea.Msg {
		UserLog.Info("scrapeNews Cmd run")
		scraping.PromptNewsURL(article, status, ctx)
		return UpdateStatusMsg(0)
	}
}

// Listen for gemini scraping protocol updates
func listenForNews(status *chan int) tea.Cmd {
	UserLog.Info("listenForNews Cmd run")
	return func() tea.Msg {
		code := <-*status
		return UpdateStatusMsg(code)
	}
}

func (n *NewsModal) styleArticle() (string, error) {
	UserLog.Info("Styling markdown")
	md, err := n.styler.Render(n.article.Content)
	if err != nil {
		UserLog.Errorf("Cannot render markdown content %s", err)
	}

	header, err := n.styler.Render(fmt.Sprintf("# %s\n## %s\n*Published: %s*", n.article.Title, n.article.Source, n.article.PublicationDate.Format("01/02/2006")))
	if err != nil {
		UserLog.Errorf("Cannot render markdown content %s", err)
	}
	return fmt.Sprintf("%s\n\n%s", header, md), nil
}

func (n *NewsModal) Init() tea.Cmd {
	// initialize viewport with full width but minimal height
	n.vp = viewport.New(n.w, 1)
	n.vp.Style = Renderer.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 0).
		Width(n.w)

	// initialize glamour renderer
	var err error
	n.styler, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(n.w-5),
	)
	if err != nil {
		UserLog.Errorf("Cannot create glamour renderer %s", err)
	}

	// if article is not readable, scrape it
	if !n.article.Readable {
		UserLog.Info("Article not readable, loading content")
		n.loading = true
		n.progressChan = make(chan int)

		n.newsCtx, n.newsCtxCancel = context.WithCancel(context.Background())
		return tea.Batch(
			scrapeNews(n.article, &n.progressChan, &n.newsCtx),
			listenForNews(&n.progressChan),
		)

	} else {
		// if you don't need to scrape
		content, err := n.styleArticle()
		n.vp.SetContent(content)
		if err != nil {
			UserLog.Errorf("Cannot render markdown content %s", err)
		}
		n.loading = false
		n.vp.Height = n.h
		return nil
	}
}

func (n *NewsModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// update viewport size
		n.w = msg.Width / 2
		n.h = int(float64(msg.Height) * .8)
		n.vp.Width = n.w
		n.vp.Height = n.h

	case ModalCloseMsg:
		// this basically checks if we've scraped the news using ai
		if n.loading {
			UserLog.Info("Closing news modal and cancelling network request")
			n.newsCtxCancel()
		}
	case UpdateContentMsg:
		UserLog.Info("Finished scraping article")
		n.vp.Height = n.h
		content, err := n.styleArticle()
		if err != nil {
			UserLog.Errorf("Cannot render markdown content %s", err)
		}
		n.loading = false
		n.vp.SetContent(content)
	case UpdateStatusMsg:
		// TODO: Modify code to constantly call Update with an UpdateStatusMsg,
		// do this until loading is finished, then call UpdateContentMsg
		var statusMsg string
		switch msg {
		case -1: // error case
			UserLog.Error(" Error occured while scraping article, closing context")
			statusMsg = " An error occured"
			n.newsCtxCancel()
		case 0:
			statusMsg = "󰖟 Initialized scraping protocol"
		case 1:
			statusMsg = "󰖟 Requesting article data"
		case 2:
			statusMsg = " Converting article to HTML file"
		case 3:
			statusMsg = " Scraping text from article"
		case 4:
			statusMsg = " Done"
			return n, func() tea.Msg { return UpdateContentMsg(*n.article) }
		}

		n.statusMessage = statusMsg
		UserLog.Info(statusMsg)
		return n, listenForNews(&n.progressChan)
	}
	n.vp, cmd = n.vp.Update(msg)
	return n, cmd
}

func (n *NewsModal) View() string {
	if n.loading {
		statusStyle := Renderer.NewStyle().
			Width(n.w).
			Height(10).
			Align(lipgloss.Center, lipgloss.Center).
			Border(lipgloss.NormalBorder())
		responseUI := fmt.Sprintf("%s\n\n%s", n.statusMessage, "Press esc to cancel")
		return statusStyle.Render(responseUI)

	} else {
		return n.vp.View()
	}
}
