import { useEffect, useState } from "react";
import { api } from "../lib/api";
import type { EditorSlashCommandConfig } from "../lib/editorSlashCommands";

export function useEditorSlashCommands() {
  const [commands, setCommands] = useState<EditorSlashCommandConfig[]>([]);

  useEffect(() => {
    let cancelled = false;
    api
      .getEditorSlashCommands()
      .then((res) => {
        if (!cancelled) setCommands(res.commands ?? []);
      })
      .catch(() => {
        if (!cancelled) setCommands([]);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return commands;
}
