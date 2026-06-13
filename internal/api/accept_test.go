package api

import "testing"

func TestNegotiateReaderFormat(t *testing.T) {
	tests := []struct {
		accept string
		want   readerFormat
	}{
		{"", readerFormatHTML},
		{"text/html", readerFormatHTML},
		{"text/markdown", readerFormatMarkdown},
		{"text/x-markdown", readerFormatMarkdown},
		{"application/json", readerFormatJSON},
		{"*/*", readerFormatHTML},
		{"text/*", readerFormatHTML},
		{"application/*", readerFormatJSON},
		{"application/json, text/html;q=0.9", readerFormatJSON},
		{"text/html, application/json;q=0.8", readerFormatHTML},
		{"text/markdown;q=0.9, text/html;q=0.8", readerFormatMarkdown},
		{"text/html;q=0.5, application/json;q=0.9", readerFormatJSON},
		{"image/png", readerFormatHTML},
	}

	for _, tc := range tests {
		t.Run(tc.accept, func(t *testing.T) {
			if got := negotiateReaderFormat(tc.accept); got != tc.want {
				t.Fatalf("negotiateReaderFormat(%q) = %v, want %v", tc.accept, got, tc.want)
			}
		})
	}
}
