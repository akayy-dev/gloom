package main

import (
	"context"
	"errors"
	"fmt"
	"gloomberg/cmd/ui/views"
	"gloomberg/internal/shared"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type Tab struct {
	name  string
	model MappedModel
}

type ModalCloseMsg bool

// The "entry" model.
type MainModel struct {
	// pointers to all the tabs
	tabs []*Tab
	// index of active tab in the list
	activeTab int
	// model responsible for showing overlay and contents "underneath" it
	overlayManager *overlay.Model
	// whether or not an overlay is open
	overlayOpen bool
	// Help Menu
	Help help.Model
	// For aligning
	Width int

	// Whether or not the prompt is open, basically makes sure that accidentally pressing q
	// won't exit the program
	PromptOpen bool
	// What the prompt is
	PromptMessage string
}

type TabChangeMsg int

type MappedModel interface {
	tea.Model
	GetKeys() []key.Binding // TODO: Change this to have actual type safety.
}

func (m MainModel) Init() tea.Cmd {
	// SECTION: Setup Configuration
	configHome, err := os.UserHomeDir()

	if err != nil {
		shared.UserLog.Fatal("Error ocurred while loading config file path: %v", err)
	}

	configFilePath := filepath.Join(configHome, ".config", "gloom", "config.json")
	shared.UserLog.Infof("Checking for config file at path %s", configFilePath)

	shared.LoadDefaultConfig()

	// Check if user config file exists
	if _, err := os.Stat(configFilePath); err == nil {
		shared.UserLog.Infof("Config file found at %s, loading...", configFilePath)
		shared.LoadUserConfig(configFilePath)
	}
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
	} else {
		// Send updates to the foreground if it's open.
		// NOTE: Did not think this through, so bugs might show up.
		_, cmd = m.overlayManager.Foreground.Update(msg)
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		shared.UserLog.Infof("Resizign width to %d", m.Width)
	case tea.KeyMsg:
		for i, _ := range m.tabs {
			// run if key index is equal to key pressed (accounting for 0 index shift)
			if keyIndex, err := strconv.Atoi(msg.String()); err == nil && i+1 == keyIndex {
				return m, func() tea.Msg { return TabChangeMsg(i) }
			}
		}
		switch msg.String() {
		case "Q":
			fallthrough
		case "q":
			if !m.PromptOpen {
				shared.UserLog.Info("Exiting on user request")
				return m, tea.Quit
			}
		case "esc":
			if m.overlayOpen {
				shared.UserLog.Info("Exiting overlay")
				m.overlayManager.Foreground.Update(ModalCloseMsg(true))
				m.overlayOpen = false
			}

			if m.PromptOpen {
				m.PromptOpen = false
			}
		}

	case tea.QuitMsg:
		// clear the screen before quitting
		return m, tea.ClearScreen

	case TabChangeMsg:
		shared.UserLog.Infof("Switching to view tabs[%d]", int(msg))
		m.activeTab = int(msg)

	case views.DisplayOverlayMsg:
		// how to display overlay messages
		// NOTE: The code for pressing escape to exit the overlay
		//  is in the keypress part of this switch statement
		if !m.overlayOpen {
			shared.UserLog.Info("displaying news overlay")
			m.overlayManager = overlay.New(msg, m, overlay.Center, overlay.Center, 0, 0)
			m.overlayOpen = true
			cmd = m.overlayManager.Foreground.Init() // so commands returned from the overlay on init run
		}
	case shared.OpenPromptMsg:
		log.Info("Prompt is open")
		m.PromptOpen = true
		m.PromptMessage = string(msg)
		return m, nil
	}

	updatedModel := MainModel{
		tabs:           m.tabs,
		activeTab:      m.activeTab,
		overlayManager: m.overlayManager,
		overlayOpen:    m.overlayOpen,
		Width:          m.Width,
		PromptOpen:     m.PromptOpen,
	}
	return updatedModel, cmd

}

func RenderHelp(keys []key.Binding, width int) string {
	var b strings.Builder

	accentColor := shared.Koanf.String("theme.accentColor")

	boldStyle := shared.Renderer.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(accentColor))
	for _, binds := range keys {
		b.WriteString(fmt.Sprintf("%s - %s ", boldStyle.Render(binds.Help().Key), binds.Help().Desc))
	}

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, b.String())
}

