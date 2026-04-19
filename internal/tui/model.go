package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/config"
)

type view int

const (
	menuView view = iota
	requestView
	responseView
)

// field beschreibt welcher Bereich der Request-View aktiv ist.
// Method ist kein eigenes Feld mehr — wird mit [/] aus der URL geändert.
type field int

const (
	fieldURL  field = iota
	fieldBody       // KV-Builder oder Raw-Textarea
)

var methods = []string{"GET", "POST", "PUT", "DELETE"}

func hasBody(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

const (
	verticalOverhead   = 13
	horizontalOverhead = 10
)

type Model struct {
	activeView view
	width      int
	height     int

	// Menü
	choices []string
	cursor  int

	// Request
	methodIdx   int
	activeField field
	urlInput    textinput.Model
	history     config.History // persistierte History
	historyIdx  int            // -1 = kein Browsing, sonst Index in history.Entries

	// Body
	bodyMode   bodyMode
	kvPairs    []kvPair
	kvCursor   int
	kvEditing  bool
	kvFocused  int // 0 = Key, 1 = Value
	kvKeyInput textinput.Model
	kvValInput textinput.Model
	bodyInput  textarea.Model

	// Response
	viewport      viewport.Model
	requestURL    string
	requestMethod string
	response      string
	contentType   string
	statusCode    int
	statusText    string
	loading       bool
	err           error
}

// New erstellt das initiale Model. history wird beim Start aus der Datei geladen
// und hier übergeben — so bleibt das Model testbar ohne Dateisystem-Zugriff.
func New(history config.History) Model {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com/users"
	ti.CharLimit = 500
	ti.Width = 50

	kvKey := textinput.New()
	kvKey.Placeholder = "key"
	kvKey.Width = 22

	kvVal := textinput.New()
	kvVal.Placeholder = "value"
	kvVal.Width = 26

	ta := textarea.New()
	ta.Placeholder = "{\n  \"key\": \"value\"\n}"
	ta.SetWidth(50)
	ta.SetHeight(8)
	ta.ShowLineNumbers = false

	return Model{
		activeView:  menuView,
		methodIdx:   0,
		activeField: fieldURL,
		historyIdx:  -1,
		bodyMode:    bodyModeKV,
		kvPairs:     []kvPair{},
		history:     history,
		choices: []string{
			"Send API Request",
			"Start Webhook Catcher",
			"Settings",
		},
		urlInput:   ti,
		kvKeyInput: kvKey,
		kvValInput: kvVal,
		bodyInput:  ta,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
