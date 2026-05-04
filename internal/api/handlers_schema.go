package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/labstack/echo/v4"
)

var validTypeName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func (h *Handlers) ListSchemas(c echo.Context) error {
	dir := filepath.Join(h.root, ".kiwi", "schemas")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return c.JSON(http.StatusOK, []string{})
	}
	var schemas []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			schemas = append(schemas, e.Name()[:len(e.Name())-5])
		}
	}
	if schemas == nil {
		schemas = []string{}
	}
	return c.JSON(http.StatusOK, schemas)
}

func (h *Handlers) GetSchema(c echo.Context) error {
	typeName := c.Param("type")
	if typeName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "type is required")
	}
	if !validTypeName.MatchString(typeName) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid type name")
	}

	path := filepath.Join(h.root, ".kiwi", "schemas", typeName+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "schema not found")
	}
	var raw json.RawMessage
	if json.Unmarshal(data, &raw) != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid schema")
	}
	return c.JSON(http.StatusOK, raw)
}

func (h *Handlers) PutSchema(c echo.Context) error {
	typeName := c.Param("type")
	if typeName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "type is required")
	}
	if !validTypeName.MatchString(typeName) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid type name")
	}

	var raw json.RawMessage
	if err := c.Bind(&raw); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON schema")
	}

	dir := filepath.Join(h.root, ".kiwi", "schemas")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	path := filepath.Join(dir, typeName+".json")
	if err := os.WriteFile(path, raw, 0644); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if h.schemaReload != nil {
		h.schemaReload()
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "type": typeName})
}
