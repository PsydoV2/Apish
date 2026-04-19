package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Response kapselt alles was wir von einem HTTP-Call zurückbekommen.
// Exportierte Felder (Großbuchstabe) sind von außen lesbar.
type Response struct {
	StatusCode int
	Status     string
	Body       string
}

// client ist unser wiederverwendbarer HTTP-Client mit Timeout.
// package-level Variable: existiert einmal für das gesamte Paket.
var client = &http.Client{
	Timeout: 15 * time.Second,
}

// Get schickt einen HTTP GET Request an die angegebene URL.
// Gibt (Response, nil) bei Erfolg oder (Response{}, error) bei Fehler zurück.
// Das ist das Standard Go-Idiom: (Wert, Fehler) als Rückgabe.
func Get(url string) (Response, error) {
	resp, err := client.Get(url)
	if err != nil {
		return Response{}, fmt.Errorf("request fehlgeschlagen: %w", err)
	}
	defer resp.Body.Close() // defer = "führe das aus wenn die Funktion endet" — wichtig gegen Memory Leaks

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("body lesen fehlgeschlagen: %w", err)
	}

	return Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       string(body),
	}, nil
}
