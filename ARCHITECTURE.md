# Architecture

This document covers internal architecture details that are intentionally kept out of the top-level quickstart flow.

## High-level model

- Build declarations are Go packages that export `Def` values.
- `nuke-build` discovers module packages and generates a temporary bootstrap under `build/_nuke/`.
- The bootstrap imports discovered packages, resolves registered target declarations, and builds executable roots.
- The build graph is evaluated as artifact sets with content-addressed caching.

```mermaid
flowchart LR
    A[nuke.go packages\nexport Def] -->|nuke-build discovers| B[generated bootstrap]
    B -->|go build| C[build.exe / build.out]
    C -->|imports modules| D[registered Def set]
    D -->|resolve roots| E[target graph]
    E -->|cache misses only| F[C++ compiler\nMSVC / clang / gcc]
    F --> G[.lib / .exe / .so]
    E -->|cache hit| G
```

## Two-phase pipeline

```mermaid
sequenceDiagram
    participant U as Developer
    participant NB as nuke-build
    participant Go as Go compiler
    participant Builder as build.exe / build.out
    participant CC as C++ compiler

    U->>NB: nuke-build --project physics-engine/
    NB->>NB: resolve project layout\ndiscover module dirs
    NB->>NB: generate bootstrap\nthat imports all modules
    NB->>Go: build/_nuke/main.go\n(generated bootstrap)
    Go-->>NB: build/build.exe (Windows) or build/build.out (Linux)
    NB->>Builder: exec build.exe / build.out [forwarded flags]
    Builder->>Builder: resolve registered Defs\nprefer executable roots
    Builder->>CC: compile/link (cache misses only)
    CC-->>Builder: .obj, .lib, .exe
    Builder-->>U: Stats + Built: path
```

## Graph and configuration semantics

- `LinkPublic` propagates public config transitively to dependents.
- `LinkPrivate` uses dependency public config when compiling this target but does not propagate it further.
- Include/config scope in fluent DSL defaults to public.
- Call `Private()` to switch scope for subsequent `With*` config calls.
- Call `Public()` to switch scope back.

## Build directory contract

- Generated bootstrap files are under `build/_nuke/`.
- Build outputs (objects/libs/exes/cache DB) are under the resolved build root.
- Source tree should remain read-only for nuke-build execution.

## Related docs

- Quickstart and first-run workflow: [README.md](README.md#quickstart)
- Full API and CLI reference: [USERGUIDE.md](USERGUIDE.md)
