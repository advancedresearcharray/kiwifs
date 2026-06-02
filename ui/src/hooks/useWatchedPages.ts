import { useCallback, useEffect, useState } from "react";

type WatchEntry = {
  path: string;
  channel?: string;
};

const BASE_KEY = "kiwifs-watched-pages";

function storageKey(space: string): string {
  return `${BASE_KEY}:${space}`;
}

function loadLocal(space: string): WatchEntry[] {
  try {
    const raw = localStorage.getItem(storageKey(space));
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed) && typeof parsed[0] === "string") {
      return parsed.map((p: string) => ({ path: p }));
    }
    return parsed as WatchEntry[];
  } catch {
    return [];
  }
}

function saveLocal(space: string, watches: WatchEntry[]) {
  try {
    localStorage.setItem(storageKey(space), JSON.stringify(watches));
  } catch {}
}

function isCloudMode(): boolean {
  return typeof window !== "undefined" && !!(window as any).__kiwi_cloud_mode__;
}

async function fetchWatches(): Promise<WatchEntry[]> {
  try {
    const res = await fetch("/api/kiwi/watches");
    if (!res.ok) return [];
    const data = await res.json();
    return (data as any[]).map((w: any) => ({ path: w.path, channel: w.channel }));
  } catch {
    return [];
  }
}

async function createWatchApi(path: string, channel?: string): Promise<boolean> {
  try {
    const res = await fetch("/api/kiwi/watches", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path, channel }),
    });
    return res.ok || res.status === 409;
  } catch {
    return false;
  }
}

async function deleteWatchApi(path: string): Promise<boolean> {
  try {
    const res = await fetch("/api/kiwi/watches", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path }),
    });
    return res.ok;
  } catch {
    return false;
  }
}

export function useWatchedPages(space: string = "default") {
  const [watched, setWatched] = useState<WatchEntry[]>(() => loadLocal(space));
  const cloud = isCloudMode();

  useEffect(() => {
    if (cloud) {
      fetchWatches().then((w) => {
        setWatched(w);
        saveLocal(space, w);
      });
    } else {
      setWatched(loadLocal(space));
    }
  }, [space, cloud]);

  const addWatch = useCallback(
    async (path: string, channel?: string) => {
      if (cloud) {
        const ok = await createWatchApi(path, channel);
        if (ok) {
          const updated = await fetchWatches();
          setWatched(updated);
          saveLocal(space, updated);
        }
      } else {
        setWatched((prev) => {
          const next = [...prev, { path, channel }];
          saveLocal(space, next);
          return next;
        });
      }
    },
    [space, cloud]
  );

  const removeWatch = useCallback(
    async (path: string) => {
      if (cloud) {
        await deleteWatchApi(path);
        const updated = await fetchWatches();
        setWatched(updated);
        saveLocal(space, updated);
      } else {
        setWatched((prev) => {
          const next = prev.filter((w) => w.path !== path);
          saveLocal(space, next);
          return next;
        });
      }
    },
    [space, cloud]
  );

  const toggle = useCallback(
    (path: string) => {
      if (watched.some((w) => w.path === path)) {
        removeWatch(path);
      } else {
        addWatch(path);
      }
    },
    [watched, addWatch, removeWatch]
  );

  const isWatched = useCallback(
    (path: string) => watched.some((w) => w.path === path),
    [watched]
  );

  return { watched, toggle, isWatched, addWatch, removeWatch, isCloud: cloud };
}
