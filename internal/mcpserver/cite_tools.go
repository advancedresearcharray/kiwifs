package mcpserver

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	defaultCrossrefWorksURL = "https://api.crossref.org/works/"
	defaultArxivQueryURL    = "https://export.arxiv.org/api/query"
	defaultCiteUserAgent    = "kiwifs/1.0 (mailto:support@kiwifs.io)"
	maxCiteIdentifierLen    = 256
)

var (
	arxivIDPattern = regexp.MustCompile(`(?i)(?:arxiv:/)?(\d{4}\.\d{4,5})(?:v\d+)?`)
	// DOI suffix allows the Crossref-registered character set; reject path/query injection.
	doiPattern = regexp.MustCompile(`(?i)^10\.\d{4,9}/[-._;()/:a-z0-9]+$`)
	unsafeCiteChars = regexp.MustCompile(`[\x00-\x1f\x7f\\]`)
)

type paperMetadata struct {
	Title     string
	Authors   []string
	Year      int
	Venue     string
	DOI       string
	ArxivID   string
	Abstract  string
	BibtexKey string
	BibTeX    string
}

type citeHTTPClient struct {
	http        *http.Client
	userAgent   string
	crossrefURL string
	arxivURL    string
}

func newDefaultCiteHTTPClient() *citeHTTPClient {
	return &citeHTTPClient{
		http:        &http.Client{Timeout: 30 * time.Second},
		userAgent:   defaultCiteUserAgent,
		crossrefURL: defaultCrossrefWorksURL,
		arxivURL:    defaultArxivQueryURL,
	}
}

func sanitizeCiteInput(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("identifier is required")
	}
	if len(s) > maxCiteIdentifierLen {
		return "", fmt.Errorf("identifier exceeds maximum length of %d", maxCiteIdentifierLen)
	}
	if unsafeCiteChars.MatchString(s) {
		return "", fmt.Errorf("identifier contains invalid characters")
	}
	// Allow https:// in URL forms; reject bare path traversal sequences.
	if !strings.Contains(s, "://") && (strings.Contains(s, "..") || strings.Contains(s, "//")) {
		return "", fmt.Errorf("identifier contains unsafe path sequences")
	}
	return s, nil
}

func normalizeDOI(raw string) string {
	s, err := sanitizeCiteInput(raw)
	if err != nil {
		return ""
	}
	s = strings.TrimPrefix(s, "doi:")
	s = strings.TrimPrefix(s, "DOI:")
	if u, err := url.Parse(s); err == nil && u.Host != "" {
		if strings.EqualFold(u.Host, "doi.org") {
			s = strings.TrimPrefix(u.Path, "/")
		} else {
			return ""
		}
	}
	if !isValidDOI(s) {
		return ""
	}
	return s
}

func isValidDOI(doi string) bool {
	return doiPattern.MatchString(doi)
}

func normalizeArxivID(raw string) string {
	s, err := sanitizeCiteInput(raw)
	if err != nil {
		return ""
	}
	if m := arxivIDPattern.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	if u, err := url.Parse(s); err == nil && strings.Contains(u.Host, "arxiv.org") {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		for i, p := range parts {
			if p == "abs" || p == "pdf" {
				if i+1 < len(parts) {
					if m := arxivIDPattern.FindStringSubmatch(parts[i+1]); len(m) > 1 {
						return m[1]
					}
				}
			}
		}
	}
	return ""
}

func isValidArxivID(id string) bool {
	return arxivIDPattern.MatchString(id)
}

