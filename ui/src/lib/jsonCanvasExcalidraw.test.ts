import { describe, expect, it } from "vitest";
import {
  excalidrawSceneToJsonCanvas,
  jsonCanvasToExcalidrawScene,
} from "./jsonCanvasExcalidraw";

describe("jsonCanvasExcalidraw", () => {
  it("converts JSON Canvas nodes and edges into Excalidraw elements", () => {
    const scene = jsonCanvasToExcalidrawScene({
      nodes: [
        { id: "a", type: "text", x: 10, y: 20, width: 160, height: 80, text: "Hello" },
        { id: "b", type: "file", x: 260, y: 20, width: 160, height: 80, file: "notes/page.md" },
      ],
      edges: [
        { id: "e1", fromNode: "a", toNode: "b", label: "links" },
      ],
    });

    expect(scene.type).toBe("excalidraw");
    expect(scene.elements).toEqual(expect.arrayContaining([
      expect.objectContaining({ id: "node:a", type: "rectangle", x: 10, y: 20, width: 160, height: 80 }),
      expect.objectContaining({ id: "label:a", type: "text", text: "Hello", containerId: "node:a" }),
      expect.objectContaining({ id: "node:b", type: "rectangle", x: 260, y: 20, width: 160, height: 80 }),
      expect.objectContaining({ id: "label:b", type: "text", text: "notes/page.md", containerId: "node:b" }),
      expect.objectContaining({ id: "edge:e1", type: "arrow", startBinding: expect.objectContaining({ elementId: "node:a" }), endBinding: expect.objectContaining({ elementId: "node:b" }) }),
      expect.objectContaining({ id: "edge-label:e1", type: "text", text: "links", containerId: "edge:e1" }),
    ]));
  });

  it("round-trips editable Excalidraw elements back to JSON Canvas", () => {
    const scene = jsonCanvasToExcalidrawScene({
      nodes: [
        { id: "a", type: "text", x: 10, y: 20, width: 160, height: 80, text: "Hello" },
        { id: "b", type: "file", x: 260, y: 20, width: 160, height: 80, file: "notes/page.md" },
      ],
      edges: [
        { id: "e1", fromNode: "a", toNode: "b", label: "links" },
      ],
    });

    const textLabel = scene.elements.find((element) => element.id === "label:a");
    if (textLabel) textLabel.text = "Updated";

    const canvas = excalidrawSceneToJsonCanvas(scene.elements);

    expect(canvas.nodes).toEqual(expect.arrayContaining([
      expect.objectContaining({ id: "a", type: "text", x: 10, y: 20, width: 160, height: 80, text: "Updated" }),
      expect.objectContaining({ id: "b", type: "file", x: 260, y: 20, width: 160, height: 80, file: "notes/page.md" }),
    ]));
    expect(canvas.edges).toEqual([
      expect.objectContaining({ id: "e1", fromNode: "a", toNode: "b", label: "links" }),
    ]);
  });
});
