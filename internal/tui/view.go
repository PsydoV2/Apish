package tui

import (
	"fmt"
	"strconv"
	"strings"

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
	default:
		return appStyle.Render(viewMenu(m))
	}
}

func header() string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		titleStyle.Render("apish"),
		"  ",
		subtitleStyle.Render("Terminal API Client"),
	)
}

func keyBind(key, desc string) string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		keyStyle.Render(key),
		" ",
		keyDescStyle.Render(desc),
		"   ",
	)
}

func keyBar(bindings ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Center, bindings...)
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
		if i == m.methodIdx {
			badges[i] = methodActiveStyle.Render(method)
		} else {
			badges[i] = methodInactiveStyle.Render(method)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, badges...)
}

// ── Menu ──────────────────────────────────────────────────────────────────────

func viewMenu(m Model) string {
	var items strings.Builder
	for i, choice := range m.choices {
		if m.cursor == i {
			items.WriteString(fmt.Sprintf("%s %s\n",
				cursorStyle.Render(">"),
				selectedItemStyle.Render(choice),
			))
		} else {
			items.WriteString(fmt.Sprintf("  %s\n", normalItemStyle.Render(choice)))
		}
	}

	available := m.width - horizontalOverhead
	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		header(),
		"",
		strings.TrimRight(items.String(), "\n"),
		"",
		wrapKeyBar(available,
			keyBind("j/k", "navigate"),
			keyBind("Enter", "select"),
			keyBind("q", "quit"),
		),
	))
}

// ── Request ───────────────────────────────────────────────────────────────────

func viewRequest(m Model) string {
	if m.loading {
		return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("API Request"),
			"",
			loadingStyle.Render("Sending request..."),
		))
	}

	method := methods[m.methodIdx]

	urlLabel := labelStyle.Render("URL")
	if m.activeField == fieldURL {
		urlLabel = titleStyle.Render("URL")
	}

	rows := []string{
		titleStyle.Render("API Request"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Center,
			labelStyle.Render("Method   "),
			methodSelector(m),
			subtitleStyle.Render("   F1–F4"),
		),
		"",
		urlLabel,
		m.urlInput.View(),
	}

	if hasBody(method) {
		bodyLabel := labelStyle.Render("Body")
		if m.activeField == fieldBody {
			bodyLabel = titleStyle.Render("Body")
		}
		rows = append(rows, "", bodyLabel, "")

		if m.bodyMode == bodyModeRaw {
			rows = append(rows,
				m.bodyInput.View(),
				"",
				subtitleStyle.Render("  Esc  back to builder"),
			)
		} else {
			rows = append(rows, viewKVBuilder(m))
		}
	}

	available := m.width - horizontalOverhead
	rows = append(rows, "", viewRequestKeyBar(m, method, available))
	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func viewRequestKeyBar(m Model, method string, maxWidth int) string {
	if m.activeField == fieldBody && hasBody(method) {
		if m.kvEditing {
			return wrapKeyBar(maxWidth,
				keyBind("Tab", "Key ↔ Value"),
				keyBind("Enter", "confirm"),
				keyBind("Esc", "cancel"),
			)
		}
		if m.bodyMode == bodyModeKV {
			return wrapKeyBar(maxWidth,
				keyBind("n", "add field"),
				keyBind("d", "delete"),
				keyBind("Enter", "edit"),
				keyBind("r", "raw JSON"),
				keyBind("ctrl+s", "send"),
				keyBind("Tab/Esc", "→ URL"),
			)
		}
	}
	bindings := []string{
		keyBind("↑/↓", "history"),
		keyBind("F1–F4", "method"),
	}
	if hasBody(method) {
		bindings = append(bindings, keyBind("Tab", "→ body"))
	}
	bindings = append(bindings,
		keyBind("Enter", "send"),
		keyBind("Esc", "back"),
	)
	return wrapKeyBar(maxWidth, bindings...)
}

func viewKVBuilder(m Model) string {
	if len(m.kvPairs) == 0 {
		return subtitleStyle.Render("  (empty — press  n  to add a field)")
	}

	lines := []string{
		lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Width(26).Render("KEY"),
			labelStyle.Render("VALUE"),
		),
		subtitleStyle.Render(strings.Repeat("─", 50)),
	}

	for i, pair := range m.kvPairs {
		selected := i == m.kvCursor

		if selected && m.kvEditing {
			lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Center,
				"  ",
				m.kvKeyInput.View(),
				"  ",
				m.kvValInput.View(),
			))
			continue
		}

		cursor := "  "
		key := normalItemStyle.Width(24).Render(pair.key)
		val := normalItemStyle.Render(pair.value)
		if selected {
			cursor = cursorStyle.Render("> ")
			key = selectedItemStyle.Width(24).Render(pair.key)
			val = selectedItemStyle.Render(pair.value)
		}
		lines = append(lines, cursor+lipgloss.JoinHorizontal(lipgloss.Left, key, val))
	}

	return strings.Join(lines, "\n")
}