func validateBibtexKey(key string) error {
	if key == "" {
		return fmt.Errorf("empty bibtex key")
	}
	if strings.Contains(key, "/") || strings.Contains(key, "\\") || strings.Contains(key, "..") {
		return fmt.Errorf("unsafe bibtex key")
	}
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`).MatchString(key) {
		return fmt.Errorf("invalid bibtex key")
	}
	return nil
}

func isArxivIdentifier(raw string) bool {
	return normalizeArxivID(raw) != ""
}

func citeErrorResult(query, msg string) *mcp.CallToolResult {
	payload, _ := json.Marshal(map[string]any{
		"success": false,
		"error":   msg,
		"query":   query,
	})
	return mcp.NewToolResultError(string(payload))
}

func (c *citeHTTPClient) assertCrossrefURL(reqURL string) error {
	if strings.HasPrefix(c.crossrefURL, defaultCrossrefWorksURL) {
		return assertCiteRequestURL(reqURL, "api.crossref.org")
	}
	return nil
}

func (c *citeHTTPClient) assertArxivURL(reqURL string) error {
	if strings.HasPrefix(c.arxivURL, defaultArxivQueryURL) {
		return assertCiteRequestURL(reqURL, "export.arxiv.org")
	}
	return nil
}

func (c *citeHTTPClient) fetchDOI(ctx context.Context, doi string) (*paperMetadata, error) {
	doi = normalizeDOI(doi)
	if doi == "" {
		return nil, fmt.Errorf("invalid DOI format")
	}

	reqURL := c.crossrefURL + url.PathEscape(doi)
	if err := c.assertCrossrefURL(reqURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("crossref request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("DOI not found")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("Crossref rate limit exceeded")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Crossref API error: HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Message struct {
			Title          []string `json:"title"`
			Author         []struct {
				Given  string `json:"given"`
				Family string `json:"family"`
			} `json:"author"`
			DOI            string `json:"DOI"`
			Abstract       string `json:"abstract"`
			ContainerTitle []string `json:"container-title"`
			Issued         struct {
				DateParts [][]int `json:"date-parts"`
			} `json:"issued"`
			PublishedPrint struct {
				DateParts [][]int `json:"date-parts"`
			} `json:"published-print"`
			PublishedOnline struct {
				DateParts [][]int `json:"date-parts"`
			} `json:"published-online"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse Crossref response: %w", err)
	}

	title := firstNonEmpty(payload.Message.Title)
	if title == "" {
		return nil, fmt.Errorf("DOI not found")
	}

	authors := make([]string, 0, len(payload.Message.Author))
	for _, a := range payload.Message.Author {
		name := strings.TrimSpace(strings.TrimSpace(a.Family) + ", " + strings.TrimSpace(a.Given))
		name = strings.Trim(name, ", ")
		if name != "" {
			authors = append(authors, name)
		}
	}

	year := yearFromDateParts(payload.Message.PublishedPrint.DateParts)
	if year == 0 {
		year = yearFromDateParts(payload.Message.PublishedOnline.DateParts)
	}
	if year == 0 {
		year = yearFromDateParts(payload.Message.Issued.DateParts)
	}

	meta := &paperMetadata{
		Title:    title,
		Authors:  authors,
		Year:     year,
		Venue:    firstNonEmpty(payload.Message.ContainerTitle),
		DOI:      payload.Message.DOI,
		Abstract: stripHTML(payload.Message.Abstract),
	}
	meta.BibtexKey = bibtexKey(meta)
	meta.BibTeX = buildBibTeX(meta)
	return meta, nil
}

func assertCiteRequestURL(rawURL, expectedHost string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid request URL: %w", err)
	}
	host := strings.ToLower(u.Hostname())
	if host != expectedHost {
		return fmt.Errorf("refusing request to unexpected host %q", host)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("refusing request with scheme %q", u.Scheme)
	}
	return nil
}

