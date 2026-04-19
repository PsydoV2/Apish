package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Response kapselt alles was wir von einem HTTP-Call zurückbekommen.
type Response struct {
	StatusCode  int
	Status      string
	ContentType string
	Body        string
}

var client = &http.Client{
	Timeout: 15 * time.Second,
}

// Do schickt einen HTTP Request mit beliebiger Method und optionalem Body.
// Body kann leer sein (für GET, DELETE). Bei gesetztem Body wird
// Content-Type: application/json automatisch gesetzt.
func Do(method, url, body string) (Response, error) {
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("request erstellen fehlgeschlagen: %w", err)
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("request fehlgeschlagen: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("body lesen fehlgeschlagen: %w", err)
	}

	return Response{
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		ContentType: resp.Header.Get("Content-Type"),
		Body:        string(respBody),
	}, nil
}
