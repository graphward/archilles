# Research: prior decision structures — cardloop-cardng9

STATUS: research, not normative. Written by an Opus agent on 2026-09-16 from repo forensics and agent memories. Nothing here is ruled.

---

I have complete evidence across both repos and the memory store. Report follows.

---

# Prior attempts at codifying architecture decisions — cardloop & cardng9

Two projects by the same author, four months apart, that tried nearly opposite approaches. cardloop's decision log is the largest and healthiest artifact of its kind here; cardng9's imported spec system died in 24 hours. The contrast is the finding.

## 1. Inventory

### cardloop (`/home/james/source/cardloop`) — 903 commits, 2026-08-15 → 2026-09-11

| Artifact | Format | Created → last touched | Enforced by | Accurate? |
|---|---|---|---|---|
| `DECISIONS.md` | Append-only log, `DEC-####`, 8,909 lines / 668KB, 152 entries | 2026-08-15 → 2026-09-10, **191 commits** | Nothing directly — but **1,436 source files cite DEC ids** | **Yes.** Append-only ("bodies below are NEVER rewritten"); all 126 DEC ids cited in code resolve; **zero dangling** |
| └ Supersession index (`DECISIONS.md:5-54`) | Hand-maintained two-column map | 2026-08-30 → 2026-08-30, **1 commit** | Nothing | **No.** Covers 59 of 152 DECs. **DEC-0115…0153 — all 39, the entire September run — are absent.** DEC-0151→0152 supersession appears only in prose |
| `CLAUDE.md` | Agent entry-point rules | 2026-08-15 → 2026-09-10, 20 commits | Partial (mirrors eslint/depcruise) | **Partial.** Names `core/src/decision/` and `decision/decide.ts`; **neither exists** — DEC-0152 moved it to `packages/decision/` *the same day* |
| `README.md` "The rules that keep it stable" | Prose rules | 2026-08-16 → 2026-09-01, 3 commits | Nothing | **Partial.** Claims "the sim's golden baseline pins behavior"; golden deleted 2026-09-01 (`chore(sim): delete the golden — the corpus is the proof`) |
| `docs/DESIGN.md` | "North star", current-law-only | 2026-08-15 → 2026-09-10, 16 commits | Nothing | **Partial.** Structural claims verified true (`state.tiles` span-indexed, laps dead). But omits clash/reel/collapse/JANI (DEC-0144…0153) while claiming "the LATEST law" |
| `docs/LEXICON.md` | House-term ↔ industry-term table | 2026-08-21 → 2026-08-30, 8 commits | Nothing | **No.** Its own "STANDING RULE: coin a term, add a row" is broken. **404 commits since.** Zero rows for `rung` (67 files), `jani` (40), `overkill` (44), `collapse` (23), `clash` (14), `reel` (8) |
| `ISSUES.md` | `ISS-####` findings register | 2026-08-15 → 2026-08-30, 19 commits | Nothing | **No / orphaned.** Superseded in practice by BACKLOG.md; referenced by **0 lines** of CLAUDE.md or BACKLOG.md. 46 ISS ids cited in code, **4 dangling** (ISS-0040/0042/0044/0061) |
| `BACKLOG.md` | Owed-work ledger | 2026-09-01 → 2026-09-09, 78 commits | Nothing | **Partial.** "Last swept: 2026-09-09"; 17 commits since |
| `docs/attic/*.md` | 4 demolished contracts, 6,354 lines | frozen 2026-08-21 / 08-30 | N/A — header says "HISTORICAL, NOT NORMATIVE" | **N/A by design** — the success case |
| `docs/research/*`, `docs/design/*` | Write-once dated research | 1–5 commits each, never revisited | N/A — self-labeled "STATUS: MUSINGS", "Nothing here is ruled" | **N/A by design** |
| `eslint.config.mjs` | 41 rules: core determinism, `max-lines` 300, `utils.ts` ban, render one-author | 2026-08-15 → 2026-09-09, 25 commits | **Yes** — CI + pre-commit | **Yes** |
| `.dependency-cruiser.cjs` | 13 layer-boundary rules, each commenting its DEC | 2026-08-15 → 2026-09-10, 5 commits | **Yes** | **Yes** |
| `scripts/check.mjs` | 3-tier gate; carries 4 dated owner quotes inline | → 2026-09-10 | **Yes** (husky + CI) | **Yes** |
| `packages/decision` (DEC-0151/2/3) | Decisions as executable JANI data | 2026-09-10 | Its own fixtures; **`docs/frames/conform.py` is in no gate** | Live |

