import { FileText } from "lucide-react";
import { format, parseISO } from "date-fns";
import { Badge } from "@kw/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@kw/components/ui/card";
import { ScrollArea } from "@kw/components/ui/scroll-area";
import { pageDotColor, type CalendarPageEntry } from "@kw/lib/calendarView";

type Props = {
  dateKey: string;
  entries: CalendarPageEntry[];
  onNavigate: (path: string) => void;
};

export function CalendarDayPanel({ dateKey, entries, onNavigate }: Props) {
  const label = format(parseISO(dateKey), "EEEE, MMMM d, yyyy");

  return (
    <Card className="h-full border-border/80">
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-semibold">{label}</CardTitle>
        <p className="text-xs text-muted-foreground">
          {entries.length} page{entries.length === 1 ? "" : "s"}
        </p>
      </CardHeader>
      <CardContent className="pt-0">
        {entries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No pages on this day.</p>
        ) : (
          <ScrollArea className="h-[min(24rem,calc(100vh-14rem))] pr-2">
            <ul className="space-y-2">
              {entries.map((entry) => (
                <li key={entry.path}>
                  <button
                    type="button"
                    onClick={() => onNavigate(entry.path)}
                    className="w-full rounded-md border border-border/70 bg-card px-3 py-2 text-left transition-colors hover:bg-accent/50"
                  >
                    <div className="flex items-start gap-2">
                      <span
                        className="mt-1 inline-block h-2 w-2 shrink-0 rounded-full"
                        style={{ backgroundColor: pageDotColor(entry) }}
                      />
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-1.5 text-sm font-medium">
                          <FileText className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                          <span className="truncate">{entry.title}</span>
                        </div>
                        <div className="mt-0.5 truncate text-xs text-muted-foreground">
                          {entry.path}
                        </div>
                        {(entry.state || entry.tags.length > 0) && (
                          <div className="mt-1.5 flex flex-wrap gap-1">
                            {entry.state && (
                              <Badge variant="outline" className="text-[10px]">
                                {entry.state}
                              </Badge>
                            )}
                            {entry.tags.slice(0, 3).map((tag) => (
                              <Badge key={tag} variant="secondary" className="text-[10px]">
                                {tag}
                              </Badge>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  );
}
