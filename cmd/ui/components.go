package main

import (
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
	divStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1, 1).
		Width(n.w).
		Height(n.h)
	return divStyle.Render(n.content)
}
