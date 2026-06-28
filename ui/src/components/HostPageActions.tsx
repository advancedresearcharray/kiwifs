import { useEffect, useState } from "react";
import * as LucideIcons from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { cn } from "@kw/lib/cn";
import { Button } from "./ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "./ui/tooltip";
import {
  dispatchPageAction,
  getPageActions,
  KIWI_PAGE_ACTION_STATE_EVENT,
  type KiwiPageActionState,
  type PageActionStateEvent,
} from "../lib/hostConfig";

function resolveLucideIcon(name: string): LucideIcon {
  const icons = LucideIcons as unknown as Record<string, LucideIcon | undefined>;
  return icons[name] ?? LucideIcons.CircleHelp;
}

interface HostPageActionButtonProps {
  action: { id: string; icon: string; activeIcon?: string; label: string; activeLabel?: string; disabled?: boolean };
  state?: KiwiPageActionState;
  path: string;
}

function HostPageActionButton({ action, state, path }: HostPageActionButtonProps) {
  const isActive = state?.active;
  const Icon = resolveLucideIcon(isActive && action.activeIcon ? action.activeIcon : action.icon);
  const label = isActive && action.activeLabel ? action.activeLabel : action.label;
  const disabled = action.disabled || state?.disabled;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className={cn("h-8 w-8", isActive && "text-emerald-500 hover:text-emerald-600")}
          aria-label={label}
          disabled={disabled}
          onClick={() => dispatchPageAction(action.id, path)}
        >
          <Icon className={cn("h-4 w-4", isActive && "fill-emerald-500/20")} />
        </Button>
      </TooltipTrigger>
      <TooltipContent side="bottom">{label}</TooltipContent>
    </Tooltip>
  );
}

/** Renders host-configured page-level action icons (same chrome as Pin/Star). */
export function HostPageActions({ path }: { path: string }) {
  const actions = getPageActions();
  const [states, setStates] = useState<Record<string, KiwiPageActionState>>({});

  useEffect(() => {
    const onState = (event: Event) => {
      const detail = (event as PageActionStateEvent).detail;
      if (detail && typeof detail === "object") {
        setStates(detail);
      }
    };
    window.addEventListener(KIWI_PAGE_ACTION_STATE_EVENT, onState);
    return () => window.removeEventListener(KIWI_PAGE_ACTION_STATE_EVENT, onState);
  }, []);

  if (actions.length === 0) return null;

  return (
    <>
      {actions.map((action) => (
        <HostPageActionButton
          key={action.id}
          action={action}
          state={states[action.id]}
          path={path}
        />
      ))}
    </>
  );
}
