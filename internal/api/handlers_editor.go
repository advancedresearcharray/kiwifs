package api

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

// slashCommandIDPattern matches IDs usable in both BlockNote and CodeMirror / menus.
var slashCommandIDPattern = regexp.MustCompile(`^[\w-]+$`)

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
//	@Tags			editor
//	@Security		BearerAuth
//	@Success		200	{object}	editorSlashCommandsResponse
//	@Router			/api/kiwi/editor/slash-commands [get]
func (h *Handlers) GetEditorSlashCommands(c echo.Context) error {
	out := make([]editorSlashCommandEntry, 0, len(h.ui.Editor.SlashCommands))
	for _, cmd := range h.ui.Editor.SlashCommands {
		id := strings.TrimSpace(cmd.ID)
		template := strings.TrimSpace(cmd.Template)
		if id == "" || template == "" || !slashCommandIDPattern.MatchString(id) {
			continue
		}
		label := strings.TrimSpace(cmd.Label)
		if label == "" {
			label = id
		}
		icon := strings.TrimSpace(cmd.Icon)
		if icon == "" {
			icon = "FileText"
		}
		out = append(out, editorSlashCommandEntry{
			ID:          id,
			Label:       label,
			Icon:        icon,
			Description: strings.TrimSpace(cmd.Description),
			Template:    template,
		})
	}
	return c.JSON(http.StatusOK, editorSlashCommandsResponse{Commands: out})
}
