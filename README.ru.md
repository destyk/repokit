# repokit

**Небольшой оркестратор репозиториев и development tooling на Go.**

`repokit` помогает работать с несколькими независимыми сервисами, сохраняя
единый стиль разработки — без превращения всего в monorepo и без навязывания
стека.

Вы сами выбираете, какие репозитории включать и какой tooling использовать.
Всё явно закрепляется веткой или тегом.

[English version](README.md) · [Техническая документация](CONTRIBUTING.ru.md)

---

## Какие проблемы решает

| Проблема                                       | Как помогает repokit                                                                                   |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| Несколько сервисов, каждый в своём репозитории | Режим workspace клонирует сервисы и генерирует workspace-файлы (`go.work`, `pnpm-workspace.yaml`, ...) |
| Нужен единый стиль (линтер, хуки, конфиги)     | Tooling подтягивается из отдельного репозитория и копируется по явным правилам                         |
| Конфиги дублируются и разъезжаются             | Один источник правды + явное версионирование через `ref` / `ref_type`                                  |
| IDE видит конфиги только в корне               | Файлы копируются куда нужно с сохранением структуры каталогов                                          |
| Нужно быстро гонять команды по сервисам        | Прокси целей: `repokit catalog/lint`, `repokit users/test ./...`                                       |
| Хочется работать только с одним сервисом       | Режим standalone — без workspace-файлов, только tooling                                                |
| Боязнь повторного setup                        | `sync` идемпотентен — можно запускать снова и снова                                                    |

---

## Быстрый старт

```bash
# Установка
go install github.com/destyk/repokit@latest

# Или сборка из исходников
make setup
make build && make install

# Стартовый конфиг
repokit init --mode standalone   # или: --mode workspace

# Отредактируйте .repokit.yml, затем:
repokit sync
repokit status
```

Примеры конфигурации: `examples/standalone/`, `examples/workspace/`.

---

## Два режима

### Standalone

Один репозиторий.

Подтягивает shared tooling и применяет file mappings. Workspace-файлы не создаёт.

### Workspace

Несколько сервисов.

Клонирует/синхронизирует их, применяет shared tooling и перегенерирует
workspace-файлы (`go.work`, `pnpm-workspace.yaml`, ...), когда срабатывают generators.

---

## Основные команды

```bash
repokit init [--mode standalone|workspace] [--force]
repokit sync
repokit status
repokit version
repokit <service>/<target> [args...]   # например catalog/lint
```

---

## Поддерживаемые runners

У каждого сервиса в workspace в `.repokit.yml` указывается **runner** — инструмент,
которым вызывается `repokit <service>/<target>`:

```yaml
repositories:
  - name: catalog
    url: git@github.com:acme/catalog.git
    ref: v1.8.0
    ref_type: tag
    runner: make # <- здесь
```

Встроенные runners:

| Runner   | Стиль вызова                  | Пример                   |
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

Поле `runner` **обязательно** для каждой записи в `repositories`. Как добавить
новый runner — в [технической документации](CONTRIBUTING.ru.md#добавление-runner).

## Почему так устроено

- **Явные пины** — каждый внешний репозиторий зафиксирован тегом или веткой; никакого плавающего `latest`.
- **Ваш tooling, не наш** — репозиторий tooling ваш, а repokit копирует только то, что вы объявили.
- **CLI vs библиотека** — вывод в консоль живёт в CLI, domain-пакеты возвращают структуры для программного вызова.
- **Безопасный повтор** — `sync` сходится к состоянию из конфига без сюрпризов.

---

## Участие в разработке

Архитектура, публичные адаптеры, загрузка конфига, runners и примеры
библиотечного API — в **[CONTRIBUTING.md](CONTRIBUTING.ru.md)**.

Начните оттуда, если хотите расширять repokit или встраивать его в свои инструменты.

---

## Лицензия

См. [LICENSE](LICENSE).
