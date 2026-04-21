package tui

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
)

// applyVars substituiert {{name}}-Platzhalter mit den Werten aus vars.
func applyVars(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

// buildFinalURL merged Query-Parameter in die Basis-URL.
func buildFinalURL(base string, params []kvPair, vars map[string]string) string {
	active := make([]kvPair, 0, len(params))
	for _, p := range params {
		if p.key != "" {
			active = append(active, p)
		}
	}
	if len(active) == 0 {
		return base
	}
	u, err := url.Parse(base)
	if err != nil {
		return base
	}
	q := u.Query()
	for _, p := range active {
		q.Set(applyVars(p.key, vars), applyVars(p.value, vars))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// buildCurlCmd baut einen curl-Befehl aus den Request-Parametern.
func buildCurlCmd(method, rawURL, body string, headers map[string]string) string {
	var b strings.Builder
	b.WriteString("curl")
	if method != "GET" {
		fmt.Fprintf(&b, " -X %s", method)
	}
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, " -H %q", k+": "+headers[k])
	}
	if body != "" {
		fmt.Fprintf(&b, " -d %q", body)
	}
	fmt.Fprintf(&b, " %q", rawURL)
	return b.String()
}

// parseCurlCmd parst einen curl-Befehl und extrahiert Method, URL, Body und Headers.
func parseCurlCmd(input string) (method, rawURL, body string, headers map[string]string, err error) {
	tokens := shellTokenize(input)
	if len(tokens) == 0 || strings.ToLower(tokens[0]) != "curl" {
		return "", "", "", nil, fmt.Errorf("kein curl-Befehl")
	}
	method = "GET"
	headers = make(map[string]string)
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		switch tok {
		case "-X", "--request":
			if i+1 < len(tokens) {
				i++
				method = strings.ToUpper(tokens[i])
			}
		case "-H", "--header":
			if i+1 < len(tokens) {
				i++
				k, v := splitHeaderStr(tokens[i])
				if k != "" {
					headers[k] = v
				}
			}
		case "-d", "--data", "--data-raw", "--data-ascii", "--data-binary":
			if i+1 < len(tokens) {
				i++
				body = tokens[i]
				if method == "GET" {
					method = "POST"
				}
			}
		case "-u", "--user":
			if i+1 < len(tokens) {
				i++
				headers["Authorization"] = "Basic " + tokens[i]
			}
		default:
			// Flags wie -L, -v, --silent etc. überspringen
			if !strings.HasPrefix(tok, "-") && rawURL == "" {
				rawURL = tok
			}
		}
	}
	if rawURL == "" {
		return "", "", "", nil, fmt.Errorf("keine URL im curl-Befehl gefunden")
	}
	return method, rawURL, body, headers, nil
}

func splitHeaderStr(h string) (key, value string) {
	idx := strings.Index(h, ":")
	if idx < 0 {
		return strings.TrimSpace(h), ""
	}
	return strings.TrimSpace(h[:idx]), strings.TrimSpace(h[idx+1:])
}

// shellTokenize zerlegt einen Shell-Befehl in Tokens (behandelt Quotes).
func shellTokenize(input string) []string {
	var tokens []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	for i := 0; i < len(input); i++ {
		c := input[i]
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case c == '\\' && inDouble && i+1 < len(input):
			i++
			cur.WriteByte(input[i])
		case (c == ' ' || c == '\t') && !inSingle && !inDouble:
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

func copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}
