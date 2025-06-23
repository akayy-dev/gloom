package main

import (
	"context"
	"errors"
	"fmt"
	"gloomberg/cmd/ui/views"
	"gloomberg/internal/utils"
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
	"github.com/charmbracelet/bubbles/textinput"

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

type OverlayWrapper struct {
	*overlay.Model
	foreground MappedModel
	background MappedModel
}

// Implement MappedModel interface for OverlayWrapper
func (o *OverlayWrapper) GetKeys() []key.Binding {
	if o.foreground != nil {
		return o.foreground.GetKeys()
	}
	return []key.Binding{}
}

type Prompt struct {
	Model    textinput.Model
	Prompt   string
	Callback func(string) tea.Msg
}

// The "entry" model.
type MainModel struct {
	// pointers to all the tabs
	tabs []*Tab
	// index of active tab in the list
	activeTab int
	// model responsible for showing overlay and contents "underneath" it
	overlayManager *OverlayWrapper
	// whether or not an overlay is open
	overlayOpen bool
	// Help Menu
	Help help.Model
	// For aligning
	Width int
	// What the prompt is
	PromptMessage string
	// Prompt Model
	input Prompt

	// The notification text displaying
	NotificationText string
	// Whether or not a notification is showing
	ShowingNotification bool
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
		utils.UserLog.Fatal("Error ocurred while loading config file path: %v", err)
	}

	configFilePath := filepath.Join(configHome, ".config", "gloom", "config.json")
	utils.UserLog.Infof("Checking for config file at path %s", configFilePath)

	utils.LoadDefaultConfig()

	// Load prompt model
	m.input = Prompt{
		Model: textinput.New(),
	}

	// Check if user config file exists
	if _, err := os.Stat(configFilePath); err == nil {
		utils.UserLog.Infof("Config file found at %s, loading...", configFilePath)
		utils.LoadUserConfig(configFilePath)
	}
	tab := m.tabs[m.activeTab].model
	return tea.Batch(tea.ClearScreen, tea.SetWindowTitle("gloom"), tab.Init())
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	tab := m.tabs[m.activeTab].model
	var cmd tea.Cmd
	if !m.overlayOpen && !m.input.Model.Focused() {
		// Only send keypresses to the current tab if we are not in a modal right now
		_, cmd = tab.Update(msg)
	} else if m.overlayOpen {
		// Send updates to the foreground if it's open
		_, cmd = m.overlayManager.Foreground.Update(msg)
		if _, ok := msg.(tea.KeyMsg); !ok {
			// If the message is not a KeyMsg, also send it to the background tab
			_, _ = tab.Update(msg)
		}
	}
	if m.input.Model.Focused() {
		if _, ok := msg.(tea.KeyMsg); !ok {
			_, cmd = tab.Update(msg)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
	case tea.KeyMsg:
		if m.input.Model.Focused() {
			// if the user presses escape break out of the prompt
			if msg.String() == "esc" {
				m.input.Model.Blur()
			} else if msg.String() == "enter" {
				m.input.Model.Blur()
				if m.input.Callback != nil {
					cmd = func() tea.Msg {
						return m.input.Callback(m.input.Model.Value())
					}
				} else {
					log.Warn("Tried to run prompt callback but was nil, did you set the value of CallbackFunc?")
				}
				break
			} else {
				m.input.Model, cmd = m.input.Model.Update(msg)
				break
			}

			break
		}

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
			if !m.input.Model.Focused() && !m.overlayOpen {
				utils.UserLog.Info("Exiting on user request")
				return m, tea.Batch(tea.ClearScreen, tea.Quit)
			}

			if m.input.Model.Focused() {
				m.input.Model.Blur()
				return m, nil
			}
		}

	case utils.ModalCloseMsg:
		utils.UserLog.Info("Exiting overlay")
		m.overlayOpen = false
		m.overlayManager.Foreground.Update(msg)
		m.overlayManager.Background = nil
		return m, nil

	case tea.QuitMsg:
		// clear the screen before quitting
		return m, tea.ClearScreen

	case TabChangeMsg:
		utils.UserLog.Infof("Switching to view tabs[%d]", int(msg))
		m.activeTab = int(msg)

	case views.DisplayOverlayMsg:
		// how to display overlay messages
		// NOTE: The code for pressing escape to exit the overlay
		//  is in the keypress part of this switch statement
		if !m.overlayOpen {
			utils.UserLog.Info("displaying overlay")
			// Create the overlay model
			overlayModel := overlay.New(msg, m, overlay.Center, overlay.Center, 0, 0)

			// Create the wrapper
			m.overlayManager = &OverlayWrapper{
				Model:      overlayModel,
				foreground: msg.(MappedModel), // Cast the message to MappedModel
				background: m,                 // MainModel already implements MappedModel
			}
			m.overlayOpen = true
			cmd = m.overlayManager.Foreground.Init()
		}
	case utils.PromptOpenMsg:
		log.Infof("PromptOpenMsg: %s", msg.Prompt)

		m.input.Model = textinput.New()
		promptWidth := m.Width - len(msg.Prompt)

		m.input.Model.Width = promptWidth
		m.input.Model.CharLimit = promptWidth
		m.input.Model.Focus()
		m.input.Prompt = msg.Prompt
		m.input.Callback = msg.CallbackFunc

	case utils.SendNotificationMsg:
		m.NotificationText = msg.Message
		m.ShowingNotification = true
		log.Info("enabled showing notification")
		cmd = tea.Tick(time.Duration(msg.DisplayTime*int(time.Millisecond)), func(t time.Time) tea.Msg {
			return utils.HideNotificationMsg{}
		})
	case utils.HideNotificationMsg:
		m.ShowingNotification = false
		log.Info("hiding notification")
	}

	updatedModel := MainModel{
		tabs:                m.tabs,
		activeTab:           m.activeTab,
		overlayManager:      m.overlayManager,
		overlayOpen:         m.overlayOpen,
		Width:               m.Width,
		input:               m.input,
		NotificationText:    m.NotificationText,
		ShowingNotification: m.ShowingNotification,
	}

	return updatedModel, cmd

}