func (c *citeHTTPClient) fetchArxiv(ctx context.Context, arxivID string) (*paperMetadata, error) {
	arxivID = normalizeArxivID(arxivID)
	if arxivID == "" || !isValidArxivID(arxivID) {
		return nil, fmt.Errorf("invalid arXiv ID format")
	}

	reqURL := c.arxivURL + "?id_list=" + url.QueryEscape(arxivID)
	if err := c.assertArxivURL(reqURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arXiv request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("arXiv API error: HTTP %d", resp.StatusCode)
	}

	var feed arxivFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parse arXiv response: %w", err)
	}
	if len(feed.Entries) == 0 {
		return nil, fmt.Errorf("arXiv ID not found")
	}

	entry := feed.Entries[0]
	title := strings.TrimSpace(strings.ReplaceAll(entry.Title, "\n", " "))
	if title == "" {
		return nil, fmt.Errorf("arXiv ID not found")
	}

	authors := make([]string, 0, len(entry.Authors))
	for _, a := range entry.Authors {
		if name := strings.TrimSpace(a.Name); name != "" {
			authors = append(authors, name)
		}
	}

	year := 0
	if entry.Published != "" {
		if t, err := time.Parse(time.RFC3339, entry.Published); err == nil {
			year = t.Year()
		}
	}

	meta := &paperMetadata{
		Title:    title,
		Authors:  authors,
		Year:     year,
		Venue:    "arXiv",
		ArxivID:  arxivID,
		DOI:      strings.TrimSpace(entry.DOI),
		Abstract: strings.TrimSpace(entry.Summary),
	}
	meta.BibtexKey = bibtexKey(meta)
	meta.BibTeX = buildBibTeX(meta)
	return meta, nil
}

type arxivFeed struct {
	Entries []arxivEntry `xml:"entry"`
}

type arxivEntry struct {
	Title     string `xml:"title"`
	Summary   string `xml:"summary"`
	Published string `xml:"published"`
	DOI       string `xml:"http://arxiv.org/schemas/atom doi"`
	Authors   []struct {
		Name string `xml:"name"`
	} `xml:"author"`
}

func firstNonEmpty(values []string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func yearFromDateParts(parts [][]int) int {
	for _, dp := range parts {
		if len(dp) > 0 && dp[0] > 0 {
			return dp[0]
		}
	}
	return 0
}

func stripHTML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	re := regexp.MustCompile(`<[^>]+>`)
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}

