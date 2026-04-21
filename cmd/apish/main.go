package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/config"
	"github.com/PsydoV2/Apish/internal/tui"
)

func main() {
	history, err := config.LoadHistory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warnung: history konnte nicht geladen werden: %v\n", err)
	}

	envs, err := config.LoadEnvs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warnung: environments konnten nicht geladen werden: %v\n", err)
	}

	p := tea.NewProgram(tui.New(history, envs))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Starten: %v\n", err)
		os.Exit(1)
	}
}
