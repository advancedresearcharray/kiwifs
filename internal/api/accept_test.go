package api

import (
	"errors"
	"strings"
	"testing"
)

func TestNegotiateReaderFormat(t *testing.T) {
	tests := []struct {
		accept string
		want   readerFormat
		err    error
	}{
		{"", readerFormatHTML, nil},
		{"text/html", readerFormatHTML, nil},
		{"text/markdown", readerFormatMarkdown, nil},
		{"text/x-markdown", readerFormatMarkdown, nil},
		{"application/json", readerFormatJSON, nil},
		{"*/*", readerFormatHTML, nil},
		{"text/*", readerFormatHTML, nil},
		{"application/*", readerFormatJSON, nil},
		{"application/json, text/html;q=0.9", readerFormatJSON, nil},
		{"text/html, application/json;q=0.8", readerFormatHTML, nil},
		{"text/markdown;q=0.9, text/html;q=0.8", readerFormatMarkdown, nil},
		{"text/html;q=0.5, application/json;q=0.9", readerFormatJSON, nil},
		{"image/png", readerFormatHTML, errAcceptNotAcceptable},
		{"text/html;q=0, application/json;q=0", readerFormatHTML, errAcceptNotAcceptable},
		{"application/xml", readerFormatHTML, errAcceptNotAcceptable},
	}

	for _, tc := range tests {
		t.Run(tc.accept, func(t *testing.T) {
			got, err := negotiateReaderFormat(tc.accept)
			if !errors.Is(err, tc.err) {
				t.Fatalf("negotiateReaderFormat(%q) err = %v, want %v", tc.accept, err, tc.err)
			}
			if got != tc.want {
				t.Fatalf("negotiateReaderFormat(%q) = %v, want %v", tc.accept, got, tc.want)
			}
		})
	}
}

func TestSanitizeAcceptHeader(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{"empty", "", "", nil},
		{"plain", "text/html", "text/html", nil},
		{"strips controls", "text\x00/html", "text/html", nil},
		{"crlf injection", "text/html\r\nX-Injected: true", "", errAcceptInvalid},
		{"truncates long header", strings.Repeat("a", maxAcceptHeaderLen+100), strings.Repeat("a", maxAcceptHeaderLen), nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := sanitizeAcceptHeader(tc.raw)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("sanitizeAcceptHeader() err = %v, want %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("sanitizeAcceptHeader() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseAcceptEntries(t *testing.T) {
	tests := []struct {
		accept string
		want   []acceptEntry
	}{
		{
			accept: "text/html;q=0.8, application/json",
			want: []acceptEntry{
				{mime: "text/html", q: 0.8},
				{mime: "application/json", q: 1.0},
			},
		},
		{
			accept: "text/html;q=bad, text/markdown",
			want: []acceptEntry{
				{mime: "text/html", q: 1.0},
				{mime: "text/markdown", q: 1.0},
			},
		},
		{
			accept: "text/html;q=2, text/markdown",
			want: []acceptEntry{
				{mime: "text/html", q: 1.0},
				{mime: "text/markdown", q: 1.0},
			},
		},
		{
			accept: "text/html, <script>, text/markdown",
			want: []acceptEntry{
				{mime: "text/html", q: 1.0},
				{mime: "text/markdown", q: 1.0},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.accept, func(t *testing.T) {
			got := parseAcceptEntries(tc.accept)
			if len(got) != len(tc.want) {
				t.Fatalf("parseAcceptEntries(%q) len = %d, want %d", tc.accept, len(got), len(tc.want))
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("entry[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseAcceptEntriesCapsEntryCount(t *testing.T) {
	var parts []string
	for i := 0; i < maxAcceptEntries+5; i++ {
		parts = append(parts, "text/html")
	}
	got := parseAcceptEntries(strings.Join(parts, ", "))
	if len(got) != maxAcceptEntries {
		t.Fatalf("len = %d, want %d", len(got), maxAcceptEntries)
	}
}

func TestIsValidMediaRange(t *testing.T) {
	valid := []string{"text/html", "text/*", "*/*", "application/json", "text/x-markdown"}
	for _, mime := range valid {
		if !isValidMediaRange(mime) {
			t.Fatalf("isValidMediaRange(%q) = false, want true", mime)
		}
	}
	invalid := []string{"", "html", "text/<script>", "text html", strings.Repeat("a", 129)}
	for _, mime := range invalid {
		if isValidMediaRange(mime) {
			t.Fatalf("isValidMediaRange(%q) = true, want false", mime)
		}
	}
}
