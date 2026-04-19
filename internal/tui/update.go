package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/config"
	"github.com/PsydoV2/Apish/internal/webhook"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.activeView == responseView {
			m.viewport.Width = msg.Width - horizontalOverhead
			m.viewport.Height = msg.Height - verticalOverhead
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case historySavedMsg:
		// Fehler beim Speichern still ignorieren — History ist nice-to-have
		return m, nil
	case responseMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.activeView = responseView
			return m, nil
		}
		m.response = msg.response.Body
		m.contentType = msg.response.ContentType
		m.statusCode = msg.response.StatusCode
		m.statusText = msg.response.Status
		vp := viewport.New(m.width-horizontalOverhead, m.height-verticalOverhead)
		vp.SetContent(highlight(m.response, m.contentType))
		m.viewport = vp
		m.activeView = responseView
		return m, nil

	case webhookRequestMsg:
		// Newest request prepended so index 0 is always the latest
		req := webhook.Request(msg)
		m.webhookRequests = append([]webhook.Request{req}, m.webhookRequests...)
		// Re-issue the wait cmd — self-renewing event loop
		return m, waitForWebhookCmd(m.webhookCh, m.webhookStop)

	case webhookServerDoneMsg:
		if msg.err != nil {
			m.webhookErr = msg.err
		}
		m.activeView = menuView
		return m, nil
	}

	switch m.activeView {
	case menuView:
		return updateMenu(m, msg)
	case requestView:
		return updateRequest(m, msg)
	case responseView:
		return updateResponse(m, msg)
	case webhookSetupView:
		return updateWebhookSetup(m, msg)
	case webhookLiveView:
		return updateWebhookLive(m, msg)
	}
	return m, nil
}

// ── Menu ──────────────────────────────────────────────────────────────────────

func updateMenu(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "enter":
		switch m.cursor {
		case 0:
			return enterRequestView(m)
		case 1:
			return enterWebhookSetup(m)
		}
	}
	return m, nil
}

func enterRequestView(m Model) (Model, tea.Cmd) {
	m.activeView = requestView
	m.activeField = fieldURL
	m.urlInput.SetValue("")
	m.bodyInput.SetValue("")
	m.kvPairs = []kvPair{}
	m.kvCursor = 0
	m.kvEditing = false
	m.bodyMode = bodyModeKV
	m.historyIdx = -1
	return m, m.urlInput.Focus()
}

// ── Request — zentraler Router ────────────────────────────────────────────────
// Jedes Feld verwaltet Tab und Esc selbst — kein globales Abfangen.

func updateRequest(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)

	// Globale Keys innerhalb der Request-View
	if isKey {
		switch key.String() {
		case "ctrl+s":
			return doSend(m)
		case "f1":
			m.methodIdx = 0
			return m, nil
		case "f2":
			m.methodIdx = 1
			return m, nil
		case "f3":
			m.methodIdx = 2
			return m, nil
		case "f4":
			m.methodIdx = 3
			return m, nil
		}
	}

	switch m.activeField {
	case fieldURL:
		return updateURL(m, msg)
	case fieldBody:
		if m.bodyMode == bodyModeRaw {
			return updateBodyRaw(m, msg)
		}
		return updateBodyKV(m, msg)
	}
	return m, nil
}

// ── URL-Feld ──────────────────────────────────────────────────────────────────

func updateURL(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if !isKey {
		var cmd tea.Cmd
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd
	}

	switch key.String() {
	case "esc":
		// URL → Menü
		m.activeView = menuView
		m.urlInput.Blur()
		return m, nil

	case "tab", "enter":
		if hasBody(methods[m.methodIdx]) {
			// URL → Body
			m.activeField = fieldBody
			m.urlInput.Blur()
			return m, nil
		}
		if key.String() == "enter" {
			return doSend(m)
		}

	case "up":
		entries := m.history.Entries
		if len(entries) == 0 {
			return m, nil
		}
		if m.historyIdx == -1 {
			m.historyIdx = len(entries) - 1
		} else if m.historyIdx > 0 {
			m.historyIdx--
		}
		return restoreHistoryEntry(m, entries[m.historyIdx])

	case "down":
		entries := m.history.Entries
		if m.historyIdx == -1 {
			return m, nil
		}
		if m.historyIdx < len(entries)-1 {
			m.historyIdx++
			return restoreHistoryEntry(m, entries[m.historyIdx])
		}
		// Ende der History: alles leeren
		m.historyIdx = -1
		m.urlInput.SetValue("")
		m.methodIdx = 0
		m.kvPairs = []kvPair{}
		m.bodyInput.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.urlInput, cmd = m.urlInput.Update(msg)
	return m, cmd
}

// ── KV-Builder ────────────────────────────────────────────────────────────────

func updateBodyKV(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)

	if !isKey {
		return forwardToKVInput(m, msg)
	}

	if m.kvEditing {
		return updateKVEditing(m, key)
	}
	return updateKVNavigating(m, key)
}

