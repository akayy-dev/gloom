package main

import (
	"fmt"

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
}

func (n *NewsModal) Init() tea.Cmd {
	return nil
}

func (n *NewsModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return n, nil
}

func (n *NewsModal) View() string {
	// format and align text
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#703FFD")).
		SetString(n.title)

	divStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1, 1).
		Width(n.w).
		MaxHeight(n.h)
		// Height(n.h)
	// controlsDiv := lipgloss.NewStyle().
	// 	Background(lipgloss.Color("#703FFD")).
	// 	Foreground(lipgloss.Color("#FFFFFF")).
	// 	SetString("<ESC> Close window")

	renderedString := fmt.Sprintf("%s\n%s", titleStyle.Render(), n.content)

	return divStyle.Render(renderedString)
}