Mechanical gates: 41 eslint rules, 13 depcruise rules, knip, jscpd, commitlint, 2 husky hooks, 90/90 coverage floor, 487 test files, 7 CI steps.

### cardng9 (`/home/james/source/cardng9`) — 288 commits, 2026-07-28 → 2026-08-13

| Artifact | Format | Created → last touched | Enforced by | Accurate? |
|---|---|---|---|---|
| `openspec/` | OpenSpec: proposal/design/tasks + 5 delta specs, `SHALL`/`WHEN`/`THEN`, **no IDs** | 2026-07-28 → **2026-07-28**, **4 commits** | **Nothing** — the `openspec` CLI is not even installed | **No.** Never completed one lifecycle: **no `openspec/specs/`**, **0 of 23 tasks ticked**, change never archived. **84% of commits came after** |
| `openspec/config.yaml` | Project constitution + scope boundary | 2026-07-28 | Nothing | **No.** "DISPLAY ONLY. No rules engine, no turn structure…" — `packages/table` ships verbs/turn/phase; `content/` holds 8 games |
| `.claude/skills/openspec-*`, `commands/opsx/*` | 12 vendored files (`author: openspec`) | 2026-07-28, 1 commit | — | Never used |
| `docs/architecture.md` | System map + placement law | 2026-07-28 → **2026-07-31**, 4 commits | Nothing | **No.** 4 of 5 prescribed concept dirs (`substrate/ actions/ derive/ chrome/`) **missing**; `apps/editor/src` is 61 flat files; **`TablePage.tsx` is 100KB** — the god-file the doc was written to prevent |
| `docs/model.md` | Format spec | → 2026-07-30, 10 commits | Nothing | **No.** "collection" doesn't exist (renamed `RosterDoc`); `CharterDoc` undocumented |
| `docs/stack.md` | ADR in all but name | → 2026-07-28, 2 commits | Nothing | **Partial.** Still names react-konva as *the* decision; `architecture.md` says "konva was dropped early"; no konva dependency |
| `docs/verbs.md` | Ruling + roadmap, written from 4 playthroughs | 2026-08-01, 2 commits | The ratchets | **Yes — 12/12 predicted verbs landed** |
| `schemas/design-file.schema.json` | Generated, 126KB | gitignored 2026-07-31 ("A generated schema is not source") | Regeneration wired to nothing | **No.** 13 days stale; **135 of 143 defs named `__schema0…134`** |
| `profile.test.ts` / `playability.test.tsx` "ratchets" | Thresholds as assertions, dated rationale per change, **plus a counter-ratchet** | → HEAD 2026-08-13 | `npm test` only — **CI never runs tests** | **Yes — still being fed on the last day** |

Absent in cardng9: no eslint, no prettier, no dependency-cruiser, no tsconfig `paths`/`references`, no nx/turbo, no git hooks, no CLAUDE.md, no ADR/RFC dir. **CI is `docker build` and nothing else.** Docs froze 2026-08-01; **144 of 288 commits (exactly half) landed after**, including the two busiest days in project history.

### Memory store (`~/.claude/projects/-home-james-source-cardloop/memory/`) — 41 files

Typed frontmatter (`feedback` 25, `project` 11, `reference` 4). `[[wikilink]]` graph: 73 links, 27 targets, **zero dangling**. `MEMORY.md` indexes 39 of 40 siblings (one orphan: `pixel-art-pipeline.md`). All DEC ids cited in memory resolve. Two decay signals: 9 files lack `node_type`, **clustered at the newest dates** (2026-09-03 ×4, 09-10, 09-11 ×2) — the schema is being abandoned, not adopted; and `backlog-file.md` is stale (below).

