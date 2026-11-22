#!/usr/bin/env bash
# Syncs hand-written code with the API type name that go-swagger generated.
#
# go-swagger derives the API type from `-A <name>` using its own initialism
# rules (e.g. "ip-ratelimit-api" -> IPRatelimitAPIAPI), so instead of
# re-implementing those rules we read the generated constructor and rewrite
# the references in the files listed below.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

OPS_DIR="internal/http/server/operations"
FILES=(internal/http/module.go)

api_file=$(grep -l -E '^func New[A-Za-z0-9_]+\(spec \*loads\.Document\)' "${OPS_DIR}"/*.go 2>/dev/null | head -n1 || true)
if [[ -z "${api_file}" ]]; then
  echo "Error: no generated API constructor found in ${OPS_DIR}; run 'make generate-api'" >&2
  exit 1
fi

api_type=$(perl -ne 'if (/^func New([A-Za-z0-9_]+)\(spec \*loads\.Document\)/) { print $1; exit }' "${api_file}")

for file in "${FILES[@]}"; do
  [[ -f "${file}" ]] || continue
  perl -pi -e "s/\\boperations\\.New[A-Za-z0-9_]+API\\(/operations.New${api_type}(/g; s/\\*operations\\.[A-Za-z0-9_]+API\\b/*operations.${api_type}/g" "${file}"
done

echo "Swagger API type: operations.${api_type}"