func slugWord(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func bibtexKey(meta *paperMetadata) string {
	authorPart := "unknown"
	if len(meta.Authors) > 0 {
		family := meta.Authors[0]
		if idx := strings.Index(family, ","); idx >= 0 {
			family = family[:idx]
		} else if parts := strings.Fields(family); len(parts) > 0 {
			family = parts[len(parts)-1]
		}
		authorPart = slugWord(family)
		if authorPart == "" {
			authorPart = "unknown"
		}
	}
	yearPart := "0000"
	if meta.Year > 0 {
		yearPart = strconv.Itoa(meta.Year)
	}
	titlePart := slugWord(meta.Title)
	if titlePart == "" {
		titlePart = "paper"
	}
	if len(titlePart) > 24 {
		titlePart = titlePart[:24]
		titlePart = strings.Trim(titlePart, "-")
	}
	return authorPart + yearPart + titlePart
}

func escapeBibTeX(s string) string {
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	return s
}

func buildBibTeX(meta *paperMetadata) string {
	entryType := "article"
	if meta.ArxivID != "" && meta.Venue == "arXiv" {
		entryType = "misc"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "@%s{%s,\n", entryType, meta.BibtexKey)
	fmt.Fprintf(&b, "  title = {%s},\n", escapeBibTeX(meta.Title))
	if len(meta.Authors) > 0 {
		fmt.Fprintf(&b, "  author = {%s},\n", escapeBibTeX(strings.Join(meta.Authors, " and ")))
	}
	if meta.Year > 0 {
		fmt.Fprintf(&b, "  year = {%d},\n", meta.Year)
	}
	if meta.Venue != "" {
		fmt.Fprintf(&b, "  journal = {%s},\n", escapeBibTeX(meta.Venue))
	}
	if meta.DOI != "" {
		fmt.Fprintf(&b, "  doi = {%s},\n", escapeBibTeX(meta.DOI))
	}
	if meta.ArxivID != "" {
		fmt.Fprintf(&b, "  eprint = {%s},\n", escapeBibTeX(meta.ArxivID))
		fmt.Fprintf(&b, "  archivePrefix = {arXiv},\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func buildPaperMarkdown(meta *paperMetadata) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %q\n", meta.Title)
	b.WriteString("authors:\n")
	for _, a := range meta.Authors {
		fmt.Fprintf(&b, "  - %q\n", a)
	}
	if meta.Year > 0 {
		fmt.Fprintf(&b, "year: %d\n", meta.Year)
	}
	if meta.Venue != "" {
		fmt.Fprintf(&b, "venue: %q\n", meta.Venue)
	}
	if meta.DOI != "" {
		fmt.Fprintf(&b, "doi: %q\n", meta.DOI)
	}
	if meta.ArxivID != "" {
		fmt.Fprintf(&b, "arxiv: %q\n", meta.ArxivID)
	}
	b.WriteString("tags: [literature]\n")
	b.WriteString("status: to-read\n")
	if strings.TrimSpace(meta.Abstract) != "" {
		b.WriteString("abstract: |\n")
		for _, line := range strings.Split(strings.TrimSpace(meta.Abstract), "\n") {
			fmt.Fprintf(&b, "  %s\n", strings.TrimSpace(line))
		}
	}
	b.WriteString("bibtex: |\n")
	for _, line := range strings.Split(strings.TrimRight(meta.BibTeX, "\n"), "\n") {
		fmt.Fprintf(&b, "  %s\n", line)
	}
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# %s\n\n", meta.Title)
	if strings.TrimSpace(meta.Abstract) != "" {
		b.WriteString("## Abstract\n\n")
		b.WriteString(strings.TrimSpace(meta.Abstract))
		if !strings.HasSuffix(meta.Abstract, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func handleCite(b Backend, client *citeHTTPClient) server.ToolHandlerFunc {
	if client == nil {
		client = newDefaultCiteHTTPClient()
	}
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		query := strings.TrimSpace(stringArg(args, "identifier"))
		if query == "" {
			query = strings.TrimSpace(stringArg(args, "doi"))
		}
		if query == "" {
			query = strings.TrimSpace(stringArg(args, "arxiv_id"))
		}
		if query == "" {
			return citeErrorResult("", "identifier, doi, or arxiv_id is required"), nil
		}
		if _, err := sanitizeCiteInput(query); err != nil {
			return citeErrorResult(query, err.Error()), nil
		}

		actor := stringArg(args, "actor")
		if actor == "" {
			actor = "mcp-agent"
		}

		var (
			meta *paperMetadata
			err  error
		)
		arxivID := normalizeArxivID(stringArg(args, "arxiv_id"))
		doi := normalizeDOI(stringArg(args, "doi"))
		rawArxiv := strings.TrimSpace(stringArg(args, "arxiv_id"))
		rawDOI := strings.TrimSpace(stringArg(args, "doi"))

		switch {
		case rawArxiv != "" && arxivID == "":
			return citeErrorResult(query, "invalid arXiv ID format"), nil
		case rawDOI != "" && doi == "":
			return citeErrorResult(query, "invalid DOI format"), nil
		case arxivID != "":
			meta, err = client.fetchArxiv(ctx, arxivID)
		case doi != "":
			meta, err = client.fetchDOI(ctx, doi)
		case isArxivIdentifier(query):
			meta, err = client.fetchArxiv(ctx, query)
		default:
			meta, err = client.fetchDOI(ctx, query)
		}
		if err != nil {
			return citeErrorResult(query, err.Error()), nil
		}
		if err := validateBibtexKey(meta.BibtexKey); err != nil {
			return citeErrorResult(query, fmt.Sprintf("generated bibtex key: %v", err)), nil
		}

		path := "papers/" + meta.BibtexKey + ".md"
		content := buildPaperMarkdown(meta)
		if _, err := b.WriteFile(ctx, path, content, actor, ""); err != nil {
			return citeErrorResult(query, fmt.Sprintf("write paper: %v", err)), nil
		}

		payload, _ := json.Marshal(map[string]any{
			"success": true,
			"path":    path,
			"query":   query,
		})
		return mcp.NewToolResultText(string(payload)), nil
	}
}

func stringArg(args map[string]any, key string) string {
	v, _ := args[key].(string)
	return v
}
