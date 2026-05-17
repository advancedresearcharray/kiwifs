import { Loader2, Plus } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@kw/components/ui/dialog";
import { Input } from "@kw/components/ui/input";
import { Label } from "@kw/components/ui/label";
import { Textarea } from "@kw/components/ui/textarea";
import { type AddCardDialogState, type AddCardMode, useKanbanStore } from "./kanbanStore";
import type { SearchResult } from "@kw/lib/api";

type Props = {
  onNavigate: (path: string) => void;
};

type ButtonVariant = "default" | "outline";

function getModeButtonVariant(currentMode: AddCardMode, targetMode: AddCardMode): ButtonVariant {
  if (currentMode === targetMode) {
    return "default";
  }
  return "outline";
}

function CreateCardButtonIcon({ busy }: { busy: boolean }) {
  if (busy) {
    return <Loader2 className="h-3.5 w-3.5 animate-spin" />;
  }
  return <Plus className="h-3.5 w-3.5" />;
}

function NewCardFields({
  addCard,
  setAddCardTitle,
  setAddCardPath,
  setAddCardBody,
}: {
  addCard: AddCardDialogState;
  setAddCardTitle: (title: string) => void;
  setAddCardPath: (path: string) => void;
  setAddCardBody: (body: string) => void;
}) {
  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <Label htmlFor="kanban-card-title">Title</Label>
        <Input
          id="kanban-card-title"
          value={addCard.title}
          onChange={(event) => setAddCardTitle(event.target.value)}
          placeholder="e.g. Draft launch note"
          disabled={addCard.busy}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="kanban-card-path">Path</Label>
        <Input
          id="kanban-card-path"
          value={addCard.path}
          onChange={(event) => setAddCardPath(event.target.value)}
          placeholder="tasks/draft-launch-note.md"
          disabled={addCard.busy}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="kanban-card-body">Body</Label>
        <Textarea
          id="kanban-card-body"
          value={addCard.body}
          onChange={(event) => setAddCardBody(event.target.value)}
          placeholder="Optional notes..."
          disabled={addCard.busy}
          rows={4}
        />
      </div>
    </div>
  );
}

function SearchResultsContent({
  results,
  busy,
  assignExistingPage,
}: {
  results: SearchResult[];
  busy: boolean;
  assignExistingPage: (path: string) => Promise<void>;
}) {
  if (results.length === 0) {
    return <div className="p-3 text-sm text-muted-foreground">No search results yet.</div>;
  }
  return (
    <>
      {results.map((result) => (
        <div key={result.path} className="flex min-w-0 items-start gap-2 p-2">
          <div className="min-w-0 flex-1 overflow-hidden">
            <div className="truncate text-sm font-medium" title={result.path}>{result.path}</div>
            {result.snippet && <div className="line-clamp-2 break-words text-xs text-muted-foreground">{result.snippet}</div>}
          </div>
          <Button className="shrink-0" size="sm" onClick={() => void assignExistingPage(result.path)} disabled={busy}>
            Add
          </Button>
        </div>
      ))}
    </>
  );
}

function ExistingPageFields({
  addCard,
  setAddCardQuery,
  searchExistingPages,
  assignExistingPage,
}: {
  addCard: AddCardDialogState;
  setAddCardQuery: (query: string) => void;
  searchExistingPages: () => Promise<void>;
  assignExistingPage: (path: string) => Promise<void>;
}) {
  return (
    <div className="min-w-0 space-y-3">
      <div className="flex min-w-0 gap-2">
        <Input
          className="min-w-0 flex-1"
          value={addCard.query}
          onChange={(event) => setAddCardQuery(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter" && !addCard.busy) void searchExistingPages();
          }}
          placeholder="Search markdown pages"
          disabled={addCard.busy}
        />
        <Button type="button" variant="outline" className="shrink-0" onClick={() => void searchExistingPages()} disabled={addCard.busy}>
          Search
        </Button>
      </div>
      <div className="max-h-56 w-full min-w-0 overflow-auto rounded-md border border-border divide-y divide-border/50">
        <SearchResultsContent
          results={addCard.results}
          busy={addCard.busy}
          assignExistingPage={assignExistingPage}
        />
      </div>
    </div>
  );
}

