package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/config"
	"github.com/PsydoV2/Apish/internal/webhook"
)

type view int

const (
	menuView view = iota
	requestView
	responseView
	webhookSetupView
	webhookLiveView
)

type field int

const (
	fieldURL  field = iota
	fieldBody
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

	// Menu
	choices []string
	cursor  int

	// Request
	methodIdx   int
	activeField field
	urlInput    textinput.Model
	history     config.History
	historyIdx  int

	// Body
	bodyMode   bodyMode
	kvPairs    []kvPair
	kvCursor   int
	kvEditing  bool
	kvFocused  int
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

	// Webhook — channels are reference types so model copies share the same channel
	webhookPortInput textinput.Model
	webhookCh        chan webhook.Request // incoming requests from server
	webhookStop      chan struct{}        // close this to shut down the server
	webhookRequests  []webhook.Request   // captured requests, newest first
	webhookCursor    int
	webhookDetail    bool           // true = showing detail of selected request
	webhookVP        viewport.Model // scrolls the detail body
	webhookPort      int
	webhookErr       error
}

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

	port := textinput.New()
	port.Placeholder = "8080"
	port.SetValue("8080")
	port.CharLimit = 5
	port.Width = 10

	return Model{
		activeView:       menuView,
		methodIdx:        0,
		activeField:      fieldURL,
		historyIdx:       -1,
		bodyMode:         bodyModeKV,
		kvPairs:          []kvPair{},
		history:          history,
		webhookPort:      8080,
		webhookPortInput: port,
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
