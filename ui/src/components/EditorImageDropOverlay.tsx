export function EditorImageDropOverlay({ active }: { active: boolean }) {
  if (!active) return null;
  return (
    <div
      className="absolute inset-0 z-20 flex items-center justify-center rounded-md border-2 border-dashed border-primary/50 bg-primary/5 pointer-events-none"
      data-testid="editor-image-drop-overlay"
    >
      <span className="text-sm text-muted-foreground">Drop image to upload</span>
    </div>
  );
}
