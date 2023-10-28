#!/usr/bin/env bash

set -o errexit -o nounset

function find_change_log() {
  local file="$1"
  local version="$2"
  local changelogs
  changelogs="$(cat "$file")"
  local i
  i="$(echo "$changelogs" | grep -n "\[$version" | cut -d: -f1)"
  if [[ -z "$i" ]]; then
    echo "cannot find version $version" >&2 && return
  fi
  local j
  j="$(echo "$changelogs" | tail -n +"$i" | grep -n "\-\-\-" | cut -d: -f1 | head -n 1)"
  if [[ -z "$j" ]]; then
    echo "cannot find the end of $version's changelog" >&2 && exit 1
  fi
  echo "$changelogs" | tail -n +"$i" | head -n "$j"
}

version=${1:-$VERSION}
changelog="$(find_change_log "./CHANGELOG.md" "$version")"

echo "writing release note for version $version"
cat <<EOF >./RELEASE_NOTES.md
## ChangeLog
$changelog
* See [CHANGELOG.md](https://github.com/lyqingye/AutoBangumi-go/blob/$version/CHANGELOG.md) for details.
EOF
