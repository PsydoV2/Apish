package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/tui"
)

func main() {
	// tui.New() gibt unser initiales Model zurück.
	// tea.NewProgram() startet die Bubble Tea Event-Loop:
	//   1. Init() aufrufen
	//   2. auf Events warten → Update() → View() → Terminal neu zeichnen
	//   3. wiederholen bis tea.Quit
	p := tea.NewProgram(tui.New())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Starten: %v\n", err)
		os.Exit(1)
	}
}