// updateKVNavigating: kein Feld hat Fokus, User bewegt den Cursor
func updateKVNavigating(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc", "tab":
		// Body → zurück zu URL
		m.activeField = fieldURL
		return m, m.urlInput.Focus()

	case "up", "k":
		if m.kvCursor > 0 {
			m.kvCursor--
		}

	case "down", "j":
		if m.kvCursor < len(m.kvPairs)-1 {
			m.kvCursor++
		}

	case "enter":
		// Ausgewähltes Paar editieren (oder erstes anlegen)
		if len(m.kvPairs) == 0 {
			m.kvPairs = append(m.kvPairs, kvPair{})
			m.kvCursor = 0
		}
		return startEditing(m, 0)

	case "n":
		// Neues Paar am Ende, sofort editieren
		m.kvPairs = append(m.kvPairs, kvPair{})
		m.kvCursor = len(m.kvPairs) - 1
		return startEditing(m, 0)

	case "d":
		if len(m.kvPairs) == 0 {
			return m, nil
		}
		m.kvPairs = append(m.kvPairs[:m.kvCursor], m.kvPairs[m.kvCursor+1:]...)
		if m.kvCursor > 0 && m.kvCursor >= len(m.kvPairs) {
			m.kvCursor = len(m.kvPairs) - 1
		}

	case "r":
		m.bodyMode = bodyModeRaw
		m.bodyInput.SetValue(kvToJSON(m.kvPairs))
		return m, m.bodyInput.Focus()
	}

	return m, nil
}

// updateKVEditing: ein Paar wird gerade bearbeitet
func updateKVEditing(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		// Editierung abbrechen ohne zu speichern
		m.kvEditing = false
		m.kvKeyInput.Blur()
		m.kvValInput.Blur()
		return m, nil

	case "tab":
		// Key ↔ Value wechseln (in beiden Richtungen)
		m = saveKVInput(m)
		if m.kvFocused == 0 {
			return startEditing(m, 1) // Key → Value
		}
		return startEditing(m, 0) // Value → Key

	case "enter":
		m = saveKVInput(m)
		if m.kvFocused == 0 {
			// Key bestätigt → weiter zu Value
			return startEditing(m, 1)
		}
		// Value bestätigt → Editierung abschließen
		m.kvEditing = false
		m.kvKeyInput.Blur()
		m.kvValInput.Blur()
		return m, nil
	}

	return forwardToKVInput(m, key)
}

func saveKVInput(m Model) Model {
	if m.kvCursor < len(m.kvPairs) {
		m.kvPairs[m.kvCursor] = kvPair{
			key:   m.kvKeyInput.Value(),
			value: m.kvValInput.Value(),
		}
	}
	return m
}

func startEditing(m Model, focusField int) (Model, tea.Cmd) {
	m.kvEditing = true
	m.kvFocused = focusField

	if m.kvCursor < len(m.kvPairs) {
		m.kvKeyInput.SetValue(m.kvPairs[m.kvCursor].key)
		m.kvValInput.SetValue(m.kvPairs[m.kvCursor].value)
	}

	if focusField == 0 {
		m.kvValInput.Blur()
		return m, m.kvKeyInput.Focus()
	}
	m.kvKeyInput.Blur()
	return m, m.kvValInput.Focus()
}

func forwardToKVInput(m Model, msg tea.Msg) (Model, tea.Cmd) {
	if !m.kvEditing {
		return m, nil
	}
	var cmd tea.Cmd
	if m.kvFocused == 0 {
		m.kvKeyInput, cmd = m.kvKeyInput.Update(msg)
	} else {
		m.kvValInput, cmd = m.kvValInput.Update(msg)
	}
	return m, cmd
}

// ── Raw-Textarea ──────────────────────────────────────────────────────────────

