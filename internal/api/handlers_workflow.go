package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/pipeline"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/kiwifs/kiwifs/internal/workflow"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

// ordinalStep is the default gap between card ordinals when assigning new
// positions within a column. Using a large step leaves room for insertions
// between existing cards without rebalancing.
const ordinalStep = 1000

const workflowRequestBodyLimit = 1 << 20

// ListWorkflows returns all workflow definitions from .kiwi/workflows/*.json.
// Broken JSON files are reported in the "errors" field so the UI can warn
// that some boards are unavailable.
func (h *Handlers) ListWorkflows(c echo.Context) error {
	result := workflow.LoadWithErrors(h.root)
	workflows := result.Workflows
	if workflows == nil {
		workflows = []workflow.Workflow{}
	}
	resp := map[string]any{"workflows": workflows}
	if len(result.Errors) > 0 {
		errs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			errs[i] = e.Error()
		}
		resp["errors"] = errs
	}
	return c.JSON(http.StatusOK, resp)
}

// GetWorkflow returns a single workflow definition
func (h *Handlers) GetWorkflow(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "workflow name required")
	}

	w, err := workflow.Get(h.root, name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, w)
}

// SaveWorkflow creates or updates a workflow definition
func (h *Handlers) SaveWorkflow(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "workflow name required")
	}

	limitedBody := http.MaxBytesReader(c.Response(), c.Request().Body, workflowRequestBodyLimit)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var w workflow.Workflow
	if err := json.Unmarshal(body, &w); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workflow JSON: "+err.Error())
	}
	w.Name = name

	if err := workflow.Save(h.root, w); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "saved", "workflow": w})
}

// DeleteWorkflow removes a workflow definition. It does not edit pages that
// reference the workflow in frontmatter.
func (h *Handlers) DeleteWorkflow(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "workflow name required")
	}

	if err := workflow.Delete(h.root, name); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "deleted", "name": name})
}

