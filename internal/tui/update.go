package tui

import (
	"encoding/base64"
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
		return m, nil
	case envsSavedMsg:
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
		m.responseHeaders = msg.response.Headers
		m.responseDuration = msg.response.Duration
		m.responseSize = msg.response.Size
		m.showHeaders = false
		m.curlCopied = false
		vp := viewport.New(m.width-horizontalOverhead, m.height-verticalOverhead)
		vp.SetContent(highlight(m.response, m.contentType))
		m.viewport = vp
		m.activeView = responseView
		return m, nil

	case webhookRequestMsg:
		req := webhook.Request(msg)
		m.webhookRequests = append([]webhook.Request{req}, m.webhookRequests...)
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
	case settingsView:
		return updateSettings(m, msg)
	case envListView:
		return updateEnvList(m, msg)
	case envEditView:
		return updateEnvEdit(m, msg)
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
		case 2:
			m.activeView = envListView
		case 3:
			m.activeView = settingsView
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
	m.queryParams = []kvPair{}
	m.qpCursor = 0
	m.qpEditing = false
	m.headersKV = []kvPair{}
	m.hdrCursor = 0
	m.hdrEditing = false
	m.bodyMode = bodyModeKV
	m.historyIdx = -1
	// Auth und Environment bleiben erhalten (session-übergreifend nützlich)
	return m, m.urlInput.Focus()
}

// ── Settings ──────────────────────────────────────────────────────────────────

func updateSettings(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "esc" || key.String() == "q" {
			m.activeView = menuView
		}
	}
	return m, nil
}

// ── Request — zentraler Router ────────────────────────────────────────────────

func updateRequest(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)

	if isKey {
		switch key.String() {
		case "ctrl+s":
			return doSend(m)
		case "ctrl+y": // curl in Clipboard kopieren
			return exportCurlFromRequest(m)
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
		case "f5":
			m.methodIdx = 4
			return m, nil
		}
	}

	switch m.activeField {
	case fieldURL:
		return updateURL(m, msg)
	case fieldParams:
		return updateParams(m, msg)
	case fieldHeaders:
		return updateHeaders(m, msg)
	case fieldAuth:
		return updateAuth(m, msg)
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
		m.activeView = menuView
		m.urlInput.Blur()
		return m, nil

	case "tab":
		m.activeField = fieldParams
		m.urlInput.Blur()
		return m, nil

	case "enter":
		urlVal := strings.TrimSpace(m.urlInput.Value())
		// curl-Import: wenn der User einen curl-Befehl einfügt, direkt parsen
		if strings.HasPrefix(strings.ToLower(urlVal), "curl ") {
			return importCurlStr(m, urlVal)
		}
		if hasBody(methods[m.methodIdx]) {
			// Enter → direkt zu Body (schneller Pfad, überspringt Params/Headers/Auth)
			m.activeField = fieldBody
			m.urlInput.Blur()
			return m, nil
		}
		return doSend(m)

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

// ── Query-Parameter KV-Builder ────────────────────────────────────────────────

func updateParams(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if !isKey {
		return forwardToQPInput(m, msg)
	}
	if m.qpEditing {
		return updateQPEditing(m, key)
	}
	return updateQPNavigating(m, key)
}

func updateQPNavigating(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.activeField = fieldURL
		return m, m.urlInput.Focus()
	case "tab":
		m.activeField = fieldHeaders
		return m, nil
	case "up", "k":
		if m.qpCursor > 0 {
			m.qpCursor--
		}
	case "down", "j":
		if m.qpCursor < len(m.queryParams)-1 {
			m.qpCursor++
		}
	case "n":
		m.queryParams = append(m.queryParams, kvPair{})
		m.qpCursor = len(m.queryParams) - 1
		return startQPEditing(m, 0)
	case "d":
		if len(m.queryParams) == 0 {
			return m, nil
		}
		m.queryParams = append(m.queryParams[:m.qpCursor], m.queryParams[m.qpCursor+1:]...)
		if m.qpCursor > 0 && m.qpCursor >= len(m.queryParams) {
			m.qpCursor = len(m.queryParams) - 1
		}
	case "enter":
		if len(m.queryParams) == 0 {
			m.queryParams = append(m.queryParams, kvPair{})
			m.qpCursor = 0
		}
		return startQPEditing(m, 0)
	}
	return m, nil
}

