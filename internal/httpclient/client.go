package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Response struct {
	StatusCode  int
	Status      string
	ContentType string
	Headers     http.Header
	Body        string
	Duration    time.Duration
	Size        int
}

var client = &http.Client{
	Timeout: 15 * time.Second,
}

func Do(method, url, body string, extraHeaders map[string]string) (Response, error) {
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

	for k, v := range extraHeaders {
		if k != "" {
			req.Header.Set(k, v)
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)
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
		Headers:     resp.Header,
		Body:        string(respBody),
		Duration:    duration,
		Size:        len(respBody),
	}, nil
}
