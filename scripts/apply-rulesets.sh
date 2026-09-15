#!/usr/bin/env bash
# Apply .github/rulesets/*.json to the repo. Idempotent: matches existing
# rulesets by name and PUTs, otherwise POSTs. App ids come from repo variables
# (or the local environment) and are substituted into bypass_actors.
#
#   BOOTSTRAP=1 scripts/apply-rulesets.sh   # pre-CI subset, main.bootstrap.json
#   scripts/apply-rulesets.sh               # full target state, main.json
set -euo pipefail

REPO="${REPO:-graphward/archilles}"
shopt -s nullglob

for f in .github/rulesets/*.json; do
  base="$(basename "$f")"
  case "$base" in
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
    verb=update; method=PUT; path="repos/$REPO/rulesets/$id"
  else
    verb=create; method=POST; path="repos/$REPO/rulesets"
  fi

  if err="$(gh api -X "$method" "$path" --input - <<<"$body" 2>&1)"; then
    echo "$verb $name${id:+ (id $id)}"
  elif grep -qE 'public repos cannot have push rules|org-owned repos can have push rules' <<<"$err"; then
    # Push rulesets need an org-owned, non-public repo. Both conditions must
    # hold; a public org repo still cannot have them. Say what is lost rather
    # than failing silently or implying the paths are protected.
    echo "SKIP  $name - GitHub refuses push rulesets here:" >&2
    sed 's/^/      /' <<<"$err" | grep -i 'push rules' >&2 || true
    echo "      => archilles/**, schema/**, .github/** are NOT protected at push." >&2
    echo "      => CODEOWNERS is the only guard, and it acts at review, not push." >&2
  else
    echo "FAIL  $name" >&2
    sed 's/^/      /' <<<"$err" >&2
    exit 1
  fi
done
