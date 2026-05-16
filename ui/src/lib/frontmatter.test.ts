import { describe, expect, it } from "vitest";
import { frontmatterToText, joinFrontmatter, splitFrontmatter } from "./frontmatter";

describe("frontmatter helpers", () => {
  it("splits only leading YAML frontmatter", () => {
    const md = "---\ntitle: Example\ntags:\n  - wiki\n---\n\n# Body\n---\nnot frontmatter\n";

    expect(splitFrontmatter(md)).toEqual({
      frontmatter: "---\ntitle: Example\ntags:\n  - wiki\n---",
      body: "# Body\n---\nnot frontmatter\n",
    });
  });

  it("does not split thematic breaks in the body", () => {
    const md = "# Body\n\n---\n\ncontent\n";
    expect(splitFrontmatter(md)).toEqual({ frontmatter: null, body: md });
  });

  it("keeps invalid leading frontmatter-looking blocks lossless", () => {
    const md = "---\nnot: [valid\n---\n\n# Body\n";
    const split = splitFrontmatter(md);

    expect(frontmatterToText(split.frontmatter)).toBe("not: [valid");
    expect(joinFrontmatter(frontmatterToText(split.frontmatter), split.body)).toBe(md);
  });

  it("joins edited frontmatter with editor body without exposing delimiters to WYSIWYG", () => {
    expect(joinFrontmatter("title: Edited", "# Body\n")).toBe("---\ntitle: Edited\n---\n\n# Body\n");
  });
});
