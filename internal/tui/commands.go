package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PsydoV2/Apish/internal/config"
	"github.com/PsydoV2/Apish/internal/httpclient"
	"github.com/PsydoV2/Apish/internal/webhook"
)

// ── HTTP ──────────────────────────────────────────────────────────────────────

type responseMsg struct {
	response httpclient.Response
	err      error
}

type historySavedMsg struct{ err error }

func sendRequest(method, url, body string) tea.Cmd {
	return func() tea.Msg {
		resp, err := httpclient.Do(method, url, body)
		return responseMsg{response: resp, err: err}
	}
}

func saveHistoryCmd(h config.History) tea.Cmd {
	return func() tea.Msg {
		return historySavedMsg{err: config.SaveHistory(h)}
	}
}

// ── Webhook ───────────────────────────────────────────────────────────────────

// webhookRequestMsg is sent to Update each time a webhook request arrives.
type webhookRequestMsg webhook.Request

// webhookServerDoneMsg is sent when the server stops (cleanly or with an error).
type webhookServerDoneMsg struct{ err error }

// startWebhookServerCmd runs the HTTP server in a Bubble Tea goroutine.
// It returns webhookServerDoneMsg when the server exits.
func startWebhookServerCmd(port int, ch chan webhook.Request, stop chan struct{}) tea.Cmd {
	return func() tea.Msg {
		err := webhook.Start(port, ch, stop)
		return webhookServerDoneMsg{err: err}
	}
}

// waitForWebhookCmd blocks until either a request arrives on ch or stop is closed.
// After handling a webhookRequestMsg, Update must re-issue this cmd to keep listening.
//
// This is the standard Bubble Tea pattern for long-running event streams:
// one cmd fires, delivers one message, then you issue it again — a self-renewing loop.
func waitForWebhookCmd(ch chan webhook.Request, stop chan struct{}) tea.Cmd {
	return func() tea.Msg {
		select {
		case req := <-ch:
			return webhookRequestMsg(req)
		case <-stop:
			// stop was closed — signal to Update that we're done
			return webhookServerDoneMsg{}
		}
	}
}
