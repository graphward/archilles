#!/usr/bin/env bash
# Apply .github/rulesets/*.json to the repo. Idempotent: matches existing
# rulesets by name and PUTs, otherwise POSTs. App ids come from repo variables
# (or the local environment) and are substituted into bypass_actors.
#
#   BOOTSTRAP=1 scripts/apply-rulesets.sh   # pre-CI subset, main.bootstrap.json
#   scripts/apply-rulesets.sh               # full target state, main.json
set -euo pipefail

REPO="${REPO:-jameslett/archilles}"
shopt -s nullglob

for f in .github/rulesets/*.json; do
  base="$(basename "$f")"
  case "$base" in
    *.org-only.json)
      echo "skip   $base (push rulesets require an org-owned repo)"
      continue
      ;;
    main.bootstrap.json)
      if [ "${BOOTSTRAP:-0}" != "1" ]; then echo "skip   $base (set BOOTSTRAP=1)"; continue; fi
      ;;
    main.json)
      if [ "${BOOTSTRAP:-0}" = "1" ]; then echo "skip   $base (BOOTSTRAP=1 in effect)"; continue; fi
      ;;
  esac

  body="$(envsubst <"$f" | jq 'del(._comment)')"

  # Drop bypass actors whose App id placeholder was never substituted, and
  # coerce the substituted ids back to numbers.
  body="$(jq '
    (.bypass_actors // []) |= map(
      select((.actor_type != "Integration") or ((.actor_id | tostring) | test("^[0-9]+$")))
      | .actor_id |= (if type == "string" then tonumber else . end)
    )' <<<"$body")"

  name="$(jq -r .name <<<"$body")"
  id="$(gh api "repos/$REPO/rulesets" --jq ".[] | select(.name==\"$name\") | .id" 2>/dev/null | head -1)"

  if [ -n "$id" ]; then
    gh api -X PUT "repos/$REPO/rulesets/$id" --input - <<<"$body" >/dev/null
    echo "update $name (id $id)"
  else
    gh api -X POST "repos/$REPO/rulesets" --input - <<<"$body" >/dev/null
    echo "create $name"
  fi
done
