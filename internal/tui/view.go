package tui

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/PsydoV2/Apish/internal/webhook"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	switch m.activeView {
	case requestView:
		return appStyle.Render(viewRequest(m))
	case responseView:
		return appStyle.Render(viewResponse(m))
	case webhookSetupView:
		return appStyle.Render(viewWebhookSetup(m))
	case webhookLiveView:
		return appStyle.Render(viewWebhookLive(m))
	case settingsView:
		return appStyle.Render(viewSettings(m))
	case envListView:
		return appStyle.Render(viewEnvList(m))
	case envEditView:
		return appStyle.Render(viewEnvEdit(m))
	default:
		return appStyle.Render(viewMenu(m))
	}
}

// ── Layout helpers ────────────────────────────────────────────────────────────

func appHeader() string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		titleStyle.Render("apish"),
		subtitleStyle.Render("  ·  Terminal API Client"),
	)
}

func divider(width int) string {
	if width < 1 {
		return ""
	}
	return subtitleStyle.Render(strings.Repeat("─", width))
}

func keyBind(key, desc string) string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		keyStyle.Render(key),
		" ",
		keyDescStyle.Render(desc),
		"   ",
	)
}

func wrapKeyBar(maxWidth int, bindings ...string) string {
	var rows [][]string
	var current []string
	currentW := 0
	for _, b := range bindings {
		w := lipgloss.Width(b)
		if currentW+w > maxWidth && len(current) > 0 {
			rows = append(rows, current)
			current = []string{b}
			currentW = w
		} else {
			current = append(current, b)
			currentW += w
		}
	}
	if len(current) > 0 {
		rows = append(rows, current)
	}
	lines := make([]string, len(rows))
	for i, row := range rows {
		lines[i] = lipgloss.JoinHorizontal(lipgloss.Center, row...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func methodSelector(m Model) string {
	badges := make([]string, len(methods))
	for i, method := range methods {
		badges[i] = methodBadge(method, i == m.methodIdx)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, badges...)
}

// ── Menu ──────────────────────────────────────────────────────────────────────

func viewMenu(m Model) string {
	var items strings.Builder
	for i, choice := range m.choices {
		if m.cursor == i {
			items.WriteString(cursorStyle.Render("▸") + "  " + selectedItemStyle.Render(choice) + "\n")
		} else {
			items.WriteString("   " + normalItemStyle.Render(choice) + "\n")
		}
	}

	available := m.width - horizontalOverhead

	rows := []string{appHeader(), ""}

	if name := m.envs.ActiveName(); name != "" {
		rows = append(rows, envBadgeStyle.Render("ENV  "+name), "")
	}

	rows = append(rows,
		strings.TrimRight(items.String(), "\n"),
		"",
		divider(available),
		wrapKeyBar(available,
			keyBind("j/k", "navigate"),
			keyBind("Enter", "select"),
			keyBind("q", "quit"),
		),
	)

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// ── Settings ──────────────────────────────────────────────────────────────────

func viewSettings(m Model) string {
	available := m.width - horizontalOverhead
	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		appHeader(),
		"",
		titleStyle.Render("Settings"),
		"",
		subtitleStyle.Render("Coming soon:"),
		"",
		normalItemStyle.Render("  • Default request timeout"),
		normalItemStyle.Render("  • Max history size"),
		normalItemStyle.Render("  • Theme"),
		"",
		divider(available),
		wrapKeyBar(available, keyBind("Esc", "back")),
	))
}

// ── Request — Accordion-Layout ────────────────────────────────────────────────

func viewRequest(m Model) string {
	if m.loading {
		return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("API Request"),
			"",
			loadingStyle.Render("  ⋯  Sending request..."),
		))
	}

	method := methods[m.methodIdx]
	available := m.width - horizontalOverhead

	// Titelzeile mit Env-Badge und curl-Feedback
	titleParts := []string{titleStyle.Render("API Request")}
	if name := m.envs.ActiveName(); name != "" {
		titleParts = append(titleParts, "  ", envBadgeStyle.Render("ENV  "+name))
	}
	if m.curlCopied {
		titleParts = append(titleParts, "  ", successStyle.Render("✓ curl copied"))
	}
	titleRow := lipgloss.JoinHorizontal(lipgloss.Left, titleParts...)

	rows := []string{
		titleRow,
		"",
		lipgloss.JoinHorizontal(lipgloss.Center,
			labelStyle.Render("Method   "),
			methodSelector(m),
			subtitleStyle.Render("   F1–F5"),
		),
		"",
	}

	// ── URL ──
	if m.activeField == fieldURL {
		rows = append(rows,
			expandedSectionHeader("URL"),
			"",
			"      "+m.urlInput.View(),
		)
	} else {
		url := m.urlInput.Value()
		if url == "" {
			url = "(leer)"
		}
		if len(url) > 52 {
			url = url[:49] + "..."
		}
		rows = append(rows, collapsedSection("URL", url))
	}
	rows = append(rows, "")

	// ── Query Params ──
	qpSum := paramsSummary(m.queryParams)
	if m.activeField == fieldParams {
		rows = append(rows,
			expandedSectionHeader("Params"),
			"",
			viewGenericKV(m.queryParams, m.qpCursor, m.qpEditing, m.qpKeyInput, m.qpValInput, "PARAM", "VALUE"),
		)
	} else {
		rows = append(rows, collapsedSection("Params", qpSum))
	}
	rows = append(rows, "")

	// ── Headers ──
	hdrSum := kvCountSummary(m.headersKV, "header")
	if m.activeField == fieldHeaders {
		rows = append(rows,
			expandedSectionHeader("Headers"),
			"",
			viewGenericKV(m.headersKV, m.hdrCursor, m.hdrEditing, m.hdrKeyInput, m.hdrValInput, "HEADER", "VALUE"),
		)
	} else {
		rows = append(rows, collapsedSection("Headers", hdrSum))
	}
	rows = append(rows, "")

	// ── Auth ──
	if m.activeField == fieldAuth {
		rows = append(rows,
			expandedSectionHeader("Auth"),
			"",
			viewAuthSection(m),
		)
	} else {
		rows = append(rows, collapsedSection("Auth", authSummary(m)))
	}
	rows = append(rows, "")

	// ── Body (nur bei Methoden mit Body) ──
	if hasBody(method) {
		if m.activeField == fieldBody {
			rows = append(rows, expandedSectionHeader("Body"), "")
			if m.bodyMode == bodyModeRaw {
				rows = append(rows, m.bodyInput.View(), "", subtitleStyle.Render("      Esc  → auth"))
			} else {
				rows = append(rows, viewGenericKV(m.kvPairs, m.kvCursor, m.kvEditing, m.kvKeyInput, m.kvValInput, "KEY", "VALUE"))
			}
		} else {
			rows = append(rows, collapsedSection("Body", bodySummaryStr(m)))
		}
		rows = append(rows, "")
	}

	rows = append(rows, divider(available), viewRequestKeyBar(m, method, available))
	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// ── Section summary helpers ───────────────────────────────────────────────────

func paramsSummary(params []kvPair) string {
	n := 0
	for _, p := range params {
		if p.key != "" {
			n++
		}
	}
	if n == 0 {
		return "(none)"
	}
	return fmt.Sprintf("%d param(s)", n)
}

func kvCountSummary(pairs []kvPair, noun string) string {
	n := len(pairs)
	if n == 0 {
		return "(none)"
	}
	return fmt.Sprintf("%d %s(s)", n, noun)
}

func authSummary(m Model) string {
	switch m.authKind {
	case authBearer:
		return "Bearer"
	case authBasic:
		return "Basic"
	case authAPIKey:
		if name := m.apiKeyNameInput.Value(); name != "" {
			return name
		}
		return "API Key"
	default:
		return "None"
	}
}

func bodySummaryStr(m Model) string {
	if m.bodyMode == bodyModeRaw {
		if m.bodyInput.Value() == "" {
			return "(empty)"
		}
		return "raw JSON"
	}
	n := 0
	for _, p := range m.kvPairs {
		if p.key != "" {
			n++
		}
	}
	if n == 0 {
		return "(empty)"
	}
	return fmt.Sprintf("%d field(s)", n)
}

// ── Shared KV-Builder renderer ────────────────────────────────────────────────

func viewGenericKV(
	pairs []kvPair,
	cursor int,
	editing bool,
	keyInput, valInput interface{ View() string },
	keyLabel, valLabel string,
) string {
	if len(pairs) == 0 {
		return subtitleStyle.Render("      (empty — press  n  to add)")
	}

	indent := "      "
	lines := []string{
		indent + lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Width(26).Render(keyLabel),
			labelStyle.Render(valLabel),
		),
		indent + subtitleStyle.Render(strings.Repeat("─", 50)),
	}

	for i, pair := range pairs {
		selected := i == cursor
		if selected && editing {
			lines = append(lines, "  "+lipgloss.JoinHorizontal(lipgloss.Center,
				"    ", keyInput.View(), "  ", valInput.View(),
			))
			continue
		}
		cur := "      "
		k := normalItemStyle.Width(24).Render(pair.key)
		v := normalItemStyle.Render(pair.value)
		if selected {
			cur = "    " + cursorStyle.Render("> ")
			k = selectedItemStyle.Width(24).Render(pair.key)
			v = selectedItemStyle.Render(pair.value)
		}
		lines = append(lines, cur+lipgloss.JoinHorizontal(lipgloss.Left, k, v))
	}
	return strings.Join(lines, "\n")
}

