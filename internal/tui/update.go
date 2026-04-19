package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// tea.WindowSizeMsg: Bubble Tea schickt diese Message beim Start
	// und bei jeder Terminal-Größenänderung — immer global behandeln.
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Globaler Quit
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// responseMsg: global, kann jederzeit ankommen
	if msg, ok := msg.(responseMsg); ok {
		m.loading = false
		m.activeView = responseView
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.response = msg.response.Body
			m.statusCode = msg.response.StatusCode
			m.statusText = msg.response.Status
		}
		return m, nil
	}

	switch m.activeView {
	case menuView:
		return updateMenu(m, msg)
	case requestView:
		return updateRequest(m, msg)
	case responseView:
		return updateResponse(m, msg)
	}

	return m, nil
}

func updateMenu(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == 0 {
				m.activeView = requestView
				m.urlInput.Focus()
				return m, nil
			}
		}
	}
	return m, nil
}

func updateRequest(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.activeView = menuView
			m.urlInput.Blur()
			return m, nil
		case "enter":
			url := m.urlInput.Value()
			if url == "" {
				return m, nil
			}
			m.loading = true
			m.err = nil
			return m, sendGetRequest(url)
		}
	}

	var cmd tea.Cmd
	m.urlInput, cmd = m.urlInput.Update(msg)
	return m, cmd
}

func updateResponse(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc", "q":
			m.activeView = menuView
			m.response = ""
			m.statusCode = 0
			m.statusText = ""
			m.err = nil
			return m, nil
		}
	}
	return m, nil
}
