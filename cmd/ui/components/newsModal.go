package components

import (
	"context"
	"fmt"
	"gloomberg/internal/scraping"
	"gloomberg/internal/utils"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// Message for when the article is finished scraping
type UpdateContentMsg scraping.NewsArticle

// Sends the progress of scraping to Update()
type UpdateStatusMsg scraping.StatusUpdate

// Pop-up model displaying news
type NewsModal struct {
	Article *scraping.NewsArticle
	// width
	W int
	H int
	// viewport model
	vp viewport.Model

	// glamour utils.Renderer
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
	progressChan chan scraping.StatusUpdate
}

// begin newsscraping
func scrapeNews(article *scraping.NewsArticle, status *chan scraping.StatusUpdate, ctx context.Context) tea.Cmd {
	log.Info("scrapeNews CMD")
	return func() tea.Msg {
		utils.UserLog.Info("scrapeNews Cmd run")
		go func() {
			for progress := range *status {
				utils.Program.Send(UpdateStatusMsg(progress))
			}
		}()
		go scraping.PromptNewsURL(article, status, ctx) // needs to run in it's own routine for listen to workk
		return nil
	}
}

func (n *NewsModal) styleArticle() (string, error) {
	utils.UserLog.Info("Styling markdown")
	md, err := n.styler.Render(n.Article.Content)
	if err != nil {
		utils.UserLog.Errorf("Cannot render markdown content %s", err)
	}

	var header string

	if len(n.Article.Bullets) > 0 {
		// TODO: build a list of bullets and render them in markdown
		var builder strings.Builder

		// building bullets as a (unrendered) list
		for i, bullet := range n.Article.Bullets {
			// don't put a newline if we are at the last bullet point
			if i < len(n.Article.Bullets)-1 {
				builder.WriteString(fmt.Sprintf("- %s\n", bullet))
			} else {
				builder.WriteString(fmt.Sprintf("- %s", bullet))
			}
		}

		// NOTE: For some reason there needs to be two newlines for summary to render on a different line
		// than published. Don't know why but if it works it works

		header, err = n.styler.Render(fmt.Sprintf("# %s\n## %s\n*Published: %s* \n\n  Summary \n %s \n ---",
			n.Article.Title,
			n.Article.Source,
			n.Article.PublicationDate.Format("01/02/2006"),
			builder.String()))

	} else {
		header, err = n.styler.Render(fmt.Sprintf("# %s\n## %s\n*Published: %s*",
			n.Article.Title,
			n.Article.Source,
			n.Article.PublicationDate.Format("01/02/2006")))

	}
	if err != nil {
		utils.UserLog.Errorf("Cannot render markdown content %s", err)
	}
	return fmt.Sprintf("%s\n%s", header, md), nil
}

func (n *NewsModal) Init() tea.Cmd {
	// SECTION: Load markdown theme based on theme options

	// initialize viewport with full width but minimal height
	n.vp = viewport.New(n.W, 1)
	n.vp.Style = utils.Renderer.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 0).
		Width(n.W)

	// initialize glamour utils.Renderer
	var err error
	n.styler, err = glamour.NewTermRenderer(
		glamour.WithStyles(utils.CreateMarkdownUserConfig()),
		glamour.WithWordWrap(n.W-5),
	)
	if err != nil {
		utils.UserLog.Errorf("Cannot create glamour utils.Renderer %s", err)
	}

	// if article is not readable, scrape it
	if !n.Article.Readable {
		utils.UserLog.Info("Article not readable, loading content")
		n.loading = true
		n.progressChan = make(chan scraping.StatusUpdate)

		var ctx context.Context
		ctx, n.newsCtxCancel = context.WithCancel(context.Background())
		n.newsCtx, n.newsCtxCancel = context.WithTimeout(ctx, 30*time.Second)

		return tea.Batch(
			scrapeNews(n.Article, &n.progressChan, n.newsCtx),
		)

	} else {
		// if you don't need to scrape
		content, err := n.styleArticle()
		n.vp.SetContent(content)
		if err != nil {
			utils.UserLog.Errorf("Cannot render markdown content %s", err)
		}
		n.loading = false
		n.vp.Height = n.H
		return nil
	}

}

func (n *NewsModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// update viewport size
		n.W = msg.Width / 2
		n.H = int(float64(msg.Height) * .8)
		n.vp.Width = n.W
		n.vp.Height = n.H

	case tea.KeyMsg:
		switch key := msg.String(); key {
		case "esc":
			return n, func() tea.Msg { return utils.ModalCloseMsg(true) }
		}

	case utils.ModalCloseMsg:
		// this basically checks if we've scraped the news using ai
		if n.loading {
			utils.UserLog.Info("Closing news modal and cancelling network request")
			n.newsCtxCancel()
		}
	case UpdateContentMsg:
		utils.UserLog.Info("Finished scraping article")
		n.vp.Height = n.H
		content, err := n.styleArticle()
		if err != nil {
			utils.UserLog.Errorf("Cannot render markdown content %s", err)
		}
		n.loading = false
		n.vp.SetContent(content)
	case UpdateStatusMsg:
		// TODO: Modify code to constantly call Update with an UpdateStatusMsg,
		// do this until loading is finished, then call UpdateContentMsg
		var statusMsg string
		switch msg.StatusCode {
		case -1: // error case
			// NOTE: Add red bold formatting to error message
			statusMsg = fmt.Sprintf(" An error occured\n%s", msg.StatusMessage)
			n.newsCtxCancel()
		case 0:
			statusMsg = "󰖟 Initialized scraping protocol"
		case 1:
			statusMsg = "󰖟 Requesting article data"
		case 2:
			statusMsg = "󰇚 Downloading article"
		case 3:
			statusMsg = " Scraping text from article"
		case 4:
			statusMsg = " Done"
			utils.UserLog.Debug(statusMsg)
			return n, func() tea.Msg { return UpdateContentMsg(*n.Article) }
		}

		n.statusMessage = statusMsg
		utils.UserLog.Info(statusMsg)
	}
	n.vp, cmd = n.vp.Update(msg)
	return n, cmd
}

func (n *NewsModal) View() string {
	if n.loading {
		statusStyle := utils.Renderer.NewStyle().
			Width(n.W).
			Height(10).
			Align(lipgloss.Center, lipgloss.Center).
			Border(lipgloss.NormalBorder())
		responseUI := fmt.Sprintf("%s\n\n%s", n.statusMessage, "Press esc to cancel")
		return statusStyle.Render(responseUI)

	} else {
		return n.vp.View()
	}
}

func (n *NewsModal) GetKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("<esc>", "close article"),
		),
		key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("<j>", "scroll down"),
		),
		key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("<k>", "scroll up"),
		),
	}
}
