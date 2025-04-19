package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type Tab struct {
	name  string
	model tea.Model
}

// The "entry" model.
type MainModel struct {
	// pointers to all the tabs
	tabs []*Tab
	// index of active tab in the list
	activeTab int
	// model responsible for showing overlay and contents "underneath" it
	overlayManager tea.Model
	// whether or not an overlay is open
	overlayOpen bool
}

type TabChangeMsg int

func (m MainModel) Init() tea.Cmd {
	tab := m.tabs[m.activeTab].model
	return tea.Batch(tea.ClearScreen, tea.SetWindowTitle("gloom"), tab.Init())
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	tab := m.tabs[m.activeTab].model
	var cmd tea.Cmd
	// NOTE: This code was meant to keep the user from being able to send keypresses
	// to the model while an overlay was open. the current implementation suspends ALL
	// messages from being sent, see if you can fix this later.
	if !m.overlayOpen {
		// only send keypresses to the current tab IF we are not in a model right now
		// BUG: When attempting to switch tabs with an ovelay open, nothing will hapen,
		// but when the user closes the modal, then the tab will switch.
		_, cmd = tab.Update(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		for i, _ := range m.tabs {
			// run if key index is equal to key pressed (accounting for 0 index shift)
			// TODO: Test if this actually preserves state.
			if keyIndex, err := strconv.Atoi(msg.String()); err == nil && i+1 == keyIndex {
				return m, func() tea.Msg { return TabChangeMsg(i) }
			}
		}
		switch msg.String() {
		case "Q":
			fallthrough
		case "q":
			log.Info("Exiting on user request")
			return m, tea.Quit
		case "esc":
			if m.overlayOpen {
				log.Info("Exiting overlay")
				m.overlayOpen = false
			}
		}

	case tea.QuitMsg:
		// clear the screen before quitting
		return m, tea.ClearScreen

	case TabChangeMsg:
		log.Infof("Switching to view tabs[%d]", int(msg))
		m.activeTab = int(msg)

	case DisplayOverlayMsg:
		// how to display overlay messages
		// NOTE: The code for pressing escape to exit the overlay
		//  is in the keypress part of this switch statement
		if !m.overlayOpen {
			log.Info("displaying news overlay")
			m.overlayManager = overlay.New(msg, m, overlay.Center, overlay.Center, 0, 0)
			m.overlayOpen = true
		}
	}

	updatedModel := MainModel{
		tabs:           m.tabs,
		activeTab:      m.activeTab,
		overlayManager: m.overlayManager,
		overlayOpen:    m.overlayOpen,
	}
	return updatedModel, cmd

}

func (m MainModel) View() string {
	tab := m.tabs[m.activeTab].model

	// build tabbar
	var b strings.Builder
	for i, t := range m.tabs {
		var tabText string
		if i == m.activeTab {
			bg := lipgloss.NewStyle().Background(lipgloss.Color("#703FFD"))
			tabText = bg.Render(fmt.Sprintf(" (%d) %s ", i+1, t.name))
		} else {
			tabText = fmt.Sprintf(" (%d) %s ", i+1, t.name)
		}
		b.WriteString(tabText)
	}
	if !m.overlayOpen {
		return lipgloss.JoinVertical(0, b.String(), tab.View())
	} else {
		return m.overlayManager.View()
	}
}

func main() {
	var dash tea.Model = &Dashboard{
		name: "Dashboard A",
	}

	var cal tea.Model = &EconomicCalendar{}

	dashTab := &Tab{
		name:  "Dashboard",
		model: dash,
	}

	calTab := &Tab{
		name:  "Calendar",
		model: cal,
	}

	m := MainModel{
		tabs:      []*Tab{dashTab, calTab},
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
