import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ErrorBoundary } from "./ErrorBoundary";
import { Check, Circle, Loader2, Save, User, X, XCircle } from "lucide-react";
import matter from "gray-matter";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { titleize } from "../lib/paths";
import { KiwiBreadcrumb } from "./KiwiBreadcrumb";
import { ExcalidrawMarkdownEditor, isExcalidrawMarkdown } from "./ExcalidrawMarkdownPreview";
import { MarkdownSourceEditor } from "./editor/MarkdownSourceEditor";
import { formatDistanceToNow } from "date-fns";

type SaveStatus = "clean" | "dirty" | "saving" | "saved" | "error";

type SaveHandle = { save: () => Promise<void> };

type Props = {
  path: string;
  tree?: import("../lib/api").TreeEntry | null;
  onClose: () => void;
  onSaved: (path: string) => void;
  onNavigate?: (path: string) => void;
  saveRef?: React.MutableRefObject<SaveHandle | null>;
};

export function KiwiEditor({ path, onClose, onSaved, onNavigate, saveRef }: Props) {
  const [initialMd, setInitialMd] = useState<string | null>(null);
  const etagRef = useRef<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isDark, setIsDark] = useState<boolean>(() =>
    typeof document !== "undefined" &&
    document.documentElement.classList.contains("dark")
  );

  useEffect(() => {
    const obs = new MutationObserver(() =>
      setIsDark(document.documentElement.classList.contains("dark"))
    );
    obs.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });
    return () => obs.disconnect();
  }, []);

  useEffect(() => {
    let cancelled = false;
    api
      .readFile(path)
      .then((r) => {
        if (cancelled) return;
        etagRef.current = r.etag;
        setInitialMd(r.content || "");
      })
      .catch((e) => {
        if (!cancelled) setError(String(e));
      });
    return () => {
      cancelled = true;
    };
  }, [path]);

  if (error) {
    return (
      <div className="p-8 text-sm text-destructive font-mono">{error}</div>
    );
  }
  if (initialMd === null) {
    return (
      <div className="p-8 text-sm text-muted-foreground">Loading editor…</div>
    );
  }

  if (isExcalidrawMarkdown(initialMd)) {
    return (
      <ExcalidrawEditorInner
        path={path}
        initialMd={initialMd}
        etagRef={etagRef}
        saving={saving}
        setSaving={setSaving}
        setError={setError}
        onClose={onClose}
        onSaved={onSaved}
        onNavigate={onNavigate}
        saveRef={saveRef}
      />
    );
  }

  return (
    <MarkdownEditorInner
      path={path}
      initialMd={initialMd}
      etagRef={etagRef}
      isDark={isDark}
      saving={saving}
      setSaving={setSaving}
      setError={setError}
      onClose={onClose}
      onSaved={onSaved}
      onNavigate={onNavigate}
      saveRef={saveRef}
    />
  );
}

