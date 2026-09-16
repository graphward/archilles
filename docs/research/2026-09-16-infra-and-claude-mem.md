# Research: prior decision structures — infra-and-claude-mem

STATUS: research, not normative. Written by an Opus agent on 2026-09-16 from repo forensics and agent memories. Nothing here is ruled.

---

## 1. Inventory (Part A)

One line per attempt. "Created/last" from `git log --follow` where a repo exists; gagett has **no git repo at its root** (deliberate — memory file `bitbucket-gagett-access.md`: *"workspace root is deliberately NOT a git repo so it won't be committed — do not `git init` it"*), so its docs have mtimes only, no history, no diff, no possible CI.

| Repo | Path | Format | Created | Last touched | Enforced by | Accurate now |
|---|---|---|---|---|---|---|
| war | `CLAUDE.md` | prose rules ("Key Design Decisions", "Key Boundaries", "Conventions") | 2026-05-02 | 2026-05-02 | nothing (no CI, no hooks, `.claude/` = permissions only) | **partial** — "`Framework/` has no `using Godot;`" holds (grep: 0 hits); but "Project Structure (Target)" lists `src/Godot/` and one extension (`Extensions/War`); reality is `src/BlackjackUI` (MonoGame), `src/CLI`, and 5 extensions + `Framework/Flow` |
| war | `docs/architecture.md` + 5 siblings (`zone-`, `match-flow`, `modifier-`, `event-`, `resolution-`) | prose + Mermaid | 2026-05-02 | 2026-05-02 | nothing | **partial** — links all resolve; content predates the 2026-06-12 refactors |
| war | `docs/api/*.md` (7 files) | generated-style API reference | 2026-05-02 | 2026-05-02 | nothing | **no** — `src/` last changed 2026-06-12 ("Refactor CLI and MonoGame screens into pure presenters", "purge game knowledge from Framework"). `InteractiveSession`, `ReplayController`, `AutoPlayController` were added 2026-06-12 and appear **nowhere** in `docs/`. `EffectResolver`/`PhaseRunner` signatures still match |
| war | `src/Data/use_cases.md` ("130 Use Cases") | use-case catalogue | 2026-05-02 | 2026-05-02 | tests loosely (262 `[Fact]/[Theory]` vs "130 use cases") | partial |
| tgviz | `CLAUDE.md` ("Architectural principles (do not violate without explicit discussion)", 8 numbered) | prose law | 2026-03-02 | 2026-05-29 | **`.coderabbit.yaml` (AI review on PRs)** + `make boundary-test`; **not** in CI | **partial** — principle 1 ("no Terragrunt in core") holds: `grep -ril terragrunt internal/` = 0 hits. But "Running Locally: `docker-compose up`" — no docker-compose file exists in the repo |
| tgviz | `DESIGN.md` | design doc + roadmap checkboxes + "Current violations to fix" | 2026-05-29 | 2026-05-31 | partially (`make boundary-test`) | **no** — lists `internal/model/types.go` pro fields and `internal/config/` K8s fields as *current violations*; both were stripped in `df889a3`/`2c4fa9f` (2026-05-31 21:29–21:31Z) and DESIGN.md was edited **after** that (`81e0d66`, 2026-06-01 00:12Z) and still says they're outstanding. Verified: no `Health/SyncPolicy/HookResult/Lock` in `internal/model/`, no `k8s.io` in `go.mod`. Roadmap boxes for both remain `[ ]` |
| tgviz | `.coderabbit.yaml` — per-path "ARCHITECTURAL RULES" / "BOUNDARY RULES" | machine-scoped prompts to an AI reviewer | 2026-05-31 | 2026-05-31 | itself the enforcer (PR-time) | **partial** — the *enforcement config is itself stale*: rule 2 for `plugins/**` says *"Currently plugins/terragrunt/ imports internal/graph/ — this is a known violation being fixed"*; `internal/graph/` was relocated in `678ee9e` and no longer exists. Rule 4 for `internal/**` tells the reviewer to flag pro-era types already deleted |
| tgviz | `Makefile: boundary-test` + `noplugins` build tag (`cmd/iacU/cmd/plugins.go` / `plugins_none.go`) | executable check | 2026-05-31 | 2026-05-31 | **nothing** — `.github/workflows/` contains only `claude.yml` and `docs.yml`; no workflow runs `go test`, `golangci-lint`, or `boundary-test` (the test/dispatcher workflows were deleted in `32b791b`) | mechanism present and correct; **unenforced** |
| tgviz | `.golangci.yml` | lint config | 2026-03-02 | 2026-03-02 | nothing (no CI job invokes it) | n/a |
| tgviz | `docs.yml` workflow | CI | 2026-05-31 | 2026-05-31 | n/a | **no** — triggers on `docs/**` and `mkdocs.yml`; neither path exists in the repo |
| tgviz | `.qc-state.json`, `.qc-reviewed.txt` | QC ledger (PR → status/notes) | 2026-03-02 | 2026-03-03 | was a gate in the pro-era repo | **no** — records PRs #165–#300 with SonarCloud scores from the deleted 91K-line operator codebase; current repo's PRs run #1–#48 |
| gagett | `docs/borrow-first.md` | doctrine + living audit tables (Borrowed / Bespoke-but-justified / Owed-a-borrow-review) | — (mtime 2026-08-01) | 2026-08-01 | nothing mechanical; guard is "justification comment lives at the code site" | **partial** — MapLibre swap confirmed (`maplibre-gl` in package.json, no Leaflet); all 10 cited code paths resolve. But its own guard fails: `spa/src/auth/AuthContext.tsx` has **no** justification comment for the bespoke JWT scheduler. And "no localStorage" is violated by `Fleet.tsx`/`DeviceDetail.tsx` tour flags |
| gagett | `docs/config-model.md` ("Decided 2026-08-01", org-vs-user table) | decision doc | — (mtime 2026-08-01) | 2026-08-01 | nothing | **partial** — `uiConfig`/`uiPrefs`/`stationary` all present in `spa/src/`; SPA's last commit is also 2026-08-01, so code and doc stopped together |
| gagett | `docs/access-architecture.html`, `docs/production-readiness.html/.drawio` | published HTML/drawio artifacts | — | 2026-07-21 / 2026-08-03 | nothing | not spot-checked beyond reference resolution |
| gagett | `docs/research/*.md` (2) | research spikes | — | 2026-08-01 | nothing | n/a |
| gagett | `.claude/` | `settings.local.json` only | — | 2026-08-28 | n/a | **no agents, skills, or rules files** |
| tg-example | `README.md` (18KB: "Best practices…", "Where to store configuration", "How is the code organized") | upstream Gruntwork convention doc | 2025-04-09 (upstream) | 2026-02-16 (James's one commit: LocalStack `root.hcl`) | `CODEOWNERS` (upstream owners `@yhakbar @denis256`) | upstream doc, unmodified; James's only change is `root.hcl` |
| claude-remote | `README.md` | install/layout notes | — (not a git repo) | 2026-06-26 | nothing | contains no "must"/"never"/"decided"/principle statements at all |

**Absent:** no `ADR/`, no `adr-*.md`, no `RFC`, no `DECISIONS.md`, no `docs/architecture/` directory, and no dependency-boundary lint (import-linter, depcruise, archunit) in any of the six named repos. `.claude/` in gagett, war and tgviz holds only permission settings — **zero agents, skills or rule files across all three**.

### Terragrunt (reference material, not James's attempt)
`/home/james/source/terragrunt` is a fork of `gruntwork-io/terragrunt` (origin upstream, `fork` = `jameslett/terragrunt`), last commit 2026-05-28.
- **RFCs are GitHub issues, not files.** The only ADR/RFC artifact in-tree is `.github/ISSUE_TEMPLATE/03-rfc.yml`, with an explicit lifecycle: labels `rfc` + `pending-decision` → `accepted` or `rejected`, implementation PRs say `Closes #<RFC number>`. No decision text ever lands in the repo.
- **Behavior is codified as per-item data, not prose**: `docs/src/data/{flags,experiments,strict-controls,commands,compatibility,changelog}/*.mdx` — one file per flag / experiment / strict control, rendered by an Astro site. These are hand-written, **not** generated from the Go source; I found no test or CI job asserting that every `internal/strict/controls` entry has a doc.
- **Enforcement is generic, not architectural**: 28 workflows (lint, markdownlint, codespell, license-check, coverage-compare, fuzz, precommit, go-mod-tidy-check), `.pre-commit-config.yaml` (tofu-fmt, goimports), `CODEOWNERS`, `.markdownlint-cli2.yaml`, SonarCloud. Doc updating is a **PR-template checkbox** (`- [ ] Update the docs.` / `- [ ] Update the changelog in the docs.`) — social, not mechanical.
- Nothing enforces layering or module boundaries.

---

## 2. Memory and observation evidence

Sources: in-repo memory `~/.claude/projects/-home-james-source-{gagett,war}/memory/`; claude-mem SQLite `~/.claude-mem/claude-mem.db` (24,582 `observations`, 2,897 `user_prompts`, 3,296 `session_summaries`; projects: cardloop 12,889 · legbone 7,846 · source 2,703 · gagett 986 · archilles 149 · james 9). **There are no observations for `war` or `tgviz`** — those projects predate or sit outside the capture window; all DB evidence comes from legbone, cardloop, gagett, source, archilles.

### Theme: docs going stale while code moved
- `~/.claude/projects/-home-james-source-war/memory/MEMORY.md` (last written 2026-06-12, the same day as the refactor): *"docs/api/*.md are stale after the 2026-06 architecture cleanup (regenerate)"* — and under User Preferences: *"Values integration testing over docs"*. The staleness was **recorded and never acted on**; three months later the docs are still 2026-05-02.
- obs `#228` [2026-07-12|legbone]: *"Book documentation stale with 59 references to old crate names… 59 hits for pre-split crate names, 37 hits for new ng9_* names; architecture/crates.md has extensive old-naming diagrams"*. Outcome, obs `#241`: *"mdBook documentation system removed; docs restart planned — Committed book deletion (17 files + config)"*. The stale doc system was **deleted, not repaired**.
- obs `#214` [2026-07-12|legbone]: stale pre-rename crate names *"in LIVE gate configs: clippy.toml, .debtmap.toml, Makefile"* — the enforcement files themselves carried dead vocabulary.
- obs `#4009` [2026-07-18|legbone]: *"Audit of architecture documentation identifies 12 spec-vs-reality divergences… seven competing docs contain obsolete crate rosters, sequence numbering, and two-bin vs one-bin conflicts… Roster counts drift: 18 vs 19 vs 22 depending on document. Migration step numbering incompatible across three docs."*
- obs `#1540` [2026-07-15|legbone]: 264 backlog tasks triaged → *"15 DONE, 30 OBSOLETE, 105+ STALE (need re-anchor)"*. obs `#3589`: *"Backlog records stale-at-velocity."*
- obs `#5698` [2026-07-21|legbone]: a task in progress since July 6 *"targets sim/src/lib.rs, which no longer exists… 222 pending tasks… may contain other path-stale work items."*

### Theme: dangling references
- `gagett-dev-dashboard-workflow.md` cites `docs/fleet-tracking-plan.md` — **file does not exist** (the only broken reference of 15 I checked across gagett memory).
- obs `#216`/`#217`/`#223` [2026-07-12|legbone]: *"CRITICAL: record_lint.py path bug — hardcoded docs/DECISIONS.md but file is at docs/archive/DECISIONS.md… Commit 97933aa5 'harness v2 demolition' archived decision ledger; tool still references old path"* — **the enforcement tool broke because the document it policed moved.**
- obs `#4083` [2026-07-18|legbone]: the published "Atlas" artifact *"contains stale 'in flight' reference to ng9_items carve"*; obs `#4025`: *"contains ng9_sim (4×) and ng9_server (2×) despite design stating 'no ng9_sim, ever'."*
- tgviz, verified directly: `docs.yml` watches paths that don't exist; `CLAUDE.md` tells you to run `docker-compose up` with no compose file; `.coderabbit.yaml` instructs the reviewer about `internal/graph/`, deleted.

### Theme: enforcement — what had teeth and what didn't
- `legbone/tools/dep_graph_lint.py` header: *"had no teeth. This gives ARCHITECTURE.md's conventions 2 and 12 a machine. Reads the tier map from ARCHITECTURE.md's 'Graph law' fenced block"* — and `ARCHITECTURE.md §5`: *"Conventions 2 and 12 stopped being prose on 2026-07-18 (incident: the items carve shipped a sideways mechanism edge only manual review caught). …**this block is the single source; the lint parses it**, so changing the law means editing this doc. …New workspace members MUST be added here with a tier, or the gate fails loud."*
- `legbone/ARCHITECTURE.md` "Principles (P-ledger)": *"Each principle names its enforcement tier: **[gate]** = mechanical check exists, **[review]** = reviewer judgment, **[scribe]** = record-time check."* (18 principles P-01..P-18.)
- obs `#4061` [2026-07-18|legbone]: *"record_lint.py moved into merge-gate to catch DECISIONS.md errors pre-merge… DEC-0152 bug: high-water DECISIONS records with no entry body silently passed merge-gate but failed at deploy. Bug silently red for 3 consecutive merges."*
- obs `#4065`: record_lint's live findings — *"digest 'CURRENT HIGH-WATER says DEC-0180 but the actual highest entry is DEC-0181'… ARCHITECTURE.md freshness stamp lags by 1 DEC… one-directional supersession links across 14 entries (marked superseded but the superseding entries don't reciprocally reference them)… Research documents missing `Distilled-into:` freshness stamps: 15 docs."*
- obs `#1259`/`#4832` [legbone]: the schema pin test — *"`abilities_schema_json_is_pinned_to_the_loader_dto` regenerates schema from AbilityDef DTO and asserts committed file matches… Failure mode: 'schema drift at abilities.schema.json: regenerated by this test run — inspect and commit'."* It actually fired on 2026-07-19 and caught real drift.
- The counter-example, obs `#723` [2026-07-13|legbone]: *"Importer columns list is documented to match schema.json but has no automated sync pin… Documentation requires importer columns to appear 'exactly as in <catalog>.schema.json' but grep finds no automated assertion… This creates a potential drift vector."*
- `gagett/docs/borrow-first.md` uses the weakest possible enforcement — *"Guard: Justification comment lives at the code site — keep it"* — and the comment is already missing from `AuthContext.tsx`.

