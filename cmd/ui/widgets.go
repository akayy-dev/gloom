package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Add this new constructor function
func CreateNewsTable() table.Model {
	cols := []table.Column{
		{Title: "Headline", Width: 60},
		{Title: "Time", Width: 15},
	}

	// test values
	rows := []table.Row{}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(false),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.Background(lipgloss.Color("#6951FE"))
	s.Selected = s.Selected.Foreground(lipgloss.Color("#703FFD"))

	t.SetStyles(s)

	return t
}
