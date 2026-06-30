package api

import (
	"net/http"

	"github.com/kiwifs/kiwifs/internal/themepresets"
	"github.com/labstack/echo/v4"
)

// GetThemePresets godoc
//
//	@Summary		List workspace theme presets
//	@Description	Loads theme preset JSON files from the directory configured via [ui.theme] presets_dir (default .kiwi/themes/). Invalid files are reported in errors without failing the request.
//	@Tags			theme
//	@Security		BearerAuth
//	@Success		200		{object}	themepresets.Result
//	@Failure		500		{object}	map[string]string
//	@Router			/api/kiwi/theme/presets [get]
func (h *Handlers) GetThemePresets(c echo.Context) error {
	result := themepresets.LoadFromDir(h.root, h.ui.Theme)
	return c.JSON(http.StatusOK, result)
}
