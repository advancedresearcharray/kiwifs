import { useEffect, useMemo, useState } from "react";
import { api } from "../lib/api";
import {
  DEFAULT_KEYBINDINGS,
  mergeKeybindings,
  type KeybindingAction,
  type KeybindingsConfig,
} from "../lib/kiwiKeybindings";

export function useKeybindings() {
  const [config, setConfig] = useState<KeybindingsConfig | null>(null);

  useEffect(() => {
    api.getKeybindings().then(setConfig).catch(() => setConfig(null));
  }, []);

  const bindings = useMemo(() => mergeKeybindings(config), [config]);
  const conflicts = config?.conflicts ?? [];
  const defaults = useMemo(
    () => mergeKeybindings({ bindings: config?.defaults ?? {}, defaults: DEFAULT_KEYBINDINGS, conflicts: [] }),
    [config?.defaults],
  );

  return { bindings, conflicts, defaults };
}

export type { KeybindingAction };