### Theme: user intent (James's own words, `user_prompts` table)
- [2026-09-16|archilles] *"okay, I have experienced much drift in my codebase, and much unmanageble and untracable decisions, written in markdown thats not verifibly, graphing the intended architecture is nice, and enforcing my decisions and being able to DETECT a change without reading a 10000 line markdown diff, is really the goal"*
- [2026-09-16|archilles] *"what I DO need to enforce is that everything related to a decision is updated without a massive codebase grep of everything referencing etc etc etc running through existing decision files updating them, not knowing where to look, not fully updating or considering the file to meet it"*
- [2026-09-16|archilles] *"codifying decision frameworks and enforcing them, I know these decision frameworks exist but I understand have largely failed and are hard to enforce"*
- [2026-07-12|legbone] *"so we have documentation gates as well right?"* and, the same day, *"if my book is stale, then how are my docs being gated???? are they seperate?"*
- [2026-07-18|legbone] *"why is it so difficult in the current repo to get you to respect the architecture"*
- [2026-07-18|legbone] *"would it be better, to squash the current docs? extract tenants? and combine decisions?"*
- [2026-08-30|cardloop] *"i need you to go through and clean up the decisions and sort of the north star, spin up an opus to read and summarize and find contradictions and old news"*
- [2026-08-25|cardloop] *"i want you to write the schema, and the decision, but schema first as the contract"*
- [2026-09-10|cardloop] *"okay, then it's a code quality measure, if it DOES not follow the same pattern to make a decision it's wrong, because then I could parse it"*
- [2026-09-11|cardloop] *"many of these 'decisions' could have been better represented and collapsed as a graph"*
- [2026-07-31|source] *"or, design good code in the first place that accurately represents the architecture"*
- [2026-09-15|archilles] *"you've changed the architecture, and an agents goal could be to get the pr mergable"* / *"we're shifting architecture left here"* / *"we have two concerns graphing the architecture change and enforcing it's intended state and you are mixing the concerns"*
- Scope caution, [2026-09-15|archilles]: *"many of our decisions we will NOT want to be the enforcer of? if a decision is 80% unit test coverage with n unit test frameworks etc"*

