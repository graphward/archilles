---
name: archilles-governance
description: Set up agent-identity governance on a repository - one GitHub App per agent role, branch/tag rulesets as code, CODEOWNERS, and the local App-auth helpers so agents commit and open PRs as their own identity rather than as the human. Use when adopting archilles on a repo, when onboarding a new agent role, or when asked to make agent actions attributable and scoped.
---

# Agent-identity governance

This skill installs the model `archilles` uses on itself: **each agent role is a GitHub App**, and it commits, pushes, opens PRs and comments under that identity. Not a shared bot user, not the human's PAT, and not a CI job — an agent running locally, authenticating as an application. Renovate's shape, driven by an agent.

The point is attribution and scope. When five roles share one identity you cannot tell from the audit log who changed what, and you cannot give the design-editing role different permissions from the code-writing one. One App per role fixes both.

## When to use this

- Adopting archilles on a new repo and wanting the design protected from the agents that read it.
- Adding a role beyond the default five.
- An agent is committing as the human and it should not be.

## The roles

The default set. Adapt the names to the project; keep the separation.

| Role | Job | Permissions | Branches | Writes design? |
|---|---|---|---|---|
| `engineer` | implement, open PRs | contents:w, pull_requests:w, issues:r, checks:r | `feat/* fix/* chore/*` | never |
| `architect` | propose design changes | contents:w, pull_requests:w | `design/*` | **only role that may** |
| `orchestrator` | plan, delegate, comment | contents:**r**, issues:w, pull_requests:w, projects:w | none | never |
| `bot` | run checks, report, drift PRs | contents:w, pull_requests:w, checks:w, issues:w | `bot/*` | never |
| `release` | tag and publish | contents:w, pull_requests:w, packages:w | `release-please--*`, tags | never |

Two deliberate asymmetries. The **orchestrator has no write access** — it coordinates, it does not commit; the moment it can make "one small fix" the separation collapses. The **architect is the only writer of the design**, and still cannot merge: it drafts, a human decides.

No role can approve. Approval is the human's, and an App must never count toward a required review.

## Setup

### 1. Create the Apps

The REST API cannot create a GitHub App from a token — GitHub only supports the [App Manifest flow](https://docs.github.com/apps/sharing-github-apps/registering-a-github-app-from-a-manifest), which needs one browser click each. `scripts/create-app.mjs` drives it so permissions come from the manifest rather than fifteen hand-set dropdowns:

```bash
for app in engineer architect orchestrator bot release; do
  node scripts/create-app.mjs "<prefix>-$app"
done
```

Each writes `.secrets/<app>/{private-key.pem,app.json,client-secret.txt,webhook-secret.txt}` at mode 0600. **Gitignore `.secrets/`.** Install each App on the repo, then record its id:

```bash
gh variable set ENGINEER_APP_ID --body <id>
```

Manifests live in `.github/app-manifests/`, one per role, permissions verbatim from the table.

### 2. Wire the agents to their identity

`scripts/as-app.sh <app> <command...>` mints a scoped, 1-hour installation token and runs the command with the App's token and bot author set, so commits are attributed to `<app>[bot]`:

```bash
scripts/as-app.sh myproj-engineer git push -u origin feat/thing
scripts/as-app.sh myproj-architect gh pr create --fill
```

`scripts/app-token.mjs <app> --env` prints exports if you need the token in a longer-lived shell. Tokens are never written to disk.

Give each agent definition (`.claude/agents/*.md`) the matching role instructions and **restrict its tools to match its permissions** — an orchestrator with `contents: read` should not have `Edit` or `Write` in its frontmatter. Platform permissions and agent tool config must agree, or the weaker one is the real policy.

### 3. Apply the rulesets

Rulesets live as JSON in `.github/rulesets/` and are applied by `scripts/apply-rulesets.sh` (idempotent — matches by name, PUTs or POSTs). Governance is reviewed as a PR like anything else.

- `main.json` — PR required, CODEOWNERS review, stale dismissal, thread resolution, squash only, linear history, no force push, no deletion, required checks. **Bypass actors: none, including admins.**
- `branch-names.json` — any branch outside the vocabulary is rejected at creation.
- `branch-owner-*.json` — per-role branch ownership. Rulesets express this by blocking the pattern for everyone and listing the one App as a bypass actor.
- `tags.json` — `v*` create-only by the release App; no update, no deletion.
- `paths.org-only.json` — restricts pushes touching the design and CI paths.

```bash
BOOTSTRAP=1 scripts/apply-rulesets.sh   # before CI jobs exist
scripts/apply-rulesets.sh               # full target state
```

## Two constraints that will bite

**Push rulesets require an org-owned repo.** The path restriction on `archilles/**`, `schema/**`, `.github/**` is a *push* ruleset, and GitHub rejects it on a personal repo:

```
422  Source public repos cannot have push rules
     Source only org-owned repos can have push rules
```

On a personal repo, `CODEOWNERS` is the only guard, and it is **review-time, not push-time**: an agent can commit to the design on its own branch and open a PR; only the merge is blocked. If push-time enforcement of the design paths is the point — and for archilles it is — the repo must live in an organization. A free org is enough.

**Required approvals deadlock solo owners.** GitHub does not let you approve your own PR. With `required_approving_review_count: 1` and one human, agent PRs work fine (the human is a different identity) but the human's own PRs can never merge, and bypass actors are empty by design. Use the bootstrap ruleset (approvals 0, everything else intact) while the human is still hand-authoring, and cut over once agents are opening the PRs.

**Commit signing**, if you enable `required_signatures`, must be configured for the human first or it rejects their commits. Apps sign automatically.

## What this does not do

A local actor with full control can bypass a local gate — unset an env var, hand-write a generated file, commit with their own key. This model does not prevent that. It makes the compliant path the fastest one and makes bypass an explicit, attributable act. Authority lives in the required checks and the rulesets, not on the developer's machine.
