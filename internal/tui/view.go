package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Laden..." // Terminal-Größe noch nicht bekannt
	}

	var content string
	switch m.activeView {
	case requestView:
		content = viewRequest(m)
	case responseView:
		content = viewResponse(m)
	default:
		content = viewMenu(m)
	}

	// appStyle gibt der gesamten App etwas Luft zum Rand
	return appStyle.Render(content)
}

// header rendert die Titelzeile die in allen Views gleich ist.
func header() string {
	title := titleStyle.Render("apish")
	subtitle := subtitleStyle.Render("Terminal API Client")
	return lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", subtitle)
}

// footer rendert die Hilfszeile mit Keybindings.
func footer(help string) string {
	return helpStyle.Render(help)
}

func viewMenu(m Model) string {
	// Menüpunkte rendern
	var items strings.Builder
	for i, choice := range m.choices {
		var line string
		if m.cursor == i {
			cursor := cursorStyle.Render(">")
			text := selectedItemStyle.Render(choice)
			line = fmt.Sprintf("%s %s", cursor, text)
		} else {
			line = fmt.Sprintf("  %s", normalItemStyle.Render(choice))
		}
		items.WriteString(line + "\n")
	}

	// Panel zusammenbauen: Header + Leerzeile + Items
	body := lipgloss.JoinVertical(lipgloss.Left,
		header(),
		"",
		strings.TrimRight(items.String(), "\n"),
		"",
		footer("j/k  navigieren   Enter  auswählen   q  beenden"),
	)

	return panelStyle.Render(body)
}

func viewRequest(m Model) string {
	var content string

	if m.loading {
		content = lipgloss.JoinVertical(lipgloss.Left,
			labelStyle.Render("URL"),
			m.urlInput.View(),
			"",
			loadingStyle.Render("Sende Request..."),
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			labelStyle.Render("URL"),
			m.urlInput.View(),
			"",
			footer("Enter  senden   Esc  zurück"),
		)
	}

	panel := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("GET Request"),
		"",
		content,
	)

	return panelStyle.Render(panel)
}

func viewResponse(m Model) string {
	if m.err != nil {
		body := lipgloss.JoinVertical(lipgloss.Left,
			errorStyle.Render("Fehler"),
			"",
			normalItemStyle.Render(m.err.Error()),
			"",
			footer("Esc  zurück"),
		)
		return panelStyle.Render(body)
	}

	// Status-Badge: Farbe je nach Code
	statusBadge := statusStyle(m.statusCode).Render(fmt.Sprintf("  %s  ", m.statusText))

	// Body kürzen wenn er zu lang ist
	// m.width - 8: Platz für Panel-Padding und Border abziehen
	maxBodyLen := (m.height - 10) * (m.width - 8)
	if maxBodyLen < 100 {
		maxBodyLen = 100
	}

	bodyText := m.response
	truncated := false
	if len(bodyText) > maxBodyLen {
		bodyText = bodyText[:maxBodyLen]
		truncated = true
	}

	lines := []string{
		statusBadge,
		"",
		labelStyle.Render("Body"),
		bodyStyle.Render(bodyText),
	}

	if truncated {
		lines = append(lines, truncatedStyle.Render(fmt.Sprintf("[%d Zeichen gekürzt]", len(m.response)-maxBodyLen)))
	}

	lines = append(lines, "", footer("Esc  zurück zum Menü"))

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