function ExcalidrawEditorInner({
  path,
  initialMd,
  etagRef,
  saving,
  setSaving,
  setError,
  onClose,
  onSaved,
  onNavigate,
  saveRef,
}: {
  path: string;
  initialMd: string;
  etagRef: React.MutableRefObject<string | null>;
  saving: boolean;
  setSaving: (v: boolean) => void;
  setError: (v: string | null) => void;
  onClose: () => void;
  onSaved: (p: string) => void;
  onNavigate?: (path: string) => void;
  saveRef?: React.MutableRefObject<SaveHandle | null>;
}) {
  const [currentMd, setCurrentMd] = useState(initialMd);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>("clean");
  const savedFlashTimer = useRef<number | null>(null);

  const fmTitle = useMemo(() => frontmatterTitle(initialMd), [initialMd]);

  const onSaveRef = useRef<(opts?: { close?: boolean }) => Promise<void>>(async () => {});
  onSaveRef.current = async (opts) => {
    setSaving(true);
    setSaveStatus("saving");
    setError(null);
    try {
      const res = await api.writeFile(path, currentMd, etagRef.current || undefined);
      etagRef.current = res.etag ? `"${res.etag}"` : null;
      setSaveStatus("saved");
      if (savedFlashTimer.current) window.clearTimeout(savedFlashTimer.current);
      savedFlashTimer.current = window.setTimeout(() => setSaveStatus("clean"), 2000);
      if (opts?.close) onSaved(path);
    } catch (e) {
      setSaveStatus("error");
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const markChanged = useCallback((nextMd: string) => {
    setCurrentMd(nextMd);
    setSaveStatus(nextMd === initialMd ? "clean" : "dirty");
  }, [initialMd]);

  useEffect(() => {
    if (!saveRef) return;
    saveRef.current = { save: () => onSaveRef.current({ close: true }) };
    return () => { saveRef.current = null; };
  }, [saveRef]);

  useEffect(() => {
    return () => {
      if (savedFlashTimer.current) window.clearTimeout(savedFlashTimer.current);
    };
  }, []);

  return (
    <EditorShell
      path={path}
      title={fmTitle || titleize(path)}
      saveStatus={saveStatus}
      saving={saving}
      canSave={saveStatus !== "clean"}
      onSave={() => onSaveRef.current({ close: true })}
      onClose={onClose}
      onNavigate={onNavigate}
    >
      <ExcalidrawMarkdownEditor markdown={initialMd} onChange={markChanged} />
    </EditorShell>
  );
}

function MarkdownEditorInner({
  path,
  initialMd,
  etagRef,
  isDark,
  saving,
  setSaving,
  setError,
  onClose,
  onSaved,
  onNavigate,
  saveRef,
}: {
  path: string;
  initialMd: string;
  etagRef: React.MutableRefObject<string | null>;
  isDark: boolean;
  saving: boolean;
  setSaving: (v: boolean) => void;
  setError: (v: string | null) => void;
  onClose: () => void;
  onSaved: (p: string) => void;
  onNavigate?: (path: string) => void;
  saveRef?: React.MutableRefObject<SaveHandle | null>;
}) {
  const [markdown, setMarkdown] = useState(initialMd);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>("clean");
  const [lastEdit, setLastEdit] = useState<{ author: string; date: string } | null>(null);
  const autoSaveTimer = useRef<number | null>(null);
  const savedFlashTimer = useRef<number | null>(null);
  const fmTitle = useMemo(() => frontmatterTitle(initialMd), [initialMd]);

  useEffect(() => {
    let cancelled = false;
    api.versions(path).then((r) => {
      if (cancelled || !r.versions.length) return;
      const v = r.versions[0];
      setLastEdit({ author: v.author, date: v.date });
    }).catch(() => {});
    return () => { cancelled = true; };
  }, [path]);

  const onSaveRef = useRef<(opts?: { close?: boolean }) => Promise<void>>(async () => {});
  onSaveRef.current = async (opts) => {
    setSaving(true);
    setSaveStatus("saving");
    setError(null);
    try {
      const res = await api.writeFile(path, markdown, etagRef.current || undefined);
      etagRef.current = res.etag ? `"${res.etag}"` : null;
      setSaveStatus("saved");
      setLastEdit({ author: "you", date: new Date().toISOString() });
      if (savedFlashTimer.current) window.clearTimeout(savedFlashTimer.current);
      savedFlashTimer.current = window.setTimeout(() => setSaveStatus("clean"), 2000);
      if (opts?.close) onSaved(path);
    } catch (e) {
      setSaveStatus("error");
      setError(String(e));
    } finally {
      setSaving(false);
    }
  };

  const markDirty = useCallback((nextMd: string) => {
    setMarkdown(nextMd);
    setSaveStatus(nextMd === initialMd ? "clean" : "dirty");
    if (autoSaveTimer.current) window.clearTimeout(autoSaveTimer.current);
    autoSaveTimer.current = window.setTimeout(() => {
      onSaveRef.current();
    }, 2000);
  }, [initialMd]);

  useEffect(() => {
    return () => {
      if (autoSaveTimer.current) window.clearTimeout(autoSaveTimer.current);
      if (savedFlashTimer.current) window.clearTimeout(savedFlashTimer.current);
    };
  }, []);

  useEffect(() => {
    if (!saveRef) return;
    saveRef.current = { save: () => onSaveRef.current({ close: true }) };
    return () => { saveRef.current = null; };
  }, [saveRef]);

  return (
    <EditorShell
      path={path}
      title={fmTitle || titleize(path)}
      saveStatus={saveStatus}
      saving={saving}
      canSave={saveStatus !== "clean"}
      onSave={() => onSaveRef.current({ close: true })}
      onClose={onClose}
      onNavigate={onNavigate}
      metadata={lastEdit && (
        <span className="flex items-center gap-1">
          <User className="h-3 w-3" />
          Last edited by {lastEdit.author} {relativeTime(lastEdit.date)}
        </span>
      )}
    >
      <div className="max-w-5xl">
        <ErrorBoundary>
          <MarkdownSourceEditor
            value={markdown}
            onChange={markDirty}
            dark={isDark}
            minHeight="60vh"
            onSaveShortcut={() => onSaveRef.current()}
          />
        </ErrorBoundary>
      </div>
    </EditorShell>
  );
}

function EditorShell({
  path,
  title,
  saveStatus,
  saving,
  canSave,
  onSave,
  onClose,
  onNavigate,
  metadata,
  children,
}: {
  path: string;
  title: string;
  saveStatus: SaveStatus;
  saving: boolean;
  canSave: boolean;
  onSave: () => void;
  onClose: () => void;
  onNavigate?: (path: string) => void;
  metadata?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col h-full">
      <div className="sticky top-0 z-10 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80 border-b border-border shrink-0">
        <div className="px-8 py-2 max-w-6xl mx-auto">
          {onNavigate
            ? <KiwiBreadcrumb path={path} onNavigate={onNavigate} />
            : <div className="text-sm text-muted-foreground font-mono truncate">{path}</div>}
        </div>
      </div>

      <div className="flex-1 overflow-auto kiwi-scroll">
        <div className="max-w-6xl mx-auto px-8 py-6">
          <div className="mb-6">
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                <h1 className="text-2xl font-bold tracking-tight text-foreground leading-tight">
                  {title}
                </h1>
                <div className="flex items-center gap-2 mt-2">
                  <SaveIndicator status={saveStatus} />
                </div>
              </div>
              <div className="flex items-center gap-2 shrink-0 pt-1">
                <Button
                  onClick={onSave}
                  disabled={saving || !canSave}
                  size="sm"
                  variant={saveStatus === "dirty" ? "default" : "outline"}
                >
                  <Save className="h-3.5 w-3.5" />
                  {saving ? "Saving…" : "Save & Close"}
                </Button>
                <Button variant="outline" size="sm" onClick={onClose}>
                  <X className="h-3.5 w-3.5" /> Close
                </Button>
              </div>
            </div>

            {metadata && (
              <div className="flex items-center gap-3 mt-3 text-xs text-muted-foreground">
                {metadata}
              </div>
            )}
          </div>

          {children}
        </div>
      </div>
    </div>
  );
}

function SaveIndicator({ status }: { status: SaveStatus }) {
  switch (status) {
    case "dirty":
      return (
        <span className="flex items-center gap-1 text-xs text-amber-500">
          <Circle className="h-2.5 w-2.5 fill-current" />
          Unsaved
        </span>
      );
    case "saving":
      return (
        <span className="flex items-center gap-1 text-xs text-muted-foreground">
          <Loader2 className="h-3 w-3 animate-spin" />
          Saving…
        </span>
      );
    case "saved":
      return (
        <span className="flex items-center gap-1 text-xs text-green-500">
          <Check className="h-3 w-3" />
          Saved
        </span>
      );
    case "error":
      return (
        <span className="flex items-center gap-1 text-xs text-destructive">
          <XCircle className="h-3 w-3" />
          Error
        </span>
      );
    default:
      return null;
  }
}

function relativeTime(d: string): string {
  try {
    const parsed = new Date(d);
    if (isNaN(parsed.getTime())) return d;
    return formatDistanceToNow(parsed, { addSuffix: true });
  } catch {
    return d;
  }
}

function frontmatterTitle(markdown: string): string | null {
  try {
    const parsed = matter(markdown);
    if (typeof parsed.data?.title === "string") return parsed.data.title;
  } catch {}
  return null;
}
