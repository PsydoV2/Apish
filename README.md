# 🐝 apish

> **The Terminal-First API Client with a Brain.** `apish` is a lightweight, blazing-fast TUI built in Go — for developers who find Postman too heavy but `curl` too manual. It lives in your terminal, respects your RAM, and comes with superpowers you won't find anywhere else.

---

### ✨ Why apish?

Postman is too heavy. `curl` is too manual. `apish` is the sweet spot.

- **Vim keybindings** — navigate requests like a pro
- **Zero bloat** — single binary, no Electron, no nonsense
- **Local-first** — your data stays on your machine
- **Persistent history** — every request saved and restored, even across restarts

---

### 🚀 Installation

```bash
go install github.com/PsydoV2/Apish/cmd/apish@latest
```

Or build from source:

```bash
git clone https://github.com/PsydoV2/Apish
cd Apish
go build ./cmd/apish
```

---

### ⌨️ Keybindings

**Menu**

| Key | Action |
|---|---|
| `j` / `k` | Navigate |
| `Enter` | Select |
| `q` | Quit |

**Request**

| Key | Action |
|---|---|
| `F1` – `F4` | Switch method (GET / POST / PUT / DELETE) |
| `↑` / `↓` | Browse history |
| `Tab` | Focus URL → Body |
| `Enter` | Send (GET/DELETE) or go to body (POST/PUT) |
| `ctrl+s` | Send from anywhere |
| `Esc` | Back |

**Body — KV Builder**

| Key | Action |
|---|---|
| `n` | Add field |
| `d` | Delete selected |
| `Enter` | Edit selected |
| `Tab` | Switch Key ↔ Value |
| `r` | Toggle raw JSON mode |
| `Tab` / `Esc` | Back to URL |

**Response**

| Key | Action |
|---|---|
| `j` / `k` | Scroll |
| `PgUp` / `PgDn` | Page up / down |
| `e` | Edit URL |
| `Esc` | Back to menu |

---

### 🗺 Roadmap

#### ✅ Phase 1 — The Core

- GET, POST, PUT, DELETE requests
- Key-Value body builder with automatic JSON generation
- Raw JSON editor as fallback
- Syntax highlighting for JSON, XML, HTML, YAML — adapts to your terminal theme
- Scrollable response view with status codes color-coded
- Persistent request history — URL, method, and body all restored on `↑`

#### 🔥 Phase 2 — Webhook Catcher *(coming next)*

Don't just send requests — receive them. Start a local listener with one keypress and inspect incoming webhooks from Stripe, GitHub, or your own microservices in a beautiful interface.

#### 🤖 Phase 3 — Local AI Debugger

Integrated with **Ollama**. When a request fails, `apish` analyses the headers and body locally on your machine. It tells you exactly *why* it failed — without ever sending data to the cloud.

#### 📡 Phase 4 — Instant Sharing

Press one key and `apish` uploads the request/response pair to your private VPS and copies a shareable link to your clipboard. No more copy-pasting JSON blocks into Slack.

---

### 💻 Tech Stack

| | |
|---|---|
| Language | Go ≥ 1.22 |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Syntax Highlighting | [Chroma](https://github.com/alecthomas/chroma) |
| AI *(planned)* | Ollama |

---

### 📁 History File

Every request you send is saved locally:

| OS | Path |
|---|---|
| Linux / macOS | `~/.config/apish/history.json` |
| Windows | `%AppData%\apish\history.json` |

Up to 100 entries. Browse with `↑` / `↓` in the URL field — restores the full request including method and body.

---

*Built with ❤️ by [PsydoV2](https://github.com/PsydoV2)*
