package api

import (
	"errors"
	"strconv"
	"strings"
)

type readerFormat int

const (
	readerFormatHTML readerFormat = iota
	readerFormatMarkdown
	readerFormatJSON
)

const (
	maxAcceptHeaderLen     = 4096
	maxAcceptEntries       = 32
	readerSupportedFormats = "text/html, text/markdown, application/json"
)

var (
	errAcceptInvalid       = errors.New("invalid Accept header")
	errAcceptNotAcceptable = errors.New("unsupported Accept header")
)

type acceptEntry struct {
	mime string
	q    float64
}

// negotiateReaderFormat picks the best response format from the Accept header.
// Returns HTML when Accept is missing. Returns errAcceptNotAcceptable when the
// client sent Accept values that match none of the supported reader formats.
func negotiateReaderFormat(rawAccept string) (readerFormat, error) {
	accept, err := sanitizeAcceptHeader(rawAccept)
	if err != nil {
		return readerFormatHTML, err
	}
	if accept == "" {
		return readerFormatHTML, nil
	}

	entries := parseAcceptEntries(accept)
	if len(entries) == 0 {
		return readerFormatHTML, errAcceptNotAcceptable
	}

	bestFormat := readerFormatHTML
	bestQ := -1.0
	found := false

	for _, entry := range entries {
		if entry.q <= 0 {
			continue
		}
		format, ok := matchReaderFormat(entry.mime)
		if !ok {
			continue
		}
		if !found || entry.q > bestQ {
			bestFormat = format
			bestQ = entry.q
			found = true
		}
	}

	if !found {
		return readerFormatHTML, errAcceptNotAcceptable
	}
	return bestFormat, nil
}

// sanitizeAcceptHeader strips control characters and enforces a length cap.
// Returns errAcceptInvalid when the raw header contains CR/LF (header injection).
func sanitizeAcceptHeader(raw string) (string, error) {
	if strings.ContainsAny(raw, "\r\n") {
		return "", errAcceptInvalid
	}
	s := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, raw)
	if len(s) > maxAcceptHeaderLen {
		s = s[:maxAcceptHeaderLen]
	}
	return strings.TrimSpace(s), nil
}

func parseAcceptEntries(accept string) []acceptEntry {
	parts := strings.Split(accept, ",")
	if len(parts) > maxAcceptEntries {
		parts = parts[:maxAcceptEntries]
	}

	var entries []acceptEntry
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		entry, ok := parseAcceptEntry(part)
		if !ok {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

func parseAcceptEntry(part string) (acceptEntry, bool) {
	mime := part
	q := 1.0
	if i := strings.Index(part, ";"); i >= 0 {
		mime = strings.TrimSpace(part[:i])
		q = parseAcceptQValue(part[i+1:])
	}
	if !isValidMediaRange(mime) {
		return acceptEntry{}, false
	}
	return acceptEntry{mime: strings.ToLower(mime), q: q}, true
}

func parseAcceptQValue(params string) float64 {
	for _, param := range strings.Split(params, ";") {
		param = strings.TrimSpace(param)
		if !strings.HasPrefix(strings.ToLower(param), "q=") {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(param[2:]), 64)
		if err != nil || v < 0 || v > 1 {
			return 1.0
		}
		return v
	}
	return 1.0
}

func isValidMediaRange(mime string) bool {
	if mime == "" || len(mime) > 128 {
		return false
	}
	for _, r := range mime {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '/', r == '.', r == '+', r == '-', r == '*', r == '_':
		default:
			return false
		}
	}
	return strings.Contains(mime, "/")
}

func matchReaderFormat(mime string) (readerFormat, bool) {
	mime = strings.ToLower(strings.TrimSpace(mime))
	switch mime {
	case "text/html", "text/*", "*/*":
		return readerFormatHTML, true
	case "text/markdown", "text/x-markdown":
		return readerFormatMarkdown, true
	case "application/json", "application/*":
		return readerFormatJSON, true
	default:
		return readerFormatHTML, false
	}
}
