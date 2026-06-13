package api

import (
	"strconv"
	"strings"
)

type readerFormat int

const (
	readerFormatHTML readerFormat = iota
	readerFormatMarkdown
	readerFormatJSON
)

// negotiateReaderFormat picks the best response format from the Accept header.
// Defaults to HTML when Accept is missing or no supported type matches.
func negotiateReaderFormat(accept string) readerFormat {
	accept = strings.TrimSpace(accept)
	if accept == "" {
		return readerFormatHTML
	}

	bestFormat := readerFormatHTML
	bestQ := -1.0

	for _, part := range strings.Split(accept, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		mime := part
		q := 1.0
		if i := strings.Index(part, ";"); i >= 0 {
			mime = strings.TrimSpace(part[:i])
			for _, param := range strings.Split(part[i+1:], ";") {
				param = strings.TrimSpace(param)
				if strings.HasPrefix(strings.ToLower(param), "q=") {
					if v, err := strconv.ParseFloat(strings.TrimSpace(param[2:]), 64); err == nil {
						q = v
					}
				}
			}
		}
		if q < 0 {
			continue
		}

		if format, ok := matchReaderFormat(mime); ok && q > bestQ {
			bestFormat = format
			bestQ = q
		}
	}

	return bestFormat
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
