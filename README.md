# uix

`uix` is a dependency-free Go CLI that converts verbose UiPath XAML into a
compact intermediate representation for humans and LLMs. It never executes a
workflow; it only performs static XML analysis.

## MVP commands

```text
uix compact [--refs] [-o DIR] PATH...
uix strip [--conservative] [-o DIR] PATH...
uix index [-o FILE] PROJECT_DIR
uix read [--refs=false] PROJECT_DIR WORKFLOW
uix graph [--format text|mermaid] PROJECT_DIR
uix version
```

### compact

Converts a XAML file or every XAML file under a directory into a compact,
indented IR. Known activities receive specialized semantic forms.

```text
Assign [customerName] = ["Makima"]
If [String.IsNullOrEmpty(apiKey)]
  Then:
    Log Error ["API Key is empty"]
```

With `--refs`, each activity includes its `WorkflowViewState.IdRef`, or a stable
XML node path when no designer ID exists. This lets an LLM refer back to the
original XAML.

### strip

Removes known noise while preserving the original XAML syntax and namespace
prefixes.

- Default mode removes designer metadata, implementation namespace/reference
  lists, screenshot metadata, and other read-only noise.
- `--conservative` removes only ViewState, HintSize, and IdRef data.
- Every result is parsed again as XML; invalid output is rejected.
- The command refuses to overwrite an input file.

`strip` is intended for lightweight, read-only inspection. Full round-trip
compatibility with UiPath Studio is not guaranteed until validated by opening,
analyzing, and saving representative real-world projects.

### index

Reads `project.json` and every XAML file, then emits JSON containing workflows,
arguments, static and dynamic invokes, and approximate activity counts.

### read

Finds a workflow by relative path or unique filename and emits its IR with
source references. Use it after `index` to load only the relevant workflow.

### graph

Emits static `InvokeWorkflowFile` relationships as text or Mermaid.

## Releases

Every push to `main`, including a merged pull request, runs the Release workflow.
Tests and `go vet` must pass before binaries are published to
[GitHub Releases](https://github.com/makima333/uipath-xaml-tools/releases).
The workflow can also be started manually on `main` from the Actions tab.

Release tags use `v0.1.<workflow run number>` (for example, `v0.1.1`). Each run
creates a new release for its exact commit; re-running the same run reuses its tag.
The version is embedded in the binary and printed by `uix version`.

Assets include Windows ZIP archives and Linux/macOS tar.gz archives, each for
AMD64 and ARM64, plus `checksums.txt` with SHA-256 hashes. Extract the archive for
your platform and run `uix` (`uix.exe` on Windows). No Go installation is needed.

To build the same archives locally, run `bash scripts/build-release.sh v0.1.1`
from the repository root with Go, Bash, zip, tar, and sha256sum installed.
Artifacts are written to `dist/`.

## Development

Requires Go 1.27 or newer.

```powershell
go test ./...
go build -buildvcs=false -o bin/uix.exe ./cmd/uix
./bin/uix.exe version
```

See [docs/design.md](docs/design.md) for the architecture and MVP boundaries.
