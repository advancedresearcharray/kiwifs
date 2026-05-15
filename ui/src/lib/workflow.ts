import type { WorkflowDef } from "./api";

export type KanbanCardDraft = {
  title: string;
  workflow: string;
  state: string;
  body?: string;
};

const DEFAULT_STATE_COLORS = ["#9B59B6", "#3498DB", "#2ECC71", "#F39C12", "#E74C3C"];

export function normalizeWorkflowName(input: string): string {
  return input.trim().replace(/[\\/]+/g, "-").replace(/\s+/g, " ");
}

export function parseWorkflowStates(input: string): string[] {
  const seen = new Set<string>();
  const states: string[] = [];

  for (const raw of input.split(/[\n,]+/)) {
    const name = raw.trim().replace(/[\\/]+/g, "-").replace(/\s+/g, " ");
    if (!name || seen.has(name)) continue;
    seen.add(name);
    states.push(name);
  }

  return states;
}

export function createDefaultWorkflow(name: string, states: string[]): WorkflowDef {
  return updateWorkflowStates(
    { name, states: [], transitions: [] },
    states.map((state, index) => ({
      name: state,
      color: DEFAULT_STATE_COLORS[index % DEFAULT_STATE_COLORS.length]!,
    })),
  );
}

export function addWorkflowState(workflow: WorkflowDef, stateName: string): WorkflowDef {
  const nextIndex = workflow.states.length;
  return updateWorkflowStates(workflow, [
    ...workflow.states,
    {
      name: stateName,
      color: DEFAULT_STATE_COLORS[nextIndex % DEFAULT_STATE_COLORS.length]!,
    },
  ]);
}

export function renameWorkflowState(workflow: WorkflowDef, from: string, to: string): WorkflowDef {
  return updateWorkflowStates(
    workflow,
    workflow.states.map((state) =>
      state.name === from ? { ...state, name: to } : state,
    ),
  );
}

export function deleteWorkflowState(workflow: WorkflowDef, stateName: string): WorkflowDef {
  return updateWorkflowStates(
    workflow,
    workflow.states.filter((state) => state.name !== stateName),
  );
}

export function updateWorkflowStates(
  workflow: WorkflowDef,
  states: WorkflowDef["states"],
): WorkflowDef {
  const seen = new Set<string>();
  const normalizedStates = states.map((state, index) => ({
    ...state,
    name: normalizeStateName(state.name),
    color: state.color || DEFAULT_STATE_COLORS[index % DEFAULT_STATE_COLORS.length]!,
  }));

  for (const state of normalizedStates) {
    if (!state.name) throw new Error("State name is required.");
    if (seen.has(state.name)) throw new Error(`Duplicate state: ${state.name}`);
    seen.add(state.name);
  }
  if (normalizedStates.length === 0) throw new Error("Add at least one state.");

  const stateNames = normalizedStates.map((state) => state.name);
  return {
    ...workflow,
    states: normalizedStates,
    transitions: createAdjacentTransitions(stateNames),
  };
}

export function defaultKanbanCardPath(title: string, workflowName: string): string {
  const folder = slugifyPathSegment(workflowName) || "kanban";
  const slug = slugifyPathSegment(title) || `card-${Date.now()}`;
  return `${folder}/${slug}.md`;
}

export function createKanbanCardMarkdown(draft: KanbanCardDraft): string {
  const title = draft.title.trim() || "Untitled card";
  const body = draft.body?.trim() ? `\n\n${draft.body.trim()}\n` : "\n";
  return [
    "---",
    `title: ${quoteYamlString(title)}`,
    `workflow: ${quoteYamlString(draft.workflow)}`,
    `state: ${quoteYamlString(draft.state)}`,
    "---",
    `# ${title}`,
  ].join("\n") + body;
}

function quoteYamlString(value: string): string {
  return JSON.stringify(value);
}

function slugifyPathSegment(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[\\/]+/g, "-")
    .replace(/[^\p{L}\p{N}]+/gu, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 80);
}

function normalizeStateName(input: string): string {
  return input.trim().replace(/[\\/]+/g, "-").replace(/\s+/g, " ");
}

function createAdjacentTransitions(states: string[]): WorkflowDef["transitions"] {
  const transitions: WorkflowDef["transitions"] = [];

  for (let index = 0; index < states.length - 1; index += 1) {
    const from = states[index]!;
    const to = states[index + 1]!;
    transitions.push({ from, to }, { from: to, to: from });
  }

  return transitions;
}
