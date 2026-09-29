#!/usr/bin/env bash
# Advisory word-count check for the ADRs in this directory.
#
# The threshold is read from the "Length and section boundaries" table in
# AGENTS.md, which is the single source of truth for the number. If that
# table's shape changes, parsing below fails loudly instead of silently
# skipping the check — update this script to match.
#
# This is Advisory only: exceeding the threshold is a signal to consider
# trimming, not a build failure (AGENTS.md itself says exceeding isn't
# automatically wrong), so this script always exits 0 unless it cannot
# parse the threshold at all.
set -euo pipefail

ADR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGENTS_MD="$ADR_DIR/AGENTS.md"

warn_above="$(awk -F'|' '/\| *Words per ADR *\|/ { gsub(/ /, "", $4); print $4 }' "$AGENTS_MD")"

if [[ -z "${warn_above}" || ! "${warn_above}" =~ ^[0-9]+$ ]]; then
  echo "error: could not parse the 'Words per ADR' threshold from AGENTS.md" >&2
  echo "       (the Length and section boundaries table may have changed shape — update check-length.sh)" >&2
  exit 1
fi

echo "Advisory: ADR word count (warn above ${warn_above} words, per AGENTS.md)"
echo

for f in "$ADR_DIR"/[0-9][0-9][0-9][0-9]-*.md; do
  words=$(wc -w < "$f" | tr -d ' ')
  name=$(basename "$f")
  if (( words > warn_above )); then
    echo "  [consider trimming] ${name}: ${words} words"
  else
    echo "  [ok]                ${name}: ${words} words"
  fi
done

exit 0
