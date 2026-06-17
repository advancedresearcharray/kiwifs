import { describe, expect, it } from "vitest";
import { FileCheck, FileText } from "lucide-react";
import {
  filterSlashCommands,
  matchesSlashQuery,
  resolveLucideIcon,
  templateLoadErrorMessage,
} from "./editorSlashCommands";

describe("editorSlashCommands", () => {
  it("resolves lucide icon names from config", () => {
    expect(resolveLucideIcon("FileCheck")).toBe(FileCheck);
    expect(resolveLucideIcon("")).toBe(FileText);
    expect(resolveLucideIcon("NotARealIcon")).toBe(FileText);
  });

  it("filters commands by id or label prefix", () => {
    const commands = [
      { id: "adr", label: "ADR", icon: "", description: "", template: "templates/adr.md" },
      { id: "runbook", label: "Runbook Step", icon: "", description: "", template: "templates/runbook.md" },
    ];
    expect(filterSlashCommands(commands, "ad")).toHaveLength(1);
    expect(filterSlashCommands(commands, "run")).toHaveLength(1);
    expect(filterSlashCommands(commands, "")).toHaveLength(2);
  });

  it("matches slash query case-insensitively", () => {
    expect(matchesSlashQuery("ADR", "ad")).toBe(true);
    expect(matchesSlashQuery("runbook", "Run")).toBe(true);
    expect(matchesSlashQuery("adr", "book")).toBe(false);
  });

  it("formats template load errors", () => {
    expect(templateLoadErrorMessage("templates/missing.md", new Error("404"))).toContain("templates/missing.md");
    expect(templateLoadErrorMessage("templates/missing.md", new Error("404"))).toContain("404");
  });
});