### Theme: frustration with the record itself
- [2026-08-21|cardloop] *"stop making stuff up **stop recording decisions**, it's a juice engine… do not let any of these agents touch the code until you actually understand"*
- [2026-09-04|cardloop] *"sorry but uh, how are you going to write decisions I did not independently authorize?"*
- [2026-08-30|cardloop] *"do not record that as a decision until you can fully restate the game loop"* and *"why are we commitying that is was already a decision, it was part of the game that just got lost"*
- [2026-09-15|archilles] *"no I did not make that decision, you are doing EXACTLY what I am fighting"*
- [2026-09-02|cardloop] *"generally speaking if I gloss over a decision it's because it was wordy and a large paragraphs, instead of a short summary of what my consent is, how to fix this?"*

### Addendum — the two deepest prior attempts, surfaced by Part B (not in the Part A scope, but on disk)
Both are far more developed than any of the six named repos, and the DB evidence is dominated by them.

**legbone** (`/home/james/source/legbone`, last commit 2026-07-27):
- `docs/archive/DECISIONS.md` — 317 KB / 2,538 lines, **181 DEC entries** (157 Accepted, 23 Superseded, 1 Rejected), append-only, explicit grammar: *"Append, never edit an Accepted body — to change a decision, add a new entry and flip the old entry's Status line to 'Superseded by DEC-NNNN'. A `> ⚠ CONFLICT` block marks two sources that disagree with no recorded adjudication."* 26 supersession links, 1 conflict block, 10 `EXCEPTION to P-NN` tags.
- **It was demoted on 2026-07-18** — its own header now reads: *"DECISIONS — append-only history (demoted 2026-07-18, docs squash). Current truth lives in repo-root `ARCHITECTURE.md`… this file is the append-only DEC ledger only: rationale archaeology, never current state."*
- `ARCHITECTURE.md` (20 KB) = conventions law (12 rules), P-ledger (18 principles, each tagged `[gate]`/`[review]`/`[scribe]`), roster, migration ledger, graph law, superseded-artifact index.
- Enforcement, `scripts/gate.sh` stage 0: `record_lint.py` (43 KB), `dep_graph_lint.py`, `convention12_lint.py`, `adopt_guard.py` + self-test, `map-lint.sh`, ast-grep (`sgconfig.yml`), `rustqual.toml` (including orphan-suppression detection that flags *stale justifications*), cargo-machete, cargo-deny, then clippy/rustdoc/coverage.
- **The decisive A/B, inside one file:** `ARCHITECTURE.md §5`'s fenced ```` ```graph-law ```` block is parsed by the lint — every crate it names still exists today. `§5b`, twenty lines below, is hand-written prose describing the "measured dependency graph" — it names `ng9_kernel`, `ng9_display`, `legbone_prelude`, `legbone_protocol`, `ng9_net`, `server-bevy`; **all six are gone from `crates/`.** Same file, same author, same day; the parsed half stayed true and the prose half rotted.

