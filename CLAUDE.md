# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`archilles` is architecture-as-data enforcement for agent-driven codebases. Humans own the design; agents read and verify it, propose changes to it, and never write it. v0 is a Go CLI that reads a design, extracts the actual dependency graph from a Go repo, diffs the two, renders Mermaid, and fails CI on drift.

`archilles-handoff.md` is the authoritative spec. It carries detail this file deliberately omits (artifact schemas, GitHub App permission matrix, ruleset definitions, the IaC/codegraph roadmap). Read the relevant section before implementing anything in that area — do not infer the design from this summary.

## Current state

Pre-code. The repo contains the handoff and this file. **Nothing in the build order below has been started.** Do not scaffold Go code before goal 1 (below) is written and reviewed.

## Goal 1 — the design comes before the code

Author `archilles/` for this repo **by hand**, in the artifact format the tool will later parse, so the format is pressure-tested on a real codebase before the parser exists:

- `archilles/design.yaml` — every component with `path`, `tags.layer`, explicit allowed edges
- `archilles/principles/` — at minimum `layer` ordering, `forbid-edge` core→adapter, `forbid-edge` anything→`cmd/`
- `archilles/decisions/` — one record per non-obvious choice already made, each citing `source: {type: human, ref: "handoff"}`
- `archilles/resolved.json` — hand-written fold; step 3's `archilles resolve` must later reproduce it byte-for-byte
- `docs/ARCHITECTURE.md` — hand-drawn Mermaid of the intended graph, in the conventions the renderer will emit; becomes the renderer's golden file

**Review the design with James before writing code.**

## Architecture

The whole tool is a pipeline over one normalized graph:

```
design.yaml + principles/ + decisions/ + exceptions/
        │ pkg/design (load, validate, resolve)
        ▼
   resolved.json ──► intended graph ─┐
                                     ├─► pkg/graph diff ──► pkg/render/{json,mermaid}
   repo source ──► pkg/adapter/* ────┘                          │
                   (actual graph)                               └─► exit code / CI
```

Layout (`pkg/` is the product; `cmd/archilles` is a thin cobra wrapper):

| Path | Role |
|---|---|
| `cmd/archilles` | cobra entrypoint only — no logic |
| `pkg/design` | schema types, loader, validator, resolver (folds principles + exceptions → rules) |
| `pkg/graph` | normalized `Node`/`Edge` model, set ops, diff. No I/O |
| `pkg/adapter` | `Adapter` interface (`Detect`, `Extract`) |
| `pkg/adapter/golang` | `go list -json -deps` extractor + path-prefix layout mapper |
| `pkg/rules` | the four rule kinds and their evaluators |
| `pkg/render/mermaid`, `pkg/render/json` | renderings of one diff struct |
| `pkg/analyzer` | `go/analysis` analyzer — compile-gate layer 2 |
| `pkg/ledger` | append-only decision/principle/exception records, fold |
| `schema/` | JSON Schema for every on-disk artifact |
| `testdata/` | fixture repos (tiny Go modules) + golden outputs |

### Dependency rules for this repo's own code

Enforced by this repo's own design from commit one — a PR that breaks them fails `archilles-self` before review:

- Layer order is `entry → adapter → core`; edges go downward only.
- Adapters depend on core. **Core never depends on an adapter.** `pkg/design`, `pkg/graph`, `pkg/rules` must not import `pkg/adapter/*`.
- Nothing imports `cmd/`.
- Edges are **default-deny** — an edge not listed in `design.yaml` is a violation, not an omission.
- **Language-aware code lives only in adapters.** Nothing in `pkg/design`, `pkg/graph`, or `pkg/rules` may know about Go, HCL, or any language.

### Invariants that constrain every change

- **Design is data, never code.** No DSL, no Rego, no plugin scripting in v0. Rule kinds are a fixed enum with typed params.
- **Exactly four principle kinds in v0**: `allow-edge`, `forbid-edge`, `layer`, `ownership`. Do not add a fifth — record the one you wanted in `docs/backlog.md`.
- Conflict resolution: specific scope beats broad; forbid beats allow; a tie is an **error at resolve time**, never a silent pick.
- **Deterministic.** Same design + same repo = byte-identical output. No timestamps anywhere except the ledger.
- **Zero network in the CLI.** Design comes from disk.
- **The graph model is multi-level from commit one** (Component / Package / File / Symbol) even though v0 only extracts and diffs at Component level. `graph.Collapse(g, Level)` folds finer to coarser via `Parent`. Do not model packages and retrofit symbols later.
- **Build `json` output first.** `text` and `mermaid` are renderings of the same struct; `json` is the CI contract and is schema-defined in `schema/`.
- Repo mode (single-repo vs design-repo vs component-repo) is a **loader concern**, not a separate command set. Same code paths.

### Borrow before building

`archilles-handoff.md` § "Borrow first" has the table. The rule: if a dependency covers >70% of a need, take it and note the gap in `docs/backlog.md`. The moat is the design model and the diff, not a bespoke graph library. Concretely — `go list`/`x/tools/go/packages` for extraction, `santhosh-tekuri/jsonschema/v5` + `goccy/go-yaml` for validation, cobra+viper for CLI, goreleaser + release-please for release.

