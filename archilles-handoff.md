# `archilles` — v0 handoff

Architecture-as-data enforcement for agent-driven codebases. Humans own the design; agents can read and verify it, propose changes to it, and never write it. v0 is a Go CLI that reads a design, extracts the actual dependency graph from a Go repo, diffs the two, renders Mermaid, and fails CI on drift.

Binary and module name: `archilles`.

## Principles (non-negotiable for the codebase itself)

- Library-first. `pkg/` is the product; `cmd/archilles` is a thin cobra wrapper. Every command is a function call away from being an API.
- Design is data, never code. No DSL, no Rego, no plugin scripting in v0. Rule kinds are a fixed enum with typed params.
- Deterministic. Same design + same repo = byte-identical output. No timestamps in outputs except the ledger.
- Language adapters are the only language-aware code. Everything else operates on the normalized graph.
- Zero network in the CLI. Design comes from disk (later: git ref).

## Repo layout

```
cmd/archilles/            cobra entrypoint only
pkg/design/          schema types, loader, validator, resolver (fold principles+exceptions -> rules)
pkg/graph/           normalized Node/Edge model, set ops, diff
pkg/adapter/         Adapter interface
pkg/adapter/golang/  go list -json -deps based extractor + layout mapper
pkg/rules/           rule kinds and evaluators (operate on graph + resolved design)
pkg/render/mermaid/  Mermaid output
pkg/render/json/     machine output for CI / PR comments
pkg/ledger/          decision/principle/exception records, append-only, fold
schema/              JSON Schema for every on-disk artifact
testdata/            fixture repos (tiny Go modules) + expected outputs
.github/workflows/   ci.yml (lint, test, self-check: archilles diff on this repo)
```

`archilles` must dogfood: this repo carries its own `archilles/design.yaml` and `archilles diff` runs in its own CI from commit one.

## Repo modes

Two modes, one tool, detected by what's on disk.

**Design repo** (org-level, `infrastructure-live`-shaped). Holds the ledger and the tree; owns no application code.

```
org.yaml                     org-wide principles, tags vocabulary
<domain>/
  domain.yaml
  <system>/
    system.yaml
    <component>/
      component.yaml         repo, path, ref, tags; may hold component-local principles
principles/  decisions/  exceptions/     flat ledger, scoped by path
resolved/<domain>/<system>/<component>.json   derived, checked in
```

`archilles resolve <path>` folds the include chain (org → domain → system → component) plus scoped ledger records into one flat resolved design for that component. Child may narrow, never loosen; loosening is an exception record.

**Component repo** (a single project). Implements one or more components; owns no design authority.

```
archilles/
  archilles.lock             design repo URL + pinned ref + component IDs this repo implements
  local.yaml                 optional: sub-component layout *inside* the owned components (packages, symbol placement), never edges to other components
  actual.json                derived on CI: extracted graph at the finest supported level; published as a build artifact
```

Component repo output contract (`archilles diff --out json`, schema `schema/diff.json`):

```json
{
  "design": {"repo": "...", "ref": "abc123", "resolved_hash": "..."},
  "components": ["payments-ledger"],
  "level": "component",
  "violations": [{"category": "forbidden-edge", "from": "...", "to": "...", "evidence": [{"file": "...", "line": 14}], "rule": "P0007"}],
  "informational": [...],
  "actual": {"artifact": "archilles/actual.json", "hash": "..."}
}
```

`actual.json` is the federation input: the design repo's bot stitches every component repo's published `actual.json` at its pinned ref into the org-wide actual graph and diffs it against the org-wide intended one. That's where cross-repo violations surface (two repos claiming a component, an edge that exists in neither, a repo touching a component it doesn't own).

Single-repo v0 collapses both modes into one directory: the repo is its own design repo with a one-level tree. `archilles.lock` is absent; ref is `HEAD`. The code paths must be the same — mode is a loader concern, not a separate command set.

## Where the code lives

Every component record answers it explicitly:

```yaml
id: payments-ledger
repo: git@github.com:org/payments.git    # omit in single-repo
path: internal/ledger                    # relative to repo root
ref: main                                # branch or tag the design expects; CI records the commit it validated
lang: go                                 # selects the adapter
tags: {layer: core, domain: payments, criticality: high}
```

