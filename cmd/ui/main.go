package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

// The "entry" model.
type MainModel struct {
	activeModel tea.Model
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, tea.SetWindowTitle("Gloomberg Terminal"))
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updatedChild, cmd := m.activeModel.Update(msg)
	updatedModel := MainModel{
		activeModel: updatedChild,
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "Q":
			fallthrough
		case "q":
			log.Info("Exiting on user request")
			return m.activeModel, tea.Quit
		}
	}

	return updatedModel, cmd
}

func (m MainModel) View() string {
	return m.activeModel.View()
}

func main() {
	m := MainModel{
		activeModel: Dashboard{
			name:  "Dashboard A",
			table: CreateNewsTable(),
		},
	}
	p := tea.NewProgram(m)
	p.Run()
}
