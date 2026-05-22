package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const specCanvasJSON = `{
  "nodes": [
    {"id": "n1", "type": "text", "x": 10, "y": 20, "width": 200, "height": 80, "text": "Hello"}
  ],
  "edges": [
    {"id": "e1", "fromNode": "n1", "toNode": "n2", "fromSide": "right", "toSide": "left", "label": "next"}
  ]
}`

func TestCanvas_WriteReadRoundtripPreservesSpec(t *testing.T) {
	s := buildTestServer(t)
	path := "boards/flow.canvas.json"

	req := httptest.NewRequest(http.MethodPut, "/api/kiwi/canvas?path="+path, strings.NewReader(specCanvasJSON))
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT canvas: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/canvas?path="+path, nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET canvas: %d %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"fromNode"`) || !strings.Contains(body, `"toNode"`) {
		t.Fatalf("roundtrip lost JSON Canvas edge fields: %s", body)
	}
	if strings.Contains(body, `"source"`) || strings.Contains(body, `"target"`) {
		t.Fatalf("roundtrip introduced non-spec edge fields: %s", body)
	}
}

func TestCanvas_ListMultiple(t *testing.T) {
	s := buildTestServer(t)
	mustPutCanvas(t, s, "a.canvas.json", specCanvasJSON)
	mustPutCanvas(t, s, "nested/b.canvas.json", specCanvasJSON)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/canvases", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET canvases: %d %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Canvases []canvasListEntry `json:"canvases"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Canvases) != 2 {
		t.Fatalf("canvases len = %d, want 2", len(resp.Canvases))
	}
	names := map[string]string{}
	for _, c := range resp.Canvases {
		names[c.Path] = c.Name
	}
	if names["a.canvas.json"] != "a" {
		t.Fatalf("display name = %q, want a", names["a.canvas.json"])
	}
}

func TestCanvas_WriteRejectsInvalidJSON(t *testing.T) {
	s := buildTestServer(t)
	req := httptest.NewRequest(http.MethodPut, "/api/kiwi/canvas?path=x.canvas.json", strings.NewReader(`{"nodes":[]}`))
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT invalid canvas: got %d, want 400", rec.Code)
	}
}

func TestCanvas_Delete(t *testing.T) {
	s := buildTestServer(t)
	mustPutCanvas(t, s, "del.canvas.json", specCanvasJSON)

	req := httptest.NewRequest(http.MethodDelete, "/api/kiwi/canvas?path=del.canvas.json", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE canvas: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/canvas?path=del.canvas.json", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET after delete: got %d, want 404", rec.Code)
	}
}

func TestCanvas_PatchAddAndRemoveNodes(t *testing.T) {
	s := buildTestServer(t)
	mustPutCanvas(t, s, "patch.canvas.json", `{"nodes":[],"edges":[]}`)

	patchBody := `{
		"operations": [
			{"op":"add_node","node":{"id":"a","type":"text","x":0,"y":0,"width":100,"height":50,"text":"A"}},
			{"op":"add_node","node":{"id":"b","type":"text","x":200,"y":0,"width":100,"height":50,"text":"B"}},
			{"op":"add_edge","edge":{"id":"e1","fromNode":"a","toNode":"b","label":"link"}}
		]
	}`
	req := httptest.NewRequest(http.MethodPatch, "/api/kiwi/canvas?path=patch.canvas.json", strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH add: %d %s", rec.Code, rec.Body.String())
	}

	// Verify
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/canvas?path=patch.canvas.json", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `"a"`) || !strings.Contains(body, `"b"`) || !strings.Contains(body, `"e1"`) {
		t.Fatalf("patch didn't stick: %s", body)
	}

	// Remove node a (should cascade edge e1)
	patchBody2 := `{"operations":[{"op":"remove_node","id":"a"}]}`
	req = httptest.NewRequest(http.MethodPatch, "/api/kiwi/canvas?path=patch.canvas.json", strings.NewReader(patchBody2))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH remove: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/canvas?path=patch.canvas.json", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	body = rec.Body.String()
	if strings.Contains(body, `"a"`) {
		t.Fatalf("node a still present: %s", body)
	}
	if strings.Contains(body, `"e1"`) {
		t.Fatalf("edge e1 should have been cascaded: %s", body)
	}
}

