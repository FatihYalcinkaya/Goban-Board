# go-kanban — Improvement & Feature Proposal

**Prepared for:** Engineering Management  
**Date:** May 9, 2026  
**Author:** Engineering Analysis  
**Status:** Draft for Review

---

## Executive Summary

`go-kanban` is a well-written terminal-based Kanban board implemented in Go using the Bubble Tea framework (~550 LOC, 7 source files). It delivers a functional MVP: task CRUD, column navigation, keyboard-driven reordering, and SQLite persistence. The code is clean and idiomatic.

However, the project has the typical gaps of a solo side project — **zero tests, no CI/CD, no stable task identity, unchecked panics, and several rough edges** that would surface under real use. Below is a prioritized proposal to harden the codebase and extend its feature set.

---

## Phase 1: Hardening & Bug Fixes (Critical — Do First)

These are issues that will cause data loss or crashes in production use.

| # | Issue | Severity | Effort |
|---|-------|----------|--------|
| 1 | **Duplicate title bug** — DB uses `title` as lookup key. Two tasks with the same name cause wrong DELETE/UPDATE/RENAME. The DB schema already has `id INTEGER PRIMARY KEY AUTOINCREMENT` but the Go code never reads or uses it. | **High** | 2–3h |
| 2 | **Unsafe type assertions** — `selectedItem.(Task)` on 4 code paths in `model.go` will panic if a non-`Task` item ever enters the list. | **High** | 1h |
| 3 | **No error surfacing** — All DB errors are `log.Printf`'d and silently swallowed. The user never knows when a save/delete/rename fails. | **Medium** | 2h |
| 4 | **Rename data-loss race** — Removing and re-adding the task during rename can lose the task if something fails mid-operation. | **Medium** | 2h |
| 5 | **No `.gitignore`** — `tasks.db` is a runtime artifact that will be tracked by git. | **Low** | 5min |

### Recommended Actions

1. Refactor all DB operations to use `id` (already in schema) instead of `title`. Pass `ID` through the `Task` struct and Bubble Tea messages.
2. Replace `task := selectedItem.(Task)` with `task, ok := selectedItem.(Task); if !ok { ... }`.
3. Introduce a lightweight error state in the model (e.g., `errMsg string` shown in status bar, cleared on next keypress).
4. Rewrite rename flow: update in-place in the list + DB; don't delete-and-reinsert.

---

## Phase 2: Testing Infrastructure (High Priority)

There are **zero test files** in the entire project. This is the single biggest quality gap.

| Area | Suggested Approach | Effort |
|------|--------------------|--------|
| DB layer | Table-driven tests with `:memory:` SQLite database | 4h |
| Model/Update | Bubble Tea's built-in `tea.NewProgram(model, tea.WithInput(nil))` for simulating keypresses | 6h |
| View rendering | Golden file / snapshot tests for View() output | 3h |
| Task & Column creation | Simple unit tests | 1h |

### Recommended Actions

1. Add `*_test.go` files for `db.go` (easiest, highest ROI).
2. Add model update tests covering: add task, delete task, rename task, move between columns, reorder within column.
3. Run tests in CI (see Phase 4).

---

## Phase 3: Code Quality & Architecture (Medium Priority)

These improvements make the codebase maintainable and safer.

| # | Improvement | Effort |
|---|-------------|--------|
| 1 | **Introduce internal packages** — split into `internal/db`, `internal/column`, `internal/task`, `internal/tui`. | 3h |
| 2 | **Remove global `db` variable** — inject DB connection via constructor / dependency injection. | 2h |
| 3 | **Cache styles** — `View()` recreates Lip Gloss styles every render cycle. Pre-build once in `RootModel`. | 1h |
| 4 | **Remove dead code** — `fixedColumnWidth` constant in `styles.go` is never used. | 5min |
| 5 | **Enable SQLite WAL mode** — prevents DB corruption on crash during write. | 15min |

### Recommended Actions

1. Refactor `package main` → structured packages. Use interfaces for DB so the TUI layer can be tested without real SQLite.
2. Remove global `var db *sql.DB`; pass `*sql.DB` or a `Database` interface to `NewRootModel()`.

---

## Phase 4: CI/CD & Developer Experience (Medium Priority)

| # | Item | Effort |
|---|------|--------|
| 1 | **GitHub Actions CI** — run `go build`, `go vet`, `go test ./...` on every push/PR. | 2h |
| 2 | **Makefile** — targets for `build`, `test`, `lint`, `clean`. | 1h |
| 3 | **Linting** — add `golangci-lint` with standard config. | 1h |
| 4 | **Pre-commit hook** — run tests + vet before commit. | 30min |

### Recommended Actions

1. Create `.github/workflows/go.yml` with Go 1.26, `go build`, `go vet`, `go test`.
2. Add `Makefile` with the common targets.
3. Add `.gitignore` for `tasks.db` and build artifacts.

---

## Phase 5: Feature Enhancements (Low Priority — Value Dependent)

These add genuine user-facing value but require the Phase 1–3 foundations first.

| Feature | Effort | Value |
|---------|--------|-------|
| **Column rename** — press `R` on a column header to rename it | 4h | High |
| **Column delete** — press `D` on a column to remove it (with confirmation) | 3h | High |
| **Delete confirmation** — confirm before permanently deleting a task | 2h | Medium |
| **Task descriptions** — set a description field; view in an expanded pane or tooltip | 6h | Medium |
| **Search/filter** — re-enable `list.Model` filtering (`/` to search) | 2h | Medium |
| **Undo** — undo last delete (hold deleted items in an undo buffer) | 4h | High |
| **Auto-increment column name** — "New Column", "New Column 2", etc. | 1h | Low |
| **Keyboard shortcut help modal** — `?` key shows a cheat sheet | 3h | Low |

### Recommended Actions

Pick the top 2–3 features based on user feedback. Column rename and delete confirmation are the most impactful relative to effort.

---

## Effort Summary

| Phase | Theme | Total Effort | Impact |
|-------|-------|-------------|--------|
| 1 | Bug fixes & hardening | ~8h | Prevents crashes & data loss |
| 2 | Testing | ~14h | Prevents regressions |
| 3 | Code quality | ~7h | Maintainability |
| 4 | CI/CD & DX | ~5h | Developer velocity |
| 5 | Feature work | ~25h | User-facing value |

**Total estimated effort: ~59 hours** (spread across phases)

---

## Recommendation

1. **Immediately** fix Phase 1 bugs (duplicate title, unsafe assertions) — these are ticking time bombs.
2. **Next sprint** establish the testing foundation (Phase 2) + CI (Phase 4). Without tests, every future change is risky.
3. **Ongoing** Phase 3 and Phase 5 work can be interleaved as capacity allows.

The project has solid bones. With ~40 hours of hardening and ~20 hours of feature work, it goes from a personal prototype to a tool I'd be comfortable shipping.
