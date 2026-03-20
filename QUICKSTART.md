# Quickstart

Build a C++ executable from scratch in under five minutes.

> **Prerequisites**: `nuke-build` installed ([INSTALL.md](INSTALL.md)), Visual Studio 2019/2022 on PATH.

---

## 1. Project layout

```
myproject/
├── go.work          (Go workspace — ties modules together)
├── app/
│   ├── go.mod       (module: myproject)
│   ├── nuke.go      (build rules)
│   └── src/
│       └── main.cpp
```

---

## 2. Create the source file

```cpp
// app/src/main.cpp
#include <cstdio>
int main() { std::puts("hello from nuke!"); }
```

---

## 3. Create the Go module

```
# app/go.mod
module myproject

go 1.21
```

---

## 4. Write the build rules

```go
// app/nuke.go
package app

import (
    "path/filepath"

    "github.com/bradphelan/nuke-engine/artifact"
    "github.com/bradphelan/nuke-engine/compiler"
    "github.com/bradphelan/nuke-engine/cpp"
    "github.com/bradphelan/nuke-engine/target"
)

func Build(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string) *target.Target[compiler.Config] {
    return cpp.NewExe("app", builder, baseCfg).
        Sources(artifact.Glob(filepath.Join(rootDir, "app", "src"), "**/*.cpp"))
}
```

---

## 5. Create the workspace

```
# go.work
go 1.21

use (
    ./app
)
```

---

## 6. Build

```powershell
nuke-build --project myproject/app
```

First run compiles everything:

```
Stats: cache_hits=0 cache_misses=3 rules_executed=3 outputs=1
Tasks executed: 3
Built: file:///.../myproject/build/bin/app.exe
```

Run again — nothing to do:

```
Stats: cache_hits=3 cache_misses=0 rules_executed=0 outputs=1
No tasks need running (up-to-date).
Built: file:///.../myproject/build/bin/app.exe
```

---

## What just happened?

```mermaid
flowchart TD
    A["nuke-build --project app/"] --> B{find nuke.go}
    B --> C[write build/_nuke/main.go\nimports your Build func]
    C --> D["go build → build/nuke-builder.exe"]
    D --> E[nuke-builder.exe runs]
    E --> F{cache hit?}
    F -- yes --> G[skip rule]
    F -- no --> H[compile / link]
    H --> I[store result in cache]
    G & I --> J[app.exe]
```

The generated `build/` directory is safe to delete at any time — it is never part of your source tree.

---

## Next steps

- Add a static library dependency → [USERGUIDE.md § Static libraries](USERGUIDE.md#static-libraries)
- Customise include paths, defines, and C++ standard → [USERGUIDE.md § Compiler config](USERGUIDE.md#compiler-config)
- Multi-module workspace layout → [USERGUIDE.md § Multi-module projects](USERGUIDE.md#multi-module-projects)
