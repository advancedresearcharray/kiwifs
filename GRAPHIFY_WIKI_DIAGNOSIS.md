# Graphify Wiki Export Diagnosis And Handoff

## Summary

`graphify . wiki` fails because the installed Graphify CLI uses `graphify <command>` syntax. The current equivalent flow is:

```bash
graphify extract .
graphify cluster-only . --no-viz
graphify export wiki
```

However, `graphify export wiki` also fails on the graph produced by `graphify extract .` in `graphifyy 0.8.8`. The export path assumes a NetworkX node-link JSON key named `links`, while the extract path writes `edges`.

This document is intended as a handoff for another agent to prepare an upstream PR if the diagnosis is accepted.

## Environment

- Project: `/Users/cinos81/works/kiwifs`
- Graphify executable: `/Users/cinos81/.local/bin/graphify`
- Graphify package version: `graphifyy 0.8.8`
- Graphify package path: `/Users/cinos81/.local/share/uv/tools/graphifyy/lib/python3.12/site-packages/graphify`
- Python used by executable: `/Users/cinos81/.local/share/uv/tools/graphifyy/bin/python`

## Reproduction

From the project root:

```bash
graphify . wiki
```

Actual result:

```text
error: unknown command '.'
Run 'graphify --help' for usage.
```

This is expected for the installed CLI shape because `.` is parsed as the command name.

The next attempted command sequence was:

```bash
graphify extract . --no-cluster
graphify cluster-only . --no-viz
graphify export wiki
```

`graphify extract . --no-cluster` produced `graphify-out/graph.json`. Semantic extraction failed because the tool environment is missing the `openai` Python package, but AST extraction still produced a graph:

```text
[graphify extract] found 489 code, 65 docs, 0 papers, 4 images
[graphify] chunk 1/1 failed: Gemini/Kimi/Ollama/OpenAI-compatible extraction requires the openai package. Run: pip install openai
[graphify extract] wrote /Users/cinos81/works/kiwifs/graphify-out/graph.json - 5438 nodes, 12335 edges (no clustering)
```

`graphify export wiki` then failed:

```text
Traceback (most recent call last):
  File "/Users/cinos81/.local/bin/graphify", line 10, in <module>
    sys.exit(main())
  File "/Users/cinos81/.local/share/uv/tools/graphifyy/lib/python3.12/site-packages/graphify/__main__.py", line 2222, in main
    G = _jg.node_link_graph(_raw, edges="links")
  File ".../networkx/readwrite/json_graph/node_link.py", line 247, in node_link_graph
    for d in data[edges]:
KeyError: 'links'
```

## Observed JSON Schema

`graphify-out/graph.json` has this top-level shape:

```text
['edges', 'hyperedges', 'input_tokens', 'nodes', 'output_tokens']
```

There is no `links` key:

```text
nodes: 5438
edges: 12335
links: 0
first edge keys: ['confidence', 'context', 'relation', 'source', 'source_file', 'source_location', 'target', 'weight']
```

## Root Causes

1. CLI documentation or skill examples are stale for direct shell usage.

The command `graphify . wiki` matches the agent skill style `/graphify <path> --wiki`, but it is not valid for the installed shell CLI. The shell CLI expects a command first, such as `extract`, `query`, `path`, `explain`, or `export`.

2. `graphify export wiki` does not normalize `edges` to `links` before loading.

In `graphify/__main__.py`, the export command loads graph JSON at lines 2217-2224:

```python
from networkx.readwrite import json_graph as _jg
from graphify.build import build_from_json as _bfj

_raw = json.loads(graph_path.read_text(encoding="utf-8"))
try:
    G = _jg.node_link_graph(_raw, edges="links")
except TypeError:
    G = _jg.node_link_graph(_raw)
```

This catches `TypeError`, but the observed failure is `KeyError: 'links'`. Other Graphify modules already contain compatibility code for the same schema mismatch.

Examples found in the installed package:

- `graphify/serve.py` normalizes `edges` to `links` before `node_link_graph`.
- `graphify/global_graph.py` normalizes `edges` to `links` before `node_link_graph`.
- `graphify/__main__.py` has similar normalization in other command branches around lines 1540, 1597, 1678, and 2037.