## 2. Memory evidence

### Theme: what made the user kill documentation

`memory/no-docs-waves.md` — the founding frustration, after ~90 minutes of agent docs waves in one day:

> The user (2026-08-19), after ~90 min of agent docs waves in one day: **"2 hours and 15 is waaay too much time."** Earlier same day they killed CONTRACT.md itself: **"why would contract md exist at all? should the schema not describe it's contract itself?"**

And the diagnosis, same file:

> **Why:** docs waves existed to keep a 4,000-line markdown contract self-consistent — a second codebase no check protects. Each ruling paid a full-file sweep, and **same-day rulings superseded each other's sweeps.** Pure ceremony once the doc is demoted.

`DECISIONS.md:3266-3268` records the same rulings verbatim, adding a third:

> **"why the hell are we writing schemas in markdown and not in.... schemas?"**

### Theme: the rule that came out of it

`DECISIONS.md:3272-3276` (DEC-0053 §1) — the load-bearing sentence in the whole corpus:

> **If it is NORMATIVE, it lives where a CHECK can fail it. If it is HISTORY, it lives in DECISIONS.md. There is no third home.** … docs/CONTRACT.md was the third home, and **everything in a third home is unverified by construction** — the day's three docs waves each re-swept it because no check could.

And §3: **"No agent ever writes documentation again."**

### Theme: user intent — who may author a decision

`memory/review-decisions-with-user.md` carries three escalating corrections, all verbatim:

> The user pushed back: **"well, you didn't review them with me."**

> **Addendum (2026-08-26): conversation is not a greenlight.** … the user interrupted: **"no that's not a greenlit that's conversation."**

> **Addendum 2 (2026-08-26): filing into the repo needs its own confirmation.** … the user interrupted: **"why are you filing anything without confirming it with me?"** A go for WORK … is not a go for FILING its artifacts into the repo.

> **Addendum 3 (2026-09-04): "put those up as decisions" means put them up FOR the owner.** I appended two DEC entries … whose clauses were mostly my proposals, not the owner's words, and moved to commit; the owner interrupted: **"how are you going to write decisions I did not independently authorize?"** … **A DEC clause is the owner's ruling quoted, or a proposal the owner has said yes to by line — never a design I filled in.**

This is why `DECISIONS.md` uses the word "verbatim" 142 times, and why one entry (`DECISIONS.md:6084`) carries a self-doubt marker: *"Recorded without an owner quote; strike this clause if wrong."*

### Theme: enforcement — what the user wanted mechanized

`memory/unify-never-bespoke.md`, an explicit mechanism preference:

> **Prefer the TYPE system over lint for enforcement** (branded knobs so only the reader unwraps a number). Order lives in ONE file (the phase list), never as a priority field on content rows.

The same hierarchy appears as a promotion in `eslint.config.mjs:402` — a convention becoming a mechanism, with an issue id recording the move:

> `// "utils.ts is a lint failure" — mechanically, not just by convention (ISS-0028).`

And `eslint.config.mjs:133` closes the escape hatch:

> `// Inline config comments (eslint-disable, rule toggles) could silently exempt a core file from the gates below; core files may never opt themselves out (ISS-0022).`

The single best artifact in either repo is `eslint.config.mjs:389` — the ruling, its provenance, and its enforcement are one object, so a violator reads the owner's own words:

> `"one author per drawn property (owner, 2026-08-30: 'the two author problem again, this is NOT the way') — seat it in the plan/place path, or ride a named channel through the property's one composer…"`

### Theme: enforcement the user tore back out

`scripts/check.mjs:50` and `:26`:

> `// THE CORPUS CHECK IS NOT HERE (owner, 2026-09-08: _"take the corpus check out, it's a balance check... I can do most of this by me playing it"_).`

> `the CONTENT LANE (owner, 2026-09-07: "I feel like we're running the test 100 times for content-only changes")`

`scripts/check.mjs:~110`:

> `(owner, 2026-09-01: "why do I keep finding agents running full gates? for 10 minutes at a time, fixing tiny fixes?")`

