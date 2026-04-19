package tui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// kvPair ist ein einzelnes Key-Value-Paar im Body-Builder.
type kvPair struct {
	key   string
	value string
}

// bodyMode beschreibt ob der Body als KV-Builder oder als Raw-Textarea angezeigt wird.
type bodyMode int

const (
	bodyModeKV  bodyMode = iota // strukturierter Key-Value-Editor
	bodyModeRaw                 // freie JSON-Textarea
)

// kvToJSON wandelt die KV-Paare in einen formatierten JSON-String um.
// Leere Keys werden übersprungen.
func kvToJSON(pairs []kvPair) string {
	if len(pairs) == 0 {
		return ""
	}

	var lines []string
	for _, p := range pairs {
		if p.key == "" {
			continue
		}
		// json.Marshal für den Key stellt sicher dass Sonderzeichen escaped werden
		keyBytes, _ := json.Marshal(p.key)
		lines = append(lines, fmt.Sprintf("  %s: %s", string(keyBytes), formatValue(p.value)))
	}

	if len(lines) == 0 {
		return ""
	}
	return "{\n" + strings.Join(lines, ",\n") + "\n}"
}

// formatValue entscheidet ob ein Wert als Zahl, Boolean, null oder String ausgegeben wird.
// Das ist Go's Weg für "duck typing" — wir versuchen verschiedene Parse-Arten.
func formatValue(v string) string {
	// Zahl?
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return v
	}
	// Boolean oder null?
	switch v {
	case "true", "false", "null":
		return v
	}
	// Alles andere: JSON-String (mit korrektem Escaping via json.Marshal)
	b, _ := json.Marshal(v)
	return string(b)
}

// jsonToKV versucht einen JSON-String in KV-Paare zu parsen.
// Funktioniert nur für flache Objekte (kein Nesting).
// Bei Fehler oder komplexen Strukturen: nil zurückgeben.
func jsonToKV(raw string) []kvPair {
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return nil
	}

	pairs := make([]kvPair, 0, len(obj))
	for k, v := range obj {
		var valueStr string
		switch val := v.(type) {
		case string:
			valueStr = val
		case float64:
			// Ganzzahl ohne Dezimalpunkt darstellen
			if val == float64(int64(val)) {
				valueStr = fmt.Sprintf("%d", int64(val))
			} else {
				valueStr = fmt.Sprintf("%g", val)
			}
		case bool:
			valueStr = fmt.Sprintf("%t", val)
		case nil:
			valueStr = "null"
		default:
			// Nested object/array → als JSON-String behalten
			b, _ := json.Marshal(val)
			valueStr = string(b)
		}
		pairs = append(pairs, kvPair{key: k, value: valueStr})
	}
	return pairs
}
