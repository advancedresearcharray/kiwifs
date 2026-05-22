import { describe, expect, it } from "vitest";
import {
  excalidrawSceneToJsonCanvas,
  jsonCanvasToExcalidrawScene,
  type JSONCanvas,
} from "./jsonCanvasExcalidraw";

describe("jsonCanvasExcalidraw", () => {
  it("round-trips JSON Canvas edge fields", () => {
    const canvas: JSONCanvas = {
      nodes: [
        {
          id: "a",
          type: "text",
          x: 0,
          y: 0,
          width: 120,
          height: 60,
          text: "Alpha",
        },
        {
          id: "b",
          type: "file",
          x: 300,
          y: 0,
          width: 200,
          height: 60,
          file: "notes/beta.md",
        },
      ],
      edges: [
        {
          id: "e1",
          fromNode: "a",
          toNode: "b",
          fromSide: "right",
          toSide: "left",
          label: "links",
        },
      ],
    };

    const scene = jsonCanvasToExcalidrawScene(canvas);
    const out = excalidrawSceneToJsonCanvas(scene.elements);

    expect(out.edges).toHaveLength(1);
    expect(out.edges[0]).toMatchObject({
      id: "e1",
      fromNode: "a",
      toNode: "b",
      label: "links",
    });
    expect(out.nodes).toHaveLength(2);
    expect(out.nodes.find((n) => n.id === "b")?.file).toBe("notes/beta.md");
  });
});