> `AN AGENT'S GATE IS ON A BUDGET (owner, 2026-09-08: _"I want better test management for agents"_, after two waves' gates put twelve cores at 100% under them)`

Gates were repeatedly *narrowed* on cost grounds — never on correctness grounds.

### Theme: the gap a decision log doesn't cover

`memory/backlog-file.md` — why a fourth record had to exist:

> On 2026-09-01 the owner asked **"track this somewhere as a running list"** … **Why:** rulings live in DECISIONS.md but **nothing listed what the law still owed**; the owner **kept re-discovering dropped asks** (map change, reopen, story, elite).

The supersession index names the same gap directly: `DEC-0065 (energy arm retires) | RULED, NOT EXECUTED — queued`.

### Theme: drift, recorded inside the memory store itself

Memory files do carry in-place supersession markers:

- `memory/platform-posture.md:11`: *"SUPERSEDED AND RE-RULED 2026-08-24 (DEC-0079). The old posture (DEC-0069…)"*
- `memory/rules-engine-dec-0152.md:38`: *"SUPERSEDED THE SAME EVENING BY DEC-0153 (commit 8a8a614): the home-grown query form goes"*

…but inconsistently. **`memory/backlog-file.md` (2026-09-02) is stale and unmarked.** It states the pre-commit hook runs `corpus:update` and stages `runs.json`, and that `corpus:check` is "a fast-tier static." Verified against code: `.husky/pre-commit` runs only `node scripts/check.mjs hook`; `STATIC_HOOK = ["lint","depcruise","typecheck"]`; `corpus` appears in `check.mjs` only in the comment explaining its removal on 2026-09-08. The code carries the correction with the owner's quote; the memory record does not.

`memory/no-flavor-naming.md`'s index line tracks dead vocabulary explicitly — *"ring/lap/lens are dead vocabulary (DEC-0092)"* — yet 88 source comment lines still say "lens", 29 say "the ring", 10 say "the lap", and 32 say "golden", all for concepts with no live code.

### Theme: velocity as the real cause

`memory/rules-engine-dec-0152.md` records a decision superseded **the same evening** it was made. DEC-0151, DEC-0152 and DEC-0153 all carry the date 2026-09-10. cardloop ruled 40 decisions in September alone. `DECISIONS.md` grew 3.6KB → 668KB in 27 days, monotonically.

## 3. Synthesis (inference)

Everything below is my interpretation, not recorded fact.

### What failed, and the mechanism

**1. Derived indexes rot; source records don't.** This is the cleanest natural experiment in the corpus. `DECISIONS.md` is append-only and was touched 191 times — **zero dangling references from 1,436 citing files.** Its hand-maintained supersession index, in the same file, was written once and covers 39% of entries. `ISSUES.md`, whose triage passes *compressed entries into an archive*, produced **4 dangling references** — because a compression pass deletes the very ids that code cites. The mechanism is simple: **a record you only ever append to cannot break an inbound link; a record you rewrite, compress, or summarize breaks links every time you tidy it.**

**2. The rot rate tracks reading frequency, not writing discipline.** `MEMORY.md` (read at every session start) indexes 39/40 files with zero dangling links. The supersession index (consulted only when someone remembers) covers 39%. Same author, same week, same hand-maintained-map format. The difference is that one is load-bearing for the next session's first action and the other is optional.

**3. A doc that names directories dies faster than one that names behavior.** cardng9's `architecture.md` prescribed five concept directories; four never existed, and the 100KB god-file it was explicitly written to prevent grew through 78 of the next 144 commits. Its sibling `docs/verbs.md` — which named *verbs* and was extracted from four observed playthroughs — went 12/12. Structure claims are falsified by every refactor; behavior claims survive them. Note cardng9's own confession at `architecture.md:53`: *"The design laws this structure serves live in the maintainer's memory and in commit messages"* — it named its own failure mode in the same paragraph.

