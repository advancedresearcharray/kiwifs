package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type editorSlashCommandEntry struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Template    string `json:"template"`
}

type editorSlashCommandsResponse struct {
	Commands []editorSlashCommandEntry `json:"commands"`
}

// GetEditorSlashCommands godoc
//
//	@Summary		Get configurable editor slash commands
//	@Description	Returns custom slash commands from [[ui.editor.slash_commands]] in config.toml. Template content is loaded separately via the file API.
//	@Tags			theme
//	@Security		BearerAuth
//	@Success		200	{object}	editorSlashCommandsResponse
//	@Router			/api/kiwi/editor/slash-commands [get]
func (h *Handlers) GetEditorSlashCommands(c echo.Context) error {
	out := make([]editorSlashCommandEntry, 0, len(h.ui.Editor.SlashCommands))
	for _, cmd := range h.ui.Editor.SlashCommands {
		if cmd.ID == "" || cmd.Template == "" {
			continue
		}
		label := cmd.Label
		if label == "" {
			label = cmd.ID
		}
		out = append(out, editorSlashCommandEntry{
			ID:          cmd.ID,
			Label:       label,
			Icon:        cmd.Icon,
			Description: cmd.Description,
			Template:    cmd.Template,
		})
	}
	return c.JSON(http.StatusOK, editorSlashCommandsResponse{Commands: out})
}
