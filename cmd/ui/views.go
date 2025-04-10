package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	tea "github.com/charmbracelet/bubbletea"
)

type Dashboard struct {
	name  string
	table table.Model
}

func (d Dashboard) Init() tea.Cmd {

	return nil
}

func (d Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var returnCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if d.table.Focused() {
				log.Info("Blurring")
				d.table.Blur()
			} else {
				log.Info("Focusing")
				d.table.Focus()

			}

		}
	}
	d.table, returnCmd = d.table.Update(msg)
	return d, returnCmd
}

func (d Dashboard) View() string {
	tableStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#FFFFFF"))
	return tableStyle.Render(d.table.View())
}