Ownership is one repo per component. Same component in two repos is a violation; a repo with files matching no owned component reports `unassigned`.

## IaC module (goal, not v0)

Terraform/Terragrunt is a target adapter, not a first-class concern of the core. Nothing in `pkg/design`, `pkg/graph`, or `pkg/rules` may know about HCL. The adapter interface is fixed in v0; this adapter lands in v1 after Go is solid. Sketch so the model isn't designed against it:

- **Nodes**: Terragrunt units (each `terragrunt.hcl` dir) at Component level; Terraform modules at Package level; resources at Symbol level (`kind: resource`, tagged with provider type).
- **Edges**: `dependency` / `dependencies` blocks → `depends`; module `source` refs → `import`; remote-state reads → `depends` with evidence. Provider-level implicit dependencies (an SG referenced by a subnet) are Symbol-level `depends` from `terraform graph`/plan JSON.
- **Placement**: a Terragrunt live tree *is* a placement structure. `placement` rules express "RDS lives under `<env>/data/`", "nothing in `<env>/app/` sources a module outside `modules/` or the approved registry".
- **Edges to app components**: an app component that reads an SSM parameter or queue owned by an IaC unit is a cross-adapter edge. Model as a `contract` edge with the parameter/queue ARN pattern as the interface; enforce presence in both graphs.
- **Extraction**: `terragrunt graph-dependencies` / `run-all plan -json` where available; parse HCL directly (`hashicorp/hcl/v2`) for the static case so CI doesn't need cloud creds. Static is default; plan-backed is `--with-plan`.

Where it lives: IaC repos are component repos like any other, with `lang: terragrunt` in the component record. The design tree for infra mirrors the live tree (same org → domain → system shape) — that mapping is the reason the design repo is `infrastructure-live`-shaped in the first place. tgviz's operator/DAG work is the natural place to run this adapter continuously instead of per-CI.

## On-disk artifacts (all schema-validated)

```
archilles/
  design.yaml          components, tags, ownership; the human-authored intent
  principles/*.yaml    scoped rules (see kinds)
  decisions/*.yaml     one-off decisions, cite principle or carry context
  exceptions/*.yaml    rule + scope + reason + requested_by + approver + expires
  resolved.json        derived: fold of all of the above; checked in; CI verifies it matches
  archilles.lock            (multi-repo, later) pinned design ref
```

### design.yaml

```yaml
version: 1
components:
  - id: cli
    path: cmd/archilles
    tags: {layer: entry}
  - id: design
    path: pkg/design
    tags: {layer: core}
  - id: adapter-go
    path: pkg/adapter/golang
    tags: {layer: adapter, lang: go}
edges:                       # explicitly allowed; default deny
  - {from: cli, to: design}
  - {from: adapter-go, to: graph}
```

A file belongs to the component whose `path` is its longest prefix match. Files matching no component are `unassigned` and reported.

### Principle kinds (v0 — exactly these four)

| kind | params | semantics |
|---|---|---|
| `allow-edge` | from, to (selectors) | permits edges matching; edges are default-deny |
| `forbid-edge` | from, to | denies even if an allow matches (forbid wins) |
| `layer` | order: [entry, adapter, core] | edges may only go downward in the order |
| `ownership` | selector, repo | component must live in this repo (no-op single-repo, validate the field) |

Selectors: `{id: x}` or `{tags: {layer: core}}`. Specific scope beats broad; forbid beats allow; ties are an error at resolve time, not a silent pick.

Do not add kinds in v0. Note the ones you wanted in `docs/backlog.md`. v0 enforcement of these four is delegated to generated go-arch-lint/depguard config where they can express it (see Borrow first); `archilles diff` covers the remainder and checks config drift.

### Records

```yaml
# principles/0001-layering.yaml
id: P0001
kind: layer
params: {order: [entry, adapter, core]}
scope: {tags: {}}            # whole repo
status: accepted             # proposed | accepted | superseded | folded-into:<id>
source: {type: human, ref: "initial design"}
author: james
approver: james
created: 2026-09-15
```