func (m MainModel) View() string {
	tab := m.tabs[m.activeTab].model

	accentColor := shared.Koanf.String("theme.accentColor")

	// build tabbar
	var b strings.Builder
	for i, t := range m.tabs {
		var tabText string
		if i == m.activeTab {
			bg := shared.Renderer.NewStyle().Background(lipgloss.Color(accentColor))
			tabText = bg.Render(fmt.Sprintf(" (%d) %s ", i+1, t.name))
		} else {
			tabText = fmt.Sprintf(" (%d) %s ", i+1, t.name)
		}
		b.WriteString(tabText)
	}
	if !m.overlayOpen {
		// NOTE: Can't hardcode the help keys forever, going to have to refactor this, and probably
		// the whole help dialog framework to make this more exendable, but this will work for now.
		if m.PromptOpen {
			promptStyle := shared.Renderer.NewStyle().Foreground(lipgloss.Color(shared.Koanf.String("theme.accentColor"))).Bold(true).SetString(m.PromptMessage)
			return lipgloss.JoinVertical(0, b.String(), tab.View(), promptStyle.Render())
		} else {
			return lipgloss.JoinVertical(0, b.String(), tab.View(), RenderHelp(tab.GetKeys(), m.Width))
		}
	} else {
		var keyBinds = []key.Binding{
			key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "exit overlay"),
			),
			key.NewBinding(
				key.WithKeys("j"),
				key.WithHelp("j", "scroll down"),
			),
			key.NewBinding(
				key.WithKeys("k"),
				key.WithHelp("k", "scroll up"),
			),
		}

		screen := m.overlayManager.View()

		lines := strings.Split(screen, "\n")

		screen = strings.Join(lines[:len(lines)-1], "\n")

		if m.PromptOpen {
			return lipgloss.JoinVertical(0, screen, "Prmpt")
		} else {
			return lipgloss.JoinVertical(0, screen, RenderHelp(keyBinds, m.Width))
		}

	}
}

// Function to setup the application as an SSH server.
func setupSSHServer(host string, port string, logFile *os.File) {
	logOutput := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(logOutput)

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbleteaMiddleware(),
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

func main() {
	logFile, err := os.OpenFile("./debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer logFile.Close()

	// SECTION: SSH Server setup
	host, hostExists := os.LookupEnv("SSH_HOST")
	port, portExists := os.LookupEnv("SSH_PORT")

	if hostExists && portExists {
		setupSSHServer(host, port, logFile)
	} else {
		// TODO: Setup program without SSH
		shared.UserLog = log.New(logFile)
		shared.UserLog.SetOutput(logFile)
		log.SetOutput(logFile)

		var dash MappedModel = &views.Dashboard{
			Name: "Dashboard A",
		}

		dashTab := &Tab{
			name:  "Dashboard",
			model: dash,
		}

		m := MainModel{
			tabs:      []*Tab{dashTab},
			activeTab: 0,
		}

		shared.Program = tea.NewProgram(m)
		shared.Program.Run()
	}
}

// Custom middleware for bubbletea, sets the Program variable.
func bubbleteaMiddleware() wish.Middleware {
	log.Info("Starting middleware")
	newProg := func(m tea.Model, opts []tea.ProgramOption) *tea.Program {
		p := tea.NewProgram(m, opts...)
		shared.Program = p
		return p
	}
	log.Info("bubbletea program created")
	teaHandler := func(s ssh.Session) *tea.Program {
		return newProg(setupSSHApplication(s))
	}

	return bm.MiddlewareWithProgramHandler(teaHandler, termenv.Ascii)

}

// Setup bubletea model to work with Wish
func setupSSHApplication(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	log.Info("setupBubbleTea")
	userString := fmt.Sprintf("%s.%s", s.User(), strings.Split(s.RemoteAddr().String(), ":")[0])
	log.Infof("Connection from %s", userString)
	// pty, _, _ := s.Pty()

	// use instead of lipgloss.NewStyle()
	shared.Renderer = bubbletea.MakeRenderer(s)

	// CREATE USER LOGGER
	/* BUG: File closes after function ends,
	making logging impossible after end of function
	Need to make a cleanup function that runs when the server closes,
	can be handled in the main function if we make the file a global variable.
	*/

	logTimeStamp := time.Now().Format("01.02.2006 15:04 MST")

	// Make the logs directory if it doesn't exist yet.
	if err := os.MkdirAll("./logs", 0755); err != nil {
		log.Error("Cannot create logs directory", "error", err)
	}

	f, err := os.OpenFile(
		fmt.Sprintf("./logs/%s %s.log",
			userString,
			logTimeStamp,
		),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

	if err != nil {
		log.Error("Cannot create log file", err)
	}

	shared.UserLog = log.New(f)
	// NOTE: Setting time format doesn't work, figure out how to fix this later.
	shared.UserLog.SetTimeFormat("2006/01/02 15:04:05")
	shared.UserLog.Info("User log created")

	// This function runs on
	go func() {
		<-s.Context().Done()
		shared.UserLog.Info("Connection closed, ending file.")
		if err := f.Close(); err != nil {
			log.Error("Error closing log file", "error", err)
		}
	}()

	var dash MappedModel = &views.Dashboard{
		Name: "Dashboard A",
	}

	dashTab := &Tab{
		name:  "Dashboard",
		model: dash,
	}

	m := MainModel{
		tabs:      []*Tab{dashTab},
		activeTab: 0,
	}

	return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithInput(s), tea.WithOutput(s)}
}