// ── Auth section ──────────────────────────────────────────────────────────────

func viewAuthSection(m Model) string {
	authTypes := []string{"None", "Bearer", "Basic", "API Key"}
	badges := make([]string, len(authTypes))
	for i, t := range authTypes {
		if authKind(i) == m.authKind {
			badges[i] = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1).
				Render(t)
		} else {
			badges[i] = lipgloss.NewStyle().
				Foreground(colorMuted).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(0, 1).
				Render(t)
		}
	}

	// PaddingLeft statt String-Konkatenation: lipgloss rückt JEDE Zeile
	// des mehrzeiligen Badge-Blocks ein, nicht nur die erste.
	selector := lipgloss.NewStyle().PaddingLeft(6).Render(
		lipgloss.JoinHorizontal(lipgloss.Left, badges...),
	)

	pad := "      "
	rows := []string{
		pad + subtitleStyle.Render("1–4  select type"),
		"",
		selector,
		"",
	}

	switch m.authKind {
	case authBearer:
		rows = append(rows, pad+labelStyle.Render("Token"), pad+m.tokenInput.View())
	case authBasic:
		rows = append(rows,
			pad+labelStyle.Render("Username"), pad+m.authUserInput.View(),
			"",
			pad+labelStyle.Render("Password"), pad+m.authPassInput.View(),
		)
	case authAPIKey:
		rows = append(rows,
			pad+labelStyle.Render("Header Name"), pad+m.apiKeyNameInput.View(),
			"",
			pad+labelStyle.Render("Value"), pad+m.apiKeyValInput.View(),
		)
	case authNone:
		rows = append(rows, pad+subtitleStyle.Render("No authentication"))
	}

	return strings.Join(rows, "\n")
}

