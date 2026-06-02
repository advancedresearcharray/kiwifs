import { useCallback, useEffect, useState } from "react";

const BASE_KEY = "kiwifs-watched-pages";

function storageKey(space: string): string {
  return `${BASE_KEY}:${space}`;
}

function load(space: string): string[] {
  try {
    const raw = localStorage.getItem(storageKey(space));
    if (!raw) return [];
    return JSON.parse(raw) as string[];
  } catch {
    return [];
  }
}

function save(space: string, pages: string[]) {
  try {
    localStorage.setItem(storageKey(space), JSON.stringify(pages));
  } catch {}
}

export function useWatchedPages(space: string = "default") {
  const [watched, setWatched] = useState<string[]>(() => load(space));

  useEffect(() => {
    setWatched(load(space));
  }, [space]);

  const toggle = useCallback(
    (path: string) => {
      setWatched((prev) => {
        const next = prev.includes(path)
          ? prev.filter((p) => p !== path)
          : [...prev, path];
        save(space, next);
        return next;
      });
    },
    [space]
  );

  const isWatched = useCallback(
    (path: string) => watched.includes(path),
    [watched]
  );

  return { watched, toggle, isWatched };
}
