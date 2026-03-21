# nuke-engine

Build C++ with plain Go.

`nuke-engine` is a replacement for CMake + Ninja when you want your build logic to be real Go code. Define C++ targets in code, compose them across modules, and get content-addressed incremental builds without a separate DSL.

It is type-safe and IDE/IntelliSense friendly. No more string-bashing macros or brittle CMake generator expressions. You can use full Go language features, normal tooling, and refactoring support for your build logic.

<p align="center">
    <a href="docs/nuke.jpg">
        <img src="docs/nuke.jpg" alt="Nuke Build Engine poster" width="760" style="max-width: 100%; height: auto; border-radius: 14px; box-shadow: 0 12px 32px rgba(0, 0, 0, 0.28);" />
    </a>
</p>

<p align="center">
    <sub><strong>Nuke Build Engine</strong> - Build like a pro in the wasteland.</sub>
</p>

## Quickstart

Build a C++ executable from scratch in under five minutes.

Prerequisites: `nuke-build` installed via [INSTALL.md](INSTALL.md), Visual Studio 2019/2022 on PATH.

### 1) Project layout

```text
myproject/
|-- go.work
`-- app/
    |-- go.mod
    |-- nuke.go
    `-- src/
        `-- main.cpp
```

### 2) Create source file

```cpp
// app/src/main.cpp
#include <cstdio>
int main() { std::puts("hello from nuke!"); }
```

### 3) Create module file

```text
# app/go.mod
module myproject

go 1.21
```

### 4) Write build rules

```go
// app/nuke.go
package app

import (
    "github.com/bradphelan/nuke-engine/cpp"
)

var Def = cpp.Define(func(self cpp.Self) *cpp.Target {
    return self.Exe().
        Sources(self.Glob("src/**/*.cpp"))
})
```

`Def` is your package-level declaration. `self` is a bound helper for the declaration currently being realized.

For library modules, include scope is fluent and type-safe (default scope is public):

```go
return self.StaticLib().
    WithIncludeDir(filepath.Join(self.Dir(), "include")).
    Private().
    WithIncludeDir(filepath.Join(self.Dir(), "src", "internal")).
    Sources(self.Glob("src/**/*.cpp"))
```

### 5) Create workspace file

```text
# go.work
go 1.21

use (
    ./app
)
```

### 6) Build

```sh
nuke-build --project myproject
```

You can also point directly at a module package:

```sh
nuke-build --project myproject/app
```

For this simple layout, both commands resolve the same build root.

First run compiles everything:

```text
Stats: cache_hits=0 cache_misses=3 rules_executed=3 outputs=1
Tasks executed: 3
Built: file:///.../myproject/build/bin/app.exe
```

Run again and unchanged work is skipped:

```text
Stats: cache_hits=3 cache_misses=0 rules_executed=0 outputs=1
No tasks need running (up-to-date).
Built: file:///.../myproject/build/bin/app.exe
```

That build command will:

1. Discover your build modules.
2. Generate a temporary bootstrap under `build/_nuke/`.
3. Import all discovered module packages.
4. Resolve registered `Def` declarations.
5. Build executable roots and their transitive dependencies.

## Why this is better than another build DSL

- Build rules are normal Go code, so composition and refactoring are just code.
- C++ targets link by importing sibling modules and referencing exported `Def` values.
- Incremental rebuilds are automatic because outputs are fingerprinted from their inputs.
- The generated bootstrap lives in `build/`, not in your source tree.
- A workspace root can contain multiple executables; `nuke-build` will discover them and build executable roots by default.

## Architecture

Detailed internals, pipeline diagrams, and component breakdown moved to [ARCHITECTURE.md](ARCHITECTURE.md).

## Common commands

```sh
# Build from a specific module package
nuke-build --project myproject/app

# Build from a workspace root containing child modules
nuke-build --project myproject
```

Both forms are valid. In workspace-root mode, `nuke-build` imports all discovered child modules and resolves every registered `Def`.

## Docs

| Document | What it covers |
|---|---|
| [INSTALL.md](INSTALL.md) | Installing `nuke-build` on your machine |
| [README.md](README.md) | Quickstart and product overview |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Internal architecture, bootstrap pipeline, and graph model |
| [USERGUIDE.md](USERGUIDE.md) | Full reference — targets, config, cache, CLI |
| [QUICKSTART.md](QUICKSTART.md) | Legacy entry point (redirects to README quickstart) |

## Supported platforms / toolchains

| Platform | Toolchain | Status |
|---|---|---|
| Windows | MSVC (VS 2019 / 2022) | ✅ |
| Linux | GCC / Clang | planned |
| macOS | Clang | planned |

## License

MIT
