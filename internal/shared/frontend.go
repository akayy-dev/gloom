// Place globals variables here.
package shared

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// Message Events
type ModalCloseMsg bool

var (
	UserLog  *log.Logger
	Program  *tea.Program
	Renderer *lipgloss.Renderer
)
