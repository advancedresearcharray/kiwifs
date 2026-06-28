import type { Meta, StoryObj } from "@storybook/react";
import { MermaidDiagram } from "./MermaidDiagram";

const meta: Meta<typeof MermaidDiagram> = {
  title: "Content/MermaidDiagram",
  component: MermaidDiagram,
  parameters: { layout: "padded" },
  decorators: [
    (Story) => (
      <div className="max-w-4xl mx-auto p-8 bg-background text-foreground">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof MermaidDiagram>;

export const Flowchart: Story = {
  args: {
    chart: `graph TD
    A[Start] --> B{Decision}
    B -->|Yes| C[Action 1]
    B -->|No| D[Action 2]
    C --> E[End]
    D --> E`,
  },
};

export const SequenceDiagram: Story = {
  args: {
    chart: `sequenceDiagram
    participant Browser
    participant KiwiFS
    participant SQLite

    Browser->>KiwiFS: GET /api/search?q=hello
    KiwiFS->>SQLite: FTS5 query
    SQLite-->>KiwiFS: ranked results
    KiwiFS-->>Browser: JSON response`,
  },
};

export const ClassDiagram: Story = {
  args: {
    chart: `classDiagram
    class KiwiFS {
      +string dataDir
      +int port
      +serve()
      +index()
    }
    class SearchEngine {
      +query(q: string)
      +reindex()
    }
    class FileWatcher {
      +watch(dir: string)
      +onChanged()
    }
    KiwiFS --> SearchEngine
    KiwiFS --> FileWatcher`,
  },
};

export const GitGraph: Story = {
  args: {
    chart: `gitGraph
    commit id: "init"
    branch feat/mermaid
    checkout feat/mermaid
    commit id: "add component"
    commit id: "fix dark mode"
    checkout main
    merge feat/mermaid
    commit id: "release"`,
  },
};

export const PieChart: Story = {
  args: {
    chart: `pie title KiwiFS File Types
    "Markdown" : 65
    "Excalidraw" : 15
    "Images" : 12
    "Other" : 8`,
  },
};

export const StateDiagram: Story = {
  args: {
    chart: `stateDiagram-v2
    [*] --> Draft
    Draft --> Published: publish
    Published --> Archived: archive
    Archived --> Published: restore
    Draft --> [*]: delete`,
  },
};

export const InvalidSyntax: Story = {
  args: {
    chart: `graph INVALID
    this is not valid mermaid --->>> syntax !!!`,
  },
};

export const Empty: Story = {
  args: {
    chart: "",
  },
};