// AssignWorkflow adds an existing page to a workflow/state column, or moves it
// into a workflow column regardless of its previous workflow membership.
func (h *Handlers) AssignWorkflow(c echo.Context) error {
	limitedBody := http.MaxBytesReader(c.Response(), c.Request().Body, workflowRequestBodyLimit)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var req struct {
		Path     string `json:"path"`
		Workflow string `json:"workflow"`
		State    string `json:"state"`
		Ordinal  *int   `json:"ordinal,omitempty"` // optional position within column
		Actor    string `json:"actor"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
	}
	if req.Path == "" || req.Workflow == "" || req.State == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path, workflow, and state are required")
	}
	if !strings.HasSuffix(strings.ToLower(req.Path), ".md") {
		return echo.NewHTTPError(http.StatusBadRequest, "workflow assignment only supports markdown pages")
	}
	if err := workflow.ValidateName(req.Workflow); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	actor := req.Actor
	if actor == "" {
		actor = c.Request().Header.Get("X-Actor")
	}
	if actor == "" {
		actor = "system"
	}

	w, err := workflow.Get(h.root, req.Workflow)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found: "+err.Error())
	}
	if !workflowHasStateCaseInsensitive(w, req.State) {
		return echo.NewHTTPError(http.StatusBadRequest, "workflow state not found: "+req.State)
	}
	// Resolve to the canonical state name for consistent storage.
	req.State = resolveStateName(w, req.State)

	content, err := h.store.Read(c.Request().Context(), req.Path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "page not found: "+err.Error())
	}
	currentETag := pipeline.ETag(content)

	// No-op: skip write when workflow and state are already correct.
	fm, _ := markdown.Frontmatter(content)
	if fm != nil {
		curWF, _ := fm["workflow"].(string)
		curState, _ := fm["state"].(string)
		if curWF == req.Workflow && curState == req.State && req.Ordinal == nil {
			return c.JSON(http.StatusOK, map[string]any{
				"path":     req.Path,
				"workflow": req.Workflow,
				"state":    req.State,
				"etag":     currentETag,
				"noop":     true,
			})
		}
	}

	fields := map[string]string{
		"workflow": req.Workflow,
		"state":    req.State,
	}
	if req.Ordinal != nil {
		fields["ordinal"] = strconv.Itoa(*req.Ordinal)
	}

	updated, err := setFrontmatterFields(content, fields)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to update workflow: "+err.Error())
	}

	// Optimistic locking: reject if the page changed between read and write.
	result, err := h.pipe.WriteWithOpts(c.Request().Context(), req.Path, updated, actor, pipeline.WriteOpts{
		IfMatch: currentETag,
	})
	if err != nil {
		if err == pipeline.ErrConflict {
			return echo.NewHTTPError(http.StatusConflict, "page was modified concurrently — please retry")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":     req.Path,
		"workflow": req.Workflow,
		"state":    req.State,
		"etag":     result.ETag,
	})
}

func workflowHasState(w workflow.Workflow, stateName string) bool {
	for _, state := range w.States {
		if state.Name == stateName {
			return true
		}
	}
	return false
}

// workflowHasStateCaseInsensitive checks whether a state exists using
// case-insensitive, whitespace-normalized comparison so that frontmatter
// values like "To Do" still match a column named "to do".
func workflowHasStateCaseInsensitive(w workflow.Workflow, stateName string) bool {
	norm := normalizeStateKey(stateName)
	for _, state := range w.States {
		if normalizeStateKey(state.Name) == norm {
			return true
		}
	}
	return false
}

// resolveStateName returns the canonical state name from the workflow
// definition that matches the given name (case-insensitive). If no match
// is found the input is returned unchanged.
func resolveStateName(w workflow.Workflow, stateName string) string {
	norm := normalizeStateKey(stateName)
	for _, state := range w.States {
		if normalizeStateKey(state.Name) == norm {
			return state.Name
		}
	}
	return stateName
}

// normalizeStateKey lowercases and collapses whitespace for fuzzy matching.
func normalizeStateKey(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// pageStem returns the filename without extension, suitable as a display
// title fallback when frontmatter title is missing.
func pageStem(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}

// AdvanceWorkflow moves a page from one workflow state to another.
//
// Request body: { "path": "...", "workflow": "...", "target_state": "...", "actor": "..." }
//
// The handler:
//  1. Reads the page's frontmatter to get current workflow+state
//  2. Loads the workflow definition
//  3. Validates the transition (terminal states, required_role)
//  4. Updates frontmatter state and writes the page with optimistic locking
func (h *Handlers) AdvanceWorkflow(c echo.Context) error {
	limitedBody := http.MaxBytesReader(c.Response(), c.Request().Body, workflowRequestBodyLimit)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var req struct {
		Path        string `json:"path"`
		Workflow    string `json:"workflow"`
		TargetState string `json:"target_state"`
		Actor       string `json:"actor"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
	}
	if req.Path == "" || req.TargetState == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path and target_state are required")
	}
	if !strings.HasSuffix(strings.ToLower(req.Path), ".md") {
		return echo.NewHTTPError(http.StatusBadRequest, "workflow advance only supports markdown pages")
	}
	actor := req.Actor
	if actor == "" {
		actor = c.Request().Header.Get("X-Actor")
	}
	if actor == "" {
		actor = "system"
	}

	content, err := h.store.Read(c.Request().Context(), req.Path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "page not found: "+err.Error())
	}
	currentETag := pipeline.ETag(content)

	fm, err := markdown.Frontmatter(content)
	if err != nil || fm == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "page has no frontmatter")
	}

	wfName, _ := fm["workflow"].(string)
	currentState, _ := fm["state"].(string)
	if wfName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "page has no 'workflow' field in frontmatter")
	}
	if currentState == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "page has no 'state' field in frontmatter")
	}

	// If the client specified a workflow, verify it matches the page's frontmatter.
	if req.Workflow != "" && req.Workflow != wfName {
		return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf(
			"page belongs to workflow %q but request specified %q", wfName, req.Workflow))
	}

	// No-op: skip write if already in target state.
	if currentState == req.TargetState {
		return c.JSON(http.StatusOK, map[string]any{
			"path":       req.Path,
			"from_state": currentState,
			"to_state":   req.TargetState,
			"etag":       currentETag,
			"noop":       true,
		})
	}

	// Load workflow definition
	w, err := workflow.Get(h.root, wfName)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found: "+err.Error())
	}

	// Validate target state exists in the workflow
	validTarget := false
	var targetState workflow.State
	for _, s := range w.States {
		if s.Name == req.TargetState {
			validTarget = true
			targetState = s
			break
		}
	}
	if !validTarget {
		return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("state %q does not exist in workflow %q", req.TargetState, w.Name))
	}
	// Enforce per-column WIP (work-in-progress) limit: reject moves that
	// would push the target column over its configured capacity.
	if targetState.WIPLimit > 0 {
		count := 0
		_ = storage.WalkAll(c.Request().Context(), h.store, "/", func(e storage.Entry) error {
			if !strings.HasSuffix(e.Path, ".md") || e.Path == req.Path {
				return nil
			}
			raw, rerr := h.store.Read(c.Request().Context(), e.Path)
			if rerr != nil {
				return nil
			}
			pfm, perr := markdown.Frontmatter(raw)
			if perr != nil || pfm == nil {
				return nil
			}
			pwf, _ := pfm["workflow"].(string)
			ps, _ := pfm["state"].(string)
			if pwf == wfName && ps == req.TargetState {
				count++
			}
			return nil
		})
		if count >= targetState.WIPLimit {
			return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf(
				"column %q has reached its WIP limit of %d", req.TargetState, targetState.WIPLimit))
		}
	}

	// Terminal states cannot have outbound transitions.
	for _, s := range w.States {
		if s.Name == currentState && s.Terminal {
			return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf(
				"state %q is terminal in workflow %q — cards cannot be moved out", currentState, w.Name))
		}
	}

	if err := workflow.ValidateTransition(w, currentState, req.TargetState); err != nil {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}

	// Update frontmatter: state + auto-stamp modified time on transition.
	updated, err := setFrontmatterFields(content, map[string]string{
		"state":    req.TargetState,
		"modified": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update state: "+err.Error())
	}

	// Optimistic locking: reject if the page changed between read and write.
	result, err := h.pipe.WriteWithOpts(c.Request().Context(), req.Path, updated, actor, pipeline.WriteOpts{
		IfMatch: currentETag,
	})
	if err != nil {
		if err == pipeline.ErrConflict {
			return echo.NewHTTPError(http.StatusConflict, "page was modified concurrently — please retry")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":       req.Path,
		"from_state": currentState,
		"to_state":   req.TargetState,
		"etag":       result.ETag,
	})
}

