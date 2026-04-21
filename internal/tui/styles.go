package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary = lipgloss.AdaptiveColor{Light: "#5B21B6", Dark: "#A78BFA"} // violet (PATCH, UI accent)
	colorGreen   = lipgloss.AdaptiveColor{Light: "#065F46", Dark: "#34D399"} // GET, 2xx
	colorYellow  = lipgloss.AdaptiveColor{Light: "#92400E", Dark: "#FCD34D"} // POST, 3xx, loading
	colorBlue    = lipgloss.AdaptiveColor{Light: "#1E40AF", Dark: "#93C5FD"} // PUT
	colorRed     = lipgloss.AdaptiveColor{Light: "#991B1B", Dark: "#F87171"} // DELETE, 4xx/5xx, errors
	colorMuted   = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"} // subtitles, inactive
	colorText    = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"} // primary text
	colorBorder  = lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#374151"} // panel border
	colorTeal    = lipgloss.AdaptiveColor{Light: "#0F766E", Dark: "#2DD4BF"} // environment badge
)

// methodColor returns the accent color for each HTTP method.
func methodColor(method string) lipgloss.TerminalColor {
	switch method {
	case "GET":
		return colorGreen
	case "POST":
		return colorYellow
	case "PUT":
		return colorBlue
	case "PATCH":
		return colorPrimary
	case "DELETE":
		return colorRed
	}
	return colorMuted
}

// ── Layout ────────────────────────────────────────────────────────────────────

var appStyle = lipgloss.NewStyle().Padding(1, 2)

var panelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(1, 3)

// ── Text ──────────────────────────────────────────────────────────────────────

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorPrimary)

var subtitleStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var labelStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Bold(true)

// ── Key bar ───────────────────────────────────────────────────────────────────

var keyStyle = lipgloss.NewStyle().
	Foreground(colorText).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(0, 1)

var keyDescStyle = lipgloss.NewStyle().Foreground(colorMuted)

// ── Menu / lists ──────────────────────────────────────────────────────────────

var selectedItemStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true)

var normalItemStyle = lipgloss.NewStyle().Foreground(colorText)

var cursorStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true)

// ── Request ───────────────────────────────────────────────────────────────────

var loadingStyle = lipgloss.NewStyle().
	Foreground(colorYellow).
	Italic(true)

// methodBadge renders a single HTTP method badge with its own accent color when active.
func methodBadge(method string, active bool) string {
	if !active {
		return lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1).
			Render(method)
	}
	color := methodColor(method)
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1).
		Render(method)
}

// ── Accordion sections ────────────────────────────────────────────────────────

// collapsedSection renders a one-line summary for an inactive section.
func collapsedSection(name, summary string) string {
	arrow := subtitleStyle.Render("  ▸  ")
	label := labelStyle.Width(9).Render(name)
	return arrow + label + subtitleStyle.Render(summary)
}

// expandedSectionHeader renders the header row of the currently active section.
func expandedSectionHeader(name string) string {
	arrow := titleStyle.Render("  ▾  ")
	return arrow + titleStyle.Render(name)
}

// ── Response ──────────────────────────────────────────────────────────────────

// statusBadge renders a background-colored status pill.
func statusBadge(code int, text string) string {
	var bg, fg lipgloss.TerminalColor
	switch {
	case code >= 200 && code < 300:
		bg = lipgloss.AdaptiveColor{Light: "#D1FAE5", Dark: "#064E3B"}
		fg = lipgloss.AdaptiveColor{Light: "#065F46", Dark: "#6EE7B7"}
	case code >= 300 && code < 400:
		bg = lipgloss.AdaptiveColor{Light: "#FEF3C7", Dark: "#451A03"}
		fg = lipgloss.AdaptiveColor{Light: "#92400E", Dark: "#FCD34D"}
	default:
		bg = lipgloss.AdaptiveColor{Light: "#FEE2E2", Dark: "#450A0A"}
		fg = lipgloss.AdaptiveColor{Light: "#991B1B", Dark: "#FCA5A5"}
	}
	return lipgloss.NewStyle().
		Bold(true).
		Background(bg).
		Foreground(fg).
		Padding(0, 1).
		Render(text)
}

// methodLabel renders a colored method name for the response view.
func methodLabel(method string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(methodColor(method)).
		Render(method)
}

var errorStyle = lipgloss.NewStyle().
	Foreground(colorRed).
	Bold(true)

var successStyle = lipgloss.NewStyle().
	Foreground(colorGreen).
	Bold(true)

var envBadgeStyle = lipgloss.NewStyle().
	Foreground(colorTeal).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorTeal).
	Padding(0, 1)
