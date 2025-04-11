package main

import (
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

// The "entry" model.
type MainModel struct {
	activeModel tea.Model
	tabs        []tea.Model
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, tea.SetWindowTitle("Gloomberg Terminal"), m.activeModel.Init())
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		for i, _ := range m.tabs {
			log.Infof("index %v", i)
			// run if key index is equal to key pressed (accounting for 0 index shift)
			if keyIndex, err := strconv.Atoi(msg.String()); err == nil && i+1 == keyIndex {
				m.activeModel = m.tabs[i]
				m.activeModel.Init()
				log.Infof("Switching to view tabs[%d]", i)
			}
		}
		switch msg.String() {
		case "Q":
			fallthrough
		case "q":
			log.Info("Exiting on user request")
			return m.activeModel, tea.Quit
		}
	case tea.QuitMsg:
		return m.activeModel, tea.ClearScreen
	}

	updatedChild, cmd := m.activeModel.Update(msg)
	updatedModel := MainModel{
		activeModel: updatedChild,
		tabs:        m.tabs,
	}
	return updatedModel, cmd
}

func (m MainModel) View() string {
	return m.activeModel.View()
}

func main() {
	m := MainModel{
		activeModel: Dashboard{
			name:      "Dashboard A",
			newsTable: CreateNewsTable(),
		},
		tabs: []tea.Model{
			Dashboard{
				name:      "Dashboard A",
				newsTable: CreateNewsTable(),
			},
			EconomicCalendar{},
		},
	}

	f, err := os.OpenFile("./debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	defer f.Close()

	p := tea.NewProgram(m)
	p.Run()
}