// ── Request key bar ───────────────────────────────────────────────────────────

func viewRequestKeyBar(m Model, method string, maxWidth int) string {
	switch m.activeField {
	case fieldParams:
		if m.qpEditing {
			return wrapKeyBar(maxWidth,
				keyBind("Tab", "Key ↔ Value"), keyBind("Enter", "confirm"), keyBind("Esc", "cancel"),
			)
		}
		return wrapKeyBar(maxWidth,
			keyBind("n", "add"), keyBind("d", "delete"), keyBind("Enter", "edit"),
			keyBind("ctrl+s", "send"), keyBind("Tab", "→ headers"), keyBind("Esc", "→ URL"),
		)

	case fieldHeaders:
		if m.hdrEditing {
			return wrapKeyBar(maxWidth,
				keyBind("Tab", "Key ↔ Value"), keyBind("Enter", "confirm"), keyBind("Esc", "cancel"),
			)
		}
		return wrapKeyBar(maxWidth,
			keyBind("n", "add"), keyBind("d", "delete"), keyBind("Enter", "edit"),
			keyBind("ctrl+s", "send"), keyBind("Tab", "→ auth"), keyBind("Esc", "→ params"),
		)

	case fieldAuth:
		base := []string{
			keyBind("1–4", "auth type"),
			keyBind("ctrl+s", "send"),
			keyBind("Esc", "→ headers"),
		}
		if hasBody(method) {
			base = append(base, keyBind("Tab", "→ body"))
		}
		return wrapKeyBar(maxWidth, base...)

	case fieldBody:
		if m.kvEditing {
			return wrapKeyBar(maxWidth,
				keyBind("Tab", "Key ↔ Value"), keyBind("Enter", "confirm"), keyBind("Esc", "cancel"),
			)
		}
		if m.bodyMode == bodyModeKV {
			return wrapKeyBar(maxWidth,
				keyBind("n", "add"), keyBind("d", "delete"), keyBind("r", "raw JSON"),
				keyBind("ctrl+s", "send"), keyBind("Esc/Tab", "→ auth"),
			)
		}
		return wrapKeyBar(maxWidth, keyBind("ctrl+s", "send"), keyBind("Esc", "→ auth"))
	}

	// URL field
	bindings := []string{
		keyBind("↑/↓", "history"),
		keyBind("F1–F5", "method"),
		keyBind("Tab", "→ params"),
		keyBind("ctrl+y", "copy curl"),
	}
	if hasBody(method) {
		bindings = append(bindings, keyBind("Enter", "→ body"))
	} else {
		bindings = append(bindings, keyBind("Enter", "send"))
	}
	bindings = append(bindings, keyBind("ctrl+s", "send"), keyBind("Esc", "back"))
	return wrapKeyBar(maxWidth, bindings...)
}