Exceptions carry `expires`. Expired exceptions are ignored by the resolver and reported by `archilles diff`.

## Normalized graph

```go
type Level int // Component, Package, File, Symbol
type Node struct {
    ID     string
    Level  Level
    Kind   string            // component | package | file | type | func | method | module | resource
    Parent string            // containment: symbol -> file -> package -> component
    Path   string
    Tags   map[string]string
}
type Edge struct { From, To string; Kind string; Evidence []Ref } // Kind: import | call | embed | depends
type Graph struct { Nodes map[string]Node; Edges []Edge }
```

Intended graph is built from `resolved.json`. Actual graph is built by the adapter. Diff operates only on these.

The model is multi-level from commit one even though v0 only *extracts* and *diffs* at Component level. `graph.Collapse(g, Level)` folds any finer graph up to a coarser one by walking `Parent`. Do not design the model around packages and retrofit symbols — that's the rewrite this section exists to prevent.

## Codegraph (goal, phased)

The interface graph is not the end state. The goal is a code graph: where code lives, where types live, where logic lives, and rules over that — not just over component edges.

| Level | Nodes | Edges | Extracted by | Phase |
|---|---|---|---|---|
| Component | design components | allowed/actual deps | path prefix | v0 |
| Package | Go packages | imports | `go list` | v0 (collapsed to component for diff) |
| File | files | imports | `go list` | v1 |
| Symbol | types, funcs, methods | calls, embeds, implements | `go/packages` + `go/types`, call graph via `golang.org/x/tools/go/callgraph` (static or CHA) | v1 |

Design stays one level above what it constrains. It never enumerates classes; it states placement and shape rules over symbol *kinds and tags*:

- `placement`: symbols matching a selector must live under a component/path — "HTTP handlers live in `adapter/http`", "anything implementing `Repository` lives in `adapter/*`", "no `func` over N lines in `core/domain`" is a lint, not this.
- `logic-boundary`: calls from tagged-`handler` symbols must go to `core` only through an interface declared in `core` (ports-and-adapters as a checkable rule).
- `interface-home`: an interface and its implementations may not share a component.

Symbol tagging comes from: path (component tags inherited), Go naming/signature heuristics in the adapter (handler = `func(http.ResponseWriter, *http.Request)` / echo/gin/chi signatures), and explicit `//archilles:tag=handler` directives as an override. Heuristics are adapter-owned and versioned; a heuristic change that alters tags is a diff, not a silent update.

`archilles graph --level symbol` renders the code graph. Mermaid will not survive a full symbol graph; at that level emit Mermaid only for a `--focus <component>` slice and DOT/JSON for the whole thing. Full-graph UI is the ReactFlow/tgviz-shaped frontend, not v0.

This is where "the orchestrator is in on it" gets caught for real: a component edge can look clean while logic quietly moved into a handler. Symbol-level `logic-boundary` is the rule that catches that.

## Adapter interface

```go
type Adapter interface {
    Detect(root string) (bool, error)
    Extract(root string, design *design.Resolved) (*graph.Graph, error)
}
```

Go adapter: `go list -json -deps ./...`, map packages to components by path prefix, collapse package-level imports to component-level edges, keep `file:line` evidence for every edge. Ignore stdlib and external modules by default (`--include-external` flag later).

## Diff categories (output of `archilles diff`)

- `missing-component`: in design, no files
- `undeclared-component`: files in `unassigned`
- `forbidden-edge`: present in actual, not allowed by resolved rules (with evidence)
- `dead-edge`: allowed in design, absent in code (informational, not failing)
- `expired-exception`: exception past `expires` still referenced
- `resolved-drift`: checked-in `resolved.json` != fold of records

Exit 0 only when no failing categories. `--fail-on` overrides the set.

## Commands (v0)

```
archilles init                      scaffold archilles/ from the current repo (proposes components from top-level packages; user edits)
archilles validate                  schema-check every artifact
archilles resolve [--write]         fold -> resolved.json; without --write, diff against checked-in
archilles extract [--out json]      actual graph only
archilles diff [--out text|json|mermaid]          default text
archilles graph [--intended|--actual|--overlay]   Mermaid to stdout, default --overlay
archilles test                      run principle fixtures in testdata/
```

