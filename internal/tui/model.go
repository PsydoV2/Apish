package tui

import (
	"net/http"
	"time"

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
	settingsView
	envListView
	envEditView
)

type field int

const (
	fieldURL field = iota
	fieldParams  // Query-Parameter
	fieldHeaders // Request-Headers
	fieldAuth    // Authentifizierung
	fieldBody
)

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

func hasBody(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

type authKind int

const (
	authNone   authKind = iota
	authBearer          // Authorization: Bearer <token>
	authBasic           // Authorization: Basic base64(user:pass)
	authAPIKey          // <header-name>: <value>
)

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

	// Request — URL
	methodIdx   int
	activeField field
	urlInput    textinput.Model
	history     config.History
	historyIdx  int

	// Query-Parameter KV-Builder
	queryParams []kvPair
	qpCursor    int
	qpEditing   bool
	qpFocused   int
	qpKeyInput  textinput.Model
	qpValInput  textinput.Model

	// Request-Headers KV-Builder
	headersKV   []kvPair
	hdrCursor   int
	hdrEditing  bool
	hdrFocused  int
	hdrKeyInput textinput.Model
	hdrValInput textinput.Model

	// Auth
	authKind        authKind
	authFocused     int // -1=Typ-Auswahl, 0=erstes Feld, 1=zweites Feld
	tokenInput      textinput.Model // Bearer Token
	authUserInput   textinput.Model // Basic: Benutzername
	authPassInput   textinput.Model // Basic: Passwort
	apiKeyNameInput textinput.Model // API Key: Header-Name
	apiKeyValInput  textinput.Model // API Key: Wert

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
	viewport        viewport.Model
	requestURL      string
	requestMethod   string
	requestBody     string
	requestHeaders  map[string]string
	response        string
	contentType     string
	statusCode      int
	statusText      string
	responseHeaders http.Header
	responseDuration time.Duration
	responseSize    int
	showHeaders     bool
	curlCopied      bool
	loading         bool
	err             error

	// Webhook
	webhookPortInput textinput.Model
	webhookCh        chan webhook.Request
	webhookStop      chan struct{}
	webhookRequests  []webhook.Request
	webhookCursor    int
	webhookDetail    bool
	webhookVP        viewport.Model
	webhookPort      int
	webhookErr       error

	// Environments
	envs          config.Environments
	envCursor     int
	envEditIdx    int
	envNameInput  textinput.Model
	envNameFocused bool
	envVarKV      []kvPair
	envVarCursor  int
	envVarEditing bool
	envVarFocused int
	envVarKeyIn   textinput.Model
	envVarValIn   textinput.Model
}

func New(history config.History, envs config.Environments) Model {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com/users  (or paste curl ...)"
	ti.CharLimit = 1000
	ti.Width = 60

	// Query Params
	qpKey := textinput.New()
	qpKey.Placeholder = "param"
	qpKey.Width = 22
	qpVal := textinput.New()
	qpVal.Placeholder = "value"
	qpVal.Width = 26

	// Headers
	hdrKey := textinput.New()
	hdrKey.Placeholder = "Header-Name"
	hdrKey.Width = 22
	hdrVal := textinput.New()
	hdrVal.Placeholder = "value"
	hdrVal.Width = 26

	// Auth
	tok := textinput.New()
	tok.Placeholder = "eyJhbGci..."
	tok.Width = 50
	tok.EchoMode = textinput.EchoPassword
	tok.EchoCharacter = '•'

	aUser := textinput.New()
	aUser.Placeholder = "username"
	aUser.Width = 30

	aPass := textinput.New()
	aPass.Placeholder = "password"
	aPass.Width = 30
	aPass.EchoMode = textinput.EchoPassword
	aPass.EchoCharacter = '•'

	apiName := textinput.New()
	apiName.Placeholder = "X-API-Key"
	apiName.Width = 22

	apiVal := textinput.New()
	apiVal.Placeholder = "sk-..."
	apiVal.Width = 30
	apiVal.EchoMode = textinput.EchoPassword
	apiVal.EchoCharacter = '•'

	// Body KV
	kvKey := textinput.New()
	kvKey.Placeholder = "key"
	kvKey.Width = 22
	kvVal := textinput.New()
	kvVal.Placeholder = "value"
	kvVal.Width = 26

	ta := textarea.New()
	ta.Placeholder = "{\n  \"key\": \"value\"\n}"
	ta.SetWidth(60)
	ta.SetHeight(8)
	ta.ShowLineNumbers = false

	// Webhook port
	port := textinput.New()
	port.Placeholder = "8080"
	port.SetValue("8080")
	port.CharLimit = 5
	port.Width = 10

	// Env editing
	envName := textinput.New()
	envName.Placeholder = "environment-name"
	envName.Width = 40

	envVarKey := textinput.New()
	envVarKey.Placeholder = "variable"
	envVarKey.Width = 22
	envVarVal := textinput.New()
	envVarVal.Placeholder = "value"
	envVarVal.Width = 26

	return Model{
		activeView:       menuView,
		methodIdx:        0,
		activeField:      fieldURL,
		historyIdx:       -1,
		bodyMode:         bodyModeKV,
		kvPairs:          []kvPair{},
		queryParams:      []kvPair{},
		headersKV:        []kvPair{},
		authKind:         authNone,
		authFocused:      -1,
		history:          history,
		envs:             envs,
		webhookPort:      8080,
		webhookPortInput: port,
		choices: []string{
			"Send API Request",
			"Start Webhook Catcher",
			"Environments",
			"Settings",
		},
		urlInput:        ti,
		qpKeyInput:      qpKey,
		qpValInput:      qpVal,
		hdrKeyInput:     hdrKey,
		hdrValInput:     hdrVal,
		tokenInput:      tok,
		authUserInput:   aUser,
		authPassInput:   aPass,
		apiKeyNameInput: apiName,
		apiKeyValInput:  apiVal,
		kvKeyInput:      kvKey,
		kvValInput:      kvVal,
		bodyInput:       ta,
		envNameInput:    envName,
		envVarKeyIn:     envVarKey,
		envVarValIn:     envVarVal,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