Most important: **v0 is not the import-rule enforcer.** `archilles resolve --emit go-arch-lint` generates `.go-arch-lint.yml` from `resolved.json`, and golangci-lint enforces it from day one. `archilles diff` then validates the generated config still matches (`config-drift`) and adds what linters cannot express (ownership, missing/undeclared components, dead edges). Ship that generator in build step 3, not step 6.

## Build order

Follow it. Do not start step 6 before 1–5 have tests.

0. Repo scaffold: `go.mod`, cobra skeleton printing version, golangci-lint config with a **hand-written** `.go-arch-lint.yml` derived from goal 1, Renovate, release-please, goreleaser, `.github/rulesets/*.json`, CODEOWNERS, CI with checks 1–3 and the `validate`-only form of check 4.
1. `pkg/design` types + JSON Schema + loader + validator; fixture tests.
2. `pkg/graph` model + diff; table-driven, no I/O.
3. `pkg/rules` four kinds + resolver + conflict detection; must-pass/must-fail fixture graphs per kind.
4. `pkg/adapter/golang`, tested against `testdata/repos/simple` and `testdata/repos/violating`.
5. `pkg/render/mermaid`, `pkg/render/json`.
6. `cmd/archilles` wiring, `--gate`, `cmd/archilles-vet` (`singlechecker`), `cmd/archilles-toolexec`. `archilles init` last.
7. Dogfood: complete check 4; confirm `resolve` reproduces the hand fold byte-for-byte and the generated go-arch-lint config matches the hand-written one.

**Out of scope for v0** — do not touch: multi-repo/federation, `archilles.lock`, GitHub App wiring, PR bot, exception workflow, any second language adapter, any rule kind beyond the four.

## Commands

Standard Go toolchain (go 1.26 installed):

```bash
go build ./...
go vet ./...
go test -race ./...
go test -race ./pkg/design -run TestResolve          # single package / single test
go test ./pkg/render/mermaid -update                 # regenerate golden files (once the flag exists)
golangci-lint run
go generate ./...                                    # runs `archilles diff --gate`
```

The tool on itself (dogfooding — every one of these runs in CI):

```bash
archilles validate                       # schema-check every artifact
archilles resolve [--write]              # fold → resolved.json; without --write, diff vs checked-in
archilles resolve --emit go-arch-lint    # generate .go-arch-lint.yml
archilles extract --out json             # actual graph only
archilles diff [--out text|json|mermaid] # default text; json is the CI contract
archilles diff --gate                    # on pass, write internal/archgate/gate_gen.go; on fail, delete it
archilles graph --overlay                # Mermaid to stdout
archilles test                           # run principle fixtures in testdata/
```

Diff categories and their severity: `missing-component`, `undeclared-component`, `forbidden-edge`, `expired-exception`, `resolved-drift` fail; `dead-edge` is informational. Exit 0 only when no failing category is present.

## Compile gate

Three layers, all shipping in v0. Layer 1 is friction and fast feedback; layers 2 and 3 are the ones that matter.

1. **Generated gate file.** `internal/archgate/gate_gen.go` holds `DesignHash`/`ActualHash` consts. Every `main` has `var _ = archgate.DesignHash`. The file is **gitignored**, written by `archilles diff --gate` on pass and deleted on fail — so a fresh clone does not compile until the diff passes once.
2. **Analyzer** (`pkg/analyzer`, via `go/analysis`). Loads `resolved.json`, maps the package under analysis to its component, reports non-allowed imports at the import line. Runs under `go vet -vettool=$(which archilles-vet)` and as a golangci-lint module plugin — so plain `go test ./...` fails on a forbidden import.
3. **`-toolexec`** (`cmd/archilles-toolexec`, `GOFLAGS` in `go.env` for CI). Intercepts `compile`, rejects before the compiler runs.

A local actor with full control can bypass a local gate. The gate's job is to make the fastest path the compliant one and make bypass explicit and loggable. Authority lives in CI.

## Working in this repo

**Commit trailers: never add `Co-Authored-By: Claude` or any Claude/Anthropic attribution to commits or PR bodies.** This is a standing instruction from the repo owner.

- Trunk-based. `main` is protected: no direct push, linear history, squash merge only. Branch from `main`, PR back.
- Branch names are enforced by ruleset: `feat/*`, `fix/*`, `chore/*` for implementation; `design/*` for changes to `archilles/**`; anything else is rejected at push.
- Conventional commits on the squash title — release-please turns them into changelog and tag.
- **Paths agents must not write** (enforced by ruleset, not just instruction): `archilles/**`, `schema/**`, `.github/**`. A design change is a proposal on a `design/*` branch for James to approve, never an edit made in passing to make a check go green. If `archilles diff` fails, fix the code — do not edit the design to permit the violation.
- `CODEOWNERS`: `archilles/**` and `schema/**` → James.
- Required PR checks: `build`, `lint`, `test`, `archilles-self`, `schema`, `fixtures`. `archilles-self` is the point of the repo and is never skipped — it runs whatever subset of the tool exists and grows as the build order advances.