// ── Response ──────────────────────────────────────────────────────────────────

func viewResponse(m Model) string {
	if m.err != nil {
		return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
			errorStyle.Render("Error"),
			"",
			normalItemStyle.Render(m.err.Error()),
			"",
			wrapKeyBar(m.width-horizontalOverhead, keyBind("Esc", "back")),
		))
	}

	url := m.requestURL
	if max := m.width - horizontalOverhead - 8; len(url) > max && max > 3 {
		url = url[:max-3] + "..."
	}

	moreIndicator := ""
	if !m.viewport.AtBottom() {
		moreIndicator = subtitleStyle.Render("  ▼  scroll for more")
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			statusStyle(m.statusCode).Render(fmt.Sprintf("  %s  ", m.statusText)),
			"  ",
			labelStyle.Render(m.requestMethod+" "),
			normalItemStyle.Render(url),
		),
		"",
		labelStyle.Render("Body"),
		"",
		m.viewport.View(),
		moreIndicator,
		"",
		wrapKeyBar(m.width-horizontalOverhead,
			keyBind("j/k", "scroll"),
			keyBind("PgUp/PgDn", "page"),
			keyBind("e", "edit URL"),
			keyBind("Esc", "back"),
		),
	))
}

// ── Webhook Setup ─────────────────────────────────────────────────────────────

func viewWebhookSetup(m Model) string {
	errLine := ""
	if m.webhookErr != nil {
		errLine = errorStyle.Render("  " + m.webhookErr.Error())
	}

	rows := []string{
		header(),
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
	rows = append(rows, "",
		wrapKeyBar(m.width-horizontalOverhead,
			keyBind("Enter", "start"),
			keyBind("Esc", "back"),
		),
	)
	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// ── Webhook Live ──────────────────────────────────────────────────────────────

func viewWebhookLive(m Model) string {
	if m.webhookDetail && len(m.webhookRequests) > 0 {
		return viewWebhookDetail(m)
	}
	return viewWebhookList(m)
}

func viewWebhookList(m Model) string {
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
			if selected {
				lines = append(lines,
					cursorStyle.Render("> ")+lipgloss.JoinHorizontal(lipgloss.Left,
						selectedItemStyle.Width(9).Render(req.Method),
						selectedItemStyle.Width(42).Render(path),
						selectedItemStyle.Render(t),
					),
				)
			} else {
				lines = append(lines,
					"  "+lipgloss.JoinHorizontal(lipgloss.Left,
						normalItemStyle.Width(9).Render(req.Method),
						normalItemStyle.Width(42).Render(path),
						subtitleStyle.Render(t),
					),
				)
			}
		}
		content = strings.Join(lines, "\n")
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			titleStyle.Render("Webhook Catcher"),
			"  ",
			subtitleStyle.Render(fmt.Sprintf("listening on :%d", m.webhookPort)),
		),
		"",
		content,
		"",
		wrapKeyBar(m.width-horizontalOverhead,
			keyBind("j/k", "navigate"),
			keyBind("Enter", "detail"),
			keyBind("s", "stop"),
		),
	))
}

func viewWebhookDetail(m Model) string {
	req := m.webhookRequests[m.webhookCursor]

	moreIndicator := ""
	if !m.webhookVP.AtBottom() {
		moreIndicator = subtitleStyle.Render("  ▼  scroll for more")
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			methodActiveStyle.Render(req.Method),
			"  ",
			normalItemStyle.Render(req.Path),
			"  ",
			subtitleStyle.Render(req.Time.Format("15:04:05")),
		),
		"",
		m.webhookVP.View(),
		moreIndicator,
		"",
		wrapKeyBar(m.width-horizontalOverhead,
			keyBind("j/k", "scroll"),
			keyBind("PgUp/PgDn", "page"),
			keyBind("Esc", "back"),
		),
	))
}

// renderWebhookDetail builds the viewport content for a single captured request.
// Headers first, then a syntax-highlighted body (if present).
func renderWebhookDetail(req webhook.Request) string {
	var sb strings.Builder

	sb.WriteString(labelStyle.Render("Headers") + "\n\n")
	for k, vals := range req.Headers {
		for _, v := range vals {
			sb.WriteString(fmt.Sprintf("  %s: %s\n",
				subtitleStyle.Render(k),
				normalItemStyle.Render(v),
			))
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