func updateQPEditing(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.qpEditing = false
		m.qpKeyInput.Blur()
		m.qpValInput.Blur()
		return m, nil
	case "tab":
		m = saveQPInput(m)
		if m.qpFocused == 0 {
			return startQPEditing(m, 1)
		}
		return startQPEditing(m, 0)
	case "enter":
		m = saveQPInput(m)
		if m.qpFocused == 0 {
			return startQPEditing(m, 1)
		}
		m.qpEditing = false
		m.qpKeyInput.Blur()
		m.qpValInput.Blur()
		return m, nil
	}
	return forwardToQPInput(m, key)
}

func saveQPInput(m Model) Model {
	if m.qpCursor < len(m.queryParams) {
		m.queryParams[m.qpCursor] = kvPair{key: m.qpKeyInput.Value(), value: m.qpValInput.Value()}
	}
	return m
}

func startQPEditing(m Model, focus int) (Model, tea.Cmd) {
	m.qpEditing = true
	m.qpFocused = focus
	if m.qpCursor < len(m.queryParams) {
		m.qpKeyInput.SetValue(m.queryParams[m.qpCursor].key)
		m.qpValInput.SetValue(m.queryParams[m.qpCursor].value)
	}
	if focus == 0 {
		m.qpValInput.Blur()
		return m, m.qpKeyInput.Focus()
	}
	m.qpKeyInput.Blur()
	return m, m.qpValInput.Focus()
}

func forwardToQPInput(m Model, msg tea.Msg) (Model, tea.Cmd) {
	if !m.qpEditing {
		return m, nil
	}
	var cmd tea.Cmd
	if m.qpFocused == 0 {
		m.qpKeyInput, cmd = m.qpKeyInput.Update(msg)
	} else {
		m.qpValInput, cmd = m.qpValInput.Update(msg)
	}
	return m, cmd
}

// ── Request-Headers KV-Builder ────────────────────────────────────────────────

func updateHeaders(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if !isKey {
		return forwardToHdrInput(m, msg)
	}
	if m.hdrEditing {
		return updateHdrEditing(m, key)
	}
	return updateHdrNavigating(m, key)
}

func updateHdrNavigating(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.activeField = fieldParams
		return m, nil
	case "tab":
		m.activeField = fieldAuth
		return enterAuthSection(m)
	case "up", "k":
		if m.hdrCursor > 0 {
			m.hdrCursor--
		}
	case "down", "j":
		if m.hdrCursor < len(m.headersKV)-1 {
			m.hdrCursor++
		}
	case "n":
		m.headersKV = append(m.headersKV, kvPair{})
		m.hdrCursor = len(m.headersKV) - 1
		return startHdrEditing(m, 0)
	case "d":
		if len(m.headersKV) == 0 {
			return m, nil
		}
		m.headersKV = append(m.headersKV[:m.hdrCursor], m.headersKV[m.hdrCursor+1:]...)
		if m.hdrCursor > 0 && m.hdrCursor >= len(m.headersKV) {
			m.hdrCursor = len(m.headersKV) - 1
		}
	case "enter":
		if len(m.headersKV) == 0 {
			m.headersKV = append(m.headersKV, kvPair{})
			m.hdrCursor = 0
		}
		return startHdrEditing(m, 0)
	}
	return m, nil
}

func updateHdrEditing(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.hdrEditing = false
		m.hdrKeyInput.Blur()
		m.hdrValInput.Blur()
		return m, nil
	case "tab":
		m = saveHdrInput(m)
		if m.hdrFocused == 0 {
			return startHdrEditing(m, 1)
		}
		return startHdrEditing(m, 0)
	case "enter":
		m = saveHdrInput(m)
		if m.hdrFocused == 0 {
			return startHdrEditing(m, 1)
		}
		m.hdrEditing = false
		m.hdrKeyInput.Blur()
		m.hdrValInput.Blur()
		return m, nil
	}
	return forwardToHdrInput(m, key)
}

