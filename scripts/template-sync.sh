#!/usr/bin/env bash
set -euo pipefail

# Template synchronization helper for go-microservice-template
# Works in downstream repositories cloned from the template.
# Provides: setup, status, fetch, diff, sync

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

TEMPLATE_REMOTE_NAME="template"
TEMPLATE_REMOTE_URL="https://github.com/andskur/go-microservice-template.git"
TEMPLATE_BRANCH="main"
TEMPLATE_VERSION_FILE=".template-version"

usage() {
  cat <<'EOF'
Usage: scripts/template-sync.sh <command> [args]

Commands:
  setup           Add template remote, fetch, create .template-version if missing
  status          Show current sync status and available updates
  fetch           Fetch latest template changes
  diff [ref]      Show differences vs template (default: template/main)
  sync [ref]      Merge template changes (default: template/main)

Examples:
  scripts/template-sync.sh setup
  scripts/template-sync.sh status
  scripts/template-sync.sh diff
  scripts/template-sync.sh diff v1.2.0
  scripts/template-sync.sh sync        # merges template/main
  scripts/template-sync.sh sync v1.2.0 # merges tag v1.2.0 from template

Notes:
- Requires a clean working tree (no uncommitted changes).
- For conflicts, resolve manually then re-run status/diff to verify.
EOF
}

require_clean_tree() {
  if [[ -n "$(git status --porcelain)" ]]; then
    echo "Error: working tree is not clean. Commit or stash changes first." >&2
    exit 1
  fi
}

ensure_remote() {
  if git remote get-url "${TEMPLATE_REMOTE_NAME}" >/dev/null 2>&1; then
    return
  fi
  echo "Adding template remote ${TEMPLATE_REMOTE_NAME} -> ${TEMPLATE_REMOTE_URL}"
  git remote add "${TEMPLATE_REMOTE_NAME}" "${TEMPLATE_REMOTE_URL}"
}

fetch_template() {
  echo "Fetching from ${TEMPLATE_REMOTE_NAME}"
  git fetch "${TEMPLATE_REMOTE_NAME}" --prune
}

init_version_file() {
  if [[ -f "${TEMPLATE_VERSION_FILE}" ]]; then
    return
  fi
  cat > "${TEMPLATE_VERSION_FILE}" <<EOF
template_repo: ${TEMPLATE_REMOTE_URL}
last_sync_version: unknown
last_sync_date: $(date -u +%Y-%m-%d)
last_sync_commit: unknown
EOF
  echo "Created ${TEMPLATE_VERSION_FILE}"
}

current_template_ref() {
  local ref=${1:-"${TEMPLATE_BRANCH}"}
  echo "${TEMPLATE_REMOTE_NAME}/${ref}"
}

show_status() {
  if [[ ! -f "${TEMPLATE_VERSION_FILE}" ]]; then
    echo ".template-version missing. Run setup first."
  else
    echo "Current template version info:"
    cat "${TEMPLATE_VERSION_FILE}"
  fi

  echo "" 
  echo "Latest template tags:"
  git tag -l --list "v*" --sort=-v:refname --merged "${TEMPLATE_REMOTE_NAME}/${TEMPLATE_BRANCH}" | head -n 5 || true

  echo "" 
  echo "Commits since last sync (if known):"
  local last_commit
  last_commit=$(awk '/last_sync_commit:/ {print $2}' "${TEMPLATE_VERSION_FILE}" 2>/dev/null || true)
  if [[ -n "${last_commit}" && "${last_commit}" != "unknown" ]]; then
    git log --oneline "${last_commit}..${TEMPLATE_REMOTE_NAME}/${TEMPLATE_BRANCH}" || true
  else
    echo "(last sync commit unknown)"
  fi
}

show_diff() {
  local ref=${1:-"${TEMPLATE_BRANCH}"}
  local template_ref
  template_ref=$(current_template_ref "${ref}")
  echo "Showing diff vs ${template_ref}"
  git diff --stat "${template_ref}" || true
  echo "" 
  echo "For detailed file diff, run: git diff ${template_ref} -- <path>"
}

sync_template() {
  local ref=${1:-"${TEMPLATE_BRANCH}"}
  local template_ref
  template_ref=$(current_template_ref "${ref}")

  require_clean_tree
  echo "Merging ${template_ref} into current branch"
  git merge "${template_ref}" || {
    echo "Merge encountered conflicts. Resolve them, then update ${TEMPLATE_VERSION_FILE} manually." >&2
    exit 1
  }

  local commit_sha
  commit_sha=$(git rev-parse "${template_ref}")
  cat > "${TEMPLATE_VERSION_FILE}" <<EOF
template_repo: ${TEMPLATE_REMOTE_URL}
last_sync_version: ${ref}
last_sync_date: $(date -u +%Y-%m-%d)
last_sync_commit: ${commit_sha}
EOF
  echo "Updated ${TEMPLATE_VERSION_FILE}"
}

main() {
  local cmd=${1:-}
  case "${cmd}" in
    setup)
      require_clean_tree
      ensure_remote
      fetch_template
      init_version_file
      ;;
    status)
      ensure_remote
      fetch_template
      show_status
      ;;
    fetch)
      ensure_remote
      fetch_template
      ;;
    diff)
      ensure_remote
      fetch_template
      show_diff "${2:-}"
      ;;
    sync)
      ensure_remote
      fetch_template
      sync_template "${2:-}"
      ;;
    -h|--help|help|"")
      usage
      ;;
    *)
      echo "Unknown command: ${cmd}" >&2
      usage
      exit 1
      ;;
  esac
}

main "$@"
