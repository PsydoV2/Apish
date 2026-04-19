package tui

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

// highlight formatiert den Body und fügt Syntax Highlighting hinzu.
// contentType ist der Wert des Content-Type Headers (kann leer sein).
func highlight(body, contentType string) string {
	lang := detectLanguage(body, contentType)

	// Schritt 1: Pretty-print wenn möglich (nur JSON via stdlib — kein extra Dep)
	body = prettyPrint(body, lang)

	// Schritt 2: Syntax Highlighting mit chroma
	return chromaHighlight(body, lang)
}

// detectLanguage leitet die Sprache aus dem Content-Type Header ab,
// oder macht Content-Sniffing als Fallback.
func detectLanguage(body, contentType string) string {
	ct := strings.ToLower(contentType)

	switch {
	case strings.Contains(ct, "json"):
		return "json"
	case strings.Contains(ct, "xml"):
		return "xml"
	case strings.Contains(ct, "html"):
		return "html"
	case strings.Contains(ct, "yaml"):
		return "yaml"
	}

	// Fallback: Content Sniffing
	trimmed := strings.TrimSpace(body)
	switch {
	case strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "["):
		return "json"
	case strings.HasPrefix(trimmed, "<"):
		return "html"
	}

	return "text"
}

// prettyPrint formatiert den Body lesbar — aktuell nur für JSON.
// encoding/json ist stdlib — kein extra Dep.
func prettyPrint(body, lang string) string {
	if lang != "json" {
		return body
	}

	var buf bytes.Buffer
	// json.Indent schreibt formatierten JSON in buf.
	// Prefix = "", Indent = zwei Spaces — Standard JSON-Formatting.
	if err := json.Indent(&buf, []byte(body), "", "  "); err != nil {
		return body // Falls kein valider JSON: Original zurückgeben
	}
	return buf.String()
}

// chromaHighlight wendet Syntax Highlighting an und gibt einen String
// mit ANSI-Escape-Codes zurück, die das Terminal als Farben interpretiert.
func chromaHighlight(body, lang string) string {
	// Lexer: versteht die Syntax der Sprache (tokenisiert den Code)
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback // Plain text als Fallback
	}

	// Style: das Farbschema — adaptiv je nach Terminal-Hintergrund
	styleName := "dracula"
	if !lipgloss.HasDarkBackground() {
		styleName = "github"
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}

	// Formatter: produziert ANSI-Terminal-Ausgabe (256 Farben)
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return body
	}

	var buf bytes.Buffer
	// Tokenize → Format ist die chroma Pipeline
	iterator, err := lexer.Tokenise(nil, body)
	if err != nil {
		return body
	}
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return body
	}

	return buf.String()
}