func saveHdrInput(m Model) Model {
	if m.hdrCursor < len(m.headersKV) {
		m.headersKV[m.hdrCursor] = kvPair{key: m.hdrKeyInput.Value(), value: m.hdrValInput.Value()}
	}
	return m
}

func startHdrEditing(m Model, focus int) (Model, tea.Cmd) {
	m.hdrEditing = true
	m.hdrFocused = focus
	if m.hdrCursor < len(m.headersKV) {
		m.hdrKeyInput.SetValue(m.headersKV[m.hdrCursor].key)
		m.hdrValInput.SetValue(m.headersKV[m.hdrCursor].value)
	}
	if focus == 0 {
		m.hdrValInput.Blur()
		return m, m.hdrKeyInput.Focus()
	}
	m.hdrKeyInput.Blur()
	return m, m.hdrValInput.Focus()
}

func forwardToHdrInput(m Model, msg tea.Msg) (Model, tea.Cmd) {
	if !m.hdrEditing {
		return m, nil
	}
	var cmd tea.Cmd
	if m.hdrFocused == 0 {
		m.hdrKeyInput, cmd = m.hdrKeyInput.Update(msg)
	} else {
		m.hdrValInput, cmd = m.hdrValInput.Update(msg)
	}
	return m, cmd
}

// ── Auth ──────────────────────────────────────────────────────────────────────

func enterAuthSection(m Model) (Model, tea.Cmd) {
	switch m.authKind {
	case authBearer:
		m.authFocused = 0
		return m, m.tokenInput.Focus()
	case authBasic:
		m.authFocused = 0
		m.authPassInput.Blur()
		return m, m.authUserInput.Focus()
	case authAPIKey:
		m.authFocused = 0
		m.apiKeyValInput.Blur()
		return m, m.apiKeyNameInput.Focus()
	default:
		m.authFocused = -1
		return m, nil
	}
}

func blurAuthAll(m Model) Model {
	m.tokenInput.Blur()
	m.authUserInput.Blur()
	m.authPassInput.Blur()
	m.apiKeyNameInput.Blur()
	m.apiKeyValInput.Blur()
	m.authFocused = -1
	return m
}

func leaveAuthToNext(m Model) (Model, tea.Cmd) {
	m = blurAuthAll(m)
	if hasBody(methods[m.methodIdx]) {
		m.activeField = fieldBody
		return m, nil
	}
	m.activeField = fieldURL
	return m, m.urlInput.Focus()
}

func updateAuth(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if !isKey {
		return forwardToAuthInput(m, msg)
	}

	switch key.String() {
	case "esc":
		m = blurAuthAll(m)
		m.activeField = fieldHeaders
		return m, nil

	case "tab":
		switch m.authKind {
		case authNone:
			return leaveAuthToNext(m)
		case authBearer:
			m.tokenInput.Blur()
			return leaveAuthToNext(m)
		case authBasic:
			if m.authFocused == 0 {
				m.authUserInput.Blur()
				m.authFocused = 1
				return m, m.authPassInput.Focus()
			}
			m.authPassInput.Blur()
			return leaveAuthToNext(m)
		case authAPIKey:
			if m.authFocused == 0 {
				m.apiKeyNameInput.Blur()
				m.authFocused = 1
				return m, m.apiKeyValInput.Focus()
			}
			m.apiKeyValInput.Blur()
			return leaveAuthToNext(m)
		}

	// Auth-Typ wechseln
	case "1":
		m = blurAuthAll(m)
		m.authKind = authNone
		return m, nil
	case "2":
		m = blurAuthAll(m)
		m.authKind = authBearer
		m.authFocused = 0
		return m, m.tokenInput.Focus()
	case "3":
		m = blurAuthAll(m)
		m.authKind = authBasic
		m.authFocused = 0
		return m, m.authUserInput.Focus()
	case "4":
		m = blurAuthAll(m)
		m.authKind = authAPIKey
		m.authFocused = 0
		return m, m.apiKeyNameInput.Focus()
	}

	return forwardToAuthInput(m, msg)
}

