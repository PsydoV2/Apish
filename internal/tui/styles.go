package tui

import "github.com/charmbracelet/lipgloss"

// AdaptiveColor: Light = heller Hintergrund, Dark = dunkler Hintergrund.
// Lip Gloss erkennt das Terminal-Theme automatisch via COLORFGBG oder OSC-Query.
var (
	colorPrimary = lipgloss.AdaptiveColor{Light: "#5B21B6", Dark: "#A78BFA"} // lila: dunkler/heller je nach BG
	colorGreen   = lipgloss.AdaptiveColor{Light: "#065F46", Dark: "#34D399"}
	colorYellow  = lipgloss.AdaptiveColor{Light: "#92400E", Dark: "#FCD34D"}
	colorRed     = lipgloss.AdaptiveColor{Light: "#991B1B", Dark: "#F87171"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
	colorText    = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"}
	colorBorder  = lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#374151"}
)

// --- Layout ---

var appStyle = lipgloss.NewStyle().Padding(1, 2)

var panelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(1, 2)

// --- Text ---

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(colorPrimary)

var subtitleStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var helpStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// keyStyle: der Key selbst — Badge mit Rounded Border und Padding
var keyStyle = lipgloss.NewStyle().
	Foreground(colorText).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(0, 1)

// keyDescStyle: die Beschreibung zum Key — gedimmt
var keyDescStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// --- Menü ---

var selectedItemStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true)

var normalItemStyle = lipgloss.NewStyle().
	Foreground(colorText)

var cursorStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true)

// --- Request ---

var labelStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Bold(true)

var loadingStyle = lipgloss.NewStyle().
	Foreground(colorYellow).
	Italic(true)

// methodActiveStyle: die aktuell gewählte HTTP-Methode
var methodActiveStyle = lipgloss.NewStyle().
	Foreground(colorText).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorPrimary).
	Padding(0, 1)

// methodInactiveStyle: alle anderen Methoden
var methodInactiveStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Padding(0, 1)

// --- Response ---

func statusStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	case code >= 300 && code < 400:
		return lipgloss.NewStyle().Bold(true).Foreground(colorYellow)
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(colorRed)
	}
}

var bodyStyle = lipgloss.NewStyle().
	Foreground(colorText)

var errorStyle = lipgloss.NewStyle().
	Foreground(colorRed).
	Bold(true)

var truncatedStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Italic(true)
