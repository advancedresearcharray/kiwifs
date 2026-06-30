import { ImagePlus } from "lucide-react";
import { cn } from "@kw/lib/cn";
import { isOsImageDrag } from "@kw/lib/editorImagePaste";

type Props = {
  active: boolean;
  className?: string;
};

export function EditorImageDropOverlay({ active, className }: Props) {
  if (!active) return null;
  return (
    <div
      className={cn(
        "pointer-events-none absolute inset-0 z-20 flex items-center justify-center rounded-md border-2 border-dashed border-primary/60 bg-primary/5",
        className,
      )}
      aria-hidden
    >
      <div className="flex flex-col items-center gap-2 text-primary">
        <ImagePlus className="h-8 w-8" />
        <span className="text-sm font-medium">Drop image to upload</span>
      </div>
    </div>
  );
}

export function shouldShowEditorImageDropOverlay(event: DragEvent): boolean {
  return isOsImageDrag(event);
}
