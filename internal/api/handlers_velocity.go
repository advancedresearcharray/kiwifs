package api

import (
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) Velocity(c echo.Context) error {
	ctx := c.Request().Context()

	period := c.QueryParam("period")
	if period == "" {
		period = "30d"
	}
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	pathPrefix := c.QueryParam("path_prefix")

	sinceArg := "--since=" + velocityParsePeriod(period)
	cmd := exec.CommandContext(ctx, "git", "log", "--numstat", "--format=%H|%an|%at", sinceArg)
	cmd.Dir = h.root
	out, err := cmd.Output()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "git log failed")
	}

	type fileChange struct {
		adds, dels int
		authors    map[string]bool
		timestamps []time.Time
	}
	files := make(map[string]*fileChange)
	var currentAuthor string
	var currentTime time.Time

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "|") && !strings.Contains(line, "\t") {
			parts := strings.SplitN(line, "|", 3)
			if len(parts) >= 3 {
				currentAuthor = parts[1]
				ts, _ := strconv.ParseInt(parts[2], 10, 64)
				currentTime = time.Unix(ts, 0)
			}
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		adds, _ := strconv.Atoi(parts[0])
		dels, _ := strconv.Atoi(parts[1])
		path := parts[2]
		if !strings.HasSuffix(path, ".md") {
			continue
		}
		if pathPrefix != "" && !strings.HasPrefix(path, pathPrefix) {
			continue
		}
		fc, ok := files[path]
		if !ok {
			fc = &fileChange{authors: make(map[string]bool)}
			files[path] = fc
		}
		fc.adds += adds
		fc.dels += dels
		fc.authors[currentAuthor] = true
		fc.timestamps = append(fc.timestamps, currentTime)
	}

	type scored struct {
		path    string
		changes int
		authors int
		lines   int
	}
	var items []scored
	totalChanges := 0
	for path, fc := range files {
		changes := len(fc.timestamps)
		totalChanges += changes
		items = append(items, scored{path: path, changes: changes, authors: len(fc.authors), lines: fc.adds + fc.dels})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].changes > items[j].changes })

	topN := limit
	if topN > len(items) {
		topN = len(items)
	}
	type hotSpot struct {
		Path         string `json:"path"`
		Changes      int    `json:"changes"`
		Authors      int    `json:"authors"`
		LinesChanged int    `json:"lines_changed"`
	}
	hotSpots := make([]hotSpot, topN)
	for i := 0; i < topN; i++ {
		hotSpots[i] = hotSpot{Path: items[i].path, Changes: items[i].changes, Authors: items[i].authors, LinesChanged: items[i].lines}
	}

	type coldSpot struct {
		Path            string `json:"path"`
		DaysSinceChange int    `json:"days_since_change"`
	}
	var coldSpots []coldSpot
	periodDaysVal := velocityParsePeriodDays(period)
	_ = storage.Walk(ctx, h.store, "/", func(e storage.Entry) error {
		if !strings.HasSuffix(e.Path, ".md") {
			return nil
		}
		if pathPrefix != "" && !strings.HasPrefix(e.Path, pathPrefix) {
			return nil
		}
		if _, ok := files[e.Path]; !ok {
			days := periodDaysVal
			gitCmd := exec.CommandContext(ctx, "git", "log", "-1", "--format=%at", "--", e.Path)
			gitCmd.Dir = h.root
			if gitOut, gerr := gitCmd.Output(); gerr == nil {
				if ts, perr := strconv.ParseInt(strings.TrimSpace(string(gitOut)), 10, 64); perr == nil && ts > 0 {
					days = int(time.Since(time.Unix(ts, 0)).Hours() / 24)
				}
			}
			coldSpots = append(coldSpots, coldSpot{Path: e.Path, DaysSinceChange: days})
		}
		return nil
	})

	type burst struct {
		Path       string  `json:"path"`
		RecentRate float64 `json:"recent_rate"`
		AvgRate    float64 `json:"avg_rate"`
	}
	var bursts []burst
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	periodDays := velocityParsePeriodDays(period)
	for _, item := range items {
		fc := files[item.path]
		recentCount := 0
		for _, ts := range fc.timestamps {
			if ts.After(sevenDaysAgo) {
				recentCount++
			}
		}
		recentRate := float64(recentCount) / 7.0
		avgRate := float64(item.changes) / float64(periodDays)
		if avgRate > 0 && recentRate > 3*avgRate {
			bursts = append(bursts, burst{Path: item.path, RecentRate: recentRate, AvgRate: avgRate})
		}
	}

	var singleAuthor []string
	for path, fc := range files {
		if len(fc.authors) == 1 {
			singleAuthor = append(singleAuthor, path)
		}
	}

	if hotSpots == nil {
		hotSpots = []hotSpot{}
	}
	if coldSpots == nil {
		coldSpots = []coldSpot{}
	}
	if bursts == nil {
		bursts = []burst{}
	}
	if singleAuthor == nil {
		singleAuthor = []string{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"period":              period,
		"total_changes":       totalChanges,
		"hot_spots":           hotSpots,
		"cold_spots":          coldSpots,
		"bursts":              bursts,
		"single_author_pages": singleAuthor,
	})
}

func velocityParsePeriod(period string) string {
	period = strings.TrimSpace(period)
	if strings.HasSuffix(period, "d") {
		return period[:len(period)-1] + " days ago"
	}
	if strings.HasSuffix(period, "w") {
		return period[:len(period)-1] + " weeks ago"
	}
	if strings.HasSuffix(period, "m") {
		return period[:len(period)-1] + " months ago"
	}
	return "30 days ago"
}

func velocityParsePeriodDays(period string) int {
	period = strings.TrimSpace(period)
	if strings.HasSuffix(period, "d") {
		n, _ := strconv.Atoi(period[:len(period)-1])
		if n > 0 {
			return n
		}
	}
	if strings.HasSuffix(period, "w") {
		n, _ := strconv.Atoi(period[:len(period)-1])
		if n > 0 {
			return n * 7
		}
	}
	if strings.HasSuffix(period, "m") {
		n, _ := strconv.Atoi(period[:len(period)-1])
		if n > 0 {
			return n * 30
		}
	}
	return 30
}
