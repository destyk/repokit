# repokit — technical documentation

This document is for contributors and anyone embedding repokit as a library.
For a product overview and “why use this”, see the [root README](../README.md).

[Русская версия](README.ru.md)

---

## Architecture

Dependency direction is inward — core does not depend on adapters.

| Layer       | Packages                                                                                              | Role                              |
| ----------- | ----------------------------------------------------------------------------------------------------- | --------------------------------- |
| Core        | `internal/repository`, `internal/tooling`, `internal/workspace`, `internal/runner`, `internal/status` | Domain logic                      |
| Application | `cli`, `internal/config`, `internal/hooks`, `internal/command`                                        | CLI, config, hooks, exec          |
| Adapters    | `repository`, `tooling`, `workspace`, `runner`, `policy`                                              | Public API for external importers |
| Entrypoint  | `main.go`                                                                                             | `main` only                       |

- **Core never imports** public packages or `cli`.
- **Public packages** are thin re-exports (type aliases + forwarding functions)
  plus small policy helpers.
- **`cli` lives outside `internal`** — it is an application adapter, not library core.
- **Config is internal only** — load `.repokit.yml` from the CLI; programmatic
  callers compose `policy` / domain specs instead.
- **Console output belongs to `cli`** — internal packages return structured
  results (`Result`, `Status`, `AppliedFile`, …).

---

## Configuration (`.repokit.yml`)

Loaded via `internal/config.FindAndLoad` (walk-up from cwd).

```yaml
version: 2
mode: standalone # or workspace

tooling:
  repository: git@github.com:your-org/shared-tooling.git
  ref: v1.2.0
  ref_type: tag
  files:
    - source: golangci/.golangci.yml
      destination: .golangci.yml
      required: true
    - source: configs/**/*.yaml
      destination: .
      required: false

hooks:
  post_sync:
    - command: make
      args: [generate]
      required: false
```

Workspace mode adds `repositories[]`:

```yaml
mode: workspace
repositories:
  - name: catalog
    url: git@github.com:acme/catalog.git
    ref: v1.8.0
    ref_type: tag
    runner: make
    required: true
  - name: users
    url: git@github.com:acme/users.git
    ref: main
    ref_type: branch
    runner: pnpm
    path: services/users
```

Full examples: [`examples/standalone/`](../examples/standalone/),
[`examples/workspace/`](../examples/workspace/).

Validation is uniform: every block implements `Validate()` with field-path
error prefixes (`tooling:`, `repositories[0]:`, ...).

---

## CLI behaviour

| Command              | Role                                                                                  |
| -------------------- | ------------------------------------------------------------------------------------- |
| `init`               | Writes a starter `.repokit.yml` from embedded templates (CLI-only; no domain package) |
| `sync`               | Idempotent full sync: tooling → apply → services → generators → hooks                 |
| `status`             | Formats `internal/status.Collect` (pure data)                                         |
| `<service>/<target>` | Looks up runner, streams child IO to the terminal                                     |
| `version` / `help`   | Meta                                                                                  |

Exit code `0` = success; non-zero = error (message on stderr).

---

## Library usage

Import only public domain packages:

```go
import (
    "github.com/destyk/repokit/policy"
    "github.com/destyk/repokit/repository"
    "github.com/destyk/repokit/runner"
    "github.com/destyk/repokit/tooling"
    "github.com/destyk/repokit/workspace"
)
```

Do **not** import `internal/...`.

### Tooling + repository

```go
ctx := context.Background()
root := "/path/to/project"

tools := policy.Standalone(
    "git@github.com:your-org/shared-tooling.git",
    "v1.2.0",
    policy.File("golangci/.golangci.yml", ".golangci.yml", true),
)

if err := tooling.Sync(ctx, root, tools); err != nil {
    panic(err)
}
applied, err := tooling.Apply(root, tools)
if err != nil {
    panic(err)
}
_ = applied

svc := policy.RequiredService(
    "catalog",
    "git@github.com:your-org/catalog.git",
    "v1.8.0",
    "make",
)
if err := repository.Sync(ctx, root, svc); err != nil {
    panic(err)
}
```

### Policy helpers

```go
tools := policy.WorkspaceTooling(
    "git@github.com:your-org/shared-tooling.git",
    "v1.2.0",
    policy.File("golangci/.golangci.yml", ".golangci.yml", true),
    policy.File("lefthook/lefthook.yml", "lefthook.yml", false),
)

services := []repository.Spec{
    policy.RequiredService("catalog", "git@github.com:acme/catalog.git", "v1.8.0", "make"),
    policy.Branch(policy.Service("users", "git@github.com:acme/users.git", "main", "pnpm")),
}
```

### Runners: stream vs capture

```go
// Capture (programmatic)
res, err := runner.Run(ctx, "make", dir, "test", nil, runner.Options{})
// res.Stdout, res.Stderr, res.ExitCode

// Stream (CLI-style)
_, err = runner.Run(ctx, "make", dir, "lint", nil, runner.Options{
    Stdout: os.Stdout,
    Stderr: os.Stderr,
})
```

---

## Idempotent `sync`

Re-running `repokit sync` is safe:

1. Tooling cache is fetched/checked out to the pinned ref
2. File mappings are applied with atomic overwrite
3. Services are cloned if missing, otherwise fetched + checked out to the pin
4. Generators rewrite workspace files
5. Post-sync hooks run as ordinary commands

Optional repositories / optional tooling files / optional hooks do not fail the
whole run when missing.

---

## Development

```bash
make setup
make test
make lint
gofmt -w .
make build
```

Pull requests should keep the dependency rule: `internal` no adapters, no `cli`.
