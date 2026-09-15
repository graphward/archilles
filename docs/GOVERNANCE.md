# Governance and agent identities

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
node scripts/create-app.mjs archilles-engineer     # browser opens, click "Create GitHub App"
```

Credentials are written to `.secrets/<app>/` (gitignored, mode 0600): `private-key.pem`, `app.json`, `client-secret.txt`, `webhook-secret.txt`. Then install the App on the repo and record its id, which the rulesets reference:

```bash
gh variable set ARCHILLES_ENGINEER_APP_ID --body <id> --repo jameslett/archilles
```

Repeat for each of the five.

### Key handling

App private keys live in the secret store the consuming runtime uses — Actions secrets for `archilles-bot` and `archilles-release`, the orchestrator runtime's secret manager for the agent Apps. One key per environment, rotated on a schedule. **An agent is never handed a key for a role it does not hold**: an engineer run has no path to an architect token. `.secrets/` is a staging area for the initial handoff into those stores, not the store itself — clear it once the keys are placed.

## Rulesets

Governance lives in [`.github/rulesets/`](../.github/rulesets/) and is applied by [`apply-rulesets.yml`](../.github/workflows/apply-rulesets.yml) on change, or locally via `scripts/apply-rulesets.sh`.

| File | Effect |
|---|---|
| `main.json` | main: PR required, 1 approval, CODEOWNERS review, stale dismissal, thread resolution, squash only, linear history, no force push, no deletion, six required checks. Bypass actors: none. |
| `main.bootstrap.json` | Same minus required checks and approvals — the form that is applicable before CI exists. |
| `branch-names.json` | Any branch outside `feat/ fix/ chore/ design/ archilles-bot/ release-please--` is rejected at creation. |
| `branch-owner-design.json` | `design/*` blocked for everyone, bypassed by `archilles-architect` and admins. |
| `branch-owner-bot.json` | `archilles-bot/*` blocked for everyone, bypassed by `archilles-bot` and admins. |
| `tags.json` | `v*` tags: creation by `archilles-release` only; no update, no deletion. |
| `paths.org-only.json` | Restricts pushes touching `archilles/**`, `schema/**`, `.github/**`. **Inert today — see below.** |

Per-actor branch ownership is expressed the way rulesets express it: block the pattern for everybody, then list the one App as a bypass actor.

## Two constraints on a personal repo

Both are consequences of `archilles` living at `jameslett/archilles` rather than in an organization. Neither is a bug in the design; they are the platform's limits.

### 1. Path restriction cannot be enforced

The handoff specifies that a write to `archilles/**`, `schema/**`, or `.github/**` by any actor other than the architect App is *rejected at push*. That is a **push ruleset**, and GitHub refuses it here:

```
422 Validation Failed
  Source public repos cannot have push rules
  Source only org-owned repos can have push rules
```

Push rulesets require an org-owned repository. Until `archilles` moves to an org, the only guard on those paths is `CODEOWNERS`, which is **review-time, not push-time**: an agent can commit to `archilles/**` on its own branch and open a PR, and nothing stops the push — only the merge. The claim "enforced by ruleset path restrictions, not by instruction" is not currently true, and `CLAUDE.md` states the restriction as an instruction as the interim fallback.

Moving the repo to a free organization restores push-time enforcement and costs nothing; `paths.org-only.json` is written and ready to apply the moment that happens.

### 2. Required approvals deadlock the owner's own PRs

GitHub does not permit approving your own pull request. With `required_approving_review_count: 1` and James as the sole human:

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
