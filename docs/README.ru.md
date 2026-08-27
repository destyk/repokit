# repokit — техническая документация

Документ для контрибьюторов и тех, кто встраивает repokit как библиотеку.
Обзор продукта и "зачем это" — в [корневом README](../README.ru.md).

[English version](README.md)

---

## Архитектура

Направление зависимостей внутрь - ядро не зависит от адаптеров.

| Слой        | Пакеты                                                                                                | Роль                               |
| ----------- | ----------------------------------------------------------------------------------------------------- | ---------------------------------- |
| Core        | `internal/repository`, `internal/tooling`, `internal/workspace`, `internal/runner`, `internal/status` | Доменная логика                    |
| Application | `cli`, `internal/config`, `internal/hooks`, `internal/command`                                        | CLI, конфиг, хуки, exec            |
| Adapters    | `repository`, `tooling`, `workspace`, `runner`, `policy`                                              | Публичный API для внешних импортов |
| Entrypoint  | `main.go`                                                                                             | только `main`                      |

- **Core никогда не импортирует** публичные пакеты или `cli`.
- **Публичные пакеты** — тонкие re-export (type aliases + forwarding) и
  небольшие policy-хелперы.
- **`cli` снаружи `internal`** — application adapter, не библиотечное ядро.
- **Config только internal** — `.repokit.yml` читает CLI, программно
  собирайте `policy` / domain specs.
- **Вывод в консоль — зона `cli`** — internal возвращает структуры
  (`Result`, `Status`, `AppliedFile`, ...).

---

## Конфигурация (`.repokit.yml`)

Загрузка через `internal/config.FindAndLoad` (поиск вверх от cwd).

```yaml
version: 2
mode: standalone # или workspace

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

В режиме workspace добавляется `repositories[]`:

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

Полные примеры: [`examples/standalone/`](../examples/standalone/),
[`examples/workspace/`](../examples/workspace/).

Валидация однообразная: у каждого блока `Validate()` с префиксами путей
полей (`tooling:`, `repositories[0]:`, …).

---

## Поведение CLI

| Команда              | Роль                                                                                 |
| -------------------- | ------------------------------------------------------------------------------------ |
| `init`               | Пишет стартовый `.repokit.yml` из embedded templates (только CLI; без domain-пакета) |
| `sync`               | Идемпотентный полный sync: tooling -> apply -> services -> generators -> hooks       |
| `status`             | Форматирует `internal/status.Collect` (чистые данные)                                |
| `<service>/<target>` | Ищет runner, стримит IO дочернего процесса в терминал                                |
| `version` / `help`   | Мета                                                                                 |

Код выхода `0` = успех; ненулевой = ошибка (сообщение в stderr).

---

## Использование как библиотеки

Импортируйте только публичные domain-пакеты:

```go
import (
    "github.com/destyk/repokit/policy"
    "github.com/destyk/repokit/repository"
    "github.com/destyk/repokit/runner"
    "github.com/destyk/repokit/tooling"
    "github.com/destyk/repokit/workspace"
)
```

**Не** импортируйте `internal/...`.

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

### Policy-хелперы

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
// Захват (программно)
res, err := runner.Run(ctx, "make", dir, "test", nil, runner.Options{})
// res.Stdout, res.Stderr, res.ExitCode

// Стрим (как CLI)
_, err = runner.Run(ctx, "make", dir, "lint", nil, runner.Options{
    Stdout: os.Stdout,
    Stderr: os.Stderr,
})
```

---

## Идемпотентный `sync`

Повторный `repokit sync` безопасен:

1. Кэш tooling fetch/checkout на pinned ref
2. File mappings — atomic overwrite
3. Сервисы: clone если нет, иначе fetch + checkout на pin
4. Generators переписывают workspace-файлы
5. Post-sync hooks — обычные команды

Опциональные репозитории / optional tooling files / optional hooks не валят
весь прогон, если недоступны.

---

## Разработка

```bash
make setup
make test
make lint
gofmt -w .
make build
```

В PR соблюдайте правило зависимостей: `internal` без адаптеров и без `cli`.
