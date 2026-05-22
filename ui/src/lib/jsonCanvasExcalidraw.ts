// Converts between JSON Canvas 1.0 documents and Excalidraw scenes.
//
// The KiwiFS canvas API stores JSON Canvas data. The UI renders it with
// Excalidraw while keeping save/load compatibility with the API format.

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

export type ExcalidrawElementLike = {
  id: string;
  type: string;
  x: number;
  y: number;
  width?: number;
  height?: number;
  text?: string;
  containerId?: string;
  customData?: Record<string, unknown>;
  isDeleted?: boolean;
  startBinding?: { elementId: string; [key: string]: unknown } | null;
  endBinding?: { elementId: string; [key: string]: unknown } | null;
  boundElements?: Array<{ id: string; type: string }> | null;
  [key: string]: unknown;
};

export type ExcalidrawSceneLike = {
  type: "excalidraw";
  version: number;
  source: string;
  elements: ExcalidrawElementLike[];
  appState: Record<string, unknown>;
  files: Record<string, unknown>;
};

const DEFAULT_WIDTH = 200;
const DEFAULT_HEIGHT = 100;

export function jsonCanvasToExcalidrawScene(canvas: JSONCanvas): ExcalidrawSceneLike {
  const elements: ExcalidrawElementLike[] = [];

  for (const node of canvas.nodes) {
    elements.push(createNodeElement(node));
    if (node.type !== "group") {
      elements.push(createNodeLabelElement(node));
    }
  }

  for (const edge of canvas.edges) {
    const fromNode = canvas.nodes.find((node) => node.id === edge.fromNode);
    const toNode = canvas.nodes.find((node) => node.id === edge.toNode);
    if (!fromNode || !toNode) continue;

    elements.push(createEdgeElement(edge, fromNode, toNode));
    if (edge.label) {
      elements.push(createEdgeLabelElement(edge, fromNode, toNode));
    }
  }

  return {
    type: "excalidraw",
    version: 2,
    source: "kiwifs-json-canvas",
    elements,
    appState: {
      viewBackgroundColor: "#ffffff",
      collaborators: new Map(),
    },
    files: {},
  };
}

export function excalidrawSceneToJsonCanvas(elements: readonly ExcalidrawElementLike[]): JSONCanvas {
  const activeElements = elements.filter((element) => element.isDeleted !== true);
  const labelsByContainer = new Map<string, ExcalidrawElementLike>();
  for (const element of activeElements) {
    if (element.type === "text" && typeof element.containerId === "string") {
      labelsByContainer.set(element.containerId, element);
    }
  }

  const nodes: JSONCanvasNode[] = [];
  const edges: JSONCanvasEdge[] = [];

  for (const element of activeElements) {
    if (!element.id.startsWith("node:")) continue;
    const customData = element.customData ?? {};
    const nodeId = getCustomString(customData, "canvasNodeId") ?? element.id.replace(/^node:/, "");
    const canvasType = getCanvasNodeType(customData);
    const label = labelsByContainer.get(element.id)?.text ?? "";
    const base = {
      id: nodeId,
      x: element.x,
      y: element.y,
      width: element.width ?? DEFAULT_WIDTH,
      height: element.height ?? DEFAULT_HEIGHT,
    };

    if (canvasType === "file") {
      nodes.push({ ...base, type: "file", file: getCustomString(customData, "kiwiPath") ?? label });
    } else if (canvasType === "link") {
      nodes.push({ ...base, type: "link", url: getCustomString(customData, "url") ?? label });
    } else if (canvasType === "group") {
      nodes.push({ ...base, type: "group", text: label || getCustomString(customData, "text") || "Group" });
    } else {
      nodes.push({ ...base, type: "text", text: label });
    }
  }

  for (const element of activeElements) {
    if (!element.id.startsWith("edge:")) continue;
    const customData = element.customData ?? {};
    const edgeId = getCustomString(customData, "canvasEdgeId") ?? element.id.replace(/^edge:/, "");
    const fromNode = bindingToNodeId(element.startBinding) ?? getCustomString(customData, "fromNode");
    const toNode = bindingToNodeId(element.endBinding) ?? getCustomString(customData, "toNode");
    if (!fromNode || !toNode) continue;
    const label = labelsByContainer.get(element.id)?.text;
    edges.push({ id: edgeId, fromNode, toNode, ...(label ? { label } : {}) });
  }

  return { nodes, edges };
}

function createNodeElement(node: JSONCanvasNode): ExcalidrawElementLike {
  const isGroup = node.type === "group";
  const labelId = `label:${node.id}`;
  return baseElement({
    id: `node:${node.id}`,
    type: isGroup ? "rectangle" : "rectangle",
    x: node.x,
    y: node.y,
    width: node.width,
    height: node.height,
    backgroundColor: nodeFill(node),
    fillStyle: isGroup ? "hachure" : "solid",
    strokeStyle: isGroup ? "dashed" : "solid",
    boundElements: isGroup ? [] : [{ id: labelId, type: "text" }],
    customData: {
      canvasNodeId: node.id,
      canvasType: node.type,
      ...(node.file ? { kiwiPath: node.file } : {}),
      ...(node.url ? { url: node.url } : {}),
      ...(node.text ? { text: node.text } : {}),
    },
  });
}

