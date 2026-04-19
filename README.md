# apish

> **The Terminal-First API Client.** A lightweight TUI (Terminal User Interface) built in Go — designed for developers who find Postman too heavy but `curl` too manual.

---

### Why apish?

- **Vim keybindings** — navigate requests like a pro
- **Zero bloat** — single binary, no dependencies, 100% Go
- **Local-first** — your data stays on your machine
- **Persistent history** — every request is saved and restored across sessions

---

### Installation

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

### Usage

```
apish
```

#### Keybindings

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

**Body (KV Builder)**

| Key | Action |
|---|---|
| `n` | Add field |
| `d` | Delete selected field |
| `Enter` | Edit selected field |
| `Tab` | Switch Key ↔ Value |
| `r` | Toggle raw JSON mode |
| `Tab` / `Esc` | Back to URL |

**Response**

| Key | Action |
|---|---|
| `j` / `k` | Scroll |
| `PgUp` / `PgDn` | Scroll by page |
| `e` | Edit URL |
| `Esc` | Back to menu |

---

### Features

#### ✅ Phase 1 — Core (complete)

- GET, POST, PUT, DELETE requests
- Key-Value body builder with auto JSON generation
- Raw JSON editor fallback
- Syntax highlighting (JSON, XML, HTML, YAML) with theme detection
- Scrollable response view
- Persistent request history (`~/.config/apish/history.json`)
- Vim keybindings throughout

#### 🔲 Phase 2 — Webhook Catcher

Start a local listener and inspect incoming webhooks from Stripe, GitHub, or your own services — right in the TUI.

#### 🔲 Phase 3 — AI Debugger

Local LLM integration via **Ollama**. When a request fails, apish analyses the response and tells you exactly what went wrong — without sending data to the cloud.

#### 🔲 Phase 4 — Cloud Share

Press one key to upload a request/response pair to a private VPS and get a shareable link — no more copy-pasting JSON into Slack.

---

### Tech Stack

| | |
|---|---|
| Language | Go ≥ 1.22 |
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Highlighting | [Chroma](https://github.com/alecthomas/chroma) |
| AI (planned) | Ollama |

---

### History

Request history is saved to:

- **Linux / macOS:** `~/.config/apish/history.json`
- **Windows:** `%AppData%\apish\history.json`

Up to 100 entries are stored. Browsing history with `↑`/`↓` restores the full request — URL, method, and body.

---

*Built by [PsydoV2](https://github.com/PsydoV2)*