func TestCanvas_PatchCreatesCanvasIfMissing(t *testing.T) {
	s := buildTestServer(t)
	patchBody := `{"operations":[{"op":"add_node","node":{"id":"a","type":"text","x":0,"y":0,"width":100,"height":50,"text":"first"}}]}`
	req := httptest.NewRequest(http.MethodPatch, "/api/kiwi/canvas?path=new.canvas.json", strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH new canvas: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCanvas_Query(t *testing.T) {
	s := buildTestServer(t)
	canvas := `{
		"nodes":[
			{"id":"a","type":"text","x":0,"y":0,"width":100,"height":50,"text":"Auth flow"},
			{"id":"b","type":"file","x":200,"y":0,"width":100,"height":50,"file":"auth.md"},
			{"id":"c","type":"text","x":0,"y":200,"width":100,"height":50,"text":"DB schema"}
		],
		"edges":[
			{"id":"e1","fromNode":"a","toNode":"b","label":"implements"},
			{"id":"e2","fromNode":"a","toNode":"c"}
		]
	}`
	mustPutCanvas(t, s, "q.canvas.json", canvas)

	// Filter by type
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/canvas/query?path=q.canvas.json&type=file", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("query type: %d %s", rec.Code, rec.Body.String())
	}
	var res struct {
		Nodes []json.RawMessage `json:"nodes"`
	}
	json.Unmarshal(rec.Body.Bytes(), &res)
	if len(res.Nodes) != 1 {
		t.Fatalf("type filter: nodes = %d, want 1", len(res.Nodes))
	}

	// Search text
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/canvas/query?path=q.canvas.json&q=auth", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	json.Unmarshal(rec.Body.Bytes(), &res)
	if len(res.Nodes) < 2 {
		t.Fatalf("search: nodes = %d, want >= 2", len(res.Nodes))
	}

	// Connected
	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/canvas/query?path=q.canvas.json&connected=a", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	var connRes struct {
		Nodes []json.RawMessage `json:"nodes"`
		Edges []json.RawMessage `json:"edges"`
	}
	json.Unmarshal(rec.Body.Bytes(), &connRes)
	if len(connRes.Edges) != 2 {
		t.Fatalf("connected edges = %d, want 2", len(connRes.Edges))
	}
}

func TestCanvas_AutoLayout(t *testing.T) {
	s := buildTestServer(t)
	canvas := `{
		"nodes":[{"id":"a","type":"text","x":0,"y":0,"width":100,"height":50,"text":"A"},{"id":"b","type":"text","x":0,"y":0,"width":100,"height":50,"text":"B"}],
		"edges":[{"id":"e1","fromNode":"a","toNode":"b"}]
	}`
	mustPutCanvas(t, s, "layout.canvas.json", canvas)

	req := httptest.NewRequest(http.MethodPost, "/api/kiwi/canvas/auto-layout?path=layout.canvas.json", strings.NewReader(`{"layout":"dot"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("auto-layout: %d %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["node_count"] != float64(2) {
		t.Fatalf("node_count = %v, want 2", resp["node_count"])
	}
}

func TestCanvas_GenerateFromLinkGraph(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)
	mustPutFile(t, s, "alpha.md", "links to [[beta]]")
	mustPutFile(t, s, "beta.md", "back to [[alpha]]")

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/graph", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("warm graph: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/kiwi/canvas/generate", strings.NewReader(`{"path":"generated.canvas.json","layout":"dot"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST canvas/generate: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/canvas?path=generated.canvas.json", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET generated: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"fromNode"`) {
		t.Fatalf("generated canvas missing fromNode: %s", rec.Body.String())
	}
}

func mustPutCanvas(t *testing.T, s *Server, path, body string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/kiwi/canvas?path="+path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT canvas %s: %d %s", path, rec.Code, rec.Body.String())
	}
}
