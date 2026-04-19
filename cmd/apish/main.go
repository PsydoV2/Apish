package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/config"
	"github.com/PsydoV2/Apish/internal/tui"
)

func main() {
	// History beim Start laden — Fehler sind nicht fatal,
	// wir starten einfach mit leerer History.
	history, err := config.LoadHistory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warnung: history konnte nicht geladen werden: %v\n", err)
	}

	p := tea.NewProgram(tui.New(history))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Starten: %v\n", err)
		os.Exit(1)
	}
}
