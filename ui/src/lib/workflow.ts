import type { WorkflowDef } from "./api";

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
  return {
    name,
    states: states.map((state, index) => ({
      name: state,
      color: DEFAULT_STATE_COLORS[index % DEFAULT_STATE_COLORS.length]!,
    })),
    transitions: createAdjacentTransitions(states),
  };
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
