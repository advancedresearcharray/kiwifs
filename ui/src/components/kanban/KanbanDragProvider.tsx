import { createContext, useCallback, useContext, useEffect, useMemo, useRef, type ReactNode } from "react";
import {
  DndContext,
  closestCorners,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";

type DragHandlers = {
  onDragStart?: (event: DragStartEvent) => void;
  onDragEnd?: (event: DragEndEvent) => void;
};

type RegisterDragHandlers = (handlers: DragHandlers | null) => void;

const KanbanDragHandlersContext = createContext<RegisterDragHandlers>(() => {});

export function KanbanDragProvider({ children }: { children: ReactNode }) {
  const handlersRef = useRef<DragHandlers | null>(null);

  const register = useCallback<RegisterDragHandlers>((handlers) => {
    handlersRef.current = handlers;
  }, []);

  const contextValue = useMemo(() => register, [register]);

  return (
    <KanbanDragHandlersContext.Provider value={contextValue}>
      <DndContext
        collisionDetection={closestCorners}
        onDragStart={(event) => handlersRef.current?.onDragStart?.(event)}
        onDragEnd={(event) => handlersRef.current?.onDragEnd?.(event)}
      >
        {children}
      </DndContext>
    </KanbanDragHandlersContext.Provider>
  );
}

export function useKanbanDragHandlers(handlers: DragHandlers) {
  const register = useContext(KanbanDragHandlersContext);

  useEffect(() => {
    register(handlers);
    return () => register(null);
  }, [handlers, register]);
}