// ── Response ──────────────────────────────────────────────────────────────────

func viewResponse(m Model) string {
	available := m.width - horizontalOverhead

	if m.err != nil {
		return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			errorStyle.Render("✗  Request failed"),
			"",
			subtitleStyle.Render(m.err.Error()),
			"",
			divider(available),
			wrapKeyBar(available, keyBind("Esc", "back")),
		))
	}

	url := m.requestURL
	if max := available - 35; len(url) > max && max > 3 {
		url = url[:max-3] + "..."
	}

	// Status + Meta-Zeile
	statusPill := statusBadge(m.statusCode, m.statusText)
	metaStr := ""
	if m.responseDuration > 0 {
		metaStr = subtitleStyle.Render(formatDuration(m.responseDuration) + "  ·  " + formatSize(m.responseSize))
	}

	curlNote := ""
	if m.curlCopied {
		curlNote = "   " + successStyle.Render("✓ curl copied")
	}

	statusRow := lipgloss.JoinHorizontal(lipgloss.Center,
		statusPill,
		"   ",
		metaStr,
	)
	urlRow := lipgloss.JoinHorizontal(lipgloss.Center,
		methodLabel(m.requestMethod)+"  ",
		normalItemStyle.Render(url),
		curlNote,
	)

	rows := []string{statusRow, urlRow, ""}

	if m.showHeaders {
		rows = append(rows,
			titleStyle.Render("Response Headers"),
			"",
			renderResponseHeaders(m.responseHeaders),
			"",
		)
	}

	moreIndicator := ""
	if !m.viewport.AtBottom() {
		moreIndicator = subtitleStyle.Render("  ▼  scroll for more")
	}

	rows = append(rows,
		labelStyle.Render("Body"),
		"",
		m.viewport.View(),
		moreIndicator,
		"",
		divider(available),
		wrapKeyBar(available,
			keyBind("j/k", "scroll"),
			keyBind("PgUp/PgDn", "page"),
			keyBind("h", "headers"),
			keyBind("c", "copy curl"),
			keyBind("e", "edit"),
			keyBind("Esc", "back"),
		),
	)

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func renderResponseHeaders(headers http.Header) string {
	if len(headers) == 0 {
		return subtitleStyle.Render("  (no headers)")
	}
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		for _, v := range headers[k] {
			lines = append(lines, fmt.Sprintf("  %s  %s",
				labelStyle.Render(k),
				normalItemStyle.Render(v),
			))
		}
	}
	return strings.Join(lines, "\n")
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func formatSize(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}

