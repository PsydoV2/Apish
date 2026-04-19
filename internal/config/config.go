package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const maxHistoryEntries = 100

// HistoryEntry speichert einen einzelnen gesendeten Request vollständig.
type HistoryEntry struct {
	URL    string `json:"url"`
	Method string `json:"method"`
	Body   string `json:"body,omitempty"` // generierter JSON-Body, leer für GET/DELETE
}

// History ist der gesamte persistierte Zustand von apish.
type History struct {
	Entries []HistoryEntry `json:"entries"`
}

// Add hängt einen neuen Eintrag an — Duplikate am Ende werden übersprungen.
func (h *History) Add(entry HistoryEntry) {
	if len(h.Entries) > 0 {
		last := h.Entries[len(h.Entries)-1]
		if last.URL == entry.URL && last.Method == entry.Method && last.Body == entry.Body {
			return
		}
	}
	h.Entries = append(h.Entries, entry)

	// Älteste Einträge entfernen wenn das Limit erreicht ist
	if len(h.Entries) > maxHistoryEntries {
		h.Entries = h.Entries[len(h.Entries)-maxHistoryEntries:]
	}
}

// LoadHistory liest die History-Datei. Gibt eine leere History zurück
// wenn die Datei noch nicht existiert (erster Start).
func LoadHistory() (History, error) {
	path, err := historyPath()
	if err != nil {
		return History{}, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return History{}, nil // Erster Start — kein Fehler
	}
	if err != nil {
		return History{}, fmt.Errorf("history lesen fehlgeschlagen: %w", err)
	}

	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return History{}, fmt.Errorf("history parsen fehlgeschlagen: %w", err)
	}
	return h, nil
}

// SaveHistory schreibt die History in die Datei.
// Erstellt das Verzeichnis automatisch falls es noch nicht existiert.
func SaveHistory(h History) error {
	path, err := historyPath()
	if err != nil {
		return err
	}

	// os.MkdirAll erstellt alle fehlenden Verzeichnisse in einem Schritt.
	// 0755 = rwxr-xr-x: Owner darf alles, andere nur lesen/ausführen.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("config-verzeichnis anlegen fehlgeschlagen: %w", err)
	}

	// MarshalIndent = menschenlesbares JSON mit Einrückung
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("history serialisieren fehlgeschlagen: %w", err)
	}

	// 0644 = rw-r--r--: nur Owner darf schreiben
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("history speichern fehlgeschlagen: %w", err)
	}
	return nil
}

// historyPath gibt den vollständigen Pfad zur History-Datei zurück.
// os.UserConfigDir() liefert:
//   - Linux:   ~/.config
//   - Windows: %AppData%
//   - macOS:   ~/Library/Application Support
func historyPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config-verzeichnis nicht gefunden: %w", err)
	}
	return filepath.Join(base, "apish", "history.json"), nil
}
