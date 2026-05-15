package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/kiwifs/kiwifs/internal/workflow"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

const workflowRequestBodyLimit = 1 << 20

// ListWorkflows returns all workflow definitions from .kiwi/workflows/*.json
func (h *Handlers) ListWorkflows(c echo.Context) error {
	workflows, err := workflow.Load(h.root)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if workflows == nil {
		workflows = []workflow.Workflow{}
	}
	return c.JSON(http.StatusOK, map[string]any{"workflows": workflows})
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
	if !workflowHasState(w, req.State) {
		return echo.NewHTTPError(http.StatusBadRequest, "workflow state not found: "+req.State)
	}

	content, err := h.store.Read(c.Request().Context(), req.Path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "page not found: "+err.Error())
	}

	updated, err := setFrontmatterFields(content, map[string]string{
		"workflow": req.Workflow,
		"state":    req.State,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to update workflow: "+err.Error())
	}

	result, err := h.pipe.Write(c.Request().Context(), req.Path, updated, actor)
	if err != nil {
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

// AdvanceWorkflow moves a page from one workflow state to another.
//
// Request body: { "path": "...", "target_state": "...", "actor": "..." }
//
// The handler:
//  1. Reads the page's frontmatter to get current workflow+state
//  2. Loads the workflow definition
//  3. Validates the transition
//  4. Updates frontmatter state and writes the page
func (h *Handlers) AdvanceWorkflow(c echo.Context) error {
	limitedBody := http.MaxBytesReader(c.Response(), c.Request().Body, workflowRequestBodyLimit)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var req struct {
		Path        string `json:"path"`
		TargetState string `json:"target_state"`
		Actor       string `json:"actor"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
	}
	if req.Path == "" || req.TargetState == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path and target_state are required")
	}
	actor := req.Actor
	if actor == "" {
		actor = c.Request().Header.Get("X-Actor")
	}
	if actor == "" {
		actor = "system"
	}

	// Read current page
	content, err := h.store.Read(c.Request().Context(), req.Path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "page not found: "+err.Error())
	}

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

	// Load workflow definition
	w, err := workflow.Get(h.root, wfName)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found: "+err.Error())
	}

	// Validate transition
	if err := workflow.ValidateTransition(w, currentState, req.TargetState); err != nil {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}

	// Update frontmatter state
	updated, err := setFrontmatterField(content, "state", req.TargetState)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update state: "+err.Error())
	}

	// Write back
	result, err := h.pipe.Write(c.Request().Context(), req.Path, updated, actor)
	if err != nil {
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

	// Walk all files and group by state
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
		if title, ok := fm["title"].(string); ok {
			entry["title"] = title
		}
		if priority, ok := fm["priority"]; ok {
			entry["priority"] = priority
		}
		board[pageState] = append(board[pageState], entry)
		return nil
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"workflow": w,
		"board":    board,
	})
}

// setFrontmatterField updates a single key in the YAML frontmatter of a
// markdown document and returns the reconstructed document. It preserves
// the body and other frontmatter fields.
func setFrontmatterField(content []byte, key, value string) ([]byte, error) {
	return setFrontmatterFields(content, map[string]string{key: value})
}

func setFrontmatterFields(content []byte, fields map[string]string) ([]byte, error) {
	fmRaw, body, err := markdown.SplitFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("split frontmatter: %w", err)
	}

	fm := map[string]any{}
	if len(fmRaw) > 0 {
		if err := yaml.Unmarshal(fmRaw, &fm); err != nil {
			return nil, fmt.Errorf("parse frontmatter: %w", err)
		}
	} else {
		body = content
	}
	for key, value := range fields {
		fm[key] = value
	}

	newFM, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(newFM)
	buf.WriteString("---\n")
	buf.Write(body)
	return buf.Bytes(), nil
}
