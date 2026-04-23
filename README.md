# 🐝 apish

> **The terminal-first API client with a brain.** Lighter than Postman, smarter than curl — a single binary that lives in your terminal and gets out of your way.

---

## ✨ Why apish?

Postman is too heavy. `curl` is too manual. `apish` is the sweet spot.

- **Zero bloat** — single Go binary, no Electron, no Docker, no cloud account
- **Local-first** — all data stays on your machine
- **Keyboard-driven** — every action reachable without touching the mouse
- **Persistent history** — every request saved and restored across restarts

---

## 🚀 Features

|                             |                                                                                                 |
| --------------------------- | ----------------------------------------------------------------------------------------------- |
| 🌐 **HTTP methods**         | GET, POST, PUT, PATCH, DELETE — each with its own accent color                                  |
| 🔗 **Query params builder** | Add, edit and delete params in a KV editor; merged into the URL at send time                    |
| 📋 **Request headers**      | Per-request custom headers with the same KV interface                                           |
| 🔐 **Authentication**       | Bearer token, HTTP Basic, or custom API Key header                                              |
| 📦 **Body editor**          | Key-value builder (auto-generates JSON) or raw textarea — toggle with `r`                       |
| 🌍 **Environments**         | Named environments with `{{variable}}` placeholders resolved at send time                       |
| 📥 **curl import**          | Paste any `curl ...` command into the URL field — method, headers and body parsed automatically |
| 📤 **curl export**          | Copy the equivalent curl command to your clipboard in one keypress                              |
| 🪝 **Webhook catcher**      | Start a local HTTP listener; inspect incoming requests in real time                             |
| ⏱ **Response metadata**     | Status code, content type, response time and size for every request                             |
| 🎨 **Syntax highlighting**  | JSON, XML, HTML and YAML — adapts to your terminal color scheme                                 |

---

## 📦 Installation

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

## ⌨️ Keybindings

### Menu

| Key       | Action      |
| --------- | ----------- |
| `j` / `k` | Move cursor |
| `Enter`   | Select      |
| `q`       | Quit        |

---

### Request view

#### Method & URL

| Key         | Action                                                   |
| ----------- | -------------------------------------------------------- |
| `F1` – `F5` | Switch method (GET / POST / PUT / PATCH / DELETE)        |
| `↑` / `↓`   | Browse history                                           |
| `Tab`       | Next section — URL → Params → Headers → Auth → Body      |
| `Esc`       | Previous section                                         |
| `Enter`     | Send (GET / DELETE) or jump to body (POST / PUT / PATCH) |
| `ctrl+s`    | Send from anywhere                                       |
| `ctrl+y`    | Copy curl command to clipboard                           |

> Paste a `curl ...` command into the URL field and press `Enter` to import the full request automatically.

#### Query Params & Request Headers

| Key     | Action                           |
| ------- | -------------------------------- |
| `n`     | Add entry                        |
| `d`     | Delete selected                  |
| `Enter` | Edit selected                    |
| `Tab`   | Switch Key ↔ Value while editing |
| `Esc`   | Cancel / back                    |

#### Auth

| Key | Action                               |
| --- | ------------------------------------ |
| `1` | No auth                              |
| `2` | Bearer token                         |
| `3` | HTTP Basic (username + password)     |
| `4` | API Key (custom header name + value) |

> Sensitive fields (token, password, API key value) are masked with `•` while typing.

#### Body _(POST / PUT / PATCH only)_

| Key     | Action                       |
| ------- | ---------------------------- |
| `n`     | Add field                    |
| `d`     | Delete selected              |
| `Enter` | Edit selected                |
| `r`     | Toggle KV builder ↔ raw JSON |

---

### Response view

| Key             | Action                         |
| --------------- | ------------------------------ |
| `j` / `k`       | Scroll                         |
| `PgUp` / `PgDn` | Page scroll                    |
| `h`             | Toggle response headers        |
| `c`             | Copy curl command to clipboard |
| `e`             | Back to request editor         |
| `Esc`           | Back to menu                   |

---

### Webhook catcher

Start a local HTTP server on any port. Incoming requests stream into a live list — select one to inspect the full headers and body.

| Key     | Action                 |
| ------- | ---------------------- |
| `Enter` | View request detail    |
| `Esc`   | Stop server and return |

---

### Environments

Define named environments with key-value variables. Any `{{variableName}}` in a URL, header, or body is replaced with the matching value from the active environment at send time.

| Key     | Action          |
| ------- | --------------- |
| `n`     | New environment |
| `Enter` | Edit selected   |
| `Space` | Set as active   |
| `d`     | Delete selected |
| `Esc`   | Back to menu    |

Inside the editor:

| Key      | Action                                      |
| -------- | ------------------------------------------- |
| `Tab`    | Switch between name field and variable list |
| `n`      | Add variable                                |
| `d`      | Delete selected variable                    |
| `Enter`  | Edit selected variable                      |
| `ctrl+s` | Save and return                             |
| `Esc`    | Discard and return                          |

---

## 📁 Local data

All data is stored on your machine — nothing leaves without your action.

| File                           | Content                                  |
| ------------------------------ | ---------------------------------------- |
| `~/.config/apish/history.json` | Sent request history (up to 100 entries) |
| `~/.config/apish/envs.json`    | Environment definitions and active index |

> On Windows: `%AppData%\apish\`

---

## 💻 Tech stack

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| Language            | Go ≥ 1.22                                                |
| TUI framework       | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling             | [Lip Gloss](https://github.com/charmbracelet/lipgloss)   |
| Syntax highlighting | [Chroma](https://github.com/alecthomas/chroma)           |
| Clipboard           | [atotto/clipboard](https://github.com/atotto/clipboard)  |

---

_Built with ❤️ by [PsydoV2](https://github.com/PsydoV2)_