// ── Webhook Setup ─────────────────────────────────────────────────────────────

func viewWebhookSetup(m Model) string {
	available := m.width - horizontalOverhead
	errLine := ""
	if m.webhookErr != nil {
		errLine = errorStyle.Render("  " + m.webhookErr.Error())
	}
	rows := []string{
		appHeader(),
		"",
		titleStyle.Render("Webhook Catcher"),
		"",
		subtitleStyle.Render("Start a local HTTP server and capture incoming requests."),
		subtitleStyle.Render("Works with Stripe, GitHub, or any webhook integration."),
		"",
		labelStyle.Render("Port"),
		m.webhookPortInput.View(),
	}
	if errLine != "" {
		rows = append(rows, "", errLine)
	}
	rows = append(rows, "", divider(available),
		wrapKeyBar(available, keyBind("Enter", "start"), keyBind("Esc", "back")),
	)
	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func viewWebhookLive(m Model) string {
	if m.webhookDetail && len(m.webhookRequests) > 0 {
		return viewWebhookDetail(m)
	}
	return viewWebhookList(m)
}

func viewWebhookList(m Model) string {
	available := m.width - horizontalOverhead
	var content string
	if len(m.webhookRequests) == 0 {
		content = subtitleStyle.Render("  Waiting for requests on :" + strconv.Itoa(m.webhookPort) + " ...")
	} else {
		lines := []string{
			lipgloss.JoinHorizontal(lipgloss.Left,
				labelStyle.Width(9).Render("METHOD"),
				labelStyle.Width(42).Render("PATH"),
				labelStyle.Render("TIME"),
			),
			subtitleStyle.Render(strings.Repeat("─", 62)),
		}
		for i, req := range m.webhookRequests {
			selected := i == m.webhookCursor
			path := req.Path
			if len(path) > 40 {
				path = path[:37] + "..."
			}
			t := req.Time.Format("15:04:05")
			meth := methodBadge(req.Method, false)
			if selected {
				meth = methodBadge(req.Method, true)
				lines = append(lines,
					cursorStyle.Render("> ")+lipgloss.JoinHorizontal(lipgloss.Left,
						lipgloss.NewStyle().Width(9).Render(meth),
						selectedItemStyle.Width(42).Render(path),
						selectedItemStyle.Render(t),
					),
				)
			} else {
				lines = append(lines,
					"  "+lipgloss.JoinHorizontal(lipgloss.Left,
						lipgloss.NewStyle().Width(9).Render(meth),
						normalItemStyle.Width(42).Render(path),
						subtitleStyle.Render(t),
					),
				)
			}
		}
		content = strings.Join(lines, "\n")
	}

	reqCount := ""
	if n := len(m.webhookRequests); n > 0 {
		reqCount = subtitleStyle.Render(fmt.Sprintf("  %d captured", n))
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			titleStyle.Render("Webhook Catcher"),
			"  ",
			subtitleStyle.Render(fmt.Sprintf(":%d", m.webhookPort)),
			reqCount,
		),
		"",
		content,
		"",
		divider(available),
		wrapKeyBar(available,
			keyBind("j/k", "navigate"),
			keyBind("Enter", "detail"),
			keyBind("r", "replay"),
			keyBind("s", "stop"),
		),
	))
}