// WorkflowBoard returns pages grouped by their workflow state (Kanban view).
// GET /api/kiwi/workflow/board/:workflow
func (h *Handlers) WorkflowBoard(c echo.Context) error {
	wfName := c.Param("workflow")
	if wfName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "workflow name required")
	}

	// Load workflow definition
	w, err := workflow.Get(h.root, wfName)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found: "+err.Error())
	}

	// Walk all files and group by state. Pages whose state doesn't match any
	// defined column are collected in an "__unmatched__" bucket so the UI can
	// surface them instead of silently dropping them.
	board := make(map[string][]map[string]any)
	for _, s := range w.States {
		board[s.Name] = []map[string]any{}
	}

	err = storage.WalkAll(c.Request().Context(), h.store, "/", func(e storage.Entry) error {
		if !strings.HasSuffix(e.Path, ".md") {
			return nil
		}
		content, readErr := h.store.Read(c.Request().Context(), e.Path)
		if readErr != nil {
			return nil
		}
		fm, fmErr := markdown.Frontmatter(content)
		if fmErr != nil || fm == nil {
			return nil
		}
		pageWF, _ := fm["workflow"].(string)
		pageState, _ := fm["state"].(string)
		if pageWF != wfName || pageState == "" {
			return nil
		}
		entry := map[string]any{
			"path":  e.Path,
			"state": pageState,
		}

		// Title with fallback to filename stem when frontmatter title is
		// missing, so cards never render blank.
		if title, ok := fm["title"].(string); ok && title != "" {
			entry["title"] = title
		} else {
			entry["title"] = pageStem(e.Path)
		}
		if priority, ok := fm["priority"]; ok {
			entry["priority"] = priority
		}

		// Tags (support both "tags" and "labels" keys, string or []any).
		if tags := extractTags(fm); len(tags) > 0 {
			entry["tags"] = tags
		}

		// Due date.
		if due, ok := fm["due"].(string); ok && due != "" {
			entry["due"] = due
		}

		// Author.
		if author, ok := fm["author"].(string); ok && author != "" {
			entry["author"] = author
		}

		// Ordinal for within-column ordering.
		if ord := fmInt(fm, "ordinal"); ord != nil {
			entry["ordinal"] = *ord
		}

		// Blocked status — a card can be flagged as blocked without moving
		// it to a different column.
		if blocked, ok := fm["blocked"].(bool); ok && blocked {
			entry["blocked"] = true
		}
		if reason, ok := fm["block_reason"].(string); ok && reason != "" {
			entry["block_reason"] = reason
		}

		// Dependencies — references to other pages this card depends on.
		if deps := extractStringList(fm, "depends_on"); len(deps) > 0 {
			entry["depends_on"] = deps
		}

		// Description: first ~120 chars of the body after frontmatter.
		body := markdown.BodyAfterFrontmatter(content)
		if desc := cardDescription(body); desc != "" {
			entry["description"] = desc
		}

		// Modified timestamp: prefer frontmatter "modified" field, fall
		// back to file system modtime.
		if fmMod, ok := fm["modified"].(string); ok && fmMod != "" {
			entry["modified"] = fmMod
		} else if !e.ModTime.IsZero() {
			entry["modified"] = e.ModTime.Format(time.RFC3339)
		}

		// Case-insensitive column matching: resolve pageState to the
		// canonical column name so cards with slightly different casing
		// still land in the correct column.
		canonState := resolveStateName(w, pageState)
		if workflowHasState(w, canonState) {
			board[canonState] = append(board[canonState], entry)
		} else {
			board["__unmatched__"] = append(board["__unmatched__"], entry)
		}
		return nil
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Sort each column's cards by ordinal (ascending). Cards without an
	// ordinal sort after all ordered cards, preserving their relative order
	// from the filesystem walk.
	for _, pages := range board {
		sort.SliceStable(pages, func(i, j int) bool {
			oi, okI := pages[i]["ordinal"].(int)
			oj, okJ := pages[j]["ordinal"].(int)
			if okI && okJ {
				return oi < oj
			}
			if okI {
				return true // ordered cards come first
			}
			return false
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"workflow": w,
		"board":    board,
	})
}

// ReorderCard updates a card's ordinal (position within its column) without
// changing its workflow state. The client sends the desired ordinal value,
// typically computed as the midpoint between the two neighbouring cards.
//
// POST /api/kiwi/workflow/reorder
// Body: { "path": "...", "ordinal": 1500, "actor": "..." }
func (h *Handlers) ReorderCard(c echo.Context) error {
	limitedBody := http.MaxBytesReader(c.Response(), c.Request().Body, workflowRequestBodyLimit)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var req struct {
		Path    string `json:"path"`
		Ordinal int    `json:"ordinal"`
		Actor   string `json:"actor"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
	}
	if req.Path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}
	actor := req.Actor
	if actor == "" {
		actor = c.Request().Header.Get("X-Actor")
	}
	if actor == "" {
		actor = "system"
	}

	content, err := h.store.Read(c.Request().Context(), req.Path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "page not found: "+err.Error())
	}
	currentETag := pipeline.ETag(content)

	// No-op: skip write if ordinal already matches.
	fm, _ := markdown.Frontmatter(content)
	if fm != nil {
		if cur := fmInt(fm, "ordinal"); cur != nil && *cur == req.Ordinal {
			return c.JSON(http.StatusOK, map[string]any{
				"path":    req.Path,
				"ordinal": req.Ordinal,
				"etag":    currentETag,
				"noop":    true,
			})
		}
	}

	updated, err := setFrontmatterField(content, "ordinal", strconv.Itoa(req.Ordinal))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update ordinal: "+err.Error())
	}

	result, err := h.pipe.WriteWithOpts(c.Request().Context(), req.Path, updated, actor, pipeline.WriteOpts{
		IfMatch: currentETag,
	})
	if err != nil {
		if err == pipeline.ErrConflict {
			return echo.NewHTTPError(http.StatusConflict, "page was modified concurrently — please retry")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"path":    req.Path,
		"ordinal": req.Ordinal,
		"etag":    result.ETag,
	})
}

// fmInt safely extracts an integer value from frontmatter. YAML may parse
// numbers as int or float64 depending on the document.
func fmInt(fm map[string]any, key string) *int {
	v, ok := fm[key]
	if !ok {
		return nil
	}
	switch n := v.(type) {
	case int:
		return &n
	case float64:
		if n == math.Trunc(n) {
			i := int(n)
			return &i
		}
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return &i
		}
	}
	return nil
}

// extractStringList pulls a []string from frontmatter, handling both
// []any{"a","b"} and a bare "a" string.
func extractStringList(fm map[string]any, key string) []string {
	val, ok := fm[key]
	if !ok {
		return nil
	}
	switch v := val.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if v != "" {
			return []string{v}
		}
	}
	return nil
}

// setFrontmatterField updates a single key in the YAML frontmatter of a
// markdown document and returns the reconstructed document. It preserves
// the body and other frontmatter fields.
func setFrontmatterField(content []byte, key, value string) ([]byte, error) {
	return setFrontmatterFields(content, map[string]string{key: value})
}

// setFrontmatterFields updates or inserts YAML keys in the frontmatter
// without round-tripping through a full parse/marshal cycle, preserving
// comments, ordering, and scalar formatting.
//
// For each key it looks for an existing "key: ..." line and replaces the
// value in place. Keys that don't already exist are appended before the
// closing "---". When the document has no frontmatter at all, a new block
// is prepended.
func setFrontmatterFields(content []byte, fields map[string]string) ([]byte, error) {
	fmRaw, body, err := markdown.SplitFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("split frontmatter: %w", err)
	}

	// No existing frontmatter — fall back to the simple marshal path
	// (there are no comments to preserve).
	if len(fmRaw) == 0 {
		fm := map[string]any{}
		for key, value := range fields {
			fm[key] = value
		}
		newFM, merr := yaml.Marshal(fm)
		if merr != nil {
			return nil, fmt.Errorf("marshal frontmatter: %w", merr)
		}
		var buf bytes.Buffer
		buf.WriteString("---\n")
		buf.Write(newFM)
		buf.WriteString("---\n")
		buf.Write(content) // entire original content is the body
		return buf.Bytes(), nil
	}

	// Line-level in-place editing: replace existing keys, track which
	// fields were handled so we can append the rest.
	lines := strings.Split(string(fmRaw), "\n")
	handled := make(map[string]bool, len(fields))

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		for key, value := range fields {
			prefix := key + ":"
			if strings.HasPrefix(trimmed, prefix) {
				rest := trimmed[len(prefix):]
				// Make sure it's actually "key:" or "key: value", not
				// "key_extra: value".
				if rest == "" || rest[0] == ' ' || rest[0] == '\t' {
					lines[i] = key + ": " + yamlScalar(value)
					handled[key] = true
					break
				}
			}
		}
	}

	// Append any fields that weren't already present.
	for key, value := range fields {
		if !handled[key] {
			lines = append(lines, key+": "+yamlScalar(value))
		}
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString(strings.Join(lines, "\n"))
	// Ensure trailing newline before closing fence.
	if len(lines) > 0 && lines[len(lines)-1] != "" {
		buf.WriteByte('\n')
	}
	buf.WriteString("---\n")
	buf.Write(body)
	return buf.Bytes(), nil
}

// yamlScalar returns a YAML-safe representation of a simple string value.
// Values that are safe bare scalars are returned unquoted; everything else
// is double-quoted.
func yamlScalar(s string) string {
	// Quick check: if the value is a plain, safe scalar, don't quote.
	safe := true
	for _, c := range s {
		if c == ':' || c == '#' || c == '"' || c == '\'' || c == '\n' || c == '{' || c == '}' || c == '[' || c == ']' {
			safe = false
			break
		}
	}
	if safe && s == strings.TrimSpace(s) && s != "" && s != "true" && s != "false" && s != "null" && s != "yes" && s != "no" {
		return s
	}
	// Fall back to double-quoted form.
	b, _ := yaml.Marshal(s)
	return strings.TrimSpace(string(b))
}

// extractTags returns tags from the "tags" or "labels" frontmatter key.
// Accepts both []any{"a","b"} and a bare string "a".
func extractTags(fm map[string]any) []string {
	val, ok := fm["tags"]
	if !ok {
		val, ok = fm["labels"]
	}
	if !ok {
		return nil
	}
	switch v := val.(type) {
	case []any:
		tags := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				tags = append(tags, s)
			}
		}
		return tags
	case string:
		if v != "" {
			return []string{v}
		}
	}
	return nil
}

// reHeading matches markdown headings like "# Title" at start of a line.
var reHeading = regexp.MustCompile(`(?m)^#{1,6}\s+`)

// cardDescription returns the first ~120 characters of the markdown body,
// stripping the leading heading line (which duplicates the title) and
// trimming whitespace. Returns "" if there is no meaningful content.
func cardDescription(body string) string {
	s := strings.TrimSpace(body)
	if s == "" {
		return ""
	}
	// Strip leading heading line (e.g. "# Title\n").
	if strings.HasPrefix(s, "#") {
		if idx := strings.IndexByte(s, '\n'); idx >= 0 {
			s = strings.TrimSpace(s[idx+1:])
		} else {
			return "" // only a heading, no body
		}
	}
	// Strip remaining markdown heading markers for a cleaner preview.
	s = reHeading.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Truncate at ~120 chars on a rune boundary.
	const maxLen = 120
	if utf8.RuneCountInString(s) > maxLen {
		runes := []rune(s)
		s = string(runes[:maxLen]) + "…"
	}
	return s
}