**4. Full-ceremony spec systems die at their least mechanical step.** OpenSpec's lifecycle is propose → apply → sync → archive. `propose` ran once. The step that produces the durable artifact (`sync`, which creates `openspec/specs/`) is pure agent prose-merge with no mechanical guarantee — and it never ran, so the system produced only a change-delta that was never folded into anything. **The whole cost was paid on day one and the entire payoff required a second visit that never happened.** Meanwhile `tasks.md` shows work that demonstrably shipped (the npm workspace, the CI workflow) with its checkbox still unticked — proof the document was write-once from the start.

**5. Rules with no checker are decoration, and the author knows it.** cardng9 stated "Nothing in the model package SHALL import React" — a one-line lint rule — and has no linter at all. Its `config.yaml` scope boundary ("no rules engine, no turn structure") was blown through within two weeks and never amended; it simply stopped being read. cardloop's contrast is stark: 41 eslint rules, 13 depcruise rules, 487 test files, and its ground rules hold on spot-check (`utils.ts`: zero instances; `Math.random`/`Date`/`console` in core: zero).

**6. Even the sanctioned home for rationale drifts.** DEC-0053 sent rationale to "site comments citing their DEC." Those comments still describe deleted concepts: 88 lines say "lens," 32 say "golden," 29 say "the ring." And `presence.ts:13` still explains a span as `lap · ringSize + pos` — two concepts DEC-0092 and DEC-0007 killed. **Comments are only reviewed when their function is edited; a comment about a concept that was deleted elsewhere is never in anyone's diff.**

**7. Unenforced record schemas decay inside a month.** 9 of 41 memory files lack `node_type` frontmatter, clustered at the newest dates. Only 60 of 122 `## DEC-` headings carry a date — dating was adopted mid-project and never backfilled. Both are format conventions with no validator.

### What succeeded, and why

**1. Citation at the point of use, with an append-only target.** 1,436 source files cite DEC ids; **all 126 distinct ids resolve.** No tooling enforces this — it works because the target never deletes. This is the single highest-value pattern in either repo, and it cost nothing to build.

**2. Co-locating the ruling, its provenance, and its enforcement.** The render one-author eslint rule, the depcruise comments citing DECs, and `check.mjs`'s inline dated owner quotes are all the same move: **the artifact that blocks you is the artifact that explains why, in the owner's words, with a date.** These are the only rule statements in either repo that cannot drift from their enforcement, because they *are* their enforcement.

**3. Explicit non-normative labeling with a stated precedence order.** `docs/attic/` ("HISTORICAL, NOT NORMATIVE… Nothing in this file governs anything"), `docs/research/` ("Nothing here is ruled until the owner rules it"), `docs/design/` ("STATUS: MUSINGS"). These are 6,354+ lines of stale prose that cause **zero** harm, because each declares its own authority as nil. `docs/DESIGN.md` goes further with a precedence rule: *"When this page and a newer DEC disagree, the DEC wins and this page is stale."* **Staleness is only a defect in a document that claims to be current.**

**4. Thresholds as assertions.** cardng9's ratchets (`expect(card.stencilsShared).toBeGreaterThanOrEqual(13); // ratcheted UP 2026-07-31: …`) are ADRs written as executable code, each threshold change carrying its dated justification — plus an explicit *counter-ratchet* guarding against the metric being gamed. Written on day 3, still being updated in the final commit at day 17, having outlived OpenSpec by 16 days and the docs by 12. This is the one thing in cardng9 that worked, and it was the cheapest thing in it. Its one weakness: `npm test` is never run by CI, so it survived on habit alone.

**5. The endpoint: decisions as executable data.** DEC-0151/0152/0153 push the logic to its conclusion — a decision that isn't expressible in the one form is *by definition* a defect ("if it DOES not follow the same pattern to make a decision it's wrong, because then I could parse it"), and two independent interpreters are held to the same fixtures. Note the honest reporting design in `conform.py`: *"A rule the GPU has ported is CHECKED here, never trusted; a rule it has not ported is COUNTED as unported, never guessed."* **Coverage gaps are reported rather than hidden** — the correct posture for any drift detector. Its weakness: it is in no gate, and the ledger stands at 9 documents against roughly 440 decisions.

