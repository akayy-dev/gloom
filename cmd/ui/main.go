package main

import (
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

// The "entry" model.
type MainModel struct {
	// pointers to all the tabs
	tabs []*tea.Model
	// index of active tab in the list
	activeTab int
}

func (m MainModel) Init() tea.Cmd {
	tab := *m.tabs[m.activeTab]
	return tea.Batch(tea.ClearScreen, tea.SetWindowTitle("Gloomberg Terminal"), tab.Init())
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	tab := *m.tabs[m.activeTab]
	tab.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		for i, _ := range m.tabs {
			// run if key index is equal to key pressed (accounting for 0 index shift)
			// TODO: Test if this actually preserves state.
			if keyIndex, err := strconv.Atoi(msg.String()); err == nil && i+1 == keyIndex {
				m.activeTab = i
				log.Infof("Switching to view tabs[%d]", i)
				return m, nil
			}
		}
		switch msg.String() {
		case "Q":
			fallthrough
		case "q":
			log.Info("Exiting on user request")
			return m, tea.Quit
		}
	case tea.QuitMsg:
		return m, tea.ClearScreen
	}

	updatedModel := MainModel{
		tabs:      m.tabs,
		activeTab: m.activeTab,
	}
	return updatedModel, nil
}

func (m MainModel) View() string {
	tab := *m.tabs[m.activeTab]

	return tab.View()
}

func main() {
	var dash tea.Model = &Dashboard{
		name: "Dashboard A",
	}

	var cal tea.Model = &EconomicCalendar{}

	m := MainModel{
		tabs:      []*tea.Model{&dash, &cal},
		activeTab: 0,
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
