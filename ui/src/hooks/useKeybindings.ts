import { useCallback, useEffect, useMemo, useState } from "react";
import { api, onSpaceChange } from "../lib/api";
import {
  buildShortcutDisplaySections,
  DEFAULT_KEYBINDINGS,
  isMacPlatform,
  mergeKeybindings,
  type KeybindingAction,
  type KeybindingsConfig,
} from "../lib/kiwiKeybindings";

export function useKeybindings() {
  const [config, setConfig] = useState<KeybindingsConfig | null>(null);

  const load = useCallback(async () => {
    try {
      const data = await api.getKeybindings();
      setConfig(data);
    } catch {
      setConfig(null);
    }
  }, []);

  useEffect(() => {
    void load();
    return onSpaceChange(() => {
      void load();
    });
  }, [load]);

  const bindings = useMemo(() => mergeKeybindings(config), [config]);
  const conflicts = config?.conflicts ?? [];
  const defaults = config?.defaults ?? DEFAULT_KEYBINDINGS;
  const mac = isMacPlatform();
  const sections = useMemo(
    () => buildShortcutDisplaySections(bindings, defaults, mac),
    [bindings, defaults, mac],
  );

  return { bindings, conflicts, defaults, sections };
}

export type { KeybindingAction };
