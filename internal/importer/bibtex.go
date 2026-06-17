package importer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nickng/bibtex"
)

// BibTeXSource implements Source for .bib reference files.
type BibTeXSource struct {
	filePath string
}

// NewBibTeX creates a BibTeX source from a .bib file path.
func NewBibTeX(filePath string) (*BibTeXSource, error) {
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("bibtex file: %w", err)
	}
	return &BibTeXSource{filePath: filePath}, nil
}

func (s *BibTeXSource) Name() string {
	base := filepath.Base(s.filePath)
	base = strings.TrimSuffix(base, ".bib")
	base = strings.TrimSuffix(base, ".bibtex")
	return base
}

func (s *BibTeXSource) Stream(ctx context.Context) (<-chan Record, <-chan error) {
	records := make(chan Record, 64)
	errs := make(chan error, 1)

	go func() {
		defer close(records)
		defer close(errs)

		data, err := os.ReadFile(s.filePath)
		if err != nil {
			errs <- fmt.Errorf("read bibtex: %w", err)
			return
		}

		parsed, err := bibtex.Parse(strings.NewReader(string(data)))
		if err != nil {
			errs <- fmt.Errorf("parse bibtex: %w", err)
			return
		}

		name := s.Name()
		for i, entry := range parsed.Entries {
			if ctx.Err() != nil {
				return
			}

			fields, rawContent := bibEntryToRecord(entry)
			pk := entry.CiteName
			if pk == "" {
				pk = fmt.Sprintf("entry_%d", i)
			}

			rec := Record{
				SourceID:   fmt.Sprintf("bibtex:%s:%s", name, pk),
				SourceDSN:  s.filePath,
				Table:      name,
				Fields:     fields,
				PrimaryKey: pk,
			}
			rec.Fields["_raw_content"] = rawContent

			select {
			case records <- rec:
			case <-ctx.Done():
				return
			}
		}
	}()
	return records, errs
}

func (s *BibTeXSource) Close() error { return nil }

var authorSplitRE = regexp.MustCompile(`(?i)\s+and\s+`)
var yamlPlainScalarRE = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func bibEntryToRecord(entry *bibtex.BibEntry) (map[string]any, string) {
	rawFields := make(map[string]string, len(entry.Fields))
	for k, v := range entry.Fields {
		rawFields[strings.ToLower(strings.TrimSpace(k))] = unescapeBibTeX(v.String())
	}

	fields := make(map[string]any, len(rawFields)+6)
	fields["bibtex_key"] = entry.CiteName
	fields["bibtex_type"] = entry.Type

	if title, ok := rawFields["title"]; ok && title != "" {
		fields["title"] = title
	}

	if authors := parseBibAuthors(rawFields["author"]); len(authors) > 0 {
		fields["authors"] = authors
	}

	if year := parseBibYear(rawFields["year"]); year > 0 {
		fields["year"] = year
	}

	venue := firstNonEmpty(
		rawFields["journal"],
		rawFields["booktitle"],
		rawFields["publisher"],
		rawFields["howpublished"],
	)
	if venue != "" {
		fields["venue"] = venue
	}

	for _, key := range []string{"doi", "url", "isbn", "issn", "abstract", "pages", "volume", "number", "month", "address", "edition", "series", "organization", "school", "institution", "chapter", "note"} {
		if val, ok := rawFields[key]; ok && val != "" {
			fields[key] = val
		}
	}

	if tags := parseBibTags(rawFields["keywords"]); len(tags) > 0 {
		fields["tags"] = tags
	}

	mapped := map[string]bool{
		"title": true, "author": true, "year": true, "journal": true, "booktitle": true,
		"publisher": true, "howpublished": true, "keywords": true,
	}
	for k, v := range rawFields {
		if mapped[k] || v == "" {
			continue
		}
		if _, exists := fields[k]; !exists {
			fields[k] = v
		}
	}

	title, _ := fields["title"].(string)
	if title == "" {
		title = entry.CiteName
	}
	rawContent := buildBibTeXMarkdown(fields, title, entry.CiteName, entry.Type, authorsFromFields(fields), venue, parseBibYear(rawFields["year"]))
	return fields, rawContent
}

func authorsFromFields(fields map[string]any) []string {
	raw, ok := fields["authors"].([]string)
	if !ok {
		return nil
	}
	return raw
}