func RenderHelp(keys []key.Binding, width int) string {
	var b strings.Builder

	accentColor := utils.Koanf.String("theme.accentColor")

	boldStyle := utils.Renderer.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(accentColor))
	for _, binds := range keys {
		b.WriteString(fmt.Sprintf("%s - %s ", boldStyle.Render(binds.Help().Key), binds.Help().Desc))
	}

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, b.String())
}

func (m MainModel) View() string {
	tab := m.tabs[m.activeTab].model

	accentColor := utils.Koanf.String("theme.accentColor")

	// build tabbar
	var b strings.Builder
	for i, t := range m.tabs {
		var tabText string
		if i == m.activeTab {
			bg := utils.Renderer.NewStyle().Background(lipgloss.Color(accentColor))
			tabText = bg.Render(fmt.Sprintf(" (%d) %s ", i+1, t.name))
		} else {
			tabText = fmt.Sprintf(" (%d) %s ", i+1, t.name)
		}
		b.WriteString(tabText)
	}

	// What text to show on the bottom
	var bottomText string
	var screen string

	if !m.overlayOpen {
		screen = tab.View()
		// if the prompt is open show it
		if m.input.Model.Focused() {
			// prompt is bold and in accent color
			styledPrompt := utils.Renderer.NewStyle().Foreground(lipgloss.Color(accentColor)).Bold(true).SetString(m.input.Prompt).Render()
			// NOTE: [:2] removes the leading "> " from the styledPrompt
			bottomText = fmt.Sprintf("%s%s", styledPrompt, m.input.Model.View()[2:])
		} else {
			// render help key when prompt is not opened
			bottomText = RenderHelp(tab.GetKeys(), m.Width)
		}

	} else {
		// FIXME: temporary solution for showing keybinds in news modal
		// in the future, this should be refatored so seach model manages THEIR OWN
		// overlay logic.
		var keyBinds = m.overlayManager.foreground.GetKeys()

		lines := strings.Split(m.overlayManager.View(), "\n")

		screen = strings.Join(lines[:len(lines)-1], "\n")
		bottomText = RenderHelp(keyBinds, m.Width)

	}
	if m.ShowingNotification && !m.input.Model.Focused() {
		log.Infof("Showing notification text: %s", m.NotificationText)
		bottomText = m.NotificationText
	}
	return lipgloss.JoinVertical(0, b.String(), screen, bottomText)
}

func (m MainModel) GetKeys() []key.Binding {
	return m.overlayManager.GetKeys()
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
		utils.UserLog = log.New(logFile)
		utils.UserLog.SetOutput(logFile)
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

		utils.Program = tea.NewProgram(m)
		utils.Program.Run()
	}
}

// Custom middleware for bubbletea, sets the Program variable.
func bubbleteaMiddleware() wish.Middleware {
	log.Info("Starting middleware")
	newProg := func(m tea.Model, opts []tea.ProgramOption) *tea.Program {
		p := tea.NewProgram(m, opts...)
		utils.Program = p
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
	utils.Renderer = bubbletea.MakeRenderer(s)

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

	utils.UserLog = log.New(f)
	// NOTE: Setting time format doesn't work, figure out how to fix this later.
	utils.UserLog.SetTimeFormat("2006/01/02 15:04:05")
	utils.UserLog.Info("User log created")

	// This function runs on
	go func() {
		<-s.Context().Done()
		utils.UserLog.Info("Connection closed, ending file.")
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
