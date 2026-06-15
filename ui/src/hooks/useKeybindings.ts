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

  return { bindings, conflicts, defaults: config?.defaults ?? DEFAULT_KEYBINDINGS };
}

export type { KeybindingAction };
