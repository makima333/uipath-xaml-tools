# uix MVP design

## Goal

Reduce UiPath XAML workflow logic to a size that can be loaded progressively
into an LLM context. The tool does not fully deserialize UiPath activity types
and never executes a UiPath package.

## Processing pipeline

```text
XAML bytes
  -> namespace-aware XML AST
  -> UiPath Normalizer
  -> canonical IR
  -> compact/read renderer

project.json + XAML AST
  -> project index
  -> JSON index / text graph / Mermaid graph
```

`strip` uses a separate pipeline because it must retain the original lexical
form, including namespace prefixes. It removes known blocks and attributes
without serializing the AST, then validates the result with the XML parser.

## Package layout

| Package | Responsibility |
|---|---|
| `cmd/uix` | Executable entry point |
| `internal/cli` | Commands, flags, file I/O, and overwrite protection |
| `internal/xaml` | Namespace-aware XML AST and parser |
| `internal/normalize` | UiPath noise removal, property normalization, and reference retention |
| `internal/ir` | Renderer-independent canonical IR |
| `internal/render` | Compact, LLM-oriented text output |
| `internal/project` | `project.json`, arguments, invokes, indexes, and graphs |
| `internal/strip` | Noise removal that preserves the original XAML syntax |
| `internal/workspace` | XAML discovery, relative paths, and workflow resolution |

## AST and IR

The XML AST retains each element's namespace URI, local name, attribute order,
text, children, and stable node path. The IR is limited to:

- activity kind and namespace;
- property roles such as `Then`, `Else`, `Condition`, and `Target`;
- display names, annotations, meaningful attributes, and expressions;
- source references based on designer IDs or XML node paths;
- disabled subtrees; and
- normalized children.

Unknown activities remain as generic nodes. A subtree is never discarded just
because its activity type is unknown.

## Normalizer

The MVP removes:

- designer state such as ViewState and HintSize;
- implementation namespace/reference lists;
- screenshot metadata;
- attributes that only express nulls or known defaults; and
- empty Argument and CursorPosition scaffolding.

The MVP has specialized forms for:

- `Assign`
- `If`
- `InvokeWorkflowFile`
- `LogMessage`
- `GetAsset`

All other data is retained in a generic `Activity [key="value"]` form.

## Command roles

1. `index`: Produce a small, machine-readable project index.
2. `graph`: Inspect static calls between workflows.
3. `read`: Load one relevant workflow with source references.
4. `compact`: Compress any selected set of XAML files.
5. `strip`: Produce lighter XAML when XML structure is still required.

## Explicit MVP non-goals

- Full round-trip compatibility with UiPath Studio
- Expression parsing or type checking
- Resolution of dynamically computed `WorkflowFileName` values
- Specialized output for every UiPath activity
- XAML editing or reverse conversion from IR to XAML

## Candidate extensions

- Measure compression and information retention against a real-project corpus.
- Add an activity-specific normalizer registry.
- Add selector aliases and a repeated-string dictionary.
- Retain source spans and support partial reads from the original XAML.
- Cache indexes and update them incrementally.
- Add a JSON representation of the canonical IR.
