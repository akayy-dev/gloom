// Place globals variables here.
package shared

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/charmbracelet/glamour/ansi"
)

// Message Events
type ModalCloseMsg bool

type KeyBinding struct {
	Key string
	// The key to display
	KeyDisplay string
	// The text to display to the user displaying the keybind
	HelpText string
}

/*
NOTE: HELPER FUNCTIONS FOR GLAMOUR THEMES“
*/
func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }

// Basically a tea.Model with a method to get the current keybindings.
type MappedModel interface {
	tea.Model
	GetKeys() key.Binding
}

var (
	UserLog  *log.Logger
	Program  *tea.Program
	Renderer *lipgloss.Renderer
)

// Modified Code from https://github.com/charmbracelet/glamour/blob/05e1d5e15ff0d26d8c0301191b9ee0e67524160a/styles/styles.go

const (
	defaultListIndent      = 2
	defaultListLevelIndent = 4
	defaultMargin          = 2
)

// Returns an ansi.StyleConfig object to be used with Glamour renderers, customized to the users config.
func CreateMarkdownUserConfig() ansi.StyleConfig {
	var UserMarkdownConfig = ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       stringPtr("#f8f8f2"),
			},
			Margin: uintPtr(defaultMargin),
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  stringPtr(Koanf.String("theme.accentColor")),
				Italic: boolPtr(true),
			},
			Indent: uintPtr(defaultMargin),
		},
		List: ansi.StyleList{
			LevelIndent: defaultMargin,
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: stringPtr("#f8f8f2"),
				},
			},
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       stringPtr(Koanf.String("theme.accentColor")),
				Bold:        boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BackgroundColor: stringPtr(Koanf.String("theme.accentColor")),
				Color:           stringPtr("#F8F8F2"),
				Bold:            boolPtr(true),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "## ",
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "### ",
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "#### ",
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "##### ",
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "###### ",
			},
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: ansi.StylePrimitive{
			Color:  stringPtr(Koanf.String("theme.accentColor")),
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Bold:  boolPtr(true),
			Color: stringPtr("#ffb86c"),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  stringPtr("#6272A4"),
			Format: "\n--------\n",
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
			Color:       stringPtr("#8be9fd"),
		},
		Task: ansi.StyleTask{
			StylePrimitive: ansi.StylePrimitive{},
			Ticked:         "[✓] ",
			Unticked:       "[ ] ",
		},
		Link: ansi.StylePrimitive{
			Color:     stringPtr("#8be9fd"),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: stringPtr("#ff79c6"),
			Bold:  boolPtr(true),
		},
		Image: ansi.StylePrimitive{
			Color:     stringPtr("#8be9fd"),
			Underline: boolPtr(true),
		},
		ImageText: ansi.StylePrimitive{
			Color:  stringPtr("#ff79c6"),
			Format: "Image: {{.text}} →",
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#50fa7b"),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: stringPtr("#ffb86c"),
				},
				Margin: uintPtr(defaultMargin),
			},
			Chroma: &ansi.Chroma{
				Text: ansi.StylePrimitive{
					Color: stringPtr("#f8f8f2"),
				},
				Error: ansi.StylePrimitive{
					Color:           stringPtr("#f8f8f2"),
					BackgroundColor: stringPtr("#ff5555"),
				},
				Comment: ansi.StylePrimitive{
					Color: stringPtr("#6272A4"),
				},
				CommentPreproc: ansi.StylePrimitive{
					Color: stringPtr("#ff79c6"),
				},
				Keyword: ansi.StylePrimitive{
					Color: stringPtr("#ff79c6"),
				},
				KeywordReserved: ansi.StylePrimitive{
					Color: stringPtr("#ff79c6"),
				},
				KeywordNamespace: ansi.StylePrimitive{
					Color: stringPtr("#ff79c6"),
				},
				KeywordType: ansi.StylePrimitive{
					Color: stringPtr("#8be9fd"),
				},
				Operator: ansi.StylePrimitive{
					Color: stringPtr("#ff79c6"),
				},
				Punctuation: ansi.StylePrimitive{
					Color: stringPtr("#f8f8f2"),
				},
				Name: ansi.StylePrimitive{
					Color: stringPtr("#8be9fd"),
				},
				NameBuiltin: ansi.StylePrimitive{
					Color: stringPtr("#8be9fd"),
				},
				NameTag: ansi.StylePrimitive{
					Color: stringPtr("#ff79c6"),
				},
				NameAttribute: ansi.StylePrimitive{
					Color: stringPtr("#50fa7b"),
				},
				NameClass: ansi.StylePrimitive{
					Color: stringPtr("#8be9fd"),
				},
				NameConstant: ansi.StylePrimitive{
					Color: stringPtr("#bd93f9"),
				},
				NameDecorator: ansi.StylePrimitive{
					Color: stringPtr("#50fa7b"),
				},
				NameFunction: ansi.StylePrimitive{
					Color: stringPtr("#50fa7b"),
				},
				LiteralNumber: ansi.StylePrimitive{
					Color: stringPtr("#6EEFC0"),
				},
				LiteralString: ansi.StylePrimitive{
					Color: stringPtr("#f1fa8c"),
				},
				LiteralStringEscape: ansi.StylePrimitive{
					Color: stringPtr("#ff79c6"),
				},
				GenericDeleted: ansi.StylePrimitive{
					Color: stringPtr("#ff5555"),
				},
				GenericEmph: ansi.StylePrimitive{
					Color:  stringPtr(Koanf.String("theme.accentColor")),
					Italic: boolPtr(true),
				},
				GenericInserted: ansi.StylePrimitive{
					Color: stringPtr("#50fa7b"),
				},
				GenericStrong: ansi.StylePrimitive{
					Color: stringPtr("#ffb86c"),
					Bold:  boolPtr(true),
				},
				GenericSubheading: ansi.StylePrimitive{
					Color: stringPtr("#bd93f9"),
				},
				Background: ansi.StylePrimitive{
					BackgroundColor: stringPtr("#282a36"),
				},
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{},
			},
		},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: "\n🠶 ",
		},
	}
	return UserMarkdownConfig
}