func parseBibAuthors(author string) []string {
	author = strings.TrimSpace(author)
	if author == "" {
		return nil
	}
	parts := authorSplitRE.Split(author, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseBibYear(year string) int {
	year = strings.TrimSpace(year)
	if year == "" {
		return 0
	}
	if n, err := strconv.Atoi(year); err == nil {
		return n
	}
	return 0
}

func parseBibTags(keywords string) []string {
	keywords = strings.TrimSpace(keywords)
	if keywords == "" {
		return nil
	}
	parts := strings.Split(keywords, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

var bibTeXAcute = map[byte]string{
	'a': "á", 'e': "é", 'i': "í", 'o': "ó", 'u': "ú", 'y': "ý",
	'A': "Á", 'E': "É", 'I': "Í", 'O': "Ó", 'U': "Ú", 'Y': "Ý",
}

var bibTeXGrave = map[byte]string{
	'a': "à", 'e': "è", 'i': "ì", 'o': "ò", 'u': "ù",
	'A': "À", 'E': "È", 'I': "Ì", 'O': "Ò", 'U': "Ù",
}

var bibTeXUmlaut = map[byte]string{
	'a': "ä", 'e': "ë", 'i': "ï", 'o': "ö", 'u': "ü", 'y': "ÿ",
	'A': "Ä", 'E': "Ë", 'I': "Ï", 'O': "Ö", 'U': "Ü", 'Y': "Ÿ",
}

func unescapeBibTeX(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		next := s[i+1]
		switch next {
		case '{', '}', '&', '%', '$', '#', '_':
			b.WriteByte(next)
			i++
		case '-':
			b.WriteByte('-')
			i++
		case '~':
			b.WriteByte(' ')
			i++
		case '\\':
			b.WriteByte('\\')
			i++
		case '\'', '`', '"', '^', 'c':
			if i+2 < len(s) {
				if ch, ok := bibTeXAccent(next, s[i+2]); ok {
					b.WriteString(ch)
					i += 2
					continue
				}
			}
			b.WriteByte('\\')
		default:
			b.WriteByte('\\')
		}
	}
	return b.String()
}

func bibTeXAccent(cmd, letter byte) (string, bool) {
	switch cmd {
	case '\'':
		if v, ok := bibTeXAcute[letter]; ok {
			return v, true
		}
	case '`':
		if v, ok := bibTeXGrave[letter]; ok {
			return v, true
		}
	case '"':
		if v, ok := bibTeXUmlaut[letter]; ok {
			return v, true
		}
	case '^':
		switch letter {
		case 'a', 'A':
			return "â", true
		case 'e', 'E':
			return "ê", true
		case 'i', 'I':
			return "î", true
		case 'o', 'O':
			return "ô", true
		case 'u', 'U':
			return "û", true
		}
	case 'c':
		switch letter {
		case 'c':
			return "ç", true
		case 'C':
			return "Ç", true
		}
	}
	return "", false
}

func buildBibTeXMarkdown(fields map[string]any, title, citeKey, entryType string, authors []string, venue string, year int) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "bibtex_key: %s\n", yamlScalar(citeKey))
	fmt.Fprintf(&b, "bibtex_type: %s\n", yamlScalar(entryType))
	if title != "" {
		fmt.Fprintf(&b, "title: %q\n", title)
	}
	if len(authors) > 0 {
		b.WriteString("authors:\n")
		for _, a := range authors {
			fmt.Fprintf(&b, "  - %q\n", a)
		}
	}
	if year > 0 {
		fmt.Fprintf(&b, "year: %d\n", year)
	}
	if venue != "" {
		fmt.Fprintf(&b, "venue: %q\n", venue)
	}
	for _, key := range []string{"doi", "url", "isbn", "abstract", "pages", "volume", "number", "month"} {
		if val, ok := fields[key].(string); ok && val != "" {
			if key == "abstract" && strings.Contains(val, "\n") {
				b.WriteString("abstract: |\n")
				for _, line := range strings.Split(strings.TrimRight(val, "\n"), "\n") {
					fmt.Fprintf(&b, "  %s\n", line)
				}
			} else {
				fmt.Fprintf(&b, "%s: %q\n", key, val)
			}
		}
	}
	if tags, ok := fields["tags"].([]string); ok && len(tags) > 0 {
		b.WriteString("tags: [")
		for i, tag := range tags {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%s", yamlScalar(tag))
		}
		b.WriteString("]\n")
	}
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# %s\n\n", title)
	b.WriteString(buildBibCitationLine(authors, year, venue))
	return b.String()
}

func yamlScalar(s string) string {
	if s == "" {
		return `""`
	}
	if yamlPlainScalarRE.MatchString(s) {
		return s
	}
	return fmt.Sprintf("%q", s)
}

func buildBibCitationLine(authors []string, year int, venue string) string {
	var b strings.Builder
	switch len(authors) {
	case 0:
	case 1:
		b.WriteString(authors[0])
	case 2:
		b.WriteString(authors[0])
		b.WriteString(" and ")
		b.WriteString(authors[1])
	default:
		b.WriteString(strings.Join(authors[:len(authors)-1], ", "))
		b.WriteString(", and ")
		b.WriteString(authors[len(authors)-1])
	}
	if year > 0 {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "(%d)", year)
	}
	if venue != "" {
		b.WriteString(". *")
		b.WriteString(venue)
		b.WriteString("*.")
	} else if b.Len() > 0 {
		b.WriteString(".")
	}
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	return b.String()
}
