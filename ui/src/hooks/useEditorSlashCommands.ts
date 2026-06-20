import { useEffect, useState } from "react";
import { api, onSpaceChange } from "../lib/api";
import type { EditorSlashCommandConfig } from "../lib/editorSlashCommands";

export function useEditorSlashCommands(): EditorSlashCommandConfig[] {
  const [commands, setCommands] = useState<EditorSlashCommandConfig[]>([]);

  useEffect(() => {
    let cancelled = false;

    const load = () => {
      api
        .getEditorSlashCommands()
        .then((res) => {
          if (!cancelled) setCommands(res.commands ?? []);
        })
        .catch(() => {
          if (!cancelled) setCommands([]);
        });
    };

    load();
    return onSpaceChange(load);
  }, []);

  return commands;
}
