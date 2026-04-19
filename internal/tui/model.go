package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	menuView view = iota
	requestView
	responseView
)

// Model speichert den gesamten Zustand der App.
type Model struct {
	activeView view

	// Terminal-Dimensionen — werden via tea.WindowSizeMsg befüllt
	width  int
	height int

	// Menü-State
	choices []string
	cursor  int

	// Request-State
	urlInput textinput.Model

	// Response-State
	response   string
	statusCode int
	statusText string
	loading    bool
	err        error
}

// New erstellt das initiale Model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com/users"
	ti.CharLimit = 500
	ti.Width = 50

	return Model{
		activeView: menuView,
		choices: []string{
			"GET Request senden",
			"POST Request senden",
			"Webhook Catcher starten",
			"Einstellungen",
		},
		urlInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
