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
type envsSavedMsg struct{ err error }

func sendRequest(method, url, body string, headers map[string]string) tea.Cmd {
	return func() tea.Msg {
		resp, err := httpclient.Do(method, url, body, headers)
		return responseMsg{response: resp, err: err}
	}
}

func saveHistoryCmd(h config.History) tea.Cmd {
	return func() tea.Msg {
		return historySavedMsg{err: config.SaveHistory(h)}
	}
}

func saveEnvsCmd(e config.Environments) tea.Cmd {
	return func() tea.Msg {
		return envsSavedMsg{err: config.SaveEnvs(e)}
	}
}

// ── Webhook ───────────────────────────────────────────────────────────────────

type webhookRequestMsg webhook.Request
type webhookServerDoneMsg struct{ err error }

func startWebhookServerCmd(port int, ch chan webhook.Request, stop chan struct{}) tea.Cmd {
	return func() tea.Msg {
		err := webhook.Start(port, ch, stop)
		return webhookServerDoneMsg{err: err}
	}
}

func waitForWebhookCmd(ch chan webhook.Request, stop chan struct{}) tea.Cmd {
	return func() tea.Msg {
		select {
		case req := <-ch:
			return webhookRequestMsg(req)
		case <-stop:
			return webhookServerDoneMsg{}
		}
	}
}
