# Governance and agent identities

> The reusable form of this lives in `.claude/skills/archilles-governance/`, which is the version other projects adopt. This document is archilles' own instantiation of it, and records the decisions specific to this repo.

The thesis of `archilles` is that humans own the design and agents never write it. That only means something if it is enforced by the platform rather than by instructions in a prompt. This document records how far that enforcement currently reaches, and where it falls back to convention.

## Agent identities

One GitHub App per agent role. No PATs, no shared bot user. Each role's installation token is what the agent runs with, so roles are distinguishable in the audit log, in ruleset bypass lists, and in ledger records (`source.type: agent, source.app: <app-slug>`).

| App | Job | Repo permissions | Branches | Approve | Bypass |
|---|---|---|---|---|---|
| `archilles-engineer` | implement tasks, open PRs | contents:write, pull_requests:write, issues:read, checks:read | `feat/*`, `fix/*`, `chore/*` | no | none |
| `archilles-architect` | propose design changes | contents:write, pull_requests:write | `design/*` | no | `design/*`, protected paths |
| `archilles-orchestrator` | plan, delegate, comment | contents:**read**, issues:write, pull_requests:write, repository_projects:write | none | no | none |
| `archilles-bot` | run diff, post comments, drift PRs | contents:write, pull_requests:write, checks:write, issues:write | `archilles-bot/*` | no | none |
| `archilles-release` | release-please + goreleaser | contents:write, pull_requests:write, packages:write | `release-please--*`, tags | no | tag creation only |

The orchestrator deliberately has no `contents: write`. It coordinates; it does not commit. Change requests are issues using the `design-change` template.

Manifests are in [`.github/app-manifests/`](../.github/app-manifests/), one file per App, permissions verbatim from the table above.

### Creating them

