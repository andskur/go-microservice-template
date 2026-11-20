#!/usr/bin/env bash
# Renames the service: Go module path, binary name and everything derived from them.
#
# Generated code (protobuf, swagger) is never edited in place: it is deleted and
# regenerated with `make generate` after the rename, so it does not matter
# whether `make generate` was run before `make rename` or not.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

usage() {
  cat <<'EOF'
Usage: make rename NEW_NAME=<module-path-or-name>

NEW_NAME is a Go module path; its last element becomes the binary name.
Examples:
  make rename NEW_NAME=my-service
  make rename NEW_NAME=github.com/your-org/my-service

Set SKIP_GENERATE=1 to skip regenerating code (not recommended).
EOF
}

NEW_MODULE=${1:-}
if [[ -z "${NEW_MODULE}" ]]; then
  usage
  exit 1
fi

# Module path: slash-separated elements of letters, digits and '.', '-', '_', '~'
# (the first element may be a domain such as github.com).
MODULE_REGEX='^[A-Za-z0-9][A-Za-z0-9._~-]*(/[A-Za-z0-9][A-Za-z0-9._~-]*)*$'
# Binary name (last element): lowercase letters, digits and dashes.
BASE_REGEX='^[a-z][a-z0-9-]*$'

if [[ ! "${NEW_MODULE}" =~ ${MODULE_REGEX} ]]; then
  echo "Error: NEW_NAME '${NEW_MODULE}' is not a valid module path (must match ${MODULE_REGEX})" >&2
  exit 1
fi

NEW_BASE=${NEW_MODULE##*/}
if [[ ! "${NEW_BASE}" =~ ${BASE_REGEX} ]]; then
  echo "Error: last path element '${NEW_BASE}' must match ${BASE_REGEX} (it becomes the binary name)" >&2
  exit 1
fi

if [[ ! -f go.mod ]]; then
  echo "Error: go.mod not found; run from repository root" >&2
  exit 1
fi

CURRENT_MODULE=$(perl -ne 'print $1 and exit if /^module\s+(\S+)/' go.mod)
if [[ -z "${CURRENT_MODULE}" ]]; then
  echo "Error: could not determine current module from go.mod" >&2
  exit 1
fi

# Let the go tool validate the module path as well (dots, case, reserved names).
cp go.mod go.mod.rename-bak
if ! go mod edit -module "${NEW_MODULE}" 2>/dev/null; then
  mv go.mod.rename-bak go.mod
  echo "Error: '${NEW_MODULE}' is rejected by 'go mod edit -module'" >&2
  exit 1
fi
mv go.mod.rename-bak go.mod

CURRENT_BASE=$(perl -ne 'print $1 and exit if /^APP:=(\S+)/' Makefile)
CURRENT_BASE=${CURRENT_BASE:-${CURRENT_MODULE##*/}}
CURRENT_DB="${CURRENT_BASE//-/_}_dev"
NEW_DB="${NEW_BASE//-/_}_dev"

cat <<EOF
Renaming project
  Module:   ${CURRENT_MODULE} -> ${NEW_MODULE}
  Binary:   ${CURRENT_BASE} -> ${NEW_BASE}
  Dev DB:   ${CURRENT_DB} -> ${NEW_DB}
EOF

# Files to rewrite: text files tracked by git (or all files when not a git
# checkout), excluding go.mod/go.sum, LICENSE, generated, vendored and subtree code.
FILE_LIST=$(mktemp)
trap 'rm -f "${FILE_LIST}"' EXIT

list_text_files() {
  if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git ls-files -co --exclude-standard
  else
    find . -type f -not -path './.git/*' | sed 's|^\./||'
  fi | grep -v -E '(^|/)vendor/|\.pb\.go$|^internal/http/(server|models)/|^bin/|^protocols/|^go\.(mod|sum)$|^LICENSE$' \
     | while IFS= read -r f; do
         if [[ -f "${f}" ]] && grep -Iq . "${f}" 2>/dev/null; then echo "${f}"; fi
       done > "${FILE_LIST}"
}

# Runs a perl substitution (reading FROM/TO from the environment) on all files.
substitute() {
  local from=$1 to=$2 expr=$3
  while IFS= read -r file; do
    [[ -f "${file}" ]] || continue # e.g. the entrypoint before it was moved
    FROM="${from}" TO="${to}" perl -pi -e "${expr}" "${file}"
  done < "${FILE_LIST}"
}

update_module_path() {
  go mod edit -module "${NEW_MODULE}"
  # "<module>/..." in Go imports, buf go_package prefix, example proto
  # go_package, Makefile and docs. Not preceded by a path character, so e.g.
  # ".../go-microservice-template/..." URLs are left alone.
  substitute "${CURRENT_MODULE}/" "${NEW_MODULE}/" 's/(?<![A-Za-z0-9_.\/-])\Q$ENV{FROM}\E/$ENV{TO}/g'
}

update_binary_name() {
  local from="cmd/${CURRENT_BASE}.go" to="cmd/${NEW_BASE}.go"
  if [[ -f "${from}" && "${from}" != "${to}" ]]; then
    if git rev-parse --is-inside-work-tree >/dev/null 2>&1 && git ls-files --error-unmatch "${from}" >/dev/null 2>&1; then
      git mv "${from}" "${to}"
    else
      mv "${from}" "${to}"
    fi
    echo "${to}" >> "${FILE_LIST}"
  fi

  perl -pi -e "s|^APP:=.+\$|APP:=${NEW_BASE}|; s|^APP_ENTRY_POINT:=cmd/.+\$|APP_ENTRY_POINT:=cmd/${NEW_BASE}.go|" Makefile

  # Binary name as a whole word: Dockerfile COPY/ENTRYPOINT, .gitignore,
  # cobra command name, docker-compose, docs.
  substitute "${CURRENT_BASE}" "${NEW_BASE}" 's/(?<![A-Za-z0-9_.-])\Q$ENV{FROM}\E(?![A-Za-z0-9_-])/$ENV{TO}/g'

  # Development database name (docker-compose, Makefile, config defaults, docs).
  substitute "${CURRENT_DB}" "${NEW_DB}" 's/\Q$ENV{FROM}\E/$ENV{TO}/g'
}

update_git_remote() {
  NEW_REMOTE=${NEW_REMOTE:-}
  if [[ -z "${NEW_REMOTE}" ]] || ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    return
  fi

  if git remote | grep -q '^origin$'; then
    if [[ "${NEW_REMOTE}" == "-" ]]; then
      git remote remove origin
      echo "Removed origin remote"
    else
      git remote set-url origin "${NEW_REMOTE}"
      echo "Updated origin to ${NEW_REMOTE}"
    fi
  else
    git remote add origin "${NEW_REMOTE}"
    echo "Added origin -> ${NEW_REMOTE}"
  fi
}

regenerate() {
  # Old generated code carries the old module path / API type name: drop it.
  make --no-print-directory clean-generated
  if [[ "${SKIP_GENERATE:-}" == "1" ]]; then
    echo "SKIP_GENERATE=1: run 'make generate' before building"
    return
  fi
  make --no-print-directory generate
  go mod tidy
}

list_text_files
update_module_path
update_binary_name
update_git_remote
regenerate

echo ""
echo "Rename complete. Review changes with 'git diff' and run 'go build ./... && go test ./...'."
