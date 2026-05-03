package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) Context(c echo.Context) error {
	read := func(rel string) string {
		data, err := os.ReadFile(filepath.Join(h.root, rel))
		if err != nil {
			return ""
		}
		return string(data)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"playbook": read(filepath.Join(".kiwi", "playbook.md")),
		"schema":   read("SCHEMA.md"),
		"index":    read("index.md"),
	})
}