**cardloop** (`/home/james/source/cardloop`, last commit 2026-09-11):
- `DECISIONS.md` — **668 KB**, 123 `##` entries numbered to DEC-0153, newest at the bottom, bodies never rewritten.
- A hand-maintained **"Supersession index"** table at the top (added 2026-08-30, obs `#18912`, in direct response to James's "find contradictions and old news"): *"A chronological log means later rulings kill earlier ones without editing them. Before acting on any entry below, check it here."* Vocabulary in use: *killed by*, *parked by / DORMANT*, *explicit repeal*, *RULED, NOT EXECUTED*, *COMPLETED, the doc is attic'd*, *"the arc self-edits"*.
- **The index is already stale:** its rows stop at DEC-0114; the log runs to DEC-0153. 39 entries uncovered. `grep DECISIONS scripts/*.mjs` → **nothing**; no gate reads it (unlike legbone's `record_lint.py`).
- The endpoint of the arc, DEC-0151/0152 (2026-09-10): decisions became **data**. `packages/decision/` — a Zod-discriminated six-kind decision schema, one interpreter (`decide.ts`), a PyTorch batch interpreter, and a conformance check (obs `#23849`: *"1106 placement cases and 18 landing cases match exactly between core and GPU"*), plus DEC-0151 §6, one graph: *"nodes (deduped states), edges (decisions), paths (recorded runs)"* — 1,714 nodes / 2,792 edges / 25 named paths. CLAUDE.md now says: *"EVERY DECISION OVER THE BOARD IS A DOCUMENT IN THE DECISION FORM… all future decisions must follow the form or are defects."*

---

## 3. Synthesis — **inference**, labeled as such

### What failed, and the mechanism
1. **Prose adjacent to code has no forcing function, so it decays at the rate the code changes.** war is the clean control: docs written 2026-05-02, code refactored 2026-06-12, zero doc commits, three new first-class framework types undocumented. gagett is the other control: SPA and docs both froze on 2026-08-01, so the docs *look* accurate — accuracy came from the code stopping, not from the doc working.
2. **The rot starts inside the enforcement artifacts themselves.** This is the non-obvious one, and it recurs in every repo: tgviz's `.coderabbit.yaml` instructs a reviewer about a package deleted the same week; tgviz's `docs.yml` watches paths that don't exist; legbone's `clippy.toml`/`.debtmap.toml`/`Makefile` carried pre-rename crate names (obs `#214`); `record_lint.py` broke outright when its target file was archived (obs `#216`). A rule file is just another document unless something reads it *and fails*.
3. **Unenforced mechanisms decay to zero even when the mechanism is correct.** tgviz built the exact right check — `noplugins` build tag, `make boundary-test`, a serve test — then deleted every CI workflow that could run it (`32b791b`). The check works today by luck.
4. **A doc that states both law and status will rot at the status half.** tgviz DESIGN.md's principles are still true; its "Current violations" list was false within three hours of being edited. legbone `§5` (law, parsed) is true; `§5b` (measured status, prose) is false. **Inference: mixing normative claims and observational claims in one artifact guarantees the artifact is partly wrong at all times, which then discredits the true half.**
5. **Append-only ledgers solve authorship and lose navigability.** 181 entries / 317 KB (legbone) and 123 entries / 668 KB (cardloop) is exactly James's *"10000 line markdown diff"*. Both projects independently invented the same rescue — a hand-maintained supersession index / a demotion to "rationale archaeology" with a separate current-state doc — and in cardloop that index itself went stale within 11 days because **nothing parsed it**. Legbone's equivalent (the digest high-water marker) *was* parsed, and stayed correct.
6. **Records of intent drift from the code independently of the decision text.** The 105+ STALE / 30 OBSOLETE backlog triage and the task pointing at `sim/src/lib.rs` (deleted) are the same failure as doc drift, in a different store: any record naming a path or symbol is a dangling pointer waiting to happen, and grep-based repair is what James says he refuses to keep doing.
7. **Agent-authored decision records poisoned trust in the record.** *"stop recording decisions"*, *"how are you going to write decisions I did not independently authorize?"*, *"no I did not make that decision, you are doing EXACTLY what I am fighting."* A store that anything can write becomes a store nobody believes. This is a failure of the *write path*, not the format.

### What succeeded, and why
1. **The doc-as-lint-input pattern — the single strongest result in the corpus.** legbone's ```` ```graph-law ```` fenced block is human-readable law *and* the lint's only input: *"this block is the single source; the lint parses it, so changing the law means editing this doc… New workspace members MUST be added here with a tier, or the gate fails loud."* It cannot drift, because drifting fails the build, and it cannot be bypassed by editing code without editing the law. Its neighbor prose in the same file rotted. Note it was born from an incident, not from planning.
2. **Pin tests (regenerate-and-compare).** `schema_json_is_pinned_to_the_loader_dto` regenerates the artifact from the source of truth, writes the new version to disk, and fails with an actionable message. It caught real drift twice. Its absence is diagnosed in the same repo (obs `#723`: *"no automated pin… schema drift is not automatically detected"*). **Inference: the test that *shows you the diff it wants* costs nothing to obey; the rule that just says "keep these in sync" is never obeyed.**
3. **Record-grammar linting at the merge gate.** `record_lint.py` checks structural properties of the ledger — high-water consistency, reference resolution, supersession reciprocity, freshness stamps — not semantics. It found 14 one-directional supersession links and a 1-DEC-lagging freshness stamp that no human would have. Critically, moving it from deploy-lane to merge-gate mattered: it had been *"silently red for 3 consecutive merges."* Enforcement at the wrong point in the pipeline ≈ no enforcement.
4. **Tiering enforcement explicitly.** `[gate]` / `[review]` / `[scribe]` per principle is the only vocabulary in the corpus that makes "we know this one isn't enforced" a first-class, visible fact rather than a silent lie. It matches James's own scope caution about decisions he does *not* want to be the enforcer of.
5. **Collapsing decisions into executable form.** cardloop's terminal move (DEC-0151/0152) — a decision schema, one interpreter, cross-implementation conformance, and a graph where edges *are* decisions — is the only attempt where a decision cannot be stale, because the code is the decision. It is also the most expensive, and it only works for decisions that are *evaluable*.
6. **Deleting a rotted record beat repairing it.** legbone's mdBook (59 stale refs) was deleted outright; the DEC ledger was demoted to archaeology. Both times, consolidating to one current-state artifact plus supersession banners on the old ones restored trust faster than incremental fixes.

### (a) How decisions should be stored — inference
- **Split the record along the "does it change when code changes?" axis.** Normative law (stable, small, parseable) separate from observed state (regenerated, never hand-written) separate from rationale (append-only, immutable, explicitly demoted to archaeology). Every failure above is a violation of this split.
- **The normative part must be structured data in a human-readable file, and must be the *only* input to its checker.** Not "the doc describes what the lint does" — the lint reads the doc. YAML/HCL/fenced-block all work; James's stated preference is YAML ("*is yaml the answer vs hcl? I find yaml easier to read*"), and `design.yaml` "*would REQUIRE a change eventually, if you change the design architecture and graph*" is exactly this pattern restated.
- **One current-state document, versioned in git, with supersession as a link, not a rewrite.** The append-only ledger is right for authorship and wrong for reading; both legbone and cardloop converged on ledger + index/current-state. The index must be **generated**, not maintained (cardloop's hand-maintained one died in 11 days).
- **Every record needs a machine-checkable anchor into the code** (crate/package/path/symbol/tier), because that is what makes "what else must change with this decision" answerable without a grep — James's explicit ask. Records that only contain prose are the ones that go dangling.
- **Guard the write path.** Given the "decisions I did not authorize" complaints, authorship/consent needs to be a field, and proposed ≠ ratified needs to be a status an agent cannot flip.

### (b) Mechanical vs review — inference
Mechanical, because they are cheap, total, and were observed to work:
- **dependency/layering edges** (dep_graph_lint: mechanism→mechanism forbidden, no engine→game edge, new members must declare a tier); tgviz's boundary test is the same idea in one build tag
- **naming/vocabulary containment** (convention12: no game name in engine sources; "no `terragrunt` in `internal/`" is a two-line grep and is still true today)
- **generated-artifact pinning** (schema ↔ DTO, docs data ↔ code registry, index ↔ ledger) — regenerate and diff
- **record integrity** — id monotonicity, every reference resolves, supersession is reciprocal, status vocabulary is closed, freshness stamps present, high-water consistent, no orphan/stale justifications
- **completeness coupling** — new unit of X must appear in Y (new crate → graph-law tier; new flag → its doc file, which Terragrunt notably does *not* enforce)

Review (or `[scribe]`), because the corpus shows attempts to mechanize them either failed or were never tried:
- semantic fidelity of a decision to its rationale; whether a new abstraction is earned ("*resist abstraction for its own sake*")
- "is this borrowed or bespoke, and is the bespoke justified" — gagett's doctrine is genuinely judgment; only the *presence* of a justification marker at the code site is mechanizable, and that's what failed silently
- ratification itself. James's *"many of our decisions we will NOT want to be the enforcer of"* is a hard scope boundary: coverage thresholds, framework choices, style — leave them declared and untested rather than half-enforced. A silently-unenforced rule is worse than an openly-unenforced one.

### (c) The relationship vocabulary that actually mattered — inference
From the records that were exercised in anger (legbone DECISIONS/ARCHITECTURE, cardloop's index, record_lint's checks), in rough order of load-bearing-ness:

1. **supersedes / superseded-by** — the only relation both projects independently built tooling around; must be **bidirectional** (record_lint's most common finding was one-directional links across 14 entries), and must carry a *kind*: repealed, killed, **parked/DORMANT** ("dormant ≠ dead: dormant entries await a fresh ruling"), **ruled-but-not-executed**, completed.
2. **decision → principle, with `EXCEPTION to P-NN`** — the relation that keeps a principle honest: it accumulates its own scars ("*mirrored onto the principle's own `Exceptions:` line so each principle displays its scars in one glance*"). 10 in use. Without it, principles quietly become false.
3. **record → enforcement tier (`[gate]` / `[review]` / `[scribe]`) → the actual check** — turns "is this enforced?" from a guess into a link, and makes unenforced rules visible instead of aspirational.
4. **record → code anchor** (crate/tier/path/symbol). This is the relation whose *absence* causes every stale-pointer failure above, and the one James names directly as the thing he needs so a decision change propagates without a repo-wide grep.
5. **composed-into / distilled-into (+ freshness stamp)** — legbone's mechanism for retiring research docs into the record: a `SUPERSEDED` banner with named drift notes plus a `Distilled-into:` stamp that record_lint audits for absence. This is what let 7 competing docs collapse into 1 without losing the reasoning.
6. **conflict** — an explicit `⚠ CONFLICT` marker for "two sources disagree, no adjudication recorded," resolved only by appending a new record. Only 1 in use in 181 entries, but it is the right primitive for the state James kept discovering by hand ("12 spec-vs-reality divergences", "find contradictions and old news").
7. **derives-from / same-concept collapse** — cardloop's late realization (*"many of these 'decisions' could have been better represented and collapsed as a graph"*, *"bands, curves, and lanes are one function concept"*). Weakly supported by tooling anywhere, but it is the relation that would have prevented 668 KB.

Not observed to matter: dates, authors-as-prose, ADR templates' "Context/Consequences" prose sections. Nothing in the corpus ever failed for lack of those; everything failed for lack of a machine that reads the record.