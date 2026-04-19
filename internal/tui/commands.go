package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/config"
	"github.com/PsydoV2/Apish/internal/httpclient"
)

// responseMsg trägt das Ergebnis eines HTTP-Calls zurück an Update().
type responseMsg struct {
	response httpclient.Response
	err      error
}

// historySavedMsg signalisiert dass die History gespeichert wurde.
// err ist nil bei Erfolg — wir loggen Fehler im Update still weg.
type historySavedMsg struct{ err error }

// sendRequest führt den HTTP-Call asynchron aus.
func sendRequest(method, url, body string) tea.Cmd {
	return func() tea.Msg {
		resp, err := httpclient.Do(method, url, body)
		return responseMsg{response: resp, err: err}
	}
}

// saveHistoryCmd schreibt die History asynchron in die Datei.
// Wir blockieren nie den UI-Thread für Disk-I/O.
func saveHistoryCmd(h config.History) tea.Cmd {
	return func() tea.Msg {
		return historySavedMsg{err: config.SaveHistory(h)}
	}
}
