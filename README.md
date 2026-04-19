# 🐝 apish

> **The Terminal-First API Client with a Brain.** `apish` is a lightweight, blazing-fast TUI (Terminal User Interface) built in Go. It’s designed for developers who are tired of heavy Electron apps but want more power than a simple `curl`.

---

### ✨ Why apish?

Postman is too heavy. `curl` is too manual. `apish` is the sweet spot. It lives in your terminal, respects your RAM, and comes with superpowers you won't find anywhere else.

- **Vim-Keybindings:** Navigate your API requests like a pro.
- **Zero-Bloat:** Single binary. No dependencies. 100% Go.
- **Local-First:** Your data stays on your machine.

---

### 🔥 The Superpowers (Upcoming)

We aren't just building another API client. `apish` is built to solve real-world debugging pain:

#### 📡 1. Instant Multiplayer (Shareable Sessions)

Got a weird 500 error? Press `Ctrl+S` and `apish` instantly uploads the request/response to your private VPS instance and copies a shareable link to your clipboard. No more copy-pasting JSON blocks into Slack.

#### 🤖 2. Local AI Debugger

Integrated with **Ollama**. If an API fails, `apish` analyzes the headers and body locally on your machine. It tells you exactly _why_ it failed (e.g., "Hey, you forgot the Bearer token\!") without ever sending data to the cloud.

#### 🎣 3. Reverse Webhook Catcher

Don't just send requests—receive them. Start a local listener with one keypress and inspect incoming webhooks from Stripe, GitHub, or your own microservices in a beautiful interface.

---

### 🛠 Installation (Planned)

```bash
# Using Go
go install github.com/PsydoV2/Apish/cmd/apish@latest
```

---

### 🗺 Roadmap

- [ ] **Phase 1: The Core** (TUI, GET/POST/PUT/DELETE, Syntax Highlighting)
- [ ] **Phase 2: Webhook Catcher** (Listen for incoming requests locally)
- [ ] **Phase 3: AI Insights** (Local LLM integration for error debugging)
- [ ] **Phase 4: Cloud Share** (Private VPS backend for instant link sharing)

---

### 💻 Tech Stack

- **Language:** Go (Golang)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) & [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **AI Integration:** Ollama
- **Backend:** Go + SQLite (hosted on Linux VPS)

---

### 🤝 Contributing

Contributions, issues, and feature requests are welcome\! Feel free to check the [issues page](https://www.google.com/search?q=https://github.com/PsydoV2/Apish/issues).

---

_Built with ❤️ by [PsydoV2](https://github.com/PsydoV2)_
