#!/usr/bin/env bash
# Apply GitHub branch ruleset for master (requires GitHub Pro or a public repo).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO="${GITHUB_REPO:-yarkingulacti/muxdev-cli}"
RULESET="$ROOT/scripts/github-ruleset-master.json"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required" >&2
  exit 1
fi

echo "Repository: $REPO"
echo "Ruleset:    $RULESET"
echo

existing=""
if rulesets="$(gh api "repos/$REPO/rulesets" --jq '.[] | select(.name == "Protect master") | .id' 2>/dev/null)"; then
  existing="$rulesets"
fi
if [[ -n "$existing" ]]; then
  echo "Updating ruleset id=$existing"
  gh api \
    --method PUT \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "repos/$REPO/rulesets/$existing" \
    --input "$RULESET"
else
  echo "Creating ruleset"
  gh api \
    --method POST \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "repos/$REPO/rulesets" \
    --input "$RULESET"
fi

echo
echo "Done. Verify:"
echo "  gh ruleset list -R $REPO"
echo "  gh ruleset check master -R $REPO"
