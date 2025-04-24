package main

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NewsModal struct {
	title   string
	url     string
	content string
	// width
	w int
	h int
	// viewport model
	vp viewport.Model
}

func (n *NewsModal) Init() tea.Cmd {
	n.vp = viewport.New(n.w, n.h)
	n.vp.SetContent(n.content)
	return nil
}

func (n *NewsModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		n.w = msg.Width / 2
		n.h = int(float64(msg.Height) * .8)
		n.vp = viewport.New(n.w, n.h)
		n.vp.SetContent(n.content)
	}
	return n, nil
}

func (n *NewsModal) View() string {
	// // format and align text
	// titleStyle := Renderer.NewStyle().
	// 	Bold(true).
	// 	Foreground(lipgloss.Color("#703FFD")).
	// 	SetString(n.title)

	divStyle := Renderer.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1, 1).
		Width(n.w).
		MaxHeight(n.h)
	// controlsDiv := lipgloss.NewStyle().
	// 	Foreground(lipgloss.Color("#FFFFFF")).
	// 	Align(lipgloss.Right).
	// 	Width(n.w - 2).
	// 	SetString(lipgloss.NewStyle().Background(lipgloss.Color("#703FFD")).Padding(0, 1).Render("<ESC> Go Back"))

	// renderedString := lipgloss.JoinVertical(lipgloss.Top, titleStyle.Render(), n.content, controlsDiv.Render())
	// return divStyle.Render(renderedString)

	return divStyle.Render(n.vp.View())
}
