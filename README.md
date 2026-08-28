# repokit

**A small Go repository and development tooling orchestrator.**

`repokit` helps you work with multiple independent services while keeping a
consistent development style — without turning everything into a monorepo and
without prescribing a particular stack.

You choose which repositories to include and which tooling to use.
Everything is explicitly pinned to a branch or tag.

[Русская версия](README.ru.md) · [Technical documentation](CONTRIBUTING.md)

---

## Problems it solves

| Problem                                          | How repokit helps                                                                                                         |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------- |
| Multiple services, each in its own repository    | Workspace mode clones services, validates presence, and generates workspace files (`go.work`, `pnpm-workspace.yaml`, ...) |
| Want a consistent style (linter, hooks, configs) | Tooling is pulled from a separate repository and copied by explicit rules                                                 |
| Configs get duplicated and drift apart           | Single source of truth + explicit versioning via `ref` / `ref_type`                                                       |
| IDE only sees configs in the root                | Files are copied to the right place while preserving directory structure                                                  |
| Need to quickly run commands across services     | Target proxy: `repokit catalog/lint`, `repokit users/test ./...`                                                          |
| Want to work on just one service                 | Standalone mode — no workspace files, only tooling                                                                        |
| Re-running setup is scary                        | `sync` is idempotent — safe to run again and again                                                                        |

---

## Quick start

```bash
# Install
go install github.com/destyk/repokit@latest

# Or build from source
make setup
make build && make install

# Create a starter config
repokit init --mode standalone   # or: --mode workspace

# Edit .repokit.yml, then:
repokit sync
repokit status
```

Configuration examples: `examples/standalone/`, `examples/workspace/`.

---

## Two modes

### Standalone

A single repository.

Pulls shared tooling and applies file mappings. Does not create workspace files.

### Workspace

Multiple services.

Clones/synchronizes them, applies shared tooling, and regenerates workspace
files (`go.work`, `pnpm-workspace.yaml`, ...) when generators match.

---

## Main commands

```bash
repokit init [--mode standalone|workspace] [--force]
repokit sync
repokit status
repokit version
repokit <service>/<target> [args...]   # e.g. catalog/lint
```

---

## Supported runners

Each workspace service declares a **runner** in `.repokit.yml` — the tool used
when you call `repokit <service>/<target>`:

```yaml
repositories:
  - name: catalog
    url: git@github.com:acme/catalog.git
    ref: v1.8.0
    ref_type: tag
    runner: make # <- here
```

Built-in runners:

| Runner   | Invocation style              | Example                  |
| -------- | ----------------------------- | ------------------------ |
| `make`   | `make <target> [args…]`       | `repokit catalog/lint`   |
| `go`     | `go <target> [args…]`         | `repokit api/test ./...` |
| `just`   | `just <target> [args…]`       |                          |
| `task`   | `task <target> [args…]`       |                          |
| `npm`    | `npm run <target> [args…]`    | `repokit web/build`      |
| `pnpm`   | `pnpm <target> [args…]`       |                          |
| `yarn`   | `yarn <target> [args…]`       |                          |
| `bun`    | `bun run <target> [args…]`    |                          |
| `cargo`  | `cargo <target> [args…]`      |                          |
| `uv`     | `uv run <target> [args…]`     |                          |
| `poetry` | `poetry run <target> [args…]` |                          |
| `mise`   | `mise run <target> [args…]`   |                          |

`runner` is **required** for every repository entry. Adding a new runner is
described in the [technical docs](CONTRIBUTING.md#adding-a-runner).

## Why this design

- **Explicit pins** — every external repo is locked to a tag or branch; no floating `latest`.
- **Your tooling, not ours** — the tooling repository is yours; repokit only copies what you declare.
- **CLI vs library** — console output lives in the CLI; domain packages return structured results for programmatic use.
- **Re-run safely** — `sync` converges to the config state without destructive surprises.

---

## Contributing

Architecture, public adapters, config loading, runners, and library examples
are documented in **[CONTRIBUTING.md](CONTRIBUTING.md)**.

Start there if you want to extend repokit or embed it in your own tools.

---

## License

See [LICENSE](LICENSE).