## Proposed Upstream Fix

Patch `graphify/__main__.py` in the `export` branch before `node_link_graph`:

```python
_raw = json.loads(graph_path.read_text(encoding="utf-8"))
if "links" not in _raw and "edges" in _raw:
    _raw = dict(_raw, links=_raw["edges"])
try:
    G = _jg.node_link_graph(_raw, edges="links")
except TypeError:
    G = _jg.node_link_graph(_raw)
```

This matches existing Graphify compatibility patterns and should preserve backward compatibility with older `links`-based files.

## Validation Already Performed

This compatibility snippet successfully loads the generated graph:

```bash
/Users/cinos81/.local/share/uv/tools/graphifyy/bin/python - <<'PY'
import json
from pathlib import Path
from networkx.readwrite import json_graph

p = Path('graphify-out/graph.json')
d = json.loads(p.read_text())
if 'links' not in d and 'edges' in d:
    d = dict(d, links=d['edges'])
G = json_graph.node_link_graph(d, edges='links')
print(G.number_of_nodes(), G.number_of_edges())
PY
```

Observed result:

```text
5727 12335
```

The node count is higher than the explicit `nodes` list because NetworkX adds endpoint nodes referenced by edges when they are missing from `nodes`. That is separate from the `links`/`edges` loader crash.

## Suggested Upstream Tests

Add or update tests to cover the export loader path with `edges`-only graph JSON.

Recommended cases:

- `graphify export wiki --graph <edges-only graph.json>` does not raise `KeyError`.
- `graphify export html --graph <edges-only graph.json>` still works.
- Existing `links`-based graph JSON still works.
- Optional: graph JSON with missing edge endpoint nodes either remains accepted as current NetworkX behavior or is explicitly validated before export.

Minimal test fixture shape:

```json
{
  "nodes": [
    {"id": "a", "label": "A", "file_type": "code", "source_file": "a.py"},
    {"id": "b", "label": "B", "file_type": "code", "source_file": "b.py"}
  ],
  "edges": [
    {"source": "a", "target": "b", "relation": "references", "confidence": "EXTRACTED", "weight": 1.0}
  ],
  "hyperedges": [],
  "input_tokens": 0,
  "output_tokens": 0
}
```

For wiki export, also provide `.graphify_analysis.json` with a non-empty `communities` map because the command intentionally refuses wiki export when community data is missing.

## Local Project Change Made

The local OpenCode Graphify plugin had a separate shell-safety issue. It injected a reminder command containing shell backticks around `graphify query "<question>"`. In shell, those backticks trigger command substitution, so a harmless reminder can accidentally execute `graphify query "<question>"` before the intended command.

Patched file:

- `.opencode/plugins/graphify.js`

Change:

- Removed shell backticks from the reminder text.
- Switched the injected prefix to `printf '%s\n' ${JSON.stringify(reminder)}` so quotes are shell-safe.

The running OpenCode session will not hot-reload plugin changes. Restart OpenCode before relying on this plugin patch.

## Handoff Instructions For Upstream PR Agent

1. Clone the upstream Graphify repository that publishes `graphifyy`.
2. Locate the `graphify/__main__.py` `export` command branch.
3. Apply the `edges` to `links` normalization before the shared `node_link_graph(..., edges="links")` call.
4. Add regression tests for `graphify export wiki` with `edges`-only graph JSON.
5. Run the upstream test suite and the focused CLI regression test.
6. Mention in the PR that this aligns export loading behavior with `serve.py`, `global_graph.py`, and other `__main__.py` branches that already normalize `edges` to `links`.

## Residual Risks

- The semantic extraction failure is environmental, not part of the wiki crash: the installed tool environment lacks the `openai` Python package required by the configured LLM client path.
- `graphify --help` does not clearly document `graphify export wiki`, even though the command exists. A separate documentation/help improvement may be useful.
- If `graphify export wiki` succeeds after the loader fix, there may still be content-quality gaps when semantic extraction was skipped or failed.