func updateBodyRaw(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if isKey {
		switch key.String() {
		case "esc":
			// Raw → zurück zu URL
			m.bodyMode = bodyModeKV
			if parsed := jsonToKV(m.bodyInput.Value()); parsed != nil {
				m.kvPairs = parsed
			}
			m.bodyInput.Blur()
			m.activeField = fieldURL
			return m, m.urlInput.Focus()
		}
	}
	var cmd tea.Cmd
	m.bodyInput, cmd = m.bodyInput.Update(msg)
	return m, cmd
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func doSend(m Model) (Model, tea.Cmd) {
	url := m.urlInput.Value()
	if url == "" {
		return m, nil
	}

	method := methods[m.methodIdx]
	body := ""
	if hasBody(method) {
		if m.bodyMode == bodyModeKV {
			body = kvToJSON(m.kvPairs)
		} else {
			body = m.bodyInput.Value()
		}
	}

	// History-Eintrag anlegen und asynchron speichern
	m.history.Add(config.HistoryEntry{URL: url, Method: method, Body: body})
	m.historyIdx = -1
	m.requestURL = url
	m.requestMethod = method
	m.loading = true
	m.err = nil

	// tea.Batch führt mehrere Commands gleichzeitig aus
	return m, tea.Batch(
		sendRequest(method, url, body),
		saveHistoryCmd(m.history),
	)
}

// restoreHistoryEntry lädt einen History-Eintrag vollständig in den Model-State.
// URL, Method und Body werden wiederhergestellt.
func restoreHistoryEntry(m Model, entry config.HistoryEntry) (Model, tea.Cmd) {
	m.urlInput.SetValue(entry.URL)

	// Methoden-Index aus dem gespeicherten String wiederherstellen
	for i, method := range methods {
		if method == entry.Method {
			m.methodIdx = i
			break
		}
	}

	// Body wiederherstellen: JSON → KV-Paare versuchen, sonst Raw-Textarea
	if entry.Body != "" {
		if pairs := jsonToKV(entry.Body); pairs != nil {
			m.kvPairs = pairs
			m.bodyMode = bodyModeKV
		} else {
			m.bodyInput.SetValue(entry.Body)
			m.bodyMode = bodyModeRaw
		}
	} else {
		m.kvPairs = []kvPair{}
		m.bodyMode = bodyModeKV
	}

	return m, nil
}

// ── Webhook Setup ─────────────────────────────────────────────────────────────

func enterWebhookSetup(m Model) (Model, tea.Cmd) {
	m.activeView = webhookSetupView
	m.webhookRequests = nil
	m.webhookCursor = 0
	m.webhookDetail = false
	m.webhookErr = nil
	return m, m.webhookPortInput.Focus()
}

func updateWebhookSetup(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if isKey {
		switch key.String() {
		case "esc":
			m.activeView = menuView
			m.webhookPortInput.Blur()
			return m, nil
		case "enter":
			port := parsePort(m.webhookPortInput.Value())
			m.webhookPort = port
			m.webhookCh = make(chan webhook.Request, 64)
			m.webhookStop = make(chan struct{})
			m.webhookVP = viewport.New(m.width-horizontalOverhead, m.height-verticalOverhead)
			m.activeView = webhookLiveView
			m.webhookPortInput.Blur()
			return m, tea.Batch(
				startWebhookServerCmd(port, m.webhookCh, m.webhookStop),
				waitForWebhookCmd(m.webhookCh, m.webhookStop),
			)
		}
	}
	var cmd tea.Cmd
	m.webhookPortInput, cmd = m.webhookPortInput.Update(msg)
	return m, cmd
}

// ── Webhook Live ──────────────────────────────────────────────────────────────

func updateWebhookLive(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if isKey {
		// Detail view: scroll viewport or go back to list
		if m.webhookDetail {
			switch key.String() {
			case "esc", "q":
				m.webhookDetail = false
				return m, nil
			}
			var cmd tea.Cmd
			m.webhookVP, cmd = m.webhookVP.Update(msg)
			return m, cmd
		}

		// List view
		switch key.String() {
		case "up", "k":
			if m.webhookCursor > 0 {
				m.webhookCursor--
			}
		case "down", "j":
			if m.webhookCursor < len(m.webhookRequests)-1 {
				m.webhookCursor++
			}
		case "enter":
			if len(m.webhookRequests) > 0 {
				content := renderWebhookDetail(m.webhookRequests[m.webhookCursor])
				m.webhookVP = viewport.New(m.width-horizontalOverhead, m.height-verticalOverhead)
				m.webhookVP.SetContent(content)
				m.webhookDetail = true
			}
		case "s", "esc":
			// Shut down: close the stop channel — unblocks both the server and the wait cmd
			if m.webhookStop != nil {
				close(m.webhookStop)
				m.webhookStop = nil
			}
			m.activeView = menuView
			return m, nil
		}
	}
	return m, nil
}

func parsePort(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 || n > 65535 {
		return 8080
	}
	return n
}

// ── Response ──────────────────────────────────────────────────────────────────

func updateResponse(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "q":
			m.activeView = menuView
			m.response, m.contentType = "", ""
			m.statusCode, m.statusText = 0, ""
			m.err, m.requestURL, m.requestMethod = nil, "", ""
			return m, nil
		case "e":
			m.activeView = requestView
			m.activeField = fieldURL
			m.urlInput.SetValue(m.requestURL)
			m.historyIdx = -1
			return m, m.urlInput.Focus()
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}
