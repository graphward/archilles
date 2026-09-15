#!/usr/bin/env bash
# Run a command with a GitHub App's identity: its installation token for
# gh/API calls, and its bot author on any commit it makes. This is how an
# archilles agent acts AS the App rather than as the human at the keyboard.
#
#   scripts/as-app.sh archilles-engineer git push origin feat/thing
#   scripts/as-app.sh archilles-architect gh pr create --fill
#   scripts/as-app.sh archilles-bot bash -c 'gh pr comment 3 --body "..."'
#
# Commits land as "<app>[bot]" in the audit log. The token is scoped to this
# repo and expires in an hour; nothing is written to disk.
set -euo pipefail

app="${1:-}"
if [ -z "$app" ] || [ $# -lt 2 ]; then
  echo "usage: scripts/as-app.sh <app-name> <command> [args...]" >&2
  exit 1
fi
shift

repo="${REPO:-jameslett/archilles}"
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

GH_TOKEN="$(node "$here/app-token.mjs" "$app" --repo "$repo")"
export GH_TOKEN
export GITHUB_TOKEN="$GH_TOKEN"

app_id="$(jq -r .id "$here/../.secrets/$app/app.json")"
slug="$(jq -r .slug "$here/../.secrets/$app/app.json")"

# GitHub attributes commits to the App when the author matches this exact form.
export GIT_AUTHOR_NAME="${slug}[bot]"
export GIT_AUTHOR_EMAIL="${app_id}+${slug}[bot]@users.noreply.github.com"
export GIT_COMMITTER_NAME="$GIT_AUTHOR_NAME"
export GIT_COMMITTER_EMAIL="$GIT_AUTHOR_EMAIL"
export ARCHILLES_ACTING_AS="$app"

# Push over HTTPS with the installation token instead of the human's SSH key.
export GIT_CONFIG_COUNT=2
export GIT_CONFIG_KEY_0="url.https://x-access-token:${GH_TOKEN}@github.com/.insteadOf"
export GIT_CONFIG_VALUE_0="git@github.com:"
export GIT_CONFIG_KEY_1="url.https://x-access-token:${GH_TOKEN}@github.com/.insteadOf"
export GIT_CONFIG_VALUE_1="https://github.com/"

exec "$@"