function createNodeLabelElement(node: JSONCanvasNode): ExcalidrawElementLike {
  const text = node.type === "file" ? node.file : node.type === "link" ? node.url : node.text;
  return baseTextElement({
    id: `label:${node.id}`,
    x: node.x + 8,
    y: node.y + 8,
    width: Math.max(1, node.width - 16),
    height: Math.max(1, node.height - 16),
    text: text ?? "",
    containerId: `node:${node.id}`,
  });
}

function createEdgeElement(edge: JSONCanvasEdge, fromNode: JSONCanvasNode, toNode: JSONCanvasNode): ExcalidrawElementLike {
  const start = sidePoint(fromNode, edge.fromSide ?? "right");
  const end = sidePoint(toNode, edge.toSide ?? "left");
  return baseElement({
    id: `edge:${edge.id}`,
    type: "arrow",
    x: start.x,
    y: start.y,
    width: end.x - start.x,
    height: end.y - start.y,
    points: [[0, 0], [end.x - start.x, end.y - start.y]],
    endArrowhead: "arrow",
    startBinding: { elementId: `node:${edge.fromNode}`, fixedPoint: sideFixedPoint(edge.fromSide ?? "right") },
    endBinding: { elementId: `node:${edge.toNode}`, fixedPoint: sideFixedPoint(edge.toSide ?? "left") },
    boundElements: edge.label ? [{ id: `edge-label:${edge.id}`, type: "text" }] : [],
    customData: {
      canvasEdgeId: edge.id,
      fromNode: edge.fromNode,
      toNode: edge.toNode,
    },
  });
}

function createEdgeLabelElement(edge: JSONCanvasEdge, fromNode: JSONCanvasNode, toNode: JSONCanvasNode): ExcalidrawElementLike {
  const start = sidePoint(fromNode, edge.fromSide ?? "right");
  const end = sidePoint(toNode, edge.toSide ?? "left");
  return baseTextElement({
    id: `edge-label:${edge.id}`,
    x: (start.x + end.x) / 2 - 40,
    y: (start.y + end.y) / 2 - 12,
    width: 80,
    height: 24,
    text: edge.label ?? "",
    containerId: `edge:${edge.id}`,
  });
}

function baseElement(element: Partial<ExcalidrawElementLike> & { id: string; type: string; x: number; y: number }): ExcalidrawElementLike {
  return {
    angle: 0,
    strokeColor: "#1e1e1e",
    backgroundColor: "transparent",
    fillStyle: "solid",
    strokeWidth: 2,
    strokeStyle: "solid",
    roughness: 1,
    opacity: 100,
    width: DEFAULT_WIDTH,
    height: DEFAULT_HEIGHT,
    seed: stableSeed(element.id),
    version: 1,
    versionNonce: stableSeed(`${element.id}:version`),
    isDeleted: false,
    groupIds: [],
    frameId: null,
    roundness: { type: 3 },
    ...element,
  };
}

function baseTextElement(element: Partial<ExcalidrawElementLike> & { id: string; x: number; y: number; text: string }): ExcalidrawElementLike {
  return baseElement({
    type: "text",
    fontSize: 18,
    fontFamily: 1,
    textAlign: "center",
    verticalAlign: "middle",
    baseline: 18,
    originalText: element.text,
    autoResize: true,
    backgroundColor: "transparent",
    fillStyle: "solid",
    strokeWidth: 1,
    roughness: 0,
    ...element,
  });
}

function nodeFill(node: JSONCanvasNode): string {
  if (node.color) return node.color;
  if (node.type === "file") return "#a5d8ff";
  if (node.type === "link") return "#b2f2bb";
  if (node.type === "group") return "transparent";
  return "#fff3bf";
}

function sidePoint(node: JSONCanvasNode, side: NonNullable<JSONCanvasEdge["fromSide"]>): { x: number; y: number } {
  if (side === "top") return { x: node.x + node.width / 2, y: node.y };
  if (side === "bottom") return { x: node.x + node.width / 2, y: node.y + node.height };
  if (side === "left") return { x: node.x, y: node.y + node.height / 2 };
  return { x: node.x + node.width, y: node.y + node.height / 2 };
}

function sideFixedPoint(side: NonNullable<JSONCanvasEdge["fromSide"]>): [number, number] {
  if (side === "top") return [0.5, 0];
  if (side === "bottom") return [0.5, 1];
  if (side === "left") return [0, 0.5];
  return [1, 0.5];
}

function bindingToNodeId(binding: ExcalidrawElementLike["startBinding"]): string | null {
  const elementId = binding?.elementId;
  return typeof elementId === "string" ? elementId.replace(/^node:/, "") : null;
}

function getCanvasNodeType(customData: Record<string, unknown>): JSONCanvasNode["type"] {
  const type = getCustomString(customData, "canvasType");
  return type === "file" || type === "link" || type === "group" ? type : "text";
}

function getCustomString(customData: Record<string, unknown>, key: string): string | undefined {
  const value = customData[key];
  return typeof value === "string" ? value : undefined;
}

function stableSeed(value: string): number {
  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = (hash * 31 + value.charCodeAt(index)) | 0;
  }
  return Math.abs(hash) || 1;
}
