package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalyticsEndpoint(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "doc1.md", "---\nstatus: active\n---\n# Doc 1\nSome content here.\n")
	mustPutFile(t, s, "doc2.md", "---\nstatus: draft\n---\n# Doc 2\nMore words.\n")

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, rec.Body.String())
	}
	if resp.TotalPages < 2 {
		t.Fatalf("expected at least 2 pages, got %d", resp.TotalPages)
	}
	if resp.TopUpdated == nil {
		t.Fatal("top_updated should not be nil")
	}
	if resp.Health.Stale.Paths == nil {
		t.Fatal("stale paths should not be nil (should be empty slice)")
	}
	if resp.Engagement.TopViewed == nil {
		t.Fatal("engagement.top_viewed should not be nil")
	}
	if resp.Engagement.FailedSearches == nil {
		t.Fatal("engagement.failed_searches should not be nil")
	}
}

func TestAnalyticsEngagementStats(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "guide.md", "# Guide\nA page about kiwi notes.\n")

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=guide.md", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET guide.md: %d %s", rec.Code, rec.Body.String())
		}
	}
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=missing-widget", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /search: %d %s", rec.Code, rec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Engagement.TotalViews != 3 {
		t.Fatalf("expected total_views=3, got %d", resp.Engagement.TotalViews)
	}
	if len(resp.Engagement.TopViewed) != 1 || resp.Engagement.TopViewed[0].Path != "guide.md" || resp.Engagement.TopViewed[0].Count != 3 {
		t.Fatalf("unexpected top_viewed: %+v", resp.Engagement.TopViewed)
	}
	if len(resp.Engagement.FailedSearches) != 1 || resp.Engagement.FailedSearches[0].Query != "missing-widget" {
		t.Fatalf("unexpected failed_searches: %+v", resp.Engagement.FailedSearches)
	}
}

func TestAnalyticsScopeFiltering(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "students/alice.md", "---\nstatus: active\n---\n# Alice\n")
	mustPutFile(t, s, "students/bob.md", "---\nstatus: active\n---\n# Bob\n")
	mustPutFile(t, s, "topics/math.md", "---\nsubject: math\n---\n# Math\n")

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics?scope=students/", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics?scope=students/: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.TotalPages != 2 {
		t.Fatalf("scoped to students/ expected 2 pages, got %d", resp.TotalPages)
	}
}

func TestAnalyticsEmptyKB(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics: %d %s", rec.Code, rec.Body.String())
	}

	var resp AnalyticsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.TotalPages != 0 {
		t.Fatalf("empty KB expected 0 pages, got %d", resp.TotalPages)
	}
	if resp.TotalWords != 0 {
		t.Fatalf("empty KB expected 0 words, got %d", resp.TotalWords)
	}
	if resp.TopUpdated == nil {
		t.Fatal("top_updated should be empty slice, not nil")
	}
	if len(resp.TopUpdated) != 0 {
		t.Fatalf("expected no top updated, got %d", len(resp.TopUpdated))
	}
}

func TestFailedSearchAnalyticsEndpoint(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "guide.md", "# Guide\nA page about searchable kiwi notes.\n")

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=missing-widget", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /search missing-widget: %d %s", rec.Code, rec.Body.String())
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/search?q=searchable", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /search searchable: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/failed-searches?top=5", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/failed-searches: %d %s", rec.Code, rec.Body.String())
	}

	var resp FailedSearchesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Top != 5 {
		t.Fatalf("expected top=5, got %d", resp.Top)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected one failed query, got %+v", resp.Results)
	}
	if got := resp.Results[0]; got.Query != "missing-widget" || got.Count != 2 || got.SearchType != "search" {
		t.Fatalf("unexpected failed search stat: %+v", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/failed-searches?since=2999-01-01T00:00:00Z", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/failed-searches since: %d %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal since: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected since filter to hide old rows, got %+v", resp.Results)
	}
}

func TestPageViewAnalyticsEndpoint(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "guide.md", "# Guide\nA page about kiwi notes.\n")
	mustPutFile(t, s, "other.md", "# Other\nAnother page.\n")

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=guide.md", nil)
		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET guide.md: %d %s", rec.Code, rec.Body.String())
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/file?path=other.md&source=ui", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET other.md: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/views?top=5", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/views: %d %s", rec.Code, rec.Body.String())
	}

	var resp PageViewsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Top != 5 {
		t.Fatalf("expected top=5, got %d", resp.Top)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected two page view rows, got %+v", resp.Results)
	}
	if got := resp.Results[0]; got.Path != "guide.md" || got.Count != 2 {
		t.Fatalf("expected guide.md first with count 2, got %+v", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/views?path=other.md", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/views path filter: %d %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal path filter: %v", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].Path != "other.md" || resp.Results[0].Count != 1 {
		t.Fatalf("unexpected path-filtered results: %+v", resp.Results)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/kiwi/analytics/views?since=2999-01-01T00:00:00Z", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /analytics/views since: %d %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal since: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected since filter to hide old rows, got %+v", resp.Results)
	}
}

func TestHealthCheckEndpoint(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	mustPutFile(t, s, "doc.md", "---\nstatus: active\n---\n# Doc\nSome words for counting.\n")

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/health-check?path=doc.md", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /health-check: %d %s", rec.Code, rec.Body.String())
	}

	var resp HealthCheckResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Path != "doc.md" {
		t.Fatalf("expected path=doc.md, got %s", resp.Path)
	}
	if resp.Issues == nil {
		t.Fatal("issues should not be nil")
	}
}

func TestHealthCheckMissingPath(t *testing.T) {
	s, _ := buildSQLiteTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/kiwi/health-check", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing path, got %d", rec.Code)
	}
}