func forwardToAuthInput(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.authKind {
	case authBearer:
		if m.authFocused == 0 {
			m.tokenInput, cmd = m.tokenInput.Update(msg)
		}
	case authBasic:
		switch m.authFocused {
		case 0:
			m.authUserInput, cmd = m.authUserInput.Update(msg)
		case 1:
			m.authPassInput, cmd = m.authPassInput.Update(msg)
		}
	case authAPIKey:
		switch m.authFocused {
		case 0:
			m.apiKeyNameInput, cmd = m.apiKeyNameInput.Update(msg)
		case 1:
			m.apiKeyValInput, cmd = m.apiKeyValInput.Update(msg)
		}
	}
	return m, cmd
}

// buildAuthHeader erzeugt den passenden Authorization-Header aus der Auth-Konfiguration.
func buildAuthHeader(m Model) (key, value string) {
	switch m.authKind {
	case authBearer:
		tok := strings.TrimSpace(m.tokenInput.Value())
		if tok != "" {
			return "Authorization", "Bearer " + tok
		}
	case authBasic:
		u := m.authUserInput.Value()
		p := m.authPassInput.Value()
		if u != "" || p != "" {
			enc := base64.StdEncoding.EncodeToString([]byte(u + ":" + p))
			return "Authorization", "Basic " + enc
		}
	case authAPIKey:
		name := strings.TrimSpace(m.apiKeyNameInput.Value())
		val := m.apiKeyValInput.Value()
		if name != "" {
			return name, val
		}
	}
	return "", ""
}

// ── Body KV-Builder ───────────────────────────────────────────────────────────

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

func updateKVNavigating(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc", "tab":
		m.activeField = fieldAuth
		return enterAuthSection(m)
	case "up", "k":
		if m.kvCursor > 0 {
			m.kvCursor--
		}
	case "down", "j":
		if m.kvCursor < len(m.kvPairs)-1 {
			m.kvCursor++
		}
	case "enter":
		if len(m.kvPairs) == 0 {
			m.kvPairs = append(m.kvPairs, kvPair{})
			m.kvCursor = 0
		}
		return startEditing(m, 0)
	case "n":
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

func updateKVEditing(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.kvEditing = false
		m.kvKeyInput.Blur()
		m.kvValInput.Blur()
		return m, nil
	case "tab":
		m = saveKVInput(m)
		if m.kvFocused == 0 {
			return startEditing(m, 1)
		}
		return startEditing(m, 0)
	case "enter":
		m = saveKVInput(m)
		if m.kvFocused == 0 {
			return startEditing(m, 1)
		}
		m.kvEditing = false
		m.kvKeyInput.Blur()
		m.kvValInput.Blur()
		return m, nil
	}
	return forwardToKVInput(m, key)
}

func saveKVInput(m Model) Model {
	if m.kvCursor < len(m.kvPairs) {
		m.kvPairs[m.kvCursor] = kvPair{key: m.kvKeyInput.Value(), value: m.kvValInput.Value()}
	}
	return m
}