Output defaults: `text` for humans on a terminal, `json` is the contract for CI and any future UI (schema in `schema/`), `mermaid` is the only visual format in v0 — free tier lives on it. Build `json` first; `text` and `mermaid` are renderings of the same struct.

Mermaid overlay conventions: allowed+present solid, forbidden `-.->` with `%% VIOLATION` comment and a `:::violation` class, dead-edge dotted gray, missing component dashed box. Keep it readable in a GitHub PR comment; no subgraph nesting beyond tags.layer.

## Compile gate

`go build` runs no hooks, so "doesn't compile unless it passes" has to be manufactured. Three layers, from cheapest to hardest to route around. All three ship in v0.

**1. Generated gate file (local, the one developers and agents hit first).**
`archilles diff` on success writes `internal/archgate/gate_gen.go`:

```go
// Code generated by archilles. DO NOT EDIT.
package archgate
const DesignHash = "sha256:…"   // resolved.json hash
const ActualHash = "sha256:…"   // extracted graph hash at time of pass
```

On failure it deletes the file. `cmd/archilles/main.go` (and every `main` in the repo) has `var _ = archgate.DesignHash`. No file → undefined identifier → build fails. The file is gitignored: a fresh clone doesn't compile until `archilles diff` passes once, and any change to design or imports that fails the diff removes it again.

`go generate ./...` runs `//go:generate archilles diff --gate` so the standard workflow regenerates it. `make build` / `mage build` are `archilles diff --gate && go build`.

An agent can hand-write the file. That's why layers 2 and 3 exist — layer 1 is friction and fast feedback, not security.

**2. Analyzer (runs inside `go vet` and `go test`, no wrapper needed).**
`pkg/analyzer` implements `golang.org/x/tools/go/analysis`. It loads `resolved.json`, maps the package under analysis to its component, and reports every import that isn't an allowed edge as a diagnostic at the import line. Also checks `archgate/gate_gen.go` hashes against the current resolved design and reports mismatch.

Wired two ways: `go vet -vettool=$(which archilles-vet) ./...` and as a golangci-lint plugin (`module` plugin system). `go test` runs vet by default, so `go test ./...` fails on a forbidden import without anyone calling archilles. This is the layer that makes it feel compile-native.

**3. `-toolexec` (optional, for the paranoid or for CI builds).**
`GOFLAGS=-toolexec=archilles-toolexec` set in `go.env` for CI and in the agent runtime's environment. The wrapper intercepts every `compile` invocation, checks the package's imports against `resolved.json`, and returns non-zero before the compiler runs. Nothing gets past this without unsetting `GOFLAGS`, which is a visible act in the audit trail because the agent runtime's env is provisioned, not agent-editable.

**What it does not do.** A local actor with full control can always bypass a local gate. The compile gate's job is to make the fastest path the compliant one and make bypass an explicit, loggable action. Authority still lives in CI (check 4) and the rulesets; an artifact built outside CI is never released (`archilles-release` builds only from tagged `main`, in Actions, with layers 2 and 3 on).

**Acceptance additions.**
- Fresh clone, `go build ./...` fails with `undefined: archgate.DesignHash`; after `archilles diff --gate` it succeeds.
- Add `import "archilles/pkg/adapter/golang"` to `pkg/design`; `go test ./...` fails from the analyzer with the file:line, without running `archilles` explicitly.
- Same edit under `-toolexec` fails at compile with the same message.

## Borrow first

Before writing any component, check this list. Build only what has no existing tool, and wrap rather than fork.