function AddCardModeFields({
  addCard,
  setAddCardTitle,
  setAddCardPath,
  setAddCardBody,
  setAddCardQuery,
  searchExistingPages,
  assignExistingPage,
}: {
  addCard: AddCardDialogState;
  setAddCardTitle: (title: string) => void;
  setAddCardPath: (path: string) => void;
  setAddCardBody: (body: string) => void;
  setAddCardQuery: (query: string) => void;
  searchExistingPages: () => Promise<void>;
  assignExistingPage: (path: string) => Promise<void>;
}) {
  if (addCard.mode === "new") {
    return (
      <NewCardFields
        addCard={addCard}
        setAddCardTitle={setAddCardTitle}
        setAddCardPath={setAddCardPath}
        setAddCardBody={setAddCardBody}
      />
    );
  }
  return (
    <ExistingPageFields
      addCard={addCard}
      setAddCardQuery={setAddCardQuery}
      searchExistingPages={searchExistingPages}
      assignExistingPage={assignExistingPage}
    />
  );
}

export function KanbanAddCardDialog({ onNavigate }: Props) {
  const addCard = useKanbanStore((state) => state.addCard);
  const setAddCardOpen = useKanbanStore((state) => state.setAddCardOpen);
  const setAddCardMode = useKanbanStore((state) => state.setAddCardMode);
  const setAddCardTitle = useKanbanStore((state) => state.setAddCardTitle);
  const setAddCardPath = useKanbanStore((state) => state.setAddCardPath);
  const setAddCardBody = useKanbanStore((state) => state.setAddCardBody);
  const setAddCardQuery = useKanbanStore((state) => state.setAddCardQuery);
  const searchExistingPages = useKanbanStore((state) => state.searchExistingPages);
  const assignExistingPage = useKanbanStore((state) => state.assignExistingPage);
  const cancelAddCard = useKanbanStore((state) => state.cancelAddCard);
  const createCard = useKanbanStore((state) => state.createCard);

  const handleCreateCard = async () => {
    const path = await createCard();
    if (path) onNavigate(path);
  };

  return (
    <Dialog open={addCard.open} onOpenChange={setAddCardOpen}>
      <DialogContent className="w-[calc(100vw-2rem)] overflow-hidden sm:max-w-lg">
        <DialogHeader className="min-w-0">
          <DialogTitle>Add card to {addCard.state}</DialogTitle>
          <DialogDescription>
            Create a new markdown page or attach an existing page by setting its workflow/state frontmatter.
          </DialogDescription>
        </DialogHeader>

        <div className="min-w-0 space-y-4 py-2">
          <div className="flex min-w-0 gap-2">
            <Button
              type="button"
              size="sm"
              variant={getModeButtonVariant(addCard.mode, "new")}
              onClick={() => setAddCardMode("new")}
              disabled={addCard.busy}
            >
              New card
            </Button>
            <Button
              type="button"
              size="sm"
              variant={getModeButtonVariant(addCard.mode, "existing")}
              onClick={() => setAddCardMode("existing")}
              disabled={addCard.busy}
            >
              Add existing page
            </Button>
          </div>

          <AddCardModeFields
            addCard={addCard}
            setAddCardTitle={setAddCardTitle}
            setAddCardPath={setAddCardPath}
            setAddCardBody={setAddCardBody}
            setAddCardQuery={setAddCardQuery}
            searchExistingPages={searchExistingPages}
            assignExistingPage={assignExistingPage}
          />

          {addCard.error && <p className="text-sm text-destructive">{addCard.error}</p>}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={cancelAddCard} disabled={addCard.busy}>
            Cancel
          </Button>
          {addCard.mode === "new" && (
            <Button onClick={() => void handleCreateCard()} disabled={addCard.busy}>
              <CreateCardButtonIcon busy={addCard.busy} />
              Create card
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
