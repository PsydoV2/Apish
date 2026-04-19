package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/httpclient"
)

// responseMsg ist unsere eigene tea.Msg — ein einfaches Struct.
// Bubble Tea schickt es an Update() wenn der HTTP-Call fertig ist.
type responseMsg struct {
	response httpclient.Response
	err      error
}

// sendGetRequest gibt eine tea.Cmd zurück.
//
// tea.Cmd ist definiert als: type Cmd func() Msg
// Das heißt: wir geben eine Funktion zurück, die Bubble Tea
// in einer separaten Goroutine ausführt — non-blocking!
//
// Die innere Funktion "schließt über" url — das ist eine Closure:
// sie merkt sich den Wert von url auch wenn sendGetRequest längst zurückgekehrt ist.
func sendGetRequest(url string) tea.Cmd {
	return func() tea.Msg {
		resp, err := httpclient.Get(url)
		return responseMsg{response: resp, err: err}
	}
}
