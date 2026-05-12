// Converts between JSON Canvas format and tldraw shapes.
//
// JSON Canvas spec:
// {
//   nodes: [{ id, x, y, width, height, type, text?, file?, color? }],
//   edges: [{ id, fromNode, toNode, label?, color? }]
// }

export type JSONCanvasNode = {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
  type: "text" | "file" | "link" | "group";
  text?: string;
  file?: string;
  url?: string;
  color?: string;
};

export type JSONCanvasEdge = {
  id: string;
  fromNode: string;
  toNode: string;
  fromSide?: "top" | "right" | "bottom" | "left";
  toSide?: "top" | "right" | "bottom" | "left";
  label?: string;
  color?: string;
};

export type JSONCanvas = {
  nodes: JSONCanvasNode[];
  edges: JSONCanvasEdge[];
};

// For tldraw v5, we use simple shape records that map to the store.
export type TldrawShapeRecord = {
  id: string;
  type: string;
  x: number;
  y: number;
  props: Record<string, unknown>;
  meta?: Record<string, unknown>;
};

/**
 * Convert a JSON Canvas document to tldraw-compatible shape records.
 * Text nodes become "text" shapes; file nodes become "note" shapes
 * with metadata referencing the kiwi page path.
 */
export function jsonCanvasToShapes(canvas: JSONCanvas): TldrawShapeRecord[] {
  const shapes: TldrawShapeRecord[] = [];

  for (const node of canvas.nodes) {
    switch (node.type) {
      case "text":
        shapes.push({
          id: `shape:${node.id}`,
          type: "note",
          x: node.x,
          y: node.y,
          props: {
            w: node.width,
            h: node.height,
            text: node.text || "",
            color: node.color || "yellow",
            size: "m",
          },
          meta: { canvasNodeId: node.id, canvasType: "text" },
        });
        break;
      case "file":
        shapes.push({
          id: `shape:${node.id}`,
          type: "note",
          x: node.x,
          y: node.y,
          props: {
            w: node.width,
            h: node.height,
            text: node.file || "",
            color: "light-blue",
            size: "m",
          },
          meta: {
            canvasNodeId: node.id,
            canvasType: "file",
            kiwiPath: node.file,
          },
        });
        break;
      case "link":
        shapes.push({
          id: `shape:${node.id}`,
          type: "note",
          x: node.x,
          y: node.y,
          props: {
            w: node.width,
            h: node.height,
            text: node.url || "",
            color: "light-green",
            size: "m",
          },
          meta: { canvasNodeId: node.id, canvasType: "link", url: node.url },
        });
        break;
      case "group":
        shapes.push({
          id: `shape:${node.id}`,
          type: "frame",
          x: node.x,
          y: node.y,
          props: {
            w: node.width,
            h: node.height,
            name: node.text || "Group",
          },
          meta: { canvasNodeId: node.id, canvasType: "group" },
        });
        break;
    }
  }

  return shapes;
}

/**
 * Convert tldraw shapes back to JSON Canvas format.
 * Currently a simplified implementation that handles note/frame shapes.
 */
export function shapesToJsonCanvas(
  shapes: TldrawShapeRecord[],
): JSONCanvas {
  const nodes: JSONCanvasNode[] = [];
  const edges: JSONCanvasEdge[] = [];

  for (const shape of shapes) {
    const meta = shape.meta || {};
    const nodeId =
      (meta.canvasNodeId as string) ||
      shape.id.replace("shape:", "");
    const canvasType = (meta.canvasType as string) || "text";

    const w = (shape.props.w as number) || 200;
    const h = (shape.props.h as number) || 100;
    const text = (shape.props.text as string) || "";

    switch (canvasType) {
      case "file":
        nodes.push({
          id: nodeId,
          x: shape.x,
          y: shape.y,
          width: w,
          height: h,
          type: "file",
          file: (meta.kiwiPath as string) || text,
        });
        break;
      case "link":
        nodes.push({
          id: nodeId,
          x: shape.x,
          y: shape.y,
          width: w,
          height: h,
          type: "link",
          url: (meta.url as string) || text,
        });
        break;
      case "group":
        nodes.push({
          id: nodeId,
          x: shape.x,
          y: shape.y,
          width: w,
          height: h,
          type: "group",
          text: (shape.props.name as string) || "",
        });
        break;
      default:
        nodes.push({
          id: nodeId,
          x: shape.x,
          y: shape.y,
          width: w,
          height: h,
          type: "text",
          text,
        });
    }
  }

  return { nodes, edges };
}
