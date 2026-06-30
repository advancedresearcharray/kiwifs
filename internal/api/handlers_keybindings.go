package api

import (
	"net/http"

	"github.com/kiwifs/kiwifs/internal/keybindings"
	"github.com/labstack/echo/v4"
)

// GetKeybindings godoc
//
//	@Summary		Get keyboard shortcut bindings
//	@Description	Returns merged keybindings from defaults, .kiwi/keybindings.json, and [ui.keybindings] in config.toml. Includes conflict warnings when multiple actions share a chord.
//	@Tags			ui
//	@Security		BearerAuth
//	@Success		200		{object}	keybindings.Resolved
//	@Failure		500		{object}	map[string]string
//	@Router			/api/kiwi/keybindings [get]
func (h *Handlers) GetKeybindings(c echo.Context) error {
	res, err := keybindings.Resolve(keybindings.Options{
		Root:              h.root,
		KeybindingsFile:   h.ui.KeybindingsFile,
		ConfigKeybindings: h.ui.Keybindings,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, res)
}