### (a) How decisions should be stored

- **Append-only, never rewritten, one id namespace.** This is the only property that produced zero broken references across 1,436 citing files. Every compression, triage, or tidy-up pass in the corpus broke something.
- **Supersession as a forward-dated new entry, plus an in-place marginal note on the old one** (the TOMBSTONE pattern). Never by editing the old entry's body.
- **Provenance is a first-class field.** Owner quote verbatim + date, or explicitly flagged as unauthorized ("strike this clause if wrong"). This was the user's single most repeated correction.
- **Derived views must be generated, not hand-maintained** — or else placed where they are read on every session, which is the only thing that kept `MEMORY.md` honest.
- **Every record declares its normative status.** Law / history / owed / musing / archived. The stale docs that caused no harm are exactly the ones that said so.
- **Separate "ruled" from "built."** BACKLOG.md exists solely because a decision log cannot answer "what does the law still owe?" — and `RULED, NOT EXECUTED` entries prove the gap is real.

### (b) Mechanical vs review

Mechanical, because they are cheap and were never bypassed: layer/import boundaries; purity constraints (no `Date`/`Math.random`/`console`); file-size and duplication ratchets; exhaustiveness (`assertNever`); banned names; **that every cited decision id resolves** (nothing does this today, and it is a five-line script); that a generated artifact is fresh; **closing the opt-out** (`noInlineConfig`) — an escape hatch converts a mechanical rule back into a convention.

Review, because they are judgment or too costly to gate: whether a decision's *intent* is honored; balance and feel (explicitly pulled out of the gate by the owner); whether a new seam should have been an extension of an existing primitive.

Two cautions the evidence supports. First, gates get narrowed on **cost**, not correctness — every gate change here was a latency complaint, so a check that is slow will be removed regardless of value. Second, **prefer types over lint where both can express the rule** — the user said so directly, and a type error cannot be waved through.

### (c) The relationship vocabulary that actually mattered

Counted in `DECISIONS.md`: `AMENDED` 29 headings (47 mentions), `supersedes`/`superseded` 41, `DORMANT` 21, `replaces` 14, `repealed` 6, `CLARIFIED` 3, `TOMBSTONE` 3, `RE-ANCHORED`, `GENERALIZED`, `COMPLETED` 1 each.

What earned its keep:

- **AMENDED** — overwhelmingly the most used, at 29 headings. DEC-0060 alone has 20 numbered amendments. Decisions are refined far more often than replaced, and the log needed a way to say so without touching the original.
- **SUPERSEDED / killed by** — but note it needs **section granularity**, not entry granularity. Real references look like `DEC-0050 §11.1`, `DEC-0089 §2`, `DEC-0063 §5`. A whole-record supersession model would have been too coarse to express most of this log.
- **DORMANT** — the genuinely novel one, and the most valuable. *"Dormant ≠ dead: dormant entries await a fresh ruling"* and *"Tombstones say dormant, never dead."* A binary alive/dead vocabulary cannot express "the intent lives, the shapes don't" (DEC-0034) — and that state is common enough to need a name.
- **RULED, NOT EXECUTED** — the decided/built distinction. Its absence is what forced BACKLOG.md into existence.
- **CLARIFIED** — narrowing scope without changing the ruling ("the cross's arity is a MAXIMUM").
- **`[[wikilinks]]`** in the memory store — 73 links, 27 targets, zero dangling. Untyped, bidirectionally useful, and the healthiest reference graph in the corpus. Its most-linked node (`review-decisions-with-user`, 10 inbound) is the one about user intent.

What did **not** matter: nothing in the corpus ever used "depends on," "relates to," or "context/consequence" section structure as a live navigational aid. The early DEC entries have Context/Decision/Consequences prose; nothing ever queried it. And OpenSpec's `ADDED`/`MODIFIED`/`REMOVED`/`RENAMED` delta vocabulary — the most formally complete relationship model in either repo — was used exactly once and never resolved, because **the operation that would have applied those deltas never ran.** A relationship vocabulary is worth only as much as the process that consumes it.