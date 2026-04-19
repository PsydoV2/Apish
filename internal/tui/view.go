package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
