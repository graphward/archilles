# Archilles enforces one thing: the map matches the territory

## Context

Every language already has an architecture linter. go-arch-lint and depguard for
Go, ArchUnit for Java, eslint-boundaries for JS. They are tested, adopted, and
better at import rules than a new tool will be. Most decisions a team records are
not enforceable by any graph tool at all: coverage thresholds, naming, performance
budgets, taste.

## Decision

Archilles owns the design, the derived graph, the lineage, and the delegation. It
enforces exactly one rule itself: every package is accounted for by a component
and every component exists. Every other rule names the tool that enforces it in
`enforced_by`, or names review, or names nobody.

## Consequences

A rule with `enforced_by: codecov` is recorded, linked, and surfaced on the PRs
that touch its subtree. Archilles never runs a coverage tool. Edge rules emit
go-arch-lint config and archilles checks the emitted file has not drifted. The
only build archilles fails directly is one where the map lies.
