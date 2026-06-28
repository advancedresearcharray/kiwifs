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

const REVEAL_FLASH_CLASS = "kiwi-tree-reveal-flash";

export function useTreeRevealTargetFocus<T extends HTMLElement>(
  revealRequest: TreeRevealRequest | null | undefined,
  path: string,
  nodeRef: RefObject<T | null>,
): void {
  const revealNonce = revealRequest?.nonce;

  useEffect(() => {
    if (!shouldFocusRevealTarget(revealRequest, path)) return;

    const id = requestAnimationFrame(() => {
      const node = nodeRef.current;
      if (!node) return;

      node.scrollIntoView({ block: "center", inline: "nearest" });
      node.focus({ preventScroll: true });

      node.classList.remove(REVEAL_FLASH_CLASS);
      void node.offsetWidth;
      node.classList.add(REVEAL_FLASH_CLASS);

      const onEnd = () => {
        node.classList.remove(REVEAL_FLASH_CLASS);
        node.removeEventListener("animationend", onEnd);
      };
      node.addEventListener("animationend", onEnd);
    });

    return () => cancelAnimationFrame(id);
  }, [nodeRef, path, revealRequest?.path, revealNonce]);
}