func startEditing(m Model, focus int) (Model, tea.Cmd) {
	m.kvEditing = true
	m.kvFocused = focus
	if m.kvCursor < len(m.kvPairs) {
		m.kvKeyInput.SetValue(m.kvPairs[m.kvCursor].key)
		m.kvValInput.SetValue(m.kvPairs[m.kvCursor].value)
	}
	if focus == 0 {
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

func updateBodyRaw(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if isKey && key.String() == "esc" {
		m.bodyMode = bodyModeKV
		if parsed := jsonToKV(m.bodyInput.Value()); parsed != nil {
			m.kvPairs = parsed
		}
		m.bodyInput.Blur()
		m.activeField = fieldAuth
		return enterAuthSection(m)
	}
	var cmd tea.Cmd
	m.bodyInput, cmd = m.bodyInput.Update(msg)
	return m, cmd
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func doSend(m Model) (Model, tea.Cmd) {
	rawURL := strings.TrimSpace(m.urlInput.Value())
	if rawURL == "" {
		return m, nil
	}

	vars := m.envs.VarMap()
	method := methods[m.methodIdx]

	// Variablen substituieren und Query-Parameter einbetten
	resolvedURL := buildFinalURL(applyVars(rawURL, vars), m.queryParams, vars)

	body := ""
	if hasBody(method) {
		if m.bodyMode == bodyModeKV {
			body = kvToJSON(m.kvPairs)
		} else {
			body = m.bodyInput.Value()
		}
		body = applyVars(body, vars)
	}

	// Headers aufbauen (manuell + Auth, Auth hat Vorrang)
	headers := make(map[string]string, len(m.headersKV)+1)
	for _, pair := range m.headersKV {
		if pair.key != "" {
			headers[applyVars(pair.key, vars)] = applyVars(pair.value, vars)
		}
	}
	if k, v := buildAuthHeader(m); k != "" {
		headers[k] = v
	}

	m.history.Add(config.HistoryEntry{URL: rawURL, Method: method, Body: body})
	m.historyIdx = -1
	m.requestURL = resolvedURL
	m.requestMethod = method
	m.requestBody = body
	m.requestHeaders = headers
	m.loading = true
	m.err = nil
	m.curlCopied = false

	return m, tea.Batch(
		sendRequest(method, resolvedURL, body, headers),
		saveHistoryCmd(m.history),
	)
}

func restoreHistoryEntry(m Model, entry config.HistoryEntry) (Model, tea.Cmd) {
	m.urlInput.SetValue(entry.URL)
	for i, method := range methods {
		if method == entry.Method {
			m.methodIdx = i
			break
		}
	}
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

// importCurlStr parst einen curl-Befehl und befüllt den Request-View.
func importCurlStr(m Model, input string) (Model, tea.Cmd) {
	method, rawURL, body, hdrs, err := parseCurlCmd(input)
	if err != nil {
		return m, nil // Fehler still ignorieren
	}

	m.urlInput.SetValue(rawURL)
	for i, meth := range methods {
		if meth == method {
			m.methodIdx = i
			break
		}
	}

	m.headersKV = make([]kvPair, 0, len(hdrs))
	for k, v := range hdrs {
		m.headersKV = append(m.headersKV, kvPair{key: k, value: v})
	}

	if body != "" {
		if pairs := jsonToKV(body); pairs != nil {
			m.kvPairs = pairs
			m.bodyMode = bodyModeKV
		} else {
			m.bodyInput.SetValue(body)
			m.bodyMode = bodyModeRaw
			m.kvPairs = []kvPair{}
		}
	}

	return m, nil
}

// exportCurlFromRequest kopiert einen curl-Befehl des aktuellen Request-Zustands.
func exportCurlFromRequest(m Model) (Model, tea.Cmd) {
	vars := m.envs.VarMap()
	method := methods[m.methodIdx]
	rawURL := applyVars(strings.TrimSpace(m.urlInput.Value()), vars)
	finalURL := buildFinalURL(rawURL, m.queryParams, vars)

	body := ""
	if hasBody(method) {
		if m.bodyMode == bodyModeKV {
			body = kvToJSON(m.kvPairs)
		} else {
			body = m.bodyInput.Value()
		}
		body = applyVars(body, vars)
	}

	headers := make(map[string]string)
	for _, pair := range m.headersKV {
		if pair.key != "" {
			headers[applyVars(pair.key, vars)] = applyVars(pair.value, vars)
		}
	}
	if k, v := buildAuthHeader(m); k != "" {
		headers[k] = v
	}

	curl := buildCurlCmd(method, finalURL, body, headers)
	_ = copyToClipboard(curl)
	m.curlCopied = true
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

func updateWebhookLive(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if isKey {
		if m.webhookDetail {
			switch key.String() {
			case "r":
				if len(m.webhookRequests) > 0 {
					return replayWebhook(m, m.webhookRequests[m.webhookCursor])
				}
			case "esc", "q":
				m.webhookDetail = false
				return m, nil
			}
			var cmd tea.Cmd
			m.webhookVP, cmd = m.webhookVP.Update(msg)
			return m, cmd
		}

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
		case "r":
			if len(m.webhookRequests) > 0 {
				return replayWebhook(m, m.webhookRequests[m.webhookCursor])
			}
		case "s", "esc":
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

func replayWebhook(m Model, req webhook.Request) (Model, tea.Cmd) {
	m.activeView = requestView
	m.activeField = fieldURL
	m.urlInput.SetValue(req.Path)
	m.historyIdx = -1

	for i, method := range methods {
		if method == req.Method {
			m.methodIdx = i
			break
		}
	}

	if req.Body != "" {
		if pairs := jsonToKV(req.Body); pairs != nil {
			m.kvPairs = pairs
			m.bodyMode = bodyModeKV
		} else {
			m.bodyInput.SetValue(req.Body)
			m.bodyMode = bodyModeRaw
			m.kvPairs = []kvPair{}
		}
	} else {
		m.kvPairs = []kvPair{}
		m.bodyMode = bodyModeKV
	}

	m.kvEditing = false
	m.headersKV = []kvPair{}
	m.hdrEditing = false
	m.queryParams = []kvPair{}
	m.qpEditing = false

	return m, m.urlInput.Focus()
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
			m.responseHeaders = nil
			m.showHeaders = false
			m.curlCopied = false
			return m, nil
		case "e":
			m.activeView = requestView
			m.activeField = fieldURL
			m.urlInput.SetValue(m.requestURL)
			m.historyIdx = -1
			return m, m.urlInput.Focus()
		case "h":
			m.showHeaders = !m.showHeaders
			return m, nil
		case "c": // curl in Clipboard kopieren
			curl := buildCurlCmd(m.requestMethod, m.requestURL, m.requestBody, m.requestHeaders)
			_ = copyToClipboard(curl)
			m.curlCopied = true
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// ── Environments ──────────────────────────────────────────────────────────────

func updateEnvList(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc", "q":
		m.activeView = menuView
	case "up", "k":
		if m.envCursor > 0 {
			m.envCursor--
		}
	case "down", "j":
		if m.envCursor < len(m.envs.Envs)-1 {
			m.envCursor++
		}
	case "enter":
		if len(m.envs.Envs) > 0 {
			m.envs.Active = m.envCursor
			return m, saveEnvsCmd(m.envs)
		}
	case "e":
		if len(m.envs.Envs) > 0 {
			return enterEnvEdit(m, m.envCursor)
		}
	case "n":
		m.envs.Envs = append(m.envs.Envs, config.Environment{Name: "new-env"})
		m.envCursor = len(m.envs.Envs) - 1
		return enterEnvEdit(m, m.envCursor)
	case "d":
		if len(m.envs.Envs) == 0 {
			return m, nil
		}
		m.envs.Envs = append(m.envs.Envs[:m.envCursor], m.envs.Envs[m.envCursor+1:]...)
		if m.envs.Active >= len(m.envs.Envs) {
			m.envs.Active = len(m.envs.Envs) - 1
		}
		if m.envCursor > 0 && m.envCursor >= len(m.envs.Envs) {
			m.envCursor--
		}
		return m, saveEnvsCmd(m.envs)
	}
	return m, nil
}

func enterEnvEdit(m Model, idx int) (Model, tea.Cmd) {
	m.activeView = envEditView
	m.envEditIdx = idx
	env := m.envs.Envs[idx]

	m.envNameInput.SetValue(env.Name)
	m.envVarKV = make([]kvPair, len(env.Vars))
	for i, v := range env.Vars {
		m.envVarKV[i] = kvPair{key: v.Key, value: v.Value}
	}
	m.envVarCursor = 0
	m.envVarEditing = false
	m.envNameFocused = true
	return m, m.envNameInput.Focus()
}

func updateEnvEdit(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)

	if m.envNameFocused {
		if isKey {
			switch key.String() {
			case "esc":
				return saveAndLeaveEnvEdit(m)
			case "tab", "enter":
				m.envNameFocused = false
				m.envNameInput.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.envNameInput, cmd = m.envNameInput.Update(msg)
		return m, cmd
	}

	// Variablen-KV-Builder
	if !isKey {
		return forwardToEnvVarInput(m, msg)
	}
	if m.envVarEditing {
		return updateEnvVarEditing(m, key)
	}
	return updateEnvVarNavigating(m, key)
}

func updateEnvVarNavigating(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		return saveAndLeaveEnvEdit(m)
	case "up", "k":
		if m.envVarCursor > 0 {
			m.envVarCursor--
		}
	case "down", "j":
		if m.envVarCursor < len(m.envVarKV)-1 {
			m.envVarCursor++
		}
	case "n":
		m.envVarKV = append(m.envVarKV, kvPair{})
		m.envVarCursor = len(m.envVarKV) - 1
		return startEnvVarEdit(m, 0)
	case "d":
		if len(m.envVarKV) == 0 {
			return m, nil
		}
		m.envVarKV = append(m.envVarKV[:m.envVarCursor], m.envVarKV[m.envVarCursor+1:]...)
		if m.envVarCursor > 0 && m.envVarCursor >= len(m.envVarKV) {
			m.envVarCursor--
		}
	case "enter":
		if len(m.envVarKV) == 0 {
			m.envVarKV = append(m.envVarKV, kvPair{})
			m.envVarCursor = 0
		}
		return startEnvVarEdit(m, 0)
	case "e":
		m.envNameFocused = true
		return m, m.envNameInput.Focus()
	}
	return m, nil
}

func updateEnvVarEditing(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.envVarEditing = false
		m.envVarKeyIn.Blur()
		m.envVarValIn.Blur()
		return m, nil
	case "tab":
		m = saveEnvVarInput(m)
		if m.envVarFocused == 0 {
			return startEnvVarEdit(m, 1)
		}
		return startEnvVarEdit(m, 0)
	case "enter":
		m = saveEnvVarInput(m)
		if m.envVarFocused == 0 {
			return startEnvVarEdit(m, 1)
		}
		m.envVarEditing = false
		m.envVarKeyIn.Blur()
		m.envVarValIn.Blur()
		return m, nil
	}
	return forwardToEnvVarInput(m, key)
}

func saveEnvVarInput(m Model) Model {
	if m.envVarCursor < len(m.envVarKV) {
		m.envVarKV[m.envVarCursor] = kvPair{key: m.envVarKeyIn.Value(), value: m.envVarValIn.Value()}
	}
	return m
}

func startEnvVarEdit(m Model, focus int) (Model, tea.Cmd) {
	m.envVarEditing = true
	m.envVarFocused = focus
	if m.envVarCursor < len(m.envVarKV) {
		m.envVarKeyIn.SetValue(m.envVarKV[m.envVarCursor].key)
		m.envVarValIn.SetValue(m.envVarKV[m.envVarCursor].value)
	}
	if focus == 0 {
		m.envVarValIn.Blur()
		return m, m.envVarKeyIn.Focus()
	}
	m.envVarKeyIn.Blur()
	return m, m.envVarValIn.Focus()
}

func forwardToEnvVarInput(m Model, msg tea.Msg) (Model, tea.Cmd) {
	if !m.envVarEditing {
		return m, nil
	}
	var cmd tea.Cmd
	if m.envVarFocused == 0 {
		m.envVarKeyIn, cmd = m.envVarKeyIn.Update(msg)
	} else {
		m.envVarValIn, cmd = m.envVarValIn.Update(msg)
	}
	return m, cmd
}

func saveAndLeaveEnvEdit(m Model) (Model, tea.Cmd) {
	if m.envVarEditing {
		m = saveEnvVarInput(m)
		m.envVarEditing = false
		m.envVarKeyIn.Blur()
		m.envVarValIn.Blur()
	}

	if m.envEditIdx < len(m.envs.Envs) {
		name := strings.TrimSpace(m.envNameInput.Value())
		if name == "" {
			name = "unnamed"
		}
		vars := make([]config.EnvVar, 0, len(m.envVarKV))
		for _, pair := range m.envVarKV {
			if pair.key != "" {
				vars = append(vars, config.EnvVar{Key: pair.key, Value: pair.value})
			}
		}
		m.envs.Envs[m.envEditIdx].Name = name
		m.envs.Envs[m.envEditIdx].Vars = vars
	}

	m.activeView = envListView
	m.envNameInput.Blur()
	return m, saveEnvsCmd(m.envs)
}
