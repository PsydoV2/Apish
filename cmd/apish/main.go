package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// 1. Das Model: Speichert den Zustand unserer App
type model struct {
	choices  []string         // Die Menüpunkte
	cursor   int              // Welcher Punkt ist gerade ausgewählt?
}

// 2. Init: Was passiert beim Start? (Erstmal nichts)
func (m model) Init() tea.Cmd {
	return nil
}

// 3. Update: Reagiert auf Tastendrücke
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q": // Beendet das Programm
			return m, tea.Quit
		case "up", "k": // Cursor hoch (inklusive Vim-Binding 'k')
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j": // Cursor runter (inklusive Vim-Binding 'j')
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// 4. View: Zeichnet das UI ins Terminal
func (m model) View() string {
	s := "Willkommen bei apish - Deinem neuen API Tool!\n\n"

	for i, choice := range m.choices {
		cursor := " " // Normaler Abstand
		if m.cursor == i {
			cursor = ">" // Das ist der ausgewählte Punkt
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\n(Drücke 'q' zum Beenden)\n"
	return s
}

// 5. Main: Startet die App
func main() {
	// Start-Zustand definieren
	initialModel := model{
		choices: []string{"GET Request", "POST Request", "Lokalen Webhook starten (Twist 3)", "Einstellungen"},
	}

	// Programm initialisieren und ausführen
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Oh nein, ein Fehler: %v", err)
		os.Exit(1)
	}
}