package main

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NewsTable struct {
	Width  int
	Height int
	table  table.Model
}

func (n NewsTable) Init() tea.Cmd {

	return nil
}

func (n NewsTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var returnCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if n.table.Focused() {
				n.table.Blur()
			} else {
				n.table.Focus()

			}

		}
	}
	n.table, returnCmd = n.table.Update(msg)
	return n, returnCmd
}

func (n NewsTable) View() string {
	return n.table.View()
}

// Add this new constructor function
func CreateNewsTable() table.Model {
	cols := []table.Column{
		{Title: "Headline", Width: 60},
		{Title: "Source", Width: 15},
		{Title: "Time", Width: 15},
	}

	// test values
	rows := []table.Row{
		{"Trump does something stupid", "The Wall Street Journal", "04/20"},
		{"Example2", "News2", "Date2"},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.Background(lipgloss.Color("#6951FE"))
	s.Selected = s.Selected.Foreground(lipgloss.Color("#703FFD"))

	t.SetStyles(s)

	return t
}