| Need | Borrow | Build |
|---|---|---|
| Go dependency extraction | `go list -json -deps`, `golang.org/x/tools/go/packages` | thin mapper to `graph.Graph` |
| Go call graph (v1) | `golang.org/x/tools/go/callgraph` (cha/vta), `go/types` | tagging + collapse only |
| Go import rules today | `go-arch-lint`, `depguard` (via golangci-lint) | nothing — use them as the *engineer-side* enforcer in v0 CI; `archilles` generates their config from `resolved.json` rather than reimplementing import checks |
| Schema validation | JSON Schema via `santhosh-tekuri/jsonschema/v5`; YAML via `goccy/go-yaml` | schemas themselves |
| Graph diff/ops | `dominikbraun/graph` or `gonum/graph` if it fits; otherwise hand-rolled sets (graph is small) | diff categories |
| Mermaid | plain string emission; validate in CI with `mermaid-cli` (`mmdc`) rendering golden files | overlay conventions |
| DOT (v1) | `emicklei/dot` | — |
| CLI | cobra + viper | — |
| Lint | golangci-lint | config |
| Release | goreleaser + release-please | — |
| Dependency updates | Renovate on this repo, day one — it's the product's reference model, so live with it | — |
| PR comments | `marocchino/sticky-pull-request-comment` | comment body |
| Branch/ruleset as code | GitHub rulesets exported to `.github/rulesets/*.json`, applied by a workflow (`gh api`) | — |
| Analyzer | `golang.org/x/tools/go/analysis`, `singlechecker`/`multichecker`, golangci-lint module plugin | the rule mapping |
| Terragrunt (v1) | `hashicorp/hcl/v2`, `terragrunt graph-dependencies`, `terraform graph` | mapper |

Rule: if a dependency covers >70% of a need, take it and note the gap in `docs/backlog.md`. The moat is the design model and the diff, not a bespoke graph library.

The most important borrow: **v0 does not need to be the import-rule enforcer**. `archilles resolve --emit go-arch-lint` writes `.go-arch-lint.yml` from the resolved design; golangci-lint enforces it on every PR from day one, before `archilles diff` exists. `archilles diff` then validates that the generated config matches (`config-drift` category) and adds what the linters can't express (ownership, missing/undeclared components, dead edges, later symbol rules). Ship the generator in step 3, not step 6.

## CI/CD and branching

Trunk-based. `main` is always releasable and always passes its own architecture check.

**Branches**
- `main`: protected, no direct push, linear history (squash merge only).
- `feat/*`, `fix/*`, `chore/*`: short-lived, from `main`, PR back to `main`. No `develop`, no release branches; tags are releases.
- Conventional commits on the squash commit title; release-please turns them into changelog + tag; goreleaser builds on tag.

**Required checks on PRs to `main`** (all must pass; no admin bypass)
1. `build` — `go build ./...`, `go vet`
2. `lint` — golangci-lint incl. `depguard`/`go-arch-lint` config *generated* from `archilles/resolved.json`
3. `test` — `go test -race ./...`, coverage artifact
4. `archilles-self` — runs with `GOFLAGS=-toolexec=archilles-toolexec` and `go vet -vettool`; `archilles validate`, `archilles resolve` (no `resolved-drift`), `archilles diff --out json` (exit 0), `archilles diff --out mermaid` posted as sticky PR comment, `mmdc` renders it without error
5. `schema` — every JSON Schema in `schema/` validates its fixtures in `testdata/`
6. `fixtures` — `archilles diff` on `testdata/repos/violating` produces the golden `diff.json` exactly

Step 4 is the point of the repo: the tool gates its own PRs with its own design. A PR that adds a forbidden import fails on the tool's own output before a human looks at it. Until step 6 of the build order exists, check 4 runs whatever subset is built (`validate` first, then `resolve`, then `diff`) — it is never skipped, it grows.

**Ownership**
- `CODEOWNERS`: `archilles/**` and `schema/**` → James. Required review from owner for those paths. Everything else: one approval.
- Agent identities (Claude Code, any bot) push branches and open PRs only; no write to `main`, no approval rights, no write to `archilles/**` (enforced by ruleset path restrictions, not by instruction).

**Identities: one GitHub App per agent role**

No PATs, no shared bot user. Each agent role gets its own GitHub App with the minimum permission set for its job; the App's installation token is what the agent runs with. Roles are distinguishable in the audit log, in `CODEOWNERS`/ruleset bypass lists, and in ledger records (`source.type: agent, source.app: <app-slug>`).

