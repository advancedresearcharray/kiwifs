package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
)

// FeedAtom generates an Atom feed of recent changes.
// GET /api/kiwi/feed.xml
func (h *Handlers) FeedAtom(c echo.Context) error {
	feed, err := h.buildFeed(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	atom, err := feed.ToAtom()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.XMLBlob(http.StatusOK, []byte(atom))
}

// FeedJSON generates a JSON feed of recent changes.
// GET /api/kiwi/feed.json
func (h *Handlers) FeedJSON(c echo.Context) error {
	feed, err := h.buildFeed(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	jsonFeed, err := feed.ToJSON()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("Content-Type", "application/feed+json; charset=utf-8")
	return c.String(http.StatusOK, jsonFeed)
}

// buildFeed creates a feeds.Feed from recent timeline events.
func (h *Handlers) buildFeed(ctx context.Context) (*feeds.Feed, error) {
	// Fetch recent timeline events (last 50)
	events, err := h.fetchTimelineEvents(ctx, 50, "", "", "")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	feed := &feeds.Feed{
		Title:       "KiwiFS Activity Feed",
		Link:        &feeds.Link{Href: h.publicURL},
		Description: "Recent changes to the knowledge base",
		Created:     now,
	}

	// Convert timeline events to feed items
	for _, event := range events {
		timestamp, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			timestamp = now
		}

		// Build item title
		title := event.Path
		if event.Type == "delete" {
			title = "Deleted: " + event.Path
		} else {
			title = "Updated: " + event.Path
		}

		// Build permalink for the file
		link := h.publicURL
		if link != "" && event.Type != "delete" {
			if link[len(link)-1] != '/' {
				link += "/"
			}
			link += event.Path
		}

		item := &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: link},
			Description: event.Message,
			Author:      &feeds.Author{Name: event.Actor},
			Created:     timestamp,
			Id:          event.Path + "@" + event.Timestamp,
		}

		feed.Items = append(feed.Items, item)
	}

	return feed, nil
}
