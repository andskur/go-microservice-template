#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

usage() {
  cat <<'EOF'
Usage: make rename NEW_NAME=<module-or-binary-name>

NEW_NAME should be a valid Go module path or name.
Examples:
  make rename NEW_NAME=my-service
  make rename NEW_NAME=github.com/yourorg/my-service
EOF
}

NEW_MODULE=${1:-}
if [[ -z "${NEW_MODULE}" ]]; then
  usage
  exit 1
fi

NAME_REGEX='^[a-z0-9][a-z0-9-]*(/[a-z0-9][a-z0-9-]*)*$'
if [[ ! "${NEW_MODULE}" =~ ${NAME_REGEX} ]]; then
  echo "Error: NEW_NAME must match ${NAME_REGEX}" >&2
  exit 1
fi

if [[ ! -f go.mod ]]; then
  echo "Error: go.mod not found; run from repository root" >&2
  exit 1
fi

CURRENT_MODULE=$(perl -ne 'print $1 and exit if /^module\s+(.+)/' go.mod)
if [[ -z "${CURRENT_MODULE}" ]]; then
  echo "Error: could not determine current module from go.mod" >&2
  exit 1
fi

CURRENT_BASE=${CURRENT_MODULE##*/}
NEW_BASE=${NEW_MODULE##*/}

cat <<EOF
About to rename project
  Current module: ${CURRENT_MODULE}
  New module:     ${NEW_MODULE}
  Current binary: ${CURRENT_BASE}
  New binary:     ${NEW_BASE}
  Entry point:    cmd/${NEW_BASE}.go
  Will update:
    - go.mod module path
    - Go imports
    - Makefile (APP, APP_ENTRY_POINT, GITVER_PKG)
    - Entry file rename (cmd/${CURRENT_BASE}.go -> cmd/${NEW_BASE}.go)
    - Cobra root command name
    - Dockerfile binary name
    - README.md and AGENTS.md references
    - Optional: git remote URL
Proceed? (y/N):
EOF

read -r CONFIRM
if [[ ! "${CONFIRM}" =~ ^[Yy](es)?$ ]]; then
  echo "Aborted"
  exit 1
fi

update_go_mod() {
  perl -pi -e "s|^module\\s+.+$|module ${NEW_MODULE}|" go.mod
}

update_go_imports() {
  find . -type f -name '*.go' -not -path './vendor/*' -print0 \
    | xargs -0 perl -pi -e "s|\Q${CURRENT_MODULE}\E|${NEW_MODULE}|g"
}

update_makefile() {
  perl -pi -e "s|^APP:=.+$|APP:=${NEW_BASE}|" Makefile
  perl -pi -e "s|^APP_ENTRY_POINT:=cmd/.+$|APP_ENTRY_POINT:=cmd/${NEW_BASE}.go|" Makefile
  perl -pi -e "s|^GITVER_PKG:=.+$|GITVER_PKG:=${NEW_MODULE}/pkg/version|" Makefile
}

rename_entrypoint() {
  local from="cmd/${CURRENT_BASE}.go"
  local to="cmd/${NEW_BASE}.go"
  if [[ -f "${from}" && "${from}" != "${to}" ]]; then
    mv "${from}" "${to}"
  fi
}

update_cli_use() {
  local file="cmd/root/root.go"
  if [[ -f "${file}" ]]; then
    perl -pi -e "s|Use: \"\Q${CURRENT_BASE}\E\"|Use: \"${NEW_BASE}\"|" "${file}"
  fi
}

update_dockerfile() {
  local file="Dockerfile"
  [[ -f "${file}" ]] || return 0
  local docker_name
  docker_name=$(perl -ne 'if(/ENTRYPOINT \["\/(.+?)"/){print $1; exit}' "${file}")
  docker_name=${docker_name:-${CURRENT_BASE}}
  perl -pi -e "s|/app/\Q${docker_name}\E|/app/${NEW_BASE}|g; s|\"/\Q${docker_name}\E\"|\"/${NEW_BASE}\"|g" "${file}"
}

update_docs() {
  for file in README.md AGENTS.md; do
    [[ -f "${file}" ]] || continue
    perl -pi -e "s|go-\Q${CURRENT_BASE}\E|go-${NEW_BASE}|g" "${file}"
    perl -pi -e "s|\Q${CURRENT_MODULE}\E|${NEW_MODULE}|g" "${file}"
    perl -pi -e "s|\Q${CURRENT_BASE}\E|${NEW_BASE}|g" "${file}"
  done
}

update_git_remote() {
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    return
  fi

  echo -n "Update git remote URL? (y/N): "
  read -r UPDATE_REMOTE
  if [[ ! "${UPDATE_REMOTE}" =~ ^[Yy](es)?$ ]]; then
    return
  fi

  echo -n "Enter new git remote URL (leave blank to remove origin): "
  read -r NEW_REMOTE
  if [[ -z "${NEW_REMOTE}" ]]; then
    if git remote | grep -q '^origin$'; then
      git remote remove origin
      echo "Removed origin remote"
    fi
  else
    if git remote | grep -q '^origin$'; then
      git remote set-url origin "${NEW_REMOTE}"
      echo "Updated origin to ${NEW_REMOTE}"
    else
      git remote add origin "${NEW_REMOTE}"
      echo "Added origin -> ${NEW_REMOTE}"
    fi
  fi
}

run_go_mod_tidy() {
  if command -v go >/dev/null 2>&1; then
    go mod tidy
  fi
}

update_go_mod
update_go_imports
update_makefile
rename_entrypoint
update_cli_use
update_dockerfile
update_docs
update_git_remote
run_go_mod_tidy

echo ""
echo "Rename complete. Review changes with 'git diff'."