| App | Job | Repo permissions | Can push to | Can approve | Bypass rulesets | Notes |
|---|---|---|---|---|---|---|
| `archilles-engineer` | implement tasks, open PRs | contents: write (branches), pull_requests: write, issues: read, checks: read | `feat/*`, `fix/*`, `chore/*` only | no | none | Path rule: cannot modify `archilles/**`, `schema/**`, `.github/**`. Enforced by ruleset, so a write to those paths on any branch is rejected at push. |
| `archilles-architect` | propose design changes | contents: write (branches), pull_requests: write | `design/*` only | no | none | The *only* App allowed to touch `archilles/**` on a branch. Still cannot merge; owner review required. |
| `archilles-orchestrator` | plan, delegate, file change requests, comment | issues: write, pull_requests: write (comment/label), contents: read, projects: write | nothing | no | none | Deliberately no `contents: write`. It coordinates; it does not commit. Change requests are issues with the `design-change` template. |
| `archilles-bot` | run diff, post comments, open drift PRs (v2) | contents: write (branches), pull_requests: write, checks: write, issues: write | `archilles-bot/*` | no | none | Renovate-shaped. Later also the federation stitcher. |
| `archilles-release` | release-please + goreleaser | contents: write, pull_requests: write, packages: write | `release-please--*`, tags | no | tag creation only | Only App with tag rights. |
| CI (Actions OIDC) | run checks | `GITHUB_TOKEN` scoped per job; `contents: read` default, `checks: write` on the archilles-self job, `pull-requests: write` only on the comment step | nothing | no | none | `permissions:` block set per job, never workflow-wide. |

Humans: James is the only approver for `archilles/**`, `schema/**`, `.github/**`. Everything else: one human approval, and Apps never count toward it.