The REST API cannot create a GitHub App from a PAT — GitHub only supports the [App Manifest flow](https://docs.github.com/apps/sharing-github-apps/registering-a-github-app-from-a-manifest), which requires a browser click. `scripts/create-app.mjs` drives that flow so the permission set comes from the manifest rather than fifteen hand-set dropdowns:

```bash
node scripts/create-app.mjs archilles-engineer --org <org>   # browser opens, click "Create GitHub App"
```

**Create these under the org, not under `jameslett`.** A private App can only be installed on the account that owns it, and ownership is fixed at creation — creating them personally and then transferring the repo means building all five again. Do the org move first and the Apps second.

Credentials are written to `.secrets/<app>/` (gitignored, mode 0600): `private-key.pem`, `app.json`, `client-secret.txt`, `webhook-secret.txt`. Then install the App on the repo and record its id, which the rulesets reference:

```bash
gh variable set ARCHILLES_ENGINEER_APP_ID --body <id> --repo graphward/archilles
```

Repeat for each of the five.

### Key handling

App private keys live in the secret store the agent runtime uses, one key per environment, rotated on a schedule. **An agent is never handed a key for a role it does not hold**: an engineer run has no path to an architect token. That separation is the whole reason for five Apps instead of one, so a runtime that mounts all five keys into every agent has thrown away the model.

`.secrets/<app>/` is where `create-app.mjs` stages a newly created key (mode 0600, gitignored). It is a staging area for the handoff into the real store, not the store — clear it once the keys are placed.

### Acting as an App

Agents authenticate locally, not through CI. `scripts/as-app.sh` mints a repo-scoped installation token that expires in an hour and runs a command with the App's token and bot author set, so the commit is attributed to `<app>[bot]`:

```bash
scripts/as-app.sh archilles-engineer git push -u origin feat/thing
scripts/as-app.sh archilles-architect gh pr create --fill
```

`scripts/app-token.mjs <app> --env` prints the exports for a longer-lived shell. Tokens are never written to disk.

Role instructions live in [`.claude/agents/`](../.claude/agents/), one per App. **Each agent's tool list must agree with its App's permissions** — the orchestrator holds `contents: read` and correspondingly has no `Edit` or `Write` tool. Where the two disagree, the weaker one is the real policy.

## Rulesets

Governance lives in [`.github/rulesets/`](../.github/rulesets/) and is applied by `scripts/apply-rulesets.sh`. This is run by an agent, not by CI — the roles below are agents that act as their App, not workflow jobs.

| File | Effect |
|---|---|
| `main.json` | main: PR required, 1 approval, CODEOWNERS review, stale dismissal, thread resolution, squash only, linear history, no force push, no deletion, six required checks. Bypass actors: none. |
| `main.bootstrap.json` | Same minus required checks and approvals — the form that is applicable before CI exists. |
| `branch-names.json` | Any branch outside `feat/ fix/ chore/ design/ archilles-bot/ release-please--` is rejected at creation. |
| `branch-owner-design.json` | `design/*` blocked for everyone, bypassed by `archilles-architect` and admins. |
| `branch-owner-bot.json` | `archilles-bot/*` blocked for everyone, bypassed by `archilles-bot` and admins. |
| `tags.json` | `v*` tags: creation by `archilles-release` only; no update, no deletion. |
| `paths.json` | Push-time restriction on `archilles/**`, `schema/**`, `.github/**`. **Not applicable while the repo is public — see below.** |

Per-actor branch ownership is expressed the way rulesets express it: block the pattern for everybody, then list the one App as a bypass actor.

## The path restriction is not enforced, and cannot be here

This is the gap between what the handoff specifies and what the platform allows. Read it before relying on the model.

The handoff says a write to `archilles/**`, `schema/**`, or `.github/**` by any actor other than the architect App is *rejected at push*. That is a **push ruleset**, and GitHub imposes two independent conditions on those:

```
422 Validation Failed
  Source public repos cannot have push rules          <- still blocking
  Source only org-owned repos can have push rules     <- fixed by the org move
```

Transferring to `graphward` satisfied the second. The first stands: **push rulesets are unavailable on public repositories, org-owned or not.** `archilles` is public by choice, so `paths.json` cannot be applied and `scripts/apply-rulesets.sh` reports it as a loud SKIP rather than a failure.

What actually guards those paths today is `CODEOWNERS`, which acts **at review, not at push**. Concretely, an agent with `contents: write` can commit to `archilles/design.yaml` on its own branch and push it; only the *merge* to `main` is blocked, pending James's review. The design cannot be changed behind your back, but it can be changed in a branch and must be caught by a human reading the diff.

The options, none of them free:

- **Stay public, accept review-time enforcement.** The agent instructions in `.claude/agents/` state the restriction, and `CODEOWNERS` catches it at the merge. This is the current state.
- **Make the repo private or internal.** Push rules become available immediately and the handoff's model holds exactly as written. Costs the public repo.
- **Split the design into its own private repo.** The handoff's multi-repo mode already describes this: a private design repo holding `archilles/**` and `schema/**`, with the public component repo consuming it at a pinned ref. Push rules protect the private side; the public side has nothing sensitive to protect. This is the only option that keeps both properties, and it is v2 work.

Do not describe the path restriction as enforced until one of the last two is done.

The org move was still necessary for the Apps: a private App can only be installed on the account that owns it and ownership is fixed at creation, so the five Apps belong to `graphward` rather than to James personally.

## Required approvals deadlock the owner's own PRs

GitHub does not permit approving your own pull request. With `required_approving_review_count: 1` and James as the sole human in `graphward`:

- an **agent's** PR works as designed — James is a different identity and can approve it;
- **James's own** PR can never reach one approval, and bypass actors are empty by design.

So the full `main.json` is correct for the steady state where agents author the changes, and blocking for the bootstrap phase where James is hand-writing `archilles/` for goal 1. `main.bootstrap.json` exists for that phase: it keeps linear history, squash-only, no force push, no deletion, and required conversation resolution, but drops the approval count to zero. Cut over to `main.json` once the step-0 CI jobs (`build`, `lint`, `test`, `archilles-self`, `schema`, `fixtures`) report and agents are opening the PRs.

## Commit signing

`main.json` omits `required_signatures` even though the handoff calls for it, because signing is not configured on this machine and the rule would reject the owner's own commits. To enable:

```bash
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519.pub
git config --global commit.gpgsign true
gh api -X POST /user/ssh_signing_keys -f title=archilles -f "key=$(cat ~/.ssh/id_ed25519.pub)"
```

Then add `{"type": "required_signatures"}` to `main.json` and re-apply. Apps sign automatically.
