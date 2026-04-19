# 🤖 AI Agent Instructions for `apish`

Welcome, AI Agent! You are assisting in the development of `apish`, a blazing-fast, lightweight Terminal User Interface (TUI) API client built in Go.

Please read and strictly adhere to the following architectural guidelines, coding standards, and project goals before suggesting any code.

## 🎯 Project Context & Vision

`apish` is designed for developers who want a local-first, zero-bloat API client. It has three core distinguishing features ("twists") that dictate its architecture:

1. **Webhook Catcher:** Can start a local server to listen for incoming requests.
2. **Local AI Debugger:** Integrates with local Ollama instances to analyze failed API calls.
3. **VPS Sharing:** Will feature a companion backend for generating shareable links of request/response payloads.

## 🏗 Architecture & Directory Structure

This project strictly follows Go standard project layout conventions:

- `/cmd/apish/main.go`: The entry point. It should ONLY initialize the TUI model and start the program. No business logic here.
- `/internal/`: All core logic goes here. Do NOT place logic in the root or `/cmd/`.
  - `/internal/tui/`: Contains the UI logic using the Bubble Tea framework.
  - `/internal/httpclient/`: Handles all HTTP networking (sending and receiving).
  - `/internal/config/`: Manages user settings, local history, and state persistence.

## 💻 Tech Stack & Libraries

- **Language:** Go (Golang) >= 1.22
- **TUI Framework:** `github.com/charmbracelet/bubbletea`
- **Styling:** `github.com/charmbracelet/lipgloss`
- **Dependencies:** Keep external dependencies to an absolute minimum. Use the Go standard library (`net/http`, `encoding/json`, `io`) wherever possible.

## 🚦 Bubble Tea Rules (CRITICAL)

When writing or modifying code in `/internal/tui/`:

1. **The Elm Architecture:** Strictly adhere to the `Model`, `Update`, and `View` paradigm.
2. **Immutability in Update:** The `Update` function must return a new copy of the model or modify it by value.
3. **Pure Views:** The `View` function MUST NOT have any side effects. It should only render the current state of the model to a string.
4. **No Blocking Operations:** NEVER run blocking operations (like HTTP requests or disk I/O) directly inside `Update`. Always use `tea.Cmd` to execute them asynchronously and return a `tea.Msg` when done.

## 🛠 Go Coding Standards

- **Error Handling:** Always check errors. Wrap errors with context using `fmt.Errorf("failed to do X: %w", err)`. Never ignore errors silently.
- **Naming Conventions:** Use short, concise variable names (e.g., `req` instead of `httpRequest`, `err` instead of `errorObject`). Use CamelCase for Go structs and variables.
- **Formatting:** Code must perfectly comply with `gofmt`.
- **Comments:** Provide concise godoc comments for exported types and functions.

## 🧠 Workflow for Feature Implementation

When asked to implement a new feature:

1. Determine if it's UI logic or Core logic.
2. If Core logic (e.g., making an HTTP call), implement it in the appropriate `/internal/` subpackage first.
3. Create a corresponding `tea.Msg` struct to pass the result to the TUI.
4. Write a `tea.Cmd` function to execute the logic asynchronously.
5. Update the `Update` function to handle the new command and the resulting message.
6. Update the `View` to reflect the changes.
