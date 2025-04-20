package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
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
	// The renderer that gets passed down to the child models
	renderer *lipgloss.Renderer
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
			bg := m.renderer.NewStyle().Background(lipgloss.Color("#703FFD"))
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
	logFile, err := os.OpenFile("./debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	logOutput := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(logOutput)
	defer logFile.Close()

	// SECTION: SSH Server setup
	host := os.Getenv("SSH_HOST")
	port := os.Getenv("SSH_PORT")

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(setupBubbleTea),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	// Done channel notifies when program is closed or killed
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Starting SSH Server", "Host", host, "Port", port)

	go func() {
		// Start SSH server and log if there is an error that causes the server to close
		if err = s.ListenAndServe(); err != nil && errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	// Code below this only runs when the server is closed.
	<-done

	log.Info("Stopping SSH Server")
	// give 30 seconds for ssh server to stop
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()

	if err := s.Shutdown(ctx); err != nil && errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

// Setup bubletea model to work with Wish
func setupBubbleTea(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	// pty, _, _ := s.Pty()

	// use instead of lipgloss.NewStyle()
	renderer := bubbletea.MakeRenderer(s)

	var dash tea.Model = &Dashboard{
		name:     "Dashboard A",
		renderer: renderer,
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
		renderer:  renderer,
	}

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