Keys: App private keys live in the secret store the agent runtime uses (Actions secrets for bot/release; the orchestrator runtime's secret manager for the agent Apps), rotated on a schedule, one key per environment. An agent is never handed a key for a role it doesn't hold — an engineer run has no way to obtain an architect token.

**Branch protections (`.github/rulesets/main.json`, applied by workflow)**

`main` ruleset:
- Restrict creations/deletions; block force push.
- Require PR; required approvals: 1; dismiss stale approvals on push; require review from CODEOWNERS; require conversation resolution.
- Require status checks (strict, up-to-date branch): `build`, `lint`, `test`, `archilles-self`, `schema`, `fixtures`.
- Require linear history; squash merge only (repo setting); require signed commits (Apps sign automatically; humans configure it).
- Bypass actors: **none**. Not even admin. Fixing the ruleset is a PR to the ruleset file.

Path ruleset (applies to all branches, evaluated on push):
- `archilles/**`, `schema/**`: only `archilles-architect` App and repo admins may push changes; every other actor is rejected.
- `.github/**`: repo admins only.
- `archilles/resolved.json`, `resolved/**`: additionally requires the `archilles-self` check to have produced the file (enforced by check 4 comparing to fold; the path rule keeps agents out, the check keeps humans honest).

Branch-name ruleset:
- `feat/*|fix/*|chore/*`: `archilles-engineer`, humans.
- `design/*`: `archilles-architect`, humans.
- `archilles-bot/*`: `archilles-bot`.
- `release-please--*`: `archilles-release`.
- Anything else: rejected.

Tag ruleset:
- `v*`: creation by `archilles-release` only; no deletion; no update.

**Rulesets as code**
- Branch protection and path rules live in `.github/rulesets/main.json`; a workflow applies them on change to that file (owner-only path). The repo's governance is reviewable in a PR like everything else.

**Release**
- release-please PR → merge → tag → goreleaser → binaries + `schema/` bundle attached. `archilles.lock` in downstream repos pins the tag.

**Renovate**
- Installed day one with `config:recommended` + Go manager + automerge for patch on green. Its dashboard issue is the UX reference for the eventual archilles bot.

This is the repo's own change-control board: design and governance paths behind owner review, everything else behind checks that include the tool itself.

## First goal: define this project's architecture

Before any Go code, author `archilles/` for this repo by hand — `design.yaml`, the principles, and the initial decision records. This is the product's first real design and its first fixture. It must be written in the artifact format the tool will later read, so the format gets pressure-tested on a real codebase before the parser exists.

Deliverables for goal 1:

- `archilles/design.yaml`: every component in the layout above, with `path`, `tags.layer`, and the explicit allowed edges.
- `archilles/principles/`: at minimum `layer` (entry → adapter → core → model, or whatever falls out), `forbid-edge` from core to adapter (adapters depend on core, never the reverse), `forbid-edge` from anything to `cmd/`.
- `archilles/decisions/`: one record per non-obvious choice already made — library-first, data-not-DSL, four kinds only, Go adapter first, Mermaid as sole v0 visual, default-deny edges. Each cites source `{type: human, ref: "handoff"}`.
- `archilles/resolved.json`: hand-written fold of the above. When step 3 lands, `archilles resolve` must reproduce it byte-for-byte; a mismatch is a bug in either the resolver or the hand fold — find out which.
- `docs/ARCHITECTURE.md`: Mermaid rendering of the intended graph, hand-drawn, in the same conventions the renderer will emit. Becomes the golden file for the renderer test.

Review the design with me before writing code. Then proceed.

## Build order

1. `pkg/design` types + JSON Schema + loader + validator. Tests on fixtures.
2. `pkg/graph` model + diff. Table-driven tests, no I/O.
3. `pkg/rules` four kinds + resolver + conflict detection. Fixture tests per kind (must-pass / must-fail graphs).
4. `pkg/adapter/golang`. Test against `testdata/repos/simple` and `testdata/repos/violating`.
5. `pkg/render/mermaid`, `pkg/render/json`.
6. `cmd/archilles` wiring, `--gate` flag, `cmd/archilles-vet` (analyzer via `singlechecker`), `cmd/archilles-toolexec`. `archilles init` last.
7. Dogfood: wire the remaining pieces of check 4 against the `archilles/` written in goal 1, confirm `resolve` reproduces the hand fold byte-for-byte, confirm the generated go-arch-lint config matches the hand-written one from step 0.

Step 0 (before 1): repo scaffold — `go.mod`, cobra skeleton that prints version, golangci-lint config with a *hand-written* `.go-arch-lint.yml` derived from goal 1's design, Renovate config, release-please, goreleaser, rulesets JSON, CODEOWNERS, CI with checks 1–3 and the `validate`-only form of check 4. Merge this first. Every later PR is gated by it.

Don't start step 6 before 1–5 have tests. Don't touch multi-repo, GitHub App, PR bot, exceptions workflow, or any second language.

## Roadmap after v0

- v1: File + Symbol extraction for Go, `placement` / `logic-boundary` / `interface-home` kinds, `--focus` Mermaid slices, DOT/JSON full codegraph.
- v1: Terragrunt/Terraform adapter, static HCL first.
- v2: `archilles.lock` + federation bot, cross-repo diff, contract edges.
- v2: PR bot (Renovate model), exception workflow, principle promotion.

## Non-goals (v0)

Hosted app, Renovate-style PR opening, ledger compaction/promotion, principle presets/`extends`, ReactFlow UI, Rego/CUE escape hatch, any *implementation* of: multi-repo federation, Terragrunt adapter, File/Symbol-level extraction, `placement` / `logic-boundary` / `interface-home` kinds. All of those have their *models and schemas* fixed in v0 (Node.Level, Node.Parent, archilles.lock, adapter interface, diff.json) so v1 adds code, not rewrites.

## Acceptance

- `archilles diff` on `testdata/repos/violating` reports exactly the seeded violations with file:line evidence, exit 1.
- `archilles diff` on this repo exits 0 in CI.
- Editing `resolved.json` by hand fails CI with `resolved-drift`.
- `archilles graph --overlay` renders in a GitHub markdown preview without edits.
- Every command has `--out json` and its schema in `schema/`.

## Conventions

Go 1.22+, cobra, no other CLI framework. `golangci-lint` with default set plus `gocritic`, `revive`. Errors wrapped with `%w`, no panics outside `main`. All fixture repos are real `go.mod` modules so `go list` works unmodified. Commit messages: conventional commits.