func viewWebhookDetail(m Model) string {
	available := m.width - horizontalOverhead
	req := m.webhookRequests[m.webhookCursor]
	moreIndicator := ""
	if !m.webhookVP.AtBottom() {
		moreIndicator = subtitleStyle.Render("  ▼  scroll for more")
	}
	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			methodBadge(req.Method, true),
			"  ",
			normalItemStyle.Render(req.Path),
			"  ",
			subtitleStyle.Render(req.Time.Format("15:04:05")),
		),
		"",
		m.webhookVP.View(),
		moreIndicator,
		"",
		divider(available),
		wrapKeyBar(available,
			keyBind("j/k", "scroll"),
			keyBind("r", "replay as request"),
			keyBind("Esc", "back"),
		),
	))
}

func renderWebhookDetail(req webhook.Request) string {
	var sb strings.Builder
	sb.WriteString(labelStyle.Render("Headers") + "\n\n")
	for k, vals := range req.Headers {
		for _, v := range vals {
			fmt.Fprintf(&sb, "  %s  %s\n", labelStyle.Render(k), normalItemStyle.Render(v))
		}
	}
	if req.Body != "" {
		sb.WriteString("\n")
		sb.WriteString(labelStyle.Render("Body") + "\n\n")
		ct := req.Headers.Get("Content-Type")
		sb.WriteString(highlight(req.Body, ct))
	}
	return sb.String()
}

// ── Environments ──────────────────────────────────────────────────────────────

func viewEnvList(m Model) string {
	available := m.width - horizontalOverhead

	var content string
	if len(m.envs.Envs) == 0 {
		content = subtitleStyle.Render("  No environments yet.")
	} else {
		lines := []string{
			lipgloss.JoinHorizontal(lipgloss.Left,
				labelStyle.Width(28).Render("NAME"),
				labelStyle.Render("VARIABLES"),
			),
			subtitleStyle.Render(strings.Repeat("─", 40)),
		}
		for i, env := range m.envs.Envs {
			selected := i == m.envCursor
			isActive := i == m.envs.Active
			name := env.Name
			if isActive {
				name += "  " + successStyle.Render("✓ active")
			}
			vars := fmt.Sprintf("%d", len(env.Vars))

			if selected {
				lines = append(lines,
					cursorStyle.Render("> ")+lipgloss.JoinHorizontal(lipgloss.Left,
						selectedItemStyle.Width(26).Render(name),
						selectedItemStyle.Render(vars),
					),
				)
			} else {
				lines = append(lines,
					"  "+lipgloss.JoinHorizontal(lipgloss.Left,
						normalItemStyle.Width(26).Render(name),
						subtitleStyle.Render(vars),
					),
				)
			}
		}
		content = strings.Join(lines, "\n")
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		appHeader(),
		"",
		titleStyle.Render("Environments"),
		"",
		content,
		"",
		divider(available),
		wrapKeyBar(available,
			keyBind("Enter", "activate"),
			keyBind("e", "edit"),
			keyBind("n", "new"),
			keyBind("d", "delete"),
			keyBind("Esc", "back"),
		),
	))
}

func viewEnvEdit(m Model) string {
	available := m.width - horizontalOverhead

	nameLabel := labelStyle.Render("Name")
	if m.envNameFocused {
		nameLabel = expandedSectionHeader("Name")
	}

	rows := []string{
		appHeader(),
		"",
		titleStyle.Render("Edit Environment"),
		"",
		nameLabel,
		"  " + m.envNameInput.View(),
		"",
	}

	if !m.envNameFocused {
		rows = append(rows,
			expandedSectionHeader("Variables"),
			"",
			viewGenericKV(m.envVarKV, m.envVarCursor, m.envVarEditing, m.envVarKeyIn, m.envVarValIn, "VARIABLE", "VALUE"),
			"",
			divider(available),
			wrapKeyBar(available,
				keyBind("n", "add"), keyBind("d", "delete"),
				keyBind("e", "edit name"), keyBind("Esc", "save & back"),
			),
		)
	} else {
		rows = append(rows,
			subtitleStyle.Render("  Press Tab or Enter to edit variables"),
			"",
			divider(available),
			wrapKeyBar(available,
				keyBind("Tab/Enter", "→ variables"),
				keyBind("Esc", "save & back"),
			),
		)
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}
