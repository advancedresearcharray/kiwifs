// Create / edit view dialog for the Bases component.

import { useEffect, useState } from "react";
import { Button } from "@kw/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@kw/components/ui/dialog";
import { Input } from "@kw/components/ui/input";
import { Label } from "@kw/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";

export type ViewDefinition = {
  name: string;
  query: string;
  layout: "table" | "cards" | "list" | "map";
  columns: { key: string; label: string; summary?: string }[];
  filters: unknown[];
  sort: unknown[];
  group_by?: string;
};

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (view: ViewDefinition) => void;
  initial?: ViewDefinition | null;
};

export function BasesViewDialog({ open, onOpenChange, onSave, initial }: Props) {
  const [name, setName] = useState("");
  const [query, setQuery] = useState("*");
  const [viewLayout, setViewLayout] = useState<ViewDefinition["layout"]>("table");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    if (initial) {
      setName(initial.name);
      setQuery(initial.query);
      setViewLayout(initial.layout);
    } else {
      setName("");
      setQuery("*");
      setViewLayout("table");
    }
    setError(null);
  }, [open, initial]);

  function handleSave() {
    const n = name.trim();
    if (!n) {
      setError("Name is required.");
      return;
    }
    onSave({
      name: n,
      query: query || "*",
      layout: viewLayout,
      columns: initial?.columns ?? [],
      filters: initial?.filters ?? [],
      sort: initial?.sort ?? [],
      group_by: initial?.group_by,
    });
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{initial ? "Edit view" : "New view"}</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="view-name">Name</Label>
            <Input
              id="view-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. All drafts"
              autoFocus
              onKeyDown={(e) => {
                if (e.key === "Enter") handleSave();
              }}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="view-query">Query</Label>
            <Input
              id="view-query"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="DQL or * for all"
              className="font-mono"
            />
          </div>
          <div className="grid gap-1.5">
            <Label>Default layout</Label>
            <Select
              value={viewLayout}
              onValueChange={(v) => setViewLayout(v as ViewDefinition["layout"])}
            >
              <SelectTrigger className="h-9">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="table">Table</SelectItem>
                <SelectItem value="cards">Cards</SelectItem>
                <SelectItem value="list">List</SelectItem>
                <SelectItem value="map">Map</SelectItem>
              </SelectContent>
            </Select>
          </div>
          {error && (
            <div className="text-sm text-destructive">{error}</div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
