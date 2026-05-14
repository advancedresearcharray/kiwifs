import { type Dispatch, type RefObject, type SetStateAction, useEffect } from "react";
import {
  nextExpandedForReveal,
  shouldFocusRevealTarget,
  type TreeRevealRequest,
} from "@kw/lib/treeReveal";

export function useTreeRevealExpansion(
  revealRequest: TreeRevealRequest | null | undefined,
  setExpanded: Dispatch<SetStateAction<Set<string>>>,
): void {
  useEffect(() => {
    setExpanded((prev) => nextExpandedForReveal(prev, revealRequest?.path));
  }, [revealRequest, setExpanded]);
}

export function useTreeRevealTargetFocus<T extends HTMLElement>(
  revealRequest: TreeRevealRequest | null | undefined,
  path: string,
  nodeRef: RefObject<T | null>,
): void {
  useEffect(() => {
    if (!shouldFocusRevealTarget(revealRequest, path)) return;

    const id = requestAnimationFrame(() => {
      const node = nodeRef.current;
      node?.scrollIntoView({ block: "center", inline: "nearest" });
      node?.focus({ preventScroll: true });
    });
    return () => cancelAnimationFrame(id);
  }, [nodeRef, path, revealRequest]);
}
