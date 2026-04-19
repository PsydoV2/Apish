package webhook

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// Request holds everything captured from a single incoming HTTP request.
type Request struct {
	Method  string
	Path    string
	Headers http.Header
	Body    string
	Time    time.Time
}

// Start launches an HTTP server on port, forwarding every incoming request to ch.
// It blocks until stop is closed or a fatal error occurs.
//
// Two goroutines are involved from the TUI side:
//  1. startWebhookServerCmd — runs this function, returns when server stops
//  2. waitForWebhookCmd     — blocks on ch, re-issues itself after each request
//
// Closing stop terminates both cleanly.
func Start(port int, ch chan<- Request, stop <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		// Non-blocking send: if TUI is slow we drop rather than block the caller
		select {
		case ch <- Request{
			Method:  r.Method,
			Path:    r.URL.RequestURI(),
			Headers: r.Header.Clone(),
			Body:    string(body),
			Time:    time.Now(),
		}:
		default:
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"received","message":"apish webhook catcher"}`)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Listen first so port-in-use errors surface immediately, before blocking.
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("port %d unavailable: %w", port, err)
	}

	go func() {
		<-stop
		srv.Close()
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
